// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package tests provides end-to-end integration tests for the cloud relay.
//
// These tests verify the complete SMTP pipeline: SMTP client -> go-smtp server
// -> session -> forwarder -> mock home device.
package tests

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/smtp"
	"strings"
	"testing"
	"time"

	"github.com/darkpipe/darkpipe/cloud-relay/relay/config"
	"github.com/darkpipe/darkpipe/cloud-relay/relay/forward"
	"github.com/darkpipe/darkpipe/cloud-relay/relay/queue"
	smtpserver "github.com/darkpipe/darkpipe/cloud-relay/relay/smtp"
)

// Helper function to start a test SMTP server
func startTestServer(t *testing.T, fwd forward.Forwarder) string {
	t.Helper()

	cfg := &config.Config{
		ListenAddr:      "127.0.0.1:0",
		MaxMessageBytes: 10 * 1024 * 1024,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
	}

	server := smtpserver.NewServer(fwd, cfg)

	// Create listener to get actual port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	addr := listener.Addr().String()

	// Start server in background
	go func() {
		if err := server.Serve(listener); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(50 * time.Millisecond)

	// Clean up server on test completion
	t.Cleanup(func() {
		server.Close()
		listener.Close()
	})

	return addr
}

func TestIntegration_SMTPRelayFlow(t *testing.T) {
	t.Parallel()

	// Create mock forwarder to capture forwarded messages
	mockFwd := forward.NewMockForwarder()

	// Start test server
	serverAddr := startTestServer(t, mockFwd)

	// Send a test email via SMTP client
	from := "sender@example.com"
	to := []string{"recipient@example.com"}
	message := []byte("Subject: Integration Test\r\n\r\nThis is a test message from the integration test.\r\n")

	err := smtp.SendMail(serverAddr, nil, from, to, message)
	if err != nil {
		t.Fatalf("SendMail: %v", err)
	}

	// Give forwarder time to process
	time.Sleep(100 * time.Millisecond)

	// Verify mock forwarder received the message
	calls := mockFwd.GetCalls()
	if len(calls) != 1 {
		t.Fatalf("forwarder calls = %d, want 1", len(calls))
	}

	call := calls[0]

	// Verify envelope
	if call.From != from {
		t.Errorf("forwarded from = %q, want %q", call.From, from)
	}

	if len(call.To) != 1 {
		t.Fatalf("forwarded to count = %d, want 1", len(call.To))
	}

	if call.To[0] != to[0] {
		t.Errorf("forwarded to = %q, want %q", call.To[0], to[0])
	}

	// Verify message body
	if !strings.Contains(call.Data, "Integration Test") {
		t.Errorf("message data doesn't contain subject: %s", call.Data)
	}

	if !strings.Contains(call.Data, "test message from the integration test") {
		t.Errorf("message data doesn't contain body: %s", call.Data)
	}
}

func TestIntegration_MultipleRecipients(t *testing.T) {
	t.Parallel()

	// Create mock forwarder
	mockFwd := forward.NewMockForwarder()

	// Start test server
	serverAddr := startTestServer(t, mockFwd)

	// Send email with multiple recipients
	from := "sender@example.com"
	to := []string{
		"recipient1@example.com",
		"recipient2@example.com",
		"recipient3@example.com",
	}
	message := []byte("Subject: Multiple Recipients\r\n\r\nTest with multiple recipients.\r\n")

	err := smtp.SendMail(serverAddr, nil, from, to, message)
	if err != nil {
		t.Fatalf("SendMail: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify all recipients were forwarded
	calls := mockFwd.GetCalls()
	if len(calls) != 1 {
		t.Fatalf("forwarder calls = %d, want 1", len(calls))
	}

	call := calls[0]

	if len(call.To) != 3 {
		t.Fatalf("forwarded to count = %d, want 3", len(call.To))
	}

	for i, expectedTo := range to {
		if call.To[i] != expectedTo {
			t.Errorf("forwarded to[%d] = %q, want %q", i, call.To[i], expectedTo)
		}
	}
}

func TestIntegration_LargeMessage(t *testing.T) {
	t.Parallel()

	// Create mock forwarder
	mockFwd := forward.NewMockForwarder()

	// Start test server
	serverAddr := startTestServer(t, mockFwd)

	// Create a moderately sized message with proper SMTP formatting
	// Note: SMTP has line length limits (typically 1000 chars), so we need to format properly
	bodyLines := make([]string, 100)
	for i := range bodyLines {
		bodyLines[i] = strings.Repeat("X", 70) // 70 chars per line, well under SMTP limit
	}
	largeBody := strings.Join(bodyLines, "\r\n")
	message := []byte("Subject: Large Message\r\n\r\n" + largeBody + "\r\n")

	from := "sender@example.com"
	to := []string{"recipient@example.com"}

	err := smtp.SendMail(serverAddr, nil, from, to, message)
	if err != nil {
		t.Fatalf("SendMail: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	// Verify message was forwarded
	calls := mockFwd.GetCalls()
	if len(calls) != 1 {
		t.Fatalf("forwarder calls = %d, want 1", len(calls))
	}

	call := calls[0]

	// Verify large body was received (check for presence of X characters)
	if !strings.Contains(call.Data, strings.Repeat("X", 70)) {
		t.Errorf("large message body not fully received (got %d bytes)", len(call.Data))
	}

	// Verify reasonable size
	if len(call.Data) < 7000 { // 100 lines * 70 chars + headers and CRLF
		t.Errorf("message body too small: got %d bytes, want at least 7000", len(call.Data))
	}
}

func TestIntegration_EphemeralBehavior(t *testing.T) {
	t.Parallel()

	// This test verifies that after forwarding, no data remains in the session
	// (ephemeral behavior per RELAY-02).

	// Create mock forwarder
	mockFwd := forward.NewMockForwarder()

	// Start test server
	serverAddr := startTestServer(t, mockFwd)

	// Send first message
	from1 := "sender1@example.com"
	to1 := []string{"recipient1@example.com"}
	message1 := []byte("Subject: First\r\n\r\nFirst message\r\n")

	err := smtp.SendMail(serverAddr, nil, from1, to1, message1)
	if err != nil {
		t.Fatalf("SendMail(1): %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Send second message
	from2 := "sender2@example.com"
	to2 := []string{"recipient2@example.com"}
	message2 := []byte("Subject: Second\r\n\r\nSecond message\r\n")

	err = smtp.SendMail(serverAddr, nil, from2, to2, message2)
	if err != nil {
		t.Fatalf("SendMail(2): %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify both messages were forwarded correctly (no cross-contamination)
	calls := mockFwd.GetCalls()
	if len(calls) != 2 {
		t.Fatalf("forwarder calls = %d, want 2", len(calls))
	}

	// First message should contain only first sender/recipient/body
	if calls[0].From != from1 {
		t.Errorf("first call from = %q, want %q", calls[0].From, from1)
	}

	if calls[0].To[0] != to1[0] {
		t.Errorf("first call to = %q, want %q", calls[0].To[0], to1[0])
	}

	if strings.Contains(calls[0].Data, "Second message") {
		t.Error("first call data contains second message (not ephemeral)")
	}

	// Second message should contain only second sender/recipient/body
	if calls[1].From != from2 {
		t.Errorf("second call from = %q, want %q", calls[1].From, from2)
	}

	if calls[1].To[0] != to2[0] {
		t.Errorf("second call to = %q, want %q", calls[1].To[0], to2[0])
	}

	if strings.Contains(calls[1].Data, "First message") {
		t.Error("second call data contains first message (not ephemeral)")
	}
}

// Phase 5 integration tests - Queue & Offline Handling

func TestIntegration_QueueOnOffline(t *testing.T) {
	// Success Criterion 1: With queuing enabled and home device offline,
	// inbound mail queues encrypted and delivers when reconnected.
	t.Parallel()

	// Create temp directory for queue keys
	tmpDir := t.TempDir()
	keyPath := tmpDir + "/identity"

	// Create MockForwarder configured to fail (offline)
	mockFwd := forward.NewMockForwarder()
	mockFwd.ForwardError = errors.New("home device offline")

	// Create message queue
	queueCfg := queue.QueueConfig{
		KeyPath:     keyPath,
		MaxRAMBytes: 1024 * 1024, // 1MB
		MaxMessages: 10,
		TTLHours:    1,
	}
	msgQueue, err := queue.NewMessageQueue(queueCfg)
	if err != nil {
		t.Fatalf("NewMessageQueue: %v", err)
	}
	defer msgQueue.Close()

	// Create QueuedForwarder with queue enabled
	queuedFwd := forward.NewQueuedForwarder(mockFwd, msgQueue, true)

	// Send message (should be queued since transport fails)
	from := "sender@example.com"
	to := []string{"recipient@example.com"}
	message := []byte("Message-ID: <test-queue-1@example.com>\r\nSubject: Queue Test\r\n\r\nTest message\r\n")

	err = queuedFwd.Forward(context.Background(), from, to, bytes.NewReader(message))
	if err != nil {
		t.Fatalf("Forward (should queue): %v", err)
	}

	// Verify MockForwarder was called and failed
	calls := mockFwd.GetCalls()
	if len(calls) != 1 {
		t.Fatalf("forwarder calls = %d, want 1 (failed attempt)", len(calls))
	}

	// Verify message is queued
	if msgQueue.Len() != 1 {
		t.Fatalf("queue length = %d, want 1", msgQueue.Len())
	}

	// Simulate reconnect: configure MockForwarder to succeed
	mockFwd.Reset()
	mockFwd.ForwardError = nil

	// Create a new forwarder for delivery simulation
	successFwd := forward.NewMockForwarder()

	// Process queue manually with success forwarder
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go msgQueue.StartProcessor(ctx, successFwd, 100*time.Millisecond)

	// Give processor time to deliver
	time.Sleep(500 * time.Millisecond)

	// Verify queue is empty (message delivered)
	if msgQueue.Len() != 0 {
		t.Errorf("queue length after reconnect = %d, want 0", msgQueue.Len())
	}

	// Verify success forwarder received the message
	deliveryCalls := successFwd.GetCalls()
	if len(deliveryCalls) != 1 {
		t.Fatalf("delivery calls = %d, want 1", len(deliveryCalls))
	}

	call := deliveryCalls[0]
	if call.From != from {
		t.Errorf("delivered from = %q, want %q", call.From, from)
	}
	if len(call.To) != 1 || call.To[0] != to[0] {
		t.Errorf("delivered to = %v, want %v", call.To, to)
	}
	if !strings.Contains(call.Data, "Queue Test") {
		t.Errorf("delivered data missing subject: %s", call.Data)
	}
}

func TestIntegration_QueueDisabledBounce(t *testing.T) {
	// Success Criterion 3: With queuing disabled, relay returns error
	// when home device unreachable.
	t.Parallel()

	// Create temp directory for queue keys
	tmpDir := t.TempDir()
	keyPath := tmpDir + "/identity"

	// Create MockForwarder configured to fail (offline)
	mockFwd := forward.NewMockForwarder()
	mockFwd.ForwardError = errors.New("home device offline")

	// Create message queue
	queueCfg := queue.QueueConfig{
		KeyPath:     keyPath,
		MaxRAMBytes: 1024 * 1024,
		MaxMessages: 10,
		TTLHours:    1,
	}
	msgQueue, err := queue.NewMessageQueue(queueCfg)
	if err != nil {
		t.Fatalf("NewMessageQueue: %v", err)
	}
	defer msgQueue.Close()

	// Create QueuedForwarder with queue DISABLED
	queuedFwd := forward.NewQueuedForwarder(mockFwd, msgQueue, false)

	// Send message (should return error since queue disabled)
	from := "sender@example.com"
	to := []string{"recipient@example.com"}
	message := []byte("Subject: Queue Disabled Test\r\n\r\nTest message\r\n")

	err = queuedFwd.Forward(context.Background(), from, to, bytes.NewReader(message))
	if err == nil {
		t.Fatal("Forward: expected error when queue disabled and transport fails, got nil")
	}

	// Verify error indicates offline
	if !strings.Contains(err.Error(), "offline") {
		t.Errorf("error = %q, want error containing 'offline'", err.Error())
	}

	// Verify nothing was queued
	if msgQueue.Len() != 0 {
		t.Errorf("queue length = %d, want 0 (nothing queued)", msgQueue.Len())
	}
}

func TestIntegration_QueueEncryption(t *testing.T) {
	// QUEUE-01 requirement: Queued messages are encrypted at rest.
	t.Parallel()

	// Create temp directory for queue keys
	tmpDir := t.TempDir()
	keyPath := tmpDir + "/identity"

	// Create MockForwarder configured to fail (offline)
	mockFwd := forward.NewMockForwarder()
	mockFwd.ForwardError = errors.New("home device offline")

	// Create message queue
	queueCfg := queue.QueueConfig{
		KeyPath:     keyPath,
		MaxRAMBytes: 1024 * 1024,
		MaxMessages: 10,
		TTLHours:    1,
	}
	msgQueue, err := queue.NewMessageQueue(queueCfg)
	if err != nil {
		t.Fatalf("NewMessageQueue: %v", err)
	}
	defer msgQueue.Close()

	// Create QueuedForwarder
	queuedFwd := forward.NewQueuedForwarder(mockFwd, msgQueue, true)

	// Send message with known plaintext
	from := "sender@example.com"
	to := []string{"recipient@example.com"}
	knownPlaintext := "Hello, this is a test message with secret content"
	message := []byte("Message-ID: <encrypt-test@example.com>\r\nSubject: Encryption Test\r\n\r\n" + knownPlaintext + "\r\n")

	err = queuedFwd.Forward(context.Background(), from, to, bytes.NewReader(message))
	if err != nil {
		t.Fatalf("Forward: %v", err)
	}

	// Verify message is queued
	if msgQueue.Len() != 1 {
		t.Fatalf("queue length = %d, want 1", msgQueue.Len())
	}

	// Dequeue message to verify encryption
	msgID := "<encrypt-test@example.com>"
	queuedMsg, plaintext, err := msgQueue.Dequeue(msgID)
	if err != nil {
		t.Fatalf("Dequeue: %v", err)
	}

	// Verify EncryptedData does NOT contain plaintext
	if queuedMsg.EncryptedData != nil && bytes.Contains(queuedMsg.EncryptedData, []byte(knownPlaintext)) {
		t.Error("EncryptedData contains plaintext (not encrypted!)")
	}

	// Verify decrypted plaintext matches original
	if !bytes.Contains(plaintext, []byte(knownPlaintext)) {
		t.Errorf("decrypted plaintext missing original content: %s", plaintext)
	}
}

func TestIntegration_QueueDedup(t *testing.T) {
	// Verify duplicate messages with same Message-ID are deduplicated.
	t.Parallel()

	// Create temp directory for queue keys
	tmpDir := t.TempDir()
	keyPath := tmpDir + "/identity"

	// Create MockForwarder configured to fail (offline)
	mockFwd := forward.NewMockForwarder()
	mockFwd.ForwardError = errors.New("home device offline")

	// Create message queue
	queueCfg := queue.QueueConfig{
		KeyPath:     keyPath,
		MaxRAMBytes: 1024 * 1024,
		MaxMessages: 10,
		TTLHours:    1,
	}
	msgQueue, err := queue.NewMessageQueue(queueCfg)
	if err != nil {
		t.Fatalf("NewMessageQueue: %v", err)
	}
	defer msgQueue.Close()

	// Create QueuedForwarder
	queuedFwd := forward.NewQueuedForwarder(mockFwd, msgQueue, true)

	// Send first message
	from := "sender@example.com"
	to := []string{"recipient@example.com"}
	messageID := "<dedup-test@example.com>"
	message1 := []byte("Message-ID: " + messageID + "\r\nSubject: First\r\n\r\nFirst body\r\n")

	err = queuedFwd.Forward(context.Background(), from, to, bytes.NewReader(message1))
	if err != nil {
		t.Fatalf("Forward (first): %v", err)
	}

	// Send second message with SAME Message-ID but different body
	message2 := []byte("Message-ID: " + messageID + "\r\nSubject: Second\r\n\r\nSecond body\r\n")

	err = queuedFwd.Forward(context.Background(), from, to, bytes.NewReader(message2))
	if err != nil {
		t.Fatalf("Forward (second): %v", err)
	}

	// Verify only one message is queued (dedup worked)
	if msgQueue.Len() != 1 {
		t.Errorf("queue length = %d, want 1 (second message deduplicated)", msgQueue.Len())
	}
}

func TestIntegration_FullSMTPWithQueue(t *testing.T) {
	// End-to-end SMTP session with queue (extends existing integration test pattern).
	t.Parallel()

	// Create temp directory for queue keys
	tmpDir := t.TempDir()
	keyPath := tmpDir + "/identity"

	// Create MockForwarder that fails initially
	mockFwd := forward.NewMockForwarder()
	mockFwd.ForwardError = errors.New("home device offline")

	// Create message queue
	queueCfg := queue.QueueConfig{
		KeyPath:     keyPath,
		MaxRAMBytes: 1024 * 1024,
		MaxMessages: 10,
		TTLHours:    1,
	}
	msgQueue, err := queue.NewMessageQueue(queueCfg)
	if err != nil {
		t.Fatalf("NewMessageQueue: %v", err)
	}
	defer msgQueue.Close()

	// Create QueuedForwarder
	queuedFwd := forward.NewQueuedForwarder(mockFwd, msgQueue, true)

	// Start SMTP server
	serverAddr := startTestServer(t, queuedFwd)

	// Send message via SMTP client
	from := "sender@example.com"
	to := []string{"recipient@example.com"}
	message := []byte("Message-ID: <smtp-queue-test@example.com>\r\nSubject: SMTP Queue Test\r\n\r\nTest message via SMTP\r\n")

	err = smtp.SendMail(serverAddr, nil, from, to, message)
	if err != nil {
		t.Fatalf("SendMail: %v", err)
	}

	// Give SMTP session time to process
	time.Sleep(100 * time.Millisecond)

	// Verify SMTP accepted (no error from client perspective)
	// Verify queue has the message
	if msgQueue.Len() != 1 {
		t.Fatalf("queue length = %d, want 1 (message queued)", msgQueue.Len())
	}

	// Simulate reconnect: process queue with succeeding MockForwarder
	successFwd := forward.NewMockForwarder()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go msgQueue.StartProcessor(ctx, successFwd, 100*time.Millisecond)

	// Wait for delivery
	time.Sleep(500 * time.Millisecond)

	// Verify message delivered
	deliveryCalls := successFwd.GetCalls()
	if len(deliveryCalls) < 1 {
		t.Fatalf("delivery calls = %d, want at least 1", len(deliveryCalls))
	}

	call := deliveryCalls[0]
	if !strings.Contains(call.Data, "SMTP Queue Test") {
		t.Errorf("delivered data missing subject: %s", call.Data)
	}

	// Verify queue is empty
	if msgQueue.Len() != 0 {
		t.Errorf("queue length after delivery = %d, want 0", msgQueue.Len())
	}
}
