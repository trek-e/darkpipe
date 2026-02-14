package cloudflare

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/cloudflare/cloudflare-go/v6/zones"
	"github.com/darkpipe/darkpipe/dns/provider"
)

// Compile-time interface check
var _ provider.DNSProvider = (*Client)(nil)

// Register the Cloudflare provider on package import
func init() {
	provider.RegisterProvider("cloudflare", func(ctx context.Context) (provider.DNSProvider, error) {
		apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
		if apiToken == "" {
			return nil, fmt.Errorf("CLOUDFLARE_API_TOKEN environment variable is required for Cloudflare DNS provider")
		}
		return NewClient(apiToken)
	})
}

// Client implements the DNSProvider interface for Cloudflare.
type Client struct {
	client *cloudflare.Client
}

// NewClient creates a new Cloudflare DNS provider client.
// apiToken should come from the CLOUDFLARE_API_TOKEN environment variable.
func NewClient(apiToken string) (*Client, error) {
	if apiToken == "" {
		return nil, fmt.Errorf("CLOUDFLARE_API_TOKEN is required")
	}

	client := cloudflare.NewClient(option.WithAPIToken(apiToken))

	return &Client{
		client: client,
	}, nil
}

// GetZoneID retrieves the zone ID for a given domain.
func (c *Client) GetZoneID(ctx context.Context, domain string) (string, error) {
	// List zones with name filter
	zoneList, err := c.client.Zones.List(ctx, zones.ZoneListParams{
		Name: cloudflare.F(domain),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list zones: %w", err)
	}

	// Get the first page of results
	if len(zoneList.Result) == 0 {
		return "", fmt.Errorf("zone not found for domain: %s", domain)
	}

	return zoneList.Result[0].ID, nil
}

// CreateRecord creates a new DNS record.
// Before creating SPF records, checks for existing SPF and updates instead.
func (c *Client) CreateRecord(ctx context.Context, rec provider.Record) error {
	zoneID, err := c.GetZoneID(ctx, extractDomain(rec.Name))
	if err != nil {
		return err
	}

	// Check for existing SPF record (pitfall #8: multiple SPF records break authentication)
	if rec.Type == "TXT" && strings.HasPrefix(rec.Content, "v=spf1") {
		existing, err := c.ListRecords(ctx, provider.RecordFilter{
			Type: "TXT",
			Name: rec.Name,
		})
		if err != nil {
			return fmt.Errorf("failed to check for existing SPF record: %w", err)
		}

		// If SPF record already exists, update it instead
		for _, existingRec := range existing {
			if strings.HasPrefix(existingRec.Content, "v=spf1") {
				return c.UpdateRecord(ctx, existingRec.ID, rec)
			}
		}
	}

	// Build DNS record params based on type
	var body dns.RecordNewParamsBodyUnion

	switch rec.Type {
	case "TXT":
		body = dns.TXTRecordParam{
			Name:    cloudflare.F(rec.Name),
			TTL:     cloudflare.F(dns.TTL(rec.TTL)),
			Type:    cloudflare.F(dns.TXTRecordTypeTXT),
			Content: cloudflare.F(rec.Content),
		}
	case "MX":
		priority := float64(10)
		if rec.Priority != nil {
			priority = float64(*rec.Priority)
		}
		body = dns.MXRecordParam{
			Name:     cloudflare.F(rec.Name),
			TTL:      cloudflare.F(dns.TTL(rec.TTL)),
			Type:     cloudflare.F(dns.MXRecordTypeMX),
			Content:  cloudflare.F(rec.Content),
			Priority: cloudflare.F(priority),
		}
	case "A":
		body = dns.ARecordParam{
			Name:    cloudflare.F(rec.Name),
			TTL:     cloudflare.F(dns.TTL(rec.TTL)),
			Type:    cloudflare.F(dns.ARecordTypeA),
			Content: cloudflare.F(rec.Content),
		}
	case "CNAME":
		body = dns.CNAMERecordParam{
			Name:    cloudflare.F(rec.Name),
			TTL:     cloudflare.F(dns.TTL(rec.TTL)),
			Type:    cloudflare.F(dns.CNAMERecordTypeCNAME),
			Content: cloudflare.F(rec.Content),
		}
	default:
		return fmt.Errorf("unsupported record type: %s", rec.Type)
	}

	params := dns.RecordNewParams{
		ZoneID: cloudflare.F(zoneID),
		Body:   body,
	}

	_, err = c.client.DNS.Records.New(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create DNS record: %w", err)
	}

	return nil
}

// UpdateRecord updates an existing DNS record.
func (c *Client) UpdateRecord(ctx context.Context, recordID string, rec provider.Record) error {
	zoneID, err := c.GetZoneID(ctx, extractDomain(rec.Name))
	if err != nil {
		return err
	}

	// Build update params based on type
	var body dns.RecordUpdateParamsBodyUnion

	switch rec.Type {
	case "TXT":
		body = dns.TXTRecordParam{
			Name:    cloudflare.F(rec.Name),
			TTL:     cloudflare.F(dns.TTL(rec.TTL)),
			Type:    cloudflare.F(dns.TXTRecordTypeTXT),
			Content: cloudflare.F(rec.Content),
		}
	case "MX":
		priority := float64(10)
		if rec.Priority != nil {
			priority = float64(*rec.Priority)
		}
		body = dns.MXRecordParam{
			Name:     cloudflare.F(rec.Name),
			TTL:      cloudflare.F(dns.TTL(rec.TTL)),
			Type:     cloudflare.F(dns.MXRecordTypeMX),
			Content:  cloudflare.F(rec.Content),
			Priority: cloudflare.F(priority),
		}
	case "A":
		body = dns.ARecordParam{
			Name:    cloudflare.F(rec.Name),
			TTL:     cloudflare.F(dns.TTL(rec.TTL)),
			Type:    cloudflare.F(dns.ARecordTypeA),
			Content: cloudflare.F(rec.Content),
		}
	case "CNAME":
		body = dns.CNAMERecordParam{
			Name:    cloudflare.F(rec.Name),
			TTL:     cloudflare.F(dns.TTL(rec.TTL)),
			Type:    cloudflare.F(dns.CNAMERecordTypeCNAME),
			Content: cloudflare.F(rec.Content),
		}
	default:
		return fmt.Errorf("unsupported record type: %s", rec.Type)
	}

	params := dns.RecordUpdateParams{
		ZoneID: cloudflare.F(zoneID),
		Body:   body,
	}

	_, err = c.client.DNS.Records.Update(ctx, recordID, params)
	if err != nil {
		return fmt.Errorf("failed to update DNS record: %w", err)
	}

	return nil
}

// ListRecords lists DNS records matching the filter.
func (c *Client) ListRecords(ctx context.Context, filter provider.RecordFilter) ([]provider.Record, error) {
	zoneID, err := c.GetZoneID(ctx, extractDomain(filter.Name))
	if err != nil {
		return nil, err
	}

	// Build list params
	params := dns.RecordListParams{
		ZoneID: cloudflare.F(zoneID),
	}

	// Add type filter
	if filter.Type != "" {
		params.Type = cloudflare.F(dns.RecordListParamsType(filter.Type))
	}

	// Note: Name filtering happens client-side as the API doesn't support exact name match in params

	// List records
	recordList, err := c.client.DNS.Records.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS records: %w", err)
	}

	// Convert to provider.Record and apply name filter
	var records []provider.Record
	for _, cfRec := range recordList.Result {
		// Apply name filter if specified
		if filter.Name != "" && cfRec.Name != filter.Name {
			continue
		}

		rec := provider.Record{
			ID:      cfRec.ID,
			Type:    string(cfRec.Type),
			Name:    cfRec.Name,
			Content: cfRec.Content,
			TTL:     int(cfRec.TTL),
		}

		// Add priority for MX records
		if cfRec.Priority != 0 {
			priority := int(cfRec.Priority)
			rec.Priority = &priority
		}

		records = append(records, rec)
	}

	return records, nil
}

// DeleteRecord deletes a DNS record by ID.
func (c *Client) DeleteRecord(ctx context.Context, recordID string) error {
	// We need zone ID for the delete operation
	// This is a limitation - we'd need to store or look up the zone
	return fmt.Errorf("DeleteRecord requires zone ID - not yet implemented")
}

// Name returns the provider name.
func (c *Client) Name() string {
	return "cloudflare"
}

// extractDomain extracts the base domain from a record name.
// For "@" or "example.com", returns "example.com".
// For "www.example.com", returns "example.com".
func extractDomain(name string) string {
	// If name is "@", we need the domain from context (will need to be passed separately)
	// For now, return as-is
	if name == "@" || !strings.Contains(name, ".") {
		// This is a limitation - we need domain context
		// Caller should pass full domain, not "@"
		return name
	}

	// Split by dots and take last two parts (domain.tld)
	parts := strings.Split(name, ".")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], ".")
	}

	return name
}
