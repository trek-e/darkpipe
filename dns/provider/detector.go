package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/darkpipe/darkpipe/dns/config"
	"github.com/miekg/dns"
)

// DetectProvider auto-detects the DNS provider for a domain by querying NS records.
// Returns provider name ("cloudflare", "route53", or "unknown").
func DetectProvider(ctx context.Context, domain string) (string, error) {
	// Create DNS client
	c := &dns.Client{}
	m := &dns.Msg{}

	// Query NS records for the domain
	m.SetQuestion(dns.Fqdn(domain), dns.TypeNS)

	// Query against Google DNS (8.8.8.8)
	r, _, err := c.Exchange(m, "8.8.8.8:53")
	if err != nil {
		return "", fmt.Errorf("failed to query NS records for %s: %w", domain, err)
	}

	// Check if we got any NS records
	if len(r.Answer) == 0 {
		return "unknown", nil
	}

	// Parse NS records to detect provider
	for _, ans := range r.Answer {
		if ns, ok := ans.(*dns.NS); ok {
			nsHost := strings.ToLower(ns.Ns)

			// Cloudflare nameservers contain "cloudflare.com"
			if strings.Contains(nsHost, "cloudflare.com") {
				return "cloudflare", nil
			}

			// AWS Route53 nameservers contain "awsdns"
			if strings.Contains(nsHost, "awsdns") {
				return "route53", nil
			}
		}
	}

	// Unknown provider (will fall back to manual guide)
	return "unknown", nil
}

// NewProviderFromDetection creates a DNS provider based on auto-detection.
// Returns nil provider with descriptive message for "unknown" providers.
func NewProviderFromDetection(ctx context.Context, domain string, cfg *config.DNSConfig) (DNSProvider, error) {
	providerName, err := DetectProvider(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to detect DNS provider: %w", err)
	}

	switch providerName {
	case "cloudflare":
		// Note: Cloudflare implementation will be added in Task 2
		return nil, fmt.Errorf("cloudflare provider not yet implemented (will be added in Task 2)")

	case "route53":
		// Note: Route53 implementation will be added in Task 2
		return nil, fmt.Errorf("route53 provider not yet implemented (will be added in Task 2)")

	case "unknown":
		// Return nil provider with descriptive message
		// This is not an error - it means we should fall back to manual guide
		return nil, fmt.Errorf("DNS provider could not be detected for domain %s. Please add DNS records manually using the generated guide", domain)

	default:
		return nil, fmt.Errorf("unexpected provider type: %s", providerName)
	}
}
