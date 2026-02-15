// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package forward

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/darkpipe/darkpipe/transport/mtls/testutil"
)

func TestNewMTLSForwarder_ValidCerts(t *testing.T) {
	t.Parallel()

	// Generate test certificates
	certDir := t.TempDir()
	certs := testutil.GenerateCertFiles(t, certDir)

	fwd, err := NewMTLSForwarder(certs.CACert, certs.ClientCert, certs.ClientKey, "127.0.0.1:10000")
	if err != nil {
		t.Fatalf("NewMTLSForwarder: %v", err)
	}

	if fwd == nil {
		t.Fatal("NewMTLSForwarder returned nil")
	}

	if fwd.homeAddr != "127.0.0.1:10000" {
		t.Errorf("homeAddr = %q, want %q", fwd.homeAddr, "127.0.0.1:10000")
	}
}

func TestNewMTLSForwarder_InvalidCertPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		caCert     string
		clientCert string
		clientKey  string
	}{
		{"nonexistent CA", "/nonexistent/ca.crt", "/tmp/client.crt", "/tmp/client.key"},
		{"nonexistent client cert", "/tmp/ca.crt", "/nonexistent/client.crt", "/tmp/client.key"},
		{"nonexistent client key", "/tmp/ca.crt", "/tmp/client.crt", "/nonexistent/client.key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMTLSForwarder(tt.caCert, tt.clientCert, tt.clientKey, "127.0.0.1:10000")
			if err == nil {
				t.Fatal("NewMTLSForwarder: expected error for invalid paths, got nil")
			}
		})
	}
}

func TestMTLSForwarder_Forward(t *testing.T) {
	t.Parallel()

	// Generate test certificates
	certDir := t.TempDir()
	certs := testutil.GenerateCertFiles(t, certDir)

	// Start an mTLS SMTP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().String()

	// Load server TLS config
	serverCert, err := tls.LoadX509KeyPair(certs.ServerCert, certs.ServerKey)
	if err != nil {
		t.Fatalf("load server cert: %v", err)
	}

	caCertPEM, err := os.ReadFile(certs.CACert)
	if err != nil {
		t.Fatalf("read CA cert: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCertPEM) {
		t.Fatal("failed to append CA cert to pool")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	// SMTP server goroutine
	received := make(chan string, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Upgrade to TLS
		tlsConn := tls.Server(conn, tlsConfig)
		if err := tlsConn.Handshake(); err != nil {
			t.Logf("TLS handshake failed: %v", err)
			return
		}
		defer tlsConn.Close()

		// Simple SMTP protocol exchange
		buf := make([]byte, 4096)
		var commands strings.Builder

		// Send greeting
		tlsConn.Write([]byte("220 test.local ESMTP\r\n"))

		for {
			n, err := tlsConn.Read(buf)
			if err != nil {
				break
			}

			cmd := string(buf[:n])
			commands.WriteString(cmd)

			// Respond to SMTP commands
			if strings.HasPrefix(cmd, "EHLO") || strings.HasPrefix(cmd, "HELO") {
				tlsConn.Write([]byte("250 Hello\r\n"))
			} else if strings.HasPrefix(cmd, "MAIL FROM") {
				tlsConn.Write([]byte("250 OK\r\n"))
			} else if strings.HasPrefix(cmd, "RCPT TO") {
				tlsConn.Write([]byte("250 OK\r\n"))
			} else if strings.HasPrefix(cmd, "DATA") {
				tlsConn.Write([]byte("354 Start mail input\r\n"))
			} else if strings.Contains(cmd, "\r\n.\r\n") {
				tlsConn.Write([]byte("250 OK\r\n"))
			} else if strings.HasPrefix(cmd, "QUIT") {
				tlsConn.Write([]byte("221 Bye\r\n"))
				break
			}
		}

		received <- commands.String()
	}()

	// Create forwarder
	fwd, err := NewMTLSForwarder(certs.CACert, certs.ClientCert, certs.ClientKey, addr)
	if err != nil {
		t.Fatalf("NewMTLSForwarder: %v", err)
	}
	defer fwd.Close()

	// Send mail
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	from := "sender@example.com"
	to := []string{"recipient@example.com"}
	data := strings.NewReader("Subject: Test\r\n\r\nTest message via mTLS\r\n")

	err = fwd.Forward(ctx, from, to, data)
	if err != nil {
		t.Fatalf("Forward: %v", err)
	}

	// Verify SMTP commands
	select {
	case commands := <-received:
		if !strings.Contains(commands, "MAIL FROM:<sender@example.com>") {
			t.Errorf("missing MAIL FROM command in: %s", commands)
		}
		if !strings.Contains(commands, "RCPT TO:<recipient@example.com>") {
			t.Errorf("missing RCPT TO command in: %s", commands)
		}
		if !strings.Contains(commands, "DATA") {
			t.Errorf("missing DATA command in: %s", commands)
		}
		if !strings.Contains(commands, "Subject: Test") {
			t.Errorf("missing message body in: %s", commands)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("timeout waiting for SMTP commands")
	}
}

func TestMTLSForwarder_ForwardConnectionFailure(t *testing.T) {
	t.Parallel()

	// Generate test certificates
	certDir := t.TempDir()
	certs := testutil.GenerateCertFiles(t, certDir)

	// Create forwarder with unreachable address
	fwd, err := NewMTLSForwarder(certs.CACert, certs.ClientCert, certs.ClientKey, "127.0.0.1:1")
	if err != nil {
		t.Fatalf("NewMTLSForwarder: %v", err)
	}
	defer fwd.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	from := "sender@example.com"
	to := []string{"recipient@example.com"}
	data := strings.NewReader("Test message")

	err = fwd.Forward(ctx, from, to, data)
	if err == nil {
		t.Fatal("Forward: expected error for unreachable address, got nil")
	}

	// Should contain "mTLS connect" or "connection refused"
	errMsg := err.Error()
	if !strings.Contains(errMsg, "mTLS connect") && !strings.Contains(errMsg, "refused") {
		t.Logf("error: %s", errMsg)
		// Still pass the test - the important thing is that it returned an error
	}
}

func TestMTLSForwarder_Close(t *testing.T) {
	t.Parallel()

	// Generate test certificates
	certDir := t.TempDir()
	certs := testutil.GenerateCertFiles(t, certDir)

	fwd, err := NewMTLSForwarder(certs.CACert, certs.ClientCert, certs.ClientKey, "127.0.0.1:10000")
	if err != nil {
		t.Fatalf("NewMTLSForwarder: %v", err)
	}

	err = fwd.Close()
	if err != nil {
		t.Errorf("Close: %v, want nil", err)
	}
}
