// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package authtest

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/darkpipe/darkpipe/dns/dkim"
	"github.com/emersion/go-smtp"
)

// TestEmailConfig holds configuration for sending a test email.
type TestEmailConfig struct {
	From       string       // Sender email address
	To         string       // Recipient email address
	RelayHost  string       // SMTP relay hostname
	RelayPort  int          // SMTP relay port (usually 25 or 587)
	DKIMKey    *rsa.PrivateKey // DKIM private key for signing
	DKIMDomain string       // Domain for DKIM signing
	DKIMSelector string     // DKIM selector
}

// SendTestEmail sends a test email for end-to-end authentication verification.
// The email is DKIM-signed and sent via the cloud relay.
// Returns instructions for checking the received email's Authentication-Results header.
func SendTestEmail(ctx context.Context, cfg TestEmailConfig) error {
	// Validate config
	if cfg.From == "" {
		return fmt.Errorf("From address is required")
	}
	if cfg.To == "" {
		return fmt.Errorf("To address is required")
	}
	if cfg.RelayHost == "" {
		return fmt.Errorf("RelayHost is required")
	}
	if cfg.RelayPort == 0 {
		cfg.RelayPort = 25 // Default SMTP port
	}

	// Construct email message
	messageID := fmt.Sprintf("<%d@%s>", time.Now().Unix(), cfg.DKIMDomain)
	date := time.Now().Format(time.RFC1123Z)

	var msgBuilder strings.Builder
	msgBuilder.WriteString(fmt.Sprintf("From: %s\r\n", cfg.From))
	msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", cfg.To))
	msgBuilder.WriteString("Subject: DarkPipe DNS Authentication Test\r\n")
	msgBuilder.WriteString(fmt.Sprintf("Date: %s\r\n", date))
	msgBuilder.WriteString(fmt.Sprintf("Message-ID: %s\r\n", messageID))
	msgBuilder.WriteString("MIME-Version: 1.0\r\n")
	msgBuilder.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString("This is an automated test email from DarkPipe to verify email authentication.\n\n")
	msgBuilder.WriteString("If you see this message, the email delivery is working.\n\n")
	msgBuilder.WriteString("To verify SPF, DKIM, and DMARC authentication:\n")
	msgBuilder.WriteString("1. View the full email headers (\"Show original\" in Gmail, \"View source\" in others)\n")
	msgBuilder.WriteString("2. Find the Authentication-Results header\n")
	msgBuilder.WriteString("3. Verify that SPF, DKIM, and DMARC all show \"pass\"\n\n")
	msgBuilder.WriteString("Example of what you should see:\n")
	msgBuilder.WriteString("Authentication-Results: mx.google.com;\n")
	msgBuilder.WriteString("  dkim=pass header.i=@yourdomain.com header.s=darkpipe-2026q1;\n")
	msgBuilder.WriteString("  spf=pass smtp.mailfrom=yourdomain.com;\n")
	msgBuilder.WriteString("  dmarc=pass (policy=none) header.from=yourdomain.com\n\n")
	msgBuilder.WriteString("--\n")
	msgBuilder.WriteString("Sent by DarkPipe dns-setup --send-test\n")

	message := msgBuilder.String()

	// Sign with DKIM if key is provided
	if cfg.DKIMKey != nil && cfg.DKIMDomain != "" && cfg.DKIMSelector != "" {
		signer := dkim.NewSigner(cfg.DKIMDomain, cfg.DKIMSelector, cfg.DKIMKey)

		signedMessage, err := signer.Sign(strings.NewReader(message))
		if err != nil {
			return fmt.Errorf("failed to sign message with DKIM: %w", err)
		}

		message = string(signedMessage)
	}

	// Send via SMTP
	addr := fmt.Sprintf("%s:%d", cfg.RelayHost, cfg.RelayPort)
	if err := smtp.SendMail(addr, nil, cfg.From, []string{cfg.To}, strings.NewReader(message)); err != nil {
		return fmt.Errorf("failed to send email via %s: %w", addr, err)
	}

	fmt.Printf("✓ Test email sent successfully!\n\n")
	fmt.Printf("From: %s\n", cfg.From)
	fmt.Printf("To: %s\n", cfg.To)
	fmt.Printf("Message-ID: %s\n\n", messageID)
	fmt.Printf("Next steps:\n")
	fmt.Printf("1. Check the inbox for %s\n", cfg.To)
	fmt.Printf("2. View the full email headers (\"Show original\" in Gmail)\n")
	fmt.Printf("3. Find the Authentication-Results header\n")
	fmt.Printf("4. Verify SPF=pass, DKIM=pass, DMARC=pass\n\n")
	fmt.Printf("If any authentication checks fail, run:\n")
	fmt.Printf("  darkpipe dns-setup --validate-only\n\n")

	return nil
}

// parseAuthResults is a helper that can be called after receiving a test email
// to parse and display the authentication results in a human-friendly format.
func DisplayAuthReport(report AuthReport, w io.Writer) {
	fmt.Fprintf(w, "Authentication Results:\n\n")

	for _, result := range report.Results {
		status := "✗ FAIL"
		if result.Result == "pass" {
			status = "✓ PASS"
		}

		fmt.Fprintf(w, "%s %s: %s\n", status, strings.ToUpper(result.Method), result.Result)
		if result.Details != "" {
			fmt.Fprintf(w, "  Details: %s\n", result.Details)
		}
	}

	fmt.Fprintf(w, "\nSummary:\n")
	fmt.Fprintf(w, "  SPF:   %s\n", passFailStatus(report.SPFPass))
	fmt.Fprintf(w, "  DKIM:  %s\n", passFailStatus(report.DKIMPass))
	fmt.Fprintf(w, "  DMARC: %s\n", passFailStatus(report.DMARCPass))

	if report.SPFPass && report.DKIMPass && report.DMARCPass {
		fmt.Fprintf(w, "\n✓ All authentication checks passed!\n")
	} else {
		fmt.Fprintf(w, "\n✗ Some authentication checks failed. Run validation:\n")
		fmt.Fprintf(w, "  darkpipe dns-setup --validate-only\n")
	}
}

func passFailStatus(pass bool) string {
	if pass {
		return "✓ PASS"
	}
	return "✗ FAIL"
}
