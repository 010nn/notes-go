package model

import "errors"

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInternal = errors.New("internal server error")
var ErrUsernameValidationFailed = errors.New("invalid username format")
var ErrPasswordValidationFailed = errors.New("invalid password format")
var ErrUserAlreadyExists = errors.New("the user already exists")
