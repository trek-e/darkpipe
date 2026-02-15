// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package qrcode

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

// TokenStore manages single-use QR code tokens with expiry.
type TokenStore interface {
	// Create generates a new single-use token for the given email.
	Create(email string, expiresAt time.Time) (token string, err error)

	// Validate checks if a token is valid and returns the associated email.
	// Tokens are single-use: on successful validation, the token is immediately marked as used.
	Validate(token string) (email string, valid bool, err error)

	// Invalidate marks a token as used (for explicit revocation).
	Invalidate(token string) error

	// Cleanup removes expired tokens. Should be called periodically.
	Cleanup()
}

// Token represents a single-use QR code token.
type Token struct {
	Email     string
	Token     string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

// MemoryTokenStore is an in-memory implementation of TokenStore with thread safety.
type MemoryTokenStore struct {
	mu     sync.RWMutex
	tokens map[string]*Token
}

// NewMemoryTokenStore creates a new in-memory token store.
func NewMemoryTokenStore() *MemoryTokenStore {
	store := &MemoryTokenStore{
		tokens: make(map[string]*Token),
	}

	// Start cleanup goroutine (runs every 5 minutes)
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			store.Cleanup()
		}
	}()

	return store
}

// Create generates a new single-use token for the given email.
func (s *MemoryTokenStore) Create(email string, expiresAt time.Time) (string, error) {
	token, err := GenerateSecureToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	t := &Token{
		Email:     email,
		Token:     token,
		ExpiresAt: expiresAt,
		Used:      false,
		CreatedAt: time.Now(),
	}

	s.mu.Lock()
	s.tokens[token] = t
	s.mu.Unlock()

	return token, nil
}

// Validate checks if a token is valid and returns the associated email.
// On successful validation, the token is immediately marked as used (single-use enforcement).
func (s *MemoryTokenStore) Validate(token string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, exists := s.tokens[token]
	if !exists {
		return "", false, nil
	}

	// Check if already used
	if t.Used {
		return "", false, nil
	}

	// Check if expired
	if time.Now().After(t.ExpiresAt) {
		return "", false, nil
	}

	// Mark as used IMMEDIATELY to prevent race conditions
	t.Used = true

	return t.Email, true, nil
}

// Invalidate marks a token as used (for explicit revocation).
func (s *MemoryTokenStore) Invalidate(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, exists := s.tokens[token]
	if !exists {
		return fmt.Errorf("token not found")
	}

	t.Used = true
	return nil
}

// Cleanup removes expired tokens from the store.
func (s *MemoryTokenStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for token, t := range s.tokens {
		if now.After(t.ExpiresAt) {
			delete(s.tokens, token)
		}
	}
}

// GenerateSecureToken generates a cryptographically secure token.
// Uses 32 bytes from crypto/rand, base64url-encoded (no padding) for 256-bit entropy per NIST recommendations.
func GenerateSecureToken() (string, error) {
	b := make([]byte, 32) // 256 bits
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("crypto/rand read: %w", err)
	}

	// Use base64 URL encoding without padding (URL-safe)
	return base64.RawURLEncoding.EncodeToString(b), nil
}
