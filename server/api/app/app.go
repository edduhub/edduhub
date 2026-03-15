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

func isWebSocketNotificationsPath(c echo.Context) bool {
	return c.Request().URL.Path == "/api/notifications/ws"
}

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

	a.e.Use(echomid.RequestLoggerWithConfig(echomid.RequestLoggerConfig{
		LogStatus:  true,
		LogMethod:  true,
		LogURI:     true,
		LogLatency: true,
		LogValuesFunc: func(c echo.Context, values echomid.RequestLoggerValues) error {
			_, err := fmt.Fprintf(
				a.e.Logger.Output(),
				"%s %d %s %s %s\n",
				values.StartTime.Format(time.RFC3339Nano),
				values.Status,
				values.Method,
				values.URI,
				values.Latency,
			)
			return err
		},
	}))

	a.e.Use(echomid.Recover())
	a.e.Use(middleware.ErrorHandlerMiddleware())
	a.e.Use(middleware.NewErrorSanitizationMiddleware().Middleware)
	a.e.Use(middleware.RecoverMiddleware())
	a.e.Use(middleware.SecurityHeaders())
	a.e.Use(middleware.ValidatorMiddleware())

	a.e.Use(echomid.CORSWithConfig(echomid.CORSConfig{
		AllowOrigins:     a.config.AppConfig.CORSOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Client-Version"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	a.e.Use(echomid.GzipWithConfig(echomid.GzipConfig{
		Level:   5,
		Skipper: isWebSocketNotificationsPath,
	}))

	a.e.Use(echomid.ContextTimeoutWithConfig(echomid.ContextTimeoutConfig{
		Skipper: isWebSocketNotificationsPath,
		Timeout: 30 * time.Second,
	}))

	a.e.Use(audit.AuditMiddleware(a.services.AuditService))

	handler.SetupRoutes(a.e, a.handlers, a.middleware.Auth, a.middleware.ParamValidator)

	a.e.Server.ReadTimeout = 10 * time.Second
	a.e.Server.WriteTimeout = 30 * time.Second
	a.e.Server.IdleTimeout = 60 * time.Second
	a.e.Server.MaxHeaderBytes = 1 << 20

	return a.e.Start(":" + a.config.AppPort)
}
