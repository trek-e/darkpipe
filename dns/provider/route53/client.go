// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package route53

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/darkpipe/darkpipe/dns/provider"
)

// Compile-time interface check
var _ provider.DNSProvider = (*Client)(nil)

// Register the Route53 provider on package import
func init() {
	provider.RegisterProvider("route53", func(ctx context.Context) (provider.DNSProvider, error) {
		// AWS SDK will automatically load credentials from environment
		// (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION)
		return NewClient(ctx)
	})
}

// Client implements the DNSProvider interface for AWS Route53.
type Client struct {
	client *route53.Client
}

// NewClient creates a new Route53 DNS provider client.
// Credentials come from AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION env vars.
func NewClient(ctx context.Context) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := route53.NewFromConfig(cfg)

	return &Client{
		client: client,
	}, nil
}

// GetZoneID retrieves the hosted zone ID for a given domain.
func (c *Client) GetZoneID(ctx context.Context, domain string) (string, error) {
	// Ensure domain ends with a dot for Route53
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}

	// List hosted zones by name
	resp, err := c.client.ListHostedZonesByName(ctx, &route53.ListHostedZonesByNameInput{
		DNSName:  aws.String(domain),
		MaxItems: aws.Int32(1),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list hosted zones: %w", err)
	}

	if len(resp.HostedZones) == 0 {
		return "", fmt.Errorf("hosted zone not found for domain: %s", domain)
	}

	if resp.HostedZones[0].Name == nil || *resp.HostedZones[0].Name != domain {
		return "", fmt.Errorf("hosted zone not found for domain: %s", domain)
	}

	// Strip leading "/" from hosted zone ID (Route53 returns "/hostedzone/Z123...")
	zoneID := *resp.HostedZones[0].Id
	zoneID = strings.TrimPrefix(zoneID, "/hostedzone/")

	return zoneID, nil
}

// CreateRecord creates a new DNS record.
// Before creating SPF records, checks for existing SPF and uses UPSERT instead.
func (c *Client) CreateRecord(ctx context.Context, rec provider.Record) error {
	domain := extractDomain(rec.Name)
	zoneID, err := c.GetZoneID(ctx, domain)
	if err != nil {
		return err
	}

	// Check for existing SPF record (pitfall #8: multiple SPF records break authentication)
	action := types.ChangeActionCreate
	if rec.Type == "TXT" && strings.HasPrefix(rec.Content, "v=spf1") {
		existing, err := c.ListRecords(ctx, provider.RecordFilter{
			Type: "TXT",
			Name: rec.Name,
		})
		if err != nil {
			return fmt.Errorf("failed to check for existing SPF record: %w", err)
		}

		// If SPF record already exists, use UPSERT instead of CREATE
		for _, existingRec := range existing {
			if strings.HasPrefix(existingRec.Content, "v=spf1") {
				action = types.ChangeActionUpsert
				break
			}
		}
	}

	// Build record name (Route53 requires FQDN with trailing dot)
	recordName := rec.Name
	if !strings.HasSuffix(recordName, ".") {
		recordName += "."
	}

	// Build resource record
	resourceRecord := buildResourceRecord(rec)

	// Create change batch
	changeBatch := &types.ChangeBatch{
		Changes: []types.Change{
			{
				Action: action,
				ResourceRecordSet: &types.ResourceRecordSet{
					Name:            aws.String(recordName),
					Type:            types.RRType(rec.Type),
					TTL:             aws.Int64(int64(rec.TTL)),
					ResourceRecords: resourceRecord,
				},
			},
		},
	}

	// Execute change
	_, err = c.client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
		ChangeBatch:  changeBatch,
	})
	if err != nil {
		return fmt.Errorf("failed to create DNS record: %w", err)
	}

	return nil
}

// UpdateRecord updates an existing DNS record using UPSERT.
// Route53 doesn't use record IDs - it identifies records by name+type.
func (c *Client) UpdateRecord(ctx context.Context, recordID string, rec provider.Record) error {
	// Note: recordID is ignored in Route53 - we use name+type for identification
	domain := extractDomain(rec.Name)
	zoneID, err := c.GetZoneID(ctx, domain)
	if err != nil {
		return err
	}

	// Build record name (Route53 requires FQDN with trailing dot)
	recordName := rec.Name
	if !strings.HasSuffix(recordName, ".") {
		recordName += "."
	}

	// Build resource record
	resourceRecord := buildResourceRecord(rec)

	// Create change batch with UPSERT action
	changeBatch := &types.ChangeBatch{
		Changes: []types.Change{
			{
				Action: types.ChangeActionUpsert,
				ResourceRecordSet: &types.ResourceRecordSet{
					Name:            aws.String(recordName),
					Type:            types.RRType(rec.Type),
					TTL:             aws.Int64(int64(rec.TTL)),
					ResourceRecords: resourceRecord,
				},
			},
		},
	}

	// Execute change
	_, err = c.client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
		ChangeBatch:  changeBatch,
	})
	if err != nil {
		return fmt.Errorf("failed to update DNS record: %w", err)
	}

	return nil
}

// ListRecords lists DNS records matching the filter.
func (c *Client) ListRecords(ctx context.Context, filter provider.RecordFilter) ([]provider.Record, error) {
	domain := extractDomain(filter.Name)
	zoneID, err := c.GetZoneID(ctx, domain)
	if err != nil {
		return nil, err
	}

	// List all resource record sets for the zone
	resp, err := c.client.ListResourceRecordSets(ctx, &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list resource record sets: %w", err)
	}

	// Filter and convert records
	var records []provider.Record
	for _, rrset := range resp.ResourceRecordSets {
		// Apply type filter
		if filter.Type != "" && string(rrset.Type) != filter.Type {
			continue
		}

		// Apply name filter
		recordName := strings.TrimSuffix(*rrset.Name, ".")
		if filter.Name != "" && recordName != filter.Name {
			continue
		}

		// Convert resource records
		for _, rr := range rrset.ResourceRecords {
			rec := provider.Record{
				ID:      recordName + ":" + string(rrset.Type), // Composite ID
				Type:    string(rrset.Type),
				Name:    recordName,
				Content: extractContent(*rr.Value, string(rrset.Type)),
				TTL:     int(*rrset.TTL),
			}

			records = append(records, rec)
		}
	}

	return records, nil
}

// DeleteRecord deletes a DNS record.
// Route53 requires full record details, not just ID.
func (c *Client) DeleteRecord(ctx context.Context, recordID string) error {
	// Route53 requires name+type+value for deletion
	// This is a limitation of the current interface design
	return fmt.Errorf("DeleteRecord not fully implemented - Route53 requires full record details")
}

// Name returns the provider name.
func (c *Client) Name() string {
	return "route53"
}

// buildResourceRecord builds a Route53 ResourceRecord from a provider.Record.
// TXT records require special quoting.
func buildResourceRecord(rec provider.Record) []types.ResourceRecord {
	value := rec.Content

	// Route53 requires TXT record values wrapped in extra quotes
	if rec.Type == "TXT" && !strings.HasPrefix(value, "\"") {
		value = fmt.Sprintf("\"%s\"", value)
	}

	// MX records need priority in the value field
	if rec.Type == "MX" && rec.Priority != nil {
		value = fmt.Sprintf("%d %s", *rec.Priority, rec.Content)
	}

	return []types.ResourceRecord{
		{
			Value: aws.String(value),
		},
	}
}

// extractContent extracts the content from a Route53 value.
// Removes extra quotes from TXT records and priority from MX records.
func extractContent(value string, recordType string) string {
	// Remove extra quotes from TXT records
	if recordType == "TXT" {
		value = strings.Trim(value, "\"")
	}

	// Extract hostname from MX records (remove priority)
	if recordType == "MX" {
		parts := strings.Fields(value)
		if len(parts) >= 2 {
			return parts[1]
		}
	}

	return value
}

// extractDomain extracts the base domain from a record name.
// For "@" or "example.com", returns "example.com".
// For "www.example.com", returns "example.com".
func extractDomain(name string) string {
	if name == "@" || !strings.Contains(name, ".") {
		return name
	}

	// Split by dots and take last two parts (domain.tld)
	parts := strings.Split(name, ".")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], ".")
	}

	return name
}
