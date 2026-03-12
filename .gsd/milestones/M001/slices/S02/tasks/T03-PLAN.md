# T03: 02-cloud-relay 03

**Slice:** S02 — **Milestone:** M001

## Description

Verify ephemeral storage guarantees, optimize container image size, and build the comprehensive test suite for the cloud relay.

Purpose: This plan closes the loop on Phase 2 by proving three critical properties: (1) no mail persists after forwarding (RELAY-02 verification), (2) the container meets the size and resource constraints (UX-02), and (3) the entire relay pipeline works end-to-end with test coverage. Per project memory rules, a test suite is required at the end of each phase.

Output: Ephemeral storage verification tool, optimized Dockerfile under 50MB, and comprehensive test suite covering all cloud-relay packages plus integration tests.

## Must-Haves

- [ ] "No mail content persists on the cloud relay filesystem after successful forwarding"
- [ ] "Container image is under 50MB"
- [ ] "Relay runs with less than 256MB RAM on a $5/month VPS"
- [ ] "All Go packages in cloud-relay/ have comprehensive test coverage"
- [ ] "End-to-end SMTP flow works: SMTP in -> Postfix -> Go daemon -> forwarder -> delivery"

## Files

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
