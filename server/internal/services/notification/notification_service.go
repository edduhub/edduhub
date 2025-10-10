package notification

import (
	"context"
	"fmt"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

type NotificationService interface {
	SendNotification(ctx context.Context, notification *models.Notification) error
	GetUserNotifications(ctx context.Context, collegeID, userID int, unreadOnly bool, limit int) ([]*models.Notification, error)
	MarkAsRead(ctx context.Context, collegeID, notificationID, userID int) error
	MarkAllAsRead(ctx context.Context, collegeID, userID int) error
	DeleteNotification(ctx context.Context, collegeID, notificationID, userID int) error
	GetUnreadCount(ctx context.Context, collegeID, userID int) (int, error)
}

type notificationService struct {
	notificationRepo repository.NotificationRepository
}

func NewNotificationService(notificationRepo repository.NotificationRepository) NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
	}
}

func (s *notificationService) SendNotification(ctx context.Context, notification *models.Notification) error {
	if notification.Title == "" {
		return fmt.Errorf("notification title is required")
	}
	if notification.UserID == 0 {
		return fmt.Errorf("user ID is required")
	}

	return s.notificationRepo.CreateNotification(ctx, notification)
}

func (s *notificationService) GetUserNotifications(ctx context.Context, collegeID, userID int, unreadOnly bool, limit int) ([]*models.Notification, error) {
	return s.notificationRepo.GetNotificationsByUser(ctx, collegeID, userID, unreadOnly, limit)
}

func (s *notificationService) MarkAsRead(ctx context.Context, collegeID, notificationID, userID int) error {
	return s.notificationRepo.MarkAsRead(ctx, collegeID, notificationID, userID)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, collegeID, userID int) error {
	return s.notificationRepo.MarkAllAsRead(ctx, collegeID, userID)
}

func (s *notificationService) DeleteNotification(ctx context.Context, collegeID, notificationID, userID int) error {
	return s.notificationRepo.DeleteNotification(ctx, collegeID, notificationID, userID)
}

func (s *notificationService) GetUnreadCount(ctx context.Context, collegeID, userID int) (int, error) {
	return s.notificationRepo.GetUnreadCount(ctx, collegeID, userID)
}
