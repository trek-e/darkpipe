package validator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// CheckResult represents the result of a single DNS record validation check.
type CheckResult struct {
	RecordType string // SPF, DKIM, DMARC, MX
	Name       string // DNS name queried
	Expected   string // Expected value or pattern
	Actual     string // Actual DNS response
	Pass       bool   // Whether the check passed
	Error      string // Error message if check failed
}

// ValidationReport aggregates all validation results.
type ValidationReport struct {
	Results   []CheckResult
	AllPassed bool
	Timestamp time.Time
}

// Checker queries DNS servers to validate records.
type Checker struct {
	client  *dns.Client
	servers []string
}

// NewChecker creates a DNS checker with specified servers.
// If servers is empty, defaults to Google (8.8.8.8:53) and Cloudflare (1.1.1.1:53).
func NewChecker(servers []string) *Checker {
	if len(servers) == 0 {
		servers = []string{"8.8.8.8:53", "1.1.1.1:53"}
	}

	return &Checker{
		client:  &dns.Client{Timeout: 5 * time.Second},
		servers: servers,
	}
}

// queryTXT queries TXT records for the given domain.
// Returns all TXT records found, or error if query fails.
func (c *Checker) queryTXT(ctx context.Context, domain string) ([]string, error) {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeTXT)

	// Try each server until one succeeds
	var lastErr error
	for _, server := range c.servers {
		r, _, err := c.client.ExchangeContext(ctx, msg, server)
		if err != nil {
			lastErr = fmt.Errorf("query to %s failed: %w", server, err)
			continue
		}

		if r.Rcode != dns.RcodeSuccess {
			lastErr = fmt.Errorf("query to %s returned %s", server, dns.RcodeToString[r.Rcode])
			continue
		}

		// Extract TXT records
		var records []string
		for _, answer := range r.Answer {
			if txt, ok := answer.(*dns.TXT); ok {
				// Join multiple strings (TXT records can be split)
				records = append(records, strings.Join(txt.Txt, ""))
			}
		}

		return records, nil
	}

	return nil, lastErr
}

// queryMX queries MX records for the given domain.
func (c *Checker) queryMX(ctx context.Context, domain string) ([]*dns.MX, error) {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeMX)

	var lastErr error
	for _, server := range c.servers {
		r, _, err := c.client.ExchangeContext(ctx, msg, server)
		if err != nil {
			lastErr = fmt.Errorf("query to %s failed: %w", server, err)
			continue
		}

		if r.Rcode != dns.RcodeSuccess {
			lastErr = fmt.Errorf("query to %s returned %s", server, dns.RcodeToString[r.Rcode])
			continue
		}

		var records []*dns.MX
		for _, answer := range r.Answer {
			if mx, ok := answer.(*dns.MX); ok {
				records = append(records, mx)
			}
		}

		return records, nil
	}

	return nil, lastErr
}

// CheckSPF validates the SPF record for the domain.
// Checks that:
// 1. An SPF record exists (starts with "v=spf1")
// 2. Only one SPF record exists (multiple SPF records is a common misconfiguration)
// 3. The expected IP is included via ip4: mechanism
func (c *Checker) CheckSPF(ctx context.Context, domain string, expectedIP string) CheckResult {
	result := CheckResult{
		RecordType: "SPF",
		Name:       domain,
		Expected:   fmt.Sprintf("ip4:%s", expectedIP),
	}

	records, err := c.queryTXT(ctx, domain)
	if err != nil {
		result.Error = fmt.Sprintf("DNS query failed: %v", err)
		return result
	}

	// Find SPF records (start with "v=spf1")
	var spfRecords []string
	for _, record := range records {
		if strings.HasPrefix(record, "v=spf1") {
			spfRecords = append(spfRecords, record)
		}
	}

	// Check for no SPF record
	if len(spfRecords) == 0 {
		result.Error = "No SPF record found"
		return result
	}

	// Check for multiple SPF records (pitfall #8)
	if len(spfRecords) > 1 {
		result.Actual = fmt.Sprintf("%d SPF records found", len(spfRecords))
		result.Error = "Multiple SPF records detected (RFC 7208 violation)"
		return result
	}

	// Check if expected IP is included
	spfRecord := spfRecords[0]
	result.Actual = spfRecord

	if !strings.Contains(spfRecord, fmt.Sprintf("ip4:%s", expectedIP)) {
		result.Error = fmt.Sprintf("SPF record does not contain ip4:%s", expectedIP)
		return result
	}

	result.Pass = true
	return result
}

// CheckDKIM validates the DKIM record for the domain with the given selector.
// Checks that:
// 1. The DKIM record exists at {selector}._domainkey.{domain}
// 2. It starts with "v=DKIM1"
// 3. It specifies k=rsa (RSA key type)
// 4. It contains a non-empty p= (public key) value
func (c *Checker) CheckDKIM(ctx context.Context, domain string, selector string) CheckResult {
	dkimDomain := fmt.Sprintf("%s._domainkey.%s", selector, domain)
	result := CheckResult{
		RecordType: "DKIM",
		Name:       dkimDomain,
		Expected:   "v=DKIM1 with k=rsa and p=<base64>",
	}

	records, err := c.queryTXT(ctx, dkimDomain)
	if err != nil {
		result.Error = fmt.Sprintf("DNS query failed: %v", err)
		return result
	}

	if len(records) == 0 {
		result.Error = "No DKIM record found"
		return result
	}

	// Use first record (should only be one)
	dkimRecord := records[0]
	result.Actual = dkimRecord

	// Check v=DKIM1
	if !strings.HasPrefix(dkimRecord, "v=DKIM1") {
		result.Error = "DKIM record does not start with v=DKIM1"
		return result
	}

	// Check k=rsa (key type)
	if !strings.Contains(dkimRecord, "k=rsa") {
		result.Error = "DKIM record does not specify k=rsa"
		return result
	}

	// Check p= contains non-empty value
	if !strings.Contains(dkimRecord, "p=") {
		result.Error = "DKIM record missing p= (public key)"
		return result
	}

	// Extract p= value and verify it's not empty
	parts := strings.Split(dkimRecord, "p=")
	if len(parts) < 2 {
		result.Error = "DKIM record p= value is malformed"
		return result
	}

	// Get the public key part (everything after p=, before next semicolon or end)
	pubKeyPart := parts[1]
	if idx := strings.Index(pubKeyPart, ";"); idx != -1 {
		pubKeyPart = pubKeyPart[:idx]
	}
	pubKeyPart = strings.TrimSpace(pubKeyPart)

	if pubKeyPart == "" {
		result.Error = "DKIM record p= value is empty"
		return result
	}

	result.Pass = true
	return result
}

// CheckDMARC validates the DMARC record for the domain.
// Checks that:
// 1. A DMARC record exists at _dmarc.{domain}
// 2. It starts with "v=DMARC1"
// 3. It contains a p= policy value
func (c *Checker) CheckDMARC(ctx context.Context, domain string) CheckResult {
	dmarcDomain := fmt.Sprintf("_dmarc.%s", domain)
	result := CheckResult{
		RecordType: "DMARC",
		Name:       dmarcDomain,
		Expected:   "v=DMARC1 with p= policy",
	}

	records, err := c.queryTXT(ctx, dmarcDomain)
	if err != nil {
		result.Error = fmt.Sprintf("DNS query failed: %v", err)
		return result
	}

	if len(records) == 0 {
		result.Error = "No DMARC record found"
		return result
	}

	// Use first record (should only be one)
	dmarcRecord := records[0]
	result.Actual = dmarcRecord

	// Check v=DMARC1
	if !strings.HasPrefix(dmarcRecord, "v=DMARC1") {
		result.Error = "DMARC record does not start with v=DMARC1"
		return result
	}

	// Check for p= policy
	if !strings.Contains(dmarcRecord, "p=") {
		result.Error = "DMARC record missing p= policy"
		return result
	}

	// Extract policy value
	var policy string
	parts := strings.Split(dmarcRecord, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "p=") {
			policy = strings.TrimPrefix(part, "p=")
			break
		}
	}

	if policy == "" {
		result.Error = "DMARC record p= policy is empty"
		return result
	}

	// Report the policy in the actual field
	result.Actual = fmt.Sprintf("%s (policy: %s)", dmarcRecord, policy)
	result.Pass = true
	return result
}

// CheckMX validates the MX record for the domain.
// Checks that at least one MX record points to the expected hostname.
func (c *Checker) CheckMX(ctx context.Context, domain string, expectedHostname string) CheckResult {
	result := CheckResult{
		RecordType: "MX",
		Name:       domain,
		Expected:   expectedHostname,
	}

	mxRecords, err := c.queryMX(ctx, domain)
	if err != nil {
		result.Error = fmt.Sprintf("DNS query failed: %v", err)
		return result
	}

	if len(mxRecords) == 0 {
		result.Error = "No MX records found"
		return result
	}

	// Build list of MX hostnames
	var hostnames []string
	for _, mx := range mxRecords {
		hostname := strings.TrimSuffix(mx.Mx, ".")
		hostnames = append(hostnames, hostname)
		if hostname == expectedHostname {
			result.Pass = true
		}
	}

	result.Actual = strings.Join(hostnames, ", ")

	if !result.Pass {
		result.Error = fmt.Sprintf("Expected hostname %s not found in MX records", expectedHostname)
	}

	return result
}

// CheckAll runs all DNS validation checks and returns an aggregated report.
func (c *Checker) CheckAll(ctx context.Context, domain, expectedIP, relayHostname, dkimSelector string) ValidationReport {
	report := ValidationReport{
		Timestamp: time.Now(),
	}

	// Run all checks
	report.Results = append(report.Results, c.CheckSPF(ctx, domain, expectedIP))
	report.Results = append(report.Results, c.CheckDKIM(ctx, domain, dkimSelector))
	report.Results = append(report.Results, c.CheckDMARC(ctx, domain))
	report.Results = append(report.Results, c.CheckMX(ctx, domain, relayHostname))

	// Check if all passed
	report.AllPassed = true
	for _, result := range report.Results {
		if !result.Pass {
			report.AllPassed = false
			break
		}
	}

	return report
}
