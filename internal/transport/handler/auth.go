package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"notes-go/internal/service"

	"github.com/gorilla/sessions"
)

type AuthHandler struct {
	service *service.UserService
	store   *sessions.CookieStore
}

func NewAuthHandler(service *service.UserService, store *sessions.CookieStore) *AuthHandler {
	return &AuthHandler{
		service: service,
		store:   store,
	}
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {

	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // максимум 1 МБ

	var auth AuthRequest

	// Декодируем JSON из тела запроса напрямую в структуру
	if err := json.NewDecoder(r.Body).Decode(&auth); err != nil {
		http.Error(w, "Invalid JSON or request size exceeded", http.StatusBadRequest)
		return
	}

	if auth.Username == "" || auth.Password == "" {
		http.Error(w, "Password or username is empty", http.StatusConflict)
		return
	}

	user, err := h.service.Login(auth.Username, auth.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	log.Printf("login user: %s %s\n", auth.Username, auth.Password)

	// Создаем сессию
	session, err := h.store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Store Get Error %v\nsession: %v", err, session)
		return
	}
	session.Values["userID"] = user.ID
	session.Values["role"] = "USER"
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Store Save Error %v", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {

	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // максимум 1 МБ

	var auth AuthRequest

	// Декодируем JSON из тела запроса напрямую в структуру
	if err := json.NewDecoder(r.Body).Decode(&auth); err != nil {
		http.Error(w, "Invalid JSON or request size exceeded", http.StatusBadRequest)
		return
	}

	if auth.Username == "" || auth.Password == "" {
		http.Error(w, "Password or username is empty", http.StatusConflict)
		return
	}

	log.Printf("register user: %s %s\n", auth.Username, auth.Password)

	if err := h.service.Register(auth.Username, auth.Password); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
