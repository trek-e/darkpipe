// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package records

import (
	"fmt"
	"strings"
	"testing"
)

func TestGenerateDKIMRecord(t *testing.T) {
	domain := "example.com"
	selector := "darkpipe-2026q1"
	publicKeyBase64 := "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA..."

	record := GenerateDKIMRecord(domain, selector, publicKeyBase64)

	// Verify domain format
	expectedDomain := "darkpipe-2026q1._domainkey.example.com"
	if record.Domain != expectedDomain {
		t.Fatalf("Domain = %s, want %s", record.Domain, expectedDomain)
	}

	// Verify selector
	if record.Selector != selector {
		t.Fatalf("Selector = %s, want %s", record.Selector, selector)
	}

	// Verify value format
	if !strings.HasPrefix(record.Value, "v=DKIM1;") {
		t.Fatalf("DKIM record does not start with 'v=DKIM1;'. Got: %s", record.Value)
	}
	if !strings.Contains(record.Value, "k=rsa;") {
		t.Fatalf("DKIM record does not contain 'k=rsa;'. Got: %s", record.Value)
	}
	if !strings.Contains(record.Value, fmt.Sprintf("p=%s", publicKeyBase64)) {
		t.Fatalf("DKIM record does not contain public key. Got: %s", record.Value)
	}

	// Verify single-line (no newlines)
	if strings.Contains(record.Value, "\n") || strings.Contains(record.Value, "\r") {
		t.Fatalf("DKIM record contains newlines (must be single-line). Got: %s", record.Value)
	}
}
