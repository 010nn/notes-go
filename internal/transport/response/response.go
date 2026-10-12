package response

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"notes-go/internal/domain"
)

var clientErrors = map[error]int{
	domain.ErrInvalidJSON:              http.StatusBadRequest,
	domain.ErrInvalidNoteId:            http.StatusBadRequest,
	domain.ErrInvalidCredentials:       http.StatusBadRequest,
	domain.ErrUsernameValidationFailed: http.StatusBadRequest,
	domain.ErrPasswordValidationFailed: http.StatusBadRequest,
	domain.ErrNoteNotFound:             http.StatusNotFound,
	domain.ErrUserNotFound:             http.StatusNotFound,
	domain.ErrVersionMismatch:          http.StatusConflict,
	domain.ErrNoteAlreadyExists:        http.StatusConflict,
	domain.ErrUserAlreadyExists:        http.StatusConflict,
}

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	ctx := r.Context()

	for clientErr, status := range clientErrors {
		if errors.Is(err, clientErr) {
			slog.WarnContext(ctx, "business error occurred",
				slog.String("details", err.Error()),
			)
			http.Error(w, clientErr.Error(), status)
			return
		}
	}

	slog.ErrorContext(ctx, "unexpected internal server error", slog.String("error", err.Error()))

	http.Error(w, "Internal server error", http.StatusInternalServerError)
}
