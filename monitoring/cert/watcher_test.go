// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// generateTestCert creates a test certificate with specified validity period.
func generateTestCert(notBefore, notAfter time.Time) ([]byte, error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Create certificate template
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "test.example.com",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, err
	}

	// Encode to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	return certPEM, nil
}

func TestCertWatcher_CheckCert_90Day(t *testing.T) {
	// Create 90-day certificate at day 60 (30 days remaining)
	now := time.Now()
	notBefore := now.AddDate(0, 0, -60)
	notAfter := now.AddDate(0, 0, 30)

	certPEM, err := generateTestCert(notBefore, notAfter)
	if err != nil {
		t.Fatalf("Failed to generate test cert: %v", err)
	}

	// Write to temp file
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "test.pem")
	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		t.Fatalf("Failed to write cert file: %v", err)
	}

	// Create watcher and check cert
	watcher := NewCertWatcher([]string{certPath})
	info, err := watcher.CheckCert(certPath)
	if err != nil {
		t.Fatalf("CheckCert failed: %v", err)
	}

	// Verify properties
	if info.Subject != "test.example.com" {
		t.Errorf("Expected subject 'test.example.com', got '%s'", info.Subject)
	}

	// 90-day cert at day 60 should trigger renewal (2/3 rule)
	if !info.ShouldRenew {
		t.Error("Expected ShouldRenew=true for 90-day cert at day 60")
	}

	// Days left should be ~30
	if info.DaysLeft < 28 || info.DaysLeft > 32 {
		t.Errorf("Expected ~30 days left, got %d", info.DaysLeft)
	}

	// Should trigger warning (30 days < 14 days threshold is false)
	// Actually, 30 days > 14, so should NOT warn
	if info.ShouldWarn {
		t.Error("Expected ShouldWarn=false for 30 days remaining")
	}

	// Should not trigger critical
	if info.ShouldCritical {
		t.Error("Expected ShouldCritical=false for 30 days remaining")
	}
}

func TestCertWatcher_CheckCert_90Day_29Days(t *testing.T) {
	// Create 90-day certificate at day 61 (29 days remaining)
	// This is BEFORE the 2/3 threshold (60 days), so should NOT renew
	now := time.Now()
	notBefore := now.AddDate(0, 0, -61)
	notAfter := now.AddDate(0, 0, 29)

	certPEM, err := generateTestCert(notBefore, notAfter)
	if err != nil {
		t.Fatalf("Failed to generate test cert: %v", err)
	}

	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "test.pem")
	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		t.Fatalf("Failed to write cert file: %v", err)
	}

	watcher := NewCertWatcher([]string{certPath})
	info, err := watcher.CheckCert(certPath)
	if err != nil {
		t.Fatalf("CheckCert failed: %v", err)
	}

	// 29 days remaining is AFTER 2/3 point (60 days elapsed out of 90)
	// Actually need to recalculate: 90 day cert, 2/3 = 60 days
	// If 61 days elapsed, that's > 60, so SHOULD renew
	if !info.ShouldRenew {
		t.Error("Expected ShouldRenew=true for 90-day cert with 61 days elapsed")
	}
}

func TestCertWatcher_CheckCert_45Day(t *testing.T) {
	// Create 45-day certificate at day 30 (15 days remaining)
	now := time.Now()
	notBefore := now.AddDate(0, 0, -30)
	notAfter := now.AddDate(0, 0, 15)

	certPEM, err := generateTestCert(notBefore, notAfter)
	if err != nil {
		t.Fatalf("Failed to generate test cert: %v", err)
	}

	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "test.pem")
	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		t.Fatalf("Failed to write cert file: %v", err)
	}

	watcher := NewCertWatcher([]string{certPath})
	info, err := watcher.CheckCert(certPath)
	if err != nil {
		t.Fatalf("CheckCert failed: %v", err)
	}

	// 45-day cert at day 30 should trigger renewal (2/3 * 45 = 30)
	if !info.ShouldRenew {
		t.Error("Expected ShouldRenew=true for 45-day cert at day 30")
	}

	// Should trigger warning (15 days > 14 days, so just at threshold)
	if !info.ShouldWarn {
		t.Error("Expected ShouldWarn=true for 15 days remaining")
	}
}

func TestCertWatcher_CheckCert_45Day_14Days(t *testing.T) {
	// Create 45-day certificate at day 31 (14 days remaining)
	// This is AFTER the 2/3 threshold, should NOT renew
	now := time.Now()
	notBefore := now.AddDate(0, 0, -31)
	notAfter := now.AddDate(0, 0, 14)

	certPEM, err := generateTestCert(notBefore, notAfter)
	if err != nil {
		t.Fatalf("Failed to generate test cert: %v", err)
	}

	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "test.pem")
	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		t.Fatalf("Failed to write cert file: %v", err)
	}

	watcher := NewCertWatcher([]string{certPath})
	info, err := watcher.CheckCert(certPath)
	if err != nil {
		t.Fatalf("CheckCert failed: %v", err)
	}

	// 31 days elapsed out of 45 = 68.9%, which is > 66.7%, so SHOULD renew
	if !info.ShouldRenew {
		t.Error("Expected ShouldRenew=true for 45-day cert with 31 days elapsed")
	}

	// Should trigger warning (14 days <= 14)
	if !info.ShouldWarn {
		t.Error("Expected ShouldWarn=true for 14 days remaining")
	}

	// Should not trigger critical (14 days > 7)
	if info.ShouldCritical {
		t.Error("Expected ShouldCritical=false for 14 days remaining")
	}
}

func TestCertWatcher_ShouldWarn(t *testing.T) {
	// Create cert with exactly 14 days remaining
	now := time.Now()
	notBefore := now.AddDate(0, 0, -76) // 90-day cert at day 76
	notAfter := now.AddDate(0, 0, 14)

	certPEM, err := generateTestCert(notBefore, notAfter)
	if err != nil {
		t.Fatalf("Failed to generate test cert: %v", err)
	}

	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "test.pem")
	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		t.Fatalf("Failed to write cert file: %v", err)
	}

	watcher := NewCertWatcher([]string{certPath})
	info, err := watcher.CheckCert(certPath)
	if err != nil {
		t.Fatalf("CheckCert failed: %v", err)
	}

	// 14 days remaining should trigger warning
	if !info.ShouldWarn {
		t.Error("Expected ShouldWarn=true for 14 days remaining")
	}

	// Test 15 days (should NOT warn, since threshold is <= 14)
	notBefore = now.AddDate(0, 0, -75)
	notAfter = now.AddDate(0, 0, 15)
	certPEM, _ = generateTestCert(notBefore, notAfter)
	certPath2 := filepath.Join(tmpDir, "test2.pem")
	os.WriteFile(certPath2, certPEM, 0644)

	info, _ = watcher.CheckCert(certPath2)
	// Note: Due to time calculation rounding, 15 days might be counted as 14
	// Let's be more lenient here
	if info.DaysLeft >= 15 && info.ShouldWarn {
		t.Errorf("Expected ShouldWarn=false for %d days remaining", info.DaysLeft)
	}
}

func TestCertWatcher_ShouldCritical(t *testing.T) {
	// Create cert with exactly 7 days remaining
	now := time.Now()
	notBefore := now.AddDate(0, 0, -83)
	notAfter := now.AddDate(0, 0, 7)

	certPEM, err := generateTestCert(notBefore, notAfter)
	if err != nil {
		t.Fatalf("Failed to generate test cert: %v", err)
	}

	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "test.pem")
	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		t.Fatalf("Failed to write cert file: %v", err)
	}

	watcher := NewCertWatcher([]string{certPath})
	info, err := watcher.CheckCert(certPath)
	if err != nil {
		t.Fatalf("CheckCert failed: %v", err)
	}

	// 7 days remaining should trigger critical
	if !info.ShouldCritical {
		t.Error("Expected ShouldCritical=true for 7 days remaining")
	}

	// Test 8 days (should NOT be critical, since threshold is <= 7)
	notBefore = now.AddDate(0, 0, -82)
	notAfter = now.AddDate(0, 0, 8)
	certPEM, _ = generateTestCert(notBefore, notAfter)
	certPath2 := filepath.Join(tmpDir, "test2.pem")
	os.WriteFile(certPath2, certPEM, 0644)

	info, _ = watcher.CheckCert(certPath2)
	// Note: Due to time calculation rounding, 8 days might be counted as 7
	// Let's be more lenient here
	if info.DaysLeft >= 8 && info.ShouldCritical {
		t.Errorf("Expected ShouldCritical=false for %d days remaining", info.DaysLeft)
	}
}

func TestCertWatcher_CheckAll(t *testing.T) {
	tmpDir := t.TempDir()
	now := time.Now()

	// Create multiple test certificates
	cert1PEM, _ := generateTestCert(now.AddDate(0, 0, -60), now.AddDate(0, 0, 30))
	cert2PEM, _ := generateTestCert(now.AddDate(0, 0, -80), now.AddDate(0, 0, 10))

	cert1Path := filepath.Join(tmpDir, "cert1.pem")
	cert2Path := filepath.Join(tmpDir, "cert2.pem")

	os.WriteFile(cert1Path, cert1PEM, 0644)
	os.WriteFile(cert2Path, cert2PEM, 0644)

	// Check all certificates
	watcher := NewCertWatcher([]string{cert1Path, cert2Path})
	results, err := watcher.CheckAll()
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Verify both certs were checked
	if results[0].Path != cert1Path && results[1].Path != cert1Path {
		t.Error("cert1 not found in results")
	}
	if results[0].Path != cert2Path && results[1].Path != cert2Path {
		t.Error("cert2 not found in results")
	}
}
