// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package mailmigrate

import (
	"testing"
)

func TestNewFolderMapper_Gmail(t *testing.T) {
	mapper := NewFolderMapper("gmail", nil)

	tests := []struct {
		source   string
		wantDest string
		wantSkip bool
	}{
		{"[Gmail]/Sent Mail", "Sent", false},
		{"[Gmail]/Drafts", "Drafts", false},
		{"[Gmail]/Trash", "Trash", false},
		{"[Gmail]/Spam", "Junk", false},
		{"[Gmail]/All Mail", "", true},
		{"[Gmail]/Important", "", true},
		{"[Gmail]/Starred", "", true},
		{"INBOX", "INBOX", false}, // Pass-through
		{"Custom Folder", "Custom Folder", false},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			dest, skip := mapper.Map(tt.source)
			if dest != tt.wantDest {
				t.Errorf("Map(%q) dest = %q, want %q", tt.source, dest, tt.wantDest)
			}
			if skip != tt.wantSkip {
				t.Errorf("Map(%q) skip = %v, want %v", tt.source, skip, tt.wantSkip)
			}
		})
	}
}

func TestNewFolderMapper_Outlook(t *testing.T) {
	mapper := NewFolderMapper("outlook", nil)

	tests := []struct {
		source   string
		wantDest string
		wantSkip bool
	}{
		{"Deleted Items", "Trash", false},
		{"Sent Items", "Sent", false},
		{"Junk Email", "Junk", false},
		{"Clutter", "", true},
		{"INBOX", "INBOX", false},
		{"Custom Folder", "Custom Folder", false},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			dest, skip := mapper.Map(tt.source)
			if dest != tt.wantDest {
				t.Errorf("Map(%q) dest = %q, want %q", tt.source, dest, tt.wantDest)
			}
			if skip != tt.wantSkip {
				t.Errorf("Map(%q) skip = %v, want %v", tt.source, skip, tt.wantSkip)
			}
		})
	}
}

func TestNewFolderMapper_Generic(t *testing.T) {
	mapper := NewFolderMapper("generic", nil)

	tests := []struct {
		source   string
		wantDest string
		wantSkip bool
	}{
		{"INBOX", "INBOX", false},
		{"Sent", "Sent", false},
		{"Drafts", "Drafts", false},
		{"Trash", "Trash", false},
		{"Custom Folder", "Custom Folder", false},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			dest, skip := mapper.Map(tt.source)
			if dest != tt.wantDest {
				t.Errorf("Map(%q) dest = %q, want %q", tt.source, dest, tt.wantDest)
			}
			if skip != tt.wantSkip {
				t.Errorf("Map(%q) skip = %v, want %v", tt.source, skip, tt.wantSkip)
			}
		})
	}
}

func TestNewFolderMapper_WithOverrides(t *testing.T) {
	overrides := map[string]string{
		"[Gmail]/Sent Mail": "MySent",     // Override default mapping
		"INBOX":             "",            // Skip INBOX
		"Custom":            "NewCustom",   // Add new mapping
	}

	mapper := NewFolderMapper("gmail", overrides)

	tests := []struct {
		source   string
		wantDest string
		wantSkip bool
	}{
		{"[Gmail]/Sent Mail", "MySent", false},       // Overridden
		{"INBOX", "", true},                          // Skipped via override
		{"Custom", "NewCustom", false},               // New mapping
		{"[Gmail]/Drafts", "Drafts", false},          // Default preserved
		{"[Gmail]/All Mail", "", true},               // Default skip preserved
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			dest, skip := mapper.Map(tt.source)
			if dest != tt.wantDest {
				t.Errorf("Map(%q) dest = %q, want %q", tt.source, dest, tt.wantDest)
			}
			if skip != tt.wantSkip {
				t.Errorf("Map(%q) skip = %v, want %v", tt.source, skip, tt.wantSkip)
			}
		})
	}
}

func TestAllMappings(t *testing.T) {
	mapper := NewFolderMapper("gmail", nil)

	all := mapper.AllMappings()

	// Verify mapped folders
	if all["[Gmail]/Sent Mail"] != "Sent" {
		t.Errorf("AllMappings [Gmail]/Sent Mail = %q, want Sent", all["[Gmail]/Sent Mail"])
	}

	// Verify skip markers
	if all["[Gmail]/All Mail"] != "(skip)" {
		t.Errorf("AllMappings [Gmail]/All Mail = %q, want (skip)", all["[Gmail]/All Mail"])
	}

	// Count entries
	expectedCount := 7 // 4 mappings + 3 skips
	if len(all) != expectedCount {
		t.Errorf("AllMappings count = %d, want %d", len(all), expectedCount)
	}
}

func TestLabelToKeyword(t *testing.T) {
	mapper := &FolderMapper{}

	tests := []struct {
		label   string
		want    string
	}{
		{"Work", "Work"},
		{"Personal Email", "Personal_Email"},
		{"Travel/Receipts", "TravelReceipts"},
		{"My Label (2026)", "My_Label_2026"},
		{"123-Numbers", "label_123-Numbers"},  // Starts with digit
		{"", "custom_label"},                  // Empty
		{"!@#$%", "custom_label"},             // All special chars
		{"label_with_underscores", "label_with_underscores"},
		{"label-with-hyphens", "label-with-hyphens"},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			got := mapper.LabelToKeyword(tt.label)
			if got != tt.want {
				t.Errorf("LabelToKeyword(%q) = %q, want %q", tt.label, got, tt.want)
			}
		})
	}
}

func TestLabelToKeyword_NoSpaces(t *testing.T) {
	mapper := &FolderMapper{}

	result := mapper.LabelToKeyword("My Label With Spaces")

	// Verify no spaces in result
	for _, ch := range result {
		if ch == ' ' {
			t.Error("LabelToKeyword result contains spaces, should be sanitized")
		}
	}

	if result != "My_Label_With_Spaces" {
		t.Errorf("LabelToKeyword result = %q, want My_Label_With_Spaces", result)
	}
}

func TestLabelToKeyword_ValidAtom(t *testing.T) {
	mapper := &FolderMapper{}

	result := mapper.LabelToKeyword("Valid-Label_123")

	// Result should only contain alphanumeric, underscore, hyphen
	for _, ch := range result {
		valid := (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '_' || ch == '-'

		if !valid {
			t.Errorf("LabelToKeyword result contains invalid char %q, should be alphanumeric/_/-", ch)
		}
	}
}

func TestLabelsAsFolders_Flag(t *testing.T) {
	mapper := NewFolderMapper("gmail", nil)

	// Default should be false
	if mapper.LabelsAsFolders {
		t.Error("LabelsAsFolders should default to false")
	}

	// Can be set
	mapper.LabelsAsFolders = true
	if !mapper.LabelsAsFolders {
		t.Error("LabelsAsFolders should be settable to true")
	}
}
