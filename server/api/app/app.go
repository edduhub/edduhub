package app

import (
	"fmt"
	"time"

	"eduhub/server/api/handler"
	"eduhub/server/internal/config"
	"eduhub/server/internal/middleware"
	"eduhub/server/internal/repository"
	"eduhub/server/internal/services"
	"eduhub/server/internal/services/audit"

	"github.com/labstack/echo/v4"
	echomid "github.com/labstack/echo/v4/middleware"
	"net/http"
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

func (a *App) Start() error {
	// OPTIMIZED: Configure Echo for better performance on low-end hardware
	
	// Disable unnecessary features for production
	a.e.HideBanner = true
	a.e.HidePort = false
	
	// OPTIMIZED: Use custom logger with reduced verbosity for better performance
	a.e.Use(echomid.LoggerWithConfig(echomid.LoggerConfig{
		Format: "${time_rfc3339_nano} ${status} ${method} ${uri} ${latency_human}\n",
		Output: a.e.Logger.Output(),
	}))
	
	a.e.Use(echomid.Recover())

	// Add custom error handler middleware for standardized error responses
	a.e.Use(middleware.ErrorHandlerMiddleware())

	// Add panic recovery middleware
	a.e.Use(middleware.RecoverMiddleware())

	// SECURITY: Add comprehensive security headers (OWASP best practices)
	a.e.Use(middleware.SecurityHeaders())

	// Add validator middleware for input validation
	a.e.Use(middleware.ValidatorMiddleware())

	// OPTIMIZED: Configure CORS with specific settings from config
	// Uses environment-configured origins or defaults to localhost:3000
	a.e.Use(echomid.CORSWithConfig(echomid.CORSConfig{
		AllowOrigins: a.config.AppConfig.CORSOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		MaxAge:       3600, // Cache preflight requests for 1 hour
	}))
	
	// OPTIMIZED: Add compression middleware for bandwidth optimization
	a.e.Use(echomid.GzipWithConfig(echomid.GzipConfig{
		Level: 5, // Balanced compression level (1-9, where 5 is a good balance)
	}))
	
	// OPTIMIZED: Add timeout middleware to prevent hanging requests
	a.e.Use(echomid.TimeoutWithConfig(echomid.TimeoutConfig{
		Timeout: 30 * time.Second, // 30 seconds max per request
	}))
	
	// OPTIMIZED: Rate limiting for resource protection (optional, can be configured per environment)
	// Uncomment in production:
	// a.e.Use(echomid.RateLimiter(echomid.NewRateLimiterMemoryStore(20))) // 20 requests per second

	// Add audit logging middleware for API routes
	a.e.Use(audit.AuditMiddleware(a.services.AuditService))

	// Setup routes
	handler.SetupRoutes(a.e, a.handlers, a.middleware.Auth)

	// OPTIMIZED: Configure server with reasonable limits
	a.e.Server.ReadTimeout = 10 * time.Second
	a.e.Server.WriteTimeout = 30 * time.Second
	a.e.Server.IdleTimeout = 60 * time.Second
	a.e.Server.MaxHeaderBytes = 1 << 20 // 1 MB

	return a.e.Start(":" + a.config.AppPort)
}
