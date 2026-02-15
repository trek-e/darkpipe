// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package alert

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimiter_ShouldSend_FirstCall(t *testing.T) {
	limiter := NewRateLimiter(1 * time.Hour)

	// First call should return true
	if !limiter.ShouldSend("cert_expiry") {
		t.Error("Expected first ShouldSend to return true")
	}
}

func TestRateLimiter_ShouldSend_WithinWindow(t *testing.T) {
	limiter := NewRateLimiter(100 * time.Millisecond)

	// First call succeeds
	if !limiter.ShouldSend("cert_expiry") {
		t.Fatal("Expected first ShouldSend to return true")
	}

	// Second call within window should be suppressed
	if limiter.ShouldSend("cert_expiry") {
		t.Error("Expected second ShouldSend to return false (within dedup window)")
	}

	// Verify suppression count incremented
	counts := limiter.GetSuppressedCount()
	if counts["cert_expiry"] != 1 {
		t.Errorf("Expected suppression count 1, got %d", counts["cert_expiry"])
	}
}

func TestRateLimiter_ShouldSend_AfterWindow(t *testing.T) {
	limiter := NewRateLimiter(50 * time.Millisecond)

	// First call
	if !limiter.ShouldSend("cert_expiry") {
		t.Fatal("Expected first ShouldSend to return true")
	}

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	// Second call after window should succeed
	if !limiter.ShouldSend("cert_expiry") {
		t.Error("Expected ShouldSend to return true after dedup window expires")
	}
}

func TestRateLimiter_IndependentTypes(t *testing.T) {
	limiter := NewRateLimiter(1 * time.Hour)

	// Send cert_expiry alert
	if !limiter.ShouldSend("cert_expiry") {
		t.Fatal("Expected first cert_expiry to return true")
	}

	// Send queue_backup alert (different type) - should not be rate-limited
	if !limiter.ShouldSend("queue_backup") {
		t.Error("Expected queue_backup to return true (different alert type)")
	}

	// Second cert_expiry should be suppressed
	if limiter.ShouldSend("cert_expiry") {
		t.Error("Expected second cert_expiry to be suppressed")
	}

	// Second queue_backup should be suppressed
	if limiter.ShouldSend("queue_backup") {
		t.Error("Expected second queue_backup to be suppressed")
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	limiter := NewRateLimiter(1 * time.Hour)

	// First call
	if !limiter.ShouldSend("cert_expiry") {
		t.Fatal("Expected first ShouldSend to return true")
	}

	// Second call should be suppressed
	if limiter.ShouldSend("cert_expiry") {
		t.Fatal("Expected second ShouldSend to be suppressed")
	}

	// Reset the rate limit
	limiter.Reset("cert_expiry")

	// Third call should succeed after reset
	if !limiter.ShouldSend("cert_expiry") {
		t.Error("Expected ShouldSend to return true after Reset")
	}

	// Verify suppression count cleared
	counts := limiter.GetSuppressedCount()
	if counts["cert_expiry"] != 0 {
		t.Errorf("Expected suppression count 0 after reset, got %d", counts["cert_expiry"])
	}
}

func TestRateLimiter_Concurrent(t *testing.T) {
	limiter := NewRateLimiter(10 * time.Millisecond)

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// Launch 100 concurrent ShouldSend calls
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.ShouldSend("cert_expiry") {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Only the first call should succeed (race detector will catch data races)
	if successCount != 1 {
		t.Errorf("Expected exactly 1 success from concurrent calls, got %d", successCount)
	}
}

func TestRateLimiter_GetSuppressedCount(t *testing.T) {
	limiter := NewRateLimiter(1 * time.Hour)

	// Send first alert (not suppressed)
	limiter.ShouldSend("cert_expiry")

	// Send 5 more alerts (all suppressed)
	for i := 0; i < 5; i++ {
		limiter.ShouldSend("cert_expiry")
	}

	// Send first queue_backup (not suppressed)
	limiter.ShouldSend("queue_backup")

	// Send 2 more queue_backup (suppressed)
	for i := 0; i < 2; i++ {
		limiter.ShouldSend("queue_backup")
	}

	counts := limiter.GetSuppressedCount()

	if counts["cert_expiry"] != 5 {
		t.Errorf("Expected cert_expiry suppression count 5, got %d", counts["cert_expiry"])
	}

	if counts["queue_backup"] != 2 {
		t.Errorf("Expected queue_backup suppression count 2, got %d", counts["queue_backup"])
	}
}

func TestRateLimiter_DefaultWindow(t *testing.T) {
	// Create limiter with zero duration (should use default 1 hour)
	limiter := NewRateLimiter(0)

	if limiter.dedupWindow != 1*time.Hour {
		t.Errorf("Expected default dedup window 1 hour, got %v", limiter.dedupWindow)
	}
}
