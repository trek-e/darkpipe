// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package apppassword

import (
	"crypto/rand"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	// Charset excludes confusing characters: no 0/O/1/I
	charset        = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	passwordLength = 16
	groupSize      = 4
)

// GenerateAppPassword generates a cryptographically secure app password
// in XXXX-XXXX-XXXX-XXXX format using crypto/rand.
func GenerateAppPassword() (string, error) {
	bytes := make([]byte, passwordLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("crypto/rand read failed: %w", err)
	}

	// Map random bytes to charset
	for i := 0; i < passwordLength; i++ {
		bytes[i] = charset[int(bytes[i])%len(charset)]
	}

	// Format as XXXX-XXXX-XXXX-XXXX
	password := string(bytes)
	groups := make([]string, 0, passwordLength/groupSize)
	for i := 0; i < passwordLength; i += groupSize {
		groups = append(groups, password[i:i+groupSize])
	}

	return strings.Join(groups, "-"), nil
}

// HashPassword hashes a password using bcrypt with cost 12.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash failed: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against a bcrypt hash.
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
