package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/services/audit"

	"github.com/labstack/echo/v4"
)

type AuditHandler struct {
	auditService audit.AuditService
}

func NewAuditHandler(auditService audit.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// GetAuditLogs retrieves audit logs with filtering
func (h *AuditHandler) GetAuditLogs(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// Parse filters
	userIDStr := c.QueryParam("user_id")
	var userID *int
	if userIDStr != "" {
		uid, err := strconv.Atoi(userIDStr)
		if err == nil {
			userID = &uid
		}
	}

	action := c.QueryParam("action")
	entity := c.QueryParam("entity")
	
	limitStr := c.QueryParam("limit")
	limit := 100
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil {
			limit = l
		}
	}

	offsetStr := c.QueryParam("offset")
	offset := 0
	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err == nil {
			offset = o
		}
	}

	logs, err := h.auditService.GetAuditLogs(c.Request().Context(), collegeID, userID, action, entity, limit, offset)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, logs, 200)
}

// GetUserActivity retrieves activity logs for a specific user
func (h *AuditHandler) GetUserActivity(c echo.Context) error {
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return helpers.Error(c, "invalid user ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	limitStr := c.QueryParam("limit")
	limit := 50
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil {
			limit = l
		}
	}

	logs, err := h.auditService.GetUserActivity(c.Request().Context(), collegeID, userID, limit)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, logs, 200)
}

// GetEntityHistory retrieves history for a specific entity
func (h *AuditHandler) GetEntityHistory(c echo.Context) error {
	entityType := c.Param("entityType")
	entityIDStr := c.Param("entityID")
	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		return helpers.Error(c, "invalid entity ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	logs, err := h.auditService.GetEntityHistory(c.Request().Context(), collegeID, entityType, entityID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, logs, 200)
}

// GetAuditStats retrieves audit statistics
func (h *AuditHandler) GetAuditStats(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	stats, err := h.auditService.GetAuditStats(c.Request().Context(), collegeID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, stats, 200)
}
