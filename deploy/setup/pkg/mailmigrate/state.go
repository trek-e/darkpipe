// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package mailmigrate

import (
	"encoding/json"
	"os"
	"sync"
)

// MigrationState tracks which items have been migrated to support resume.
type MigrationState struct {
	mu               sync.RWMutex
	CalEvents        map[string]bool `json:"cal_events"`        // UID -> migrated
	Contacts         map[string]bool `json:"contacts"`          // email -> migrated
	Messages         map[string]bool `json:"messages"`          // Message-ID -> migrated
	Folders          map[string]bool `json:"folders"`           // folder path -> migrated
	path             string
	autoSaveInterval int // save every N operations (0 = manual only)
	opsSinceLastSave int
}

// NewMigrationState creates a new migration state.
// If path exists, loads existing state; otherwise starts fresh.
func NewMigrationState(path string) (*MigrationState, error) {
	s := &MigrationState{
		CalEvents:        make(map[string]bool),
		Contacts:         make(map[string]bool),
		Messages:         make(map[string]bool),
		Folders:          make(map[string]bool),
		path:             path,
		autoSaveInterval: 100, // save every 100 operations by default
	}

	// Try to load existing state
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Fresh start - no error
			return s, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, s); err != nil {
		return nil, err
	}

	return s, nil
}

// IsCalEventMigrated checks if a calendar event has been migrated.
func (s *MigrationState) IsCalEventMigrated(uid string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CalEvents[uid]
}

// MarkCalEventMigrated marks a calendar event as migrated.
func (s *MigrationState) MarkCalEventMigrated(uid string) error {
	s.mu.Lock()
	s.CalEvents[uid] = true
	s.opsSinceLastSave++
	shouldSave := s.autoSaveInterval > 0 && s.opsSinceLastSave >= s.autoSaveInterval
	s.mu.Unlock()

	if shouldSave {
		return s.Save()
	}
	return nil
}

// IsContactMigrated checks if a contact has been migrated.
func (s *MigrationState) IsContactMigrated(email string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Contacts[email]
}

// MarkContactMigrated marks a contact as migrated.
func (s *MigrationState) MarkContactMigrated(email string) error {
	s.mu.Lock()
	s.Contacts[email] = true
	s.opsSinceLastSave++
	shouldSave := s.autoSaveInterval > 0 && s.opsSinceLastSave >= s.autoSaveInterval
	s.mu.Unlock()

	if shouldSave {
		return s.Save()
	}
	return nil
}

// IsMessageMigrated checks if a message has been migrated.
func (s *MigrationState) IsMessageMigrated(messageID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Messages[messageID]
}

// MarkMessageMigrated marks a message as migrated.
func (s *MigrationState) MarkMessageMigrated(messageID string) error {
	s.mu.Lock()
	s.Messages[messageID] = true
	s.opsSinceLastSave++
	shouldSave := s.autoSaveInterval > 0 && s.opsSinceLastSave >= s.autoSaveInterval
	s.mu.Unlock()

	if shouldSave {
		return s.Save()
	}
	return nil
}

// IsFolderMigrated checks if a folder has been migrated.
func (s *MigrationState) IsFolderMigrated(path string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Folders[path]
}

// MarkFolderMigrated marks a folder as migrated.
func (s *MigrationState) MarkFolderMigrated(path string) error {
	s.mu.Lock()
	s.Folders[path] = true
	s.opsSinceLastSave++
	shouldSave := s.autoSaveInterval > 0 && s.opsSinceLastSave >= s.autoSaveInterval
	s.mu.Unlock()

	if shouldSave {
		return s.Save()
	}
	return nil
}

// Save writes the state to disk.
func (s *MigrationState) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	s.opsSinceLastSave = 0
	return os.WriteFile(s.path, data, 0600)
}

// Stats returns migration statistics.
func (s *MigrationState) Stats() MigrationStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return MigrationStats{
		CalEvents: len(s.CalEvents),
		Contacts:  len(s.Contacts),
		Messages:  len(s.Messages),
		Folders:   len(s.Folders),
	}
}

// MigrationStats holds migration statistics.
type MigrationStats struct {
	CalEvents int
	Contacts  int
	Messages  int
	Folders   int
}
