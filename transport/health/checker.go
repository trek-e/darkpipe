// Package health provides a unified transport health checker that reports
// the status of whichever transport is active (WireGuard or mTLS). This
// gives Phase 9 (monitoring) a single interface to query regardless of
// which transport the user selected.
package health

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/darkpipe/darkpipe/transport/wireguard/monitor"
)

// TransportType identifies which transport is being health-checked.
type TransportType string

const (
	// WireGuard transport using kernel/userspace WireGuard tunnel.
	WireGuard TransportType = "wireguard"
	// MTLS transport using mutual TLS over persistent TCP connection.
	MTLS TransportType = "mtls"
)

// HealthStatus holds the result of a single health check.
type HealthStatus struct {
	Transport TransportType
	Healthy   bool
	LastCheck time.Time
	Error     string
	Details   map[string]string
}

// WireGuardConfig holds the parameters for checking WireGuard health.
type WireGuardConfig struct {
	DeviceName      string        // e.g. "wg0"
	MaxHandshakeAge time.Duration // Default: 5 minutes
}

// MTLSConfig holds the parameters for checking mTLS health.
type MTLSConfig struct {
	ServerAddr string // host:port of the mTLS server
	CACertPath string // PEM CA cert for server verification
	CertPath   string // Client cert PEM path
	KeyPath    string // Client key PEM path
}

// Check performs a health check for the specified transport type.
//
// For WireGuard, config should be a *WireGuardConfig.
// For mTLS, config should be an *MTLSConfig.
//
// Returns a HealthStatus with details about the check result.
func Check(transport TransportType, config interface{}) HealthStatus {
	now := time.Now()
	switch transport {
	case WireGuard:
		return checkWireGuard(config, now)
	case MTLS:
		return checkMTLS(config, now)
	default:
		return HealthStatus{
			Transport: transport,
			Healthy:   false,
			LastCheck: now,
			Error:     fmt.Sprintf("unknown transport type: %s", transport),
			Details:   map[string]string{},
		}
	}
}

func checkWireGuard(config interface{}, now time.Time) HealthStatus {
	status := HealthStatus{
		Transport: WireGuard,
		LastCheck: now,
		Details:   make(map[string]string),
	}

	wgCfg, ok := config.(*WireGuardConfig)
	if !ok || wgCfg == nil {
		status.Error = "invalid config: expected *WireGuardConfig"
		return status
	}

	if wgCfg.DeviceName == "" {
		wgCfg.DeviceName = "wg0"
	}
	if wgCfg.MaxHandshakeAge <= 0 {
		wgCfg.MaxHandshakeAge = monitor.DefaultMaxHandshakeAge
	}

	status.Details["device"] = wgCfg.DeviceName
	status.Details["max_handshake_age"] = wgCfg.MaxHandshakeAge.String()

	// Use GetTunnelHealth for structured data.
	health := monitor.GetTunnelHealth(wgCfg.DeviceName, wgCfg.MaxHandshakeAge)
	status.Healthy = health.Healthy
	status.Error = health.Error

	status.Details["peer_count"] = fmt.Sprintf("%d", len(health.Peers))
	for i, peer := range health.Peers {
		prefix := fmt.Sprintf("peer_%d", i)
		status.Details[prefix+"_pubkey"] = peer.PublicKey[:16] + "..."
		status.Details[prefix+"_healthy"] = fmt.Sprintf("%t", peer.Healthy)
		if peer.HandshakeAge > 0 {
			status.Details[prefix+"_handshake_age"] = peer.HandshakeAge.Round(time.Second).String()
		}
	}

	return status
}

func checkMTLS(config interface{}, now time.Time) HealthStatus {
	status := HealthStatus{
		Transport: MTLS,
		LastCheck: now,
		Details:   make(map[string]string),
	}

	mtlsCfg, ok := config.(*MTLSConfig)
	if !ok || mtlsCfg == nil {
		status.Error = "invalid config: expected *MTLSConfig"
		return status
	}

	if mtlsCfg.ServerAddr == "" {
		status.Error = "server address is required"
		return status
	}

	status.Details["server"] = mtlsCfg.ServerAddr

	// Check local certificate validity first (no network needed).
	if mtlsCfg.CertPath != "" {
		certExpiry, err := checkCertExpiry(mtlsCfg.CertPath)
		if err != nil {
			status.Error = fmt.Sprintf("check client cert: %v", err)
			return status
		}
		status.Details["cert_expiry"] = certExpiry.Format(time.RFC3339)
		status.Details["cert_remaining"] = time.Until(certExpiry).Round(time.Minute).String()

		if time.Until(certExpiry) <= 0 {
			status.Error = "client certificate has expired"
			return status
		}
	}

	// Attempt TLS connection to verify server is reachable and certs are valid.
	tlsConfig, err := buildTLSConfig(mtlsCfg)
	if err != nil {
		status.Error = fmt.Sprintf("build TLS config: %v", err)
		return status
	}

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	start := time.Now()
	conn, err := tls.DialWithDialer(
		dialer,
		"tcp",
		mtlsCfg.ServerAddr,
		tlsConfig,
	)
	if err != nil {
		status.Error = fmt.Sprintf("TLS handshake failed: %v", err)
		return status
	}
	defer conn.Close()

	latency := time.Since(start)
	status.Details["connection_latency"] = latency.Round(time.Millisecond).String()

	// Extract server certificate info.
	state := conn.ConnectionState()
	if len(state.PeerCertificates) > 0 {
		serverCert := state.PeerCertificates[0]
		status.Details["server_cert_cn"] = serverCert.Subject.CommonName
		status.Details["server_cert_expiry"] = serverCert.NotAfter.Format(time.RFC3339)
	}
	status.Details["tls_version"] = tlsVersionString(state.Version)
	status.Details["negotiated_protocol"] = state.NegotiatedProtocol

	status.Healthy = true
	return status
}

// checkCertExpiry reads a PEM certificate file and returns the expiry time.
func checkCertExpiry(certPath string) (time.Time, error) {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return time.Time{}, fmt.Errorf("read %s: %w", certPath, err)
	}

	// Parse the first certificate in the PEM chain.
	rest := data
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err == nil {
				return cert.NotAfter, nil
			}
		}
	}

	return time.Time{}, fmt.Errorf("no certificate found in %s", certPath)
}

func buildTLSConfig(cfg *MTLSConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Load CA cert if provided.
	if cfg.CACertPath != "" {
		caCert, err := os.ReadFile(cfg.CACertPath)
		if err != nil {
			return nil, fmt.Errorf("read CA cert: %w", err)
		}

		rootCAs := x509.NewCertPool()
		if !rootCAs.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("parse CA cert from %s", cfg.CACertPath)
		}
		tlsConfig.RootCAs = rootCAs
	}

	// Load client cert/key if provided.
	if cfg.CertPath != "" && cfg.KeyPath != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("load client cert/key: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

func tlsVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("unknown (0x%04x)", version)
	}
}
