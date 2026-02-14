package autoconfig

import (
	"encoding/xml"
	"strings"
	"testing"
)

// Structs for parsing autoconfig XML
type ClientConfig struct {
	XMLName       xml.Name      `xml:"clientConfig"`
	Version       string        `xml:"version,attr"`
	EmailProvider EmailProvider `xml:"emailProvider"`
}

type EmailProvider struct {
	ID              string          `xml:"id,attr"`
	Domain          string          `xml:"domain"`
	DisplayName     string          `xml:"displayName"`
	IncomingServer  IncomingServer  `xml:"incomingServer"`
	OutgoingServer  OutgoingServer  `xml:"outgoingServer"`
}

type IncomingServer struct {
	Type           string `xml:"type,attr"`
	Hostname       string `xml:"hostname"`
	Port           int    `xml:"port"`
	SocketType     string `xml:"socketType"`
	Authentication string `xml:"authentication"`
	Username       string `xml:"username"`
}

type OutgoingServer struct {
	Type           string `xml:"type,attr"`
	Hostname       string `xml:"hostname"`
	Port           int    `xml:"port"`
	SocketType     string `xml:"socketType"`
	Authentication string `xml:"authentication"`
	Username       string `xml:"username"`
}

func TestGenerateAutoconfig_Basic(t *testing.T) {
	data, err := GenerateAutoconfig("example.com", "mail.example.com")
	if err != nil {
		t.Fatalf("GenerateAutoconfig failed: %v", err)
	}

	// Parse XML
	var config ClientConfig
	if err := xml.Unmarshal(data, &config); err != nil {
		t.Fatalf("Unmarshal XML failed: %v", err)
	}

	// Verify version
	if config.Version != "1.1" {
		t.Errorf("Version = %s, want 1.1", config.Version)
	}

	// Verify domain
	if config.EmailProvider.Domain != "example.com" {
		t.Errorf("Domain = %s, want example.com", config.EmailProvider.Domain)
	}
}

func TestGenerateAutoconfig_IMAPSettings(t *testing.T) {
	data, err := GenerateAutoconfig("test.org", "mail.test.org")
	if err != nil {
		t.Fatalf("GenerateAutoconfig failed: %v", err)
	}

	var config ClientConfig
	if err := xml.Unmarshal(data, &config); err != nil {
		t.Fatalf("Unmarshal XML failed: %v", err)
	}

	incoming := config.EmailProvider.IncomingServer

	// Verify IMAP settings
	if incoming.Type != "imap" {
		t.Errorf("IncomingServer Type = %s, want imap", incoming.Type)
	}

	if incoming.Hostname != "mail.test.org" {
		t.Errorf("IncomingServer Hostname = %s, want mail.test.org", incoming.Hostname)
	}

	if incoming.Port != 993 {
		t.Errorf("IncomingServer Port = %d, want 993", incoming.Port)
	}

	if incoming.SocketType != "SSL" {
		t.Errorf("IncomingServer SocketType = %s, want SSL", incoming.SocketType)
	}

	if incoming.Authentication != "password-cleartext" {
		t.Errorf("IncomingServer Authentication = %s, want password-cleartext", incoming.Authentication)
	}

	if incoming.Username != "%EMAILADDRESS%" {
		t.Errorf("IncomingServer Username = %s, want %%EMAILADDRESS%%", incoming.Username)
	}
}

func TestGenerateAutoconfig_SMTPSettings(t *testing.T) {
	data, err := GenerateAutoconfig("test.org", "mail.test.org")
	if err != nil {
		t.Fatalf("GenerateAutoconfig failed: %v", err)
	}

	var config ClientConfig
	if err := xml.Unmarshal(data, &config); err != nil {
		t.Fatalf("Unmarshal XML failed: %v", err)
	}

	outgoing := config.EmailProvider.OutgoingServer

	// Verify SMTP settings
	if outgoing.Type != "smtp" {
		t.Errorf("OutgoingServer Type = %s, want smtp", outgoing.Type)
	}

	if outgoing.Hostname != "mail.test.org" {
		t.Errorf("OutgoingServer Hostname = %s, want mail.test.org", outgoing.Hostname)
	}

	if outgoing.Port != 587 {
		t.Errorf("OutgoingServer Port = %d, want 587", outgoing.Port)
	}

	if outgoing.SocketType != "STARTTLS" {
		t.Errorf("OutgoingServer SocketType = %s, want STARTTLS", outgoing.SocketType)
	}

	if outgoing.Authentication != "password-cleartext" {
		t.Errorf("OutgoingServer Authentication = %s, want password-cleartext", outgoing.Authentication)
	}

	if outgoing.Username != "%EMAILADDRESS%" {
		t.Errorf("OutgoingServer Username = %s, want %%EMAILADDRESS%%", outgoing.Username)
	}
}

func TestGenerateAutoconfig_XMLFormat(t *testing.T) {
	data, err := GenerateAutoconfig("example.com", "mail.example.com")
	if err != nil {
		t.Fatalf("GenerateAutoconfig failed: %v", err)
	}

	dataStr := string(data)

	// Verify XML declaration
	if !strings.HasPrefix(dataStr, `<?xml version="1.0" encoding="UTF-8"?>`) {
		t.Error("Missing XML declaration")
	}

	// Verify contains expected elements
	expectedElements := []string{
		"<clientConfig",
		"<emailProvider",
		"<incomingServer",
		"<outgoingServer",
		"<hostname>mail.example.com</hostname>",
		"<domain>example.com</domain>",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(dataStr, elem) {
			t.Errorf("XML missing element: %s", elem)
		}
	}
}
