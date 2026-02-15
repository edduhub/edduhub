package push

import (
	"context"
	"testing"

	"eduhub/server/internal/services/integrations"
)

func TestPushServiceDisabled(t *testing.T) {
	svc := NewPushNotificationService(Config{Enabled: false})

	err := svc.SendPushNotification(context.Background(), "token", "title", "body", nil)
	if err == nil {
		t.Fatalf("expected error when push integration is disabled")
	}
	if !integrations.IsDisabled(err) {
		t.Fatalf("expected disabled integration error, got %v", err)
	}
}

func TestPushServiceMisconfigured(t *testing.T) {
	svc := NewPushNotificationService(Config{Enabled: true, ServerKey: "", ProjectID: ""})

	err := svc.SendTopicNotification(context.Background(), "topic", "title", "body", nil)
	if err == nil {
		t.Fatalf("expected error when push integration config is incomplete")
	}
	if !integrations.IsMisconfigured(err) {
		t.Fatalf("expected misconfigured integration error, got %v", err)
	}
}

func TestValidateDeviceTokenFailsFastWhenPushDisabled(t *testing.T) {
	svc := NewPushNotificationService(Config{Enabled: false})

	ok, err := svc.ValidateDeviceToken(context.Background(), "token")
	if err == nil {
		t.Fatalf("expected error when push integration is disabled")
	}
	if ok {
		t.Fatalf("expected token validation to return false when disabled")
	}
}
