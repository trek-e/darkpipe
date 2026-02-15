// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package records

// MXRecord represents an MX (mail exchange) record.
type MXRecord struct {
	Domain   string // "@" or domain name
	Hostname string // Mail server hostname (FQDN)
	Priority int    // MX priority (lower = higher priority)
}

// GenerateMX generates an MX record pointing to the cloud relay.
// Default priority is 10 (standard for primary mail server).
func GenerateMX(domain string, relayHostname string, priority int) MXRecord {
	if priority == 0 {
		priority = 10 // Default priority
	}

	return MXRecord{
		Domain:   domain,
		Hostname: relayHostname,
		Priority: priority,
	}
}
