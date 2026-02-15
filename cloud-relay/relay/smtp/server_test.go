// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package smtp

import (
	"testing"
	"time"

	"github.com/darkpipe/darkpipe/cloud-relay/relay/config"
	"github.com/darkpipe/darkpipe/cloud-relay/relay/forward"
)

func TestNewBackend(t *testing.T) {
	t.Parallel()

	mockFwd := forward.NewMockForwarder()
	backend := NewBackend(mockFwd)

	if backend == nil {
		t.Fatal("NewBackend returned nil")
	}

	if backend.forwarder == nil {
		t.Error("backend.forwarder is nil")
	}
}

func TestBackend_NewSession(t *testing.T) {
	t.Parallel()

	mockFwd := forward.NewMockForwarder()
	backend := NewBackend(mockFwd)

	session, err := backend.NewSession(nil)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	if session == nil {
		t.Fatal("NewSession returned nil session")
	}

	// Verify it's the correct type
	s, ok := session.(*Session)
	if !ok {
		t.Errorf("session type = %T, want *Session", session)
	}

	if s.forwarder == nil {
		t.Error("session.forwarder is nil")
	}
}

func TestNewServer(t *testing.T) {
	t.Parallel()

	mockFwd := forward.NewMockForwarder()
	cfg := &config.Config{
		ListenAddr:      "127.0.0.1:10025",
		MaxMessageBytes: 10 * 1024 * 1024,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
	}

	server := NewServer(mockFwd, cfg)

	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	// Verify server configuration
	if server.Addr != cfg.ListenAddr {
		t.Errorf("server.Addr = %q, want %q", server.Addr, cfg.ListenAddr)
	}

	if server.Domain != "relay.darkpipe.local" {
		t.Errorf("server.Domain = %q, want %q", server.Domain, "relay.darkpipe.local")
	}

	if server.MaxMessageBytes != cfg.MaxMessageBytes {
		t.Errorf("server.MaxMessageBytes = %d, want %d", server.MaxMessageBytes, cfg.MaxMessageBytes)
	}

	if server.MaxRecipients != 100 {
		t.Errorf("server.MaxRecipients = %d, want %d", server.MaxRecipients, 100)
	}

	if server.ReadTimeout != cfg.ReadTimeout {
		t.Errorf("server.ReadTimeout = %v, want %v", server.ReadTimeout, cfg.ReadTimeout)
	}

	if server.WriteTimeout != cfg.WriteTimeout {
		t.Errorf("server.WriteTimeout = %v, want %v", server.WriteTimeout, cfg.WriteTimeout)
	}

	if !server.AllowInsecureAuth {
		t.Error("server.AllowInsecureAuth = false, want true (localhost-only)")
	}
}

func TestNewServer_WithDifferentConfig(t *testing.T) {
	t.Parallel()

	mockFwd := forward.NewMockForwarder()
	cfg := &config.Config{
		ListenAddr:      "0.0.0.0:2525",
		MaxMessageBytes: 50 * 1024 * 1024,
		ReadTimeout:     60 * time.Second,
		WriteTimeout:    60 * time.Second,
	}

	server := NewServer(mockFwd, cfg)

	if server.Addr != "0.0.0.0:2525" {
		t.Errorf("server.Addr = %q, want %q", server.Addr, "0.0.0.0:2525")
	}

	if server.MaxMessageBytes != 50*1024*1024 {
		t.Errorf("server.MaxMessageBytes = %d, want %d", server.MaxMessageBytes, 50*1024*1024)
	}

	if server.ReadTimeout != 60*time.Second {
		t.Errorf("server.ReadTimeout = %v, want %v", server.ReadTimeout, 60*time.Second)
	}

	if server.WriteTimeout != 60*time.Second {
		t.Errorf("server.WriteTimeout = %v, want %v", server.WriteTimeout, 60*time.Second)
	}
}
