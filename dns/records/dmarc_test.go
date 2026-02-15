// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package records

import (
	"strings"
	"testing"
)

func TestGenerateDMARC(t *testing.T) {
	tests := []struct {
		name         string
		domain       string
		opts         DMARCOptions
		wantDomain   string
		wantContains []string
	}{
		{
			name:   "basic policy none",
			domain: "example.com",
			opts: DMARCOptions{
				Policy: "none",
			},
			wantDomain:   "_dmarc.example.com",
			wantContains: []string{"v=DMARC1", "p=none", "sp=quarantine"},
		},
		{
			name:   "quarantine with reports",
			domain: "example.com",
			opts: DMARCOptions{
				Policy: "quarantine",
				RUA:    "dmarc-reports@example.com",
				RUF:    "dmarc-forensic@example.com",
			},
			wantDomain:   "_dmarc.example.com",
			wantContains: []string{"v=DMARC1", "p=quarantine", "sp=quarantine", "rua=mailto:dmarc-reports@example.com", "ruf=mailto:dmarc-forensic@example.com"},
		},
		{
			name:   "reject with custom subdomain policy",
			domain: "example.com",
			opts: DMARCOptions{
				Policy:          "reject",
				SubdomainPolicy: "reject",
				RUA:             "dmarc@example.com",
			},
			wantDomain:   "_dmarc.example.com",
			wantContains: []string{"v=DMARC1", "p=reject", "sp=reject", "rua=mailto:dmarc@example.com"},
		},
		{
			name:   "with percentage",
			domain: "example.com",
			opts: DMARCOptions{
				Policy:     "quarantine",
				Percentage: 50,
			},
			wantDomain:   "_dmarc.example.com",
			wantContains: []string{"v=DMARC1", "p=quarantine", "pct=50"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := GenerateDMARC(tt.domain, tt.opts)

			if record.Domain != tt.wantDomain {
				t.Fatalf("Domain = %s, want %s", record.Domain, tt.wantDomain)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(record.Value, want) {
					t.Fatalf("DMARC record does not contain %q. Got: %s", want, record.Value)
				}
			}

			// Verify format
			if !strings.HasPrefix(record.Value, "v=DMARC1") {
				t.Fatalf("DMARC record does not start with 'v=DMARC1'. Got: %s", record.Value)
			}

			// Verify required tags present
			if !strings.Contains(record.Value, "p=") {
				t.Fatal("DMARC record missing p= tag")
			}
			if !strings.Contains(record.Value, "sp=") {
				t.Fatal("DMARC record missing sp= tag (subdomain protection)")
			}
		})
	}
}
