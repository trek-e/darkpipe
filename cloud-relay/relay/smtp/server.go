// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package smtp provides the SMTP server backend for the cloud relay daemon.
package smtp

import (
	"log"

	"github.com/darkpipe/darkpipe/cloud-relay/relay/config"
	"github.com/darkpipe/darkpipe/cloud-relay/relay/forward"
	"github.com/emersion/go-smtp"
)

// Backend implements smtp.Backend from emersion/go-smtp.
type Backend struct {
	forwarder forward.Forwarder
	debug     bool
}

// NewBackend creates a new SMTP backend.
func NewBackend(forwarder forward.Forwarder, debug bool) *Backend {
	return &Backend{
		forwarder: forwarder,
		debug:     debug,
	}
}

// NewSession creates a new SMTP session.
func (b *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &Session{
		forwarder: b.forwarder,
		debug:     b.debug,
	}, nil
}

// NewServer creates a configured SMTP server.
func NewServer(forwarder forward.Forwarder, cfg *config.Config) *smtp.Server {
	backend := NewBackend(forwarder, cfg.RelayDebug)

	s := smtp.NewServer(backend)
	s.Addr = cfg.ListenAddr
	s.Domain = "relay.darkpipe.local"
	s.ReadTimeout = cfg.ReadTimeout
	s.WriteTimeout = cfg.WriteTimeout
	s.MaxMessageBytes = cfg.MaxMessageBytes
	s.MaxRecipients = 100
	s.AllowInsecureAuth = true // We're localhost-only, Postfix handles internet-facing SMTP

	s.ErrorLog = log.Default()

	return s
}
