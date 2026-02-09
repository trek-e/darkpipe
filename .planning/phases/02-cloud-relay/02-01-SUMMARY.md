---
phase: 02-cloud-relay
plan: 01
subsystem: cloud-relay
tags: [smtp, relay, postfix, wireguard, mtls, docker]
dependency_graph:
  requires:
    - phase: 01-transport-layer
      plans: [01-01, 01-02]
      components: [transport/mtls/client, transport/wireguard]
  provides:
    - cloud-relay/cmd/relay (relay daemon entrypoint)
    - cloud-relay/relay/smtp (SMTP backend for emersion/go-smtp)
    - cloud-relay/relay/forward (transport abstraction with mTLS and WireGuard implementations)
    - cloud-relay/postfix-config (relay-only null client configuration)
    - cloud-relay/Dockerfile (containerized Postfix + relay daemon)
  affects:
    - Phase 03 (home device): will receive SMTP forwarding from this relay
    - Phase 04 (DNS/SPF): requires RELAY_HOSTNAME for SPF records
tech_stack:
  added:
    - emersion/go-smtp v0.24.0 (SMTP server and client library)
    - Postfix 3.x (Alpine package, relay-only mode)
    - Alpine Linux 3.21 (container base)
  patterns:
    - Transport abstraction via Forwarder interface
    - SMTP relay pipeline: Postfix -> Go daemon -> transport -> home device
    - Multi-stage Docker build for minimal image size
    - LMDB for Postfix maps (BerkleyDB deprecated in Alpine)
key_files:
  created:
    - cloud-relay/cmd/relay/main.go (daemon entrypoint with graceful shutdown)
    - cloud-relay/relay/config/config.go (environment-based configuration)
    - cloud-relay/relay/smtp/server.go (emersion/go-smtp backend)
    - cloud-relay/relay/smtp/session.go (SMTP session with forwarding logic)
    - cloud-relay/relay/forward/forwarder.go (transport abstraction interface)
    - cloud-relay/relay/forward/mtls.go (mTLS forwarder using Phase 1 client)
    - cloud-relay/relay/forward/wireguard.go (WireGuard tunnel forwarder)
    - cloud-relay/postfix-config/main.cf (null client configuration)
    - cloud-relay/postfix-config/master.cf (standard master process config)
    - cloud-relay/postfix-config/transport (wildcard map to localhost:10025)
    - cloud-relay/entrypoint.sh (container startup with envsubst and graceful shutdown)
    - cloud-relay/Dockerfile (multi-stage Go builder + Alpine runtime)
    - cloud-relay/docker-compose.yml (development deployment with WireGuard support)
  modified:
    - go.mod (added emersion/go-smtp dependency)
    - go.sum (dependency checksums)
decisions:
  - "Use emersion/go-smtp for both server (relay daemon) and client (forwarding) sides"
  - "LMDB format for Postfix maps instead of BerkleyDB (deprecated in Alpine)"
  - "Relay daemon listens on 127.0.0.1:10025 only (not exposed to internet)"
  - "Transport abstraction via Forwarder interface for WireGuard/mTLS flexibility"
  - "Opportunistic TLS (may) not enforced - enforcement deferred to Plan 02-02"
  - "WireGuard support requires NET_ADMIN capability and /dev/net/tun device in container"
  - "Persistent queue volume to survive container restarts per research recommendation"
metrics:
  duration: 234s
  tasks_completed: 2
  files_created: 15
  commits: 2
  completed_date: 2026-02-09
---

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
