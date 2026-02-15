// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package testutil provides helpers for generating ephemeral TLS certificates
// used in mTLS tests. All certificates are created in memory using
// crypto/x509 and crypto/ecdsa from the standard library.
package testutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// CertFiles holds the PEM file paths for a CA, server, and client cert set.
type CertFiles struct {
	CACert     string
	ServerCert string
	ServerKey  string
	ClientCert string
	ClientKey  string
}

// GenerateCertFiles creates an ephemeral CA, server cert, and client cert
// written to PEM files inside dir. The files are automatically cleaned up
// when the test finishes (via t.TempDir or t.Cleanup).
func GenerateCertFiles(t *testing.T, dir string) CertFiles {
	t.Helper()

	// ── CA key and self-signed root cert ──────────────────────────────
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate CA key: %v", err)
	}

	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"DarkPipe Test CA"},
			CommonName:   "DarkPipe Test CA",
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create CA cert: %v", err)
	}

	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatalf("parse CA cert: %v", err)
	}

	caCertPath := writePEM(t, dir, "ca.crt", "CERTIFICATE", caDER)

	// ── Server cert signed by CA ─────────────────────────────────────
	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate server key: %v", err)
	}

	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
		NotBefore:   time.Now().Add(-1 * time.Hour),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	serverDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create server cert: %v", err)
	}

	serverCertPath := writePEM(t, dir, "server.crt", "CERTIFICATE", serverDER)
	serverKeyPath := writeKeyPEM(t, dir, "server.key", serverKey)

	// ── Client cert signed by CA ─────────────────────────────────────
	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate client key: %v", err)
	}

	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			CommonName: "darkpipe-client",
		},
		NotBefore:   time.Now().Add(-1 * time.Hour),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	clientDER, err := x509.CreateCertificate(rand.Reader, clientTemplate, caCert, &clientKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create client cert: %v", err)
	}

	clientCertPath := writePEM(t, dir, "client.crt", "CERTIFICATE", clientDER)
	clientKeyPath := writeKeyPEM(t, dir, "client.key", clientKey)

	return CertFiles{
		CACert:     caCertPath,
		ServerCert: serverCertPath,
		ServerKey:  serverKeyPath,
		ClientCert: clientCertPath,
		ClientKey:  clientKeyPath,
	}
}

// writePEM encodes raw DER bytes as a PEM file.
func writePEM(t *testing.T, dir, name, blockType string, der []byte) string {
	t.Helper()

	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", name, err)
	}
	defer f.Close()

	if err := pem.Encode(f, &pem.Block{Type: blockType, Bytes: der}); err != nil {
		t.Fatalf("encode PEM %s: %v", name, err)
	}

	return path
}

// writeKeyPEM marshals an ECDSA private key to a PEM file.
func writeKeyPEM(t *testing.T, dir, name string, key *ecdsa.PrivateKey) string {
	t.Helper()

	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}

	return writePEM(t, dir, name, "EC PRIVATE KEY", der)
}
