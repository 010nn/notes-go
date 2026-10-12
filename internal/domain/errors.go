package domain

import "errors"

var (
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrInternal                 = errors.New("internal server error")
	ErrUsernameValidationFailed = errors.New("invalid username format")
	ErrPasswordValidationFailed = errors.New("invalid password format")
	ErrUserAlreadyExists        = errors.New("user already exists")

	ErrNoteNotFound      = errors.New("note not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrVersionMismatch   = errors.New("note version mismatch")
	ErrNoteAlreadyExists = errors.New("note already exists")
)
