// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package qrcode

import (
	"sync"
	"testing"
	"time"
)

func TestGenerateSecureToken(t *testing.T) {
	token1, err := GenerateSecureToken()
	if err != nil {
		t.Fatalf("GenerateSecureToken failed: %v", err)
	}

	if token1 == "" {
		t.Fatal("GenerateSecureToken returned empty string")
	}

	// Should be at least 43 characters (32 bytes base64url = 43 chars without padding)
	if len(token1) < 43 {
		t.Errorf("Token too short: got %d, want >= 43", len(token1))
	}

	// Generate second token to verify uniqueness
	token2, err := GenerateSecureToken()
	if err != nil {
		t.Fatalf("GenerateSecureToken failed on second call: %v", err)
	}

	if token1 == token2 {
		t.Error("GenerateSecureToken produced duplicate tokens")
	}
}

func TestMemoryTokenStoreCreate(t *testing.T) {
	store := NewMemoryTokenStore()
	email := "test@example.com"
	expiresAt := time.Now().Add(15 * time.Minute)

	token, err := store.Create(email, expiresAt)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if token == "" {
		t.Fatal("Create returned empty token")
	}

	// Verify token was stored
	store.mu.RLock()
	stored, exists := store.tokens[token]
	store.mu.RUnlock()

	if !exists {
		t.Fatal("Token not found in store")
	}

	if stored.Email != email {
		t.Errorf("Email mismatch: got %s, want %s", stored.Email, email)
	}

	if stored.Used {
		t.Error("New token should not be marked as used")
	}
}

func TestMemoryTokenStoreValidate(t *testing.T) {
	store := NewMemoryTokenStore()
	email := "test@example.com"
	expiresAt := time.Now().Add(15 * time.Minute)

	token, err := store.Create(email, expiresAt)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test valid token
	validEmail, valid, err := store.Validate(token)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if !valid {
		t.Error("Token should be valid")
	}

	if validEmail != email {
		t.Errorf("Email mismatch: got %s, want %s", validEmail, email)
	}

	// Test single-use enforcement: second validation should fail
	_, valid, err = store.Validate(token)
	if err != nil {
		t.Fatalf("Second Validate failed: %v", err)
	}

	if valid {
		t.Error("Token should be invalid after first use (single-use enforcement)")
	}
}

func TestMemoryTokenStoreValidateExpired(t *testing.T) {
	store := NewMemoryTokenStore()
	email := "test@example.com"
	expiresAt := time.Now().Add(-1 * time.Minute) // Already expired

	token, err := store.Create(email, expiresAt)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Validate expired token
	_, valid, err := store.Validate(token)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if valid {
		t.Error("Expired token should be invalid")
	}
}

func TestMemoryTokenStoreValidateNonExistent(t *testing.T) {
	store := NewMemoryTokenStore()

	_, valid, err := store.Validate("nonexistent-token")
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if valid {
		t.Error("Non-existent token should be invalid")
	}
}

func TestMemoryTokenStoreInvalidate(t *testing.T) {
	store := NewMemoryTokenStore()
	email := "test@example.com"
	expiresAt := time.Now().Add(15 * time.Minute)

	token, err := store.Create(email, expiresAt)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Invalidate token
	err = store.Invalidate(token)
	if err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}

	// Validate should fail after invalidation
	_, valid, err := store.Validate(token)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if valid {
		t.Error("Token should be invalid after Invalidate")
	}
}

func TestMemoryTokenStoreCleanup(t *testing.T) {
	store := NewMemoryTokenStore()

	// Create expired token
	expiredEmail := "expired@example.com"
	expiredToken, err := store.Create(expiredEmail, time.Now().Add(-1*time.Minute))
	if err != nil {
		t.Fatalf("Create expired token failed: %v", err)
	}

	// Create valid token
	validEmail := "valid@example.com"
	validToken, err := store.Create(validEmail, time.Now().Add(15*time.Minute))
	if err != nil {
		t.Fatalf("Create valid token failed: %v", err)
	}

	// Run cleanup
	store.Cleanup()

	// Expired token should be removed
	store.mu.RLock()
	_, expiredExists := store.tokens[expiredToken]
	_, validExists := store.tokens[validToken]
	store.mu.RUnlock()

	if expiredExists {
		t.Error("Cleanup did not remove expired token")
	}

	if !validExists {
		t.Error("Cleanup removed valid token")
	}
}

func TestMemoryTokenStoreConcurrency(t *testing.T) {
	store := NewMemoryTokenStore()
	email := "test@example.com"
	expiresAt := time.Now().Add(15 * time.Minute)

	var wg sync.WaitGroup
	numGoroutines := 10

	// Create tokens concurrently
	tokens := make([]string, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			token, err := store.Create(email, expiresAt)
			if err != nil {
				t.Errorf("Concurrent Create failed: %v", err)
				return
			}
			tokens[idx] = token
		}(i)
	}

	wg.Wait()

	// Validate tokens concurrently
	results := make([]bool, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, valid, err := store.Validate(tokens[idx])
			if err != nil {
				t.Errorf("Concurrent Validate failed: %v", err)
				return
			}
			results[idx] = valid
		}(i)
	}

	wg.Wait()

	// All validations should succeed (first use)
	for i, valid := range results {
		if !valid {
			t.Errorf("Token %d validation failed", i)
		}
	}
}
