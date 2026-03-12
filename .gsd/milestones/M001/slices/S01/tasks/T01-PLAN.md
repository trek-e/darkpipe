# T01: 01-transport-layer 01

**Slice:** S01 — **Milestone:** M001

## Description

Establish the WireGuard tunnel foundation: Go module initialization, WireGuard keypair generation, config file generation for hub (cloud) and spoke (home) topologies, and deployment scripts with systemd auto-restart.

Purpose: WireGuard is the primary encrypted transport between cloud relay and home device. This plan creates the tooling to generate, deploy, and manage WireGuard tunnels with proper NAT traversal (PersistentKeepalive=25) and systemd lifecycle management.

Output: Go module with WireGuard config generation library, keypair generator, deployment scripts for both cloud and home sides, systemd override for auto-restart on failure.

## Must-Haves

- [ ] "WireGuard config files are generated for both cloud (hub) and home (spoke) with correct key references and PersistentKeepalive=25 on home peer"
- [ ] "Key generation produces valid WireGuard keypairs that can be used in configs"
- [ ] "Deployment scripts install WireGuard, apply generated configs, and enable systemd service with auto-restart"

## Files

- `go.mod`
- `go.sum`
- `transport/wireguard/config/templates.go`
- `transport/wireguard/config/templates_test.go`
- `transport/wireguard/keygen/keygen.go`
- `transport/wireguard/keygen/keygen_test.go`
- `deploy/wireguard/cloud-setup.sh`
- `deploy/wireguard/home-setup.sh`
- `deploy/wireguard/systemd/wg-quick@wg0.service.d/override.conf`
