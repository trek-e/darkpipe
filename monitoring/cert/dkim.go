package cert

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/darkpipe/darkpipe/dns/dkim"
)

// DKIMRotationConfig configures DKIM key rotation.
type DKIMRotationConfig struct {
	Domain           string // Domain name for DKIM signing
	Prefix           string // Selector prefix (default: "darkpipe")
	KeyDir           string // Directory to store DKIM keys
	DNSProviderType  string // DNS provider type (cloudflare/route53/manual)
	LastRotationPath string // Path to file storing last rotation timestamp
}

// RotateDKIM performs quarterly DKIM key rotation.
// Steps:
// 1. Check if rotation is due (quarterly)
// 2. Generate new selector for next quarter
// 3. Generate new 2048-bit RSA key
// 4. Write key files with correct permissions
// 5. Output DNS record instructions
// 6. Update last rotation timestamp
//
// Note: Implements dual-key overlap strategy (keep old key for 7 days).
func RotateDKIM(config DKIMRotationConfig) error {
	// Set defaults
	if config.Prefix == "" {
		config.Prefix = "darkpipe"
	}
	if config.LastRotationPath == "" {
		config.LastRotationPath = filepath.Join(config.KeyDir, ".last-rotation")
	}

	// Check if rotation is due
	shouldRotate, err := ShouldRotateDKIM(config.LastRotationPath)
	if err != nil {
		return fmt.Errorf("failed to check rotation status: %w", err)
	}

	if !shouldRotate {
		return fmt.Errorf("rotation not due (still in same quarter)")
	}

	// Generate new selector for current quarter
	selector := dkim.GetCurrentSelector(config.Prefix)

	// Generate new 2048-bit RSA key pair
	privateKey, err := dkim.GenerateKeyPair(2048)
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Save key pair to disk
	if err := dkim.SaveKeyPair(privateKey, config.KeyDir, selector); err != nil {
		return fmt.Errorf("failed to save key pair: %w", err)
	}

	// Get public key base64 for DNS record
	publicKeyBase64, err := dkim.PublicKeyBase64(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to encode public key: %w", err)
	}

	// Output DNS record instructions
	fmt.Printf("\n=== DKIM Key Rotation Complete ===\n")
	fmt.Printf("Domain: %s\n", config.Domain)
	fmt.Printf("Selector: %s\n", selector)
	fmt.Printf("\nDNS Record to add:\n")
	fmt.Printf("Name: %s._domainkey.%s\n", selector, config.Domain)
	fmt.Printf("Type: TXT\n")
	fmt.Printf("Value: v=DKIM1; k=rsa; p=%s\n", publicKeyBase64)
	fmt.Printf("\nKey files:\n")
	fmt.Printf("Private: %s/%s.private.pem (0600)\n", config.KeyDir, selector)
	fmt.Printf("Public: %s/%s.public.pem (0644)\n", config.KeyDir, selector)
	fmt.Printf("\nDual-key overlap: Keep old key active for 7 days.\n")
	fmt.Printf("Update mail server signing config to use new selector.\n\n")

	// Update last rotation timestamp
	now := time.Now()
	if err := os.WriteFile(config.LastRotationPath, []byte(now.Format(time.RFC3339)), 0644); err != nil {
		return fmt.Errorf("failed to write last rotation timestamp: %w", err)
	}

	return nil
}

// ShouldRotateDKIM checks if DKIM rotation is due based on last rotation timestamp.
// Returns true if we're in a different quarter than the last rotation.
func ShouldRotateDKIM(lastRotationPath string) (bool, error) {
	// Read last rotation timestamp
	data, err := os.ReadFile(lastRotationPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No previous rotation - rotation is due
			return true, nil
		}
		return false, fmt.Errorf("failed to read last rotation file: %w", err)
	}

	// Parse timestamp
	lastRotation, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return false, fmt.Errorf("failed to parse last rotation timestamp: %w", err)
	}

	// Check if rotation is due using dkim.ShouldRotate
	return dkim.ShouldRotate(lastRotation), nil
}
