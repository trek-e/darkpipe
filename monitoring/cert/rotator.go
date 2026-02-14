package cert

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// RenewalType identifies the certificate authority type.
type RenewalType string

const (
	LetsEncrypt RenewalType = "letsencrypt"
	StepCA      RenewalType = "stepca"
	SelfSigned  RenewalType = "selfsigned"
)

// RotatorConfig configures certificate renewal behavior.
type RotatorConfig struct {
	CertName    string      // Certificate name (for Let's Encrypt)
	RenewalType RenewalType // Type of renewal to perform
	DryRun      bool        // Dry-run mode (no actual changes)
	StagingMode bool        // Use Let's Encrypt staging server (for testing)
	CertPath    string      // Path to certificate file (for step-ca)
	KeyPath     string      // Path to key file (for step-ca)
}

// RenewWithRetry attempts to renew a certificate with exponential backoff retry.
// Retries up to 3 times for transient failures.
// Returns permanent errors immediately without retry.
func RenewWithRetry(config RotatorConfig) error {
	operation := func() error {
		return renew(config)
	}

	// Configure exponential backoff
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 5 * time.Minute

	// Wrap with retry limit
	boWithMaxRetries := backoff.WithMaxRetries(bo, 3)

	// Wrap operation to check for permanent errors
	err := backoff.RetryNotify(
		operation,
		boWithMaxRetries,
		func(err error, duration time.Duration) {
			// Log retry attempt (in production, use proper logging)
			fmt.Printf("Renewal failed, retrying in %v: %v\n", duration, err)
		},
	)

	return err
}

// renew performs the actual certificate renewal.
func renew(config RotatorConfig) error {
	var cmd *exec.Cmd

	switch config.RenewalType {
	case LetsEncrypt:
		args := []string{"renew", "--cert-name", config.CertName}
		if config.DryRun {
			args = append(args, "--dry-run")
		}
		if config.StagingMode {
			args = append(args, "--staging")
		}
		cmd = exec.Command("certbot", args...)

	case StepCA:
		args := []string{"ca", "renew", config.CertPath, config.KeyPath, "--force"}
		if config.DryRun {
			// step-ca doesn't have a dry-run flag, so skip execution
			return nil
		}
		cmd = exec.Command("step", args...)

	case SelfSigned:
		if config.DryRun {
			return nil
		}
		// Generate new self-signed certificate
		// This is a simplified example - production would need proper parameters
		args := []string{
			"req", "-x509", "-newkey", "rsa:2048",
			"-keyout", config.KeyPath,
			"-out", config.CertPath,
			"-days", "365",
			"-nodes",
			"-subj", fmt.Sprintf("/CN=%s", config.CertName),
		}
		cmd = exec.Command("openssl", args...)

	default:
		return fmt.Errorf("unsupported renewal type: %s", config.RenewalType)
	}

	// Execute renewal command
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if this is a permanent error
		if isPermanentError(output) {
			return backoff.Permanent(fmt.Errorf("permanent renewal error: %w\nOutput: %s", err, output))
		}
		return fmt.Errorf("renewal failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// isPermanentError checks if an error is permanent (should not retry).
// Permanent errors include ACME account issues, invalid domain errors, etc.
func isPermanentError(output []byte) bool {
	outputStr := string(output)

	// List of permanent error indicators
	permanentErrors := []string{
		"account does not exist",
		"invalid domain",
		"domain ownership verification failed",
		"CAA record forbids issuance",
		"too many certificates already issued",
		"certificate not found",
		"authorization deactivated",
	}

	for _, errMsg := range permanentErrors {
		if strings.Contains(strings.ToLower(outputStr), strings.ToLower(errMsg)) {
			return true
		}
	}

	return false
}

// RenewIfNeeded combines certificate checking and renewal.
// Checks if renewal is needed, and if so, attempts renewal with retry.
func RenewIfNeeded(watcher *CertWatcher, certPath string, config RotatorConfig) (renewed bool, err error) {
	// Check certificate status
	info, err := watcher.CheckCert(certPath)
	if err != nil {
		return false, fmt.Errorf("failed to check certificate: %w", err)
	}

	// If renewal not needed, return early
	if !info.ShouldRenew {
		return false, nil
	}

	// Attempt renewal
	if err := RenewWithRetry(config); err != nil {
		return false, fmt.Errorf("renewal failed after retries: %w", err)
	}

	return true, nil
}
