package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"notes-go/internal/model"
	"notes-go/internal/service"
	"time"
)

type AuthHandler struct {
	service *service.UserService
}

func NewAuthHandler(service *service.UserService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {

	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // максимум 1 МБ

	var auth model.AuthRequest

	// Декодируем JSON из тела запроса напрямую в структуру
	if err := json.NewDecoder(r.Body).Decode(&auth); err != nil {
		http.Error(w, "Invalid JSON or request size exceeded", http.StatusBadRequest)
		return
	}

	log.Printf("login user: %s %s\n", auth.Username, auth.Password)

	if _, err := h.service.Login(auth); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "SUPERTOKEN",
		Value:    "nice",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {

	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // максимум 1 МБ

	var auth model.AuthRequest

	// Декодируем JSON из тела запроса напрямую в структуру
	if err := json.NewDecoder(r.Body).Decode(&auth); err != nil {
		http.Error(w, "Invalid JSON or request size exceeded", http.StatusBadRequest)
		return
	}

	log.Printf("register user: %s %s\n", auth.Username, auth.Password)

	if err := h.service.Register(auth); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
