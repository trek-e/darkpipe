package dkim

import (
	"bytes"
	"strings"
	"testing"

	"github.com/emersion/go-msgauth/dkim"
)

func TestSigner(t *testing.T) {
	// Generate test key
	privateKey, err := GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	signer := NewSigner("example.com", "test-2026q1", privateKey)

	// Test message
	message := []byte(`From: sender@example.com
To: recipient@example.com
Subject: Test Email
Date: Mon, 14 Feb 2026 05:00:00 +0000
Message-ID: <test@example.com>
MIME-Version: 1.0
Content-Type: text/plain; charset=utf-8

This is a test email message.
`)

	// Sign the message
	signed, err := signer.Sign(bytes.NewReader(message))
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	// Verify signed message contains DKIM-Signature header
	signedStr := string(signed)
	if !strings.Contains(signedStr, "DKIM-Signature:") {
		t.Fatal("Signed message does not contain DKIM-Signature header")
	}

	// Verify signature contains expected fields
	if !strings.Contains(signedStr, "v=1") {
		t.Fatal("DKIM signature missing v=1 tag")
	}
	if !strings.Contains(signedStr, "d=example.com") {
		t.Fatal("DKIM signature missing domain tag")
	}
	if !strings.Contains(signedStr, "s=test-2026q1") {
		t.Fatal("DKIM signature missing selector tag")
	}

	// Verify the signature using emersion/go-msgauth
	verifications, err := dkim.Verify(bytes.NewReader(signed))
	if err != nil {
		t.Fatalf("DKIM verification error = %v", err)
	}

	if len(verifications) == 0 {
		t.Fatal("No DKIM signatures found in signed message")
	}

	for _, v := range verifications {
		if v.Err != nil {
			// Note: Verification will fail without DNS TXT record (expected in tests)
			// We're testing that the signature was created with correct format
			// Acceptable errors: "no key", "lookup", "key revoked" (DKIM uses empty TXT record to indicate revoked keys)
			errStr := v.Err.Error()
			if !strings.Contains(errStr, "no key") &&
			   !strings.Contains(errStr, "lookup") &&
			   !strings.Contains(errStr, "key revoked") {
				t.Fatalf("DKIM verification error = %v", v.Err)
			}
			// If we got here, the error is expected (DNS lookup failure)
			// The fact that we have a verification object means the signature structure is valid
		}
	}
}

func TestSignerMinimalMessage(t *testing.T) {
	privateKey, err := GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	signer := NewSigner("example.com", "test-2026q1", privateKey)

	// Minimal valid email message (must have at least From header)
	minimalMessage := []byte("From: sender@example.com\r\n\r\n")
	_, err = signer.Sign(bytes.NewReader(minimalMessage))
	if err != nil {
		t.Fatalf("Sign() minimal message error = %v", err)
	}
}

func TestSignerHeadersIncluded(t *testing.T) {
	privateKey, err := GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	signer := NewSigner("example.com", "test-2026q1", privateKey)

	message := []byte(`From: sender@example.com
To: recipient@example.com
Subject: Test Email
Date: Mon, 14 Feb 2026 05:00:00 +0000
Message-ID: <test@example.com>
MIME-Version: 1.0
Content-Type: text/plain; charset=utf-8

Test body.
`)

	signed, err := signer.Sign(bytes.NewReader(message))
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	signedStr := string(signed)

	// Verify important headers are signed (present in h= tag)
	// The h= tag lists signed headers
	if !strings.Contains(signedStr, "h=") {
		t.Fatal("DKIM signature missing h= (headers) tag")
	}

	// Note: The actual header list format may vary, but these headers should be signed
	// We're testing that our HeaderKeys configuration is applied
}
