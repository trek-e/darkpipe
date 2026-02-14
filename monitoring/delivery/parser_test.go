package delivery

import (
	"strings"
	"testing"
)

func TestParser_ParseLine_Sent(t *testing.T) {
	parser := NewParser()
	line := "Feb 14 10:23:45 mail postfix/smtp[1234]: ABCDEF1234: to=<user@example.com>, relay=gmail-smtp-in.l.google.com[142.250.1.26]:25, delay=1.2, delays=0.1/0/0.5/0.6, dsn=2.0.0, status=sent (250 2.0.0 OK)"

	entry, err := parser.ParseLine(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry == nil {
		t.Fatal("expected entry, got nil")
	}

	if entry.QueueID != "ABCDEF1234" {
		t.Errorf("expected queue_id ABCDEF1234, got %s", entry.QueueID)
	}

	if entry.To != "user@example.com" {
		t.Errorf("expected to user@example.com, got %s", entry.To)
	}

	if entry.Status != "delivered" {
		t.Errorf("expected status delivered, got %s", entry.Status)
	}

	if entry.Relay != "gmail-smtp-in.l.google.com[142.250.1.26]:25" {
		t.Errorf("expected relay gmail-smtp-in.l.google.com[142.250.1.26]:25, got %s", entry.Relay)
	}

	if entry.Delay != "1.2" {
		t.Errorf("expected delay 1.2, got %s", entry.Delay)
	}

	if entry.DSN != "2.0.0" {
		t.Errorf("expected dsn 2.0.0, got %s", entry.DSN)
	}

	if !strings.Contains(entry.StatusDetail, "250 2.0.0 OK") {
		t.Errorf("expected status_detail to contain '250 2.0.0 OK', got %s", entry.StatusDetail)
	}

	if entry.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestParser_ParseLine_Deferred(t *testing.T) {
	parser := NewParser()
	line := "Feb 14 10:23:46 mail postfix/smtp[1234]: ABCDEF1235: to=<user@failing.com>, relay=none, delay=300, dsn=4.4.1, status=deferred (connect to failing.com[1.2.3.4]:25: Connection timed out)"

	entry, err := parser.ParseLine(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry == nil {
		t.Fatal("expected entry, got nil")
	}

	if entry.QueueID != "ABCDEF1235" {
		t.Errorf("expected queue_id ABCDEF1235, got %s", entry.QueueID)
	}

	if entry.Status != "deferred" {
		t.Errorf("expected status deferred, got %s", entry.Status)
	}

	if entry.Relay != "none" {
		t.Errorf("expected relay none, got %s", entry.Relay)
	}

	if !strings.Contains(entry.StatusDetail, "Connection timed out") {
		t.Errorf("expected status_detail to contain 'Connection timed out', got %s", entry.StatusDetail)
	}
}

func TestParser_ParseLine_Bounced(t *testing.T) {
	parser := NewParser()
	line := "Feb 14 10:23:47 mail postfix/smtp[1234]: ABCDEF1236: to=<user@invalid.com>, relay=none, delay=0, dsn=5.1.1, status=bounced (user unknown)"

	entry, err := parser.ParseLine(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry == nil {
		t.Fatal("expected entry, got nil")
	}

	if entry.Status != "bounced" {
		t.Errorf("expected status bounced, got %s", entry.Status)
	}

	if entry.DSN != "5.1.1" {
		t.Errorf("expected dsn 5.1.1, got %s", entry.DSN)
	}

	if !strings.Contains(entry.StatusDetail, "user unknown") {
		t.Errorf("expected status_detail to contain 'user unknown', got %s", entry.StatusDetail)
	}
}

func TestParser_ParseLine_Expired(t *testing.T) {
	parser := NewParser()
	line := "Feb 14 10:23:48 mail postfix/qmgr[5678]: ABCDEF1237: to=<user@gone.com>, status=expired (message has been in queue too long)"

	entry, err := parser.ParseLine(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry == nil {
		t.Fatal("expected entry, got nil")
	}

	if entry.Status != "expired" {
		t.Errorf("expected status expired, got %s", entry.Status)
	}
}

func TestParser_ParseLine_NonDeliveryLine(t *testing.T) {
	parser := NewParser()

	testCases := []string{
		"Feb 14 10:23:45 mail postfix/smtpd[1234]: connect from unknown[1.2.3.4]",
		"Feb 14 10:23:45 mail postfix/cleanup[5678]: ABCDEF1234: message-id=<test@example.com>",
		"Feb 14 10:23:45 mail postfix/qmgr[9012]: ABCDEF1234: from=<sender@example.com>, size=1234, nrcpt=1",
	}

	for _, line := range testCases {
		entry, err := parser.ParseLine(line)
		if err != nil {
			t.Errorf("unexpected error for line '%s': %v", line, err)
		}
		if entry != nil {
			t.Errorf("expected nil for non-delivery line, got %+v", entry)
		}
	}
}

func TestParser_ParseLine_WithFrom(t *testing.T) {
	parser := NewParser()
	line := "Feb 14 10:23:45 mail postfix/smtp[1234]: ABCDEF1234: to=<user@example.com>, from=<sender@test.org>, relay=example.com[1.2.3.4]:25, status=sent (250 OK)"

	entry, err := parser.ParseLine(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry == nil {
		t.Fatal("expected entry, got nil")
	}

	if entry.From != "sender@test.org" {
		t.Errorf("expected from sender@test.org, got %s", entry.From)
	}
}

func TestParser_ParseLine_EmptyFrom(t *testing.T) {
	parser := NewParser()
	// Bounce messages have empty from=<>
	line := "Feb 14 10:23:45 mail postfix/smtp[1234]: ABCDEF1234: to=<sender@example.com>, from=<>, relay=example.com[1.2.3.4]:25, status=sent (250 OK)"

	entry, err := parser.ParseLine(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry == nil {
		t.Fatal("expected entry, got nil")
	}

	if entry.From != "" {
		t.Errorf("expected empty from, got %s", entry.From)
	}
}

func TestParser_ParseLine_MalformedLine(t *testing.T) {
	parser := NewParser()

	testCases := []string{
		"",                    // Empty
		"short line",          // Too short
		"Feb 14 status=sent",  // Missing queue ID context
	}

	for _, line := range testCases {
		entry, _ := parser.ParseLine(line)
		// Should either return nil with no error (skipped) or an error
		if entry != nil {
			t.Errorf("expected nil entry for malformed line '%s', got %+v", line, entry)
		}
		// We allow errors for malformed lines
	}
}

func TestParser_ParseLine_MultipleRecipients(t *testing.T) {
	parser := NewParser()
	// Postfix logs one line per recipient
	line1 := "Feb 14 10:23:45 mail postfix/smtp[1234]: ABCDEF1234: to=<user1@example.com>, relay=example.com[1.2.3.4]:25, status=sent (250 OK)"
	line2 := "Feb 14 10:23:46 mail postfix/smtp[1234]: ABCDEF1234: to=<user2@example.com>, relay=example.com[1.2.3.4]:25, status=sent (250 OK)"

	entry1, err := parser.ParseLine(line1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry2, err := parser.ParseLine(line2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both should have the same queue ID but different recipients
	if entry1.QueueID != entry2.QueueID {
		t.Errorf("expected same queue ID, got %s and %s", entry1.QueueID, entry2.QueueID)
	}

	if entry1.To == entry2.To {
		t.Error("expected different recipients")
	}
}

func TestParseTimestampAndQueueID(t *testing.T) {
	line := "Feb 14 10:23:45 mail postfix/smtp[1234]: ABCDEF1234: to=<user@example.com>"

	timestamp, queueID, err := parseTimestampAndQueueID(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if queueID != "ABCDEF1234" {
		t.Errorf("expected queue_id ABCDEF1234, got %s", queueID)
	}

	if timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}

	// Check timestamp components
	if timestamp.Month().String() != "February" {
		t.Errorf("expected February, got %s", timestamp.Month())
	}
	if timestamp.Day() != 14 {
		t.Errorf("expected day 14, got %d", timestamp.Day())
	}
	if timestamp.Hour() != 10 {
		t.Errorf("expected hour 10, got %d", timestamp.Hour())
	}
	if timestamp.Minute() != 23 {
		t.Errorf("expected minute 23, got %d", timestamp.Minute())
	}
	if timestamp.Second() != 45 {
		t.Errorf("expected second 45, got %d", timestamp.Second())
	}
}
