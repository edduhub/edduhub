package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"eduhub/server/internal/services/integrations"
)

// SMSService defines the interface for SMS operations
type SMSService interface {
	SendSMS(ctx context.Context, toPhoneNumber, message string) error
	SendBulkSMS(ctx context.Context, phoneNumbers []string, message string) ([]string, []string)
	SendTemplateSMS(ctx context.Context, toPhoneNumber, templateName string, variables map[string]string) error
	SendFeeReminder(ctx context.Context, toPhoneNumber, studentName string, amount float64, dueDate string) error
	SendAttendanceAlert(ctx context.Context, toPhoneNumber, studentName string, date string, status string) error
	SendGradeNotification(ctx context.Context, toPhoneNumber, studentName, courseName string, grade string) error
	SendExamReminder(ctx context.Context, toPhoneNumber, studentName, examName string, examDate string) error
	VerifyPhoneNumber(ctx context.Context, phoneNumber string) (bool, error)
	GetSMSStatus(ctx context.Context, messageSID string) (string, error)
}

// Config holds SMS service configuration
type Config struct {
	AccountSID      string
	AuthToken       string
	FromPhoneNumber string
	Enabled         bool
	BaseURL         string
}

// twilioResponse represents Twilio API response
type twilioResponse struct {
	SID          string `json:"sid"`
	Status       string `json:"status"`
	To           string `json:"to"`
	From         string `json:"from"`
	Body         string `json:"body"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// httpSMSService implements SMSService using HTTP requests to Twilio API
type httpSMSService struct {
	config Config
	client *http.Client
}

// NewSMSService creates a new SMS service instance
func NewSMSService(config Config) SMSService {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.twilio.com/2010-04-01"
	}

	return &httpSMSService{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *httpSMSService) validateConfig() error {
	if !s.config.Enabled {
		return integrations.NewDisabledError("sms")
	}

	missing := make([]string, 0)
	if strings.TrimSpace(s.config.AccountSID) == "" {
		missing = append(missing, "TWILIO_ACCOUNT_SID")
	}
	if strings.TrimSpace(s.config.AuthToken) == "" {
		missing = append(missing, "TWILIO_AUTH_TOKEN")
	}
	if strings.TrimSpace(s.config.FromPhoneNumber) == "" {
		missing = append(missing, "TWILIO_PHONE_NUMBER")
	}
	if len(missing) > 0 {
		return integrations.NewMisconfiguredError("sms", missing...)
	}

	return nil
}

// NewSMSServiceFromEnv creates SMS service from environment variables
func NewSMSServiceFromEnv() SMSService {
	config := Config{
		AccountSID:      os.Getenv("TWILIO_ACCOUNT_SID"),
		AuthToken:       os.Getenv("TWILIO_AUTH_TOKEN"),
		FromPhoneNumber: os.Getenv("TWILIO_PHONE_NUMBER"),
		Enabled:         os.Getenv("SMS_ENABLED") == "true",
		BaseURL:         "https://api.twilio.com/2010-04-01",
	}

	return NewSMSService(config)
}

// SendSMS sends a single SMS message via Twilio API
func (s *httpSMSService) SendSMS(ctx context.Context, toPhoneNumber, message string) error {
	if err := s.validateConfig(); err != nil {
		return err
	}

	apiURL := fmt.Sprintf("%s/Accounts/%s/Messages.json", s.config.BaseURL, s.config.AccountSID)

	data := url.Values{}
	data.Set("To", toPhoneNumber)
	data.Set("From", s.config.FromPhoneNumber)
	data.Set("Body", message)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.config.AccountSID, s.config.AuthToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SMS request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("SMS API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result twilioResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if result.ErrorCode != "" {
		return fmt.Errorf("SMS error %s: %s", result.ErrorCode, result.ErrorMessage)
	}

	return nil
}

// SendBulkSMS sends SMS to multiple recipients
func (s *httpSMSService) SendBulkSMS(ctx context.Context, phoneNumbers []string, message string) ([]string, []string) {
	if err := s.validateConfig(); err != nil {
		return nil, append([]string{}, phoneNumbers...)
	}

	var successList []string
	var failedList []string

	for _, phone := range phoneNumbers {
		if err := s.SendSMS(ctx, phone, message); err != nil {
			failedList = append(failedList, phone)
		} else {
			successList = append(successList, phone)
		}
	}

	return successList, failedList
}

// SendTemplateSMS sends an SMS using a predefined template
func (s *httpSMSService) SendTemplateSMS(ctx context.Context, toPhoneNumber, templateName string, variables map[string]string) error {
	templates := map[string]string{
		"fee_reminder":       "Hi {{studentName}}, this is a reminder that your fee payment of ${{amount}} is due on {{dueDate}}. Please pay to avoid late charges. -EduHub",
		"attendance_alert":   "Alert: {{studentName}} was marked {{status}} on {{date}}. Please contact the college if you have any concerns. -EduHub",
		"grade_notification": "Hi {{studentName}}, your grade for {{courseName}} has been published: {{grade}}. Check the portal for details. -EduHub",
		"exam_reminder":      "Reminder: {{examName}} is scheduled for {{examDate}}. Be prepared and arrive on time. Good luck! -EduHub",
		"welcome":            "Welcome to EduHub! Your account has been created. Login to access your dashboard.",
	}

	template, exists := templates[templateName]
	if !exists {
		return fmt.Errorf("template not found: %s", templateName)
	}

	message := template
	for key, value := range variables {
		placeholder := "{{" + key + "}}"
		message = strings.ReplaceAll(message, placeholder, value)
	}

	return s.SendSMS(ctx, toPhoneNumber, message)
}

// SendFeeReminder sends a fee reminder SMS
func (s *httpSMSService) SendFeeReminder(ctx context.Context, toPhoneNumber, studentName string, amount float64, dueDate string) error {
	return s.SendTemplateSMS(ctx, toPhoneNumber, "fee_reminder", map[string]string{
		"studentName": studentName,
		"amount":      fmt.Sprintf("%.2f", amount),
		"dueDate":     dueDate,
	})
}

// SendAttendanceAlert sends an attendance alert SMS
func (s *httpSMSService) SendAttendanceAlert(ctx context.Context, toPhoneNumber, studentName string, date string, status string) error {
	return s.SendTemplateSMS(ctx, toPhoneNumber, "attendance_alert", map[string]string{
		"studentName": studentName,
		"date":        date,
		"status":      status,
	})
}

// SendGradeNotification sends a grade notification SMS
func (s *httpSMSService) SendGradeNotification(ctx context.Context, toPhoneNumber, studentName, courseName string, grade string) error {
	return s.SendTemplateSMS(ctx, toPhoneNumber, "grade_notification", map[string]string{
		"studentName": studentName,
		"courseName":  courseName,
		"grade":       grade,
	})
}

// SendExamReminder sends an exam reminder SMS
func (s *httpSMSService) SendExamReminder(ctx context.Context, toPhoneNumber, studentName, examName string, examDate string) error {
	return s.SendTemplateSMS(ctx, toPhoneNumber, "exam_reminder", map[string]string{
		"studentName": studentName,
		"examName":    examName,
		"examDate":    examDate,
	})
}

// VerifyPhoneNumber validates a phone number format
func (s *httpSMSService) VerifyPhoneNumber(ctx context.Context, phoneNumber string) (bool, error) {
	if err := s.validateConfig(); err != nil {
		return false, err
	}
	if len(phoneNumber) < 10 {
		return false, nil
	}
	return true, nil
}

// GetSMSStatus retrieves the status of a sent SMS
func (s *httpSMSService) GetSMSStatus(ctx context.Context, messageSID string) (string, error) {
	if err := s.validateConfig(); err != nil {
		return "", err
	}

	apiURL := fmt.Sprintf("%s/Accounts/%s/Messages/%s.json", s.config.BaseURL, s.config.AccountSID, messageSID)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(s.config.AccountSID, s.config.AuthToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get SMS status: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("SMS API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result twilioResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Status, nil
}
