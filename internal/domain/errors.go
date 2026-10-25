package domain

import "errors"

var (
	ErrInternal = errors.New("internal server error")

	ErrInvalidJSON              = errors.New("invalid json")
	ErrInvalidNoteId            = errors.New("invalid id")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrValidation               = errors.New("validation error")
	ErrUsernameValidationFailed = errors.New("invalid username format")
	ErrPasswordValidationFailed = errors.New("invalid password format")
	ErrNoteNotFound             = errors.New("note not found")
	ErrUserNotFound             = errors.New("user not found")
	ErrVersionMismatch          = errors.New("note version mismatch")
	ErrNoteAlreadyExists        = errors.New("note already exists")
	ErrUserAlreadyExists        = errors.New("user already exists")
)
