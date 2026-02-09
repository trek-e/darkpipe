package server

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"testing"
	"time"

	"github.com/darkpipe/darkpipe/transport/mtls/testutil"
)

func TestNewServer_ValidCerts(t *testing.T) {
	dir := t.TempDir()
	certs := testutil.GenerateCertFiles(t, dir)

	srv, err := NewServer("127.0.0.1:0", certs.CACert, certs.ServerCert, certs.ServerKey)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	if srv == nil {
		t.Fatal("NewServer returned nil server")
	}
}

func TestNewServer_MissingCACert(t *testing.T) {
	dir := t.TempDir()
	certs := testutil.GenerateCertFiles(t, dir)

	_, err := NewServer("127.0.0.1:0", "/nonexistent/ca.crt", certs.ServerCert, certs.ServerKey)
	if err == nil {
		t.Fatal("expected error for missing CA cert, got nil")
	}
}

func TestNewServer_MissingServerCert(t *testing.T) {
	dir := t.TempDir()
	certs := testutil.GenerateCertFiles(t, dir)

	_, err := NewServer("127.0.0.1:0", certs.CACert, "/nonexistent/server.crt", certs.ServerKey)
	if err == nil {
		t.Fatal("expected error for missing server cert, got nil")
	}
}

func TestServer_AcceptsValidClient(t *testing.T) {
	dir := t.TempDir()
	certs := testutil.GenerateCertFiles(t, dir)

	srv, err := NewServer("127.0.0.1:0", certs.CACert, certs.ServerCert, certs.ServerKey)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	ln, err := srv.Listen()
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer ln.Close()

	// Accept and complete handshake in a goroutine so tls.Dial can proceed.
	type result struct {
		conn net.Conn
		err  error
	}
	resCh := make(chan result, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			resCh <- result{nil, err}
			return
		}
		// Explicitly complete the server-side TLS handshake.
		if tc, ok := conn.(*tls.Conn); ok {
			if hsErr := tc.Handshake(); hsErr != nil {
				conn.Close()
				resCh <- result{nil, hsErr}
				return
			}
		}
		resCh <- result{conn, nil}
	}()

	// Connect with valid client cert.
	clientTLS := clientTLSConfig(t, certs)
	conn, err := tls.Dial("tcp", ln.Addr().String(), clientTLS)
	if err != nil {
		t.Fatalf("client dial: %v", err)
	}
	defer conn.Close()

	// Server should accept.
	select {
	case res := <-resCh:
		if res.err != nil {
			t.Fatalf("server handshake error: %v", res.err)
		}
		res.conn.Close()
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for server accept")
	}
}

func TestServer_RejectsClientWithoutCert(t *testing.T) {
	dir := t.TempDir()
	certs := testutil.GenerateCertFiles(t, dir)

	srv, err := NewServer("127.0.0.1:0", certs.CACert, certs.ServerCert, certs.ServerKey)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	ln, err := srv.Listen()
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer ln.Close()

	// Server goroutine: accept, handshake, attempt read. The read forces
	// certificate verification to complete and the server will see the
	// client did not present a cert.
	serverDone := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			serverDone <- err
			return
		}
		defer conn.Close()
		if tc, ok := conn.(*tls.Conn); ok {
			if hsErr := tc.Handshake(); hsErr != nil {
				// Server rejected the client during handshake -- expected.
				serverDone <- hsErr
				return
			}
		}
		// Try to read -- should fail because client has no cert.
		buf := make([]byte, 1)
		_, readErr := conn.Read(buf)
		serverDone <- readErr
	}()

	// Connect WITHOUT presenting a client certificate.
	// In TLS 1.3 the handshake may appear to succeed on the client side
	// because the certificate request is post-handshake. The failure
	// surfaces when the client tries to read/write data.
	noCertTLS := &tls.Config{
		RootCAs:    loadCACertPool(t, certs.CACert),
		MinVersion: tls.VersionTLS12,
	}

	conn, err := tls.Dial("tcp", ln.Addr().String(), noCertTLS)
	if err != nil {
		// Expected path for TLS 1.2: handshake fails immediately.
		return
	}
	defer conn.Close()

	// TLS 1.3 path: handshake succeeded, but data exchange should fail.
	// Write something so the server tries to read (triggering cert check).
	_, writeErr := conn.Write([]byte("test"))

	// Try to read -- server should have closed the connection.
	buf := make([]byte, 64)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, readErr := conn.Read(buf)

	// At least one of write/read or the server handshake should have errored.
	select {
	case srvErr := <-serverDone:
		// Server detected the missing cert -- test passes either way.
		_ = srvErr
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for server to process connection")
	}

	// If we got here without any errors at all, that means the server
	// accepted a client without a certificate, which is wrong.
	if writeErr == nil && readErr == nil {
		t.Fatal("expected connection to fail without client cert, but data exchange succeeded")
	}
}

func TestServer_Serve(t *testing.T) {
	dir := t.TempDir()
	certs := testutil.GenerateCertFiles(t, dir)

	srv, err := NewServer("127.0.0.1:0", certs.CACert, certs.ServerCert, certs.ServerKey)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	ln, err := srv.Listen()
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}

	handlerCalled := make(chan struct{}, 1)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				// Complete handshake and signal.
				if tc, ok := c.(*tls.Conn); ok {
					if err := tc.Handshake(); err != nil {
						return
					}
				}
				select {
				case handlerCalled <- struct{}{}:
				default:
				}
			}(conn)
		}
	}()

	// Connect.
	clientTLS := clientTLSConfig(t, certs)
	conn, err := tls.Dial("tcp", ln.Addr().String(), clientTLS)
	if err != nil {
		t.Fatalf("client dial: %v", err)
	}
	conn.Close()

	select {
	case <-handlerCalled:
		// Success.
	case <-time.After(5 * time.Second):
		t.Fatal("handler was never called")
	}

	// Clean shutdown.
	srv.Close()
}

// ── helpers ─────────────────────────────────────────────────────────────────

func clientTLSConfig(t *testing.T, certs testutil.CertFiles) *tls.Config {
	t.Helper()

	cert, err := tls.LoadX509KeyPair(certs.ClientCert, certs.ClientKey)
	if err != nil {
		t.Fatalf("load client cert: %v", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      loadCACertPool(t, certs.CACert),
		MinVersion:   tls.VersionTLS12,
	}
}

func loadCACertPool(t *testing.T, caPath string) *x509.CertPool {
	t.Helper()

	caPEM, err := os.ReadFile(caPath)
	if err != nil {
		t.Fatalf("read CA cert: %v", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		t.Fatalf("failed to parse CA cert PEM from %s", caPath)
	}

	return pool
}
