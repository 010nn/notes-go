package handler

import (
	"encoding/json"
	"net/http"
	"notes-go/internal/domain"
	"notes-go/internal/service"
	"notes-go/internal/transport/response"
	"time"

	"github.com/go-playground/validator/v10"
)

// DTO
type authRequest struct {
	Username string `json:"username" validate:"required,max=32"`
	Password string `json:"password" validate:"required,max=32"`
}

var validate = validator.New()

func (req authRequest) Validate() error {
	return validate.Struct(req)
}

// Handlers
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

	var auth authRequest

	// Декодируем JSON из тела запроса напрямую в структуру
	if err := json.NewDecoder(r.Body).Decode(&auth); err != nil {
		response.HandleError(w, r, domain.ErrInvalidJSON)
		return
	}
	// Validation
	if err := auth.Validate(); err != nil {
		response.HandleError(w, r, domain.ErrInvalidJSON)
		return
	}

	token, err := h.service.Login(r.Context(), auth.Username, auth.Password)
	if err != nil {
		response.HandleError(w, r, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {

	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // максимум 1 МБ

	var auth authRequest

	// Декодируем JSON из тела запроса напрямую в структуру
	if err := json.NewDecoder(r.Body).Decode(&auth); err != nil {
		response.HandleError(w, r, domain.ErrInvalidJSON)
		return
	}
	// Validation
	if err := auth.Validate(); err != nil {
		response.HandleError(w, r, domain.ErrInvalidJSON)
		return
	}

	if err := h.service.Register(r.Context(), auth.Username, auth.Password); err != nil {
		response.HandleError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
