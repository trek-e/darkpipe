// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package validate

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// ValidateSMTPPort checks if SMTP port 25 is accessible
func ValidateSMTPPort(hostname string, port int, timeout time.Duration) error {
	address := net.JoinHostPort(hostname, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return fmt.Errorf("port %d not accessible: %w - many VPS providers block SMTP. See VPS provider guide", port, err)
	}
	defer conn.Close()

	return nil
}

// ValidateSMTPBanner connects to port 25 and reads the SMTP banner
func ValidateSMTPBanner(hostname string, timeout time.Duration) (string, error) {
	address := net.JoinHostPort(hostname, "25")
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return "", fmt.Errorf("failed to connect to SMTP port: %w", err)
	}
	defer conn.Close()

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(timeout))

	// Read SMTP banner (should start with "220")
	reader := bufio.NewReader(conn)
	banner, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read SMTP banner: %w", err)
	}

	banner = strings.TrimSpace(banner)
	if !strings.HasPrefix(banner, "220") {
		return "", fmt.Errorf("invalid SMTP banner (expected 220, got: %s)", banner)
	}

	return banner, nil
}
