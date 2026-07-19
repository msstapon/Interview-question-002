package dto

import (
	"time"

	"example.com/interview-question-002/internal/domain"
)

// RegisterRequest is the IT 02-2 form payload. confirm_password must equal password
// (validated here AND on the client).
type RegisterRequest struct {
	Username        string `json:"username" validate:"required,min=3,max=100"`
	Password        string `json:"password" validate:"required,min=8,max=128"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

// LoginRequest is the IT 02-1 form payload.
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse is returned on successful authentication.
type LoginResponse struct {
	AccessToken string   `json:"access_token"`
	ExpiresIn   int      `json:"expires_in"`
	User        *UserDTO `json:"user"`
}

// UserDTO is the external user shape.
type UserDTO struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
}

func NewUserDTO(u *domain.User) UserDTO {
	return UserDTO{
		ID:        u.ID,
		Username:  u.Username,
		CreatedAt: u.CreatedAt.UTC().Format(time.RFC3339),
	}
}
