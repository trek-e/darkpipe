// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package compose

import (
	"fmt"
	"os"

	"github.com/darkpipe/darkpipe/deploy/setup/pkg/config"
	"gopkg.in/yaml.v3"
)

// Generate creates a docker-compose.yml file based on the configuration
func Generate(cfg *config.Config, outputPath string) error {
	compose := &ComposeFile{
		Version:  "3.8",
		Services: make(map[string]*ComposeService),
		Networks: make(map[string]*ComposeNetwork),
		Volumes:  make(map[string]*ComposeVolume),
		Secrets:  make(map[string]*ComposeSecret),
	}

	// Always include cloud relay
	compose.Services["relay"] = cloudRelayService(cfg)

	// Add selected mail server
	switch cfg.MailServer {
	case "stalwart":
		compose.Services["stalwart"] = stalwartService(cfg)
	case "maddy":
		compose.Services["maddy"] = maddyService(cfg)
	case "postfix-dovecot":
		compose.Services["postfix-dovecot"] = postfixDovecotService(cfg)
	}

	// Add webmail if selected
	switch cfg.Webmail {
	case "roundcube":
		compose.Services["roundcube"] = roundcubeService(cfg)
		compose.Services["caddy"] = caddyService(cfg)
	case "snappymail":
		compose.Services["snappymail"] = snappymailService(cfg)
		compose.Services["caddy"] = caddyService(cfg)
	}

	// Add calendar if selected (and not builtin)
	if cfg.Calendar == "radicale" {
		compose.Services["radicale"] = radicaleService(cfg)
		// Add Caddy if not already added
		if cfg.Webmail == "none" {
			compose.Services["caddy"] = caddyService(cfg)
		}
	}

	// Always include spam filtering services
	compose.Services["rspamd"] = rspamdService()
	compose.Services["redis"] = redisService()

	// Networks
	compose.Networks["darkpipe"] = &ComposeNetwork{
		Driver: "bridge",
	}

	// Volumes
	compose.Volumes["postfix-queue"] = &ComposeVolume{Driver: "local"}
	compose.Volumes["certbot-etc"] = &ComposeVolume{Driver: "local"}
	compose.Volumes["queue-data"] = &ComposeVolume{Driver: "local"}
	compose.Volumes["mail-data"] = &ComposeVolume{Driver: "local"}
	compose.Volumes["rspamd-data"] = &ComposeVolume{Driver: "local"}
	compose.Volumes["redis-data"] = &ComposeVolume{Driver: "local"}

	if cfg.MailServer == "postfix-dovecot" {
		compose.Volumes["mail-config"] = &ComposeVolume{Driver: "local"}
	}

	if cfg.Webmail == "roundcube" {
		compose.Volumes["roundcube-data"] = &ComposeVolume{Driver: "local"}
	}

	if cfg.Webmail == "snappymail" {
		compose.Volumes["snappymail-data"] = &ComposeVolume{Driver: "local"}
	}

	if cfg.Calendar == "radicale" {
		compose.Volumes["radicale-data"] = &ComposeVolume{Driver: "local"}
	}

	if cfg.Webmail != "none" || cfg.Calendar == "radicale" {
		compose.Volumes["caddy-data"] = &ComposeVolume{Driver: "local"}
		compose.Volumes["caddy-config"] = &ComposeVolume{Driver: "local"}
	}

	// Docker secrets
	compose.Secrets["admin_password"] = &ComposeSecret{
		File: "./secrets/admin_password.txt",
	}
	compose.Secrets["dkim_private_key"] = &ComposeSecret{
		File: "./secrets/dkim_private_key.pem",
	}

	// Marshal to YAML
	data, err := yaml.Marshal(compose)
	if err != nil {
		return fmt.Errorf("failed to marshal compose file: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write compose file: %w", err)
	}

	return nil
}
