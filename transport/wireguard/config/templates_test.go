// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateCloudConfig(t *testing.T) {
	cfg := CloudConfig{
		PrivateKey: "cF7z8J9L2X5N4M1K6H3G8D9S0A2Q5R7T8W1Y4U3V6B=",
		HomePubKey: "xA1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U=",
	}

	content, err := GenerateCloudConfig(cfg)
	if err != nil {
		t.Fatalf("GenerateCloudConfig failed: %v", err)
	}

	// Verify content is not empty
	if content == "" {
		t.Error("generated config is empty")
	}

	// Verify format contains required sections
	if !strings.Contains(content, "[Interface]") {
		t.Error("config missing [Interface] section")
	}
	if !strings.Contains(content, "[Peer]") {
		t.Error("config missing [Peer] section")
	}

	// Verify required fields are present
	if !strings.Contains(content, "PrivateKey = "+cfg.PrivateKey) {
		t.Error("config missing PrivateKey")
	}
	if !strings.Contains(content, "PublicKey = "+cfg.HomePubKey) {
		t.Error("config missing home PublicKey")
	}
	if !strings.Contains(content, "ListenPort = 51820") {
		t.Error("config missing default ListenPort")
	}
	if !strings.Contains(content, "Address = 10.8.0.1/24") {
		t.Error("config missing default Address")
	}
	if !strings.Contains(content, "AllowedIPs = 10.8.0.2/32") {
		t.Error("config missing default AllowedIPs")
	}

	// Verify cloud config does NOT have PersistentKeepalive
	if strings.Contains(content, "PersistentKeepalive") {
		t.Error("cloud config should not contain PersistentKeepalive")
	}

	t.Logf("Generated cloud config:\n%s", content)
}

func TestGenerateHomeConfig(t *testing.T) {
	cfg := HomeConfig{
		PrivateKey:    "pR1v2A3t4E5K6E7y8H9O0m1E2D3e4V5i6C7e8K9E0Y=",
		CloudPubKey:   "cL0u1D2P3u4B5l6I7c8K9E0y1A2B3C4D5E6F7G8H9I=",
		CloudEndpoint: "203.0.113.1:51820",
	}

	content, err := GenerateHomeConfig(cfg)
	if err != nil {
		t.Fatalf("GenerateHomeConfig failed: %v", err)
	}

	// Verify content is not empty
	if content == "" {
		t.Error("generated config is empty")
	}

	// Verify format contains required sections
	if !strings.Contains(content, "[Interface]") {
		t.Error("config missing [Interface] section")
	}
	if !strings.Contains(content, "[Peer]") {
		t.Error("config missing [Peer] section")
	}

	// Verify required fields are present
	if !strings.Contains(content, "PrivateKey = "+cfg.PrivateKey) {
		t.Error("config missing PrivateKey")
	}
	if !strings.Contains(content, "PublicKey = "+cfg.CloudPubKey) {
		t.Error("config missing cloud PublicKey")
	}
	if !strings.Contains(content, "Endpoint = "+cfg.CloudEndpoint) {
		t.Error("config missing Endpoint")
	}
	if !strings.Contains(content, "Address = 10.8.0.2/24") {
		t.Error("config missing default Address")
	}
	if !strings.Contains(content, "AllowedIPs = 10.8.0.1/32") {
		t.Error("config missing default AllowedIPs")
	}

	// CRITICAL: Verify PersistentKeepalive = 25 is present (NAT traversal)
	if !strings.Contains(content, "PersistentKeepalive = 25") {
		t.Error("home config MUST contain PersistentKeepalive = 25 for NAT traversal")
	}

	t.Logf("Generated home config:\n%s", content)
}

func TestGenerateCloudConfigWithCustomValues(t *testing.T) {
	cfg := CloudConfig{
		PrivateKey:     "cF7z8J9L2X5N4M1K6H3G8D9S0A2Q5R7T8W1Y4U3V6B=",
		Address:        "192.168.1.1/24",
		ListenPort:     12345,
		HomePubKey:     "xA1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U=",
		HomeAllowedIPs: "192.168.1.2/32",
	}

	content, err := GenerateCloudConfig(cfg)
	if err != nil {
		t.Fatalf("GenerateCloudConfig failed: %v", err)
	}

	// Verify custom values are used
	if !strings.Contains(content, "Address = 192.168.1.1/24") {
		t.Error("config missing custom Address")
	}
	if !strings.Contains(content, "ListenPort = 12345") {
		t.Error("config missing custom ListenPort")
	}
	if !strings.Contains(content, "AllowedIPs = 192.168.1.2/32") {
		t.Error("config missing custom AllowedIPs")
	}
}

func TestGenerateCloudConfigMissingRequiredFields(t *testing.T) {
	// Test missing PrivateKey
	cfg := CloudConfig{
		HomePubKey: "xA1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U=",
	}
	_, err := GenerateCloudConfig(cfg)
	if err == nil {
		t.Error("expected error for missing PrivateKey, got nil")
	}

	// Test missing HomePubKey
	cfg = CloudConfig{
		PrivateKey: "cF7z8J9L2X5N4M1K6H3G8D9S0A2Q5R7T8W1Y4U3V6B=",
	}
	_, err = GenerateCloudConfig(cfg)
	if err == nil {
		t.Error("expected error for missing HomePubKey, got nil")
	}
}

func TestGenerateHomeConfigMissingRequiredFields(t *testing.T) {
	// Test missing PrivateKey
	cfg := HomeConfig{
		CloudPubKey:   "cL0u1D2P3u4B5l6I7c8K9E0y1A2B3C4D5E6F7G8H9I=",
		CloudEndpoint: "203.0.113.1:51820",
	}
	_, err := GenerateHomeConfig(cfg)
	if err == nil {
		t.Error("expected error for missing PrivateKey, got nil")
	}

	// Test missing CloudPubKey
	cfg = HomeConfig{
		PrivateKey:    "pR1v2A3t4E5K6E7y8H9O0m1E2D3e4V5i6C7e8K9E0Y=",
		CloudEndpoint: "203.0.113.1:51820",
	}
	_, err = GenerateHomeConfig(cfg)
	if err == nil {
		t.Error("expected error for missing CloudPubKey, got nil")
	}

	// Test missing CloudEndpoint
	cfg = HomeConfig{
		PrivateKey:  "pR1v2A3t4E5K6E7y8H9O0m1E2D3e4V5i6C7e8K9E0Y=",
		CloudPubKey: "cL0u1D2P3u4B5l6I7c8K9E0y1A2B3C4D5E6F7G8H9I=",
	}
	_, err = GenerateHomeConfig(cfg)
	if err == nil {
		t.Error("expected error for missing CloudEndpoint, got nil")
	}
}

func TestWriteConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-wg0.conf")

	content := "[Interface]\nPrivateKey = testkey123\n"

	err := WriteConfig(content, testFile)
	if err != nil {
		t.Fatalf("WriteConfig failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Verify file permissions are 0600
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("failed to stat config file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("config file has incorrect permissions: got %o, want 0600", info.Mode().Perm())
	}

	// Verify file content
	readContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	if string(readContent) != content {
		t.Errorf("config file content mismatch: got %q, want %q", string(readContent), content)
	}
}

func TestWriteConfigEmptyPath(t *testing.T) {
	content := "[Interface]\nPrivateKey = testkey123\n"
	err := WriteConfig(content, "")
	if err == nil {
		t.Error("expected error for empty path, got nil")
	}
}
