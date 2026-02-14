package records

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestGenerateGuide(t *testing.T) {
	records := AllRecords{
		Domain: "example.com",
		SPF: SPFRecord{
			Domain: "example.com",
			Value:  "v=spf1 ip4:203.0.113.10 -all",
		},
		DKIM: DKIMRecord{
			Domain:   "darkpipe-2026q1._domainkey.example.com",
			Selector: "darkpipe-2026q1",
			Value:    "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A...",
		},
		DMARC: DMARCRecord{
			Domain: "_dmarc.example.com",
			Value:  "v=DMARC1; p=none; sp=quarantine; rua=mailto:dmarc@example.com",
		},
		MX: MXRecord{
			Domain:   "example.com",
			Hostname: "relay.darkpipe.com",
			Priority: 10,
		},
	}

	guide := GenerateGuide(records)

	// Verify guide contains all record types
	requiredSections := []string{
		"SPF Record",
		"DKIM Record",
		"DMARC Record",
		"MX Record",
	}

	for _, section := range requiredSections {
		if !strings.Contains(guide, section) {
			t.Fatalf("Guide does not contain section: %s", section)
		}
	}

	// Verify record values are present
	if !strings.Contains(guide, records.SPF.Value) {
		t.Fatal("Guide does not contain SPF value")
	}
	if !strings.Contains(guide, records.DKIM.Value) {
		t.Fatal("Guide does not contain DKIM value")
	}
	if !strings.Contains(guide, records.DMARC.Value) {
		t.Fatal("Guide does not contain DMARC value")
	}
	if !strings.Contains(guide, records.MX.Hostname) {
		t.Fatal("Guide does not contain MX hostname")
	}

	// Verify provider documentation links
	providers := []string{
		"Cloudflare",
		"Route53",
		"GoDaddy",
		"Namecheap",
	}

	for _, provider := range providers {
		if !strings.Contains(guide, provider) {
			t.Fatalf("Guide does not contain provider documentation for: %s", provider)
		}
	}

	// Verify verification section
	if !strings.Contains(guide, "Verification") {
		t.Fatal("Guide does not contain verification section")
	}
	if !strings.Contains(guide, "MXToolbox") {
		t.Fatal("Guide does not contain MXToolbox link")
	}
	if !strings.Contains(guide, "mail-tester.com") {
		t.Fatal("Guide does not contain mail-tester.com link")
	}
	if !strings.Contains(guide, "darkpipe dns-setup --validate-only") {
		t.Fatal("Guide does not contain built-in validation command")
	}
}

func TestPrintRecords(t *testing.T) {
	records := AllRecords{
		Domain: "example.com",
		SPF: SPFRecord{
			Domain: "example.com",
			Value:  "v=spf1 ip4:203.0.113.10 -all",
		},
		DKIM: DKIMRecord{
			Domain:   "darkpipe-2026q1._domainkey.example.com",
			Selector: "darkpipe-2026q1",
			Value:    "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A...",
		},
		DMARC: DMARCRecord{
			Domain: "_dmarc.example.com",
			Value:  "v=DMARC1; p=none; sp=quarantine",
		},
		MX: MXRecord{
			Domain:   "example.com",
			Hostname: "relay.darkpipe.com",
			Priority: 10,
		},
	}

	var buf bytes.Buffer
	PrintRecords(&buf, records, false) // No color for testing

	output := buf.String()

	// Verify output contains record information
	if !strings.Contains(output, "SPF Record") {
		t.Fatal("Terminal output does not contain SPF Record")
	}
	if !strings.Contains(output, "DKIM Record") {
		t.Fatal("Terminal output does not contain DKIM Record")
	}
	if !strings.Contains(output, "DMARC Record") {
		t.Fatal("Terminal output does not contain DMARC Record")
	}
	if !strings.Contains(output, "MX Record") {
		t.Fatal("Terminal output does not contain MX Record")
	}

	// Verify next steps are included
	if !strings.Contains(output, "Next Steps") {
		t.Fatal("Terminal output does not contain Next Steps section")
	}
}

func TestPrintJSON(t *testing.T) {
	records := AllRecords{
		Domain: "example.com",
		SPF: SPFRecord{
			Domain: "example.com",
			Value:  "v=spf1 ip4:203.0.113.10 -all",
		},
		DKIM: DKIMRecord{
			Domain:   "darkpipe-2026q1._domainkey.example.com",
			Selector: "darkpipe-2026q1",
			Value:    "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A...",
		},
		DMARC: DMARCRecord{
			Domain: "_dmarc.example.com",
			Value:  "v=DMARC1; p=none; sp=quarantine",
		},
		MX: MXRecord{
			Domain:   "example.com",
			Hostname: "relay.darkpipe.com",
			Priority: 10,
		},
	}

	var buf bytes.Buffer
	if err := PrintJSON(&buf, records); err != nil {
		t.Fatalf("PrintJSON() error = %v", err)
	}

	// Parse JSON to verify validity
	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("JSON output is invalid: %v", err)
	}

	// Verify structure
	if parsed["domain"] != "example.com" {
		t.Fatalf("JSON domain = %v, want example.com", parsed["domain"])
	}

	recordsObj, ok := parsed["records"].(map[string]interface{})
	if !ok {
		t.Fatal("JSON does not contain records object")
	}

	// Verify each record type exists
	for _, recordType := range []string{"spf", "dkim", "dmarc", "mx"} {
		if _, ok := recordsObj[recordType]; !ok {
			t.Fatalf("JSON records missing %s", recordType)
		}
	}
}

func TestSaveGuide(t *testing.T) {
	records := AllRecords{
		Domain: "example.com",
		SPF: SPFRecord{
			Domain: "example.com",
			Value:  "v=spf1 ip4:203.0.113.10 -all",
		},
		DKIM: DKIMRecord{
			Domain:   "darkpipe-2026q1._domainkey.example.com",
			Selector: "darkpipe-2026q1",
			Value:    "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A...",
		},
		DMARC: DMARCRecord{
			Domain: "_dmarc.example.com",
			Value:  "v=DMARC1; p=none; sp=quarantine",
		},
		MX: MXRecord{
			Domain:   "example.com",
			Hostname: "relay.darkpipe.com",
			Priority: 10,
		},
	}

	// Create temp file
	tmpFile := t.TempDir() + "/DNS-RECORDS.md"

	if err := SaveGuide(tmpFile, records); err != nil {
		t.Fatalf("SaveGuide() error = %v", err)
	}

	// Verify file exists
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read saved guide: %v", err)
	}

	// Verify content matches GenerateGuide output
	expected := GenerateGuide(records)
	if string(content) != expected {
		t.Fatal("Saved guide content does not match GenerateGuide output")
	}
}
