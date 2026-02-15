// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package qrcode

import (
	"fmt"
	"time"

	"github.com/skip2/go-qrcode"
)

const (
	// DefaultTokenExpiry is the default expiration time for QR code tokens (15 minutes).
	DefaultTokenExpiry = 15 * time.Minute

	// DefaultQRSize is the default QR code image size in pixels.
	DefaultQRSize = 256
)

// GenerateQRCode creates a single-use token and returns the URL to embed in a QR code.
// The URL format is: https://<profileBaseURL>/profile/download?token=<token>
func GenerateQRCode(profileBaseURL, email string, store TokenStore) (string, error) {
	expiresAt := time.Now().Add(DefaultTokenExpiry)
	token, err := store.Create(email, expiresAt)
	if err != nil {
		return "", fmt.Errorf("create token: %w", err)
	}

	url := fmt.Sprintf("https://%s/profile/download?token=%s", profileBaseURL, token)
	return url, nil
}

// GenerateQRCodePNG generates a QR code as PNG bytes.
// Uses skip2/go-qrcode with Medium error correction.
// Default size is 256x256 pixels.
func GenerateQRCodePNG(url string, size int) ([]byte, error) {
	if size <= 0 {
		size = DefaultQRSize
	}

	// Medium error correction (~15% recovery)
	png, err := qrcode.Encode(url, qrcode.Medium, size)
	if err != nil {
		return nil, fmt.Errorf("encode QR code: %w", err)
	}

	return png, nil
}

// GenerateQRCodeTerminal generates a QR code as ASCII art for CLI display.
func GenerateQRCodeTerminal(url string) (string, error) {
	qr, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return "", fmt.Errorf("create QR code: %w", err)
	}

	// ToString produces ASCII art representation
	ascii := qr.ToString(false) // false = don't invert colors
	return ascii, nil
}
