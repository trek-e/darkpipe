// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package mailmigrate

import (
	"strings"
	"testing"

	"github.com/emersion/go-ical"
)

func TestExtractEventUID(t *testing.T) {
	tests := []struct {
		name    string
		ical    string
		wantUID string
		wantErr bool
	}{
		{
			name: "valid event with UID",
			ical: `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
UID:test-event-123
SUMMARY:Test Event
DTSTART:20260214T100000Z
DTEND:20260214T110000Z
END:VEVENT
END:VCALENDAR`,
			wantUID: "test-event-123",
			wantErr: false,
		},
		{
			name: "event without UID",
			ical: `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
SUMMARY:Test Event
END:VEVENT
END:VCALENDAR`,
			wantUID: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := ical.NewDecoder(strings.NewReader(tt.ical))
			cal, err := dec.Decode()
			if err != nil {
				t.Fatalf("failed to decode iCal: %v", err)
			}

			uid, err := extractEventUID(cal)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractEventUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if uid != tt.wantUID {
				t.Errorf("extractEventUID() = %v, want %v", uid, tt.wantUID)
			}
		})
	}
}

func TestCalDAVStateIntegration(t *testing.T) {
	// Test that CalDAV operations integrate with state tracking
	state := &MigrationState{
		CalEvents: make(map[string]bool),
	}

	uid := "test-event-123"

	// Initially not migrated
	if state.IsCalEventMigrated(uid) {
		t.Error("expected event to not be migrated initially")
	}

	// Mark as migrated
	if err := state.MarkCalEventMigrated(uid); err != nil {
		t.Fatalf("failed to mark event as migrated: %v", err)
	}

	// Now should be migrated
	if !state.IsCalEventMigrated(uid) {
		t.Error("expected event to be migrated after marking")
	}
}
