// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package tls provides TLS monitoring and strict mode configuration for Postfix.
package tls

import (
	"fmt"
	"os"
	"os/exec"
)

// StrictMode manages Postfix TLS policy enforcement.
type StrictMode struct {
	Enabled            bool   // Whether strict mode is enabled
	RefuseAllPlaintext bool   // Refuse ALL plaintext connections (not per-domain)
	PolicyMapPath      string // Path to Postfix TLS policy map
}

// NewStrictMode creates a new strict mode manager with default settings.
func NewStrictMode(enabled bool) *StrictMode {
	return &StrictMode{
		Enabled:            enabled,
		RefuseAllPlaintext: enabled, // If strict mode is enabled, refuse all plaintext by default
		PolicyMapPath:      "/etc/postfix/tls_policy",
	}
}

// GeneratePolicyMap creates a Postfix TLS policy map file if strict mode requires it.
// When RefuseAllPlaintext is true, generates a catch-all "* encrypt" rule.
func (s *StrictMode) GeneratePolicyMap() error {
	if !s.RefuseAllPlaintext {
		// No policy map needed if not refusing all plaintext
		return nil
	}

	// Create policy map with catch-all encrypt rule
	content := "# DarkPipe TLS strict mode policy map\n"
	content += "# Generated automatically - do not edit manually\n"
	content += "* encrypt\n"

	if err := os.WriteFile(s.PolicyMapPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write policy map: %w", err)
	}

	// Hash the policy map using postmap (LMDB format)
	cmd := exec.Command("postmap", fmt.Sprintf("lmdb:%s", s.PolicyMapPath))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run postmap: %w (output: %s)", err, output)
	}

	return nil
}

// ApplyToPostfix applies strict mode settings to Postfix configuration.
// Uses postconf to dynamically update settings without editing main.cf.
func (s *StrictMode) ApplyToPostfix() error {
	if !s.Enabled {
		return s.DisableStrictMode()
	}

	// Set TLS security level to 'encrypt' for both inbound and outbound
	// This means TLS is REQUIRED for all connections
	commands := [][]string{
		{"postconf", "-e", "smtp_tls_security_level=encrypt"},   // Outbound (relay to remote MTAs)
		{"postconf", "-e", "smtpd_tls_security_level=encrypt"},  // Inbound (accept from internet)
	}

	if s.RefuseAllPlaintext {
		// Add policy map reference (though security_level=encrypt already enforces it globally)
		commands = append(commands, []string{"postconf", "-e", fmt.Sprintf("smtp_tls_policy_maps=lmdb:%s", s.PolicyMapPath)})
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to run %v: %w (output: %s)", cmdArgs, err, output)
		}
	}

	// Reload Postfix to apply changes
	cmd := exec.Command("postfix", "reload")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reload postfix: %w (output: %s)", err, output)
	}

	return nil
}

// DisableStrictMode reverts Postfix to opportunistic TLS ('may' security level).
func (s *StrictMode) DisableStrictMode() error {
	// Set TLS security level to 'may' (opportunistic) for both inbound and outbound
	commands := [][]string{
		{"postconf", "-e", "smtp_tls_security_level=may"},
		{"postconf", "-e", "smtpd_tls_security_level=may"},
		{"postconf", "-e", "smtp_tls_policy_maps="},  // Clear policy maps
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to run %v: %w (output: %s)", cmdArgs, err, output)
		}
	}

	// Reload Postfix to apply changes
	cmd := exec.Command("postfix", "reload")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reload postfix: %w (output: %s)", err, output)
	}

	return nil
}
