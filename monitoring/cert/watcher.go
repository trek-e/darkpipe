// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package cert provides certificate lifecycle management including monitoring, renewal, and rotation.
package cert

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

// CertInfo contains information about a certificate and its lifecycle status.
type CertInfo struct {
	Path          string        // Path to certificate file
	Subject       string        // Certificate subject (CN)
	NotBefore     time.Time     // Certificate valid from
	NotAfter      time.Time     // Certificate valid until
	DaysLeft      int           // Days until expiry
	Lifetime      time.Duration // Total certificate lifetime
	ShouldRenew   bool          // True if cert should be renewed (2/3 of lifetime elapsed)
	ShouldWarn    bool          // True if warn alert should fire
	ShouldCritical bool         // True if critical alert should fire
}

// CertWatcher monitors certificates for expiry and renewal timing.
type CertWatcher struct {
	certPaths       []string // Paths to certificate files to monitor
	warnDays        int      // Days before expiry to warn (default: 14)
	criticalDays    int      // Days before expiry to alert critical (default: 7)
	renewalFraction float64  // Fraction of lifetime before renewal (default: 2/3)
}

// NewCertWatcher creates a certificate watcher with default thresholds.
// Defaults: warnDays=14, criticalDays=7, renewalFraction=2/3
func NewCertWatcher(paths []string) *CertWatcher {
	return &CertWatcher{
		certPaths:       paths,
		warnDays:        14,
		criticalDays:    7,
		renewalFraction: 2.0 / 3.0,
	}
}

// WithThresholds sets custom alert thresholds.
func (w *CertWatcher) WithThresholds(warnDays, criticalDays int) *CertWatcher {
	w.warnDays = warnDays
	w.criticalDays = criticalDays
	return w
}

// WithRenewalFraction sets custom renewal fraction (default: 2/3).
func (w *CertWatcher) WithRenewalFraction(fraction float64) *CertWatcher {
	w.renewalFraction = fraction
	return w
}

// CheckCert parses a certificate file and returns lifecycle information.
func (w *CertWatcher) CheckCert(certPath string) (*CertInfo, error) {
	// Read certificate file
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	// Parse PEM block
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Calculate lifecycle metrics
	now := time.Now()
	lifetime := cert.NotAfter.Sub(cert.NotBefore)
	elapsed := now.Sub(cert.NotBefore)
	remaining := cert.NotAfter.Sub(now)
	daysLeft := int(remaining.Hours() / 24)

	// Determine renewal and alert status
	shouldRenew := elapsed > time.Duration(float64(lifetime)*w.renewalFraction)
	shouldWarn := daysLeft <= w.warnDays
	shouldCritical := daysLeft <= w.criticalDays

	return &CertInfo{
		Path:           certPath,
		Subject:        cert.Subject.CommonName,
		NotBefore:      cert.NotBefore,
		NotAfter:       cert.NotAfter,
		DaysLeft:       daysLeft,
		Lifetime:       lifetime,
		ShouldRenew:    shouldRenew,
		ShouldWarn:     shouldWarn,
		ShouldCritical: shouldCritical,
	}, nil
}

// CheckAll checks all registered certificate paths.
func (w *CertWatcher) CheckAll() ([]CertInfo, error) {
	results := make([]CertInfo, 0, len(w.certPaths))

	for _, path := range w.certPaths {
		info, err := w.CheckCert(path)
		if err != nil {
			// Log error but continue checking other certificates
			continue
		}
		results = append(results, *info)
	}

	return results, nil
}

// ShouldRenew returns true if the certificate should be renewed based on the 2/3 lifetime rule.
// For a 90-day cert, renewal triggers at 60 days elapsed (30 days remaining).
// For a 45-day cert, renewal triggers at 30 days elapsed (15 days remaining).
func ShouldRenew(cert *CertInfo) bool {
	return cert.ShouldRenew
}
