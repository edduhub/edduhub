package repository

import (
	"context"
	"time"

	"eduhub/server/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
)

type NotificationRepository interface {
	CreateNotification(ctx context.Context, notification *models.Notification) error
	GetNotificationsByUser(ctx context.Context, collegeID, userID int, unreadOnly bool, limit int) ([]*models.Notification, error)
	MarkAsRead(ctx context.Context, collegeID, notificationID, userID int) error
	MarkAllAsRead(ctx context.Context, collegeID, userID int) error
	DeleteNotification(ctx context.Context, collegeID, notificationID, userID int) error
	GetUnreadCount(ctx context.Context, collegeID, userID int) (int, error)
}

type notificationRepository struct {
	DB *DB
}

func NewNotificationRepository(db *DB) NotificationRepository {
	return &notificationRepository{DB: db}
}

func (r *notificationRepository) CreateNotification(ctx context.Context, notification *models.Notification) error {
	now := time.Now()
	notification.CreatedAt = now

	sql := `INSERT INTO notifications (user_id, college_id, title, message, type, is_read, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	var id int
	err := r.DB.Pool.QueryRow(ctx, sql,
		notification.UserID,
		notification.CollegeID,
		notification.Title,
		notification.Message,
		notification.Type,
		false,
		notification.CreatedAt,
	).Scan(&id)

	if err != nil {
		return err
	}

	notification.ID = id
	return nil
}

func (r *notificationRepository) GetNotificationsByUser(ctx context.Context, collegeID, userID int, unreadOnly bool, limit int) ([]*models.Notification, error) {
	var sql string
	var args []interface{}

	if unreadOnly {
		sql = `SELECT id, user_id, college_id, title, message, type, is_read, created_at
			   FROM notifications WHERE college_id = $1 AND user_id = $2 AND is_read = false
			   ORDER BY created_at DESC LIMIT $3`
		args = []interface{}{collegeID, userID, limit}
	} else {
		sql = `SELECT id, user_id, college_id, title, message, type, is_read, created_at
			   FROM notifications WHERE college_id = $1 AND user_id = $2
			   ORDER BY created_at DESC LIMIT $3`
		args = []interface{}{collegeID, userID, limit}
	}

	var notifications []*models.Notification
	err := pgxscan.Select(ctx, r.DB.Pool, &notifications, sql, args...)
	return notifications, err
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, collegeID, notificationID, userID int) error {
	sql := `UPDATE notifications SET is_read = true WHERE id = $1 AND college_id = $2 AND user_id = $3`
	_, err := r.DB.Pool.Exec(ctx, sql, notificationID, collegeID, userID)
	return err
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, collegeID, userID int) error {
	sql := `UPDATE notifications SET is_read = true WHERE college_id = $1 AND user_id = $2 AND is_read = false`
	_, err := r.DB.Pool.Exec(ctx, sql, collegeID, userID)
	return err
}

func (r *notificationRepository) DeleteNotification(ctx context.Context, collegeID, notificationID, userID int) error {
	sql := `DELETE FROM notifications WHERE id = $1 AND college_id = $2 AND user_id = $3`
	_, err := r.DB.Pool.Exec(ctx, sql, notificationID, collegeID, userID)
	return err
}

func (r *notificationRepository) GetUnreadCount(ctx context.Context, collegeID, userID int) (int, error) {
	sql := `SELECT COUNT(*) FROM notifications WHERE college_id = $1 AND user_id = $2 AND is_read = false`
	
	var count int
	err := r.DB.Pool.QueryRow(ctx, sql, collegeID, userID).Scan(&count)
	return count, err
}
