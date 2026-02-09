package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/darkpipe/darkpipe/transport/mtls/testutil"
)

// startTestServer starts an mTLS server on a random port and returns its
// address. The server calls handler for every accepted connection.
func startTestServer(t *testing.T, certs testutil.CertFiles, handler func(net.Conn)) net.Addr {
	t.Helper()

	caPEM, err := os.ReadFile(certs.CACert)
	if err != nil {
		t.Fatalf("read CA cert: %v", err)
	}
	clientCAs := x509.NewCertPool()
	clientCAs.AppendCertsFromPEM(caPEM)

	serverCert, err := tls.LoadX509KeyPair(certs.ServerCert, certs.ServerKey)
	if err != nil {
		t.Fatalf("load server cert: %v", err)
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientCAs,
		MinVersion:   tls.VersionTLS12,
	}

	ln, err := tls.Listen("tcp", "127.0.0.1:0", tlsCfg)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	t.Cleanup(func() { ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handler(conn)
		}
	}()

	return ln.Addr()
}

func TestNewClient_ValidCerts(t *testing.T) {
	dir := t.TempDir()
	certs := testutil.GenerateCertFiles(t, dir)

	c, err := NewClient("127.0.0.1:0", certs.CACert, certs.ClientCert, certs.ClientKey)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestNewClient_MissingCACert(t *testing.T) {
	dir := t.TempDir()
	certs := testutil.GenerateCertFiles(t, dir)

	_, err := NewClient("127.0.0.1:0", "/nonexistent/ca.crt", certs.ClientCert, certs.ClientKey)
	if err == nil {
		t.Fatal("expected error for missing CA cert, got nil")
	}
}

func TestNewClient_MissingClientCert(t *testing.T) {
	dir := t.TempDir()
	certs := testutil.GenerateCertFiles(t, dir)

	_, err := NewClient("127.0.0.1:0", certs.CACert, "/nonexistent/client.crt", certs.ClientKey)
	if err == nil {
		t.Fatal("expected error for missing client cert, got nil")
	}
}

func TestConnect_Success(t *testing.T) {
	dir := t.TempDir()
	certs := testutil.GenerateCertFiles(t, dir)

	addr := startTestServer(t, certs, func(conn net.Conn) {
		defer conn.Close()
		// Echo server: just hold the connection briefly.
		buf := make([]byte, 64)
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		conn.Read(buf)
	})

	c, err := NewClient(addr.String(), certs.CACert, certs.ClientCert, certs.ClientKey)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := c.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer conn.Close()

	// Verify connection state.
	state := conn.ConnectionState()
	if !state.HandshakeComplete {
		t.Error("expected handshake to be complete")
	}
}

func TestMaintainConnection_RetriesOnFailure(t *testing.T) {
	dir := t.TempDir()
	certs := testutil.GenerateCertFiles(t, dir)

	var attempts atomic.Int32

	addr := startTestServer(t, certs, func(conn net.Conn) {
		n := attempts.Add(1)
		if n < 3 {
			// First 2 connections: close immediately to trigger retry.
			conn.Close()
			return
		}
		// Third connection: hold it until client is done.
		defer conn.Close()
		buf := make([]byte, 64)
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		conn.Read(buf)
	})

	c, err := NewClient(addr.String(), certs.CACert, certs.ClientCert, certs.ClientKey)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	handlerCalled := make(chan struct{}, 1)

	go func() {
		_ = c.MaintainConnection(ctx, func(conn net.Conn) error {
			n := attempts.Load()
			if n < 3 {
				return errors.New("simulated handler failure")
			}
			// Signal success on the third connection.
			select {
			case handlerCalled <- struct{}{}:
			default:
			}
			// Hold connection until context cancelled.
			<-ctx.Done()
			return ctx.Err()
		})
	}()

	select {
	case <-handlerCalled:
		// MaintainConnection successfully retried and connected.
	case <-time.After(8 * time.Second):
		t.Fatal("timed out waiting for MaintainConnection to succeed after retries")
	}

	cancel()
}

func TestMaintainConnection_RespectsContextCancellation(t *testing.T) {
	dir := t.TempDir()
	certs := testutil.GenerateCertFiles(t, dir)

	// Server that just holds connections.
	addr := startTestServer(t, certs, func(conn net.Conn) {
		defer conn.Close()
		buf := make([]byte, 64)
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		conn.Read(buf)
	})

	c, err := NewClient(addr.String(), certs.CACert, certs.ClientCert, certs.ClientKey)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	doneCh := make(chan error, 1)
	go func() {
		doneCh <- c.MaintainConnection(ctx, func(conn net.Conn) error {
			// Hold connection until context is cancelled.
			<-ctx.Done()
			return ctx.Err()
		})
	}()

	// Give it a moment to establish the connection.
	time.Sleep(500 * time.Millisecond)

	cancel()

	select {
	case err := <-doneCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("MaintainConnection did not return after context cancellation")
	}
}
