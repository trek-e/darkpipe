// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package health

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/darkpipe/darkpipe/transport/health"
)

// CheckTunnel performs a health check on the configured transport tunnel
// (WireGuard or mTLS) by leveraging the transport/health package.
func CheckTunnel(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:   "tunnel",
		Status: "ok",
	}

	// Determine transport type from environment
	transportType := os.Getenv("TRANSPORT_TYPE")
	if transportType == "" {
		transportType = "wireguard" // default
	}

	var status health.HealthStatus

	switch transportType {
	case "wireguard":
		cfg := &health.WireGuardConfig{
			DeviceName:      os.Getenv("WG_DEVICE_NAME"),
			MaxHandshakeAge: 5 * time.Minute,
		}
		if cfg.DeviceName == "" {
			cfg.DeviceName = "wg0"
		}
		status = health.Check(health.WireGuard, cfg)

	case "mtls":
		cfg := &health.MTLSConfig{
			ServerAddr: os.Getenv("MTLS_SERVER"),
			CACertPath: os.Getenv("MTLS_CA_CERT"),
			CertPath:   os.Getenv("MTLS_CLIENT_CERT"),
			KeyPath:    os.Getenv("MTLS_CLIENT_KEY"),
		}
		status = health.Check(health.MTLS, cfg)

	default:
		result.Status = "error"
		result.Message = fmt.Sprintf("unknown transport type: %s", transportType)
		return result
	}

	// Convert transport HealthStatus to CheckResult
	if !status.Healthy {
		result.Status = "error"
		result.Message = status.Error
	} else {
		result.Message = fmt.Sprintf("%s tunnel healthy", transportType)
	}

	return result
}
