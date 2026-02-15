// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package forward

import (
	"bytes"
	"context"
	"io"

	"github.com/darkpipe/darkpipe/cloud-relay/relay/queue"
)

// QueuedForwarder wraps a transport forwarder and queues messages
// when the transport fails (home device offline).
type QueuedForwarder struct {
	transport Forwarder
	queue     *queue.MessageQueue
	enabled   bool
}

// NewQueuedForwarder creates a new queued forwarder.
func NewQueuedForwarder(transport Forwarder, q *queue.MessageQueue, enabled bool) *QueuedForwarder {
	return &QueuedForwarder{
		transport: transport,
		queue:     q,
		enabled:   enabled,
	}
}

// Forward attempts to forward the message via the transport.
// If transport fails and queuing is enabled, queues the message for later delivery.
// If transport fails and queuing is disabled, returns the error (triggers 4xx from Postfix).
func (f *QueuedForwarder) Forward(ctx context.Context, from string, to []string, data io.Reader) error {
	// Read data into buffer (io.Reader not seekable, need copy for potential queue)
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, data); err != nil {
		return err
	}

	// Attempt transport delivery
	if err := f.transport.Forward(ctx, from, to, bytes.NewReader(buf.Bytes())); err != nil {
		// Transport failed
		if !f.enabled {
			// Queuing disabled - return error to Postfix (triggers 4xx retry/bounce)
			return err
		}

		// Queuing enabled - enqueue the message
		if enqErr := f.queue.Enqueue(ctx, from, to, bytes.NewReader(buf.Bytes())); enqErr != nil {
			// Queue full - return original transport error
			return err
		}

		// Message queued successfully - return nil (message accepted)
		return nil
	}

	// Transport succeeded - message delivered
	return nil
}

// Close closes the transport and queue.
func (f *QueuedForwarder) Close() error {
	// Close transport first
	if err := f.transport.Close(); err != nil {
		// Still try to close queue even if transport close fails
		f.queue.Close()
		return err
	}

	// Close queue
	return f.queue.Close()
}

// Queue returns the underlying message queue (for processor startup in main.go).
func (f *QueuedForwarder) Queue() *queue.MessageQueue {
	return f.queue
}
