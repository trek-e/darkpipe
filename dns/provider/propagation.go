// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// DefaultDNSServers are public DNS resolvers used for propagation checks.
var DefaultDNSServers = []string{
	"8.8.8.8:53",       // Google
	"1.1.1.1:53",       // Cloudflare
	"208.67.222.222:53", // OpenDNS
}

// WaitForPropagation polls DNS servers until all confirm the expected record value
// or timeout is reached. Returns error if timeout occurs before full propagation.
func WaitForPropagation(ctx context.Context, name string, recordType string, expectedValue string, timeout time.Duration, servers []string) error {
	if len(servers) == 0 {
		servers = DefaultDNSServers
	}

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Normalize expected value for comparison
	expectedNorm := normalizeValue(recordType, expectedValue)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			confirmed := 0
			for _, server := range servers {
				if checkRecord(name, recordType, expectedNorm, server) {
					confirmed++
				}
			}

			fmt.Printf("Waiting for DNS propagation... (%d/%d servers confirmed)\n", confirmed, len(servers))

			if confirmed == len(servers) {
				fmt.Printf("DNS propagation complete! All servers confirmed.\n")
				return nil
			}

			if time.Now().After(deadline) {
				return fmt.Errorf("DNS propagation timeout after %v (only %d/%d servers confirmed)", timeout, confirmed, len(servers))
			}
		}
	}
}

// checkRecord queries a specific DNS server to check if a record has propagated.
func checkRecord(name string, recordType string, expectedValue string, server string) bool {
	c := &dns.Client{
		Timeout: 3 * time.Second,
	}
	m := &dns.Msg{}

	// Map record type string to DNS type
	var dnsType uint16
	switch strings.ToUpper(recordType) {
	case "TXT":
		dnsType = dns.TypeTXT
	case "MX":
		dnsType = dns.TypeMX
	case "A":
		dnsType = dns.TypeA
	case "CNAME":
		dnsType = dns.TypeCNAME
	default:
		return false
	}

	m.SetQuestion(dns.Fqdn(name), dnsType)

	r, _, err := c.Exchange(m, server)
	if err != nil {
		return false
	}

	// Check answers for expected value
	for _, ans := range r.Answer {
		actualValue := extractValue(ans, recordType)
		if actualValue != "" && actualValue == expectedValue {
			return true
		}
	}

	return false
}

// extractValue extracts the value from a DNS answer based on record type.
func extractValue(ans dns.RR, recordType string) string {
	switch strings.ToUpper(recordType) {
	case "TXT":
		if txt, ok := ans.(*dns.TXT); ok {
			// Join multiple TXT strings (some DNS servers split long records)
			return normalizeValue("TXT", strings.Join(txt.Txt, ""))
		}
	case "MX":
		if mx, ok := ans.(*dns.MX); ok {
			return strings.TrimSuffix(mx.Mx, ".")
		}
	case "A":
		if a, ok := ans.(*dns.A); ok {
			return a.A.String()
		}
	case "CNAME":
		if cname, ok := ans.(*dns.CNAME); ok {
			return strings.TrimSuffix(cname.Target, ".")
		}
	}
	return ""
}

// normalizeValue normalizes values for comparison (removes whitespace, quotes, etc.)
func normalizeValue(recordType string, value string) string {
	value = strings.TrimSpace(value)

	// For TXT records, remove surrounding quotes if present
	if strings.ToUpper(recordType) == "TXT" {
		value = strings.Trim(value, "\"")
	}

	return value
}
