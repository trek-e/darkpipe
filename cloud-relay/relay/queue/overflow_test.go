// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package queue

import (
	"context"
	"testing"
)

// ObjectStore interface defines the methods needed for overflow storage.
// This allows testing queue overflow logic without requiring a real S3 connection.
type ObjectStore interface {
	Upload(ctx context.Context, messageID string, data []byte) (string, error)
	Download(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]string, error)
}

// Verify OverflowStorage implements ObjectStore interface
var _ ObjectStore = (*OverflowStorage)(nil)

func TestNewOverflowStorage_InvalidEndpoint(t *testing.T) {
	// Test that invalid endpoint is handled gracefully
	_, err := NewOverflowStorage("invalid-endpoint-that-does-not-exist:9999", "key", "secret", "bucket", false)
	if err == nil {
		t.Error("expected error for invalid endpoint, got nil")
	}
}

func TestOverflowKeyGeneration(t *testing.T) {
	tests := []struct {
		name      string
		messageID string
		wantLen   int
	}{
		{
			name:      "Standard Message-ID",
			messageID: "<abc123@example.com>",
			wantLen:   79, // "darkpipe/queue/" (15) + 64 hex chars
		},
		{
			name:      "Message-ID with special characters",
			messageID: "<test+special@example.com>",
			wantLen:   79,
		},
		{
			name:      "Very long Message-ID",
			messageID: "<" + string(make([]byte, 1000)) + "@example.com>",
			wantLen:   79, // Hash always produces same length
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := sanitizeMessageIDToKey(tt.messageID)

			// Verify key format
			if len(key) != tt.wantLen {
				t.Errorf("key length = %d, want %d", len(key), tt.wantLen)
			}

			// Verify prefix
			if key[:15] != "darkpipe/queue/" {
				t.Errorf("key prefix = %q, want %q", key[:15], "darkpipe/queue/")
			}

			// Verify no special characters (only hex and slashes)
			for i, c := range key {
				if i < 15 {
					continue // Skip prefix
				}
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					t.Errorf("key contains non-hex character at position %d: %c", i, c)
				}
			}
		})
	}
}

func TestSanitizeForFilesystem(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Remove angle brackets",
			input: "<message@example.com>",
			want:  "message_example.com",
		},
		{
			name:  "Replace @ symbol",
			input: "user@domain.com",
			want:  "user_domain.com",
		},
		{
			name:  "Keep safe characters",
			input: "safe-file_name.123",
			want:  "safe-file_name.123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeForFilesystem(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeForFilesystem(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// Note: Full S3 integration tests (actual upload/download/delete) require a live S3/MinIO endpoint.
// These tests can be added with build tag `//go:build integration` for CI/CD environments
// that provide test S3 infrastructure.
