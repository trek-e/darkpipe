package provider

import "context"

// DNSProvider defines the interface for DNS provider implementations.
// This interface allows community contributors to add new DNS providers
// (e.g., Namecheap, GoDaddy, DigitalOcean) without modifying core code.
type DNSProvider interface {
	// CreateRecord creates a new DNS record.
	CreateRecord(ctx context.Context, rec Record) error

	// UpdateRecord updates an existing DNS record by ID.
	UpdateRecord(ctx context.Context, recordID string, rec Record) error

	// ListRecords lists DNS records matching the provided filter.
	ListRecords(ctx context.Context, filter RecordFilter) ([]Record, error)

	// DeleteRecord deletes a DNS record by ID.
	DeleteRecord(ctx context.Context, recordID string) error

	// GetZoneID retrieves the zone ID for a given domain.
	GetZoneID(ctx context.Context, domain string) (string, error)

	// Name returns the provider name (e.g., "cloudflare", "route53", "unknown").
	Name() string
}

// Record represents a DNS record.
type Record struct {
	ID       string  // Provider-specific record ID
	Type     string  // Record type (TXT, MX, A, CNAME, etc.)
	Name     string  // Record name (e.g., "@", "www", "_dmarc")
	Content  string  // Record content/value
	TTL      int     // Time to live in seconds
	Priority *int    // Priority (for MX records, nil for others)
}

// RecordFilter specifies criteria for filtering DNS records.
type RecordFilter struct {
	Type string // Filter by record type (empty = all types)
	Name string // Filter by record name (empty = all names)
}
