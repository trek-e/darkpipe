package tls

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/darkpipe/darkpipe/cloud-relay/relay/notify"
)

// mockNotifier captures events for testing.
type mockNotifier struct {
	events []notify.Event
}

func (m *mockNotifier) Send(ctx context.Context, event notify.Event) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockNotifier) Close() error {
	return nil
}

func TestTLSMonitor_DetectsTLSRequiredPattern(t *testing.T) {
	logLine := "postfix/smtp[12345]: TLS is required, but was not offered by mail.example.com[1.2.3.4] for to=<user@example.com>"
	logReader := strings.NewReader(logLine)

	notifier := &mockNotifier{}
	monitor := NewTLSMonitor(logReader, notifier)

	ctx := context.Background()
	monitor.Monitor(ctx)

	if len(notifier.events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(notifier.events))
	}

	event := notifier.events[0]
	if event.Type != "tls_failure" {
		t.Errorf("Expected type 'tls_failure', got '%s'", event.Type)
	}
	if event.Domain != "example.com" {
		t.Errorf("Expected domain 'example.com', got '%s'", event.Domain)
	}
}

func TestTLSMonitor_DetectsTLSHandshakeFailure(t *testing.T) {
	logLine := "postfix/smtp[12345]: TLS handshake failed to mail.fail.com[1.2.3.4] for to=<admin@fail.com>"
	logReader := strings.NewReader(logLine)

	notifier := &mockNotifier{}
	monitor := NewTLSMonitor(logReader, notifier)

	ctx := context.Background()
	monitor.Monitor(ctx)

	if len(notifier.events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(notifier.events))
	}

	event := notifier.events[0]
	if event.Type != "tls_warning" {
		t.Errorf("Expected type 'tls_warning', got '%s'", event.Type)
	}
	if event.Domain != "fail.com" {
		t.Errorf("Expected domain 'fail.com', got '%s'", event.Domain)
	}
}

func TestTLSMonitor_ExtractsDomainFromLogLine(t *testing.T) {
	tests := []struct {
		logLine        string
		expectedDomain string
	}{
		{
			logLine:        "postfix/smtp[123]: message to=<user@example.com>",
			expectedDomain: "example.com",
		},
		{
			logLine:        "postfix/smtp[456]: connect from mail.server.org[1.2.3.4]",
			expectedDomain: "mail.server.org",
		},
		{
			logLine:        "postfix/smtp[789]: some other log without domain info",
			expectedDomain: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expectedDomain, func(t *testing.T) {
			domain := extractDomain(tt.logLine)
			if domain != tt.expectedDomain {
				t.Errorf("Expected domain '%s', got '%s'", tt.expectedDomain, domain)
			}
		})
	}
}

func TestTLSMonitor_IgnoresSuccessfulTLS(t *testing.T) {
	logLine := "postfix/smtp[12345]: Anonymous TLS connection established to mail.example.com[1.2.3.4]"
	logReader := strings.NewReader(logLine)

	notifier := &mockNotifier{}
	monitor := NewTLSMonitor(logReader, notifier)

	ctx := context.Background()
	monitor.Monitor(ctx)

	// Successful TLS should NOT trigger a notification
	if len(notifier.events) != 0 {
		t.Errorf("Expected 0 events for successful TLS, got %d", len(notifier.events))
	}
}

func TestTLSMonitor_ContextCancellation(t *testing.T) {
	// Create a long log stream that would take time to process
	var buf bytes.Buffer
	for i := 0; i < 1000; i++ {
		buf.WriteString("postfix/smtp[123]: some log line\n")
	}

	notifier := &mockNotifier{}
	monitor := NewTLSMonitor(&buf, notifier)

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	// Monitor should return without processing all lines
	err := monitor.Monitor(ctx)
	if err != nil {
		t.Errorf("Expected nil error on context cancellation, got %v", err)
	}
}

func TestTLSMonitor_DetectsCertificateVerificationFailure(t *testing.T) {
	logLine := "postfix/smtp[12345]: certificate verification failed: untrusted issuer for mail.bad.com[1.2.3.4] to=<user@bad.com>"
	logReader := strings.NewReader(logLine)

	notifier := &mockNotifier{}
	monitor := NewTLSMonitor(logReader, notifier)

	ctx := context.Background()
	monitor.Monitor(ctx)

	if len(notifier.events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(notifier.events))
	}

	event := notifier.events[0]
	if event.Type != "tls_warning" {
		t.Errorf("Expected type 'tls_warning', got '%s'", event.Type)
	}
	if !strings.Contains(event.Message, "certificate verification failed") {
		t.Errorf("Expected message about certificate verification, got '%s'", event.Message)
	}
}
