// Package config provides configuration management for the cloud relay daemon.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the relay daemon configuration.
type Config struct {
	ListenAddr    string
	TransportType string // "wireguard" or "mtls"
	HomeDeviceAddr string

	// mTLS-specific fields
	CACertPath     string
	ClientCertPath string
	ClientKeyPath  string

	// Message handling
	MaxMessageBytes int64
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration

	// TLS monitoring and notifications
	StrictModeEnabled bool
	WebhookURL        string

	// Queue configuration (QUEUE-01, QUEUE-02, QUEUE-03)
	QueueEnabled      bool   // QUEUE-03: false = reject when offline
	QueueKeyPath      string // Path to age identity file for encryption
	QueueMaxRAMBytes  int64  // Max RAM for queue (default 200MB)
	QueueMaxMessages  int    // Max messages in queue (default 10000)
	QueueTTLHours     int    // Max age before purge (default 168 = 7 days)
	QueueSnapshotPath string // Path for queue metadata snapshots

	// S3 overflow configuration (QUEUE-02, optional)
	OverflowEnabled   bool
	OverflowEndpoint  string // e.g., "gateway.storjshare.io" or "s3.amazonaws.com"
	OverflowBucket    string
	OverflowAccessKey string
	OverflowSecretKey string
	OverflowUseSSL    bool // default true
}

// LoadFromEnv loads configuration from environment variables with sensible defaults.
func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		ListenAddr:        getEnv("RELAY_LISTEN_ADDR", "127.0.0.1:10025"),
		TransportType:     getEnv("RELAY_TRANSPORT", "wireguard"),
		HomeDeviceAddr:    getEnv("RELAY_HOME_ADDR", "10.8.0.2:25"),
		CACertPath:        getEnv("RELAY_CA_CERT", ""),
		ClientCertPath:    getEnv("RELAY_CLIENT_CERT", ""),
		ClientKeyPath:     getEnv("RELAY_CLIENT_KEY", ""),
		MaxMessageBytes:   getEnvInt64("RELAY_MAX_MESSAGE_BYTES", 50*1024*1024), // 50MB
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		StrictModeEnabled: getEnvBool("RELAY_STRICT_MODE", false),
		WebhookURL:        getEnv("RELAY_WEBHOOK_URL", ""),
		QueueEnabled:      getEnvBool("RELAY_QUEUE_ENABLED", true),                                 // Enabled by default
		QueueKeyPath:      getEnv("RELAY_QUEUE_KEY_PATH", "/data/queue-keys/identity"),         // Default path
		QueueMaxRAMBytes:  getEnvInt64("RELAY_QUEUE_MAX_RAM", 200*1024*1024),                   // 200MB
		QueueMaxMessages:  int(getEnvInt64("RELAY_QUEUE_MAX_MESSAGES", 10000)),                 // 10k messages
		QueueTTLHours:     int(getEnvInt64("RELAY_QUEUE_TTL_HOURS", 168)),                      // 7 days
		QueueSnapshotPath: getEnv("RELAY_QUEUE_SNAPSHOT_PATH", "/data/queue-state/snapshot.json"), // Default path
		OverflowEnabled:   getEnvBool("RELAY_OVERFLOW_ENABLED", false),                         // Disabled by default (requires S3 credentials)
		OverflowEndpoint:  getEnv("RELAY_OVERFLOW_ENDPOINT", ""),
		OverflowBucket:    getEnv("RELAY_OVERFLOW_BUCKET", "darkpipe-queue"),
		OverflowAccessKey: getEnv("RELAY_OVERFLOW_ACCESS_KEY", ""),
		OverflowSecretKey: getEnv("RELAY_OVERFLOW_SECRET_KEY", ""),
		OverflowUseSSL:    getEnvBool("RELAY_OVERFLOW_USE_SSL", true),
	}

	// Validate based on transport type
	if cfg.TransportType != "wireguard" && cfg.TransportType != "mtls" {
		return nil, fmt.Errorf("invalid RELAY_TRANSPORT: %s (must be 'wireguard' or 'mtls')", cfg.TransportType)
	}

	if cfg.TransportType == "mtls" {
		if cfg.CACertPath == "" || cfg.ClientCertPath == "" || cfg.ClientKeyPath == "" {
			return nil, fmt.Errorf("mTLS transport requires RELAY_CA_CERT, RELAY_CLIENT_CERT, and RELAY_CLIENT_KEY")
		}
	}

	if cfg.HomeDeviceAddr == "" {
		return nil, fmt.Errorf("RELAY_HOME_ADDR is required")
	}

	// Validate overflow configuration if enabled
	if cfg.OverflowEnabled {
		if cfg.OverflowEndpoint == "" {
			return nil, fmt.Errorf("RELAY_OVERFLOW_ENDPOINT is required when overflow is enabled")
		}
		if cfg.OverflowAccessKey == "" {
			return nil, fmt.Errorf("RELAY_OVERFLOW_ACCESS_KEY is required when overflow is enabled")
		}
		if cfg.OverflowSecretKey == "" {
			return nil, fmt.Errorf("RELAY_OVERFLOW_SECRET_KEY is required when overflow is enabled")
		}
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
