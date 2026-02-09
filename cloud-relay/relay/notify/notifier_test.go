package notify

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockNotifier is a test notifier that records calls and can simulate errors.
type mockNotifier struct {
	sendCalls []Event
	sendErr   error
	closed    bool
}

func (m *mockNotifier) Send(ctx context.Context, event Event) error {
	m.sendCalls = append(m.sendCalls, event)
	return m.sendErr
}

func (m *mockNotifier) Close() error {
	m.closed = true
	return nil
}

func TestMultiNotifier_DispatchesToAllBackends(t *testing.T) {
	mock1 := &mockNotifier{}
	mock2 := &mockNotifier{}
	multi := NewMultiNotifier(mock1, mock2)

	event := Event{
		Type:      "tls_warning",
		Domain:    "example.com",
		Message:   "Test event",
		Timestamp: time.Now(),
	}

	err := multi.Send(context.Background(), event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(mock1.sendCalls) != 1 {
		t.Errorf("Expected 1 call to mock1, got %d", len(mock1.sendCalls))
	}
	if len(mock2.sendCalls) != 1 {
		t.Errorf("Expected 1 call to mock2, got %d", len(mock2.sendCalls))
	}
}

func TestMultiNotifier_CollectsErrors(t *testing.T) {
	mock1 := &mockNotifier{sendErr: errors.New("backend1 failed")}
	mock2 := &mockNotifier{}
	multi := NewMultiNotifier(mock1, mock2)

	event := Event{
		Type:   "tls_failure",
		Domain: "fail.com",
	}

	err := multi.Send(context.Background(), event)
	if err == nil {
		t.Fatal("Expected error from failing backend")
	}

	// Both backends should be called despite the error
	if len(mock1.sendCalls) != 1 {
		t.Errorf("Expected 1 call to mock1 despite error, got %d", len(mock1.sendCalls))
	}
	if len(mock2.sendCalls) != 1 {
		t.Errorf("Expected 1 call to mock2, got %d", len(mock2.sendCalls))
	}
}

func TestMultiNotifier_Close(t *testing.T) {
	mock1 := &mockNotifier{}
	mock2 := &mockNotifier{}
	multi := NewMultiNotifier(mock1, mock2)

	err := multi.Close()
	if err != nil {
		t.Fatalf("Expected no error on close, got %v", err)
	}

	if !mock1.closed {
		t.Error("Expected mock1 to be closed")
	}
	if !mock2.closed {
		t.Error("Expected mock2 to be closed")
	}
}

func TestWebhookNotifier_RateLimiting(t *testing.T) {
	webhook := NewWebhookNotifier("http://localhost:9999/webhook")
	webhook.dedupeWindow = 100 * time.Millisecond // Short window for testing

	event := Event{
		Type:      "tls_warning",
		Domain:    "spam.example.com",
		Message:   "Repeated failure",
		Timestamp: time.Now(),
	}

	// First send should succeed (or fail with connection error, but that's OK for this test)
	ctx := context.Background()
	webhook.Send(ctx, event)

	// Check that domain was recorded
	lastTime, ok := webhook.lastNotified.Load(event.Domain)
	if !ok {
		t.Fatal("Expected domain to be recorded in rate limit map")
	}

	// Immediate second send should be suppressed (no error, just silently dropped)
	err := webhook.Send(ctx, event)
	if err != nil {
		// We expect no error even if the notification is suppressed
		t.Fatalf("Expected no error from rate-limited send, got %v", err)
	}

	// Wait for dedupe window to expire
	time.Sleep(150 * time.Millisecond)

	// Third send should be allowed
	webhook.Send(ctx, event)

	newLastTime, _ := webhook.lastNotified.Load(event.Domain)
	if newLastTime == lastTime {
		t.Error("Expected last notification time to be updated after dedupe window expired")
	}
}
