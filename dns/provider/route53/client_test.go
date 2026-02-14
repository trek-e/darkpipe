package route53

import (
	"testing"

	"github.com/darkpipe/darkpipe/dns/provider"
)

// Test compile-time interface compliance
func TestClient_ImplementsDNSProvider(t *testing.T) {
	// This will fail at compile time if Client doesn't implement DNSProvider
	var _ provider.DNSProvider = (*Client)(nil)
}

func TestClient_Name(t *testing.T) {
	client := &Client{}
	if client.Name() != "route53" {
		t.Errorf("Expected name 'route53', got '%s'", client.Name())
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

func TestBuildResourceRecord_TXT(t *testing.T) {
	// Test TXT record quoting
	rec := provider.Record{
		Type:    "TXT",
		Content: "v=spf1 -all",
		TTL:     3600,
	}

	resourceRecords := buildResourceRecord(rec)
	if len(resourceRecords) != 1 {
		t.Errorf("Expected 1 resource record, got %d", len(resourceRecords))
	}

	expected := "\"v=spf1 -all\""
	if *resourceRecords[0].Value != expected {
		t.Errorf("Expected TXT value %q, got %q", expected, *resourceRecords[0].Value)
	}
}

func TestBuildResourceRecord_MX(t *testing.T) {
	// Test MX record with priority
	priority := 10
	rec := provider.Record{
		Type:     "MX",
		Content:  "mail.example.com",
		TTL:      3600,
		Priority: &priority,
	}

	resourceRecords := buildResourceRecord(rec)
	if len(resourceRecords) != 1 {
		t.Errorf("Expected 1 resource record, got %d", len(resourceRecords))
	}

	expected := "10 mail.example.com"
	if *resourceRecords[0].Value != expected {
		t.Errorf("Expected MX value %q, got %q", expected, *resourceRecords[0].Value)
	}
}

func TestExtractContent_TXT(t *testing.T) {
	// Test TXT record quote removal
	value := "\"v=spf1 -all\""
	result := extractContent(value, "TXT")
	expected := "v=spf1 -all"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestExtractContent_MX(t *testing.T) {
	// Test MX record priority removal
	value := "10 mail.example.com"
	result := extractContent(value, "MX")
	expected := "mail.example.com"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestExtractContent_A(t *testing.T) {
	// Test A record (no transformation)
	value := "192.0.2.1"
	result := extractContent(value, "A")
	if result != value {
		t.Errorf("Expected %q, got %q", value, result)
	}
}

// Note: Full integration tests would require actual AWS credentials
// and a test hosted zone. These tests verify the structure and interface compliance.
// Real API testing should be done in a separate integration test suite.
