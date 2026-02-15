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

// DovecotStore implements Store using a JSON file.
// Used for Postfix+Dovecot backend.
type DovecotStore struct {
	FilePath string
	mu       sync.RWMutex
}

type dovecotStorage struct {
	Passwords []AppPassword `json:"passwords"`
}

// NewDovecotStore creates a file-based store for Dovecot.
func NewDovecotStore(filePath string) *DovecotStore {
	if filePath == "" {
		filePath = "/data/app-passwords.json"
	}
	return &DovecotStore{FilePath: filePath}
}

func (d *DovecotStore) load() (*dovecotStorage, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	file, err := os.Open(d.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &dovecotStorage{Passwords: []AppPassword{}}, nil
		}
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	// Acquire shared lock
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_SH); err != nil {
		return nil, fmt.Errorf("flock: %w", err)
	}
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

	var storage dovecotStorage
	if err := json.NewDecoder(file).Decode(&storage); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	return &storage, nil
}

func (d *DovecotStore) save(storage *dovecotStorage) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(d.FilePath), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	file, err := os.OpenFile(d.FilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
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

func (d *DovecotStore) Create(email, deviceName, plainPassword string) (*AppPassword, error) {
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

	storage, err := d.load()
	if err != nil {
		return nil, fmt.Errorf("load storage: %w", err)
	}

	storage.Passwords = append(storage.Passwords, *appPassword)

	if err := d.save(storage); err != nil {
		return nil, fmt.Errorf("save storage: %w", err)
	}

	return appPassword, nil
}

func (d *DovecotStore) List(email string) ([]AppPassword, error) {
	storage, err := d.load()
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

func (d *DovecotStore) Revoke(id string) error {
	storage, err := d.load()
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

	if err := d.save(storage); err != nil {
		return fmt.Errorf("save storage: %w", err)
	}

	return nil
}

func (d *DovecotStore) Verify(email, plainPassword string) (bool, error) {
	storage, err := d.load()
	if err != nil {
		return false, fmt.Errorf("load storage: %w", err)
	}

	for i, pw := range storage.Passwords {
		if pw.Email == email && VerifyPassword(plainPassword, pw.HashedPassword) {
			// Update LastUsedAt
			storage.Passwords[i].LastUsedAt = time.Now()
			if err := d.save(storage); err != nil {
				// Log error but don't fail verification
				fmt.Fprintf(os.Stderr, "Warning: failed to update LastUsedAt: %v\n", err)
			}
			return true, nil
		}
	}

	return false, nil
}
