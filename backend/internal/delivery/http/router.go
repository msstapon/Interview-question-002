package http

import (
	"github.com/labstack/echo/v4"

	"example.com/interview-question-002/config"
	"example.com/interview-question-002/internal/delivery/http/handler"
	"example.com/interview-question-002/internal/delivery/http/middleware"
	"example.com/interview-question-002/pkg/jwt"
)

type Deps struct {
	Cfg        *config.Config
	Auth       *handler.AuthHandler
	User       *handler.UserHandler
	Health     *handler.HealthHandler
	JWTManager *jwt.Manager
}

func RegisterRoutes(e *echo.Echo, d Deps) {
	e.GET("/healthz", d.Health.Live)
	e.GET("/readyz", d.Health.Ready)

	api := e.Group("/api/v1")

	auth := api.Group("/auth")
	auth.POST("/register", d.Auth.Register)
	auth.POST("/login", d.Auth.Login)

	// protected — requires a valid Bearer access token
	protected := api.Group("", middleware.JWTAuth(d.JWTManager))
	protected.GET("/me", d.User.Me)
}
