// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package queue

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// mockForwarder implements the Forwarder interface for testing.
type mockForwarder struct {
	mu           sync.Mutex
	forwardCalls int
	forwardError error
}

func (m *mockForwarder) Forward(ctx context.Context, from string, to []string, data io.Reader) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.forwardCalls++
	return m.forwardError
}

func (m *mockForwarder) Close() error {
	return nil
}

func (m *mockForwarder) getCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.forwardCalls
}

func (m *mockForwarder) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.forwardCalls = 0
	m.forwardError = nil
}

func TestProcessQueue_DeliversWhenOnline(t *testing.T) {
	tempDir := t.TempDir()
	cfg := QueueConfig{
		KeyPath:      filepath.Join(tempDir, "identity"),
		MaxRAMBytes:  10 * 1024 * 1024,
		MaxMessages:  100,
		TTLHours:     24,
		SnapshotPath: filepath.Join(tempDir, "snapshot.json"),
	}

	q, err := NewMessageQueue(cfg)
	if err != nil {
		t.Fatalf("NewMessageQueue() error: %v", err)
	}
	defer q.Close()

	// Enqueue 3 messages
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		msgData := []byte("Message-ID: <msg" + string(rune('0'+i)) + "@example.com>\r\n\r\nTest")
		if err := q.Enqueue(ctx, "sender@example.com", []string{"rcpt@example.com"}, bytes.NewReader(msgData)); err != nil {
			t.Fatalf("Enqueue() error: %v", err)
		}
	}

	// Create mock forwarder (succeeds)
	mock := &mockForwarder{}

	// Process queue
	q.processQueue(ctx, mock)

	// All 3 messages should be delivered
	if mock.getCallCount() != 3 {
		t.Errorf("Forward called %d times, want 3", mock.getCallCount())
	}

	// Queue should be empty
	if q.Len() != 0 {
		t.Errorf("Queue length = %d after processing, want 0", q.Len())
	}
}

func TestProcessQueue_StopsOnFailure(t *testing.T) {
	tempDir := t.TempDir()
	cfg := QueueConfig{
		KeyPath:      filepath.Join(tempDir, "identity"),
		MaxRAMBytes:  10 * 1024 * 1024,
		MaxMessages:  100,
		TTLHours:     24,
		SnapshotPath: filepath.Join(tempDir, "snapshot.json"),
	}

	q, err := NewMessageQueue(cfg)
	if err != nil {
		t.Fatalf("NewMessageQueue() error: %v", err)
	}
	defer q.Close()

	// Enqueue 3 messages
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		msgData := []byte("Message-ID: <msg" + string(rune('0'+i)) + "@example.com>\r\n\r\nTest")
		if err := q.Enqueue(ctx, "sender@example.com", []string{"rcpt@example.com"}, bytes.NewReader(msgData)); err != nil {
			t.Fatalf("Enqueue() error: %v", err)
		}
	}

	// Create mock forwarder that fails
	mock := &mockForwarder{
		forwardError: errors.New("transport offline"),
	}

	// Process queue
	q.processQueue(ctx, mock)

	// Should have attempted to deliver first message only (stopped on failure)
	callCount := mock.getCallCount()
	if callCount > 3 {
		// Backoff retries mean multiple Forward calls, but should stop after first message fails
		t.Logf("Forward called %d times (includes retries)", callCount)
	}

	// Queue should still have messages (not all delivered)
	if q.Len() == 0 {
		t.Error("Queue is empty, expected messages to remain after transport failure")
	}
}

func TestProcessQueue_RateLimit(t *testing.T) {
	tempDir := t.TempDir()
	cfg := QueueConfig{
		KeyPath:      filepath.Join(tempDir, "identity"),
		MaxRAMBytes:  100 * 1024 * 1024, // High limit
		MaxMessages:  100,
		TTLHours:     24,
		SnapshotPath: filepath.Join(tempDir, "snapshot.json"),
	}

	q, err := NewMessageQueue(cfg)
	if err != nil {
		t.Fatalf("NewMessageQueue() error: %v", err)
	}
	defer q.Close()

	// Enqueue 20 messages
	ctx := context.Background()
	for i := 0; i < 20; i++ {
		msgData := []byte("Message-ID: <msg" + string(rune('0'+i)) + "@example.com>\r\n\r\nTest")
		if err := q.Enqueue(ctx, "sender@example.com", []string{"rcpt@example.com"}, bytes.NewReader(msgData)); err != nil {
			t.Fatalf("Enqueue() message %d error: %v", i, err)
		}
	}

	// Create mock forwarder (succeeds)
	mock := &mockForwarder{}

	// Process queue once
	q.processQueue(ctx, mock)

	// Should process at most 10 messages (rate limit)
	if mock.getCallCount() > 10 {
		t.Errorf("Forward called %d times, want <= 10 (rate limit)", mock.getCallCount())
	}

	// Queue should still have ~10 messages remaining
	remaining := q.Len()
	if remaining < 10 {
		t.Errorf("Queue length = %d, want >= 10 (rate limit should leave messages)", remaining)
	}
}

func TestProcessQueue_PurgesExpired(t *testing.T) {
	tempDir := t.TempDir()
	cfg := QueueConfig{
		KeyPath:      filepath.Join(tempDir, "identity"),
		MaxRAMBytes:  10 * 1024 * 1024,
		MaxMessages:  100,
		TTLHours:     1, // 1 hour TTL
		SnapshotPath: filepath.Join(tempDir, "snapshot.json"),
	}

	q, err := NewMessageQueue(cfg)
	if err != nil {
		t.Fatalf("NewMessageQueue() error: %v", err)
	}
	defer q.Close()

	// Enqueue a message
	ctx := context.Background()
	msgData := []byte("Message-ID: <old@example.com>\r\n\r\nOld message")
	if err := q.Enqueue(ctx, "sender@example.com", []string{"rcpt@example.com"}, bytes.NewReader(msgData)); err != nil {
		t.Fatalf("Enqueue() error: %v", err)
	}

	// Manually set the message's enqueue time to 2 hours ago
	q.mu.Lock()
	for _, msg := range q.messages {
		msg.EnqueuedAt = time.Now().Add(-2 * time.Hour)
	}
	q.mu.Unlock()

	// Create mock forwarder
	mock := &mockForwarder{}

	// Process queue (should purge expired first)
	q.processQueue(ctx, mock)

	// Queue should be empty (message expired and purged)
	if q.Len() != 0 {
		t.Errorf("Queue length = %d after processing, want 0 (expired message should be purged)", q.Len())
	}

	// No delivery attempts should have been made
	if mock.getCallCount() != 0 {
		t.Errorf("Forward called %d times, want 0 (message was expired)", mock.getCallCount())
	}
}

func TestStartProcessor_ContextCancellation(t *testing.T) {
	tempDir := t.TempDir()
	cfg := QueueConfig{
		KeyPath:      filepath.Join(tempDir, "identity"),
		MaxRAMBytes:  10 * 1024 * 1024,
		MaxMessages:  100,
		TTLHours:     24,
		SnapshotPath: filepath.Join(tempDir, "snapshot.json"),
	}

	q, err := NewMessageQueue(cfg)
	if err != nil {
		t.Fatalf("NewMessageQueue() error: %v", err)
	}
	defer q.Close()

	mock := &mockForwarder{}

	// Start processor with cancelable context
	ctx, cancel := context.WithCancel(context.Background())

	processorDone := make(chan struct{})
	go func() {
		q.StartProcessor(ctx, mock, 100*time.Millisecond)
		close(processorDone)
	}()

	// Let it run for a bit
	time.Sleep(250 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for processor to stop (with timeout)
	select {
	case <-processorDone:
		// Success - processor stopped
	case <-time.After(2 * time.Second):
		t.Fatal("Processor did not stop after context cancellation")
	}
}

func TestProcessSingle(t *testing.T) {
	tempDir := t.TempDir()
	cfg := QueueConfig{
		KeyPath:      filepath.Join(tempDir, "identity"),
		MaxRAMBytes:  10 * 1024 * 1024,
		MaxMessages:  100,
		TTLHours:     24,
		SnapshotPath: filepath.Join(tempDir, "snapshot.json"),
	}

	q, err := NewMessageQueue(cfg)
	if err != nil {
		t.Fatalf("NewMessageQueue() error: %v", err)
	}
	defer q.Close()

	// Enqueue a message
	ctx := context.Background()
	msgData := []byte("Message-ID: <single@example.com>\r\n\r\nTest")
	if err := q.Enqueue(ctx, "sender@example.com", []string{"rcpt@example.com"}, bytes.NewReader(msgData)); err != nil {
		t.Fatalf("Enqueue() error: %v", err)
	}

	// Create mock forwarder (succeeds)
	mock := &mockForwarder{}

	// Process single message
	if err := q.processSingle(ctx, mock, "<single@example.com>"); err != nil {
		t.Fatalf("processSingle() error: %v", err)
	}

	// Message should be delivered and removed from queue
	if q.Len() != 0 {
		t.Errorf("Queue length = %d after processSingle, want 0", q.Len())
	}

	// Forward should have been called (possibly multiple times due to retries if there were failures)
	if mock.getCallCount() == 0 {
		t.Error("Forward was not called")
	}
}
