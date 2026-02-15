// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package records

import (
	"fmt"
	"strings"
)

// DMARCRecord represents a DMARC TXT record.
type DMARCRecord struct {
	Domain string // _dmarc.{domain}
	Value  string // DMARC record content
}

// DMARCOptions configures DMARC policy settings.
type DMARCOptions struct {
	Policy           string // "none", "quarantine", or "reject"
	SubdomainPolicy  string // Subdomain policy (sp=), defaults to "quarantine"
	Percentage       int    // Percentage of messages to apply policy to (pct=), 0-100
	RUA              string // Aggregate report email (rua=)
	RUF              string // Forensic report email (ruf=)
}

// GenerateDMARC generates a DMARC TXT record with the specified options.
// Includes sp= tag for subdomain protection (pitfall #5).
// Recommended progression: none -> quarantine -> reject (pitfall #4).
func GenerateDMARC(domain string, opts DMARCOptions) DMARCRecord {
	// Start with v=DMARC1
	parts := []string{"v=DMARC1"}

	// Policy (required)
	policy := opts.Policy
	if policy == "" {
		policy = "none" // Default to monitoring mode
	}
	parts = append(parts, fmt.Sprintf("p=%s", policy))

	// Subdomain policy (important for security)
	subdomainPolicy := opts.SubdomainPolicy
	if subdomainPolicy == "" {
		subdomainPolicy = "quarantine" // More restrictive default for subdomains
	}
	parts = append(parts, fmt.Sprintf("sp=%s", subdomainPolicy))

	// Percentage (optional, defaults to 100 if omitted)
	if opts.Percentage > 0 && opts.Percentage < 100 {
		parts = append(parts, fmt.Sprintf("pct=%d", opts.Percentage))
	}

	// Aggregate reports (optional but recommended)
	if opts.RUA != "" {
		parts = append(parts, fmt.Sprintf("rua=mailto:%s", opts.RUA))
	}

	// Forensic reports (optional)
	if opts.RUF != "" {
		parts = append(parts, fmt.Sprintf("ruf=mailto:%s", opts.RUF))
	}

	// Subdomain: _dmarc.{domain}
	recordDomain := fmt.Sprintf("_dmarc.%s", domain)

	return DMARCRecord{
		Domain: recordDomain,
		Value:  strings.Join(parts, "; "),
	}
}
