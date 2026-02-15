// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package queue provides encrypted in-memory message queuing with background
// delivery for handling offline home devices.
package queue

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"io"
	"os"

	"filippo.io/age"
)

// GenerateIdentity creates a new age X25519 keypair.
func GenerateIdentity() (*age.X25519Identity, error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("generate identity: %w", err)
	}
	return identity, nil
}

// LoadOrCreateIdentity loads an age identity from the specified path,
// or creates a new one if the file doesn't exist.
// The identity file is created with 0600 permissions to protect the private key.
func LoadOrCreateIdentity(path string) (*age.X25519Identity, error) {
	// Try to load existing identity
	data, err := os.ReadFile(path)
	if err == nil {
		// Parse existing identity
		identity, err := age.ParseX25519Identity(string(data))
		if err != nil {
			return nil, fmt.Errorf("parse identity from %s: %w", path, err)
		}
		return identity, nil
	}

	// If file doesn't exist, generate new identity
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read identity file %s: %w", path, err)
	}

	// Generate new identity
	identity, err := GenerateIdentity()
	if err != nil {
		return nil, err
	}

	// Write to file with 0600 permissions
	if err := os.WriteFile(path, []byte(identity.String()), 0600); err != nil {
		return nil, fmt.Errorf("write identity to %s: %w", path, err)
	}

	return identity, nil
}

// Encrypt encrypts data using age and returns the ciphertext and CRC32 checksum.
func Encrypt(data []byte, recipient age.Recipient) ([]byte, uint32, error) {
	buf := &bytes.Buffer{}

	w, err := age.Encrypt(buf, recipient)
	if err != nil {
		return nil, 0, fmt.Errorf("create age encryptor: %w", err)
	}

	if _, err := w.Write(data); err != nil {
		return nil, 0, fmt.Errorf("write plaintext: %w", err)
	}

	if err := w.Close(); err != nil {
		return nil, 0, fmt.Errorf("close age encryptor: %w", err)
	}

	ciphertext := buf.Bytes()
	checksum := crc32.ChecksumIEEE(ciphertext)

	return ciphertext, checksum, nil
}

// Decrypt decrypts age-encrypted data after verifying the CRC32 checksum.
// Returns an error if the checksum doesn't match or decryption fails.
func Decrypt(encrypted []byte, checksum uint32, identity age.Identity) ([]byte, error) {
	// Fast rejection: verify CRC32 first
	actualChecksum := crc32.ChecksumIEEE(encrypted)
	if actualChecksum != checksum {
		return nil, fmt.Errorf("checksum mismatch: expected %d, got %d", checksum, actualChecksum)
	}

	// Decrypt
	r, err := age.Decrypt(bytes.NewReader(encrypted), identity)
	if err != nil {
		// Poly1305 MAC failure indicates corruption or tampering
		return nil, fmt.Errorf("age decrypt (possible corruption/tampering): %w", err)
	}

	plaintext, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read decrypted data: %w", err)
	}

	return plaintext, nil
}
