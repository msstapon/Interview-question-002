package handler

import (
	"errors"

	"github.com/labstack/echo/v4"

	"example.com/interview-question-002/internal/delivery/http/dto"
	"example.com/interview-question-002/internal/domain"
	"example.com/interview-question-002/internal/usecase"
	"example.com/interview-question-002/pkg/response"
)

type AuthHandler struct {
	auth *usecase.AuthUsecase
}

func NewAuth(a *usecase.AuthUsecase) *AuthHandler { return &AuthHandler{auth: a} }

// Register handles POST /api/v1/auth/register (IT 02-2).
func (h *AuthHandler) Register(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, "validation failed", dto.ValidationErrors(err))
	}
	u, err := h.auth.Register(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return response.Conflict(c, "USERNAME_TAKEN", "username already registered")
		}
		return response.Internal(c, "register failed")
	}
	return response.Created(c, dto.NewUserDTO(u))
}

// Login handles POST /api/v1/auth/login (IT 02-1) and returns an access token.
func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, "validation failed", dto.ValidationErrors(err))
	}
	pair, err := h.auth.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			// Combined message avoids leaking which field was wrong (user enumeration).
			return response.Unauthorized(c, "username or password is incorrect")
		}
		return response.Internal(c, "login failed")
	}
	d := dto.NewUserDTO(pair.User)
	return response.OK(c, dto.LoginResponse{
		AccessToken: pair.AccessToken,
		ExpiresIn:   pair.ExpiresIn,
		User:        &d,
	})
}
