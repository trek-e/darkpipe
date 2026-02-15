// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package queue

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateIdentity(t *testing.T) {
	identity, err := GenerateIdentity()
	if err != nil {
		t.Fatalf("GenerateIdentity() error: %v", err)
	}
	if identity == nil {
		t.Fatal("GenerateIdentity() returned nil identity")
	}

	// Verify we can get the recipient
	recipient := identity.Recipient()
	if recipient == nil {
		t.Fatal("identity.Recipient() returned nil")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	identity, err := GenerateIdentity()
	if err != nil {
		t.Fatalf("GenerateIdentity() error: %v", err)
	}

	plaintext := []byte("This is a test message for encryption")
	recipient := identity.Recipient()

	// Encrypt
	ciphertext, checksum, err := Encrypt(plaintext, recipient)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}
	if len(ciphertext) == 0 {
		t.Fatal("Encrypt() returned empty ciphertext")
	}
	if checksum == 0 {
		t.Fatal("Encrypt() returned zero checksum")
	}

	// Verify ciphertext is different from plaintext
	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("Ciphertext equals plaintext (not encrypted)")
	}

	// Decrypt
	decrypted, err := Decrypt(ciphertext, checksum, identity)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}

	// Verify round-trip
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted data doesn't match original.\nWant: %q\nGot:  %q", plaintext, decrypted)
	}
}

func TestDecryptCorruptedData(t *testing.T) {
	identity, err := GenerateIdentity()
	if err != nil {
		t.Fatalf("GenerateIdentity() error: %v", err)
	}

	plaintext := []byte("Test message")
	recipient := identity.Recipient()

	ciphertext, checksum, err := Encrypt(plaintext, recipient)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	// Corrupt the ciphertext (flip a byte)
	corrupted := make([]byte, len(ciphertext))
	copy(corrupted, ciphertext)
	corrupted[len(corrupted)/2] ^= 0xFF

	// Decrypt should fail with checksum mismatch
	_, err = Decrypt(corrupted, checksum, identity)
	if err == nil {
		t.Fatal("Decrypt() succeeded on corrupted data, expected checksum error")
	}
	// Should get checksum error (fast rejection)
	if !bytes.Contains([]byte(err.Error()), []byte("checksum mismatch")) {
		t.Errorf("Expected checksum mismatch error, got: %v", err)
	}
}

func TestDecryptWrongChecksum(t *testing.T) {
	identity, err := GenerateIdentity()
	if err != nil {
		t.Fatalf("GenerateIdentity() error: %v", err)
	}

	plaintext := []byte("Test message")
	recipient := identity.Recipient()

	ciphertext, _, err := Encrypt(plaintext, recipient)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	// Use wrong checksum
	wrongChecksum := uint32(12345)

	// Should fail with checksum error (fast rejection, doesn't attempt decrypt)
	_, err = Decrypt(ciphertext, wrongChecksum, identity)
	if err == nil {
		t.Fatal("Decrypt() succeeded with wrong checksum, expected error")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("checksum mismatch")) {
		t.Errorf("Expected checksum mismatch error, got: %v", err)
	}
}

func TestLoadOrCreateIdentity(t *testing.T) {
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "identity")

	// First call should create new identity
	identity1, err := LoadOrCreateIdentity(keyPath)
	if err != nil {
		t.Fatalf("LoadOrCreateIdentity() first call error: %v", err)
	}
	if identity1 == nil {
		t.Fatal("LoadOrCreateIdentity() returned nil identity")
	}

	// Verify file was created with correct permissions
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("Stat identity file error: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Identity file permissions = %o, want 0600", info.Mode().Perm())
	}

	// Second call should load existing identity
	identity2, err := LoadOrCreateIdentity(keyPath)
	if err != nil {
		t.Fatalf("LoadOrCreateIdentity() second call error: %v", err)
	}
	if identity2 == nil {
		t.Fatal("LoadOrCreateIdentity() returned nil identity on second call")
	}

	// Verify both identities are the same (can encrypt/decrypt between them)
	plaintext := []byte("Test consistency")
	ciphertext, checksum, err := Encrypt(plaintext, identity1.Recipient())
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, checksum, identity2)
	if err != nil {
		t.Fatalf("Decrypt() with loaded identity error: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("Loaded identity doesn't match created identity")
	}
}

func TestLoadOrCreateIdentity_InvalidPath(t *testing.T) {
	// Try to read from a directory (should fail)
	tempDir := t.TempDir()

	_, err := LoadOrCreateIdentity(tempDir)
	if err == nil {
		t.Fatal("LoadOrCreateIdentity() succeeded on directory path, expected error")
	}
}
