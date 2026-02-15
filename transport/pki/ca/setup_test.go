// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package ca

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// generateTestCert creates a self-signed PEM certificate in dir and returns
// the file path. The cert expires at the given notAfter time.
func generateTestCert(t *testing.T, dir string, notAfter time.Time) string {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test-cert",
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  notAfter,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	certPath := filepath.Join(dir, "test.crt")
	f, err := os.Create(certPath)
	if err != nil {
		t.Fatalf("create cert file: %v", err)
	}
	defer f.Close()

	if err := pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		t.Fatalf("encode PEM: %v", err)
	}

	return certPath
}

func TestCheckCertExpiry(t *testing.T) {
	dir := t.TempDir()
	expectedExpiry := time.Now().Add(24 * time.Hour).Truncate(time.Second)

	certPath := generateTestCert(t, dir, expectedExpiry)

	got, err := CheckCertExpiry(certPath)
	if err != nil {
		t.Fatalf("CheckCertExpiry: %v", err)
	}

	// x509 certificates store time at second granularity.
	gotTrunc := got.Truncate(time.Second)
	if !gotTrunc.Equal(expectedExpiry) {
		t.Errorf("expiry = %v, want %v", gotTrunc, expectedExpiry)
	}
}

func TestCheckCertExpiry_MissingFile(t *testing.T) {
	_, err := CheckCertExpiry("/nonexistent/path/cert.pem")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestCheckCertExpiry_InvalidPEM(t *testing.T) {
	dir := t.TempDir()
	badPath := filepath.Join(dir, "bad.crt")
	if err := os.WriteFile(badPath, []byte("not a PEM"), 0600); err != nil {
		t.Fatalf("write bad file: %v", err)
	}

	_, err := CheckCertExpiry(badPath)
	if err == nil {
		t.Fatal("expected error for invalid PEM, got nil")
	}
}

func TestInitCA_MissingStepBinary(t *testing.T) {
	// Override PATH so that step CLI cannot be found.
	t.Setenv("PATH", t.TempDir())

	err := InitCA(CAConfig{})
	if err == nil {
		t.Fatal("expected error when step binary is missing, got nil")
	}
}

func TestCAConfigDefaults(t *testing.T) {
	cfg := CAConfig{}
	cfg.applyDefaults()

	if cfg.Name != "DarkPipe Internal CA" {
		t.Errorf("Name = %q, want %q", cfg.Name, "DarkPipe Internal CA")
	}
	if cfg.DNS != "ca.darkpipe.internal" {
		t.Errorf("DNS = %q, want %q", cfg.DNS, "ca.darkpipe.internal")
	}
	if cfg.Address != ":8443" {
		t.Errorf("Address = %q, want %q", cfg.Address, ":8443")
	}
	if cfg.Provisioner != "darkpipe-acme" {
		t.Errorf("Provisioner = %q, want %q", cfg.Provisioner, "darkpipe-acme")
	}
	if cfg.CertDir != "/etc/darkpipe/certs" {
		t.Errorf("CertDir = %q, want %q", cfg.CertDir, "/etc/darkpipe/certs")
	}
}

func TestCAConfigPreserveCustomValues(t *testing.T) {
	cfg := CAConfig{
		Name:        "Custom CA",
		DNS:         "custom.dns",
		Address:     ":9999",
		Provisioner: "custom-prov",
		CertDir:     "/custom/certs",
	}
	cfg.applyDefaults()

	if cfg.Name != "Custom CA" {
		t.Errorf("Name overwritten: got %q", cfg.Name)
	}
	if cfg.DNS != "custom.dns" {
		t.Errorf("DNS overwritten: got %q", cfg.DNS)
	}
	if cfg.Address != ":9999" {
		t.Errorf("Address overwritten: got %q", cfg.Address)
	}
	if cfg.Provisioner != "custom-prov" {
		t.Errorf("Provisioner overwritten: got %q", cfg.Provisioner)
	}
	if cfg.CertDir != "/custom/certs" {
		t.Errorf("CertDir overwritten: got %q", cfg.CertDir)
	}
}
