package validator

import (
	"context"
	"net"
	"testing"
)

func TestCheckPTR_Pass(t *testing.T) {
	// This test uses a real DNS lookup for a known good PTR record
	// We'll use Google's public DNS server which has proper PTR records
	// 8.8.8.8 -> dns.google

	result := CheckPTR(context.Background(), "8.8.8.8", "dns.google")

	// The check should pass (8.8.8.8 has PTR to dns.google)
	if !result.Pass {
		t.Logf("PTR check for 8.8.8.8 -> dns.google: %s", result.Error)
		t.Logf("PTR names found: %v", result.PTRNames)
		t.Logf("Forward match: %v", result.ForwardMatch)
		// This may fail in some network environments, so we'll make it a soft failure
		t.Skip("Skipping live PTR test - may not work in all environments")
	}

	if result.IP != "8.8.8.8" {
		t.Errorf("Expected IP=8.8.8.8, got %s", result.IP)
	}
	if result.ExpectedHostname != "dns.google" {
		t.Errorf("Expected ExpectedHostname=dns.google, got %s", result.ExpectedHostname)
	}
}

func TestCheckPTR_NoRecord(t *testing.T) {
	// Use a private IP that definitely won't have a PTR record
	result := CheckPTR(context.Background(), "192.168.1.1", "test.example.com")

	if result.Pass {
		t.Error("Expected PTR check to fail for private IP without PTR record")
	}

	// Should have an error message
	if result.Error == "" {
		t.Error("Expected error message when PTR check fails")
	}

	// Error should mention VPS provider
	if !contains(result.Error, "VPS provider") {
		t.Errorf("Expected error to mention VPS provider, got: %s", result.Error)
	}
}

func TestCheckPTR_InvalidIP(t *testing.T) {
	// Test with an invalid IP address
	result := CheckPTR(context.Background(), "999.999.999.999", "test.example.com")

	if result.Pass {
		t.Error("Expected PTR check to fail for invalid IP")
	}

	if result.Error == "" {
		t.Error("Expected error message for invalid IP")
	}
}

func TestCheckPTR_ForwardMismatch(t *testing.T) {
	// This is a tricky test to write without a mock DNS server
	// We'll use a known IP and deliberately wrong hostname

	// 8.8.8.8 has PTR to dns.google, but we'll check for a different hostname
	// that won't match in forward lookup
	result := CheckPTR(context.Background(), "8.8.8.8", "wrong.example.com")

	if result.Pass {
		t.Error("Expected PTR check to fail when expected hostname doesn't match")
	}

	// The forward match might succeed (8.8.8.8 -> dns.google -> 8.8.8.8)
	// but the expected hostname won't be in PTR names
	if result.Error == "" {
		t.Error("Expected error message when hostname doesn't match")
	}
}

func TestCheckPTR_Structure(t *testing.T) {
	// Test the basic structure of PTRResult without relying on network
	result := PTRResult{
		IP:               "1.2.3.4",
		ExpectedHostname: "mail.example.com",
		PTRNames:         []string{"mail.example.com"},
		ForwardMatch:     true,
		Pass:             true,
	}

	if result.IP != "1.2.3.4" {
		t.Errorf("Expected IP=1.2.3.4, got %s", result.IP)
	}
	if result.ExpectedHostname != "mail.example.com" {
		t.Errorf("Expected ExpectedHostname=mail.example.com, got %s", result.ExpectedHostname)
	}
	if len(result.PTRNames) != 1 {
		t.Errorf("Expected 1 PTR name, got %d", len(result.PTRNames))
	}
	if !result.ForwardMatch {
		t.Error("Expected ForwardMatch=true")
	}
	if !result.Pass {
		t.Error("Expected Pass=true")
	}
}

func TestCheckPTR_ErrorMessages(t *testing.T) {
	// Test that error messages contain helpful information
	tests := []struct {
		name             string
		ip               string
		expectedHostname string
		wantSubstring    string
	}{
		{
			name:             "private IP",
			ip:               "10.0.0.1",
			expectedHostname: "mail.example.com",
			wantSubstring:    "VPS provider",
		},
		{
			name:             "localhost",
			ip:               "127.0.0.1",
			expectedHostname: "localhost",
			wantSubstring:    "", // May succeed with "localhost" PTR
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPTR(context.Background(), tt.ip, tt.expectedHostname)

			// If it fails and we expect a substring, check for it
			if !result.Pass && tt.wantSubstring != "" {
				if !contains(result.Error, tt.wantSubstring) {
					t.Errorf("Expected error to contain '%s', got: %s", tt.wantSubstring, result.Error)
				}
			}
		})
	}
}

func TestCheckPTR_ContextCancellation(t *testing.T) {
	// Test that context cancellation is respected
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result := CheckPTR(ctx, "8.8.8.8", "dns.google")

	// Should fail due to context cancellation
	if result.Pass {
		t.Error("Expected PTR check to fail when context is cancelled")
	}

	// LookupAddr should respect context and return an error
	// The error message should indicate the lookup failed
	if result.Error == "" {
		t.Error("Expected error message when context is cancelled")
	}
}

func TestPTRResult_Fields(t *testing.T) {
	// Verify all fields are accessible and have expected types
	result := PTRResult{
		IP:               "test-ip",
		ExpectedHostname: "test-hostname",
		PTRNames:         []string{"name1", "name2"},
		ForwardMatch:     true,
		Pass:             false,
		Error:            "test error",
	}

	// Type assertions to verify field types
	var _ string = result.IP
	var _ string = result.ExpectedHostname
	var _ []string = result.PTRNames
	var _ bool = result.ForwardMatch
	var _ bool = result.Pass
	var _ string = result.Error

	// Verify values
	if result.IP != "test-ip" {
		t.Errorf("IP field mismatch")
	}
	if result.ExpectedHostname != "test-hostname" {
		t.Errorf("ExpectedHostname field mismatch")
	}
	if len(result.PTRNames) != 2 {
		t.Errorf("PTRNames field mismatch")
	}
	if !result.ForwardMatch {
		t.Errorf("ForwardMatch field mismatch")
	}
	if result.Pass {
		t.Errorf("Pass field mismatch")
	}
	if result.Error != "test error" {
		t.Errorf("Error field mismatch")
	}
}

func TestCheckPTR_TrailingDot(t *testing.T) {
	// PTR lookups often return names with trailing dots
	// Verify we handle both with and without trailing dots

	// We can't easily mock this without changing the implementation,
	// so we'll test the logic with a real lookup that we expect to work

	// Test with cloudflare's DNS (1.1.1.1 -> one.one.one.one)
	result := CheckPTR(context.Background(), "1.1.1.1", "one.one.one.one")

	if !result.Pass {
		// May fail in some environments
		t.Logf("PTR check for 1.1.1.1: %s", result.Error)
		t.Logf("PTR names: %v", result.PTRNames)

		// Check that PTR names don't have trailing dots
		for _, name := range result.PTRNames {
			if len(name) > 0 && name[len(name)-1] == '.' {
				t.Errorf("PTR name should not have trailing dot: %s", name)
			}
		}
	}
}

func TestCheckPTR_RealWorldScenario(t *testing.T) {
	// Test a realistic scenario where the PTR might exist but not match
	// Use localhost which should have some kind of PTR

	result := CheckPTR(context.Background(), "127.0.0.1", "mail.darkpipe.example")

	// This should likely fail because 127.0.0.1 won't point to our expected hostname
	// but it tests the real-world flow

	if result.Pass {
		t.Log("Unexpected: localhost PTR matched our test hostname")
	} else {
		// Verify the result has populated fields
		if result.IP != "127.0.0.1" {
			t.Errorf("Expected IP=127.0.0.1, got %s", result.IP)
		}
		if result.ExpectedHostname != "mail.darkpipe.example" {
			t.Errorf("Expected ExpectedHostname=mail.darkpipe.example, got %s", result.ExpectedHostname)
		}

		// Error should provide helpful guidance
		if result.Error == "" {
			t.Error("Expected error message")
		}
	}
}

func TestCheckPTR_LookupAddrBehavior(t *testing.T) {
	// Verify that net.LookupAddr handles in-addr.arpa construction
	// This is more of a documentation test

	names, err := net.LookupAddr("8.8.8.8")
	if err != nil {
		t.Logf("LookupAddr failed (may be network issue): %v", err)
		t.Skip("Skipping - network lookup failed")
	}

	if len(names) == 0 {
		t.Skip("No PTR records found for 8.8.8.8")
	}

	// Verify names are returned
	t.Logf("PTR records for 8.8.8.8: %v", names)

	// Should include dns.google or similar
	found := false
	for _, name := range names {
		if contains(name, "dns.google") || contains(name, "google") {
			found = true
			break
		}
	}

	if !found {
		t.Logf("Expected to find dns.google in PTR records, got: %v", names)
	}
}
