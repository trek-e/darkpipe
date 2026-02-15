// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package cloudflare

import (
	"testing"

	"github.com/darkpipe/darkpipe/dns/provider"
)

// Test compile-time interface compliance
func TestClient_ImplementsDNSProvider(t *testing.T) {
	// This will fail at compile time if Client doesn't implement DNSProvider
	var _ provider.DNSProvider = (*Client)(nil)
}

func TestNewClient_RequiresAPIToken(t *testing.T) {
	// Test that empty API token returns error
	_, err := NewClient("")
	if err == nil {
		t.Error("Expected error when API token is empty")
	}

	expectedMsg := "CLOUDFLARE_API_TOKEN is required"
	if err != nil && err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestNewClient_Success(t *testing.T) {
	// Test that non-empty API token creates client
	client, err := NewClient("test-api-token")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if client == nil {
		t.Error("Expected non-nil client")
	}

	if client.Name() != "cloudflare" {
		t.Errorf("Expected provider name 'cloudflare', got '%s'", client.Name())
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"@", "@"},
		{"example.com", "example.com"},
		{"www.example.com", "example.com"},
		{"mail.example.com", "example.com"},
		{"_dmarc.example.com", "example.com"},
		{"sub.domain.example.com", "example.com"},
	}

	for _, tt := range tests {
		result := extractDomain(tt.input)
		if result != tt.expected {
			t.Errorf("extractDomain(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// Note: Full integration tests would require actual Cloudflare API credentials
// and a test domain. These tests verify the structure and interface compliance.
// Real API testing should be done in a separate integration test suite.

func TestClient_Name(t *testing.T) {
	client := &Client{}
	if client.Name() != "cloudflare" {
		t.Errorf("Expected name 'cloudflare', got '%s'", client.Name())
	}
}

// Mock-based tests for SPF duplicate detection would go here
// For now, we verify the logic structure is sound by reviewing the code
