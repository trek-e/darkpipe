// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package alert

import (
	"sync"
	"time"
)

// RateLimiter provides per-alert-type rate limiting with deduplication.
// Simple map-based implementation for single-process use case (no external dependencies).
type RateLimiter struct {
	mu           sync.Mutex
	lastSent     map[string]time.Time // alertType -> last sent timestamp
	dedupWindow  time.Duration        // Suppression window
	suppressions map[string]int       // alertType -> suppression count
}

// NewRateLimiter creates a new rate limiter with the specified deduplication window.
// Default: 1 hour per user decision.
func NewRateLimiter(dedupWindow time.Duration) *RateLimiter {
	if dedupWindow == 0 {
		dedupWindow = 1 * time.Hour
	}
	return &RateLimiter{
		lastSent:     make(map[string]time.Time),
		dedupWindow:  dedupWindow,
		suppressions: make(map[string]int),
	}
}

// ShouldSend returns true if an alert of this type should be sent.
// Returns false if the same alert type was sent within the deduplication window.
// Thread-safe.
func (r *RateLimiter) ShouldSend(alertType string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// Check if we've sent this alert type recently
	if lastTime, exists := r.lastSent[alertType]; exists {
		if now.Sub(lastTime) < r.dedupWindow {
			// Suppressed - increment counter
			r.suppressions[alertType]++
			return false
		}
	}

	// Update last sent timestamp
	r.lastSent[alertType] = now
	return true
}

// Reset clears the rate limit for a specific alert type.
// Useful for tests and manual override.
func (r *RateLimiter) Reset(alertType string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.lastSent, alertType)
	delete(r.suppressions, alertType)
}

// GetSuppressedCount returns suppression counts for all alert types.
// Used for observability (how many alerts were rate-limited).
func (r *RateLimiter) GetSuppressedCount() map[string]int {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Return a copy to avoid concurrent map access
	result := make(map[string]int)
	for k, v := range r.suppressions {
		result[k] = v
	}
	return result
}
