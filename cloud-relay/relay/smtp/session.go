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
	"github.com/darkpipe/darkpipe/cloud-relay/relay/logutil"
	"github.com/emersion/go-smtp"
)

// Session implements smtp.Session from emersion/go-smtp.
type Session struct {
	forwarder forward.Forwarder
	from      string
	to        []string
	debug     bool
}

// logFrom returns the from address for logging — redacted unless debug mode.
func (s *Session) logFrom() string {
	if s.debug {
		return s.from
	}
	return logutil.RedactEmail(s.from)
}

// logTo returns the recipient list for logging — redacted unless debug mode.
func (s *Session) logTo() []string {
	if s.debug {
		return s.to
	}
	return logutil.RedactEmails(s.to)
}

// Mail is called when the client sends MAIL FROM.
func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	s.from = from
	log.Printf("MAIL FROM: %s", s.logFrom())
	return nil
}

// Rcpt is called when the client sends RCPT TO.
func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.to = append(s.to, to)
	if s.debug {
		log.Printf("RCPT TO: %s", to)
	} else {
		log.Printf("RCPT TO: %s", logutil.RedactEmail(to))
	}
	return nil
}

// Data is called when the client sends DATA. This is where we forward
// the mail to the home device via the configured transport.
func (s *Session) Data(r io.Reader) error {
	log.Printf("DATA: forwarding from=%s to=%v via transport", s.logFrom(), s.logTo())

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

	log.Printf("SUCCESS: forwarded message from=%s to=%v", s.logFrom(), s.logTo())
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
