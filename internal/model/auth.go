package model

import "github.com/go-playground/validator/v10"

type AuthRequest struct {
	Username string `json:"username" validate:"required,max=32"`
	Password string `json:"password" validate:"required,max=32"`
}

var validate = validator.New()

func (req AuthRequest) Validate() error {
	return validate.Struct(req)
}
