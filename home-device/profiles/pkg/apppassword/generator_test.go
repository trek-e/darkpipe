// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package apppassword

import (
	"regexp"
	"strings"
	"testing"
)

func TestGenerateAppPassword_Format(t *testing.T) {
	password, err := GenerateAppPassword()
	if err != nil {
		t.Fatalf("GenerateAppPassword failed: %v", err)
	}

	// Check format: XXXX-XXXX-XXXX-XXXX
	pattern := regexp.MustCompile(`^[A-Z2-9]{4}-[A-Z2-9]{4}-[A-Z2-9]{4}-[A-Z2-9]{4}$`)
	if !pattern.MatchString(password) {
		t.Errorf("Password format incorrect: %s", password)
	}

	// Check length (19 chars including hyphens)
	if len(password) != 19 {
		t.Errorf("Password length incorrect: got %d, want 19", len(password))
	}

	// Check hyphen positions
	if password[4] != '-' || password[9] != '-' || password[14] != '-' {
		t.Errorf("Hyphens in wrong positions: %s", password)
	}
}

func TestGenerateAppPassword_Charset(t *testing.T) {
	password, err := GenerateAppPassword()
	if err != nil {
		t.Fatalf("GenerateAppPassword failed: %v", err)
	}

	// Remove hyphens for charset check
	chars := strings.ReplaceAll(password, "-", "")

	// Check all characters are in charset
	for _, ch := range chars {
		if !strings.ContainsRune(charset, ch) {
			t.Errorf("Invalid character in password: %c", ch)
		}
	}

	// Check no confusing characters (0, O, 1, I)
	confusing := "01OI"
	for _, ch := range chars {
		if strings.ContainsRune(confusing, ch) {
			t.Errorf("Password contains confusing character: %c", ch)
		}
	}
}

func TestGenerateAppPassword_Uniqueness(t *testing.T) {
	// Generate 100 passwords and check for duplicates
	passwords := make(map[string]bool)
	for i := 0; i < 100; i++ {
		password, err := GenerateAppPassword()
		if err != nil {
			t.Fatalf("GenerateAppPassword failed on iteration %d: %v", i, err)
		}
		if passwords[password] {
			t.Errorf("Duplicate password generated: %s", password)
		}
		passwords[password] = true
	}
}

func TestHashPassword_Roundtrip(t *testing.T) {
	password := "TEST-PASS-WORD-1234"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("Hash is empty")
	}

	// Verify correct password
	if !VerifyPassword(password, hash) {
		t.Error("VerifyPassword failed for correct password")
	}
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	password := "TEST-PASS-WORD-1234"
	wrongPassword := "WRONG-PASS-WORD-5678"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Verify wrong password fails
	if VerifyPassword(wrongPassword, hash) {
		t.Error("VerifyPassword succeeded for wrong password")
	}
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	password := "TEST-PASS-WORD-1234"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("First HashPassword failed: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Second HashPassword failed: %v", err)
	}

	// Bcrypt should produce different salts/hashes
	if hash1 == hash2 {
		t.Error("Two hashes of same password are identical (salt issue)")
	}

	// But both should verify
	if !VerifyPassword(password, hash1) || !VerifyPassword(password, hash2) {
		t.Error("Hash verification failed")
	}
}
