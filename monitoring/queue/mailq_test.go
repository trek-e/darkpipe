// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package queue

import (
	"fmt"
	"testing"
	"time"
)

// MockPostqueueExecutor allows injecting test data.
type MockPostqueueExecutor struct {
	Output []byte
	Err    error
}

func (m *MockPostqueueExecutor) Execute() ([]byte, error) {
	return m.Output, m.Err
}

func TestGetQueueStats_EmptyQueue(t *testing.T) {
	executor := &MockPostqueueExecutor{
		Output: []byte(""),
		Err:    nil,
	}

	stats, err := GetQueueStatsWithExecutor(executor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.Depth != 0 {
		t.Errorf("expected depth 0, got %d", stats.Depth)
	}
	if stats.Active != 0 {
		t.Errorf("expected active 0, got %d", stats.Active)
	}
	if stats.Deferred != 0 {
		t.Errorf("expected deferred 0, got %d", stats.Deferred)
	}
	if stats.Stuck != 0 {
		t.Errorf("expected stuck 0, got %d", stats.Stuck)
	}
}

func TestGetQueueStats_ParsesJSON(t *testing.T) {
	// Sample postqueue -j output (NDJSON format)
	// Note: arrival_time is Unix epoch integer, not RFC3339 string
	now := time.Now().Unix()
	jsonOutput := fmt.Sprintf(`{"queue_id":"ABC123","queue_name":"active","arrival_time":%d,"message_size":1234,"sender":"user@example.com","recipients":[{"address":"dest@test.com"}]}
{"queue_id":"DEF456","queue_name":"deferred","arrival_time":%d,"message_size":5678,"sender":"sender@example.org","recipients":[{"address":"rcpt@example.net","delay_reason":"Connection timed out"}]}
`, now, now-3600)

	executor := &MockPostqueueExecutor{
		Output: []byte(jsonOutput),
		Err:    nil,
	}

	stats, err := GetQueueStatsWithExecutor(executor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.Depth != 2 {
		t.Errorf("expected depth 2, got %d", stats.Depth)
	}
	if stats.Active != 1 {
		t.Errorf("expected active 1, got %d", stats.Active)
	}
	if stats.Deferred != 1 {
		t.Errorf("expected deferred 1, got %d", stats.Deferred)
	}
	if stats.Hold != 0 {
		t.Errorf("expected hold 0, got %d", stats.Hold)
	}

	// Check oldest message
	if stats.OldestMessage.IsZero() {
		t.Error("expected non-zero oldest message time")
	}
}

func TestGetQueueStats_StuckDetection(t *testing.T) {
	// Create a message that's 25 hours old (stuck)
	stuckTime := time.Now().Add(-25 * time.Hour).Unix()
	recentTime := time.Now().Add(-1 * time.Hour).Unix()

	jsonOutput := fmt.Sprintf(`{"queue_id":"OLD123","queue_name":"deferred","arrival_time":%d,"message_size":1000,"sender":"old@example.com","recipients":[{"address":"dest@test.com"}]}
{"queue_id":"NEW456","queue_name":"active","arrival_time":%d,"message_size":2000,"sender":"new@example.com","recipients":[{"address":"dest@test.com"}]}
`, stuckTime, recentTime)

	executor := &MockPostqueueExecutor{
		Output: []byte(jsonOutput),
		Err:    nil,
	}

	stats, err := GetQueueStatsWithExecutor(executor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.Depth != 2 {
		t.Errorf("expected depth 2, got %d", stats.Depth)
	}
	if stats.Stuck != 1 {
		t.Errorf("expected 1 stuck message, got %d", stats.Stuck)
	}

	// Verify oldest message is the stuck one
	expectedOldest := time.Unix(stuckTime, 0)
	if !stats.OldestMessage.Equal(expectedOldest) {
		t.Errorf("expected oldest message at %v, got %v", expectedOldest, stats.OldestMessage)
	}
}

func TestGetQueueStats_HoldQueue(t *testing.T) {
	now := time.Now().Unix()
	jsonOutput := fmt.Sprintf(`{"queue_id":"HOLD1","queue_name":"hold","arrival_time":%d,"message_size":1234,"sender":"hold@example.com","recipients":[{"address":"dest@test.com"}]}
`, now)

	executor := &MockPostqueueExecutor{
		Output: []byte(jsonOutput),
		Err:    nil,
	}

	stats, err := GetQueueStatsWithExecutor(executor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.Hold != 1 {
		t.Errorf("expected hold 1, got %d", stats.Hold)
	}
}

func TestParsePostqueueJSON_MalformedLine(t *testing.T) {
	// Include one malformed line - should skip it and parse the valid one
	now := time.Now().Unix()
	jsonOutput := fmt.Sprintf(`{"invalid json line
{"queue_id":"VALID1","queue_name":"active","arrival_time":%d,"message_size":1234,"sender":"user@example.com","recipients":[]}
`, now)

	messages, err := parsePostqueueJSON([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have parsed the valid line
	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}

	if messages[0].QueueID != "VALID1" {
		t.Errorf("expected queue_id VALID1, got %s", messages[0].QueueID)
	}
}

func TestGetDetailedQueue(t *testing.T) {
	now := time.Now().Unix()
	jsonOutput := fmt.Sprintf(`{"queue_id":"ABC123","queue_name":"active","arrival_time":%d,"message_size":1234,"sender":"user@example.com","recipients":[{"address":"dest@test.com"}]}
{"queue_id":"DEF456","queue_name":"deferred","arrival_time":%d,"message_size":5678,"sender":"sender@example.org","recipients":[{"address":"rcpt@example.net","delay_reason":"Deferred"}]}
`, now, now-1800)

	executor := &MockPostqueueExecutor{
		Output: []byte(jsonOutput),
		Err:    nil,
	}

	snapshot, err := GetDetailedQueueWithExecutor(executor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if snapshot.Depth != 2 {
		t.Errorf("expected depth 2, got %d", snapshot.Depth)
	}

	if len(snapshot.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(snapshot.Messages))
	}

	// Verify message details
	msg := snapshot.Messages[0]
	if msg.QueueID != "ABC123" {
		t.Errorf("expected queue_id ABC123, got %s", msg.QueueID)
	}
	if msg.Sender != "user@example.com" {
		t.Errorf("expected sender user@example.com, got %s", msg.Sender)
	}
	if len(msg.Recipients) != 1 {
		t.Errorf("expected 1 recipient, got %d", len(msg.Recipients))
	}

	// Verify deferred message
	msg2 := snapshot.Messages[1]
	if msg2.QueueName != "deferred" {
		t.Errorf("expected queue_name deferred, got %s", msg2.QueueName)
	}
	if msg2.Recipients[0].DelayReason != "Deferred" {
		t.Errorf("expected delay_reason 'Deferred', got '%s'", msg2.Recipients[0].DelayReason)
	}
}

func TestGetDetailedQueue_Empty(t *testing.T) {
	executor := &MockPostqueueExecutor{
		Output: []byte(""),
		Err:    nil,
	}

	snapshot, err := GetDetailedQueueWithExecutor(executor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if snapshot.Depth != 0 {
		t.Errorf("expected depth 0, got %d", snapshot.Depth)
	}

	if len(snapshot.Messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(snapshot.Messages))
	}
}

func TestQueueMessage_Recipients(t *testing.T) {
	now := time.Now().Unix()
	jsonOutput := fmt.Sprintf(`{"queue_id":"MULTI1","queue_name":"active","arrival_time":%d,"message_size":1234,"sender":"user@example.com","recipients":[{"address":"rcpt1@test.com"},{"address":"rcpt2@test.com","delay_reason":"Greylisted"}]}
`, now)

	messages, err := parsePostqueueJSON([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if len(msg.Recipients) != 2 {
		t.Fatalf("expected 2 recipients, got %d", len(msg.Recipients))
	}

	if msg.Recipients[0].Address != "rcpt1@test.com" {
		t.Errorf("expected rcpt1@test.com, got %s", msg.Recipients[0].Address)
	}

	if msg.Recipients[1].DelayReason != "Greylisted" {
		t.Errorf("expected delay_reason Greylisted, got %s", msg.Recipients[1].DelayReason)
	}
}
