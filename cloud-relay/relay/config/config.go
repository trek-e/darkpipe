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
