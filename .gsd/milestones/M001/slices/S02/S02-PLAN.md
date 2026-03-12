# S02: Cloud Relay

**Goal:** Build the core cloud relay: a Postfix relay-only container with a Go SMTP relay daemon that bridges internet-facing SMTP to the home device via WireGuard or mTLS transport.
**Demo:** Build the core cloud relay: a Postfix relay-only container with a Go SMTP relay daemon that bridges internet-facing SMTP to the home device via WireGuard or mTLS transport.

## Must-Haves


## Tasks

- [x] **T01: 02-cloud-relay 01**
  - Build the core cloud relay: a Postfix relay-only container with a Go SMTP relay daemon that bridges internet-facing SMTP to the home device via WireGuard or mTLS transport.

Purpose: This is the foundational plan for Phase 2. It establishes the inbound mail flow (internet -> Postfix -> Go daemon -> transport -> home device) and outbound mail flow (home device -> transport -> Postfix -> direct MTA delivery). Without this, no mail flows through the system.

Output: A working Docker container running Postfix in relay-only mode and a Go relay daemon that forwards mail to the home device via Phase 1's transport layer.
- [x] **T02: 02-cloud-relay 02**
  - Add TLS enforcement, strict mode, notification system, and Let's Encrypt certificate automation to the cloud relay.

Purpose: This plan fulfills RELAY-04 (TLS enforced on all connections), RELAY-05 (optional strict mode to refuse plaintext-only peers), RELAY-06 (user notified when remote server lacks TLS), and CERT-01 (Let's Encrypt certificates for public-facing TLS). Without TLS, the relay cannot participate in modern email delivery where major providers (Gmail, Outlook) expect encrypted connections.

Output: TLS-enabled Postfix with automated Let's Encrypt certificates, configurable strict mode, and real-time TLS quality monitoring with webhook notifications.
- [x] **T03: 02-cloud-relay 03**
  - Verify ephemeral storage guarantees, optimize container image size, and build the comprehensive test suite for the cloud relay.

Purpose: This plan closes the loop on Phase 2 by proving three critical properties: (1) no mail persists after forwarding (RELAY-02 verification), (2) the container meets the size and resource constraints (UX-02), and (3) the entire relay pipeline works end-to-end with test coverage. Per project memory rules, a test suite is required at the end of each phase.

Output: Ephemeral storage verification tool, optimized Dockerfile under 50MB, and comprehensive test suite covering all cloud-relay packages plus integration tests.

## Files Likely Touched

- `cloud-relay/cmd/relay/main.go`
- `cloud-relay/relay/smtp/server.go`
- `cloud-relay/relay/smtp/session.go`
- `cloud-relay/relay/forward/forwarder.go`
- `cloud-relay/relay/forward/mtls.go`
- `cloud-relay/relay/forward/wireguard.go`
- `cloud-relay/relay/config/config.go`
- `cloud-relay/postfix-config/main.cf`
- `cloud-relay/postfix-config/master.cf`
- `cloud-relay/postfix-config/transport`
- `cloud-relay/Dockerfile`
- `cloud-relay/docker-compose.yml`
- `cloud-relay/entrypoint.sh`
- `go.mod`
- `go.sum`
- `cloud-relay/relay/tls/monitor.go`
- `cloud-relay/relay/tls/monitor_test.go`
- `cloud-relay/relay/tls/strict.go`
- `cloud-relay/relay/tls/strict_test.go`
- `cloud-relay/relay/notify/notifier.go`
- `cloud-relay/relay/notify/notifier_test.go`
- `cloud-relay/relay/notify/webhook.go`
- `cloud-relay/certbot/docker-compose.certbot.yml`
- `cloud-relay/certbot/renew-hook.sh`
- `cloud-relay/postfix-config/main.cf`
- `cloud-relay/docker-compose.yml`
- `cloud-relay/relay/config/config.go`
- `cloud-relay/cmd/relay/main.go`
- `cloud-relay/relay/ephemeral/verify.go`
- `cloud-relay/relay/ephemeral/verify_test.go`
- `cloud-relay/Dockerfile`
- `cloud-relay/docker-compose.yml`
- `cloud-relay/relay/smtp/server_test.go`
- `cloud-relay/relay/smtp/session_test.go`
- `cloud-relay/relay/forward/forwarder_test.go`
- `cloud-relay/relay/forward/mtls_test.go`
- `cloud-relay/relay/forward/wireguard_test.go`
- `cloud-relay/relay/config/config_test.go`
- `cloud-relay/tests/integration_test.go`
