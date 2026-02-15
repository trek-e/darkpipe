// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package records

import (
	"fmt"
)

// DKIMRecord represents a DKIM TXT record.
type DKIMRecord struct {
	Domain   string // {selector}._domainkey.{domain}
	Selector string // DKIM selector
	Value    string // DKIM record content
}

// GenerateDKIMRecord generates a DKIM TXT record from a base64-encoded public key.
// The record is formatted as a single-line string (no newlines) per RFC 6376.
func GenerateDKIMRecord(domain string, selector string, publicKeyBase64 string) DKIMRecord {
	// Format: v=DKIM1; k=rsa; p={base64PublicKey}
	// Single-line output is critical for DNS TXT record compatibility
	value := fmt.Sprintf("v=DKIM1; k=rsa; p=%s", publicKeyBase64)

	// Subdomain format: {selector}._domainkey.{domain}
	recordDomain := fmt.Sprintf("%s._domainkey.%s", selector, domain)

	return DKIMRecord{
		Domain:   recordDomain,
		Selector: selector,
		Value:    value,
	}
}
