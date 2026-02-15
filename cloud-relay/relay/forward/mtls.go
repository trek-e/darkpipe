// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package forward

import (
	"context"
	"fmt"
	"io"

	"github.com/darkpipe/darkpipe/transport/mtls/client"
	"github.com/emersion/go-smtp"
)

// MTLSForwarder forwards mail to the home device via Phase 1's mTLS transport.
type MTLSForwarder struct {
	mtlsClient *client.Client
	homeAddr   string
}

// NewMTLSForwarder creates a forwarder that uses mTLS transport.
func NewMTLSForwarder(caCert, clientCert, clientKey, homeAddr string) (*MTLSForwarder, error) {
	mtlsClient, err := client.NewClient(homeAddr, caCert, clientCert, clientKey)
	if err != nil {
		return nil, fmt.Errorf("create mTLS client: %w", err)
	}

	return &MTLSForwarder{
		mtlsClient: mtlsClient,
		homeAddr:   homeAddr,
	}, nil
}

// Forward sends the mail envelope via mTLS connection to home device's SMTP server.
func (f *MTLSForwarder) Forward(ctx context.Context, from string, to []string, data io.Reader) error {
	// Establish mTLS connection
	conn, err := f.mtlsClient.Connect(ctx)
	if err != nil {
		return fmt.Errorf("mTLS connect: %w", err)
	}
	defer conn.Close()

	// Create SMTP client over the mTLS connection
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

// Close releases mTLS resources.
func (f *MTLSForwarder) Close() error {
	// Phase 1 mTLS client doesn't expose a Close method (it's stateless).
	// No cleanup needed.
	return nil
}
