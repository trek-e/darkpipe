// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package forward

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

func TestNewWireGuardForwarder(t *testing.T) {
	t.Parallel()

	homeAddr := "10.8.0.2:25"
	fwd := NewWireGuardForwarder(homeAddr)

	if fwd == nil {
		t.Fatal("NewWireGuardForwarder returned nil")
	}

	if fwd.homeAddr != homeAddr {
		t.Errorf("homeAddr = %q, want %q", fwd.homeAddr, homeAddr)
	}
}

func TestWireGuardForwarder_Forward(t *testing.T) {
	t.Parallel()

	// Start a test SMTP server to receive the forwarded mail
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().String()

	// SMTP server goroutine
	received := make(chan string, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Simple SMTP protocol exchange
		buf := make([]byte, 4096)
		var commands strings.Builder

		// Send greeting
		conn.Write([]byte("220 test.local ESMTP\r\n"))

		for {
			n, err := conn.Read(buf)
			if err != nil {
				break
			}

			cmd := string(buf[:n])
			commands.WriteString(cmd)

			// Respond to SMTP commands
			if strings.HasPrefix(cmd, "EHLO") || strings.HasPrefix(cmd, "HELO") {
				conn.Write([]byte("250 Hello\r\n"))
			} else if strings.HasPrefix(cmd, "MAIL FROM") {
				conn.Write([]byte("250 OK\r\n"))
			} else if strings.HasPrefix(cmd, "RCPT TO") {
				conn.Write([]byte("250 OK\r\n"))
			} else if strings.HasPrefix(cmd, "DATA") {
				conn.Write([]byte("354 Start mail input\r\n"))
			} else if strings.Contains(cmd, "\r\n.\r\n") {
				conn.Write([]byte("250 OK\r\n"))
			} else if strings.HasPrefix(cmd, "QUIT") {
				conn.Write([]byte("221 Bye\r\n"))
				break
			}
		}

		received <- commands.String()
	}()

	// Create forwarder and send mail
	fwd := NewWireGuardForwarder(addr)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	from := "sender@example.com"
	to := []string{"recipient1@example.com", "recipient2@example.com"}
	data := strings.NewReader("Subject: Test\r\n\r\nTest message body\r\n")

	err = fwd.Forward(ctx, from, to, data)
	if err != nil {
		t.Fatalf("Forward: %v", err)
	}

	// Verify SMTP commands were sent correctly
	select {
	case commands := <-received:
		if !strings.Contains(commands, "MAIL FROM:<sender@example.com>") {
			t.Errorf("missing MAIL FROM command in: %s", commands)
		}
		if !strings.Contains(commands, "RCPT TO:<recipient1@example.com>") {
			t.Errorf("missing RCPT TO for recipient1 in: %s", commands)
		}
		if !strings.Contains(commands, "RCPT TO:<recipient2@example.com>") {
			t.Errorf("missing RCPT TO for recipient2 in: %s", commands)
		}
		if !strings.Contains(commands, "DATA") {
			t.Errorf("missing DATA command in: %s", commands)
		}
		if !strings.Contains(commands, "Subject: Test") {
			t.Errorf("missing message body in: %s", commands)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for SMTP commands")
	}
}

func TestWireGuardForwarder_ForwardUnreachable(t *testing.T) {
	t.Parallel()

	// Use an unreachable address
	fwd := NewWireGuardForwarder("127.0.0.1:1")

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	from := "sender@example.com"
	to := []string{"recipient@example.com"}
	data := strings.NewReader("Test message")

	err := fwd.Forward(ctx, from, to, data)
	if err == nil {
		t.Fatal("Forward: expected error for unreachable address, got nil")
	}

	// Should contain "dial" or "connection refused"
	errMsg := err.Error()
	if !strings.Contains(errMsg, "dial") && !strings.Contains(errMsg, "refused") {
		t.Errorf("error = %q, want error containing 'dial' or 'refused'", errMsg)
	}
}

func TestWireGuardForwarder_Close(t *testing.T) {
	t.Parallel()

	fwd := NewWireGuardForwarder("10.8.0.2:25")

	err := fwd.Close()
	if err != nil {
		t.Errorf("Close: %v, want nil", err)
	}
}
