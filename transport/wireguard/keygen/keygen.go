// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package keygen

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GenerateKeyPair generates a WireGuard private/public keypair using the wg CLI.
// Returns base64-encoded private and public keys.
// Requires wireguard-tools to be installed (wg command available).
func GenerateKeyPair() (privateKey, publicKey string, err error) {
	// Check if wg command is available
	if _, err := exec.LookPath("wg"); err != nil {
		return "", "", fmt.Errorf("wg command not found: install wireguard-tools (apt install wireguard-tools)")
	}

	// Generate private key using wg genkey
	genKeyCmd := exec.Command("wg", "genkey")
	var privKeyOut bytes.Buffer
	genKeyCmd.Stdout = &privKeyOut
	if err := genKeyCmd.Run(); err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}
	privateKey = strings.TrimSpace(privKeyOut.String())

	// Generate public key from private key using wg pubkey
	pubKeyCmd := exec.Command("wg", "pubkey")
	pubKeyCmd.Stdin = strings.NewReader(privateKey)
	var pubKeyOut bytes.Buffer
	pubKeyCmd.Stdout = &pubKeyOut
	if err := pubKeyCmd.Run(); err != nil {
		return "", "", fmt.Errorf("failed to generate public key: %w", err)
	}
	publicKey = strings.TrimSpace(pubKeyOut.String())

	return privateKey, publicKey, nil
}

// GeneratePreSharedKey generates a WireGuard pre-shared key for optional PSK support.
// Returns base64-encoded pre-shared key.
// Requires wireguard-tools to be installed (wg command available).
func GeneratePreSharedKey() (string, error) {
	// Check if wg command is available
	if _, err := exec.LookPath("wg"); err != nil {
		return "", fmt.Errorf("wg command not found: install wireguard-tools (apt install wireguard-tools)")
	}

	// Generate pre-shared key using wg genpsk
	genPskCmd := exec.Command("wg", "genpsk")
	var pskOut bytes.Buffer
	genPskCmd.Stdout = &pskOut
	if err := genPskCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to generate pre-shared key: %w", err)
	}

	return strings.TrimSpace(pskOut.String()), nil
}
