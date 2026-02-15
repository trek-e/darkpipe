// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package health

import (
	"context"
	"fmt"
	"net"
	"time"
)

// CheckPostfix performs a TCP connection check to the Postfix SMTP port.
// This validates that Postfix is listening and accepting connections.
func CheckPostfix(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:   "postfix",
		Status: "ok",
	}

	// Create dialer with timeout
	var d net.Dialer
	d.Timeout = 2 * time.Second

	conn, err := d.DialContext(ctx, "tcp", "localhost:25")
	if err != nil {
		result.Status = "error"
		result.Message = fmt.Sprintf("failed to connect: %v", err)
		return result
	}
	defer conn.Close()

	result.Message = "SMTP port responsive"
	return result
}
