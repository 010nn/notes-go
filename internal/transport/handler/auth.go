package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"notes-go/internal/domain"
	"notes-go/internal/service"
	"notes-go/internal/transport"
	"notes-go/internal/transport/response"
	"time"

	"github.com/go-playground/validator/v10"
)

// DTO
type authRequest struct {
	Username string `json:"username" validate:"required,max=32"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type authResponse struct {
	Username  string `json:"username"`
	NotesSalt string `json:"notesSalt"` // base64
}

func (req authRequest) Validate(validate *validator.Validate) error {
	return validate.Struct(req)
}

// Handlers
type AuthHandler struct {
	service    *service.UserService
	validate   *validator.Validate
	jwtManager *transport.JWTManager
}

func NewAuthHandler(service *service.UserService, jwtManager *transport.JWTManager) *AuthHandler {
	return &AuthHandler{
		service:    service,
		validate:   validator.New(),
		jwtManager: jwtManager,
	}
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username, ok := transport.UserFromCtx(ctx)
	if !ok {
		response.HandleError(w, r, domain.ErrInternal)
		return
	}

	user, err := h.service.GetUser(ctx, username)
	if err != nil {
		response.HandleError(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, authResponse{
		Username:  user.Username,
		NotesSalt: base64.StdEncoding.EncodeToString(user.NotesSalt),
	})
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
	if err := auth.Validate(h.validate); err != nil {
		response.HandleError(w, r, fmt.Errorf("%w: %w", domain.ErrValidation, err))
		return
	}

	user, token, err := h.service.Login(r.Context(), auth.Username, auth.Password)
	if err != nil {
		response.HandleError(w, r, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(h.jwtManager.AccessTTL),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	response.JSON(w, http.StatusOK, authResponse{
		Username:  user.Username,
		NotesSalt: base64.StdEncoding.EncodeToString(user.NotesSalt),
	})
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
	if err := auth.Validate(h.validate); err != nil {
		response.HandleError(w, r, fmt.Errorf("%w: %w", domain.ErrValidation, err))
		return
	}

	if err := h.service.Register(r.Context(), auth.Username, auth.Password); err != nil {
		response.HandleError(w, r, err)
		return
	}

	user, token, err := h.service.Login(r.Context(), auth.Username, auth.Password)
	if err != nil {
		response.HandleError(w, r, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(h.jwtManager.AccessTTL),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	response.JSON(w, http.StatusOK, authResponse{
		Username:  user.Username,
		NotesSalt: base64.StdEncoding.EncodeToString(user.NotesSalt),
	})
}
