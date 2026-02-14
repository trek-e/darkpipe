package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// DNSConfig holds configuration for DNS tools, loaded from environment variables.
type DNSConfig struct {
	// Domain settings
	Domain string

	// Cloud relay settings
	RelayHostname string // FQDN of cloud relay (e.g., "relay.darkpipe.com")
	RelayIP       string // Public IP of cloud relay

	// DKIM settings
	DKIMKeyBits        int    // Key size in bits (default: 2048)
	DKIMKeyDir         string // Directory for DKIM keys (default: /etc/darkpipe/dkim)
	DKIMSelectorPrefix string // Prefix for selector naming (default: "darkpipe")

	// DMARC settings
	DMARCPolicy      string // Policy: "none", "quarantine", or "reject" (default: "none")
	DMARCRua         string // Aggregate report email address
	DMARCRuf         string // Forensic report email address

	// Output settings
	OutputFormat string // "text" or "json" (default: "text")
	RecordsFile  string // Path to DNS-RECORDS.md file (default: "DNS-RECORDS.md")

	// Validation settings
	PropagationTimeout time.Duration // Timeout for DNS propagation checks (default: 5 minutes)
	DNSServers         []string      // DNS servers for validation queries
}

// LoadFromEnv loads configuration from environment variables.
// Required variables: DARKPIPE_DOMAIN, RELAY_HOSTNAME, RELAY_IP
func LoadFromEnv() (*DNSConfig, error) {
	cfg := &DNSConfig{
		Domain:             getEnv("DARKPIPE_DOMAIN", ""),
		RelayHostname:      getEnv("RELAY_HOSTNAME", ""),
		RelayIP:            getEnv("RELAY_IP", ""),
		DKIMKeyBits:        getEnvInt("DKIM_KEY_BITS", 2048),
		DKIMKeyDir:         getEnv("DKIM_KEY_DIR", "/etc/darkpipe/dkim"),
		DKIMSelectorPrefix: getEnv("DKIM_SELECTOR_PREFIX", "darkpipe"),
		DMARCPolicy:        getEnv("DMARC_POLICY", "none"),
		DMARCRua:           getEnv("DMARC_RUA", ""),
		DMARCRuf:           getEnv("DMARC_RUF", ""),
		OutputFormat:       getEnv("OUTPUT_FORMAT", "text"),
		RecordsFile:        getEnv("DNS_RECORDS_FILE", "DNS-RECORDS.md"),
		PropagationTimeout: getEnvDuration("DNS_PROPAGATION_TIMEOUT", 5*time.Minute),
		DNSServers:         getEnvSlice("DNS_SERVERS", []string{"8.8.8.8:53", "1.1.1.1:53"}),
	}

	// Validate required fields
	if cfg.Domain == "" {
		return nil, fmt.Errorf("DARKPIPE_DOMAIN environment variable is required")
	}
	if cfg.RelayHostname == "" {
		return nil, fmt.Errorf("RELAY_HOSTNAME environment variable is required")
	}
	if cfg.RelayIP == "" {
		return nil, fmt.Errorf("RELAY_IP environment variable is required")
	}

	// Validate DMARC policy
	if cfg.DMARCPolicy != "none" && cfg.DMARCPolicy != "quarantine" && cfg.DMARCPolicy != "reject" {
		return nil, fmt.Errorf("DMARC_POLICY must be 'none', 'quarantine', or 'reject' (got %s)", cfg.DMARCPolicy)
	}

	// Validate output format
	if cfg.OutputFormat != "text" && cfg.OutputFormat != "json" {
		return nil, fmt.Errorf("OUTPUT_FORMAT must be 'text' or 'json' (got %s)", cfg.OutputFormat)
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	// Simple comma-separated parsing
	// For more complex needs, use a proper CSV parser
	if value := os.Getenv(key); value != "" {
		// This is a simplified version - for production you might want to handle escaping
		return []string{value} // Single value for now
	}
	return defaultValue
}
