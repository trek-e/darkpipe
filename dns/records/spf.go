// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package records

import (
	"fmt"
	"strings"
)

// SPFRecord represents an SPF TXT record.
type SPFRecord struct {
	Domain string // "@" or domain name
	Value  string // SPF record content
}

// GenerateSPF generates an SPF record for the domain.
// Uses explicit ip4: mechanism to minimize DNS lookup count (RFC 7208).
// Always uses -all (hard fail) for maximum protection.
func GenerateSPF(domain string, relayIP string, additionalIPs []string) SPFRecord {
	// Start with v=spf1
	parts := []string{"v=spf1"}

	// Add relay IP
	parts = append(parts, fmt.Sprintf("ip4:%s", relayIP))

	// Add additional IPs
	for _, ip := range additionalIPs {
		parts = append(parts, fmt.Sprintf("ip4:%s", ip))
	}

	// Terminate with -all (hard fail)
	parts = append(parts, "-all")

	// Warning if too many IPs (SPF has 10 DNS lookup limit, but ip4: doesn't count as lookup)
	// This is mainly for documentation purposes
	if len(additionalIPs) > 8 {
		// Note: ip4: mechanisms don't count toward lookup limit,
		// but extremely long SPF records may hit DNS UDP size limits
		// Log warning in real implementation
	}

	return SPFRecord{
		Domain: domain,
		Value:  strings.Join(parts, " "),
	}
}
