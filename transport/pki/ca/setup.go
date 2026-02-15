// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package ca provides helpers for initializing a step-ca private certificate
// authority and issuing/inspecting certificates. All certificate operations
// shell out to the step CLI rather than re-implementing crypto -- the official
// tools are maintained by Smallstep and handle edge cases (key storage, ACME
// provisioners, audit logging) that a hand-rolled solution would miss.
package ca

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// CAConfig holds the parameters needed to initialise a step-ca instance.
// Zero-value fields are replaced with sensible defaults before use.
type CAConfig struct {
	Name        string // CA display name (default "DarkPipe Internal CA")
	DNS         string // CA DNS name    (default "ca.darkpipe.internal")
	Address     string // Listen address (default ":8443")
	Provisioner string // Provisioner    (default "darkpipe-acme")
	CertDir     string // Cert output dir(default "/etc/darkpipe/certs")
}

// applyDefaults fills zero-value fields with the documented defaults.
func (c *CAConfig) applyDefaults() {
	if c.Name == "" {
		c.Name = "DarkPipe Internal CA"
	}
	if c.DNS == "" {
		c.DNS = "ca.darkpipe.internal"
	}
	if c.Address == "" {
		c.Address = ":8443"
	}
	if c.Provisioner == "" {
		c.Provisioner = "darkpipe-acme"
	}
	if c.CertDir == "" {
		c.CertDir = "/etc/darkpipe/certs"
	}
}

// InitCA initialises a new step-ca certificate authority by shelling out to
// `step ca init`. This creates the CA database, root certificate, and
// intermediate certificate on disk.
//
// Returns a clear error if the step CLI binary cannot be found.
func InitCA(cfg CAConfig) error {
	cfg.applyDefaults()

	stepBin, err := exec.LookPath("step")
	if err != nil {
		return fmt.Errorf("step CLI not found in PATH: install step-ca (https://smallstep.com/docs/step-ca/installation/): %w", err)
	}

	var stderr bytes.Buffer
	cmd := exec.Command(stepBin, "ca", "init",
		"--name", cfg.Name,
		"--dns", cfg.DNS,
		"--address", cfg.Address,
		"--provisioner", cfg.Provisioner,
		"--provisioner-type", "ACME",
	)
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("step ca init failed: %w\nstderr: %s", err, stderr.String())
	}

	return nil
}

// IssueCertificate requests a new certificate from a running step-ca instance.
// It shells out to `step ca certificate` which connects to caURL, authenticates
// via the configured provisioner, and writes the signed cert and private key to
// certPath and keyPath respectively.
func IssueCertificate(caURL, commonName, certPath, keyPath string) error {
	stepBin, err := exec.LookPath("step")
	if err != nil {
		return fmt.Errorf("step CLI not found in PATH: install step-ca (https://smallstep.com/docs/step-ca/installation/): %w", err)
	}

	var stderr bytes.Buffer
	cmd := exec.Command(stepBin, "ca", "certificate",
		commonName,
		certPath,
		keyPath,
		"--ca-url", caURL,
		"--force",
	)
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("step ca certificate failed for %q: %w\nstderr: %s", commonName, err, stderr.String())
	}

	return nil
}

// CheckCertExpiry reads a PEM-encoded certificate from certPath and returns
// the NotAfter time (expiry). Uses crypto/x509 from the standard library --
// no external dependencies.
func CheckCertExpiry(certPath string) (time.Time, error) {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return time.Time{}, fmt.Errorf("read certificate %s: %w", certPath, err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return time.Time{}, fmt.Errorf("no PEM block found in %s", certPath)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse certificate %s: %w", certPath, err)
	}

	return cert.NotAfter, nil
}
