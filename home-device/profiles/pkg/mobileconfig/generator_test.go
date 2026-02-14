package mobileconfig

import (
	"strings"
	"testing"

	"github.com/micromdm/plist"
)

func TestGenerateProfile_WithAllPayloads(t *testing.T) {
	gen := &ProfileGenerator{}
	cfg := ProfileConfig{
		Domain:       "example.com",
		MailHostname: "mail.example.com",
		Email:        "user@example.com",
		AppPassword:  "TEST-PASS-WORD-1234",
		CalDAVURL:    "https://mail.example.com:5232/user@example.com/",
		CardDAVURL:   "https://mail.example.com:5232/user@example.com/",
		CalDAVPort:   5232,
		CardDAVPort:  5232,
	}

	data, err := gen.GenerateProfile(cfg)
	if err != nil {
		t.Fatalf("GenerateProfile failed: %v", err)
	}

	// Parse back to verify structure
	var profile MobileConfigProfile
	if err := plist.Unmarshal(data, &profile); err != nil {
		t.Fatalf("Unmarshal plist failed: %v", err)
	}

	// Verify profile metadata
	if profile.PayloadType != "Configuration" {
		t.Errorf("PayloadType = %s, want Configuration", profile.PayloadType)
	}

	if profile.PayloadOrganization != "DarkPipe" {
		t.Errorf("PayloadOrganization = %s, want DarkPipe", profile.PayloadOrganization)
	}

	if profile.PayloadVersion != 1 {
		t.Errorf("PayloadVersion = %d, want 1", profile.PayloadVersion)
	}

	// Verify 3 payloads (Email + CalDAV + CardDAV)
	if len(profile.PayloadContent) != 3 {
		t.Fatalf("PayloadContent length = %d, want 3", len(profile.PayloadContent))
	}

	// Verify data contains expected strings
	dataStr := string(data)
	if !strings.Contains(dataStr, "com.apple.mail.managed") {
		t.Error("Profile missing Email payload type")
	}

	if !strings.Contains(dataStr, "com.apple.caldav.account") {
		t.Error("Profile missing CalDAV payload type")
	}

	if !strings.Contains(dataStr, "com.apple.carddav.account") {
		t.Error("Profile missing CardDAV payload type")
	}

	if !strings.Contains(dataStr, "mail.example.com") {
		t.Error("Profile missing mail hostname")
	}

	if !strings.Contains(dataStr, "user@example.com") {
		t.Error("Profile missing email address")
	}

	if !strings.Contains(dataStr, "TEST-PASS-WORD-1234") {
		t.Error("Profile missing app password")
	}

	// Verify port numbers
	if !strings.Contains(dataStr, "<integer>993</integer>") {
		t.Error("Profile missing IMAP port 993")
	}

	if !strings.Contains(dataStr, "<integer>587</integer>") {
		t.Error("Profile missing SMTP port 587")
	}

	if !strings.Contains(dataStr, "<integer>5232</integer>") {
		t.Error("Profile missing CalDAV/CardDAV port 5232")
	}
}

func TestGenerateProfile_EmailOnly(t *testing.T) {
	gen := &ProfileGenerator{}
	cfg := ProfileConfig{
		Domain:       "example.com",
		MailHostname: "mail.example.com",
		Email:        "user@example.com",
		AppPassword:  "TEST-PASS-WORD-1234",
		// No CalDAV/CardDAV
	}

	data, err := gen.GenerateProfile(cfg)
	if err != nil {
		t.Fatalf("GenerateProfile failed: %v", err)
	}

	// Parse back
	var profile MobileConfigProfile
	if err := plist.Unmarshal(data, &profile); err != nil {
		t.Fatalf("Unmarshal plist failed: %v", err)
	}

	// Verify only 1 payload (Email)
	if len(profile.PayloadContent) != 1 {
		t.Errorf("PayloadContent length = %d, want 1", len(profile.PayloadContent))
	}

	// Verify no CalDAV/CardDAV in output
	dataStr := string(data)
	if strings.Contains(dataStr, "caldav") {
		t.Error("Profile should not contain CalDAV")
	}

	if strings.Contains(dataStr, "carddav") {
		t.Error("Profile should not contain CardDAV")
	}
}

func TestGenerateProfile_Identifiers(t *testing.T) {
	gen := &ProfileGenerator{}
	cfg := ProfileConfig{
		Domain:       "test.org",
		MailHostname: "mail.test.org",
		Email:        "admin@test.org",
		AppPassword:  "ABCD-EFGH-JKLM-NPQR",
		CalDAVURL:    "https://mail.test.org/caldav/",
		CalDAVPort:   443,
	}

	data, err := gen.GenerateProfile(cfg)
	if err != nil {
		t.Fatalf("GenerateProfile failed: %v", err)
	}

	dataStr := string(data)

	// Verify domain-specific identifiers
	if !strings.Contains(dataStr, "com.darkpipe.test.org") {
		t.Error("Profile missing domain-specific identifier")
	}

	if !strings.Contains(dataStr, "com.darkpipe.test.org.email") {
		t.Error("Profile missing email payload identifier")
	}

	if !strings.Contains(dataStr, "com.darkpipe.test.org.caldav") {
		t.Error("Profile missing CalDAV payload identifier")
	}
}

func TestGenerateProfile_SSLSettings(t *testing.T) {
	gen := &ProfileGenerator{}
	cfg := ProfileConfig{
		Domain:       "example.com",
		MailHostname: "mail.example.com",
		Email:        "user@example.com",
		AppPassword:  "TEST-PASS-WORD-1234",
		CalDAVURL:    "https://mail.example.com/caldav/",
		CalDAVPort:   443, // Should set UseSSL=true
	}

	data, err := gen.GenerateProfile(cfg)
	if err != nil {
		t.Fatalf("GenerateProfile failed: %v", err)
	}

	dataStr := string(data)

	// IMAP should use SSL (port 993)
	if !strings.Contains(dataStr, "<key>IncomingMailServerUseSSL</key>") {
		t.Error("Profile missing IncomingMailServerUseSSL key")
	}

	// CalDAV on port 443 should use SSL
	if cfg.CalDAVPort == 443 && !strings.Contains(dataStr, "<key>CalDAVUseSSL</key>") {
		t.Error("Profile missing CalDAVUseSSL key")
	}
}
