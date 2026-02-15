// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package config

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
)

// CloudConfig represents the WireGuard configuration for the cloud hub.
type CloudConfig struct {
	PrivateKey     string
	Address        string // Default: "10.8.0.1/24"
	ListenPort     int    // Default: 51820
	HomePubKey     string
	HomeAllowedIPs string // Default: "10.8.0.2/32"
}

// HomeConfig represents the WireGuard configuration for the home spoke.
type HomeConfig struct {
	PrivateKey          string
	Address             string // Default: "10.8.0.2/24"
	CloudPubKey         string
	CloudEndpoint       string
	CloudAllowedIPs     string // Default: "10.8.0.1/32"
	PersistentKeepalive int    // Default: 25
}

const cloudTemplate = `[Interface]
PrivateKey = {{.PrivateKey}}
Address = {{.Address}}
ListenPort = {{.ListenPort}}

[Peer]
PublicKey = {{.HomePubKey}}
AllowedIPs = {{.HomeAllowedIPs}}
`

const homeTemplate = `[Interface]
PrivateKey = {{.PrivateKey}}
Address = {{.Address}}

[Peer]
PublicKey = {{.CloudPubKey}}
Endpoint = {{.CloudEndpoint}}
AllowedIPs = {{.CloudAllowedIPs}}
PersistentKeepalive = {{.PersistentKeepalive}}
`

// GenerateCloudConfig generates a WireGuard configuration file for the cloud hub.
// Returns the configuration content as a string.
func GenerateCloudConfig(cfg CloudConfig) (string, error) {
	// Apply defaults
	if cfg.Address == "" {
		cfg.Address = "10.8.0.1/24"
	}
	if cfg.ListenPort == 0 {
		cfg.ListenPort = 51820
	}
	if cfg.HomeAllowedIPs == "" {
		cfg.HomeAllowedIPs = "10.8.0.2/32"
	}

	// Validate required fields
	if cfg.PrivateKey == "" {
		return "", fmt.Errorf("PrivateKey is required")
	}
	if cfg.HomePubKey == "" {
		return "", fmt.Errorf("HomePubKey is required")
	}

	// Parse and execute template
	tmpl, err := template.New("cloud").Parse(cloudTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse cloud template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return "", fmt.Errorf("failed to execute cloud template: %w", err)
	}

	return buf.String(), nil
}

// GenerateHomeConfig generates a WireGuard configuration file for the home spoke.
// Returns the configuration content as a string.
// Always includes PersistentKeepalive for NAT traversal.
func GenerateHomeConfig(cfg HomeConfig) (string, error) {
	// Apply defaults
	if cfg.Address == "" {
		cfg.Address = "10.8.0.2/24"
	}
	if cfg.CloudAllowedIPs == "" {
		cfg.CloudAllowedIPs = "10.8.0.1/32"
	}
	if cfg.PersistentKeepalive == 0 {
		cfg.PersistentKeepalive = 25
	}

	// Validate required fields
	if cfg.PrivateKey == "" {
		return "", fmt.Errorf("PrivateKey is required")
	}
	if cfg.CloudPubKey == "" {
		return "", fmt.Errorf("CloudPubKey is required")
	}
	if cfg.CloudEndpoint == "" {
		return "", fmt.Errorf("CloudEndpoint is required")
	}

	// Parse and execute template
	tmpl, err := template.New("home").Parse(homeTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse home template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return "", fmt.Errorf("failed to execute home template: %w", err)
	}

	return buf.String(), nil
}

// WriteConfig writes a WireGuard configuration to a file with secure permissions (0600).
// The file is created with mode 0600 to protect the private key.
func WriteConfig(content string, path string) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}

	// Write file with mode 0600 (read/write for owner only)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write config to %s: %w", path, err)
	}

	return nil
}
