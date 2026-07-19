package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"example.com/interview-question-002/config"
	httpdelivery "example.com/interview-question-002/internal/delivery/http"
	"example.com/interview-question-002/internal/delivery/http/handler"
	appmw "example.com/interview-question-002/internal/delivery/http/middleware"
	"example.com/interview-question-002/internal/infrastructure/db"
	"example.com/interview-question-002/internal/infrastructure/logger"
	pgrepo "example.com/interview-question-002/internal/repository/postgres"
	"example.com/interview-question-002/internal/usecase"
	jwtpkg "example.com/interview-question-002/pkg/jwt"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	zlog, err := logger.New(cfg.Obs.LogLevel, cfg.Obs.LogFormat)
	if err != nil {
		log.Fatalf("logger: %v", err)
	}
	defer func() { _ = zlog.Sync() }()

	zlog.Info("starting", zap.String("env", cfg.App.Env), zap.Int("port", cfg.App.Port))

	gormDB, err := db.NewPostgres(cfg.DB, cfg.IsProd())
	if err != nil {
		zlog.Fatal("postgres", zap.Error(err))
	}

	tm, err := jwtpkg.New(cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath, cfg.JWT.Issuer, cfg.JWT.AccessTTL)
	if err != nil {
		zlog.Fatal("jwt", zap.Error(err))
	}

	// wiring: repository -> usecase -> handler
	userRepo := pgrepo.NewUserRepo(gormDB)
	authUC := usecase.NewAuth(userRepo, tm)
	authH := handler.NewAuth(authUC)
	userH := handler.NewUser(authUC)
	healthH := handler.NewHealth(gormDB)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Validator = httpdelivery.NewValidator()
	e.Use(echomw.RequestID())
	e.Use(echomw.Recover())
	e.Use(echomw.Logger())
	e.Use(appmw.CORS(cfg.App.AllowedOrigins))

	httpdelivery.RegisterRoutes(e, httpdelivery.Deps{
		Cfg: cfg, Auth: authH, User: userH, Health: healthH, JWTManager: tm,
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.App.Port),
		Handler:           e,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		zlog.Info("listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zlog.Fatal("server", zap.Error(err))
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	zlog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		zlog.Error("server shutdown", zap.Error(err))
	}
	if sqlDB, err := gormDB.DB(); err == nil {
		_ = sqlDB.Close()
	}
	zlog.Info("bye")
}
