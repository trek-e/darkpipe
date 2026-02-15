package mailmigrate

import (
	"os"
	"testing"
)

func TestParseVCFFile_MultiContact(t *testing.T) {
	// Create temp VCF file with multiple contacts
	content := `BEGIN:VCARD
VERSION:3.0
FN:John Doe
EMAIL:john@example.com
TEL:555-1234
END:VCARD
BEGIN:VCARD
VERSION:3.0
FN:Jane Doe
EMAIL:jane@example.com
ORG:Acme Corp
END:VCARD`

	tmpfile, err := os.CreateTemp("", "test-*.vcf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cards, err := parseVCFFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("parseVCFFile failed: %v", err)
	}

	if len(cards) != 2 {
		t.Errorf("expected 2 cards, got %d", len(cards))
	}

	// Verify first card
	if extractPrimaryEmail(cards[0]) != "john@example.com" {
		t.Error("first card email mismatch")
	}

	// Verify second card
	if extractPrimaryEmail(cards[1]) != "jane@example.com" {
		t.Error("second card email mismatch")
	}
}

func TestParseICSFile_MultiEvent(t *testing.T) {
	// Create temp ICS file with multiple events
	content := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
UID:event-1
SUMMARY:Event 1
DTSTART:20260214T100000Z
DTEND:20260214T110000Z
END:VEVENT
BEGIN:VEVENT
UID:event-2
SUMMARY:Event 2
DTSTART:20260215T100000Z
DTEND:20260215T110000Z
END:VEVENT
END:VCALENDAR`

	tmpfile, err := os.CreateTemp("", "test-*.ics")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	calendars, err := parseICSFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("parseICSFile failed: %v", err)
	}

	if len(calendars) != 1 {
		t.Errorf("expected 1 calendar, got %d", len(calendars))
	}

	// Count events
	eventCount := 0
	for _, comp := range calendars[0].Children {
		if comp.Name == "VEVENT" {
			eventCount++
		}
	}

	if eventCount != 2 {
		t.Errorf("expected 2 events, got %d", eventCount)
	}
}

func TestDryRunVCF(t *testing.T) {
	content := `BEGIN:VCARD
VERSION:3.0
FN:With Email
EMAIL:test@example.com
END:VCARD
BEGIN:VCARD
VERSION:3.0
FN:Without Email
TEL:555-1234
END:VCARD`

	tmpfile, err := os.CreateTemp("", "test-*.vcf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	result, err := DryRunVCF(tmpfile.Name())
	if err != nil {
		t.Fatalf("DryRunVCF failed: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("expected total=2, got %d", result.Total)
	}

	if result.WithEmail != 1 {
		t.Errorf("expected withEmail=1, got %d", result.WithEmail)
	}

	if result.WithoutEmail != 1 {
		t.Errorf("expected withoutEmail=1, got %d", result.WithoutEmail)
	}
}

func TestDryRunICS(t *testing.T) {
	content := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
UID:event-1
SUMMARY:Event 1
END:VEVENT
BEGIN:VEVENT
UID:event-2
SUMMARY:Event 2
END:VEVENT
BEGIN:VEVENT
UID:event-3
SUMMARY:Event 3
END:VEVENT
END:VCALENDAR`

	tmpfile, err := os.CreateTemp("", "test-*.ics")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	result, err := DryRunICS(tmpfile.Name())
	if err != nil {
		t.Fatalf("DryRunICS failed: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("expected total=3, got %d", result.Total)
	}
}

func TestParseVCFFile_Empty(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test-*.vcf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	cards, err := parseVCFFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("parseVCFFile failed on empty file: %v", err)
	}

	if len(cards) != 0 {
		t.Errorf("expected 0 cards from empty file, got %d", len(cards))
	}
}

func TestParseVCFFile_SingleContact(t *testing.T) {
	content := `BEGIN:VCARD
VERSION:3.0
FN:Single Contact
EMAIL:single@example.com
END:VCARD`

	tmpfile, err := os.CreateTemp("", "test-*.vcf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cards, err := parseVCFFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("parseVCFFile failed: %v", err)
	}

	if len(cards) != 1 {
		t.Errorf("expected 1 card, got %d", len(cards))
	}
}
