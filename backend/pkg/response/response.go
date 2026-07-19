// Package response provides standard HTTP response shapes.
package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Envelope struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func OK(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, Envelope{Success: true, Data: data})
}

func Created(c echo.Context, data any) error {
	return c.JSON(http.StatusCreated, Envelope{Success: true, Data: data})
}

func NoContent(c echo.Context) error { return c.NoContent(http.StatusNoContent) }

func Fail(c echo.Context, status int, code, message string, details ...any) error {
	e := &Error{Code: code, Message: message}
	if len(details) > 0 {
		e.Details = details[0]
	}
	return c.JSON(status, Envelope{Success: false, Error: e})
}

func BadRequest(c echo.Context, msg string, details ...any) error {
	return Fail(c, http.StatusBadRequest, "BAD_REQUEST", msg, details...)
}
func Unauthorized(c echo.Context, msg string) error {
	return Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", msg)
}
func Forbidden(c echo.Context, msg string) error {
	return Fail(c, http.StatusForbidden, "FORBIDDEN", msg)
}
func NotFound(c echo.Context, msg string) error {
	return Fail(c, http.StatusNotFound, "NOT_FOUND", msg)
}
func Conflict(c echo.Context, code, msg string) error {
	return Fail(c, http.StatusConflict, code, msg)
}
func Internal(c echo.Context, msg string) error {
	return Fail(c, http.StatusInternalServerError, "INTERNAL", msg)
}
