package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"eduhub/server/api/handler"
	"eduhub/server/internal/config"
	"eduhub/server/internal/middleware"
	"eduhub/server/internal/repository"
	"eduhub/server/internal/services"
	"eduhub/server/internal/services/audit"

	"github.com/labstack/echo/v4"
	echomid "github.com/labstack/echo/v4/middleware"
)

type App struct {
	e          *echo.Echo
	db         *repository.DB
	config     *config.Config
	services   *services.Services
	handlers   *handler.Handlers
	middleware *middleware.Middleware
}

func New() (*App, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	if cfg.DB == nil || cfg.DB.Pool == nil {
		return nil, fmt.Errorf("database connection pool is nil")
	}
	// Initialize auth service
	services := services.NewServices(cfg)
	handlers := handler.NewHandlers(services)
	// repos := repository.NewRepository(cfg.DB)
	mid := middleware.NewMiddleware(services)

	return &App{
		e:          echo.New(),
		db:         cfg.DB,
		config:     cfg,
		services:   services,
		handlers:   handlers,
		middleware: mid,
	}, nil
}

func (a *App) Shutdown(ctx context.Context) error {
	a.services.WebSocketService.Stop()
	return a.e.Shutdown(ctx)
}

func (a *App) Start() error {
	a.e.HideBanner = true
	a.e.HidePort = false

	a.e.Use(echomid.LoggerWithConfig(echomid.LoggerConfig{
		Format: "${time_rfc3339_nano} ${status} ${method} ${uri} ${latency_human}\n",
		Output: a.e.Logger.Output(),
	}))

	a.e.Use(echomid.Recover())
	a.e.Use(middleware.ErrorHandlerMiddleware())
	a.e.Use(middleware.RecoverMiddleware())
	a.e.Use(middleware.SecurityHeaders())
	a.e.Use(middleware.ValidatorMiddleware())

	a.e.Use(echomid.CORSWithConfig(echomid.CORSConfig{
		AllowOrigins: a.config.AppConfig.CORSOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		MaxAge:       3600,
	}))

	a.e.Use(echomid.GzipWithConfig(echomid.GzipConfig{
		Level: 5,
	}))

	a.e.Use(echomid.TimeoutWithConfig(echomid.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))

	a.e.Use(audit.AuditMiddleware(a.services.AuditService))

	handler.SetupRoutes(a.e, a.handlers, a.middleware.Auth)

	a.e.Server.ReadTimeout = 10 * time.Second
	a.e.Server.WriteTimeout = 30 * time.Second
	a.e.Server.IdleTimeout = 60 * time.Second
	a.e.Server.MaxHeaderBytes = 1 << 20

	return a.e.Start(":" + a.config.AppPort)
}
