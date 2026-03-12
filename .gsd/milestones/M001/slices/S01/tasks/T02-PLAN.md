# T02: 01-transport-layer 02

**Slice:** S01 — **Milestone:** M001

## Description

Establish the mTLS alternative transport and internal PKI: step-ca deployment as private CA, mTLS server (cloud relay) and client (home device) with mutual certificate verification, persistent connection with exponential backoff reconnection, and automated certificate renewal via systemd timers.

Purpose: mTLS provides an application-level alternative to WireGuard for users who prefer it or cannot use kernel WireGuard. The internal CA (step-ca) issues short-lived certificates distinct from public-facing TLS, supporting both WireGuard cert needs and mTLS transport with automated renewal.

Output: Go packages for mTLS server/client with auto-reconnect, step-ca setup tooling, certificate renewal automation via systemd, deployment script for CA initialization.

## Must-Haves

- [ ] "Internal transport certificates are issued by a private CA (step-ca) and are distinct from public-facing TLS certificates"
- [ ] "User can select mTLS as an alternative transport and traffic flows encrypted between cloud and home"
- [ ] "Certificate renewal is automated via systemd timers with randomized jitter"

## Files

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
