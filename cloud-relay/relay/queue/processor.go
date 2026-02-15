// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package queue

import (
	"bytes"
	"context"
	"io"
	"log"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// Forwarder defines the interface for forwarding mail.
// This mirrors the interface in cloud-relay/relay/forward/forwarder.go
type Forwarder interface {
	Forward(ctx context.Context, from string, to []string, data io.Reader) error
	Close() error
}

// StartProcessor starts a background goroutine that drains the queue
// by attempting to deliver messages via the transport.
// It ticks at the specified interval and processes up to 10 messages per tick.
func (q *MessageQueue) StartProcessor(ctx context.Context, transport Forwarder, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Queue processor started (interval=%s)", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Queue processor stopped")
			return
		case <-ticker.C:
			// Purge expired messages first
			purged := q.PurgeExpired()
			if purged > 0 {
				log.Printf("Purged %d expired messages from queue", purged)
			}

			// Process queue if not empty
			if q.Len() > 0 {
				q.processQueue(ctx, transport)
			}
		}
	}
}

// processQueue attempts to deliver queued messages in FIFO order.
// Processes at most 10 messages per call to prevent thundering herd.
// Stops on first transport failure (home device likely still offline).
func (q *MessageQueue) processQueue(ctx context.Context, transport Forwarder) {
	const maxBatchSize = 10

	// Purge expired messages first (happens before getting batch)
	purged := q.PurgeExpired()
	if purged > 0 {
		log.Printf("Purged %d expired messages from queue during batch processing", purged)
	}

	q.mu.RLock()
	// Get up to 10 message IDs in FIFO order
	batch := make([]string, 0, maxBatchSize)
	for i := 0; i < len(q.order) && i < maxBatchSize; i++ {
		batch = append(batch, q.order[i])
	}
	q.mu.RUnlock()

	if len(batch) == 0 {
		return
	}

	log.Printf("Processing %d queued messages...", len(batch))

	delivered := 0
	for _, id := range batch {
		if err := q.processSingle(ctx, transport, id); err != nil {
			// Transport failure - stop processing (home device likely still offline)
			log.Printf("Queue delivery failed for %s: %v (stopping batch)", id, err)
			break
		}
		delivered++
	}

	if delivered > 0 {
		log.Printf("Delivered %d/%d queued messages", delivered, len(batch))
	}
}

// processSingle decrypts and delivers a single queued message.
func (q *MessageQueue) processSingle(ctx context.Context, transport Forwarder, id string) error {
	// Check if message is in overflow (need to delete from S3 after delivery)
	q.mu.RLock()
	msg, exists := q.messages[id]
	var inOverflow bool
	var overflowKey string
	if exists {
		inOverflow = msg.InOverflow
		overflowKey = msg.OverflowKey
	}
	q.mu.RUnlock()

	// Dequeue and decrypt
	msg, plaintext, err := q.Dequeue(id)
	if err != nil {
		// Message not found (might have been dequeued by another process)
		return nil
	}

	// Attempt delivery with exponential backoff
	deliveryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 1 * time.Second
	expBackoff.MaxInterval = 3 * time.Second
	expBackoff.MaxElapsedTime = 10 * time.Second

	retryFunc := func() error {
		return transport.Forward(deliveryCtx, msg.From, msg.To, bytes.NewReader(plaintext))
	}

	if err := backoff.Retry(retryFunc, backoff.WithContext(expBackoff, deliveryCtx)); err != nil {
		// Delivery failed - re-enqueue the message
		log.Printf("Failed to deliver queued message %s, re-enqueuing: %v", id, err)
		if enqErr := q.Enqueue(ctx, msg.From, msg.To, bytes.NewReader(plaintext)); enqErr != nil {
			log.Printf("ERROR: Failed to re-enqueue message %s: %v", id, enqErr)
		}
		return err
	}

	// Delivery succeeded - clean up overflow storage if needed
	if inOverflow && overflowKey != "" {
		q.mu.RLock()
		overflow := q.overflow
		q.mu.RUnlock()

		if overflow != nil {
			if err := overflow.Delete(ctx, overflowKey); err != nil {
				log.Printf("WARNING: Failed to delete overflow message %s from S3: %v", overflowKey, err)
			}
		}
	}

	return nil
}
