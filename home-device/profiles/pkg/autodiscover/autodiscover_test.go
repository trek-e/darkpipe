package autodiscover

import (
	"encoding/xml"
	"strings"
	"testing"
)

// Structs for parsing autodiscover XML
type Autodiscover struct {
	XMLName  xml.Name `xml:"Autodiscover"`
	Response Response `xml:"Response"`
}

type Response struct {
	Account Account `xml:"Account"`
}

type Account struct {
	AccountType string     `xml:"AccountType"`
	Action      string     `xml:"Action"`
	Protocols   []Protocol `xml:"Protocol"`
}

type Protocol struct {
	Type           string `xml:"Type"`
	Server         string `xml:"Server"`
	Port           int    `xml:"Port"`
	DomainRequired string `xml:"DomainRequired"`
	LoginName      string `xml:"LoginName"`
	SPA            string `xml:"SPA"`
	SSL            string `xml:"SSL,omitempty"`
	Encryption     string `xml:"Encryption,omitempty"`
	AuthRequired   string `xml:"AuthRequired"`
	UsePOPAuth     string `xml:"UsePOPAuth,omitempty"`
	SMTPLast       string `xml:"SMTPLast,omitempty"`
}

func TestGenerateAutodiscover_Basic(t *testing.T) {
	data, err := GenerateAutodiscover("user@example.com", "mail.example.com")
	if err != nil {
		t.Fatalf("GenerateAutodiscover failed: %v", err)
	}

	// Parse XML
	var autodiscover Autodiscover
	if err := xml.Unmarshal(data, &autodiscover); err != nil {
		t.Fatalf("Unmarshal XML failed: %v", err)
	}

	// Verify account type
	if autodiscover.Response.Account.AccountType != "email" {
		t.Errorf("AccountType = %s, want email", autodiscover.Response.Account.AccountType)
	}

	// Verify action
	if autodiscover.Response.Account.Action != "settings" {
		t.Errorf("Action = %s, want settings", autodiscover.Response.Account.Action)
	}

	// Verify we have 2 protocols (IMAP + SMTP)
	protocols := autodiscover.Response.Account.Protocols
	if len(protocols) != 2 {
		t.Fatalf("Expected 2 protocols, got %d", len(protocols))
	}
}

func TestGenerateAutodiscover_IMAPSettings(t *testing.T) {
	data, err := GenerateAutodiscover("admin@test.org", "mail.test.org")
	if err != nil {
		t.Fatalf("GenerateAutodiscover failed: %v", err)
	}

	var autodiscover Autodiscover
	if err := xml.Unmarshal(data, &autodiscover); err != nil {
		t.Fatalf("Unmarshal XML failed: %v", err)
	}

	// Find IMAP protocol
	var imap *Protocol
	for i, p := range autodiscover.Response.Account.Protocols {
		if p.Type == "IMAP" {
			imap = &autodiscover.Response.Account.Protocols[i]
			break
		}
	}

	if imap == nil {
		t.Fatal("IMAP protocol not found")
	}

	// Verify IMAP settings
	if imap.Server != "mail.test.org" {
		t.Errorf("IMAP Server = %s, want mail.test.org", imap.Server)
	}

	if imap.Port != 993 {
		t.Errorf("IMAP Port = %d, want 993", imap.Port)
	}

	if imap.SSL != "on" {
		t.Errorf("IMAP SSL = %s, want on", imap.SSL)
	}

	if imap.SPA != "off" {
		t.Errorf("IMAP SPA = %s, want off", imap.SPA)
	}

	if imap.LoginName != "admin@test.org" {
		t.Errorf("IMAP LoginName = %s, want admin@test.org", imap.LoginName)
	}

	if imap.AuthRequired != "on" {
		t.Errorf("IMAP AuthRequired = %s, want on", imap.AuthRequired)
	}
}

func TestGenerateAutodiscover_SMTPSettings(t *testing.T) {
	data, err := GenerateAutodiscover("admin@test.org", "mail.test.org")
	if err != nil {
		t.Fatalf("GenerateAutodiscover failed: %v", err)
	}

	var autodiscover Autodiscover
	if err := xml.Unmarshal(data, &autodiscover); err != nil {
		t.Fatalf("Unmarshal XML failed: %v", err)
	}

	// Find SMTP protocol
	var smtp *Protocol
	for i, p := range autodiscover.Response.Account.Protocols {
		if p.Type == "SMTP" {
			smtp = &autodiscover.Response.Account.Protocols[i]
			break
		}
	}

	if smtp == nil {
		t.Fatal("SMTP protocol not found")
	}

	// Verify SMTP settings
	if smtp.Server != "mail.test.org" {
		t.Errorf("SMTP Server = %s, want mail.test.org", smtp.Server)
	}

	if smtp.Port != 587 {
		t.Errorf("SMTP Port = %d, want 587", smtp.Port)
	}

	if smtp.Encryption != "TLS" {
		t.Errorf("SMTP Encryption = %s, want TLS", smtp.Encryption)
	}

	if smtp.SPA != "off" {
		t.Errorf("SMTP SPA = %s, want off", smtp.SPA)
	}

	if smtp.LoginName != "admin@test.org" {
		t.Errorf("SMTP LoginName = %s, want admin@test.org", smtp.LoginName)
	}

	if smtp.AuthRequired != "on" {
		t.Errorf("SMTP AuthRequired = %s, want on", smtp.AuthRequired)
	}
}

func TestGenerateAutodiscover_XMLFormat(t *testing.T) {
	data, err := GenerateAutodiscover("user@example.com", "mail.example.com")
	if err != nil {
		t.Fatalf("GenerateAutodiscover failed: %v", err)
	}

	dataStr := string(data)

	// Verify XML declaration
	if !strings.HasPrefix(dataStr, `<?xml version="1.0" encoding="utf-8"?>`) {
		t.Error("Missing XML declaration")
	}

	// Verify contains expected elements
	expectedElements := []string{
		"<Autodiscover",
		"<Response",
		"<Account>",
		"<Protocol>",
		"<Type>IMAP</Type>",
		"<Type>SMTP</Type>",
		"<Server>mail.example.com</Server>",
		"<LoginName>user@example.com</LoginName>",
		"<Port>993</Port>",
		"<Port>587</Port>",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(dataStr, elem) {
			t.Errorf("XML missing element: %s", elem)
		}
	}
}
