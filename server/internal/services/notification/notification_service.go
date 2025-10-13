package notification

import (
	"context"
	"fmt"
	"log"

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
	BroadcastNotification(ctx context.Context, collegeID int, notification *models.Notification) error
	BroadcastToUser(ctx context.Context, collegeID, userID int, notification *models.Notification) error
	BroadcastToUsers(ctx context.Context, collegeID int, userIDs []int, notification *models.Notification) error
}

type notificationService struct {
	notificationRepo repository.NotificationRepository
	websocketService WebSocketService
}

func NewNotificationService(notificationRepo repository.NotificationRepository, websocketService WebSocketService) NotificationService {
	return &notificationService{
		notificationRepo:  notificationRepo,
		websocketService: websocketService,
	}
}

func (s *notificationService) SendNotification(ctx context.Context, notification *models.Notification) error {
	if notification.Title == "" {
		return fmt.Errorf("notification title is required")
	}
	if notification.UserID == 0 {
		return fmt.Errorf("user ID is required")
	}

	// Create notification in database
	if err := s.notificationRepo.CreateNotification(ctx, notification); err != nil {
		return err
	}

	// Broadcast to WebSocket clients
	go func() {
		if err := s.websocketService.BroadcastToUser(ctx, notification.CollegeID, notification.UserID, notification); err != nil {
			log.Printf("Failed to broadcast notification: %v", err)
		}
	}()

	return nil
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

func (s *notificationService) BroadcastNotification(ctx context.Context, collegeID int, notification *models.Notification) error {
	return s.websocketService.BroadcastNotification(ctx, collegeID, notification)
}

func (s *notificationService) BroadcastToUser(ctx context.Context, collegeID, userID int, notification *models.Notification) error {
	return s.websocketService.BroadcastToUser(ctx, collegeID, userID, notification)
}

func (s *notificationService) BroadcastToUsers(ctx context.Context, collegeID int, userIDs []int, notification *models.Notification) error {
	return s.websocketService.BroadcastToUsers(ctx, collegeID, userIDs, notification)
}
