// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package queue provides Postfix mail queue monitoring via postqueue -j JSON output.
package queue

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// QueueMessage represents a single message in the Postfix queue.
type QueueMessage struct {
	QueueID     string      `json:"queue_id"`
	QueueName   string      `json:"queue_name"`   // active, deferred, hold, etc.
	ArrivalTime int64       `json:"arrival_time"` // Unix epoch timestamp
	MessageSize int         `json:"message_size"`
	Sender      string      `json:"sender"`
	Recipients  []Recipient `json:"recipients"`
}

// Recipient represents a message recipient with delivery status.
type Recipient struct {
	Address     string `json:"address"`
	DelayReason string `json:"delay_reason,omitempty"`
}

// PostqueueExecutor defines the interface for executing postqueue commands.
// This allows for testing with mock implementations.
type PostqueueExecutor interface {
	Execute() ([]byte, error)
}

// RealPostqueueExecutor executes the actual postqueue command.
type RealPostqueueExecutor struct{}

// Execute runs postqueue -j and returns the output.
func (r *RealPostqueueExecutor) Execute() ([]byte, error) {
	cmd := exec.Command("postqueue", "-j")
	return cmd.CombinedOutput()
}

// defaultExecutor is the default postqueue executor.
var defaultExecutor PostqueueExecutor = &RealPostqueueExecutor{}

// SetPostqueueExecutor allows injecting a custom executor (primarily for testing).
func SetPostqueueExecutor(executor PostqueueExecutor) {
	defaultExecutor = executor
}

// GetQueueStats retrieves queue statistics from Postfix.
func GetQueueStats() (*QueueStats, error) {
	return GetQueueStatsWithExecutor(defaultExecutor)
}

// GetQueueStatsWithExecutor retrieves queue stats using the provided executor.
func GetQueueStatsWithExecutor(executor PostqueueExecutor) (*QueueStats, error) {
	output, err := executor.Execute()
	if err != nil {
		// Empty queue returns exit code 0 with no output
		// Non-zero exit with no output likely means queue is empty but command succeeded
		if len(output) == 0 {
			return &QueueStats{
				Depth:     0,
				Active:    0,
				Deferred:  0,
				Hold:      0,
				Stuck:     0,
				Timestamp: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("postqueue failed: %w (output: %s)", err, string(output))
	}

	// Parse NDJSON output
	messages, err := parsePostqueueJSON(output)
	if err != nil {
		return nil, fmt.Errorf("parse postqueue output: %w", err)
	}

	// Calculate statistics
	stats := calculateStats(messages)
	return stats, nil
}

// parsePostqueueJSON parses NDJSON output from postqueue -j.
func parsePostqueueJSON(data []byte) ([]QueueMessage, error) {
	if len(data) == 0 {
		return []QueueMessage{}, nil
	}

	var messages []QueueMessage
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var msg QueueMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			// Skip malformed lines rather than failing completely
			continue
		}

		messages = append(messages, msg)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan postqueue output: %w", err)
	}

	return messages, nil
}

// calculateStats computes queue statistics from parsed messages.
func calculateStats(messages []QueueMessage) *QueueStats {
	stats := &QueueStats{
		Timestamp:      time.Now(),
		StuckThreshold: 24 * time.Hour, // Default 24-hour threshold
	}

	if len(messages) == 0 {
		return stats
	}

	stats.Depth = len(messages)

	now := time.Now()
	var oldestTime time.Time

	for _, msg := range messages {
		// Count by queue name
		switch msg.QueueName {
		case "active":
			stats.Active++
		case "deferred":
			stats.Deferred++
		case "hold":
			stats.Hold++
		}

		// Check for stuck messages (older than threshold)
		arrivalTime := time.Unix(msg.ArrivalTime, 0)
		age := now.Sub(arrivalTime)
		if age > stats.StuckThreshold {
			stats.Stuck++
		}

		// Track oldest message
		if oldestTime.IsZero() || arrivalTime.Before(oldestTime) {
			oldestTime = arrivalTime
		}
	}

	stats.OldestMessage = oldestTime
	return stats
}
