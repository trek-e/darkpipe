package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	// CurrentVersion is the schema version for configuration
	CurrentVersion = "1"

	// ConfigFile is the default configuration file name
	ConfigFile = ".darkpipe.yml"

	// SecretsDir is the directory for Docker secrets
	SecretsDir = "secrets"
)

// Config represents the DarkPipe setup configuration
type Config struct {
	Version       string   `yaml:"version"`        // Schema version for migration
	MailDomain    string   `yaml:"mail_domain"`    // Primary mail domain (e.g., example.com)
	RelayHostname string   `yaml:"relay_hostname"` // Cloud relay FQDN (e.g., relay.example.com)
	MailServer    string   `yaml:"mail_server"`    // stalwart|maddy|postfix-dovecot
	Webmail       string   `yaml:"webmail"`        // none|roundcube|snappymail
	Calendar      string   `yaml:"calendar"`       // none|radicale|builtin
	AdminEmail    string   `yaml:"admin_email"`    // Admin email address
	ExtraDomains  []string `yaml:"extra_domains"`  // Additional mail domains
	Transport     string   `yaml:"transport"`      // wireguard|mtls
	QueueEnabled  bool     `yaml:"queue_enabled"`  // Enable message queuing
	StrictMode    bool     `yaml:"strict_mode"`    // TLS strict mode
}

// DefaultConfig returns a configuration with opinionated defaults
func DefaultConfig() *Config {
	return &Config{
		Version:      CurrentVersion,
		MailServer:   "stalwart",
		Webmail:      "snappymail",
		Calendar:     "builtin",
		Transport:    "wireguard",
		QueueEnabled: true,
		StrictMode:   false,
		ExtraDomains: []string{},
	}
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(config *Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
