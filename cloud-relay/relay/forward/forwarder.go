// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package forward provides transport abstraction for forwarding mail
// from the cloud relay to the home device.
package forward

import (
	"context"
	"io"
)

// Forwarder defines the interface for forwarding mail through different
// transport layers (WireGuard or mTLS).
type Forwarder interface {
	// Forward sends the mail envelope and message data to the home device's
	// SMTP server via the configured transport.
	Forward(ctx context.Context, from string, to []string, data io.Reader) error

	// Close releases any resources held by the forwarder.
	Close() error
}
