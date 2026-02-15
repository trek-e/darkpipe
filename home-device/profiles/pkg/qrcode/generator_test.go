// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package qrcode

import (
	"strings"
	"testing"
)

func TestGenerateQRCode(t *testing.T) {
	store := NewMemoryTokenStore()
	profileBaseURL := "mail.example.com"
	email := "test@example.com"

	url, err := GenerateQRCode(profileBaseURL, email, store)
	if err != nil {
		t.Fatalf("GenerateQRCode failed: %v", err)
	}

	// Verify URL format
	expectedPrefix := "https://mail.example.com/profile/download?token="
	if !strings.HasPrefix(url, expectedPrefix) {
		t.Errorf("URL format incorrect: got %s, want prefix %s", url, expectedPrefix)
	}

	// Extract token from URL
	token := strings.TrimPrefix(url, expectedPrefix)
	if token == "" {
		t.Error("Token not found in URL")
	}

	// Verify token exists in store and is valid
	validEmail, valid, err := store.Validate(token)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if !valid {
		t.Error("Generated token should be valid")
	}

	if validEmail != email {
		t.Errorf("Email mismatch: got %s, want %s", validEmail, email)
	}
}

func TestGenerateQRCodePNG(t *testing.T) {
	url := "https://mail.example.com/profile/download?token=test123"

	// Test with default size
	png, err := GenerateQRCodePNG(url, 0)
	if err != nil {
		t.Fatalf("GenerateQRCodePNG failed: %v", err)
	}

	if len(png) == 0 {
		t.Error("GenerateQRCodePNG returned empty bytes")
	}

	// Verify PNG magic number (first 8 bytes)
	if len(png) < 8 {
		t.Fatal("PNG too short to verify magic number")
	}

	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	for i := 0; i < 8; i++ {
		if png[i] != pngMagic[i] {
			t.Errorf("PNG magic number mismatch at byte %d: got %x, want %x", i, png[i], pngMagic[i])
		}
	}

	// Test with custom size
	png256, err := GenerateQRCodePNG(url, 256)
	if err != nil {
		t.Fatalf("GenerateQRCodePNG with size 256 failed: %v", err)
	}

	if len(png256) == 0 {
		t.Error("GenerateQRCodePNG with size 256 returned empty bytes")
	}

	// Test with larger size
	png512, err := GenerateQRCodePNG(url, 512)
	if err != nil {
		t.Fatalf("GenerateQRCodePNG with size 512 failed: %v", err)
	}

	if len(png512) <= len(png256) {
		t.Error("Larger QR code should produce larger PNG")
	}
}

func TestGenerateQRCodeTerminal(t *testing.T) {
	url := "https://mail.example.com/profile/download?token=test123"

	ascii, err := GenerateQRCodeTerminal(url)
	if err != nil {
		t.Fatalf("GenerateQRCodeTerminal failed: %v", err)
	}

	if ascii == "" {
		t.Error("GenerateQRCodeTerminal returned empty string")
	}

	// ASCII art should contain block characters or spaces
	// Just verify we got something reasonable
	if len(ascii) < 50 {
		t.Errorf("ASCII art too short: got %d chars", len(ascii))
	}

	// Should contain newlines (multi-line output)
	if !strings.Contains(ascii, "\n") {
		t.Error("ASCII art should contain newlines")
	}
}

func TestGenerateQRCodePNGWithDifferentURLs(t *testing.T) {
	url1 := "https://mail.example.com/profile/download?token=abc123"
	url2 := "https://mail.example.com/profile/download?token=xyz789"

	png1, err := GenerateQRCodePNG(url1, 256)
	if err != nil {
		t.Fatalf("GenerateQRCodePNG for url1 failed: %v", err)
	}

	png2, err := GenerateQRCodePNG(url2, 256)
	if err != nil {
		t.Fatalf("GenerateQRCodePNG for url2 failed: %v", err)
	}

	// Different URLs should produce different QR codes
	if len(png1) == len(png2) {
		// Check if content is different (not just length)
		same := true
		for i := 0; i < len(png1); i++ {
			if png1[i] != png2[i] {
				same = false
				break
			}
		}
		if same {
			t.Error("Different URLs should produce different QR codes")
		}
	}
}
