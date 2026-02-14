package validator

import (
	"context"
	"fmt"
	"net"
	"strings"
)

// PTRResult represents the result of a PTR (reverse DNS) verification.
type PTRResult struct {
	IP               string   // IP address being checked
	ExpectedHostname string   // Expected hostname
	PTRNames         []string // Hostnames returned from reverse DNS lookup
	ForwardMatch     bool     // Whether forward DNS confirms the match
	Pass             bool     // Overall pass/fail
	Error            string   // Error message if check failed
}

// CheckPTR performs PTR reverse DNS verification.
// This implements the RFC-standard PTR check:
// 1. Reverse DNS lookup: IP -> hostnames (using net.LookupAddr)
// 2. For each returned hostname, forward DNS lookup: hostname -> IPs
// 3. Verify original IP appears in forward lookup results
// 4. Verify expectedHostname appears in PTR names
func CheckPTR(ctx context.Context, ip string, expectedHostname string) PTRResult {
	result := PTRResult{
		IP:               ip,
		ExpectedHostname: expectedHostname,
	}

	// Step 1: Reverse DNS lookup (IP -> hostnames)
	// net.LookupAddr handles the in-addr.arpa construction automatically
	names, err := net.DefaultResolver.LookupAddr(ctx, ip)
	if err != nil {
		result.Error = fmt.Sprintf("Reverse DNS lookup failed: %v\n\nPTR records are set by your VPS provider, not your DNS provider.\nConsult your VPS provider's documentation for PTR record configuration.\nCommon providers:\n- DigitalOcean: https://docs.digitalocean.com/products/networking/dns/how-to/manage-records/#ptr-records\n- Linode: https://www.linode.com/docs/guides/configure-your-linode-for-reverse-dns/\n- Vultr: https://www.vultr.com/docs/how-to-setup-reverse-dns-on-vultr/\n- Hetzner: https://docs.hetzner.com/dns-console/dns/general/reverse-dns/", err)
		return result
	}

	if len(names) == 0 {
		result.Error = fmt.Sprintf("No PTR records found for IP %s\n\nPTR records are set by your VPS provider, not your DNS provider.\nYou need to configure reverse DNS (rDNS) in your VPS provider's control panel.\nSet the PTR record to point to: %s", ip, expectedHostname)
		return result
	}

	// Store PTR names (remove trailing dots)
	for _, name := range names {
		result.PTRNames = append(result.PTRNames, strings.TrimSuffix(name, "."))
	}

	// Step 2 & 3: For each PTR name, do forward lookup and verify IP match
	result.ForwardMatch = false
	for _, name := range result.PTRNames {
		// Forward DNS lookup (hostname -> IPs)
		ips, err := net.DefaultResolver.LookupHost(ctx, name)
		if err != nil {
			continue // Try next name
		}

		// Check if original IP is in the forward lookup results
		for _, forwardIP := range ips {
			if forwardIP == ip {
				result.ForwardMatch = true
				break
			}
		}

		if result.ForwardMatch {
			break
		}
	}

	if !result.ForwardMatch {
		result.Error = fmt.Sprintf("Forward DNS lookup does not confirm reverse DNS match. PTR records: %s", strings.Join(result.PTRNames, ", "))
		return result
	}

	// Step 4: Verify expected hostname appears in PTR names
	hostnameFound := false
	for _, name := range result.PTRNames {
		if name == expectedHostname || name == expectedHostname+"." {
			hostnameFound = true
			break
		}
	}

	if !hostnameFound {
		result.Error = fmt.Sprintf("Expected hostname %s not found in PTR records. Found: %s\n\nUpdate your PTR record at your VPS provider to: %s", expectedHostname, strings.Join(result.PTRNames, ", "), expectedHostname)
		return result
	}

	result.Pass = true
	return result
}
