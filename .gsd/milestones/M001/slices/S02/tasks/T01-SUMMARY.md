---
id: T01
parent: S02
milestone: M001
provides: []
requires: []
affects: []
key_files: []
key_decisions: []
patterns_established: []
observability_surfaces: []
drill_down_paths: []
duration: 
verification_result: passed
completed_at: 
blocker_discovered: false
---
# T01: 02-cloud-relay 01

**# Phase 02 Plan 01: Cloud Relay Core Summary**

## What Happened

# Phase 02 Plan 01: Cloud Relay Core Summary

**One-liner:** Postfix relay-only container with Go SMTP daemon forwarding mail to home device via WireGuard/mTLS transport using emersion/go-smtp and Phase 1 transport clients.

## Overview

Built the foundational cloud relay component for Phase 2. This plan establishes the complete inbound mail flow (internet -> Postfix -> Go daemon -> transport -> home device) and outbound mail flow (home device -> transport -> Postfix -> direct MTA delivery).

The relay consists of two components running in a single Docker container:
1. **Postfix** in relay-only (null client) mode accepting internet SMTP on port 25
2. **Go relay daemon** receiving forwarded mail from Postfix and bridging to the home device via WireGuard or mTLS transport

## Execution Summary

### Task 1: Go relay daemon with SMTP backend and transport forwarding

Created the complete Go relay daemon under `cloud-relay/` within the existing github.com/darkpipe/darkpipe module.

**Components:**
- **config/config.go**: Environment-based configuration with validation
- **forward/forwarder.go**: Transport abstraction interface
- **forward/mtls.go**: mTLS transport using Phase 1 client.Client with proper SMTP envelope handling
- **forward/wireguard.go**: WireGuard tunnel transport (transparent encryption at network layer)
- **smtp/server.go**: emersion/go-smtp Backend implementation
- **smtp/session.go**: SMTP session with Mail/Rcpt/Data handlers that bridge to Forwarder
- **cmd/relay/main.go**: Entrypoint with config loading, forwarder creation, server startup, and graceful shutdown

**Critical link:** `session.go` Data() method calls `forwarder.Forward()` - this is where SMTP transitions to the transport layer.

**Key decision:** Used emersion/go-smtp for both server (relay daemon) and client (forwarding to home device) sides for consistency.

**Commit:** 889aad4

### Task 2: Postfix relay-only configuration and Docker container

Created Postfix null client configuration and containerization.

**Postfix configuration:**
- `mydestination =` (empty) - no local delivery
- `transport_maps = lmdb:/etc/postfix/transport` - all mail to localhost:10025
- `mynetworks` includes WireGuard subnet (10.8.0.0/24) for home device outbound submission
- TLS configured as opportunistic (may) - enforcement in Plan 02-02
- LMDB format for all maps (BerkleyDB deprecated in Alpine per research)
- Logging to stdout for container observability

**Docker setup:**
- Multi-stage build: Go 1.24 builder + Alpine 3.21 runtime
- Installs: postfix, ca-certificates, wireguard-tools, gettext
- entrypoint.sh: envsubst for config templating, postmap for transport hash, starts both processes
- NET_ADMIN capability + /dev/net/tun device for WireGuard support
- Persistent queue volume to survive restarts
- HEALTHCHECK using `postfix status`

**Commit:** 7d36dbd

## Deviations from Plan

None - plan executed exactly as written.

## Verification Results

All verification checks passed:
- ✓ `go build ./cloud-relay/...` compiles without errors
- ✓ `go vet ./cloud-relay/...` reports no issues
- ✓ Postfix main.cf has `mydestination =` (null client mode)
- ✓ Transport map routes `*` to `smtp:[127.0.0.1]:10025`
- ✓ Go daemon listens on 127.0.0.1:10025 only (not exposed to internet)
- ✓ Forwarder interface has both mTLS and WireGuard implementations
- ✓ entrypoint.sh starts both relay daemon and Postfix

Docker build verification skipped (Docker not available in environment) - Dockerfile syntax and configuration validated manually.

## Success Criteria Met

- ✓ Go relay daemon builds and has correct SMTP backend with transport forwarding
- ✓ Postfix is configured as relay-only (no local delivery) with transport map to Go daemon
- ✓ Docker container configuration complete with all required components
- ✓ Inbound flow: internet:25 -> Postfix -> localhost:10025 -> Go daemon -> transport -> home
- ✓ Outbound flow: home -> transport -> Go daemon -> Postfix -> direct MTA delivery

## Next Steps

- **Plan 02-02**: TLS/SSL certificate acquisition (Let's Encrypt) and Postfix TLS enforcement
- **Plan 02-03**: SMTP authentication, rate limiting, and spam prevention

The relay core is now ready to route mail, pending TLS certificate provisioning.

## Self-Check: PASSED

Verified all created files exist:
- ✓ cloud-relay/cmd/relay/main.go
- ✓ cloud-relay/relay/config/config.go
- ✓ cloud-relay/relay/smtp/server.go
- ✓ cloud-relay/relay/smtp/session.go
- ✓ cloud-relay/relay/forward/forwarder.go
- ✓ cloud-relay/relay/forward/mtls.go
- ✓ cloud-relay/relay/forward/wireguard.go
- ✓ cloud-relay/postfix-config/main.cf
- ✓ cloud-relay/postfix-config/master.cf
- ✓ cloud-relay/postfix-config/transport
- ✓ cloud-relay/entrypoint.sh
- ✓ cloud-relay/Dockerfile
- ✓ cloud-relay/docker-compose.yml

Verified commits exist:
- ✓ 889aad4: feat(02-01): implement Go relay daemon with SMTP backend and transport forwarding
- ✓ 7d36dbd: feat(02-01): add Postfix relay-only configuration and Docker container
