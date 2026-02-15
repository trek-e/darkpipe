// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package notify provides a notification system for TLS events and warnings.
package notify

import (
	"context"
	"time"
)

// Event represents a notification event from the relay system.
type Event struct {
	Type      string            // "tls_warning", "tls_failure"
	Domain    string            // Remote server domain that triggered the event
	Message   string            // Human-readable description
	Timestamp time.Time         // When the event occurred
	Details   map[string]string // Additional metadata (IP, port, TLS version, etc.)
}

// Notifier is the interface for notification backends.
type Notifier interface {
	// Send sends an event notification.
	Send(ctx context.Context, event Event) error
	// Close closes the notifier and releases resources.
	Close() error
}

// MultiNotifier dispatches events to multiple notification backends (fan-out).
type MultiNotifier struct {
	notifiers []Notifier
}

// NewMultiNotifier creates a new MultiNotifier that dispatches to all provided backends.
func NewMultiNotifier(notifiers ...Notifier) *MultiNotifier {
	return &MultiNotifier{
		notifiers: notifiers,
	}
}

// Send dispatches the event to all backends and collects errors.
// Returns an error if any backend fails, but continues sending to all backends.
func (m *MultiNotifier) Send(ctx context.Context, event Event) error {
	var firstErr error
	for _, n := range m.notifiers {
		if err := n.Send(ctx, event); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			// Log but continue to other backends
		}
	}
	return firstErr
}

// Close closes all notification backends.
func (m *MultiNotifier) Close() error {
	var firstErr error
	for _, n := range m.notifiers {
		if err := n.Close(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
