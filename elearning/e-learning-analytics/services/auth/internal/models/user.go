// internal/models/user.go
package models

import (
	"github.com/go-playground/validator/v10"
)

// Package-level validator instance for better performance
var validate = validator.New()

// User represents the persisted user document.
type User struct {
	ID       string `bson:"_id,omitempty" json:"id,omitempty"`
	Email    string `bson:"email" json:"email" validate:"required,email"`
	Password string `bson:"password" json:"-"` // hashed password
	Role     string `bson:"role" json:"role"`
	Created  any    `bson:"created" json:"created"`
}

// RegisterRequest is the payload for creating a new user.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// Validate runs request-level validation.
func (r RegisterRequest) Validate() error {
	return validate.Struct(r)
}

// LoginRequest is the payload for authentication.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (l LoginRequest) Validate() error {
	return validate.Struct(l)
}
