// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package server provides an mTLS listener for the cloud relay side of
// DarkPipe's alternative transport. The server requires mutual TLS -- every
// connecting client must present a certificate signed by the same internal CA.
//
// TLS configuration deliberately relies on Go's defaults for cipher suites
// so that TLS 1.3 (including post-quantum key exchange in Go 1.24+) is
// negotiated automatically.
package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

// Server is an mTLS listener that enforces RequireAndVerifyClientCert.
type Server struct {
	listenAddr string
	tlsConfig  *tls.Config

	mu       sync.Mutex
	listener net.Listener
	closed   bool
}

// NewServer creates a new mTLS server. It loads the CA certificate used to
// verify connecting clients and the server's own certificate/key pair.
//
// caCertPath:     PEM file containing the CA certificate (used to verify client certs)
// serverCertPath: PEM file containing the server's certificate
// serverKeyPath:  PEM file containing the server's private key
func NewServer(listenAddr, caCertPath, serverCertPath, serverKeyPath string) (*Server, error) {
	// Load CA certificate for client verification.
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("load CA cert %s: %w", caCertPath, err)
	}

	clientCAs := x509.NewCertPool()
	if !clientCAs.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate from %s", caCertPath)
	}

	// Load server certificate and key.
	serverCert, err := tls.LoadX509KeyPair(serverCertPath, serverKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load server cert/key: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientCAs,
		MinVersion:   tls.VersionTLS12,
		// Cipher suites left as default -- Go negotiates TLS 1.3 automatically.
	}

	return &Server{
		listenAddr: listenAddr,
		tlsConfig:  tlsConfig,
	}, nil
}

// Listen returns a TLS listener bound to the server's address.
// Callers are responsible for closing the listener.
func (s *Server) Listen() (net.Listener, error) {
	ln, err := tls.Listen("tcp", s.listenAddr, s.tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("tls listen on %s: %w", s.listenAddr, err)
	}

	s.mu.Lock()
	s.listener = ln
	s.mu.Unlock()

	return ln, nil
}

// Serve accepts connections in a loop and dispatches each to handler in a new
// goroutine. It returns when the listener is closed. Connection-level errors
// are logged but never crash the server.
func (s *Server) Serve(handler func(net.Conn)) error {
	ln, err := s.Listen()
	if err != nil {
		return err
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			s.mu.Lock()
			closed := s.closed
			s.mu.Unlock()
			if closed {
				return nil // clean shutdown
			}
			log.Printf("mtls server: accept error: %v", err)
			continue
		}
		go handler(conn)
	}
}

// Close shuts down the listener if one is active.
func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closed = true
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
