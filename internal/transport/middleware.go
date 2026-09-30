package transport

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

type Middleware struct {
	store *sessions.CookieStore
}

func NewMiddleware(store *sessions.CookieStore) *Middleware {
	return &Middleware{store: store}
}

func (m *Middleware) AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := m.store.Get(r, "session-name")

		// Проверяем, есть ли ID пользователя в сессии
		userID, ok := session.Values["userID"].(int)
		if !ok || userID == 0 {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Кладем userID в контекст запроса, чтобы хэндлеры ниже могли его достать
		ctx := context.WithValue(r.Context(), "userID", userID)

		// Передаем управление следующему обработчику с обновленным контекстом
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("%s %s %s %v", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
		next.ServeHTTP(w, r)
	})
}
