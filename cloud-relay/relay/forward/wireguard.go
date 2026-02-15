// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package forward

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/emersion/go-smtp"
)

// WireGuardForwarder forwards mail to the home device via WireGuard tunnel.
// The tunnel provides transparent encryption at the network layer, so this
// forwarder just dials the home device's SMTP server directly through the
// tunnel.
type WireGuardForwarder struct {
	homeAddr string
}

// NewWireGuardForwarder creates a forwarder that uses WireGuard transport.
func NewWireGuardForwarder(homeAddr string) *WireGuardForwarder {
	return &WireGuardForwarder{
		homeAddr: homeAddr,
	}
}

// Forward sends the mail envelope via WireGuard tunnel to home device's SMTP server.
func (f *WireGuardForwarder) Forward(ctx context.Context, from string, to []string, data io.Reader) error {
	// Dial home device's SMTP server through WireGuard tunnel.
	// The WireGuard kernel module handles encryption transparently.
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", f.homeAddr)
	if err != nil {
		return fmt.Errorf("dial via WireGuard: %w", err)
	}
	defer conn.Close()

	// Create SMTP client over the WireGuard connection
	smtpClient := smtp.NewClient(conn)
	defer smtpClient.Close()

	// SMTP HELO/EHLO
	if err := smtpClient.Hello("relay.darkpipe.local"); err != nil {
		return fmt.Errorf("HELO: %w", err)
	}

	// Send MAIL FROM
	if err := smtpClient.Mail(from, nil); err != nil {
		return fmt.Errorf("MAIL FROM: %w", err)
	}

	// Send RCPT TO for each recipient
	for _, rcpt := range to {
		if err := smtpClient.Rcpt(rcpt, nil); err != nil {
			return fmt.Errorf("RCPT TO %s: %w", rcpt, err)
		}
	}

	// Send DATA
	w, err := smtpClient.Data()
	if err != nil {
		return fmt.Errorf("DATA: %w", err)
	}
	defer w.Close()

	if _, err := io.Copy(w, data); err != nil {
		return fmt.Errorf("write message data: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("complete DATA: %w", err)
	}

	return smtpClient.Quit()
}

// Close releases resources (none needed for WireGuard).
func (f *WireGuardForwarder) Close() error {
	return nil
}
