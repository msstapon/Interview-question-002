package handler

import (
	"github.com/labstack/echo/v4"

	"example.com/interview-question-002/internal/delivery/http/dto"
	"example.com/interview-question-002/internal/delivery/http/middleware"
	"example.com/interview-question-002/internal/usecase"
	"example.com/interview-question-002/pkg/response"
)

type UserHandler struct {
	auth *usecase.AuthUsecase
}

func NewUser(a *usecase.AuthUsecase) *UserHandler { return &UserHandler{auth: a} }

// Me handles GET /api/v1/me (IT 02-3). The JWT is already validated by JWTAuth;
// here we reload the user so the response reflects current state.
func (h *UserHandler) Me(c echo.Context) error {
	uid, _ := c.Get(middleware.ContextUserID).(string)
	if uid == "" {
		return response.Unauthorized(c, "invalid token")
	}
	u, err := h.auth.GetUser(c.Request().Context(), uid)
	if err != nil {
		return response.Unauthorized(c, "user not found")
	}
	return response.OK(c, dto.NewUserDTO(u))
}
