// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package queue

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEnqueueDequeue(t *testing.T) {
	tempDir := t.TempDir()
	cfg := QueueConfig{
		KeyPath:       filepath.Join(tempDir, "identity"),
		MaxRAMBytes:   10 * 1024 * 1024, // 10MB
		MaxMessages:   100,
		TTLHours:      24,
		SnapshotPath:  filepath.Join(tempDir, "snapshot.json"),
	}

	q, err := NewMessageQueue(cfg)
	if err != nil {
		t.Fatalf("NewMessageQueue() error: %v", err)
	}
	defer q.Close()

	// Enqueue a message
	from := "sender@example.com"
	to := []string{"recipient@example.com"}
	msgData := []byte("Message-ID: <test123@example.com>\r\n\r\nTest body")

	ctx := context.Background()
	if err := q.Enqueue(ctx, from, to, bytes.NewReader(msgData)); err != nil {
		t.Fatalf("Enqueue() error: %v", err)
	}

	// Verify queue length
	if q.Len() != 1 {
		t.Errorf("Queue length = %d, want 1", q.Len())
	}

	// Verify RAM usage
	if q.RAMUsage() == 0 {
		t.Error("RAM usage is 0 after enqueue")
	}

	// Dequeue the message
	msg, plaintext, err := q.Dequeue("<test123@example.com>")
	if err != nil {
		t.Fatalf("Dequeue() error: %v", err)
	}

	// Verify message metadata
	if msg.From != from {
		t.Errorf("Message from = %q, want %q", msg.From, from)
	}
	if len(msg.To) != 1 || msg.To[0] != to[0] {
		t.Errorf("Message to = %v, want %v", msg.To, to)
	}

	// Verify plaintext
	if !bytes.Equal(plaintext, msgData) {
		t.Errorf("Decrypted data doesn't match original.\nWant: %q\nGot:  %q", msgData, plaintext)
	}

	// Verify queue is empty
	if q.Len() != 0 {
		t.Errorf("Queue length = %d after dequeue, want 0", q.Len())
	}

	// Verify RAM usage is 0
	if q.RAMUsage() != 0 {
		t.Errorf("RAM usage = %d after dequeue, want 0", q.RAMUsage())
	}
}

func TestDedup(t *testing.T) {
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

	msgData := []byte("Message-ID: <duplicate@example.com>\r\n\r\nTest")
	ctx := context.Background()

	// Enqueue same message twice
	if err := q.Enqueue(ctx, "sender@example.com", []string{"rcpt@example.com"}, bytes.NewReader(msgData)); err != nil {
		t.Fatalf("First Enqueue() error: %v", err)
	}

	if err := q.Enqueue(ctx, "sender@example.com", []string{"rcpt@example.com"}, bytes.NewReader(msgData)); err != nil {
		t.Fatalf("Second Enqueue() error: %v", err)
	}

	// Should only have one message (deduplicated)
	if q.Len() != 1 {
		t.Errorf("Queue length = %d, want 1 (deduplication failed)", q.Len())
	}
}

func TestFallbackMessageID(t *testing.T) {
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

	// Message without Message-ID header
	msgData := []byte("Subject: Test\r\n\r\nBody without Message-ID")
	ctx := context.Background()

	if err := q.Enqueue(ctx, "sender@example.com", []string{"rcpt@example.com"}, bytes.NewReader(msgData)); err != nil {
		t.Fatalf("Enqueue() error: %v", err)
	}

	// Should have generated fallback ID
	if q.Len() != 1 {
		t.Errorf("Queue length = %d, want 1", q.Len())
	}

	// Check that the message ID is a fallback (contains "fallback-")
	q.mu.RLock()
	foundFallback := false
	for id := range q.messages {
		if bytes.Contains([]byte(id), []byte("fallback-")) {
			foundFallback = true
			break
		}
	}
	q.mu.RUnlock()

	if !foundFallback {
		t.Error("Expected fallback Message-ID to be generated")
	}
}

func TestQueueFull(t *testing.T) {
	tempDir := t.TempDir()
	cfg := QueueConfig{
		KeyPath:      filepath.Join(tempDir, "identity"),
		MaxRAMBytes:  1024, // 1KB limit (very small)
		MaxMessages:  100,
		TTLHours:     24,
		SnapshotPath: filepath.Join(tempDir, "snapshot.json"),
	}

	q, err := NewMessageQueue(cfg)
	if err != nil {
		t.Fatalf("NewMessageQueue() error: %v", err)
	}
	defer q.Close()

	ctx := context.Background()

	// Try to enqueue a large message that exceeds RAM limit
	largeMsg := bytes.Repeat([]byte("x"), 2048) // 2KB message
	msgData := append([]byte("Message-ID: <large@example.com>\r\n\r\n"), largeMsg...)

	err = q.Enqueue(ctx, "sender@example.com", []string{"rcpt@example.com"}, bytes.NewReader(msgData))
	if err != ErrQueueFull {
		t.Errorf("Enqueue() error = %v, want ErrQueueFull", err)
	}

	// Queue should be empty
	if q.Len() != 0 {
		t.Errorf("Queue length = %d, want 0 (message should not be enqueued)", q.Len())
	}
}

func TestPurgeExpired(t *testing.T) {
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

	ctx := context.Background()

	// Enqueue a message
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

	// Purge expired messages
	purged := q.PurgeExpired()
	if purged != 1 {
		t.Errorf("PurgeExpired() = %d, want 1", purged)
	}

	// Queue should be empty
	if q.Len() != 0 {
		t.Errorf("Queue length = %d after purge, want 0", q.Len())
	}

	// RAM usage should be 0
	if q.RAMUsage() != 0 {
		t.Errorf("RAM usage = %d after purge, want 0", q.RAMUsage())
	}
}

func TestSnapshot(t *testing.T) {
	tempDir := t.TempDir()
	snapshotPath := filepath.Join(tempDir, "snapshot.json")
	cfg := QueueConfig{
		KeyPath:      filepath.Join(tempDir, "identity"),
		MaxRAMBytes:  10 * 1024 * 1024,
		MaxMessages:  100,
		TTLHours:     24,
		SnapshotPath: snapshotPath,
	}

	q, err := NewMessageQueue(cfg)
	if err != nil {
		t.Fatalf("NewMessageQueue() error: %v", err)
	}

	// Enqueue a message
	msgData := []byte("Message-ID: <snapshot-test@example.com>\r\n\r\nTest")
	ctx := context.Background()
	if err := q.Enqueue(ctx, "sender@example.com", []string{"rcpt@example.com"}, bytes.NewReader(msgData)); err != nil {
		t.Fatalf("Enqueue() error: %v", err)
	}

	// Write snapshot
	if err := q.Snapshot(); err != nil {
		t.Fatalf("Snapshot() error: %v", err)
	}

	// Verify snapshot file exists
	if _, err := os.Stat(snapshotPath); err != nil {
		t.Fatalf("Snapshot file not created: %v", err)
	}

	// Verify snapshot content is valid JSON
	data, err := os.ReadFile(snapshotPath)
	if err != nil {
		t.Fatalf("Read snapshot file error: %v", err)
	}

	if !strings.Contains(string(data), "snapshot-test@example.com") {
		t.Error("Snapshot doesn't contain expected message ID")
	}

	q.Close()
}

func TestQueueRAMTracking(t *testing.T) {
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

	ctx := context.Background()

	// Initial RAM should be 0
	if q.RAMUsage() != 0 {
		t.Errorf("Initial RAM usage = %d, want 0", q.RAMUsage())
	}

	// Enqueue multiple messages
	for i := 0; i < 3; i++ {
		msgData := []byte("Message-ID: <msg" + string(rune('0'+i)) + "@example.com>\r\n\r\nTest")
		if err := q.Enqueue(ctx, "sender@example.com", []string{"rcpt@example.com"}, bytes.NewReader(msgData)); err != nil {
			t.Fatalf("Enqueue() message %d error: %v", i, err)
		}
	}

	ramAfterEnqueue := q.RAMUsage()
	if ramAfterEnqueue == 0 {
		t.Error("RAM usage is 0 after enqueuing 3 messages")
	}

	// Dequeue one message
	q.mu.RLock()
	var firstID string
	if len(q.order) > 0 {
		firstID = q.order[0]
	}
	q.mu.RUnlock()

	if _, _, err := q.Dequeue(firstID); err != nil {
		t.Fatalf("Dequeue() error: %v", err)
	}

	ramAfterDequeue := q.RAMUsage()
	if ramAfterDequeue >= ramAfterEnqueue {
		t.Errorf("RAM usage after dequeue (%d) >= before dequeue (%d)", ramAfterDequeue, ramAfterEnqueue)
	}
}

func TestDequeueNotFound(t *testing.T) {
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

	// Try to dequeue non-existent message
	_, _, err = q.Dequeue("<nonexistent@example.com>")
	if err != ErrNotFound {
		t.Errorf("Dequeue() error = %v, want ErrNotFound", err)
	}
}

func TestQueueMaxMessages(t *testing.T) {
	tempDir := t.TempDir()
	cfg := QueueConfig{
		KeyPath:      filepath.Join(tempDir, "identity"),
		MaxRAMBytes:  100 * 1024 * 1024, // High RAM limit
		MaxMessages:  2,                 // Only 2 messages allowed
		TTLHours:     24,
		SnapshotPath: filepath.Join(tempDir, "snapshot.json"),
	}

	q, err := NewMessageQueue(cfg)
	if err != nil {
		t.Fatalf("NewMessageQueue() error: %v", err)
	}
	defer q.Close()

	ctx := context.Background()

	// Enqueue 2 messages (should succeed)
	for i := 0; i < 2; i++ {
		msgData := []byte("Message-ID: <msg" + string(rune('0'+i)) + "@example.com>\r\n\r\nTest")
		if err := q.Enqueue(ctx, "sender@example.com", []string{"rcpt@example.com"}, bytes.NewReader(msgData)); err != nil {
			t.Fatalf("Enqueue() message %d error: %v", i, err)
		}
	}

	// Try to enqueue 3rd message (should fail)
	msgData := []byte("Message-ID: <msg3@example.com>\r\n\r\nTest")
	err = q.Enqueue(ctx, "sender@example.com", []string{"rcpt@example.com"}, bytes.NewReader(msgData))
	if err != ErrQueueFull {
		t.Errorf("Enqueue() 3rd message error = %v, want ErrQueueFull", err)
	}
}
