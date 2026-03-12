# T01: 02-cloud-relay 01

**Slice:** S02 — **Milestone:** M001

## Description

Build the core cloud relay: a Postfix relay-only container with a Go SMTP relay daemon that bridges internet-facing SMTP to the home device via WireGuard or mTLS transport.

Purpose: This is the foundational plan for Phase 2. It establishes the inbound mail flow (internet -> Postfix -> Go daemon -> transport -> home device) and outbound mail flow (home device -> transport -> Postfix -> direct MTA delivery). Without this, no mail flows through the system.

Output: A working Docker container running Postfix in relay-only mode and a Go relay daemon that forwards mail to the home device via Phase 1's transport layer.

## Must-Haves

- [ ] "Postfix accepts inbound SMTP on port 25 and forwards all mail to Go relay daemon on localhost:10025"
- [ ] "Go relay daemon receives SMTP envelope from Postfix and forwards message data through WireGuard/mTLS transport to home device"
- [ ] "Outbound SMTP from home device routes through Postfix for direct MTA delivery to destination mail servers"
- [ ] "Container builds and runs with Postfix + Go relay daemon + WireGuard tools on Alpine"

## Files

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
