# Типы и структуры в Go-бэкенде: где что объявлять

Разберу по слоям. Главная мысль: **типы принадлежат слою, который их создаёт или потребляет**. Не «все модели в одном пакете», а «каждый слой имеет свои типы, и между слоями они преобразуются».

## Слои и их типы

```
HTTP handler    →  Request / Response DTO       (транспорт)
                    ↓ toInput / toResponse
Service         →  Input / domain errors        (сценарий)
                    ↓
Domain          →  Note, User                   (сущность)
                    ↑ fromRepo / toRepo
Repository      →  DB row / scan targets        (хранение)
```

Плюс `domain` — общий пакет, который импортируют все, но который **никого не импортирует**.

## Пакеты

```
/internal/
  domain/           ← сущности и ошибки домена
    note.go
    user.go
    errors.go
  service/          ← бизнес-логика
    note_service.go
    dto.go          ← Input/Output типы сервиса
  repository/       ← доступ к данным
    note_repo.go
    models.go       ← структуры, специфичные для хранилища
  transport/http/   ← HTTP-слой
    note_handler.go
    dto.go          ← Request/Response
```

Правило импортов:
- `domain` — **никого не импортирует** (только стандартную библиотеку и, может, `uuid`).
- `service` — импортирует `domain`.
- `repository` — импортирует `domain`.
- `transport/http` — импортирует `service` и `domain` (для маппинга ошибок).
- `service` **не** импортирует `transport` и **не** импортирует `repository` (только интерфейс).

## 1. Domain — сущность

Здесь живут **бизнес-сущности** в том виде, в каком они существуют в предметной области. Без тегов JSON, без SQL-тегов, без HTTP-полей.

```go
// internal/domain/note.go
package domain

import (
    "time"
    "github.com/google/uuid"
)

type Note struct {
    ID        uuid.UUID
    UserID    uuid.UUID
    Title     string
    Content   string
    Order     int64
    Version   int
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt *time.Time
}
```

Почему без тегов:
- `json:"..."` — забота транспорта, не домена.
- `db:"..."` — забота репозитория, не домена.
- Домен — «чистое» представление. Меняется редко, стабильно.

Исключение: если используешь `sqlx` с тегами прямо на домене — можно, но это компромисс. Правильнее — отдельная модель в репозитории (см. п. 4).

### Ошибки домена

```go
// internal/domain/errors.go
package domain

import "errors"

var (
    ErrNoteNotFound = errors.New("note not found")
    ErrVersionConflict = errors.New("version conflict")
    ErrForbidden = errors.New("forbidden")
    ErrInvalidInput = errors.New("invalid input")
)
```

Это **сентинел-ошибки** — сервис их возвращает, транспорт маппит в HTTP-статусы. Живут в домене, потому что это часть контракта бизнес-логики, а не деталь реализации.

### Методы на сущности

Можно добавлять **доменные методы** — то, что относится к самой сущности:

```go
func (n *Note) IsDeleted() bool {
    return n.DeletedAt != nil
}

func (n *Note) Rename(title string) error {
    if len(title) > 500 {
        return domain.ErrInvalidInput
    }
    n.Title = title
    return nil
}
```

Но **не** добавляй сюда `Save()`, `ToJSON()`, `ValidateForDB()` — это чужие слои.

## 2. Service — типы сценариев

Сервис описывает **сценарии**: «создать заметку», «сохранить заметку», «переместить». Каждому сценарию — свои Input/Output.

```go
// internal/service/dto.go
package service

import "github.com/google/uuid"

type SaveNoteInput struct {
    ID      uuid.UUID
    UserID  uuid.UUID
    Version int
    Title   string
    Content string
}

type MoveNoteInput struct {
    UserID       uuid.UUID
    NoteID       uuid.UUID
    BeforeNoteID *uuid.UUID  // nil = в конец
}

type ListNotesOutput struct {
    Notes []*domain.Note
}
```

Почему отдельные типы, а не `domain.Note`:
- Клиент может прислать только часть полей (нельзя задать `CreatedAt` — это решает сервер).
- Input ≠ сущность. Input — это «что хочет сделать вызывающий».
- Так проще менять API, не ломая домен.

### Интерфейсы репозиториев

Живут в сервисе — **там, где потребляются**. Это Go-идиома: интерфейс объявляется на стороне потребителя, а не поставщика.

```go
// internal/service/note_service.go
package service

import (
    "context"
    "github.com/google/uuid"
    "myapp/internal/domain"
)

type NoteRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Note, error)
    InsertIfAbsent(ctx context.Context, n *domain.Note) (*domain.Note, error)
    UpdateWithVersion(ctx context.Context, id uuid.UUID, version int, patch NotePatch) (*domain.Note, error)
    ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Note, error)
    Upsert(ctx context.Context, n *domain.Note, expectedVersion int) (*domain.Note, error)
}

type NoteService struct {
    repo NoteRepository
}
```

`NotePatch` — тоже тип сервиса:

```go
type NotePatch struct {
    Title   *string
    Content *string
}
```

Указатели — чтобы различать «не менять» и «поставить пустую строку».

## 3. Transport (HTTP) — Request/Response

Здесь живут DTO, которые **ходят по проводу**. JSON-теги, валидация формата, HTTP-специфика.

```go
// internal/transport/http/dto.go
package http

import "github.com/google/uuid"

type SaveNoteRequest struct {
    ID      string `json:"id"`
    Version int    `json:"version"`
    Title   string `json:"title"`
    Content string `json:"content"`
}

type NoteResponse struct {
    ID        string  `json:"id"`
    Title     string  `json:"title"`
    Content   string  `json:"content"`
    Order     int64   `json:"order"`
    Version   int     `json:"version"`
    CreatedAt string  `json:"createdAt"`
    UpdatedAt string  `json:"updatedAt"`
    DeletedAt *string `json:"deletedAt"`
}
```

Почему не отдавать `domain.Note` напрямую:
- Домен не знает про JSON-формат дат (`RFC3339` vs unix).
- Некоторые поля не должны уходить клиенту (`UserID` — лишнее).
- Меняется HTTP-контракт — не трогаем домен.

### Маппинг между слоями

Это **функции-переходники**, живут в транспортном пакете:

```go
// internal/transport/http/mappers.go

func (r SaveNoteRequest) ToInput(userID uuid.UUID) (service.SaveNoteInput, error) {
    id, err := uuid.Parse(r.ID)
    if err != nil {
        return service.SaveNoteInput{}, fmt.Errorf("invalid id: %w", err)
    }
    return service.SaveNoteInput{
        ID:      id,
        UserID:  userID,
        Version: r.Version,
        Title:   r.Title,
        Content: r.Content,
    }, nil
}

func toNoteResponse(n *domain.Note) NoteResponse {
    resp := NoteResponse{
        ID:        n.ID.String(),
        Title:     n.Title,
        Content:   n.Content,
        Order:     n.Order,
        Version:   n.Version,
        CreatedAt: n.CreatedAt.Format(time.RFC3339),
        UpdatedAt: n.UpdatedAt.Format(time.RFC3339),
    }
    if n.DeletedAt != nil {
        s := n.DeletedAt.Format(time.RFC3339)
        resp.DeletedAt = &s
    }
    return resp
}
```

Маппинг — это и есть граница между слоями. Всё, что не маппится — либо лишнее, либо слои неправильно разделены.

## 4. Repository — модели хранилища

Тут два варианта: **использовать `domain.Note` напрямую** или **отдельная модель для БД**.

### Вариант A: работать с `domain.Note`

```go
// internal/repository/note_repo.go
func (r *noteRepo) scan(row pgx.Row) (*domain.Note, error) {
    var n domain.Note
    err := row.Scan(&n.ID, &n.UserID, &n.Title, &n.Content,
        &n.Order, &n.Version, &n.CreatedAt, &n.UpdatedAt, &n.DeletedAt)
    return &n, err
}
```

Подходит, когда таблица 1-в-1 соответствует домену. Меньше кода, меньше маппинга.

### Вариант B: отдельная `NoteRow`

```go
// internal/repository/models.go
type noteRow struct {
    ID        uuid.UUID  `db:"id"`
    UserID    uuid.UUID  `db:"user_id"`
    Title     string     `db:"title"`
    Content   string     `db:"content"`
    Order     int64      `db:"order"`
    Version   int        `db:"version"`
    CreatedAt time.Time  `db:"created_at"`
    UpdatedAt time.Time  `db:"updated_at"`
    DeletedAt *time.Time `db:"deleted_at"`
}

func (r noteRow) toDomain() *domain.Note {
    return &domain.Note{
        ID: r.ID, UserID: r.UserID, Title: r.Title,
        Content: r.Content, Order: r.Order, Version: r.Version,
        CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt, DeletedAt: r.DeletedAt,
    }
}
```

Нужен, когда:
- БД имеет колонки/связи, которых нет в домене (denormalized fields, join-результаты).
- Домен хочется держать независимым от `sqlx`/`gorm` тегов.
- Одна таблица раскладывается на несколько доменных сущностей.

**Для заметок** — я бы начал с **варианта A**. Переусложнять не нужно. Перейдёшь на B, когда появится реальная причина.

### Промежуточные типы для операций

В репозитории могут быть свои типы для сложных запросов:

```go
// internal/repository/note_repo.go
type NeighborsResult struct {
    Prev *domain.Note
    Next *domain.Note
}
```

Не выносятся в сервис — используются только внутри репозитория.

## 5. Тонкие моменты

### Где объявлять интерфейс?

**Всегда на стороне потребителя.** `NoteRepository` в `service`, потому что сервис его вызывает. Репозиторий-пакет его не объявляет.

```go
// ❌ Неправильно: интерфейс в repository
package repository
type NoteRepository interface { ... }
func NewPostgresNoteRepo() NoteRepository { ... }

// ✅ Правильно: интерфейс в service (потребителе)
package service
type NoteRepository interface { ... }
```

Почему: тестируя сервис, ты подсовываешь мок — и он должен реализовывать интерфейс, объявленный сервисом. Если интерфейс в репозитории, сервис вынужден импортировать пакет репозитория только ради типа — плохо.

### Где объявлять конструктор?

Конструкторы (`NewXxx`) — в **пакете самого типа**:

```go
// internal/repository/postgres/note_repo.go
package postgres
func NewNoteRepo(db *pgxpool.Pool) *NoteRepo { ... }
```

`main.go` собирает всё вместе:

```go
// cmd/server/main.go
db := pgxpool.New(...)
repo := postgres.NewNoteRepo(db)
svc := service.NewNoteService(repo)
h := http.NewNoteHandler(svc)
```

### Где объявлять ошибки?

- **Доменные ошибки** (`ErrNoteNotFound`) — в `domain`.
- **Ошибки сервиса** (сценария, `ErrConflict` отличается от домена) — можно тоже в `domain`, если они часть контракта.
- **Ошибки транспорта** — в `transport`, если специфичны (например, «невалидный JSON»).

Правило: ошибка живёт там, где её **будут проверять через `errors.Is`**. Если сервис проверяет `errors.Is(err, domain.ErrNotFound)` — значит в `domain`.

### Где валидация?

- **Формат** (UUID парсится, JSON валиден, длина строки) — транспорт/хендлер.
- **Бизнес-правила** (заметка не может принадлежать чужому юзеру, лимит на 1000 заметок) — сервис.
- **Инварианты сущности** (title не пустой после trim) — метод на `domain.Note`.

Не смешивай: «UUID валиден» — это не бизнес-правило.

## Полная карта типов

| Тип | Где живёт | Знает про |
|---|---|---|
| `domain.Note` | `internal/domain` | Бизнес-поля |
| `domain.ErrNotFound` | `internal/domain` | — |
| `service.SaveNoteInput` | `internal/service` | Что хочет вызывающий |
| `service.NoteRepository` (интерфейс) | `internal/service` | Что нужно сервису от БД |
| `service.NotePatch` | `internal/service` | Частичное обновление |
| `http.SaveNoteRequest` | `internal/transport/http` | JSON, HTTP-контракт |
| `http.NoteResponse` | `internal/transport/http` | JSON, HTTP-контракт |
| `repository.NoteRow` (если нужен) | `internal/repository` | SQL-колонки |
| `repository.NeighborsResult` | `internal/repository` | Внутренняя деталь SQL |

## Движение данных

```
HTTP-запрос (JSON)
    ↓ SaveNoteRequest (transport)
    ↓ ToInput()
service.SaveNoteInput (service)
    ↓ вызов репозитория
repo.Upsert(note *domain.Note, ...)
    ↓ SQL
(БД)
    ↓ RETURNING
*domain.Note (domain)
    ↓ ToNoteResponse()
NoteResponse (transport)
    ↓ JSON
HTTP-ответ
```

Домен **проходит через все слои**, но каждый слой добавляет/убирает свой контекст. Не наоборот — не тащите `SaveNoteRequest` внутрь сервиса.

## Частые ошибки

1. **Все модели в одном пакете `models`.** Тогда непонятно, где какой слой и что можно менять. Слои деградируют в один большой.
2. **DTO транспорта в сервисе.** Сервис начинает знать про JSON-теги и HTTP.
3. **Домен с `json`/`db` тегами.** Домен становится зависимым от инфраструктуры.
4. **Один `User` на все слои.** Начинает обрастать полями «для БД», «для API», «для логов». Через месяц непонятно, какие поля зачем.
5. **Возвращать `domain.Note` из HTTP-хендлера напрямую.** Клиент видит внутренние поля, нельзя поменять формат даты без ломки домена.
6. **Интерфейс репозитория в пакете репозитория.** Сервис вынужден импортировать инфраструктуру.

## Правила в двух строках

- **Domain** — то, что не зависит ни от чего. Сущности и ошибки.
- **Каждый слой** имеет свои типы для «входа» и «выхода» и маппит их на границах.
- **Интерфейс — у потребителя**, реализация — у поставщика.
- **Теги** (`json`, `db`) — только на типах того слоя, где они нужны.
- **`main.go`** склеивает всё, больше никто не знает про конкретные реализации.

Если хочешь, могу показать, как это выглядит в реальном проекте — например, как `NoteService.Save` работает с `NoteRepository` и `domain.Note` от начала до конца, с моками для тестов.
