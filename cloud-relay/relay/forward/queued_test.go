// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package forward

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/darkpipe/darkpipe/cloud-relay/relay/queue"
)

func TestQueuedForwarder_TransportSuccess(t *testing.T) {
	// Create mock transport (succeeds)
	mock := NewMockForwarder()

	// Create queue
	tempDir := t.TempDir()
	cfg := queue.QueueConfig{
		KeyPath:      filepath.Join(tempDir, "identity"),
		MaxRAMBytes:  10 * 1024 * 1024,
		MaxMessages:  100,
		TTLHours:     24,
		SnapshotPath: filepath.Join(tempDir, "snapshot.json"),
	}
	q, err := queue.NewMessageQueue(cfg)
	if err != nil {
		t.Fatalf("NewMessageQueue() error: %v", err)
	}
	defer q.Close()

	// Create queued forwarder
	qf := NewQueuedForwarder(mock, q, true)

	// Forward a message
	msgData := []byte("Message-ID: <test@example.com>\r\n\r\nTest body")
	ctx := context.Background()
	err = qf.Forward(ctx, "sender@example.com", []string{"rcpt@example.com"}, bytes.NewReader(msgData))
	if err != nil {
		t.Fatalf("Forward() error: %v", err)
	}

	// Transport should have been called
	calls := mock.GetCalls()
	if len(calls) != 1 {
		t.Errorf("Transport Forward called %d times, want 1", len(calls))
	}

	// Queue should be empty (message delivered, not queued)
	if q.Len() != 0 {
		t.Errorf("Queue length = %d, want 0 (message should not be queued on success)", q.Len())
	}
}

func TestQueuedForwarder_TransportFailure_QueueEnabled(t *testing.T) {
	// Create mock transport (fails)
	mock := NewMockForwarder()
	mock.ForwardError = errors.New("transport offline")

	// Create queue
	tempDir := t.TempDir()
	cfg := queue.QueueConfig{
		KeyPath:      filepath.Join(tempDir, "identity"),
		MaxRAMBytes:  10 * 1024 * 1024,
		MaxMessages:  100,
		TTLHours:     24,
		SnapshotPath: filepath.Join(tempDir, "snapshot.json"),
	}
	q, err := queue.NewMessageQueue(cfg)
	if err != nil {
		t.Fatalf("NewMessageQueue() error: %v", err)
	}
	defer q.Close()

	// Create queued forwarder with queue enabled
	qf := NewQueuedForwarder(mock, q, true)

	// Forward a message
	msgData := []byte("Message-ID: <queued@example.com>\r\n\r\nTest body")
	ctx := context.Background()
	err = qf.Forward(ctx, "sender@example.com", []string{"rcpt@example.com"}, bytes.NewReader(msgData))
	if err != nil {
		t.Fatalf("Forward() returned error %v, want nil (message should be queued)", err)
	}

	// Queue should have the message
	if q.Len() != 1 {
		t.Errorf("Queue length = %d, want 1 (message should be queued)", q.Len())
	}
}

func TestQueuedForwarder_TransportFailure_QueueDisabled(t *testing.T) {
	// Create mock transport (fails)
	mock := NewMockForwarder()
	mock.ForwardError = errors.New("transport offline")

	// Create queue
	tempDir := t.TempDir()
	cfg := queue.QueueConfig{
		KeyPath:      filepath.Join(tempDir, "identity"),
		MaxRAMBytes:  10 * 1024 * 1024,
		MaxMessages:  100,
		TTLHours:     24,
		SnapshotPath: filepath.Join(tempDir, "snapshot.json"),
	}
	q, err := queue.NewMessageQueue(cfg)
	if err != nil {
		t.Fatalf("NewMessageQueue() error: %v", err)
	}
	defer q.Close()

	// Create queued forwarder with queue DISABLED
	qf := NewQueuedForwarder(mock, q, false)

	// Forward a message
	msgData := []byte("Message-ID: <bounce@example.com>\r\n\r\nTest body")
	ctx := context.Background()
	err = qf.Forward(ctx, "sender@example.com", []string{"rcpt@example.com"}, bytes.NewReader(msgData))
	if err == nil {
		t.Fatal("Forward() returned nil, want error (queue disabled)")
	}

	// Queue should be empty (message not queued when queue disabled)
	if q.Len() != 0 {
		t.Errorf("Queue length = %d, want 0 (message should not be queued when disabled)", q.Len())
	}
}

func TestQueuedForwarder_QueueFull(t *testing.T) {
	// Create mock transport (fails)
	mock := NewMockForwarder()
	mock.ForwardError = errors.New("transport offline")

	// Create queue with very small limit
	tempDir := t.TempDir()
	cfg := queue.QueueConfig{
		KeyPath:      filepath.Join(tempDir, "identity"),
		MaxRAMBytes:  100, // Very small - will fill quickly
		MaxMessages:  100,
		TTLHours:     24,
		SnapshotPath: filepath.Join(tempDir, "snapshot.json"),
	}
	q, err := queue.NewMessageQueue(cfg)
	if err != nil {
		t.Fatalf("NewMessageQueue() error: %v", err)
	}
	defer q.Close()

	// Create queued forwarder
	qf := NewQueuedForwarder(mock, q, true)

	// Try to forward a large message
	largeMsg := bytes.Repeat([]byte("x"), 10000)
	msgData := append([]byte("Message-ID: <large@example.com>\r\n\r\n"), largeMsg...)
	ctx := context.Background()
	err = qf.Forward(ctx, "sender@example.com", []string{"rcpt@example.com"}, bytes.NewReader(msgData))
	if err == nil {
		t.Fatal("Forward() returned nil, want error (queue should be full)")
	}

	// Should get transport error back (not queue full error)
	// This allows Postfix to send 4xx to sender
}

func TestQueuedForwarder_ImplementsForwarderInterface(t *testing.T) {
	// Compile-time check that QueuedForwarder implements Forwarder
	var _ Forwarder = (*QueuedForwarder)(nil)
}

func TestQueuedForwarder_Close(t *testing.T) {
	mock := NewMockForwarder()

	tempDir := t.TempDir()
	cfg := queue.QueueConfig{
		KeyPath:      filepath.Join(tempDir, "identity"),
		MaxRAMBytes:  10 * 1024 * 1024,
		MaxMessages:  100,
		TTLHours:     24,
		SnapshotPath: filepath.Join(tempDir, "snapshot.json"),
	}
	q, err := queue.NewMessageQueue(cfg)
	if err != nil {
		t.Fatalf("NewMessageQueue() error: %v", err)
	}

	qf := NewQueuedForwarder(mock, q, true)

	// Close should not error
	if err := qf.Close(); err != nil {
		t.Errorf("Close() error: %v", err)
	}
}

func TestQueuedForwarder_QueueAccessor(t *testing.T) {
	mock := NewMockForwarder()

	tempDir := t.TempDir()
	cfg := queue.QueueConfig{
		KeyPath:      filepath.Join(tempDir, "identity"),
		MaxRAMBytes:  10 * 1024 * 1024,
		MaxMessages:  100,
		TTLHours:     24,
		SnapshotPath: filepath.Join(tempDir, "snapshot.json"),
	}
	q, err := queue.NewMessageQueue(cfg)
	if err != nil {
		t.Fatalf("NewMessageQueue() error: %v", err)
	}
	defer q.Close()

	qf := NewQueuedForwarder(mock, q, true)

	// Queue() should return the queue
	if qf.Queue() != q {
		t.Error("Queue() didn't return the correct queue")
	}
}
