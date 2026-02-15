// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package provider

import (
	"context"
	"testing"
)

// mockProvider is a test implementation of DNSProvider
type mockProvider struct {
	createCalled bool
	updateCalled bool
	deleteCalled bool
	listCalled   bool
	zoneCalled   bool
}

func (m *mockProvider) CreateRecord(ctx context.Context, rec Record) error {
	m.createCalled = true
	return nil
}

func (m *mockProvider) UpdateRecord(ctx context.Context, recordID string, rec Record) error {
	m.updateCalled = true
	return nil
}

func (m *mockProvider) ListRecords(ctx context.Context, filter RecordFilter) ([]Record, error) {
	m.listCalled = true
	return []Record{{ID: "test-1", Type: "TXT", Name: "@", Content: "test"}}, nil
}

func (m *mockProvider) DeleteRecord(ctx context.Context, recordID string) error {
	m.deleteCalled = true
	return nil
}

func (m *mockProvider) GetZoneID(ctx context.Context, domain string) (string, error) {
	m.zoneCalled = true
	return "zone-123", nil
}

func (m *mockProvider) Name() string {
	return "mock"
}

func TestDryRunProvider_InterceptsWrites(t *testing.T) {
	mock := &mockProvider{}
	dryRun := NewDryRunProvider(mock, true)

	ctx := context.Background()
	priority := 10

	// Test CreateRecord in dry-run mode
	err := dryRun.CreateRecord(ctx, Record{Type: "TXT", Name: "@", Content: "v=spf1 -all"})
	if err != nil {
		t.Errorf("CreateRecord returned error: %v", err)
	}
	if mock.createCalled {
		t.Error("CreateRecord should not call underlying provider in dry-run mode")
	}

	// Test UpdateRecord in dry-run mode
	err = dryRun.UpdateRecord(ctx, "rec-123", Record{Type: "MX", Name: "@", Content: "mail.example.com", Priority: &priority})
	if err != nil {
		t.Errorf("UpdateRecord returned error: %v", err)
	}
	if mock.updateCalled {
		t.Error("UpdateRecord should not call underlying provider in dry-run mode")
	}

	// Test DeleteRecord in dry-run mode
	err = dryRun.DeleteRecord(ctx, "rec-456")
	if err != nil {
		t.Errorf("DeleteRecord returned error: %v", err)
	}
	if mock.deleteCalled {
		t.Error("DeleteRecord should not call underlying provider in dry-run mode")
	}
}

func TestDryRunProvider_PassesReads(t *testing.T) {
	mock := &mockProvider{}
	dryRun := NewDryRunProvider(mock, true)

	ctx := context.Background()

	// Test ListRecords passes through even in dry-run mode
	records, err := dryRun.ListRecords(ctx, RecordFilter{Type: "TXT"})
	if err != nil {
		t.Errorf("ListRecords returned error: %v", err)
	}
	if !mock.listCalled {
		t.Error("ListRecords should call underlying provider even in dry-run mode")
	}
	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}

	// Test GetZoneID passes through even in dry-run mode
	zoneID, err := dryRun.GetZoneID(ctx, "example.com")
	if err != nil {
		t.Errorf("GetZoneID returned error: %v", err)
	}
	if !mock.zoneCalled {
		t.Error("GetZoneID should call underlying provider even in dry-run mode")
	}
	if zoneID != "zone-123" {
		t.Errorf("Expected zone-123, got %s", zoneID)
	}
}

func TestDryRunProvider_ApplyMode(t *testing.T) {
	mock := &mockProvider{}
	dryRun := NewDryRunProvider(mock, false) // Apply mode (not dry-run)

	ctx := context.Background()

	// Test CreateRecord in apply mode
	err := dryRun.CreateRecord(ctx, Record{Type: "TXT", Name: "@", Content: "test"})
	if err != nil {
		t.Errorf("CreateRecord returned error: %v", err)
	}
	if !mock.createCalled {
		t.Error("CreateRecord should call underlying provider in apply mode")
	}

	// Test UpdateRecord in apply mode
	err = dryRun.UpdateRecord(ctx, "rec-123", Record{Type: "TXT", Name: "@", Content: "test"})
	if err != nil {
		t.Errorf("UpdateRecord returned error: %v", err)
	}
	if !mock.updateCalled {
		t.Error("UpdateRecord should call underlying provider in apply mode")
	}

	// Test DeleteRecord in apply mode
	err = dryRun.DeleteRecord(ctx, "rec-456")
	if err != nil {
		t.Errorf("DeleteRecord returned error: %v", err)
	}
	if !mock.deleteCalled {
		t.Error("DeleteRecord should call underlying provider in apply mode")
	}
}

func TestDryRunProvider_IsDryRun(t *testing.T) {
	mock := &mockProvider{}

	dryRunTrue := NewDryRunProvider(mock, true)
	if !dryRunTrue.IsDryRun() {
		t.Error("IsDryRun should return true when created with dryRun=true")
	}

	dryRunFalse := NewDryRunProvider(mock, false)
	if dryRunFalse.IsDryRun() {
		t.Error("IsDryRun should return false when created with dryRun=false")
	}
}

func TestDryRunProvider_Name(t *testing.T) {
	mock := &mockProvider{}
	dryRun := NewDryRunProvider(mock, true)

	if name := dryRun.Name(); name != "mock" {
		t.Errorf("Expected name 'mock', got '%s'", name)
	}
}
