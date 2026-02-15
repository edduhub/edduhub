package sms

import (
	"context"
	"testing"

	"eduhub/server/internal/services/integrations"
)

func TestSMSServiceDisabled(t *testing.T) {
	svc := NewSMSService(Config{Enabled: false})

	err := svc.SendSMS(context.Background(), "+15551234567", "hello")
	if err == nil {
		t.Fatalf("expected error when SMS integration is disabled")
	}
	if !integrations.IsDisabled(err) {
		t.Fatalf("expected disabled integration error, got %v", err)
	}
}

func TestSMSServiceMisconfigured(t *testing.T) {
	svc := NewSMSService(Config{Enabled: true, AccountSID: "", AuthToken: "", FromPhoneNumber: ""})

	err := svc.SendSMS(context.Background(), "+15551234567", "hello")
	if err == nil {
		t.Fatalf("expected error when SMS integration config is incomplete")
	}
	if !integrations.IsMisconfigured(err) {
		t.Fatalf("expected misconfigured integration error, got %v", err)
	}
}

func TestSMSBulkSendReturnsFailedWhenDisabled(t *testing.T) {
	svc := NewSMSService(Config{Enabled: false})
	success, failed := svc.SendBulkSMS(context.Background(), []string{"+15550000001", "+15550000002"}, "hello")

	if len(success) != 0 {
		t.Fatalf("expected zero successful sends, got %d", len(success))
	}
	if len(failed) != 2 {
		t.Fatalf("expected all recipients to fail, got %d", len(failed))
	}
}
