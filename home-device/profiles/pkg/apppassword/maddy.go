// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package apppassword

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
)

// MaddyStore implements Store using a JSON file.
// Used for Maddy backend.
type MaddyStore struct {
	FilePath string
	mu       sync.RWMutex
}

type maddyStorage struct {
	Passwords []AppPassword `json:"passwords"`
}

// NewMaddyStore creates a file-based store for Maddy.
func NewMaddyStore(filePath string) *MaddyStore {
	if filePath == "" {
		filePath = "/data/app-passwords.json"
	}
	return &MaddyStore{FilePath: filePath}
}

func (m *MaddyStore) load() (*maddyStorage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	file, err := os.Open(m.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &maddyStorage{Passwords: []AppPassword{}}, nil
		}
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	// Acquire shared lock
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_SH); err != nil {
		return nil, fmt.Errorf("flock: %w", err)
	}
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

	var storage maddyStorage
	if err := json.NewDecoder(file).Decode(&storage); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	return &storage, nil
}

func (m *MaddyStore) save(storage *maddyStorage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(m.FilePath), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	file, err := os.OpenFile(m.FilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	// Acquire exclusive lock
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("flock: %w", err)
	}
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(storage); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	return nil
}

func (m *MaddyStore) Create(email, deviceName, plainPassword string) (*AppPassword, error) {
	hash, err := HashPassword(plainPassword)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	appPassword := &AppPassword{
		ID:             uuid.New().String(),
		Email:          email,
		DeviceName:     deviceName,
		HashedPassword: hash,
		CreatedAt:      time.Now(),
		LastUsedAt:     time.Time{},
	}

	storage, err := m.load()
	if err != nil {
		return nil, fmt.Errorf("load storage: %w", err)
	}

	storage.Passwords = append(storage.Passwords, *appPassword)

	if err := m.save(storage); err != nil {
		return nil, fmt.Errorf("save storage: %w", err)
	}

	return appPassword, nil
}

func (m *MaddyStore) List(email string) ([]AppPassword, error) {
	storage, err := m.load()
	if err != nil {
		return nil, fmt.Errorf("load storage: %w", err)
	}

	var result []AppPassword
	for _, pw := range storage.Passwords {
		if pw.Email == email {
			result = append(result, pw)
		}
	}

	return result, nil
}

func (m *MaddyStore) Revoke(id string) error {
	storage, err := m.load()
	if err != nil {
		return fmt.Errorf("load storage: %w", err)
	}

	var filtered []AppPassword
	found := false
	for _, pw := range storage.Passwords {
		if pw.ID != id {
			filtered = append(filtered, pw)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("app password not found: %s", id)
	}

	storage.Passwords = filtered

	if err := m.save(storage); err != nil {
		return fmt.Errorf("save storage: %w", err)
	}

	return nil
}

func (m *MaddyStore) Verify(email, plainPassword string) (bool, error) {
	storage, err := m.load()
	if err != nil {
		return false, fmt.Errorf("load storage: %w", err)
	}

	for i, pw := range storage.Passwords {
		if pw.Email == email && VerifyPassword(plainPassword, pw.HashedPassword) {
			// Update LastUsedAt
			storage.Passwords[i].LastUsedAt = time.Now()
			if err := m.save(storage); err != nil {
				// Log error but don't fail verification
				fmt.Fprintf(os.Stderr, "Warning: failed to update LastUsedAt: %v\n", err)
			}
			return true, nil
		}
	}

	return false, nil
}
