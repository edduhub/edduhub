package push

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// PushNotificationService defines the interface for push notification operations
type PushNotificationService interface {
	SendPushNotification(ctx context.Context, deviceToken, title, body string, data map[string]string) error
	SendMulticast(ctx context.Context, deviceTokens []string, title, body string, data map[string]string) (*MulticastResponse, error)
	SendTopicNotification(ctx context.Context, topic, title, body string, data map[string]string) error
	SubscribeToTopic(ctx context.Context, deviceTokens []string, topic string) error
	UnsubscribeFromTopic(ctx context.Context, deviceTokens []string, topic string) error
	SendGradeNotification(ctx context.Context, deviceToken, studentName, courseName, grade string) error
	SendFeeReminder(ctx context.Context, deviceToken, studentName string, amount float64, dueDate string) error
	SendAttendanceAlert(ctx context.Context, deviceToken, studentName, date, status string) error
	SendExamReminder(ctx context.Context, deviceToken, examName, examDate, venue string) error
	SendAnnouncement(ctx context.Context, deviceTokens []string, title, message string, priority string) error
	ValidateDeviceToken(ctx context.Context, deviceToken string) (bool, error)
}

// Config holds push notification service configuration
type Config struct {
	ServerKey string
	ProjectID string
	Enabled   bool
	BaseURL   string
}

// FCMMessage represents a Firebase Cloud Messaging message
type FCMMessage struct {
	Message struct {
		Token        string `json:"token,omitempty"`
		Topic        string `json:"topic,omitempty"`
		Notification struct {
			Title string `json:"title"`
			Body  string `json:"body"`
		} `json:"notification,omitempty"`
		Data    map[string]string `json:"data,omitempty"`
		Android struct {
			Priority     string `json:"priority,omitempty"`
			Notification struct {
				ChannelID string `json:"channel_id,omitempty"`
				Sound     string `json:"sound,omitempty"`
			} `json:"notification,omitempty"`
		} `json:"android,omitempty"`
		APNS struct {
			Payload struct {
				APS struct {
					Sound string `json:"sound,omitempty"`
					Badge int    `json:"badge,omitempty"`
				} `json:"aps"`
			} `json:"aps"`
		} `json:"apns,omitempty"`
	} `json:"message"`
}

// FCMResponse represents FCM API response
type FCMResponse struct {
	Name string `json:"name"`
}

// FCMError represents FCM error response
type FCMError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// MulticastResponse represents response for multicast send
type MulticastResponse struct {
	SuccessCount int      `json:"success_count"`
	FailureCount int      `json:"failure_count"`
	FailedTokens []string `json:"failed_tokens"`
}

// fcmPushService implements PushNotificationService using Firebase Cloud Messaging
type fcmPushService struct {
	config Config
	client *http.Client
}

// NewPushNotificationService creates a new push notification service instance
func NewPushNotificationService(config Config) PushNotificationService {
	if !config.Enabled || config.ServerKey == "" {
		return &mockPushService{}
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://fcm.googleapis.com/v1"
	}

	return &fcmPushService{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewPushNotificationServiceFromEnv creates push service from environment variables
func NewPushNotificationServiceFromEnv() PushNotificationService {
	config := Config{
		ServerKey: os.Getenv("FCM_SERVER_KEY"),
		ProjectID: os.Getenv("FCM_PROJECT_ID"),
		Enabled:   os.Getenv("PUSH_ENABLED") == "true",
		BaseURL:   "https://fcm.googleapis.com/v1",
	}

	return NewPushNotificationService(config)
}

// SendPushNotification sends a push notification to a single device
func (s *fcmPushService) SendPushNotification(ctx context.Context, deviceToken, title, body string, data map[string]string) error {
	if data == nil {
		data = make(map[string]string)
	}

	message := FCMMessage{}
	message.Message.Token = deviceToken
	message.Message.Notification.Title = title
	message.Message.Notification.Body = body
	message.Message.Data = data
	message.Message.Android.Priority = "high"
	message.Message.Android.Notification.ChannelID = "eduhub_general"
	message.Message.Android.Notification.Sound = "default"
	message.Message.APNS.Payload.APS.Sound = "default"
	message.Message.APNS.Payload.APS.Badge = 1

	apiURL := fmt.Sprintf("%s/projects/%s/messages:send", s.config.BaseURL, s.config.ProjectID)

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.ServerKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send push notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("FCM API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// SendMulticast sends push notifications to multiple devices
func (s *fcmPushService) SendMulticast(ctx context.Context, deviceTokens []string, title, body string, data map[string]string) (*MulticastResponse, error) {
	response := &MulticastResponse{}

	for _, token := range deviceTokens {
		if err := s.SendPushNotification(ctx, token, title, body, data); err != nil {
			response.FailureCount++
			response.FailedTokens = append(response.FailedTokens, token)
		} else {
			response.SuccessCount++
		}
	}

	return response, nil
}

// SendTopicNotification sends a notification to all devices subscribed to a topic
func (s *fcmPushService) SendTopicNotification(ctx context.Context, topic, title, body string, data map[string]string) error {
	if data == nil {
		data = make(map[string]string)
	}

	message := FCMMessage{}
	message.Message.Topic = topic
	message.Message.Notification.Title = title
	message.Message.Notification.Body = body
	message.Message.Data = data
	message.Message.Android.Priority = "high"
	message.Message.Android.Notification.ChannelID = "eduhub_general"
	message.Message.Android.Notification.Sound = "default"

	apiURL := fmt.Sprintf("%s/projects/%s/messages:send", s.config.BaseURL, s.config.ProjectID)

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.ServerKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send topic notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("FCM API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// SubscribeToTopic subscribes devices to a topic
func (s *fcmPushService) SubscribeToTopic(ctx context.Context, deviceTokens []string, topic string) error {
	apiURL := fmt.Sprintf("%s/projects/%s/topics:batchAdd", s.config.BaseURL, s.config.ProjectID)

	payload := map[string]interface{}{
		"tokens": deviceTokens,
		"topic":  topic,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.ServerKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("FCM API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// UnsubscribeFromTopic unsubscribes devices from a topic
func (s *fcmPushService) UnsubscribeFromTopic(ctx context.Context, deviceTokens []string, topic string) error {
	apiURL := fmt.Sprintf("%s/projects/%s/topics:batchRemove", s.config.BaseURL, s.config.ProjectID)

	payload := map[string]interface{}{
		"tokens": deviceTokens,
		"topic":  topic,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.ServerKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe from topic: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("FCM API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// SendGradeNotification sends a grade notification push
func (s *fcmPushService) SendGradeNotification(ctx context.Context, deviceToken, studentName, courseName, grade string) error {
	data := map[string]string{
		"type":        "grade",
		"courseName":  courseName,
		"grade":       grade,
		"studentName": studentName,
	}

	return s.SendPushNotification(ctx, deviceToken,
		"Grade Published",
		fmt.Sprintf("Your grade for %s has been published: %s", courseName, grade),
		data)
}

// SendFeeReminder sends a fee reminder push notification
func (s *fcmPushService) SendFeeReminder(ctx context.Context, deviceToken, studentName string, amount float64, dueDate string) error {
	data := map[string]string{
		"type":        "fee",
		"amount":      fmt.Sprintf("%.2f", amount),
		"dueDate":     dueDate,
		"studentName": studentName,
	}

	return s.SendPushNotification(ctx, deviceToken,
		"Fee Payment Reminder",
		fmt.Sprintf("Fee payment of $%.2f is due on %s", amount, dueDate),
		data)
}

// SendAttendanceAlert sends an attendance alert push notification
func (s *fcmPushService) SendAttendanceAlert(ctx context.Context, deviceToken, studentName, date, status string) error {
	data := map[string]string{
		"type":        "attendance",
		"date":        date,
		"status":      status,
		"studentName": studentName,
	}

	return s.SendPushNotification(ctx, deviceToken,
		"Attendance Alert",
		fmt.Sprintf("%s was marked %s on %s", studentName, status, date),
		data)
}

// SendExamReminder sends an exam reminder push notification
func (s *fcmPushService) SendExamReminder(ctx context.Context, deviceToken, examName, examDate, venue string) error {
	data := map[string]string{
		"type":     "exam",
		"examName": examName,
		"examDate": examDate,
		"venue":    venue,
	}

	return s.SendPushNotification(ctx, deviceToken,
		"Exam Reminder",
		fmt.Sprintf("%s is scheduled for %s at %s", examName, examDate, venue),
		data)
}

// SendAnnouncement sends an announcement to multiple devices
func (s *fcmPushService) SendAnnouncement(ctx context.Context, deviceTokens []string, title, message string, priority string) error {
	data := map[string]string{
		"type":     "announcement",
		"priority": priority,
		"title":    title,
	}

	_, err := s.SendMulticast(ctx, deviceTokens, title, message, data)
	return err
}

// ValidateDeviceToken checks if a device token is valid
func (s *fcmPushService) ValidateDeviceToken(ctx context.Context, deviceToken string) (bool, error) {
	return len(deviceToken) > 50, nil
}

// mockPushService is a mock implementation for development/testing
type mockPushService struct{}

func (m *mockPushService) SendPushNotification(ctx context.Context, deviceToken, title, body string, data map[string]string) error {
	return nil
}

func (m *mockPushService) SendMulticast(ctx context.Context, deviceTokens []string, title, body string, data map[string]string) (*MulticastResponse, error) {
	return &MulticastResponse{SuccessCount: len(deviceTokens)}, nil
}

func (m *mockPushService) SendTopicNotification(ctx context.Context, topic, title, body string, data map[string]string) error {
	return nil
}

func (m *mockPushService) SubscribeToTopic(ctx context.Context, deviceTokens []string, topic string) error {
	return nil
}

func (m *mockPushService) UnsubscribeFromTopic(ctx context.Context, deviceTokens []string, topic string) error {
	return nil
}

func (m *mockPushService) SendGradeNotification(ctx context.Context, deviceToken, studentName, courseName, grade string) error {
	return nil
}

func (m *mockPushService) SendFeeReminder(ctx context.Context, deviceToken, studentName string, amount float64, dueDate string) error {
	return nil
}

func (m *mockPushService) SendAttendanceAlert(ctx context.Context, deviceToken, studentName, date, status string) error {
	return nil
}

func (m *mockPushService) SendExamReminder(ctx context.Context, deviceToken, examName, examDate, venue string) error {
	return nil
}

func (m *mockPushService) SendAnnouncement(ctx context.Context, deviceTokens []string, title, message string, priority string) error {
	return nil
}

func (m *mockPushService) ValidateDeviceToken(ctx context.Context, deviceToken string) (bool, error) {
	return true, nil
}
