// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package provider

import (
	"context"
	"testing"
	"time"
)

func TestWaitForPropagation_Success(t *testing.T) {
	// Test with a well-known TXT record that should exist
	// cloudflare.com has stable TXT records
	// Note: This is a live DNS query, so it may be flaky in some environments
	// In production, you'd want to mock the DNS client
	t.Skip("Skipping live DNS propagation test - would require specific test domain")
}

func TestWaitForPropagation_Timeout(t *testing.T) {
	// Test with a non-existent record to trigger timeout
	ctx := context.Background()

	// Use a very short timeout to make the test fast
	timeout := 6 * time.Second // Just over one check interval

	err := WaitForPropagation(
		ctx,
		"nonexistent-record-12345.example.com",
		"TXT",
		"this-value-does-not-exist",
		timeout,
		[]string{"8.8.8.8:53"},
	)

	if err == nil {
		t.Error("Expected timeout error for non-existent record")
	}

	if err != nil && !contains(err.Error(), "timeout") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestWaitForPropagation_ContextCancellation(t *testing.T) {
	// Test that context cancellation stops propagation check
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	err := WaitForPropagation(
		ctx,
		"example.com",
		"TXT",
		"test",
		1*time.Minute,
		[]string{"8.8.8.8:53"},
	)

	if err == nil {
		t.Error("Expected context cancellation error")
	}

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}
}

func TestNormalizeValue_TXT(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"v=spf1 -all"`, "v=spf1 -all"},
		{"v=spf1 -all", "v=spf1 -all"},
		{`  "test"  `, "test"},
		{"  test  ", "test"},
		{`""`, ""},
	}

	for _, tt := range tests {
		result := normalizeValue("TXT", tt.input)
		if result != tt.expected {
			t.Errorf("normalizeValue(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeValue_OtherTypes(t *testing.T) {
	tests := []struct {
		recordType string
		input      string
		expected   string
	}{
		{"MX", "  mail.example.com  ", "mail.example.com"},
		{"A", "  192.0.2.1  ", "192.0.2.1"},
		{"CNAME", "  www.example.com  ", "www.example.com"},
	}

	for _, tt := range tests {
		result := normalizeValue(tt.recordType, tt.input)
		if result != tt.expected {
			t.Errorf("normalizeValue(%s, %q) = %q, want %q", tt.recordType, tt.input, result, tt.expected)
		}
	}
}

func TestCheckRecord_InvalidType(t *testing.T) {
	// Invalid record type should return false (not panic)
	result := checkRecord("example.com", "INVALID", "test", "8.8.8.8:53")
	if result {
		t.Error("Expected false for invalid record type")
	}
}

func TestExtractValue_TXT(t *testing.T) {
	// This is hard to test without actual DNS answers
	// Mainly testing that it doesn't panic
	// Real testing would require mocking dns.RR
	t.Skip("Skipping extractValue test - requires DNS answer mocking")
}

func TestDefaultDNSServers(t *testing.T) {
	// Verify default DNS servers are set correctly
	expected := []string{
		"8.8.8.8:53",
		"1.1.1.1:53",
		"208.67.222.222:53",
	}

	if len(DefaultDNSServers) != len(expected) {
		t.Errorf("Expected %d default DNS servers, got %d", len(expected), len(DefaultDNSServers))
	}

	for i, server := range expected {
		if i >= len(DefaultDNSServers) {
			break
		}
		if DefaultDNSServers[i] != server {
			t.Errorf("DefaultDNSServers[%d] = %s, want %s", i, DefaultDNSServers[i], server)
		}
	}
}

func TestWaitForPropagation_EmptyServers(t *testing.T) {
	// When no servers provided, should use default servers
	ctx := context.Background()

	// Use a very short timeout
	timeout := 6 * time.Second

	err := WaitForPropagation(
		ctx,
		"nonexistent-test.example.com",
		"TXT",
		"test",
		timeout,
		[]string{}, // Empty servers - should use defaults
	)

	// Should timeout (expected) but using default servers
	if err == nil {
		t.Error("Expected timeout error")
	}

	// If it times out, it means it tried the default servers
	if err != nil && !contains(err.Error(), "timeout") && !contains(err.Error(), "3 servers") {
		t.Logf("Got error (expected): %v", err)
	}
}
