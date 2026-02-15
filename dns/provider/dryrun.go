// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package provider

import (
	"context"
	"fmt"
)

// DryRunProvider wraps a DNSProvider and intercepts write operations
// to print what would happen without actually making changes.
// Read operations always pass through to the underlying provider.
type DryRunProvider struct {
	underlying DNSProvider
	dryRun     bool
}

// NewDryRunProvider creates a new DryRunProvider wrapping the given provider.
// If dryRun is true, write operations are intercepted and logged.
// If dryRun is false, all operations pass through to the underlying provider.
func NewDryRunProvider(underlying DNSProvider, dryRun bool) *DryRunProvider {
	return &DryRunProvider{
		underlying: underlying,
		dryRun:     dryRun,
	}
}

// CreateRecord creates a DNS record (or logs what would be created in dry-run mode).
func (p *DryRunProvider) CreateRecord(ctx context.Context, rec Record) error {
	if p.dryRun {
		priority := ""
		if rec.Priority != nil {
			priority = fmt.Sprintf(" (priority: %d)", *rec.Priority)
		}
		fmt.Printf("[DRY RUN] Would create %s record: %s -> %s%s\n", rec.Type, rec.Name, rec.Content, priority)
		return nil
	}
	return p.underlying.CreateRecord(ctx, rec)
}

// UpdateRecord updates a DNS record (or logs what would be updated in dry-run mode).
func (p *DryRunProvider) UpdateRecord(ctx context.Context, recordID string, rec Record) error {
	if p.dryRun {
		priority := ""
		if rec.Priority != nil {
			priority = fmt.Sprintf(" (priority: %d)", *rec.Priority)
		}
		fmt.Printf("[DRY RUN] Would update %s record (ID: %s): %s -> %s%s\n", rec.Type, recordID, rec.Name, rec.Content, priority)
		return nil
	}
	return p.underlying.UpdateRecord(ctx, recordID, rec)
}

// DeleteRecord deletes a DNS record (or logs what would be deleted in dry-run mode).
func (p *DryRunProvider) DeleteRecord(ctx context.Context, recordID string) error {
	if p.dryRun {
		fmt.Printf("[DRY RUN] Would delete record (ID: %s)\n", recordID)
		return nil
	}
	return p.underlying.DeleteRecord(ctx, recordID)
}

// ListRecords always passes through to the underlying provider (reads are safe).
func (p *DryRunProvider) ListRecords(ctx context.Context, filter RecordFilter) ([]Record, error) {
	return p.underlying.ListRecords(ctx, filter)
}

// GetZoneID always passes through to the underlying provider (reads are safe).
func (p *DryRunProvider) GetZoneID(ctx context.Context, domain string) (string, error) {
	return p.underlying.GetZoneID(ctx, domain)
}

// Name returns the underlying provider's name.
func (p *DryRunProvider) Name() string {
	return p.underlying.Name()
}

// IsDryRun returns whether the provider is in dry-run mode.
func (p *DryRunProvider) IsDryRun() bool {
	return p.dryRun
}
