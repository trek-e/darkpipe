# S01: Transport Layer

**Goal:** Establish the WireGuard tunnel foundation: Go module initialization, WireGuard keypair generation, config file generation for hub (cloud) and spoke (home) topologies, and deployment scripts with systemd auto-restart.
**Demo:** Establish the WireGuard tunnel foundation: Go module initialization, WireGuard keypair generation, config file generation for hub (cloud) and spoke (home) topologies, and deployment scripts with systemd auto-restart.

## Must-Haves


## Tasks

- [x] **T01: 01-transport-layer 01**
  - Establish the WireGuard tunnel foundation: Go module initialization, WireGuard keypair generation, config file generation for hub (cloud) and spoke (home) topologies, and deployment scripts with systemd auto-restart.

Purpose: WireGuard is the primary encrypted transport between cloud relay and home device. This plan creates the tooling to generate, deploy, and manage WireGuard tunnels with proper NAT traversal (PersistentKeepalive=25) and systemd lifecycle management.

Output: Go module with WireGuard config generation library, keypair generator, deployment scripts for both cloud and home sides, systemd override for auto-restart on failure.
- [x] **T02: 01-transport-layer 02** `est:7min`
  - Establish the mTLS alternative transport and internal PKI: step-ca deployment as private CA, mTLS server (cloud relay) and client (home device) with mutual certificate verification, persistent connection with exponential backoff reconnection, and automated certificate renewal via systemd timers.

Purpose: mTLS provides an application-level alternative to WireGuard for users who prefer it or cannot use kernel WireGuard. The internal CA (step-ca) issues short-lived certificates distinct from public-facing TLS, supporting both WireGuard cert needs and mTLS transport with automated renewal.

Output: Go packages for mTLS server/client with auto-reconnect, step-ca setup tooling, certificate renewal automation via systemd, deployment script for CA initialization.
- [x] **T03: 01-transport-layer 03** `est:3min`
  - Harden auto-reconnection, implement transport health monitoring, create integration test scripts for outage simulation, and document VPS provider selection for port 25 SMTP.

Purpose: The transport layer must survive real-world internet interruptions without manual intervention. This plan adds the monitoring and self-healing that turns a working tunnel into a resilient one. VPS provider documentation ensures users can deploy on providers that allow SMTP traffic.

Output: WireGuard health monitor (handshake-based), unified transport health checker, outage simulation script, tunnel connectivity test, VPS provider guide.

## Files Likely Touched

- `go.mod`
- `go.sum`
- `transport/wireguard/config/templates.go`
- `transport/wireguard/config/templates_test.go`
- `transport/wireguard/keygen/keygen.go`
- `transport/wireguard/keygen/keygen_test.go`
- `deploy/wireguard/cloud-setup.sh`
- `deploy/wireguard/home-setup.sh`
- `deploy/wireguard/systemd/wg-quick@wg0.service.d/override.conf`
- `transport/pki/ca/setup.go`
- `transport/pki/ca/setup_test.go`
- `transport/pki/renewal/systemd/cert-renewer@.service`
- `transport/pki/renewal/systemd/cert-renewer@.timer`
- `transport/pki/renewal/hooks/reload.sh`
- `transport/mtls/server/listener.go`
- `transport/mtls/server/listener_test.go`
- `transport/mtls/client/connector.go`
- `transport/mtls/client/connector_test.go`
- `deploy/pki/step-ca-setup.sh`
- `transport/wireguard/monitor/health.go`
- `transport/wireguard/monitor/health_test.go`
- `transport/wireguard/monitor/reconnect.sh`
- `transport/health/checker.go`
- `transport/health/checker_test.go`
- `deploy/test/tunnel-test.sh`
- `deploy/test/outage-sim.sh`
- `docs/vps-providers.md`
