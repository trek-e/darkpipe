package records

import (
	"strings"
	"testing"
)

func TestGenerateSPF(t *testing.T) {
	tests := []struct {
		name          string
		domain        string
		relayIP       string
		additionalIPs []string
		wantContains  []string
	}{
		{
			name:          "single IP",
			domain:        "example.com",
			relayIP:       "203.0.113.10",
			additionalIPs: nil,
			wantContains:  []string{"v=spf1", "ip4:203.0.113.10", "-all"},
		},
		{
			name:          "multiple IPs",
			domain:        "example.com",
			relayIP:       "203.0.113.10",
			additionalIPs: []string{"203.0.113.20", "203.0.113.30"},
			wantContains:  []string{"v=spf1", "ip4:203.0.113.10", "ip4:203.0.113.20", "ip4:203.0.113.30", "-all"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := GenerateSPF(tt.domain, tt.relayIP, tt.additionalIPs)

			if record.Domain != tt.domain {
				t.Fatalf("Domain = %s, want %s", record.Domain, tt.domain)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(record.Value, want) {
					t.Fatalf("SPF record does not contain %q. Got: %s", want, record.Value)
				}
			}

			// Verify format
			if !strings.HasPrefix(record.Value, "v=spf1") {
				t.Fatalf("SPF record does not start with 'v=spf1'. Got: %s", record.Value)
			}
			if !strings.HasSuffix(record.Value, "-all") {
				t.Fatalf("SPF record does not end with '-all'. Got: %s", record.Value)
			}
		})
	}
}
