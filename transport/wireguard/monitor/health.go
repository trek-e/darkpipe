// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package monitor provides WireGuard tunnel health monitoring via handshake
// timestamp inspection. It uses wgctrl to read the kernel (or userspace)
// WireGuard device state and alerts when the latest handshake exceeds a
// configurable age threshold.
//
// With PersistentKeepalive=25, a healthy tunnel refreshes handshakes roughly
// every two minutes. The default alert threshold is five minutes -- long enough
// to avoid false positives, short enough to catch real outages quickly.
package monitor

import (
	"context"
	"fmt"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// DefaultMaxHandshakeAge is the default threshold for stale handshakes.
// With PersistentKeepalive=25s, handshakes refresh every ~2 minutes.
// Five minutes gives comfortable margin before alerting.
const DefaultMaxHandshakeAge = 5 * time.Minute

// DefaultMonitorInterval is the default tick interval for the Monitor loop.
const DefaultMonitorInterval = 30 * time.Second

// PeerHealth holds the health state of a single WireGuard peer.
type PeerHealth struct {
	PublicKey    string
	HandshakeAge time.Duration
	Healthy      bool
}

// TunnelHealth holds the aggregated health state of a WireGuard device.
type TunnelHealth struct {
	DeviceName string
	Peers      []PeerHealth
	Healthy    bool
	Error      string
}

// wgClient abstracts wgctrl.Client for testing. In production this is a
// real wgctrl.Client; in tests it can be replaced with a mock.
type wgClient interface {
	Device(name string) (*wgtypes.Device, error)
	Close() error
}

// newClient is the default factory function. Tests override this via
// checkTunnelHealthWithClient.
var newClient = func() (wgClient, error) {
	return wgctrl.New()
}

// CheckTunnelHealth inspects the WireGuard device identified by deviceName and
// returns an error if any peer's last handshake is older than maxHandshakeAge,
// or if no peers are configured. Returns nil when all peers are healthy.
func CheckTunnelHealth(deviceName string, maxHandshakeAge time.Duration) error {
	client, err := newClient()
	if err != nil {
		return fmt.Errorf("wgctrl init: %w", err)
	}
	defer client.Close()

	return checkHealth(client, deviceName, maxHandshakeAge)
}

// checkTunnelHealthWithClient is an internal function that accepts an existing
// client. Used for testing.
func checkTunnelHealthWithClient(client wgClient, deviceName string, maxHandshakeAge time.Duration) error {
	return checkHealth(client, deviceName, maxHandshakeAge)
}

func checkHealth(client wgClient, deviceName string, maxHandshakeAge time.Duration) error {
	device, err := client.Device(deviceName)
	if err != nil {
		return fmt.Errorf("get device %s: %w", deviceName, err)
	}

	if len(device.Peers) == 0 {
		return fmt.Errorf("no peers configured on %s", deviceName)
	}

	for _, peer := range device.Peers {
		// A zero handshake time means no handshake has ever completed.
		if peer.LastHandshakeTime.IsZero() {
			return fmt.Errorf("peer %s: no handshake completed yet",
				peer.PublicKey.String()[:16])
		}

		age := time.Since(peer.LastHandshakeTime)
		if age > maxHandshakeAge {
			return fmt.Errorf("peer %s: handshake too old (%s > %s)",
				peer.PublicKey.String()[:16], age.Round(time.Second), maxHandshakeAge)
		}
	}

	return nil
}

// GetTunnelHealth returns a structured TunnelHealth report for the given device.
// Unlike CheckTunnelHealth which returns only an error, this provides per-peer
// detail useful for monitoring dashboards.
func GetTunnelHealth(deviceName string, maxHandshakeAge time.Duration) TunnelHealth {
	client, err := newClient()
	if err != nil {
		return TunnelHealth{
			DeviceName: deviceName,
			Healthy:    false,
			Error:      fmt.Sprintf("wgctrl init: %v", err),
		}
	}
	defer client.Close()

	return getTunnelHealthWithClient(client, deviceName, maxHandshakeAge)
}

func getTunnelHealthWithClient(client wgClient, deviceName string, maxHandshakeAge time.Duration) TunnelHealth {
	device, err := client.Device(deviceName)
	if err != nil {
		return TunnelHealth{
			DeviceName: deviceName,
			Healthy:    false,
			Error:      fmt.Sprintf("get device %s: %v", deviceName, err),
		}
	}

	if len(device.Peers) == 0 {
		return TunnelHealth{
			DeviceName: deviceName,
			Healthy:    false,
			Error:      fmt.Sprintf("no peers configured on %s", deviceName),
		}
	}

	th := TunnelHealth{
		DeviceName: deviceName,
		Healthy:    true,
	}

	for _, peer := range device.Peers {
		ph := PeerHealth{
			PublicKey: peer.PublicKey.String(),
		}

		if peer.LastHandshakeTime.IsZero() {
			ph.Healthy = false
			th.Healthy = false
			if th.Error == "" {
				th.Error = fmt.Sprintf("peer %s: no handshake completed", ph.PublicKey[:16])
			}
		} else {
			ph.HandshakeAge = time.Since(peer.LastHandshakeTime)
			ph.Healthy = ph.HandshakeAge <= maxHandshakeAge
			if !ph.Healthy {
				th.Healthy = false
				if th.Error == "" {
					th.Error = fmt.Sprintf("peer %s: handshake too old (%s)",
						ph.PublicKey[:16], ph.HandshakeAge.Round(time.Second))
				}
			}
		}

		th.Peers = append(th.Peers, ph)
	}

	return th
}

// Monitor continuously checks the health of a WireGuard device at the given
// interval. When a health check fails, alertFn is called with the error.
// Monitor respects context cancellation for clean shutdown.
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	Monitor(ctx, "wg0", 30*time.Second, func(err error) {
//	    log.Printf("ALERT: %v", err)
//	})
func Monitor(ctx context.Context, deviceName string, interval time.Duration, alertFn func(error)) {
	if interval <= 0 {
		interval = DefaultMonitorInterval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := CheckTunnelHealth(deviceName, DefaultMaxHandshakeAge); err != nil {
				alertFn(err)
			}
		}
	}
}

// MonitorWithClient is like Monitor but accepts a custom client factory for
// testing. The factory is called once at monitor start.
func MonitorWithClient(ctx context.Context, clientFactory func() (wgClient, error), deviceName string, interval time.Duration, maxAge time.Duration, alertFn func(error)) {
	if interval <= 0 {
		interval = DefaultMonitorInterval
	}
	if maxAge <= 0 {
		maxAge = DefaultMaxHandshakeAge
	}

	client, err := clientFactory()
	if err != nil {
		alertFn(fmt.Errorf("wgctrl init: %w", err))
		return
	}
	defer client.Close()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := checkHealth(client, deviceName, maxAge); err != nil {
				alertFn(err)
			}
		}
	}
}
