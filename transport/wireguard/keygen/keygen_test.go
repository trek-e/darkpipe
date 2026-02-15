// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package keygen

import (
	"os/exec"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	// Skip test if wg binary is not available
	if _, err := exec.LookPath("wg"); err != nil {
		t.Skip("wg command not found, skipping test (install wireguard-tools)")
	}

	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	// Verify private key
	if privateKey == "" {
		t.Error("private key is empty")
	}
	if len(privateKey) != 44 {
		t.Errorf("private key has incorrect length: got %d, want 44", len(privateKey))
	}

	// Verify public key
	if publicKey == "" {
		t.Error("public key is empty")
	}
	if len(publicKey) != 44 {
		t.Errorf("public key has incorrect length: got %d, want 44", len(publicKey))
	}

	// Verify keys are different
	if privateKey == publicKey {
		t.Error("private key and public key are the same")
	}

	t.Logf("Generated keys:\nPrivate: %s\nPublic: %s", privateKey, publicKey)
}

func TestGeneratePreSharedKey(t *testing.T) {
	// Skip test if wg binary is not available
	if _, err := exec.LookPath("wg"); err != nil {
		t.Skip("wg command not found, skipping test (install wireguard-tools)")
	}

	psk, err := GeneratePreSharedKey()
	if err != nil {
		t.Fatalf("GeneratePreSharedKey failed: %v", err)
	}

	// Verify pre-shared key
	if psk == "" {
		t.Error("pre-shared key is empty")
	}
	if len(psk) != 44 {
		t.Errorf("pre-shared key has incorrect length: got %d, want 44", len(psk))
	}

	t.Logf("Generated pre-shared key: %s", psk)
}

func TestGenerateKeyPairWithoutWG(t *testing.T) {
	// This test verifies error handling when wg is not available
	// We can't actually test this without removing wg from PATH
	// So we'll skip if wg is available
	if _, err := exec.LookPath("wg"); err == nil {
		t.Skip("wg command is available, cannot test error case")
	}

	_, _, err := GenerateKeyPair()
	if err == nil {
		t.Error("expected error when wg is not available, got nil")
	}
}
