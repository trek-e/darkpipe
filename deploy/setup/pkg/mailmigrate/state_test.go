package mailmigrate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewMigrationState_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "nonexistent.json")

	state, err := NewMigrationState(statePath)
	if err != nil {
		t.Fatalf("NewMigrationState on missing file failed: %v", err)
	}

	// Verify empty state is returned
	stats := state.Stats()
	if stats.Messages != 0 {
		t.Errorf("new state Messages count = %d, want 0", stats.Messages)
	}
	if stats.CalEvents != 0 {
		t.Errorf("new state CalEvents count = %d, want 0", stats.CalEvents)
	}
	if stats.Contacts != 0 {
		t.Errorf("new state Contacts count = %d, want 0", stats.Contacts)
	}
	if stats.Folders != 0 {
		t.Errorf("new state Folders count = %d, want 0", stats.Folders)
	}
}

func TestSaveAndLoadState_Roundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	// Create state with some data
	state, err := NewMigrationState(statePath)
	if err != nil {
		t.Fatalf("NewMigrationState failed: %v", err)
	}

	// Add some data
	if err := state.MarkMessageMigrated("<msg1@example.com>"); err != nil {
		t.Fatalf("MarkMessageMigrated failed: %v", err)
	}
	if err := state.MarkMessageMigrated("<msg2@example.com>"); err != nil {
		t.Fatalf("MarkMessageMigrated failed: %v", err)
	}
	if err := state.MarkCalEventMigrated("event-uid-123"); err != nil {
		t.Fatalf("MarkCalEventMigrated failed: %v", err)
	}
	if err := state.MarkContactMigrated("alice@example.com"); err != nil {
		t.Fatalf("MarkContactMigrated failed: %v", err)
	}
	if err := state.MarkFolderMigrated("INBOX"); err != nil {
		t.Fatalf("MarkFolderMigrated failed: %v", err)
	}

	// Save manually
	if err := state.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file permissions
	info, err := os.Stat(statePath)
	if err != nil {
		t.Fatalf("stat state file failed: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("state file permissions = %o, want 0600", info.Mode().Perm())
	}

	// Load in new state instance
	loadedState, err := NewMigrationState(statePath)
	if err != nil {
		t.Fatalf("NewMigrationState (load) failed: %v", err)
	}

	// Verify data integrity
	stats := loadedState.Stats()
	if stats.Messages != 2 {
		t.Errorf("loaded Messages count = %d, want 2", stats.Messages)
	}
	if stats.CalEvents != 1 {
		t.Errorf("loaded CalEvents count = %d, want 1", stats.CalEvents)
	}
	if stats.Contacts != 1 {
		t.Errorf("loaded Contacts count = %d, want 1", stats.Contacts)
	}
	if stats.Folders != 1 {
		t.Errorf("loaded Folders count = %d, want 1", stats.Folders)
	}

	// Verify specific entries
	if !loadedState.IsMessageMigrated("<msg1@example.com>") {
		t.Error("loaded state missing msg1")
	}
	if !loadedState.IsMessageMigrated("<msg2@example.com>") {
		t.Error("loaded state missing msg2")
	}
	if !loadedState.IsCalEventMigrated("event-uid-123") {
		t.Error("loaded state missing event")
	}
	if !loadedState.IsContactMigrated("alice@example.com") {
		t.Error("loaded state missing contact")
	}
	if !loadedState.IsFolderMigrated("INBOX") {
		t.Error("loaded state missing folder")
	}
}

func TestIsMessageMigrated_AndMarkMessageMigrated(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	state, err := NewMigrationState(statePath)
	if err != nil {
		t.Fatalf("NewMigrationState failed: %v", err)
	}

	messageID := "<test123@example.com>"

	// Initially not migrated
	if state.IsMessageMigrated(messageID) {
		t.Error("message should not be migrated initially")
	}

	// Mark migrated
	if err := state.MarkMessageMigrated(messageID); err != nil {
		t.Fatalf("MarkMessageMigrated failed: %v", err)
	}

	// Now should be migrated
	if !state.IsMessageMigrated(messageID) {
		t.Error("message should be migrated after marking")
	}

	// Different message ID should not be migrated
	if state.IsMessageMigrated("<other@example.com>") {
		t.Error("different message should not be migrated")
	}
}

func TestCalEvent_AndContact_AndFolder_Methods(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	state, err := NewMigrationState(statePath)
	if err != nil {
		t.Fatalf("NewMigrationState failed: %v", err)
	}

	// Test calendar events
	eventUID := "event-123"
	if state.IsCalEventMigrated(eventUID) {
		t.Error("event should not be migrated initially")
	}
	if err := state.MarkCalEventMigrated(eventUID); err != nil {
		t.Fatalf("MarkCalEventMigrated failed: %v", err)
	}
	if !state.IsCalEventMigrated(eventUID) {
		t.Error("event should be migrated after marking")
	}

	// Test contacts
	email := "alice@example.com"
	if state.IsContactMigrated(email) {
		t.Error("contact should not be migrated initially")
	}
	if err := state.MarkContactMigrated(email); err != nil {
		t.Fatalf("MarkContactMigrated failed: %v", err)
	}
	if !state.IsContactMigrated(email) {
		t.Error("contact should be migrated after marking")
	}

	// Test folders
	folder := "INBOX"
	if state.IsFolderMigrated(folder) {
		t.Error("folder should not be migrated initially")
	}
	if err := state.MarkFolderMigrated(folder); err != nil {
		t.Fatalf("MarkFolderMigrated failed: %v", err)
	}
	if !state.IsFolderMigrated(folder) {
		t.Error("folder should be migrated after marking")
	}
}

func TestStats(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	state, err := NewMigrationState(statePath)
	if err != nil {
		t.Fatalf("NewMigrationState failed: %v", err)
	}

	// Add some items
	state.MarkMessageMigrated("<msg1@example.com>")
	state.MarkMessageMigrated("<msg2@example.com>")
	state.MarkCalEventMigrated("event-1")
	state.MarkContactMigrated("alice@example.com")
	state.MarkFolderMigrated("INBOX")

	// Get stats
	stats := state.Stats()

	// Verify counts
	if stats.Messages != 2 {
		t.Errorf("Stats Messages = %d, want 2", stats.Messages)
	}
	if stats.CalEvents != 1 {
		t.Errorf("Stats CalEvents = %d, want 1", stats.CalEvents)
	}
	if stats.Contacts != 1 {
		t.Errorf("Stats Contacts = %d, want 1", stats.Contacts)
	}
	if stats.Folders != 1 {
		t.Errorf("Stats Folders = %d, want 1", stats.Folders)
	}
}

func TestAutoSave(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	state, err := NewMigrationState(statePath)
	if err != nil {
		t.Fatalf("NewMigrationState failed: %v", err)
	}

	// Auto-save triggers every 100 operations by default
	// Add 100 messages to trigger auto-save
	for i := 0; i < 100; i++ {
		messageID := "<msg" + string(rune(i)) + "@example.com>"
		if err := state.MarkMessageMigrated(messageID); err != nil {
			t.Fatalf("MarkMessageMigrated failed: %v", err)
		}
	}

	// Verify file was auto-saved
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("state file should exist after 100 operations (auto-save)")
	}

	// Load and verify count
	loadedState, err := NewMigrationState(statePath)
	if err != nil {
		t.Fatalf("NewMigrationState (load) failed: %v", err)
	}

	stats := loadedState.Stats()
	if stats.Messages != 100 {
		t.Errorf("auto-saved Messages count = %d, want 100", stats.Messages)
	}
}
