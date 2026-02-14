package dkim

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	tests := []struct {
		name    string
		bits    int
		wantErr bool
	}{
		{
			name:    "2048 bit key",
			bits:    2048,
			wantErr: false,
		},
		{
			name:    "4096 bit key",
			bits:    4096,
			wantErr: false,
		},
		{
			name:    "1024 bit key (too small)",
			bits:    1024,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privateKey, err := GenerateKeyPair(tt.bits)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GenerateKeyPair() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if privateKey == nil {
					t.Fatal("GenerateKeyPair() returned nil private key")
				}
				if privateKey.N.BitLen() != tt.bits {
					t.Fatalf("GenerateKeyPair() key size = %d, want %d", privateKey.N.BitLen(), tt.bits)
				}
			}
		})
	}
}

func TestSaveAndLoadKeyPair(t *testing.T) {
	// Generate a test key
	privateKey, err := GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	// Create temp directory
	tmpDir := t.TempDir()
	selector := "test-2026q1"

	// Save key pair
	if err := SaveKeyPair(privateKey, tmpDir, selector); err != nil {
		t.Fatalf("SaveKeyPair() error = %v", err)
	}

	// Verify files exist with correct permissions
	privateKeyPath := filepath.Join(tmpDir, selector+".private.pem")
	publicKeyPath := filepath.Join(tmpDir, selector+".public.pem")

	privateKeyInfo, err := os.Stat(privateKeyPath)
	if err != nil {
		t.Fatalf("Private key file not found: %v", err)
	}
	if privateKeyInfo.Mode().Perm() != 0600 {
		t.Fatalf("Private key file permissions = %o, want 0600", privateKeyInfo.Mode().Perm())
	}

	publicKeyInfo, err := os.Stat(publicKeyPath)
	if err != nil {
		t.Fatalf("Public key file not found: %v", err)
	}
	if publicKeyInfo.Mode().Perm() != 0644 {
		t.Fatalf("Public key file permissions = %o, want 0644", publicKeyInfo.Mode().Perm())
	}

	// Load private key back
	loadedKey, err := LoadPrivateKey(privateKeyPath)
	if err != nil {
		t.Fatalf("LoadPrivateKey() error = %v", err)
	}

	// Verify loaded key matches original
	if !privateKey.Equal(loadedKey) {
		t.Fatal("Loaded private key does not match original")
	}
}

func TestPublicKeyBase64(t *testing.T) {
	// Generate a test key
	privateKey, err := GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	// Get base64 encoding
	base64Str, err := PublicKeyBase64(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("PublicKeyBase64() error = %v", err)
	}

	// Verify it's valid base64
	decoded, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		t.Fatalf("Base64 decode error = %v", err)
	}

	// Verify decoded data is non-empty
	if len(decoded) == 0 {
		t.Fatal("PublicKeyBase64() returned empty base64 string")
	}
}

func TestLoadPrivateKeyErrors(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "non-existent file",
			path:    "/nonexistent/path/key.pem",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadPrivateKey(tt.path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
