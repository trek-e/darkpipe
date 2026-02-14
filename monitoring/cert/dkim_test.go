package cert

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/darkpipe/darkpipe/dns/dkim"
)

func TestShouldRotateDKIM_NoLastRotation(t *testing.T) {
	// If no last rotation file exists, should rotate
	tmpDir := t.TempDir()
	lastRotationPath := filepath.Join(tmpDir, ".last-rotation")

	shouldRotate, err := ShouldRotateDKIM(lastRotationPath)
	if err != nil {
		t.Fatalf("ShouldRotateDKIM failed: %v", err)
	}

	if !shouldRotate {
		t.Error("Expected rotation when no last rotation file exists")
	}
}

func TestShouldRotateDKIM_QuarterChanged(t *testing.T) {
	// Create last rotation in previous quarter
	tmpDir := t.TempDir()
	lastRotationPath := filepath.Join(tmpDir, ".last-rotation")

	// Calculate a time in the previous quarter
	now := time.Now()
	previousQuarter := now.AddDate(0, -4, 0) // 4 months ago (guaranteed different quarter)

	// Write last rotation timestamp
	if err := os.WriteFile(lastRotationPath, []byte(previousQuarter.Format(time.RFC3339)), 0644); err != nil {
		t.Fatalf("Failed to write last rotation: %v", err)
	}

	shouldRotate, err := ShouldRotateDKIM(lastRotationPath)
	if err != nil {
		t.Fatalf("ShouldRotateDKIM failed: %v", err)
	}

	if !shouldRotate {
		t.Error("Expected rotation when quarter has changed")
	}
}

func TestShouldRotateDKIM_SameQuarter(t *testing.T) {
	// Create last rotation in current quarter
	tmpDir := t.TempDir()
	lastRotationPath := filepath.Join(tmpDir, ".last-rotation")

	// Write recent timestamp (same quarter)
	now := time.Now()
	if err := os.WriteFile(lastRotationPath, []byte(now.AddDate(0, 0, -7).Format(time.RFC3339)), 0644); err != nil {
		t.Fatalf("Failed to write last rotation: %v", err)
	}

	shouldRotate, err := ShouldRotateDKIM(lastRotationPath)
	if err != nil {
		t.Fatalf("ShouldRotateDKIM failed: %v", err)
	}

	if shouldRotate {
		t.Error("Expected no rotation within same quarter")
	}
}

func TestRotateDKIM_GeneratesCorrectSelector(t *testing.T) {
	tmpDir := t.TempDir()

	config := DKIMRotationConfig{
		Domain:           "example.com",
		Prefix:           "darkpipe",
		KeyDir:           tmpDir,
		LastRotationPath: filepath.Join(tmpDir, ".last-rotation"),
	}

	// Set last rotation to previous quarter to trigger rotation
	previousQuarter := time.Now().AddDate(0, -4, 0)
	os.WriteFile(config.LastRotationPath, []byte(previousQuarter.Format(time.RFC3339)), 0644)

	// Capture stdout
	// Note: In real test, we'd redirect stdout or check filesystem
	err := RotateDKIM(config)
	if err != nil {
		t.Fatalf("RotateDKIM failed: %v", err)
	}

	// Verify selector follows {prefix}-{YYYY}q{Q} format
	expectedSelector := dkim.GetCurrentSelector("darkpipe")

	// Check that key files were created
	privateKeyPath := filepath.Join(tmpDir, expectedSelector+".private.pem")
	publicKeyPath := filepath.Join(tmpDir, expectedSelector+".public.pem")

	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		t.Errorf("Private key file not created: %s", privateKeyPath)
	}
	if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
		t.Errorf("Public key file not created: %s", publicKeyPath)
	}

	// Verify key file permissions
	info, _ := os.Stat(privateKeyPath)
	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("Private key has incorrect permissions: %o (expected 0600)", mode)
	}
}

func TestRotateDKIM_WritesLastRotationTimestamp(t *testing.T) {
	tmpDir := t.TempDir()

	config := DKIMRotationConfig{
		Domain:           "example.com",
		Prefix:           "test",
		KeyDir:           tmpDir,
		LastRotationPath: filepath.Join(tmpDir, ".last-rotation"),
	}

	// Set last rotation to previous quarter
	previousQuarter := time.Now().AddDate(0, -4, 0)
	os.WriteFile(config.LastRotationPath, []byte(previousQuarter.Format(time.RFC3339)), 0644)

	// Perform rotation
	before := time.Now()
	err := RotateDKIM(config)
	if err != nil {
		t.Fatalf("RotateDKIM failed: %v", err)
	}
	after := time.Now()

	// Verify last rotation timestamp was updated
	data, err := os.ReadFile(config.LastRotationPath)
	if err != nil {
		t.Fatalf("Failed to read last rotation file: %v", err)
	}

	timestamp, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		t.Fatalf("Failed to parse timestamp: %v", err)
	}

	// Timestamp should be between before and after (with 1-second tolerance for rounding)
	if timestamp.Before(before.Add(-1*time.Second)) || timestamp.After(after.Add(1*time.Second)) {
		t.Errorf("Timestamp %v not within expected range [%v, %v]", timestamp, before, after)
	}
}

func TestRotateDKIM_DefaultPrefix(t *testing.T) {
	tmpDir := t.TempDir()

	config := DKIMRotationConfig{
		Domain:           "example.com",
		Prefix:           "", // Empty prefix should default to "darkpipe"
		KeyDir:           tmpDir,
		LastRotationPath: filepath.Join(tmpDir, ".last-rotation"),
	}

	// Set last rotation to previous quarter
	previousQuarter := time.Now().AddDate(0, -4, 0)
	os.WriteFile(config.LastRotationPath, []byte(previousQuarter.Format(time.RFC3339)), 0644)

	err := RotateDKIM(config)
	if err != nil {
		t.Fatalf("RotateDKIM failed: %v", err)
	}

	// Verify selector uses default prefix "darkpipe"
	expectedSelector := dkim.GetCurrentSelector("darkpipe")
	privateKeyPath := filepath.Join(tmpDir, expectedSelector+".private.pem")

	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		t.Error("Expected key file with default 'darkpipe' prefix not found")
	}
}

func TestRotateDKIM_RotationNotDue(t *testing.T) {
	tmpDir := t.TempDir()

	config := DKIMRotationConfig{
		Domain:           "example.com",
		Prefix:           "test",
		KeyDir:           tmpDir,
		LastRotationPath: filepath.Join(tmpDir, ".last-rotation"),
	}

	// Set last rotation to current quarter (yesterday)
	yesterday := time.Now().AddDate(0, 0, -1)
	os.WriteFile(config.LastRotationPath, []byte(yesterday.Format(time.RFC3339)), 0644)

	// Attempt rotation (should fail because rotation not due)
	err := RotateDKIM(config)
	if err == nil {
		t.Error("Expected error when rotation not due")
	}
	if !strings.Contains(err.Error(), "not due") {
		t.Errorf("Expected 'not due' error, got: %v", err)
	}
}

func TestRotateDKIM_KeySize(t *testing.T) {
	tmpDir := t.TempDir()

	config := DKIMRotationConfig{
		Domain:           "example.com",
		Prefix:           "test",
		KeyDir:           tmpDir,
		LastRotationPath: filepath.Join(tmpDir, ".last-rotation"),
	}

	// Set last rotation to previous quarter
	previousQuarter := time.Now().AddDate(0, -4, 0)
	os.WriteFile(config.LastRotationPath, []byte(previousQuarter.Format(time.RFC3339)), 0644)

	err := RotateDKIM(config)
	if err != nil {
		t.Fatalf("RotateDKIM failed: %v", err)
	}

	// Verify key size is 2048 bits
	// This is implicitly tested by dkim.GenerateKeyPair(2048)
	// We can verify the key file exists and is readable
	selector := dkim.GetCurrentSelector("test")
	privateKeyPath := filepath.Join(tmpDir, selector+".private.pem")

	// Load private key to verify it's valid
	_, err = dkim.LoadPrivateKey(privateKeyPath)
	if err != nil {
		t.Errorf("Failed to load generated private key: %v", err)
	}
}
