// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package forward

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
)

// ForwardCall records the arguments of a Forward call.
type ForwardCall struct {
	From string
	To   []string
	Data string // Message body as string for easier assertion
}

// MockForwarder implements the Forwarder interface for testing.
// It records all Forward calls and can be configured to return errors.
type MockForwarder struct {
	mu           sync.Mutex
	ForwardCalls []ForwardCall
	ForwardError error // If set, Forward will return this error
	CloseError   error // If set, Close will return this error
}

// NewMockForwarder creates a new mock forwarder for testing.
func NewMockForwarder() *MockForwarder {
	return &MockForwarder{
		ForwardCalls: []ForwardCall{},
	}
}

// Forward records the call and returns the configured error (if any).
func (m *MockForwarder) Forward(ctx context.Context, from string, to []string, data io.Reader) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Read data into buffer
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, data); err != nil {
		return fmt.Errorf("mock read data: %w", err)
	}

	// Record the call
	call := ForwardCall{
		From: from,
		To:   append([]string{}, to...), // Copy slice to avoid mutation
		Data: buf.String(),
	}
	m.ForwardCalls = append(m.ForwardCalls, call)

	return m.ForwardError
}

// Close returns the configured error (if any).
func (m *MockForwarder) Close() error {
	return m.CloseError
}

// GetCalls returns a copy of all recorded calls (thread-safe).
func (m *MockForwarder) GetCalls() []ForwardCall {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Return a copy to prevent race conditions
	calls := make([]ForwardCall, len(m.ForwardCalls))
	copy(calls, m.ForwardCalls)
	return calls
}

// Reset clears all recorded calls and errors.
func (m *MockForwarder) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ForwardCalls = []ForwardCall{}
	m.ForwardError = nil
	m.CloseError = nil
}
