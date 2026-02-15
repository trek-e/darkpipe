// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package smtp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/darkpipe/darkpipe/cloud-relay/relay/forward"
	"github.com/emersion/go-smtp"
)

// Session implements smtp.Session from emersion/go-smtp.
type Session struct {
	forwarder forward.Forwarder
	from      string
	to        []string
}

// Mail is called when the client sends MAIL FROM.
func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	log.Printf("MAIL FROM: %s", from)
	s.from = from
	return nil
}

// Rcpt is called when the client sends RCPT TO.
func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	log.Printf("RCPT TO: %s", to)
	s.to = append(s.to, to)
	return nil
}

// Data is called when the client sends DATA. This is where we forward
// the mail to the home device via the configured transport.
func (s *Session) Data(r io.Reader) error {
	log.Printf("DATA: forwarding from=%s to=%v via transport", s.from, s.to)

	// Read message data into buffer (needed because we may need to retry
	// and io.Reader is not seekable).
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, r); err != nil {
		log.Printf("ERROR: failed to read message data: %v", err)
		return fmt.Errorf("read message data: %w", err)
	}

	// Forward to home device via transport layer
	ctx := context.Background()
	if err := s.forwarder.Forward(ctx, s.from, s.to, buf); err != nil {
		log.Printf("ERROR: forward failed: %v", err)
		return fmt.Errorf("forward to home device: %w", err)
	}

	log.Printf("SUCCESS: forwarded message from=%s to=%v", s.from, s.to)
	return nil
}

// Reset is called when the client sends RSET.
func (s *Session) Reset() {
	s.from = ""
	s.to = nil
}

// Logout is called when the client sends QUIT.
func (s *Session) Logout() error {
	return nil
}
