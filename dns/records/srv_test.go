package records

import (
	"strings"
	"testing"
)

func TestGenerateSRVRecords(t *testing.T) {
	domain := "example.com"
	mailHostname := "mail.example.com"

	records := GenerateSRVRecords(domain, mailHostname)

	// Should return 3 records (_imaps, _imap, _submission)
	if len(records) != 3 {
		t.Fatalf("Expected 3 SRV records, got %d", len(records))
	}

	// Test _imaps record
	imaps := records[0]
	if imaps.Service != "_imaps" {
		t.Errorf("Expected Service '_imaps', got '%s'", imaps.Service)
	}
	if imaps.Proto != "_tcp" {
		t.Errorf("Expected Proto '_tcp', got '%s'", imaps.Proto)
	}
	if imaps.Port != 993 {
		t.Errorf("Expected Port 993, got %d", imaps.Port)
	}
	if imaps.Priority != 0 {
		t.Errorf("Expected Priority 0 for preferred service, got %d", imaps.Priority)
	}
	if imaps.Weight != 1 {
		t.Errorf("Expected Weight 1, got %d", imaps.Weight)
	}
	if imaps.Target != mailHostname {
		t.Errorf("Expected Target '%s', got '%s'", mailHostname, imaps.Target)
	}

	// Test _imap record (unavailable)
	imap := records[1]
	if imap.Service != "_imap" {
		t.Errorf("Expected Service '_imap', got '%s'", imap.Service)
	}
	if imap.Port != 143 {
		t.Errorf("Expected Port 143, got %d", imap.Port)
	}
	if imap.Target != "." {
		t.Errorf("Expected Target '.' (unavailable), got '%s'", imap.Target)
	}
	if imap.Priority != 10 {
		t.Errorf("Expected Priority 10 for unavailable service, got %d", imap.Priority)
	}

	// Test _submission record
	submission := records[2]
	if submission.Service != "_submission" {
		t.Errorf("Expected Service '_submission', got '%s'", submission.Service)
	}
	if submission.Port != 587 {
		t.Errorf("Expected Port 587, got %d", submission.Port)
	}
	if submission.Priority != 0 {
		t.Errorf("Expected Priority 0 for preferred service, got %d", submission.Priority)
	}
	if submission.Target != mailHostname {
		t.Errorf("Expected Target '%s', got '%s'", mailHostname, submission.Target)
	}
}

func TestSRVRecordString(t *testing.T) {
	record := SRVRecord{
		Service:  "_imaps",
		Proto:    "_tcp",
		Domain:   "example.com",
		Priority: 0,
		Weight:   1,
		Port:     993,
		Target:   "mail.example.com",
	}

	result := record.String()

	// Expected format: _imaps._tcp.example.com. IN SRV 0 1 993 mail.example.com.
	expected := "_imaps._tcp.example.com. IN SRV 0 1 993 mail.example.com."
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Verify components are present
	if !strings.Contains(result, "_imaps._tcp.example.com.") {
		t.Error("Result should contain service and domain")
	}
	if !strings.Contains(result, "IN SRV") {
		t.Error("Result should contain 'IN SRV'")
	}
	if !strings.Contains(result, "0 1 993") {
		t.Error("Result should contain priority, weight, and port")
	}
	if !strings.Contains(result, "mail.example.com.") {
		t.Error("Result should contain target with trailing dot")
	}
}

func TestSRVRecordStringWithTrailingDot(t *testing.T) {
	// Test that target without trailing dot gets one added
	record := SRVRecord{
		Service:  "_submission",
		Proto:    "_tcp",
		Domain:   "example.org",
		Priority: 0,
		Weight:   1,
		Port:     587,
		Target:   "mail.example.org", // No trailing dot
	}

	result := record.String()

	if !strings.HasSuffix(result, "mail.example.org.") {
		t.Errorf("Target should have trailing dot, got: %s", result)
	}

	// Test that target with trailing dot is preserved
	record.Target = "mail.example.org."
	result = record.String()

	if !strings.HasSuffix(result, "mail.example.org.") {
		t.Errorf("Target with trailing dot should be preserved, got: %s", result)
	}

	// Should not have double dot
	if strings.HasSuffix(result, "..") {
		t.Error("Should not have double trailing dot")
	}
}

func TestGenerateAutodiscoverCNAME(t *testing.T) {
	domain := "example.com"
	relayHostname := "relay.example.net"

	records := GenerateAutodiscoverCNAME(domain, relayHostname)

	// Should return 2 CNAME records (autoconfig and autodiscover)
	if len(records) != 2 {
		t.Fatalf("Expected 2 CNAME records, got %d", len(records))
	}

	// Test autoconfig CNAME
	autoconfig := records[0]
	if autoconfig.Name != "autoconfig.example.com" {
		t.Errorf("Expected Name 'autoconfig.example.com', got '%s'", autoconfig.Name)
	}
	if autoconfig.Type != "CNAME" {
		t.Errorf("Expected Type 'CNAME', got '%s'", autoconfig.Type)
	}
	if autoconfig.Target != relayHostname {
		t.Errorf("Expected Target '%s', got '%s'", relayHostname, autoconfig.Target)
	}

	// Test autodiscover CNAME
	autodiscover := records[1]
	if autodiscover.Name != "autodiscover.example.com" {
		t.Errorf("Expected Name 'autodiscover.example.com', got '%s'", autodiscover.Name)
	}
	if autodiscover.Type != "CNAME" {
		t.Errorf("Expected Type 'CNAME', got '%s'", autodiscover.Type)
	}
	if autodiscover.Target != relayHostname {
		t.Errorf("Expected Target '%s', got '%s'", relayHostname, autodiscover.Target)
	}
}

func TestDNSRecordString(t *testing.T) {
	record := DNSRecord{
		Name:   "autoconfig.example.com",
		Type:   "CNAME",
		Target: "relay.example.net",
	}

	result := record.String()

	// Expected format: autoconfig.example.com. IN CNAME relay.example.net.
	expected := "autoconfig.example.com. IN CNAME relay.example.net."
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Verify trailing dots
	if !strings.Contains(result, "autoconfig.example.com.") {
		t.Error("Name should have trailing dot")
	}
	if !strings.HasSuffix(result, "relay.example.net.") {
		t.Error("Target should have trailing dot")
	}
}

func TestEnsureTrailingDot(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"example.com", "example.com."},
		{"example.com.", "example.com."},
		{"", ""},
		{".", "."},
		{"mail.example.org", "mail.example.org."},
	}

	for _, tt := range tests {
		result := ensureTrailingDot(tt.input)
		if result != tt.expected {
			t.Errorf("ensureTrailingDot(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
