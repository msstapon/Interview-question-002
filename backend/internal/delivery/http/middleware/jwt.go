package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"

	"example.com/interview-question-002/pkg/jwt"
	"example.com/interview-question-002/pkg/response"
)

const (
	ContextUserID   = "uid"
	ContextUsername = "username"
)

// JWTAuth validates the Bearer access token and stashes uid/username in the context.
func JWTAuth(tm *jwt.Manager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				return response.Unauthorized(c, "missing bearer token")
			}
			claims, err := tm.Parse(strings.TrimPrefix(auth, "Bearer "))
			if err != nil {
				return response.Unauthorized(c, "invalid token")
			}
			c.Set(ContextUserID, claims.UserID)
			c.Set(ContextUsername, claims.Username)
			return next(c)
		}
	}
}
