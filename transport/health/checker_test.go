package health

import (
	"crypto/tls"
	"net"
	"os"
	"testing"
	"time"

	"github.com/darkpipe/darkpipe/transport/mtls/testutil"
)

func TestCheck_WireGuard_ReturnsCorrectTransportType(t *testing.T) {
	cfg := &WireGuardConfig{DeviceName: "wg-nonexistent"}
	status := Check(WireGuard, cfg)

	if status.Transport != WireGuard {
		t.Fatalf("expected transport %q, got %q", WireGuard, status.Transport)
	}
}

func TestCheck_WireGuard_UnhealthyIncludesError(t *testing.T) {
	// A non-existent device should yield an unhealthy status with error.
	cfg := &WireGuardConfig{DeviceName: "wg-nonexistent"}
	status := Check(WireGuard, cfg)

	if status.Healthy {
		t.Fatal("expected unhealthy status for non-existent device")
	}
	if status.Error == "" {
		t.Fatal("expected error message for non-existent device")
	}
	t.Logf("WireGuard error (expected): %s", status.Error)
}

func TestCheck_WireGuard_DetailsPopulated(t *testing.T) {
	cfg := &WireGuardConfig{DeviceName: "wg-nonexistent"}
	status := Check(WireGuard, cfg)

	if status.Details == nil {
		t.Fatal("expected Details map to be non-nil")
	}
	if status.Details["device"] != "wg-nonexistent" {
		t.Fatalf("expected device=wg-nonexistent, got %q", status.Details["device"])
	}
	if status.Details["max_handshake_age"] == "" {
		t.Fatal("expected max_handshake_age to be populated")
	}
}

func TestCheck_WireGuard_InvalidConfig(t *testing.T) {
	// Pass wrong config type.
	status := Check(WireGuard, "not-a-config")
	if status.Healthy {
		t.Fatal("expected unhealthy with invalid config")
	}
	if status.Error == "" {
		t.Fatal("expected error message for invalid config")
	}
}

func TestCheck_WireGuard_NilConfig(t *testing.T) {
	var cfg *WireGuardConfig
	status := Check(WireGuard, cfg)
	if status.Healthy {
		t.Fatal("expected unhealthy with nil config")
	}
}

func TestCheck_MTLS_ReturnsCorrectTransportType(t *testing.T) {
	cfg := &MTLSConfig{ServerAddr: "127.0.0.1:0"}
	status := Check(MTLS, cfg)

	if status.Transport != MTLS {
		t.Fatalf("expected transport %q, got %q", MTLS, status.Transport)
	}
}

func TestCheck_MTLS_InvalidConfig(t *testing.T) {
	status := Check(MTLS, "not-a-config")
	if status.Healthy {
		t.Fatal("expected unhealthy with invalid config")
	}
	if status.Error == "" {
		t.Fatal("expected error for invalid config type")
	}
}

func TestCheck_MTLS_EmptyServerAddr(t *testing.T) {
	cfg := &MTLSConfig{}
	status := Check(MTLS, cfg)
	if status.Healthy {
		t.Fatal("expected unhealthy with empty server address")
	}
	if status.Error == "" {
		t.Fatal("expected error for empty server address")
	}
}

func TestCheck_MTLS_UnreachableServer(t *testing.T) {
	// Use a port that nothing is listening on.
	cfg := &MTLSConfig{ServerAddr: "127.0.0.1:19999"}
	status := Check(MTLS, cfg)
	if status.Healthy {
		t.Fatal("expected unhealthy when server unreachable")
	}
	if status.Error == "" {
		t.Fatal("expected error for unreachable server")
	}
	if status.Details == nil {
		t.Fatal("expected Details map to be non-nil")
	}
	if status.Details["server"] != "127.0.0.1:19999" {
		t.Fatalf("expected server detail, got %q", status.Details["server"])
	}
}

func TestCheck_MTLS_HealthyConnection(t *testing.T) {
	// Start a real mTLS server and verify the checker succeeds.
	dir := t.TempDir()
	certs := testutil.GenerateCertFiles(t, dir)

	// Start TLS server.
	serverCert, err := tls.LoadX509KeyPair(certs.ServerCert, certs.ServerKey)
	if err != nil {
		t.Fatalf("load server cert: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert, // Simplified for health check test.
		MinVersion:   tls.VersionTLS12,
	}

	ln, err := tls.Listen("tcp", "127.0.0.1:0", tlsConfig)
	if err != nil {
		t.Fatalf("tls listen: %v", err)
	}
	defer ln.Close()

	// Accept connections in background.
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			// Perform handshake then close.
			if tc, ok := conn.(*tls.Conn); ok {
				tc.Handshake()
			}
			conn.Close()
		}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	cfg := &MTLSConfig{
		ServerAddr: addr.String(),
		CACertPath: certs.CACert,
		CertPath:   certs.ClientCert,
		KeyPath:    certs.ClientKey,
	}

	status := Check(MTLS, cfg)
	if !status.Healthy {
		t.Fatalf("expected healthy mTLS connection, got error: %s", status.Error)
	}
	if status.Details["connection_latency"] == "" {
		t.Fatal("expected connection_latency detail")
	}
	if status.Details["tls_version"] == "" {
		t.Fatal("expected tls_version detail")
	}
	if status.Details["server_cert_cn"] != "localhost" {
		t.Fatalf("expected server_cert_cn=localhost, got %q", status.Details["server_cert_cn"])
	}
	t.Logf("mTLS health check passed: latency=%s, tls=%s",
		status.Details["connection_latency"], status.Details["tls_version"])
}

func TestCheck_MTLS_CertExpiryCheck(t *testing.T) {
	dir := t.TempDir()
	certs := testutil.GenerateCertFiles(t, dir)

	// Verify cert expiry details are populated.
	cfg := &MTLSConfig{
		ServerAddr: "127.0.0.1:19999", // Unreachable, but cert check runs first.
		CertPath:   certs.ClientCert,
		KeyPath:    certs.ClientKey,
	}

	status := Check(MTLS, cfg)
	// Should fail on TLS connection (unreachable), but cert expiry should be populated.
	if status.Details["cert_expiry"] == "" {
		t.Fatal("expected cert_expiry detail to be populated")
	}
	if status.Details["cert_remaining"] == "" {
		t.Fatal("expected cert_remaining detail to be populated")
	}
	t.Logf("cert_expiry=%s, cert_remaining=%s",
		status.Details["cert_expiry"], status.Details["cert_remaining"])
}

func TestCheck_UnknownTransport(t *testing.T) {
	status := Check("unknown", nil)
	if status.Healthy {
		t.Fatal("expected unhealthy for unknown transport")
	}
	if status.Error == "" {
		t.Fatal("expected error for unknown transport")
	}
	if status.Transport != "unknown" {
		t.Fatalf("expected transport=unknown, got %q", status.Transport)
	}
}

func TestCheckCertExpiry_ValidCert(t *testing.T) {
	dir := t.TempDir()
	certs := testutil.GenerateCertFiles(t, dir)

	expiry, err := checkCertExpiry(certs.ClientCert)
	if err != nil {
		t.Fatalf("checkCertExpiry failed: %v", err)
	}

	// The test cert is valid for 24 hours.
	remaining := time.Until(expiry)
	if remaining <= 0 {
		t.Fatal("expected cert to not be expired")
	}
	if remaining > 25*time.Hour {
		t.Fatalf("unexpected remaining time: %s", remaining)
	}
	t.Logf("cert expires in %s", remaining.Round(time.Minute))
}

func TestCheckCertExpiry_MissingFile(t *testing.T) {
	_, err := checkCertExpiry("/nonexistent/cert.pem")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestCheckCertExpiry_InvalidPEM(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/bad.pem"

	if err := writeTestFile(path, "not a pem file"); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	_, err := checkCertExpiry(path)
	if err == nil {
		t.Fatal("expected error for invalid PEM")
	}
}

func TestTlsVersionString(t *testing.T) {
	tests := []struct {
		version uint16
		want    string
	}{
		{tls.VersionTLS10, "TLS 1.0"},
		{tls.VersionTLS11, "TLS 1.1"},
		{tls.VersionTLS12, "TLS 1.2"},
		{tls.VersionTLS13, "TLS 1.3"},
		{0x0999, "unknown (0x0999)"},
	}

	for _, tt := range tests {
		got := tlsVersionString(tt.version)
		if got != tt.want {
			t.Errorf("tlsVersionString(0x%04x) = %q, want %q", tt.version, got, tt.want)
		}
	}
}

func writeTestFile(path, content string) error {
	return writeFileHelper(path, []byte(content))
}

func writeFileHelper(path string, data []byte) error {
	return writeFileWithPerm(path, data, 0644)
}

func writeFileWithPerm(path string, data []byte, perm uint32) error {
	return os.WriteFile(path, data, os.FileMode(perm))
}
