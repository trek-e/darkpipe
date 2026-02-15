// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package monitor

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// mockClient implements wgClient for testing without requiring root access
// or a real WireGuard kernel device.
type mockClient struct {
	device *wgtypes.Device
	err    error
}

func (m *mockClient) Device(name string) (*wgtypes.Device, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.device == nil {
		return nil, fmt.Errorf("device %s not found", name)
	}
	return m.device, nil
}

func (m *mockClient) Close() error { return nil }

// testKey generates a deterministic wgtypes.Key for testing.
func testKey(b byte) wgtypes.Key {
	var k wgtypes.Key
	k[0] = b
	return k
}

func TestCheckTunnelHealth_HealthyPeer(t *testing.T) {
	client := &mockClient{
		device: &wgtypes.Device{
			Name: "wg0",
			Peers: []wgtypes.Peer{
				{
					PublicKey:          testKey(1),
					LastHandshakeTime: time.Now().Add(-1 * time.Minute),
				},
			},
		},
	}

	err := checkTunnelHealthWithClient(client, "wg0", DefaultMaxHandshakeAge)
	if err != nil {
		t.Fatalf("expected healthy peer, got error: %v", err)
	}
}

func TestCheckTunnelHealth_StaleHandshake(t *testing.T) {
	client := &mockClient{
		device: &wgtypes.Device{
			Name: "wg0",
			Peers: []wgtypes.Peer{
				{
					PublicKey:          testKey(1),
					LastHandshakeTime: time.Now().Add(-10 * time.Minute),
				},
			},
		},
	}

	err := checkTunnelHealthWithClient(client, "wg0", DefaultMaxHandshakeAge)
	if err == nil {
		t.Fatal("expected error for stale handshake, got nil")
	}
	t.Logf("expected error: %v", err)
}

func TestCheckTunnelHealth_NoPeers(t *testing.T) {
	client := &mockClient{
		device: &wgtypes.Device{
			Name:  "wg0",
			Peers: []wgtypes.Peer{},
		},
	}

	err := checkTunnelHealthWithClient(client, "wg0", DefaultMaxHandshakeAge)
	if err == nil {
		t.Fatal("expected error for no peers, got nil")
	}
	t.Logf("expected error: %v", err)
}

func TestCheckTunnelHealth_NoHandshakeYet(t *testing.T) {
	client := &mockClient{
		device: &wgtypes.Device{
			Name: "wg0",
			Peers: []wgtypes.Peer{
				{
					PublicKey:          testKey(1),
					LastHandshakeTime: time.Time{}, // zero value
				},
			},
		},
	}

	err := checkTunnelHealthWithClient(client, "wg0", DefaultMaxHandshakeAge)
	if err == nil {
		t.Fatal("expected error for zero handshake time, got nil")
	}
	t.Logf("expected error: %v", err)
}

func TestCheckTunnelHealth_DeviceNotFound(t *testing.T) {
	client := &mockClient{
		err: fmt.Errorf("device wg0 not found"),
	}

	err := checkTunnelHealthWithClient(client, "wg0", DefaultMaxHandshakeAge)
	if err == nil {
		t.Fatal("expected error for missing device, got nil")
	}
	t.Logf("expected error: %v", err)
}

func TestCheckTunnelHealth_MultiplePeers_OneStale(t *testing.T) {
	client := &mockClient{
		device: &wgtypes.Device{
			Name: "wg0",
			Peers: []wgtypes.Peer{
				{
					PublicKey:          testKey(1),
					LastHandshakeTime: time.Now().Add(-1 * time.Minute),
				},
				{
					PublicKey:          testKey(2),
					LastHandshakeTime: time.Now().Add(-10 * time.Minute),
				},
			},
		},
	}

	err := checkTunnelHealthWithClient(client, "wg0", DefaultMaxHandshakeAge)
	if err == nil {
		t.Fatal("expected error when one peer is stale, got nil")
	}
	t.Logf("expected error: %v", err)
}

func TestGetTunnelHealth_Healthy(t *testing.T) {
	client := &mockClient{
		device: &wgtypes.Device{
			Name: "wg0",
			Peers: []wgtypes.Peer{
				{
					PublicKey:          testKey(1),
					LastHandshakeTime: time.Now().Add(-1 * time.Minute),
				},
			},
		},
	}

	th := getTunnelHealthWithClient(client, "wg0", DefaultMaxHandshakeAge)
	if !th.Healthy {
		t.Fatalf("expected healthy tunnel, got error: %s", th.Error)
	}
	if len(th.Peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(th.Peers))
	}
	if !th.Peers[0].Healthy {
		t.Fatal("expected peer to be healthy")
	}
}

func TestGetTunnelHealth_Unhealthy(t *testing.T) {
	client := &mockClient{
		device: &wgtypes.Device{
			Name: "wg0",
			Peers: []wgtypes.Peer{
				{
					PublicKey:          testKey(1),
					LastHandshakeTime: time.Now().Add(-10 * time.Minute),
				},
			},
		},
	}

	th := getTunnelHealthWithClient(client, "wg0", DefaultMaxHandshakeAge)
	if th.Healthy {
		t.Fatal("expected unhealthy tunnel")
	}
	if th.Error == "" {
		t.Fatal("expected error message")
	}
	if len(th.Peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(th.Peers))
	}
	if th.Peers[0].Healthy {
		t.Fatal("expected peer to be unhealthy")
	}
}

func TestGetTunnelHealth_DeviceError(t *testing.T) {
	client := &mockClient{
		err: fmt.Errorf("not available"),
	}

	th := getTunnelHealthWithClient(client, "wg0", DefaultMaxHandshakeAge)
	if th.Healthy {
		t.Fatal("expected unhealthy status on device error")
	}
	if th.Error == "" {
		t.Fatal("expected error message")
	}
}

func TestMonitorWithClient_AlertOnFailure(t *testing.T) {
	client := &mockClient{
		device: &wgtypes.Device{
			Name: "wg0",
			Peers: []wgtypes.Peer{
				{
					PublicKey:          testKey(1),
					LastHandshakeTime: time.Now().Add(-10 * time.Minute),
				},
			},
		},
	}

	var mu sync.Mutex
	var alerts []error

	ctx, cancel := context.WithCancel(context.Background())

	factory := func() (wgClient, error) {
		return client, nil
	}

	go MonitorWithClient(ctx, factory, "wg0", 10*time.Millisecond, DefaultMaxHandshakeAge, func(err error) {
		mu.Lock()
		alerts = append(alerts, err)
		mu.Unlock()
	})

	// Wait for at least one tick.
	time.Sleep(50 * time.Millisecond)
	cancel()

	mu.Lock()
	count := len(alerts)
	mu.Unlock()

	if count == 0 {
		t.Fatal("expected at least one alert, got none")
	}
	t.Logf("received %d alerts", count)
}

func TestMonitorWithClient_NoAlertWhenHealthy(t *testing.T) {
	client := &mockClient{
		device: &wgtypes.Device{
			Name: "wg0",
			Peers: []wgtypes.Peer{
				{
					PublicKey:          testKey(1),
					LastHandshakeTime: time.Now().Add(-1 * time.Minute),
				},
			},
		},
	}

	var mu sync.Mutex
	var alerts []error

	ctx, cancel := context.WithCancel(context.Background())

	factory := func() (wgClient, error) {
		return client, nil
	}

	go MonitorWithClient(ctx, factory, "wg0", 10*time.Millisecond, DefaultMaxHandshakeAge, func(err error) {
		mu.Lock()
		alerts = append(alerts, err)
		mu.Unlock()
	})

	time.Sleep(50 * time.Millisecond)
	cancel()

	mu.Lock()
	count := len(alerts)
	mu.Unlock()

	if count != 0 {
		t.Fatalf("expected no alerts for healthy tunnel, got %d", count)
	}
}

func TestMonitorWithClient_ContextCancellation(t *testing.T) {
	client := &mockClient{
		device: &wgtypes.Device{
			Name: "wg0",
			Peers: []wgtypes.Peer{
				{
					PublicKey:          testKey(1),
					LastHandshakeTime: time.Now().Add(-1 * time.Minute),
				},
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	factory := func() (wgClient, error) {
		return client, nil
	}

	go func() {
		MonitorWithClient(ctx, factory, "wg0", 10*time.Millisecond, DefaultMaxHandshakeAge, func(err error) {})
		close(done)
	}()

	// Cancel immediately.
	cancel()

	select {
	case <-done:
		// Monitor exited cleanly.
	case <-time.After(2 * time.Second):
		t.Fatal("Monitor did not exit after context cancellation")
	}
}

func TestMonitorWithClient_ClientFactoryError(t *testing.T) {
	var mu sync.Mutex
	var alerts []error

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	factory := func() (wgClient, error) {
		return nil, fmt.Errorf("permission denied")
	}

	done := make(chan struct{})
	go func() {
		MonitorWithClient(ctx, factory, "wg0", 10*time.Millisecond, DefaultMaxHandshakeAge, func(err error) {
			mu.Lock()
			alerts = append(alerts, err)
			mu.Unlock()
		})
		close(done)
	}()

	select {
	case <-done:
		// Expected -- factory error causes immediate return.
	case <-time.After(2 * time.Second):
		t.Fatal("Monitor did not exit after factory error")
	}

	mu.Lock()
	count := len(alerts)
	mu.Unlock()

	if count != 1 {
		t.Fatalf("expected exactly 1 alert for factory error, got %d", count)
	}
}

func TestCheckTunnelHealth_CustomMaxAge(t *testing.T) {
	client := &mockClient{
		device: &wgtypes.Device{
			Name: "wg0",
			Peers: []wgtypes.Peer{
				{
					PublicKey:          testKey(1),
					LastHandshakeTime: time.Now().Add(-3 * time.Minute),
				},
			},
		},
	}

	// With a 2-minute threshold, a 3-minute-old handshake should fail.
	err := checkTunnelHealthWithClient(client, "wg0", 2*time.Minute)
	if err == nil {
		t.Fatal("expected error with 2min max age on 3min old handshake")
	}

	// With a 10-minute threshold, the same handshake should pass.
	err = checkTunnelHealthWithClient(client, "wg0", 10*time.Minute)
	if err != nil {
		t.Fatalf("expected no error with 10min max age, got: %v", err)
	}
}
