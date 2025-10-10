package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/notification"

	"github.com/labstack/echo/v4"
)

type NotificationHandler struct {
	notificationService notification.NotificationService
}

func NewNotificationHandler(notificationService notification.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// GetNotifications retrieves notifications for the current user
func (h *NotificationHandler) GetNotifications(c echo.Context) error {
	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "user ID required", 401)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	unreadOnly := c.QueryParam("unread") == "true"

	limitStr := c.QueryParam("limit")
	limit := 50
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil {
			limit = l
		}
	}

	notifications, err := h.notificationService.GetUserNotifications(c.Request().Context(), collegeID, userID, unreadOnly, limit)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, notifications, 200)
}

// MarkAsRead marks a notification as read
func (h *NotificationHandler) MarkAsRead(c echo.Context) error {
	notificationIDStr := c.Param("notificationID")
	notificationID, err := strconv.Atoi(notificationIDStr)
	if err != nil {
		return helpers.Error(c, "invalid notification ID", 400)
	}

	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "user ID required", 401)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	err = h.notificationService.MarkAsRead(c.Request().Context(), collegeID, notificationID, userID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Notification marked as read", 200)
}

// MarkAllAsRead marks all notifications as read
func (h *NotificationHandler) MarkAllAsRead(c echo.Context) error {
	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "user ID required", 401)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	err = h.notificationService.MarkAllAsRead(c.Request().Context(), collegeID, userID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "All notifications marked as read", 200)
}

// DeleteNotification deletes a notification
func (h *NotificationHandler) DeleteNotification(c echo.Context) error {
	notificationIDStr := c.Param("notificationID")
	notificationID, err := strconv.Atoi(notificationIDStr)
	if err != nil {
		return helpers.Error(c, "invalid notification ID", 400)
	}

	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "user ID required", 401)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	err = h.notificationService.DeleteNotification(c.Request().Context(), collegeID, notificationID, userID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Notification deleted", 200)
}

// SendNotification sends a notification (Admin/Faculty only)
func (h *NotificationHandler) SendNotification(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req models.Notification
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	req.CollegeID = collegeID

	err = h.notificationService.SendNotification(c.Request().Context(), &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, "Notification sent successfully", 201)
}

// GetUnreadCount retrieves the count of unread notifications
func (h *NotificationHandler) GetUnreadCount(c echo.Context) error {
	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "user ID required", 401)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	count, err := h.notificationService.GetUnreadCount(c.Request().Context(), collegeID, userID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, map[string]int{"unread_count": count}, 200)
}
