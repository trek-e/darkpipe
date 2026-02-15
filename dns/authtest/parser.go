// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package authtest

import (
	"fmt"
	"strings"

	"github.com/emersion/go-msgauth/authres"
)

// AuthResult represents a single authentication result (SPF, DKIM, or DMARC).
type AuthResult struct {
	Method  string // "dkim", "spf", "dmarc"
	Result  string // "pass", "fail", "none", etc.
	Details string // Additional information (e.g., "header.s=darkpipe-2026q1")
}

// AuthReport aggregates all authentication results from an email.
type AuthReport struct {
	Results    []AuthResult
	SPFPass    bool
	DKIMPass   bool
	DMARCPass  bool
	RawHeader  string // Original Authentication-Results header
}

// ParseAuthResults parses an Authentication-Results header from an email.
// Uses emersion/go-msgauth/authres for RFC 8601 compliant parsing.
func ParseAuthResults(header string) (AuthReport, error) {
	report := AuthReport{
		RawHeader: header,
	}

	// Parse using emersion/go-msgauth
	_, results, err := authres.Parse(header)
	if err != nil {
		return report, fmt.Errorf("failed to parse Authentication-Results header: %w", err)
	}

	// Extract DKIM, SPF, and DMARC results
	for _, result := range results {
		var authResult AuthResult

		switch r := result.(type) {
		case *authres.DKIMResult:
			authResult.Method = "dkim"
			authResult.Result = string(r.Value)

			// Extract details (selector, domain)
			var details []string
			if r.Reason != "" {
				details = append(details, fmt.Sprintf("reason=%s", r.Reason))
			}
			if r.Domain != "" {
				details = append(details, fmt.Sprintf("header.d=%s", r.Domain))
			}
			if r.Identifier != "" {
				details = append(details, fmt.Sprintf("header.i=%s", r.Identifier))
			}
			authResult.Details = strings.Join(details, "; ")

			if r.Value == "pass" {
				report.DKIMPass = true
			}

		case *authres.SPFResult:
			authResult.Method = "spf"
			authResult.Result = string(r.Value)

			// Extract details (sender IP, envelope)
			var details []string
			if r.Reason != "" {
				details = append(details, fmt.Sprintf("reason=%s", r.Reason))
			}
			if r.From != "" {
				details = append(details, fmt.Sprintf("smtp.mailfrom=%s", r.From))
			}
			if r.Helo != "" {
				details = append(details, fmt.Sprintf("smtp.helo=%s", r.Helo))
			}
			authResult.Details = strings.Join(details, "; ")

			if r.Value == "pass" {
				report.SPFPass = true
			}

		case *authres.DMARCResult:
			authResult.Method = "dmarc"
			authResult.Result = string(r.Value)

			// Extract details (policy, alignment)
			var details []string
			if r.Reason != "" {
				details = append(details, fmt.Sprintf("reason=%s", r.Reason))
			}
			if r.From != "" {
				details = append(details, fmt.Sprintf("header.from=%s", r.From))
			}
			authResult.Details = strings.Join(details, "; ")

			if r.Value == "pass" {
				report.DMARCPass = true
			}

		default:
			// Skip unknown methods
			continue
		}

		report.Results = append(report.Results, authResult)
	}

	return report, nil
}
