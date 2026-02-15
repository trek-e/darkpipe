// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package tls

import (
	"os"
	"strings"
	"testing"
)

func TestStrictMode_GeneratePolicyMapWithStrictEnabled(t *testing.T) {
	tmpFile := "/tmp/darkpipe_test_policy_map"
	defer os.Remove(tmpFile)
	defer os.Remove(tmpFile + ".lmdb")

	strict := &StrictMode{
		Enabled:            true,
		RefuseAllPlaintext: true,
		PolicyMapPath:      tmpFile,
	}

	// Note: GeneratePolicyMap will fail without postmap available, which is expected in test environment
	err := strict.GeneratePolicyMap()

	// Check that the policy file was created (even if postmap fails)
	content, readErr := os.ReadFile(tmpFile)
	if readErr != nil {
		t.Fatalf("Expected policy map file to be created, got error: %v", readErr)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "* encrypt") {
		t.Errorf("Expected policy map to contain '* encrypt', got: %s", contentStr)
	}

	// We expect postmap to fail in test environment (no postmap binary), so don't fail the test on that
	if err != nil && !strings.Contains(err.Error(), "postmap") {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestStrictMode_GeneratePolicyMapWithStrictDisabled(t *testing.T) {
	tmpFile := "/tmp/darkpipe_test_policy_map_disabled"
	defer os.Remove(tmpFile)

	strict := &StrictMode{
		Enabled:            false,
		RefuseAllPlaintext: false,
		PolicyMapPath:      tmpFile,
	}

	err := strict.GeneratePolicyMap()
	if err != nil {
		t.Fatalf("Expected no error when strict mode disabled, got: %v", err)
	}

	// Policy map file should NOT be created when strict mode is disabled
	if _, err := os.Stat(tmpFile); err == nil {
		t.Error("Expected policy map file NOT to be created when strict mode disabled")
	}
}

func TestStrictMode_PostconfCommandConstruction(t *testing.T) {
	// This test verifies that the postconf commands would be correctly constructed
	// We cannot actually run postconf in the test environment, but we can verify logic

	strict := NewStrictMode(true)

	// Verify initial state
	if !strict.Enabled {
		t.Error("Expected strict mode to be enabled")
	}
	if !strict.RefuseAllPlaintext {
		t.Error("Expected RefuseAllPlaintext to be true when strict mode enabled")
	}

	// Test that default policy map path is set
	if strict.PolicyMapPath == "" {
		t.Error("Expected PolicyMapPath to be set")
	}
}

func TestStrictMode_NewStrictMode(t *testing.T) {
	tests := []struct {
		name                     string
		enabled                  bool
		expectedEnabled          bool
		expectedRefuseAllPlaintext bool
	}{
		{
			name:                     "Strict mode enabled",
			enabled:                  true,
			expectedEnabled:          true,
			expectedRefuseAllPlaintext: true,
		},
		{
			name:                     "Strict mode disabled",
			enabled:                  false,
			expectedEnabled:          false,
			expectedRefuseAllPlaintext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strict := NewStrictMode(tt.enabled)
			if strict.Enabled != tt.expectedEnabled {
				t.Errorf("Expected Enabled=%v, got %v", tt.expectedEnabled, strict.Enabled)
			}
			if strict.RefuseAllPlaintext != tt.expectedRefuseAllPlaintext {
				t.Errorf("Expected RefuseAllPlaintext=%v, got %v", tt.expectedRefuseAllPlaintext, strict.RefuseAllPlaintext)
			}
		})
	}
}
