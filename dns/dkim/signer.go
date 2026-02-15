// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package dkim

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"io"

	"github.com/emersion/go-msgauth/dkim"
)

// Signer wraps emersion/go-msgauth for DKIM message signing.
type Signer struct {
	domain     string
	selector   string
	privateKey *rsa.PrivateKey
}

// NewSigner creates a new DKIM signer.
func NewSigner(domain, selector string, privateKey *rsa.PrivateKey) *Signer {
	return &Signer{
		domain:     domain,
		selector:   selector,
		privateKey: privateKey,
	}
}

// Sign signs an email message with DKIM-Signature header.
// Uses relaxed/relaxed canonicalization per RFC 6376 best practices.
// Signs headers: From, To, Subject, Date, Message-ID, MIME-Version, Content-Type.
func (s *Signer) Sign(r io.Reader) ([]byte, error) {
	// Read the message into memory
	message, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	options := &dkim.SignOptions{
		Domain:   s.domain,
		Selector: s.selector,
		Signer:   s.privateKey,
		// Use relaxed/relaxed canonicalization (default, most compatible)
		HeaderCanonicalization: dkim.CanonicalizationRelaxed,
		BodyCanonicalization:   dkim.CanonicalizationRelaxed,
		// Sign important headers
		HeaderKeys: []string{
			"From",
			"To",
			"Subject",
			"Date",
			"Message-ID",
			"MIME-Version",
			"Content-Type",
		},
	}

	var signed bytes.Buffer
	if err := dkim.Sign(&signed, bytes.NewReader(message), options); err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}

	return signed.Bytes(), nil
}
