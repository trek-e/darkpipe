// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package cert

import (
	"fmt"
	"os/exec"
	"strings"
)

// ServiceConfig defines a service that needs reloading after cert renewal.
type ServiceConfig struct {
	Name          string // Service name (e.g., "postfix", "caddy")
	ReloadCommand string // Command to reload service (e.g., "postfix reload")
}

// ServiceReloader orchestrates service reloads after certificate changes.
type ServiceReloader struct {
	services []ServiceConfig
}

// NewServiceReloader creates a service reloader with the given services.
func NewServiceReloader(services []ServiceConfig) *ServiceReloader {
	return &ServiceReloader{
		services: services,
	}
}

// DefaultServices returns the standard DarkPipe services that need reload.
func DefaultServices() []ServiceConfig {
	return []ServiceConfig{
		{
			Name:          "postfix",
			ReloadCommand: "postfix reload",
		},
		{
			Name:          "caddy",
			ReloadCommand: "caddy reload --config /etc/caddy/Caddyfile",
		},
	}
}

// ReloadServices executes reload commands for all configured services.
// Uses hot reload (SIGHUP/reload command) to avoid service interruption.
// Executes sequentially to avoid race conditions.
func ReloadServices(services []ServiceConfig) error {
	var errors []string

	for _, svc := range services {
		// Parse command into args
		args := strings.Fields(svc.ReloadCommand)
		if len(args) == 0 {
			errors = append(errors, fmt.Sprintf("%s: empty reload command", svc.Name))
			continue
		}

		// Execute reload command
		cmd := exec.Command(args[0], args[1:]...)
		if err := cmd.Run(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", svc.Name, err))
			continue
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("reload failures: %s", strings.Join(errors, "; "))
	}

	return nil
}

// ReloadServices is a convenience method on ServiceReloader.
func (r *ServiceReloader) ReloadServices() error {
	return ReloadServices(r.services)
}

// GenerateSystemdPathUnit generates a systemd .path unit for certificate monitoring.
// The .path unit watches certificate files and triggers a reload service when they change.
//
// Example output:
// [Unit]
// Description=Watch certificate files for changes
//
// [Path]
// PathChanged=/etc/letsencrypt/live/example.com/fullchain.pem
// PathChanged=/etc/letsencrypt/live/example.com/privkey.pem
// Unit=reload-services.service
//
// [Install]
// WantedBy=multi-user.target
func GenerateSystemdPathUnit(certPaths []string, reloadService string) string {
	var builder strings.Builder

	builder.WriteString("[Unit]\n")
	builder.WriteString("Description=Watch certificate files for changes\n")
	builder.WriteString("\n")
	builder.WriteString("[Path]\n")

	for _, path := range certPaths {
		builder.WriteString(fmt.Sprintf("PathChanged=%s\n", path))
	}

	builder.WriteString(fmt.Sprintf("Unit=%s\n", reloadService))
	builder.WriteString("\n")
	builder.WriteString("[Install]\n")
	builder.WriteString("WantedBy=multi-user.target\n")

	return builder.String()
}
