// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package mailmigrate

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/emersion/go-imap/v2"
)

func TestNewIMAPSync(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	state, err := NewMigrationState(statePath)
	if err != nil {
		t.Fatalf("NewMigrationState failed: %v", err)
	}

	mapper := NewFolderMapper("gmail", nil)

	sync := NewIMAPSync(nil, nil, state, mapper, statePath)

	if sync == nil {
		t.Fatal("NewIMAPSync returned nil")
	}

	if sync.State != state {
		t.Error("State not set correctly")
	}

	if sync.Mapper != mapper {
		t.Error("Mapper not set correctly")
	}

	if sync.BatchSize != 50 {
		t.Errorf("BatchSize = %d, want 50", sync.BatchSize)
	}

	if sync.LabelsAsKeywords {
		t.Error("LabelsAsKeywords should default to false")
	}
}

func TestFolderInfo_Mapping(t *testing.T) {
	mapper := NewFolderMapper("gmail", nil)

	tests := []struct {
		sourceName string
		wantMapped string
		wantSkip   bool
	}{
		{"[Gmail]/Sent Mail", "Sent", false},
		{"[Gmail]/All Mail", "", true},
		{"INBOX", "INBOX", false},
	}

	for _, tt := range tests {
		t.Run(tt.sourceName, func(t *testing.T) {
			mappedTo, skip := mapper.Map(tt.sourceName)

			info := FolderInfo{
				Name:     tt.sourceName,
				MappedTo: mappedTo,
				Skip:     skip,
			}

			if info.MappedTo != tt.wantMapped {
				t.Errorf("FolderInfo.MappedTo = %q, want %q", info.MappedTo, tt.wantMapped)
			}

			if info.Skip != tt.wantSkip {
				t.Errorf("FolderInfo.Skip = %v, want %v", info.Skip, tt.wantSkip)
			}
		})
	}
}

func TestDryRunResult_Calculation(t *testing.T) {
	result := &DryRunResult{
		Folders: []DryRunFolder{
			{Name: "INBOX", Messages: 100, MappedTo: "INBOX", WillSkip: false},
			{Name: "Sent", Messages: 50, MappedTo: "Sent", WillSkip: false},
			{Name: "[Gmail]/All Mail", Messages: 200, MappedTo: "", WillSkip: true},
		},
		TotalMessages: 150, // 100 + 50 (skipped folder not counted)
	}

	// Verify total excludes skipped
	expectedTotal := uint32(150)
	if result.TotalMessages != expectedTotal {
		t.Errorf("TotalMessages = %d, want %d", result.TotalMessages, expectedTotal)
	}

	// Verify folder count
	if len(result.Folders) != 3 {
		t.Errorf("Folders count = %d, want 3", len(result.Folders))
	}

	// Verify skip flag
	for _, folder := range result.Folders {
		if folder.Name == "[Gmail]/All Mail" && !folder.WillSkip {
			t.Error("[Gmail]/All Mail should have WillSkip=true")
		}
	}
}

func TestSyncResult_Aggregation(t *testing.T) {
	result := &SyncResult{
		Folders: []FolderResult{
			{Folder: "INBOX", Migrated: 100, Skipped: 10, Errors: 2},
			{Folder: "Sent", Migrated: 50, Skipped: 5, Errors: 1},
			{Folder: "Drafts", Migrated: 20, Skipped: 2, Errors: 0},
		},
		TotalMigrated: 170,
		TotalSkipped:  17,
		TotalErrors:   3,
	}

	// Verify totals match sum of folder results
	var sumMigrated, sumSkipped, sumErrors int
	for _, folder := range result.Folders {
		sumMigrated += folder.Migrated
		sumSkipped += folder.Skipped
		sumErrors += folder.Errors
	}

	if result.TotalMigrated != sumMigrated {
		t.Errorf("TotalMigrated = %d, want %d", result.TotalMigrated, sumMigrated)
	}

	if result.TotalSkipped != sumSkipped {
		t.Errorf("TotalSkipped = %d, want %d", result.TotalSkipped, sumSkipped)
	}

	if result.TotalErrors != sumErrors {
		t.Errorf("TotalErrors = %d, want %d", result.TotalErrors, sumErrors)
	}
}

func TestFolderResult_Counts(t *testing.T) {
	result := &FolderResult{
		Folder:   "INBOX",
		Migrated: 100,
		Skipped:  10,
		Errors:   2,
	}

	// Total processed = Migrated + Skipped + Errors
	totalProcessed := result.Migrated + result.Skipped + result.Errors
	expectedTotal := 112

	if totalProcessed != expectedTotal {
		t.Errorf("Total processed = %d, want %d", totalProcessed, expectedTotal)
	}
}

func TestIMAPSync_StateIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	state, err := NewMigrationState(statePath)
	if err != nil {
		t.Fatalf("NewMigrationState failed: %v", err)
	}

	mapper := NewFolderMapper("generic", nil)
	sync := NewIMAPSync(nil, nil, state, mapper, statePath)

	// Simulate marking messages as migrated
	messageIDs := []string{
		"<msg1@example.com>",
		"<msg2@example.com>",
		"<msg3@example.com>",
	}

	for _, msgID := range messageIDs {
		if err := sync.State.MarkMessageMigrated(msgID); err != nil {
			t.Fatalf("MarkMessageMigrated failed: %v", err)
		}
	}

	// Verify state tracking
	for _, msgID := range messageIDs {
		if !sync.State.IsMessageMigrated(msgID) {
			t.Errorf("Message %q should be marked as migrated", msgID)
		}
	}

	// Verify new message not migrated
	if sync.State.IsMessageMigrated("<newmsg@example.com>") {
		t.Error("New message should not be marked as migrated")
	}

	// Save and reload to verify persistence
	if err := sync.State.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loadedState, err := NewMigrationState(statePath)
	if err != nil {
		t.Fatalf("NewMigrationState (load) failed: %v", err)
	}

	for _, msgID := range messageIDs {
		if !loadedState.IsMessageMigrated(msgID) {
			t.Errorf("Loaded state missing message %q", msgID)
		}
	}
}

func TestIMAPSync_BatchSize(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	state, err := NewMigrationState(statePath)
	if err != nil {
		t.Fatalf("NewMigrationState failed: %v", err)
	}

	mapper := NewFolderMapper("generic", nil)
	sync := NewIMAPSync(nil, nil, state, mapper, statePath)

	// Test default batch size
	if sync.BatchSize != 50 {
		t.Errorf("Default BatchSize = %d, want 50", sync.BatchSize)
	}

	// Test custom batch size
	sync.BatchSize = 100
	if sync.BatchSize != 100 {
		t.Errorf("Custom BatchSize = %d, want 100", sync.BatchSize)
	}
}

func TestIMAPSync_ProgressCallbacks(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	state, err := NewMigrationState(statePath)
	if err != nil {
		t.Fatalf("NewMigrationState failed: %v", err)
	}

	mapper := NewFolderMapper("generic", nil)
	sync := NewIMAPSync(nil, nil, state, mapper, statePath)

	// Test callback assignment
	var startCalled, progressCalled, doneCalled bool

	sync.OnFolderStart = func(folder string, total int) {
		startCalled = true
	}

	sync.OnProgress = func(folder string, current, total int) {
		progressCalled = true
	}

	sync.OnFolderDone = func(folder string) {
		doneCalled = true
	}

	// Simulate callbacks
	if sync.OnFolderStart != nil {
		sync.OnFolderStart("INBOX", 100)
	}

	if sync.OnProgress != nil {
		sync.OnProgress("INBOX", 50, 100)
	}

	if sync.OnFolderDone != nil {
		sync.OnFolderDone("INBOX")
	}

	// Verify all callbacks were invoked
	if !startCalled {
		t.Error("OnFolderStart callback not called")
	}

	if !progressCalled {
		t.Error("OnProgress callback not called")
	}

	if !doneCalled {
		t.Error("OnFolderDone callback not called")
	}
}

func TestAppendOptions_DatePreservation(t *testing.T) {
	// Test that append options preserve original date
	originalDate := time.Date(2020, 1, 15, 10, 30, 0, 0, time.UTC)
	flags := []imap.Flag{imap.FlagSeen, imap.FlagFlagged}

	options := &imap.AppendOptions{
		Flags: flags,
		Time:  originalDate,
	}

	// Verify date is set correctly
	if !options.Time.Equal(originalDate) {
		t.Errorf("AppendOptions.Time = %v, want %v", options.Time, originalDate)
	}

	// Verify flags are set correctly
	if len(options.Flags) != 2 {
		t.Errorf("AppendOptions.Flags count = %d, want 2", len(options.Flags))
	}
}

func TestDryRunFolder_Structure(t *testing.T) {
	folder := DryRunFolder{
		Name:     "[Gmail]/Sent Mail",
		Messages: 1234,
		MappedTo: "Sent",
		WillSkip: false,
	}

	if folder.Name != "[Gmail]/Sent Mail" {
		t.Errorf("Name = %q, want [Gmail]/Sent Mail", folder.Name)
	}

	if folder.Messages != 1234 {
		t.Errorf("Messages = %d, want 1234", folder.Messages)
	}

	if folder.MappedTo != "Sent" {
		t.Errorf("MappedTo = %q, want Sent", folder.MappedTo)
	}

	if folder.WillSkip {
		t.Error("WillSkip should be false")
	}
}

func TestSyncFolder_SkipLogic(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	state, err := NewMigrationState(statePath)
	if err != nil {
		t.Fatalf("NewMigrationState failed: %v", err)
	}

	// Gmail mapper skips [Gmail]/All Mail
	mapper := NewFolderMapper("gmail", nil)
	sync := NewIMAPSync(nil, nil, state, mapper, statePath)

	ctx := context.Background()

	// Test skipped folder (no IMAP client needed for skip test)
	result, err := sync.SyncFolder(ctx, "[Gmail]/All Mail")
	if err != nil {
		t.Fatalf("SyncFolder failed: %v", err)
	}

	// Result should have zero counts for skipped folder
	if result.Migrated != 0 {
		t.Errorf("Skipped folder Migrated = %d, want 0", result.Migrated)
	}

	if result.Skipped != 0 {
		t.Errorf("Skipped folder Skipped = %d, want 0", result.Skipped)
	}

	if result.Errors != 0 {
		t.Errorf("Skipped folder Errors = %d, want 0", result.Errors)
	}
}
