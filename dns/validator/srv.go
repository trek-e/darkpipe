package validator

import (
	"context"
	"fmt"

	"github.com/miekg/dns"
)

// querySRV queries SRV records for the given service, protocol, and domain.
func (c *Checker) querySRV(ctx context.Context, service, proto, domain string) ([]*dns.SRV, error) {
	qname := fmt.Sprintf("%s.%s.%s", service, proto, domain)
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(qname), dns.TypeSRV)

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

		// Extract SRV records
		var records []*dns.SRV
		for _, answer := range r.Answer {
			if srv, ok := answer.(*dns.SRV); ok {
				records = append(records, srv)
			}
		}

		return records, nil
	}

	return nil, lastErr
}

// queryCNAME queries CNAME records for the given domain.
func (c *Checker) queryCNAME(ctx context.Context, domain string) (string, error) {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeCNAME)

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

		// Extract CNAME record
		for _, answer := range r.Answer {
			if cname, ok := answer.(*dns.CNAME); ok {
				return cname.Target, nil
			}
		}

		// No CNAME found
		return "", nil
	}

	return "", lastErr
}

// CheckSRV validates that the required SRV records exist for email autodiscovery.
func (c *Checker) CheckSRV(ctx context.Context, domain string) CheckResult {
	// Check _imaps._tcp SRV record
	imapsSrvs, err := c.querySRV(ctx, "_imaps", "_tcp", domain)
	if err != nil {
		return CheckResult{
			RecordType: "SRV",
			Name:       fmt.Sprintf("_imaps._tcp.%s", domain),
			Expected:   "SRV record with port 993",
			Actual:     "",
			Pass:       false,
			Error:      fmt.Sprintf("Query failed: %v", err),
		}
	}

	if len(imapsSrvs) == 0 {
		return CheckResult{
			RecordType: "SRV",
			Name:       fmt.Sprintf("_imaps._tcp.%s", domain),
			Expected:   "SRV record with port 993",
			Actual:     "No SRV record found",
			Pass:       false,
			Error:      "IMAPS SRV record not found",
		}
	}

	// Verify IMAPS port is 993
	imapsSrv := imapsSrvs[0]
	if imapsSrv.Port != 993 {
		return CheckResult{
			RecordType: "SRV",
			Name:       fmt.Sprintf("_imaps._tcp.%s", domain),
			Expected:   "Port 993",
			Actual:     fmt.Sprintf("Port %d", imapsSrv.Port),
			Pass:       false,
			Error:      "IMAPS SRV record has wrong port",
		}
	}

	// Check _submission._tcp SRV record
	submissionSrvs, err := c.querySRV(ctx, "_submission", "_tcp", domain)
	if err != nil {
		return CheckResult{
			RecordType: "SRV",
			Name:       fmt.Sprintf("_submission._tcp.%s", domain),
			Expected:   "SRV record with port 587",
			Actual:     "",
			Pass:       false,
			Error:      fmt.Sprintf("Query failed: %v", err),
		}
	}

	if len(submissionSrvs) == 0 {
		return CheckResult{
			RecordType: "SRV",
			Name:       fmt.Sprintf("_submission._tcp.%s", domain),
			Expected:   "SRV record with port 587",
			Actual:     "No SRV record found",
			Pass:       false,
			Error:      "SUBMISSION SRV record not found",
		}
	}

	// Verify SUBMISSION port is 587
	submissionSrv := submissionSrvs[0]
	if submissionSrv.Port != 587 {
		return CheckResult{
			RecordType: "SRV",
			Name:       fmt.Sprintf("_submission._tcp.%s", domain),
			Expected:   "Port 587",
			Actual:     fmt.Sprintf("Port %d", submissionSrv.Port),
			Pass:       false,
			Error:      "SUBMISSION SRV record has wrong port",
		}
	}

	return CheckResult{
		RecordType: "SRV",
		Name:       fmt.Sprintf("_imaps._tcp.%s and _submission._tcp.%s", domain, domain),
		Expected:   "SRV records for IMAPS (993) and SUBMISSION (587)",
		Actual:     fmt.Sprintf("Found IMAPS -> %s:%d and SUBMISSION -> %s:%d", imapsSrv.Target, imapsSrv.Port, submissionSrv.Target, submissionSrv.Port),
		Pass:       true,
		Error:      "",
	}
}

// CheckAutodiscoverCNAMEs validates that autoconfig and autodiscover CNAME records exist.
func (c *Checker) CheckAutodiscoverCNAMEs(ctx context.Context, domain, expectedTarget string) CheckResult {
	// Check autoconfig CNAME
	autoconfigDomain := fmt.Sprintf("autoconfig.%s", domain)
	autoconfigTarget, err := c.queryCNAME(ctx, autoconfigDomain)
	if err != nil {
		return CheckResult{
			RecordType: "CNAME",
			Name:       autoconfigDomain,
			Expected:   fmt.Sprintf("CNAME to %s", expectedTarget),
			Actual:     "",
			Pass:       false,
			Error:      fmt.Sprintf("Query failed: %v", err),
		}
	}

	// Check autodiscover CNAME
	autodiscoverDomain := fmt.Sprintf("autodiscover.%s", domain)
	autodiscoverTarget, err := c.queryCNAME(ctx, autodiscoverDomain)
	if err != nil {
		return CheckResult{
			RecordType: "CNAME",
			Name:       autodiscoverDomain,
			Expected:   fmt.Sprintf("CNAME to %s", expectedTarget),
			Actual:     "",
			Pass:       false,
			Error:      fmt.Sprintf("Query failed: %v", err),
		}
	}

	// Verify both exist
	if autoconfigTarget == "" && autodiscoverTarget == "" {
		return CheckResult{
			RecordType: "CNAME",
			Name:       fmt.Sprintf("autoconfig.%s and autodiscover.%s", domain, domain),
			Expected:   fmt.Sprintf("CNAME records to %s", expectedTarget),
			Actual:     "No CNAME records found",
			Pass:       false,
			Error:      "Autodiscover CNAME records not found",
		}
	}

	if autoconfigTarget == "" {
		return CheckResult{
			RecordType: "CNAME",
			Name:       autoconfigDomain,
			Expected:   fmt.Sprintf("CNAME to %s", expectedTarget),
			Actual:     "No CNAME record found",
			Pass:       false,
			Error:      "autoconfig CNAME record not found",
		}
	}

	if autodiscoverTarget == "" {
		return CheckResult{
			RecordType: "CNAME",
			Name:       autodiscoverDomain,
			Expected:   fmt.Sprintf("CNAME to %s", expectedTarget),
			Actual:     "No CNAME record found",
			Pass:       false,
			Error:      "autodiscover CNAME record not found",
		}
	}

	return CheckResult{
		RecordType: "CNAME",
		Name:       fmt.Sprintf("autoconfig.%s and autodiscover.%s", domain, domain),
		Expected:   fmt.Sprintf("CNAME records to %s", expectedTarget),
		Actual:     fmt.Sprintf("autoconfig -> %s, autodiscover -> %s", autoconfigTarget, autodiscoverTarget),
		Pass:       true,
		Error:      "",
	}
}
