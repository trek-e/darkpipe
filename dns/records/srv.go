package records

import "fmt"

// SRVRecord represents an RFC 2782 SRV DNS record.
type SRVRecord struct {
	Service  string // e.g., "_imaps"
	Proto    string // e.g., "_tcp"
	Domain   string // e.g., "example.com"
	Priority uint16
	Weight   uint16
	Port     uint16
	Target   string // e.g., "mail.example.com."
}

// String formats the SRV record as a DNS zone file line.
// Format: _service._proto.domain. IN SRV priority weight port target.
func (r SRVRecord) String() string {
	return fmt.Sprintf("%s.%s.%s. IN SRV %d %d %d %s",
		r.Service, r.Proto, r.Domain,
		r.Priority, r.Weight, r.Port,
		ensureTrailingDot(r.Target))
}

// GenerateSRVRecords generates standard email SRV records per RFC 6186.
// Returns records for IMAPS (preferred), IMAP (unavailable), and SUBMISSION.
func GenerateSRVRecords(domain, mailHostname string) []SRVRecord {
	return []SRVRecord{
		// IMAPS (preferred, TLS-only)
		{
			Service:  "_imaps",
			Proto:    "_tcp",
			Domain:   domain,
			Priority: 0,
			Weight:   1,
			Port:     993,
			Target:   mailHostname,
		},
		// IMAP (unavailable - DarkPipe only supports TLS)
		{
			Service:  "_imap",
			Proto:    "_tcp",
			Domain:   domain,
			Priority: 10,
			Weight:   0,
			Port:     143,
			Target:   ".", // "." indicates service unavailable per RFC 2782
		},
		// SUBMISSION (mail submission over TLS)
		{
			Service:  "_submission",
			Proto:    "_tcp",
			Domain:   domain,
			Priority: 0,
			Weight:   1,
			Port:     587,
			Target:   mailHostname,
		},
	}
}

// DNSRecord represents a generic DNS record (for CNAME, etc).
type DNSRecord struct {
	Name   string
	Type   string
	Target string
}

// String formats the DNS record as a DNS zone file line.
func (r DNSRecord) String() string {
	return fmt.Sprintf("%s IN %s %s",
		ensureTrailingDot(r.Name),
		r.Type,
		ensureTrailingDot(r.Target))
}

// GenerateAutodiscoverCNAME generates CNAME records for autoconfig and autodiscover subdomains.
// These point to the cloud relay hostname for Thunderbird/Outlook autodiscovery.
func GenerateAutodiscoverCNAME(domain, relayHostname string) []DNSRecord {
	return []DNSRecord{
		{
			Name:   fmt.Sprintf("autoconfig.%s", domain),
			Type:   "CNAME",
			Target: relayHostname,
		},
		{
			Name:   fmt.Sprintf("autodiscover.%s", domain),
			Type:   "CNAME",
			Target: relayHostname,
		},
	}
}

// ensureTrailingDot adds a trailing dot to a hostname if not present.
func ensureTrailingDot(hostname string) string {
	if hostname == "" || hostname == "." {
		return hostname
	}
	if hostname[len(hostname)-1] != '.' {
		return hostname + "."
	}
	return hostname
}
