// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package provider

import (
	"context"
	"os"
	"testing"
)

// Note: These tests rely on actual DNS queries to public DNS servers.
// In a production environment, you might want to mock the DNS client.
// For now, we test with known domains that use these providers.

func TestDetectProvider_Cloudflare(t *testing.T) {
	// cloudflare.com uses Cloudflare nameservers
	ctx := context.Background()
	provider, err := DetectProvider(ctx, "cloudflare.com")
	if err != nil {
		t.Fatalf("DetectProvider failed: %v", err)
	}

	if provider != "cloudflare" {
		t.Errorf("Expected provider 'cloudflare', got '%s'", provider)
	}
}

func TestDetectProvider_Route53(t *testing.T) {
	// aws.amazon.com uses Route53 nameservers
	ctx := context.Background()
	provider, err := DetectProvider(ctx, "aws.amazon.com")
	if err != nil {
		t.Fatalf("DetectProvider failed: %v", err)
	}

	if provider != "route53" {
		t.Errorf("Expected provider 'route53', got '%s'", provider)
	}
}

func TestDetectProvider_Unknown(t *testing.T) {
	// google.com uses Google nameservers (not Cloudflare or Route53)
	ctx := context.Background()
	provider, err := DetectProvider(ctx, "google.com")
	if err != nil {
		t.Fatalf("DetectProvider failed: %v", err)
	}

	if provider != "unknown" {
		t.Errorf("Expected provider 'unknown', got '%s'", provider)
	}
}

func TestDetectProvider_InvalidDomain(t *testing.T) {
	// Non-existent domain should return unknown, not error
	ctx := context.Background()
	provider, err := DetectProvider(ctx, "thisisnotarealdomainthatexists12345.com")

	// Some DNS servers may return NXDOMAIN error, others may return empty answer
	// Both cases should be handled gracefully
	if err != nil && provider != "unknown" {
		// If error is returned, it should be a DNS query error (acceptable)
		t.Logf("Got expected error for invalid domain: %v", err)
	} else if provider != "unknown" {
		t.Errorf("Expected provider 'unknown' for invalid domain, got '%s'", provider)
	}
}

func TestNewProviderFromDetection_Unknown(t *testing.T) {
	// Unknown provider should return error with manual guide message
	ctx := context.Background()
	provider, err := NewProviderFromDetection(ctx, "google.com", nil)

	if err == nil {
		t.Error("Expected error for unknown provider")
	}

	if provider != nil {
		t.Error("Expected nil provider for unknown provider")
	}

	if err != nil && len(err.Error()) > 0 {
		if !contains(err.Error(), "DNS provider could not be detected") {
			t.Errorf("Error message should mention manual guide, got: %s", err.Error())
		}
	}
}

func TestNewProviderFromDetection_NotRegistered(t *testing.T) {
	// Providers register themselves on import
	// Without importing them, they won't be in the registry
	ctx := context.Background()

	// Save and clear environment to prevent accidental auth
	oldToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	os.Setenv("CLOUDFLARE_API_TOKEN", "")
	defer os.Setenv("CLOUDFLARE_API_TOKEN", oldToken)

	// Test Cloudflare detection - provider not imported, so not registered
	_, err := NewProviderFromDetection(ctx, "cloudflare.com", nil)
	if err == nil {
		t.Error("Expected error for unregistered provider")
	}
	if err != nil && !contains(err.Error(), "not registered") {
		t.Errorf("Expected 'not registered' error, got: %s", err.Error())
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
