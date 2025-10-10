package handler

import (
	"context"
	"time"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/repository"

	"github.com/labstack/echo/v4"
)

type SystemHandler struct {
	db *repository.DB
}

func NewSystemHandler(db *repository.DB) *SystemHandler {
	return &SystemHandler{
		db: db,
	}
}

// HealthCheck performs a health check on the system
func (h *SystemHandler) HealthCheck(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	status := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "eduhub-api",
	}

	// Check database connection
	if h.db != nil && h.db.Pool != nil {
		err := h.db.Pool.Ping(ctx)
		if err != nil {
			status["status"] = "unhealthy"
			status["database"] = "unavailable"
			status["error"] = err.Error()
			return helpers.Success(c, status, 503)
		}
		status["database"] = "connected"
	}

	return helpers.Success(c, status, 200)
}

// ReadinessCheck checks if the service is ready to serve requests
func (h *SystemHandler) ReadinessCheck(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	// Check if database is ready
	if h.db != nil && h.db.Pool != nil {
		err := h.db.Pool.Ping(ctx)
		if err != nil {
			return helpers.Error(c, "service not ready", 503)
		}
	}

	return helpers.Success(c, map[string]string{"status": "ready"}, 200)
}

// LivenessCheck checks if the service is alive
func (h *SystemHandler) LivenessCheck(c echo.Context) error {
	return helpers.Success(c, map[string]string{"status": "alive"}, 200)
}
