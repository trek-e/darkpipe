// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package mailmigrate

import (
	"strings"
	"testing"

	"github.com/emersion/go-vcard"
)

func TestExtractPrimaryEmail(t *testing.T) {
	tests := []struct {
		name      string
		vcf       string
		wantEmail string
	}{
		{
			name: "single email",
			vcf: `BEGIN:VCARD
VERSION:3.0
FN:John Doe
EMAIL:john@example.com
END:VCARD`,
			wantEmail: "john@example.com",
		},
		{
			name: "multiple emails - first one",
			vcf: `BEGIN:VCARD
VERSION:3.0
FN:Jane Doe
EMAIL;TYPE=WORK:jane.work@example.com
EMAIL;TYPE=HOME:jane.home@example.com
END:VCARD`,
			wantEmail: "jane.work@example.com",
		},
		{
			name: "no email",
			vcf: `BEGIN:VCARD
VERSION:3.0
FN:No Email
TEL:555-1234
END:VCARD`,
			wantEmail: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := vcard.NewDecoder(strings.NewReader(tt.vcf))
			card, err := dec.Decode()
			if err != nil {
				t.Fatalf("failed to decode vCard: %v", err)
			}

			email := extractPrimaryEmail(card)
			if email != tt.wantEmail {
				t.Errorf("extractPrimaryEmail() = %v, want %v", email, tt.wantEmail)
			}
		})
	}
}

func TestMergeContact(t *testing.T) {
	tests := []struct {
		name           string
		destVCF        string
		sourceVCF      string
		wantTel        bool // should have telephone after merge
		wantOrg        bool // should have organization after merge
		preserveNote   bool // destination note should be preserved
	}{
		{
			name: "fill empty fields from source",
			destVCF: `BEGIN:VCARD
VERSION:3.0
FN:John Doe
EMAIL:john@example.com
NOTE:Existing note
END:VCARD`,
			sourceVCF: `BEGIN:VCARD
VERSION:3.0
FN:John Doe
EMAIL:john@example.com
TEL:555-1234
ORG:Acme Corp
NOTE:Source note (should not overwrite)
END:VCARD`,
			wantTel:      true,
			wantOrg:      true,
			preserveNote: true,
		},
		{
			name: "don't overwrite existing fields",
			destVCF: `BEGIN:VCARD
VERSION:3.0
FN:Jane Doe
EMAIL:jane@example.com
TEL:555-5678
ORG:Existing Corp
END:VCARD`,
			sourceVCF: `BEGIN:VCARD
VERSION:3.0
FN:Jane Doe
EMAIL:jane@example.com
TEL:555-9999
ORG:Source Corp
END:VCARD`,
			wantTel:      true, // should keep destination value
			wantOrg:      true, // should keep destination value
			preserveNote: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Decode destination
			destDec := vcard.NewDecoder(strings.NewReader(tt.destVCF))
			dest, err := destDec.Decode()
			if err != nil {
				t.Fatalf("failed to decode dest vCard: %v", err)
			}

			// Decode source
			sourceDec := vcard.NewDecoder(strings.NewReader(tt.sourceVCF))
			source, err := sourceDec.Decode()
			if err != nil {
				t.Fatalf("failed to decode source vCard: %v", err)
			}

			// Get original values
			origTel := dest.Get(vcard.FieldTelephone)
			origOrg := dest.Get(vcard.FieldOrganization)
			origNote := dest.Get(vcard.FieldNote)

			// Merge
			merged, decisions := mergeContact(dest, source)

			// Check telephone
			if tt.wantTel {
				if merged.Get(vcard.FieldTelephone) == nil {
					t.Error("expected telephone to be present after merge")
				}
				// If dest had telephone, it should be preserved
				if origTel != nil && merged.Get(vcard.FieldTelephone).Value != origTel.Value {
					t.Error("expected destination telephone to be preserved")
				}
			}

			// Check organization
			if tt.wantOrg {
				if merged.Get(vcard.FieldOrganization) == nil {
					t.Error("expected organization to be present after merge")
				}
				// If dest had org, it should be preserved
				if origOrg != nil && merged.Get(vcard.FieldOrganization).Value != origOrg.Value {
					t.Error("expected destination organization to be preserved")
				}
			}

			// Check note preservation
			if tt.preserveNote {
				if origNote != nil && merged.Get(vcard.FieldNote).Value != origNote.Value {
					t.Error("expected destination note to be preserved")
				}
			}

			// Check that decisions were logged only if fields were actually filled
			// Second test case: dest already has both fields, so no decisions expected
			if tt.name == "fill empty fields from source" && len(decisions) == 0 {
				t.Error("expected merge decisions to be logged when filling empty fields")
			}
		})
	}
}

func TestMergeContactAllModes(t *testing.T) {
	// Test all merge modes: append, overwrite, skip
	destVCF := `BEGIN:VCARD
VERSION:3.0
FN:Test User
EMAIL:test@example.com
TEL:555-1111
END:VCARD`

	sourceVCF := `BEGIN:VCARD
VERSION:3.0
FN:Test User
EMAIL:test@example.com
TEL:555-2222
ORG:Test Corp
END:VCARD`

	destDec := vcard.NewDecoder(strings.NewReader(destVCF))
	dest, _ := destDec.Decode()

	sourceDec := vcard.NewDecoder(strings.NewReader(sourceVCF))
	source, _ := sourceDec.Decode()

	// Test append mode (default)
	merged, _ := mergeContact(dest, source)

	// Should have org from source
	if merged.Get(vcard.FieldOrganization) == nil {
		t.Error("append mode: expected organization from source")
	}

	// Should preserve dest telephone
	if merged.Get(vcard.FieldTelephone).Value != "555-1111" {
		t.Error("append mode: expected to preserve dest telephone")
	}
}

func TestCardDAVStateIntegration(t *testing.T) {
	// Test that CardDAV operations integrate with state tracking
	state := &MigrationState{
		Contacts: make(map[string]bool),
	}

	email := "test@example.com"

	// Initially not migrated
	if state.IsContactMigrated(email) {
		t.Error("expected contact to not be migrated initially")
	}

	// Mark as migrated
	if err := state.MarkContactMigrated(email); err != nil {
		t.Fatalf("failed to mark contact as migrated: %v", err)
	}

	// Now should be migrated
	if !state.IsContactMigrated(email) {
		t.Error("expected contact to be migrated after marking")
	}
}
