package notification

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"eduhub/server/internal/models"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockNotificationRepository struct {
	mock.Mock
}

func (m *mockNotificationRepository) CreateNotification(ctx context.Context, notification *models.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *mockNotificationRepository) GetNotificationsByUser(ctx context.Context, collegeID, userID int, unreadOnly bool, limit int) ([]*models.Notification, error) {
	args := m.Called(ctx, collegeID, userID, unreadOnly, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Notification), args.Error(1)
}

func (m *mockNotificationRepository) MarkAsRead(ctx context.Context, collegeID, notificationID, userID int) error {
	args := m.Called(ctx, collegeID, notificationID, userID)
	return args.Error(0)
}

func (m *mockNotificationRepository) MarkAllAsRead(ctx context.Context, collegeID, userID int) error {
	args := m.Called(ctx, collegeID, userID)
	return args.Error(0)
}

func (m *mockNotificationRepository) DeleteNotification(ctx context.Context, collegeID, notificationID, userID int) error {
	args := m.Called(ctx, collegeID, notificationID, userID)
	return args.Error(0)
}

func (m *mockNotificationRepository) GetUnreadCount(ctx context.Context, collegeID, userID int) (int, error) {
	args := m.Called(ctx, collegeID, userID)
	return args.Int(0), args.Error(1)
}

func TestWebSocketService_New(t *testing.T) {
	mockRepo := new(mockNotificationRepository)

	service := NewWebSocketService(mockRepo, []string{"http://localhost:3000"})

	assert.NotNil(t, service)
}

func TestWebSocketService_GetConnectedUsers_Empty(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	service := NewWebSocketService(mockRepo, []string{"http://localhost:3000"})

	users := service.GetConnectedUsers(1)

	assert.Nil(t, users)
}

func TestWebSocketService_BroadcastToUser_NoConnectedUsers(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	service := NewWebSocketService(mockRepo, []string{"http://localhost:3000"})

	ctx := context.Background()
	notification := &models.Notification{Title: "Test"}

	err := service.BroadcastToUser(ctx, 1, 1, notification)

	assert.NoError(t, err)
}

func TestWebSocketService_BroadcastToUsers_NoConnectedUsers(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	service := NewWebSocketService(mockRepo, []string{"http://localhost:3000"})

	ctx := context.Background()
	notification := &models.Notification{Title: "Test"}

	err := service.BroadcastToUsers(ctx, 1, []int{1, 2, 3}, notification)

	assert.NoError(t, err)
}

func TestWebSocketService_BroadcastTypingIndicator_NoConnectedUsers(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	service := NewWebSocketService(mockRepo, []string{"http://localhost:3000"})

	ctx := context.Background()

	err := service.BroadcastTypingIndicator(ctx, 1, 1, true)

	assert.NoError(t, err)
}

func TestWebSocketService_BroadcastPresence_NoConnectedUsers(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	service := NewWebSocketService(mockRepo, []string{"http://localhost:3000"})

	ctx := context.Background()

	err := service.BroadcastPresence(ctx, 1, 1, "online")

	assert.NoError(t, err)
}

func TestWebSocketService_GetConnectionStats(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	service := NewWebSocketService(mockRepo, []string{"http://localhost:3000"})

	assert.NotPanics(t, func() {
		stats := service.GetConnectionStats()
		assert.NotNil(t, stats)
		assert.Contains(t, stats, "total_connections")
		assert.Contains(t, stats, "colleges")
		assert.Equal(t, 0, stats["total_connections"])
	})
}

func TestWebSocketService_StopIsSafe(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	service := NewWebSocketService(mockRepo, []string{"http://localhost:3000"})

	assert.NotPanics(t, func() {
		service.Stop()
		service.Stop()
	})
}

func TestExtractCollegeIDFromContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/api/notifications/ws", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("college_id", 42)

	collegeID, err := extractCollegeIDFromContext(c)

	assert.NoError(t, err)
	assert.Equal(t, 42, collegeID)
}

func TestExtractCollegeIDFromContext_Missing(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/api/notifications/ws", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	collegeID, err := extractCollegeIDFromContext(c)

	assert.Error(t, err)
	assert.Zero(t, collegeID)
}

func TestNewWebSocketMessage_UsesEnvelopeFields(t *testing.T) {
	nowBefore := time.Now()
	notification := &models.Notification{ID: 7, Title: "Demo"}

	message := newWebSocketMessage("notification", 9, 13, notification, map[string]string{"state": "ok"})

	assert.Equal(t, "notification", message.Type)
	assert.Equal(t, 9, message.CollegeID)
	assert.Equal(t, 13, message.UserID)
	assert.Equal(t, notification, message.Notification)
	assert.Equal(t, map[string]string{"state": "ok"}, message.Data)
	assert.WithinDuration(t, nowBefore, message.Timestamp, time.Second)
}
