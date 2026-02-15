// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package smtp

import (
	"errors"
	"strings"
	"testing"

	"github.com/darkpipe/darkpipe/cloud-relay/relay/forward"
)

func TestSession_Mail(t *testing.T) {
	t.Parallel()

	mockFwd := forward.NewMockForwarder()
	session := &Session{forwarder: mockFwd}

	from := "sender@example.com"
	err := session.Mail(from, nil)
	if err != nil {
		t.Fatalf("Mail: %v", err)
	}

	if session.from != from {
		t.Errorf("session.from = %q, want %q", session.from, from)
	}
}

func TestSession_Rcpt(t *testing.T) {
	t.Parallel()

	mockFwd := forward.NewMockForwarder()
	session := &Session{forwarder: mockFwd}

	// Add first recipient
	err := session.Rcpt("recipient1@example.com", nil)
	if err != nil {
		t.Fatalf("Rcpt(1): %v", err)
	}

	if len(session.to) != 1 {
		t.Fatalf("len(session.to) = %d, want 1", len(session.to))
	}

	if session.to[0] != "recipient1@example.com" {
		t.Errorf("session.to[0] = %q, want %q", session.to[0], "recipient1@example.com")
	}

	// Add second recipient
	err = session.Rcpt("recipient2@example.com", nil)
	if err != nil {
		t.Fatalf("Rcpt(2): %v", err)
	}

	if len(session.to) != 2 {
		t.Fatalf("len(session.to) = %d, want 2", len(session.to))
	}

	if session.to[1] != "recipient2@example.com" {
		t.Errorf("session.to[1] = %q, want %q", session.to[1], "recipient2@example.com")
	}
}

func TestSession_Data(t *testing.T) {
	t.Parallel()

	mockFwd := forward.NewMockForwarder()
	session := &Session{forwarder: mockFwd}

	// Set up envelope
	session.from = "sender@example.com"
	session.to = []string{"recipient1@example.com", "recipient2@example.com"}

	// Send message data
	messageBody := "Subject: Test\r\n\r\nTest message body"
	err := session.Data(strings.NewReader(messageBody))
	if err != nil {
		t.Fatalf("Data: %v", err)
	}

	// Verify forwarder was called
	calls := mockFwd.GetCalls()
	if len(calls) != 1 {
		t.Fatalf("forwarder calls = %d, want 1", len(calls))
	}

	call := calls[0]
	if call.From != "sender@example.com" {
		t.Errorf("forwarded from = %q, want %q", call.From, "sender@example.com")
	}

	if len(call.To) != 2 {
		t.Fatalf("forwarded to count = %d, want 2", len(call.To))
	}

	if call.To[0] != "recipient1@example.com" {
		t.Errorf("forwarded to[0] = %q, want %q", call.To[0], "recipient1@example.com")
	}

	if call.To[1] != "recipient2@example.com" {
		t.Errorf("forwarded to[1] = %q, want %q", call.To[1], "recipient2@example.com")
	}

	if call.Data != messageBody {
		t.Errorf("forwarded data = %q, want %q", call.Data, messageBody)
	}
}

func TestSession_DataWithForwarderError(t *testing.T) {
	t.Parallel()

	mockFwd := forward.NewMockForwarder()
	mockFwd.ForwardError = errors.New("connection refused")

	session := &Session{forwarder: mockFwd}
	session.from = "sender@example.com"
	session.to = []string{"recipient@example.com"}

	err := session.Data(strings.NewReader("Test message"))
	if err == nil {
		t.Fatal("Data: expected error from forwarder, got nil")
	}

	// Error should be wrapped with context
	errMsg := err.Error()
	if !strings.Contains(errMsg, "forward to home device") {
		t.Errorf("error message = %q, want to contain 'forward to home device'", errMsg)
	}
}

func TestSession_Reset(t *testing.T) {
	t.Parallel()

	mockFwd := forward.NewMockForwarder()
	session := &Session{forwarder: mockFwd}

	// Set up session state
	session.from = "sender@example.com"
	session.to = []string{"recipient@example.com"}

	// Reset should clear state
	session.Reset()

	if session.from != "" {
		t.Errorf("session.from after reset = %q, want empty", session.from)
	}

	if session.to != nil {
		t.Errorf("session.to after reset = %v, want nil", session.to)
	}
}

func TestSession_Logout(t *testing.T) {
	t.Parallel()

	mockFwd := forward.NewMockForwarder()
	session := &Session{forwarder: mockFwd}

	err := session.Logout()
	if err != nil {
		t.Errorf("Logout: %v, want nil", err)
	}
}

func TestSession_FullLifecycle(t *testing.T) {
	t.Parallel()

	mockFwd := forward.NewMockForwarder()
	session := &Session{forwarder: mockFwd}

	// Simulate full SMTP transaction
	if err := session.Mail("sender@example.com", nil); err != nil {
		t.Fatalf("Mail: %v", err)
	}

	if err := session.Rcpt("recipient1@example.com", nil); err != nil {
		t.Fatalf("Rcpt(1): %v", err)
	}

	if err := session.Rcpt("recipient2@example.com", nil); err != nil {
		t.Fatalf("Rcpt(2): %v", err)
	}

	messageBody := "Subject: Full Lifecycle Test\r\n\r\nThis is a test"
	if err := session.Data(strings.NewReader(messageBody)); err != nil {
		t.Fatalf("Data: %v", err)
	}

	// Verify complete envelope was forwarded
	calls := mockFwd.GetCalls()
	if len(calls) != 1 {
		t.Fatalf("forwarder calls = %d, want 1", len(calls))
	}

	call := calls[0]
	if call.From != "sender@example.com" {
		t.Errorf("forwarded from = %q, want %q", call.From, "sender@example.com")
	}

	if len(call.To) != 2 {
		t.Errorf("forwarded to count = %d, want 2", len(call.To))
	}

	if !strings.Contains(call.Data, "Full Lifecycle Test") {
		t.Errorf("forwarded data doesn't contain subject")
	}

	// Test reset
	session.Reset()

	if session.from != "" || session.to != nil {
		t.Error("Reset did not clear session state")
	}

	// Test logout
	if err := session.Logout(); err != nil {
		t.Errorf("Logout: %v", err)
	}
}

func TestSession_MultipleMessages(t *testing.T) {
	t.Parallel()

	mockFwd := forward.NewMockForwarder()
	session := &Session{forwarder: mockFwd}

	// First message
	session.Mail("sender1@example.com", nil)
	session.Rcpt("recipient1@example.com", nil)
	session.Data(strings.NewReader("First message"))
	session.Reset()

	// Second message
	session.Mail("sender2@example.com", nil)
	session.Rcpt("recipient2@example.com", nil)
	session.Data(strings.NewReader("Second message"))

	// Verify both messages were forwarded
	calls := mockFwd.GetCalls()
	if len(calls) != 2 {
		t.Fatalf("forwarder calls = %d, want 2", len(calls))
	}

	if calls[0].From != "sender1@example.com" {
		t.Errorf("first message from = %q, want %q", calls[0].From, "sender1@example.com")
	}

	if calls[1].From != "sender2@example.com" {
		t.Errorf("second message from = %q, want %q", calls[1].From, "sender2@example.com")
	}

	if calls[0].Data != "First message" {
		t.Errorf("first message data = %q, want %q", calls[0].Data, "First message")
	}

	if calls[1].Data != "Second message" {
		t.Errorf("second message data = %q, want %q", calls[1].Data, "Second message")
	}
}
