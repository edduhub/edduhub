package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
)

type EmailService interface {
	SendEmail(ctx context.Context, to, subject, body string) error
	SendTemplateEmail(ctx context.Context, to, subject, templateName string, data any) error
	SendBulkEmail(ctx context.Context, recipients []string, subject, body string) error
	SendWelcomeEmail(ctx context.Context, to, name string) error
	SendPasswordResetEmail(ctx context.Context, to, resetLink string) error
	SendGradeNotification(ctx context.Context, to, studentName, courseName string, grade float64) error
	SendAnnouncementEmail(ctx context.Context, recipients []string, announcement string) error
}

type emailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromAddress  string
	templates    map[string]*template.Template
}

func NewEmailService(host, port, username, password, fromAddress string) EmailService {
	service := &emailService{
		smtpHost:     host,
		smtpPort:     port,
		smtpUsername: username,
		smtpPassword: password,
		fromAddress:  fromAddress,
		templates:    make(map[string]*template.Template),
	}

	// Load email templates
	service.loadTemplates()

	return service
}

func (s *emailService) loadTemplates() {
	// Welcome email template
	welcomeTmpl := `
	<html>
		<body>
			<h2>Welcome to EduHub, {{.Name}}!</h2>
			<p>Your account has been successfully created.</p>
			<p>You can now log in and access all features.</p>
		</body>
	</html>
	`
	s.templates["welcome"], _ = template.New("welcome").Parse(welcomeTmpl)

	// Password reset template
	resetTmpl := `
	<html>
		<body>
			<h2>Password Reset Request</h2>
			<p>Click the link below to reset your password:</p>
			<p><a href="{{.ResetLink}}">Reset Password</a></p>
			<p>This link will expire in 24 hours.</p>
		</body>
	</html>
	`
	s.templates["reset"], _ = template.New("reset").Parse(resetTmpl)

	// Grade notification template
	gradeTmpl := `
	<html>
		<body>
			<h2>New Grade Posted</h2>
			<p>Hello {{.StudentName}},</p>
			<p>A new grade has been posted for <strong>{{.CourseName}}</strong>.</p>
			<p>Your grade: <strong>{{.Grade}}</strong></p>
		</body>
	</html>
	`
	s.templates["grade"], _ = template.New("grade").Parse(gradeTmpl)
}

func (s *emailService) SendEmail(ctx context.Context, to, subject, body string) error {
	if s.smtpHost == "" {
		// Email not configured, return error instead of failing silently
		return fmt.Errorf("SMTP not configured: cannot send email to %s", to)
	}

	// Compose email
	msg := fmt.Appendf(nil, "From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", s.fromAddress, to, subject, body)

	// SMTP authentication
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	// Send email
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	err := smtp.SendMail(addr, auth, s.fromAddress, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *emailService) SendTemplateEmail(ctx context.Context, to, subject, templateName string, data any) error {
	tmpl, ok := s.templates[templateName]
	if !ok {
		return fmt.Errorf("template %s not found", templateName)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return err
	}

	return s.SendEmail(ctx, to, subject, body.String())
}

func (s *emailService) SendBulkEmail(ctx context.Context, recipients []string, subject, body string) error {
	if len(recipients) == 0 {
		return nil
	}

	var firstError error
	for _, recipient := range recipients {
		if err := s.SendEmail(ctx, recipient, subject, body); err != nil {
			if firstError == nil {
				firstError = err
			}
		}
	}

	return firstError
}

func (s *emailService) SendWelcomeEmail(ctx context.Context, to, name string) error {
	data := map[string]string{"Name": name}
	return s.SendTemplateEmail(ctx, to, "Welcome to EduHub", "welcome", data)
}

func (s *emailService) SendPasswordResetEmail(ctx context.Context, to, resetLink string) error {
	data := map[string]string{"ResetLink": resetLink}
	return s.SendTemplateEmail(ctx, to, "Password Reset Request", "reset", data)
}

func (s *emailService) SendGradeNotification(ctx context.Context, to, studentName, courseName string, grade float64) error {
	data := map[string]any{
		"StudentName": studentName,
		"CourseName":  courseName,
		"Grade":       grade,
	}
	return s.SendTemplateEmail(ctx, to, "New Grade Posted", "grade", data)
}

func (s *emailService) SendAnnouncementEmail(ctx context.Context, recipients []string, announcement string) error {
	body := fmt.Sprintf("<html><body><h2>New Announcement</h2><p>%s</p></body></html>", announcement)
	return s.SendBulkEmail(ctx, recipients, "New Announcement", body)
}
