// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package records

import (
	"testing"
)

func TestGenerateMX(t *testing.T) {
	tests := []struct {
		name          string
		domain        string
		relayHostname string
		priority      int
		wantPriority  int
	}{
		{
			name:          "default priority",
			domain:        "example.com",
			relayHostname: "relay.darkpipe.com",
			priority:      0,
			wantPriority:  10,
		},
		{
			name:          "custom priority",
			domain:        "example.com",
			relayHostname: "relay.darkpipe.com",
			priority:      20,
			wantPriority:  20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := GenerateMX(tt.domain, tt.relayHostname, tt.priority)

			if record.Domain != tt.domain {
				t.Fatalf("Domain = %s, want %s", record.Domain, tt.domain)
			}
			if record.Hostname != tt.relayHostname {
				t.Fatalf("Hostname = %s, want %s", record.Hostname, tt.relayHostname)
			}
			if record.Priority != tt.wantPriority {
				t.Fatalf("Priority = %d, want %d", record.Priority, tt.wantPriority)
			}
		})
	}
}
