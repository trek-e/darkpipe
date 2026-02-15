// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package apppassword

import (
	"time"
)

// AppPassword represents a stored app-specific password.
type AppPassword struct {
	ID             string    // UUIDv7
	Email          string    // User email (e.g., user@example.com)
	DeviceName     string    // Human label (e.g., "iPhone 15", "Thunderbird Desktop")
	HashedPassword string    // Bcrypt hash
	CreatedAt      time.Time
	LastUsedAt     time.Time // Updated on successful auth
}

// Store defines the interface for app password storage backends.
type Store interface {
	// Create generates and stores a new app password.
	Create(email, deviceName, plainPassword string) (*AppPassword, error)

	// List returns all app passwords for a given email.
	List(email string) ([]AppPassword, error)

	// Revoke removes an app password by ID.
	Revoke(id string) error

	// Verify checks if a plain password is valid for the given email.
	// Returns true if valid, updates LastUsedAt on success.
	Verify(email, plainPassword string) (bool, error)
}
