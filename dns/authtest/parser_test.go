// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package authtest

import (
	"strings"
	"testing"
)

func TestParseAuthResults_AllPass(t *testing.T) {
	// Gmail-style Authentication-Results header with all checks passing
	header := `mx.google.com;
       dkim=pass header.i=@example.com header.s=darkpipe-2026q1 header.b=ABC123;
       spf=pass (google.com: domain of sender@example.com designates 1.2.3.4 as permitted sender) smtp.mailfrom=sender@example.com;
       dmarc=pass (p=NONE sp=QUARANTINE dis=NONE) header.from=example.com`

	report, err := ParseAuthResults(header)
	if err != nil {
		t.Fatalf("Failed to parse Authentication-Results: %v", err)
	}

	// Check that all three methods are present
	if len(report.Results) != 3 {
		t.Errorf("Expected 3 results (DKIM, SPF, DMARC), got %d", len(report.Results))
	}

	// Verify convenience booleans
	if !report.SPFPass {
		t.Error("Expected SPFPass=true")
	}
	if !report.DKIMPass {
		t.Error("Expected DKIMPass=true")
	}
	if !report.DMARCPass {
		t.Error("Expected DMARCPass=true")
	}

	// Verify each result
	methods := make(map[string]AuthResult)
	for _, result := range report.Results {
		methods[result.Method] = result
	}

	// Check DKIM
	if dkim, ok := methods["dkim"]; ok {
		if dkim.Result != "pass" {
			t.Errorf("Expected DKIM result=pass, got %s", dkim.Result)
		}
	} else {
		t.Error("DKIM result not found")
	}

	// Check SPF
	if spf, ok := methods["spf"]; ok {
		if spf.Result != "pass" {
			t.Errorf("Expected SPF result=pass, got %s", spf.Result)
		}
	} else {
		t.Error("SPF result not found")
	}

	// Check DMARC
	if dmarc, ok := methods["dmarc"]; ok {
		if dmarc.Result != "pass" {
			t.Errorf("Expected DMARC result=pass, got %s", dmarc.Result)
		}
	} else {
		t.Error("DMARC result not found")
	}
}

func TestParseAuthResults_MixedResults(t *testing.T) {
	// SPF passes, DKIM fails, DMARC passes
	header := `mx.example.com;
       spf=pass smtp.mailfrom=sender@example.com;
       dkim=fail (signature verification failed) header.i=@example.com;
       dmarc=pass header.from=example.com`

	report, err := ParseAuthResults(header)
	if err != nil {
		t.Fatalf("Failed to parse Authentication-Results: %v", err)
	}

	if !report.SPFPass {
		t.Error("Expected SPFPass=true")
	}
	if report.DKIMPass {
		t.Error("Expected DKIMPass=false")
	}
	if !report.DMARCPass {
		t.Error("Expected DMARCPass=true")
	}

	// Verify DKIM fail result
	for _, result := range report.Results {
		if result.Method == "dkim" {
			if result.Result != "fail" {
				t.Errorf("Expected DKIM result=fail, got %s", result.Result)
			}
		}
	}
}

func TestParseAuthResults_OnlySPF(t *testing.T) {
	// Header with only SPF result
	header := `mx.example.com; spf=pass smtp.mailfrom=sender@example.com`

	report, err := ParseAuthResults(header)
	if err != nil {
		t.Fatalf("Failed to parse Authentication-Results: %v", err)
	}

	if len(report.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(report.Results))
	}

	if !report.SPFPass {
		t.Error("Expected SPFPass=true")
	}
	if report.DKIMPass {
		t.Error("Expected DKIMPass=false (DKIM not present)")
	}
	if report.DMARCPass {
		t.Error("Expected DMARCPass=false (DMARC not present)")
	}
}

func TestParseAuthResults_OnlyDKIM(t *testing.T) {
	// Header with only DKIM result
	header := `mx.example.com; dkim=pass header.i=@example.com header.s=selector1`

	report, err := ParseAuthResults(header)
	if err != nil {
		t.Fatalf("Failed to parse Authentication-Results: %v", err)
	}

	if len(report.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(report.Results))
	}

	if report.SPFPass {
		t.Error("Expected SPFPass=false (SPF not present)")
	}
	if !report.DKIMPass {
		t.Error("Expected DKIMPass=true")
	}
	if report.DMARCPass {
		t.Error("Expected DMARCPass=false (DMARC not present)")
	}
}

func TestParseAuthResults_AllFail(t *testing.T) {
	// All checks fail
	header := `mx.example.com;
       spf=fail smtp.mailfrom=sender@example.com;
       dkim=fail header.i=@example.com;
       dmarc=fail header.from=example.com`

	report, err := ParseAuthResults(header)
	if err != nil {
		t.Fatalf("Failed to parse Authentication-Results: %v", err)
	}

	if report.SPFPass {
		t.Error("Expected SPFPass=false")
	}
	if report.DKIMPass {
		t.Error("Expected DKIMPass=false")
	}
	if report.DMARCPass {
		t.Error("Expected DMARCPass=false")
	}
}

func TestParseAuthResults_NoneResults(t *testing.T) {
	// Authentication not performed (none)
	header := `mx.example.com;
       spf=none smtp.mailfrom=sender@example.com;
       dkim=none;
       dmarc=none`

	report, err := ParseAuthResults(header)
	if err != nil {
		t.Fatalf("Failed to parse Authentication-Results: %v", err)
	}

	// "none" means not authenticated, so should be false
	if report.SPFPass {
		t.Error("Expected SPFPass=false for 'none' result")
	}
	if report.DKIMPass {
		t.Error("Expected DKIMPass=false for 'none' result")
	}
	if report.DMARCPass {
		t.Error("Expected DMARCPass=false for 'none' result")
	}

	// Verify results are captured
	if len(report.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(report.Results))
	}

	for _, result := range report.Results {
		if result.Result != "none" {
			t.Errorf("Expected result=none, got %s for %s", result.Result, result.Method)
		}
	}
}

func TestParseAuthResults_Malformed(t *testing.T) {
	// Test with malformed header
	header := "this is not a valid authentication results header"

	_, err := ParseAuthResults(header)
	if err == nil {
		t.Error("Expected error for malformed header")
	}
}

func TestParseAuthResults_Empty(t *testing.T) {
	// Test with empty header
	header := ""

	report, err := ParseAuthResults(header)
	// Empty header may or may not error depending on the parser implementation
	// If it doesn't error, it should return an empty report
	if err == nil {
		if len(report.Results) != 0 {
			t.Error("Expected empty results for empty header")
		}
	}
}

func TestParseAuthResults_RealGmailExample(t *testing.T) {
	// Real-world Gmail Authentication-Results header format
	header := `mx.google.com;
       dkim=pass header.i=@darkpipe.com header.s=darkpipe-2026q1 header.b=K8sY9XZm;
       spf=pass (google.com: domain of user@darkpipe.com designates 203.0.113.42 as permitted sender) smtp.mailfrom=user@darkpipe.com;
       dmarc=pass (p=NONE sp=QUARANTINE dis=NONE) header.from=darkpipe.com`

	report, err := ParseAuthResults(header)
	if err != nil {
		t.Fatalf("Failed to parse real Gmail header: %v", err)
	}

	if !report.SPFPass || !report.DKIMPass || !report.DMARCPass {
		t.Errorf("Expected all checks to pass. SPF=%v, DKIM=%v, DMARC=%v",
			report.SPFPass, report.DKIMPass, report.DMARCPass)
	}

	// Verify raw header is stored
	if report.RawHeader != header {
		t.Error("RawHeader not stored correctly")
	}
}

func TestParseAuthResults_WithDetails(t *testing.T) {
	// Header with detailed properties
	header := `mx.example.com;
       dkim=pass header.i=@example.com header.s=selector1 header.d=example.com;
       spf=pass smtp.mailfrom=sender@example.com smtp.helo=relay.example.com`

	report, err := ParseAuthResults(header)
	if err != nil {
		t.Fatalf("Failed to parse Authentication-Results: %v", err)
	}

	// Check that details are extracted
	for _, result := range report.Results {
		if result.Method == "dkim" {
			// Should have some details about selector or domain
			if result.Details == "" {
				t.Error("Expected DKIM details to be non-empty")
			}
		}
	}
}

func TestAuthResult_Structure(t *testing.T) {
	// Test the basic structure of AuthResult
	result := AuthResult{
		Method:  "dkim",
		Result:  "pass",
		Details: "header.s=selector1; header.d=example.com",
	}

	if result.Method != "dkim" {
		t.Errorf("Expected Method=dkim, got %s", result.Method)
	}
	if result.Result != "pass" {
		t.Errorf("Expected Result=pass, got %s", result.Result)
	}
	if result.Details == "" {
		t.Error("Expected non-empty Details")
	}
}

func TestAuthReport_Structure(t *testing.T) {
	// Test the basic structure of AuthReport
	report := AuthReport{
		Results: []AuthResult{
			{Method: "spf", Result: "pass"},
			{Method: "dkim", Result: "pass"},
			{Method: "dmarc", Result: "pass"},
		},
		SPFPass:   true,
		DKIMPass:  true,
		DMARCPass: true,
		RawHeader: "test header",
	}

	if len(report.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(report.Results))
	}
	if !report.SPFPass || !report.DKIMPass || !report.DMARCPass {
		t.Error("Expected all pass flags to be true")
	}
	if report.RawHeader != "test header" {
		t.Error("RawHeader not set correctly")
	}
}

func TestDisplayAuthReport(t *testing.T) {
	// Test the display function
	report := AuthReport{
		Results: []AuthResult{
			{Method: "spf", Result: "pass", Details: "smtp.mailfrom=sender@example.com"},
			{Method: "dkim", Result: "pass", Details: "header.s=selector1"},
			{Method: "dmarc", Result: "pass", Details: "p=NONE"},
		},
		SPFPass:   true,
		DKIMPass:  true,
		DMARCPass: true,
	}

	var buf strings.Builder
	DisplayAuthReport(report, &buf)

	output := buf.String()

	// Check that output contains expected elements
	if !strings.Contains(output, "SPF") {
		t.Error("Output should contain SPF")
	}
	if !strings.Contains(output, "DKIM") {
		t.Error("Output should contain DKIM")
	}
	if !strings.Contains(output, "DMARC") {
		t.Error("Output should contain DMARC")
	}
	if !strings.Contains(output, "PASS") {
		t.Error("Output should contain PASS")
	}
	if !strings.Contains(output, "All authentication checks passed") {
		t.Error("Output should contain success message")
	}
}

func TestDisplayAuthReport_WithFailures(t *testing.T) {
	// Test display with failures
	report := AuthReport{
		Results: []AuthResult{
			{Method: "spf", Result: "fail"},
			{Method: "dkim", Result: "pass"},
			{Method: "dmarc", Result: "fail"},
		},
		SPFPass:   false,
		DKIMPass:  true,
		DMARCPass: false,
	}

	var buf strings.Builder
	DisplayAuthReport(report, &buf)

	output := buf.String()

	// Check for failure indicators
	if !strings.Contains(output, "FAIL") {
		t.Error("Output should contain FAIL")
	}
	if !strings.Contains(output, "darkpipe dns-setup --validate-only") {
		t.Error("Output should contain validation command suggestion")
	}
	if strings.Contains(output, "All authentication checks passed") {
		t.Error("Output should NOT contain success message when there are failures")
	}
}
