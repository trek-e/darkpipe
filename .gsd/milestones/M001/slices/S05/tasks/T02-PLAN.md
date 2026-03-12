# T02: 05-queue-offline-handling 02

**Slice:** S05 — **Milestone:** M001

## Description

Add S3-compatible overflow storage for the message queue and create the phase integration test suite.

Purpose: Fulfills QUEUE-02 (overflow to Storj/S3 when RAM queue full). Also creates the end-of-phase test suite per project workflow rules ("Create test suite at end of each phase"). The integration tests validate all three phase success criteria.

Output: overflow.go (MinIO SDK S3 client), queue.go modified with overflow integration, updated config and main.go, docker-compose volume for queue data, integration tests proving queue-on-offline + auto-delivery + disabled-bounce behaviors.

## Must-Haves

- [ ] "When the cloud relay RAM queue exceeds its threshold, overflow messages store encrypted in S3-compatible storage"
- [ ] "Overflow messages are retrieved and delivered when the home device reconnects"
- [ ] "S3 overflow is optional — queue works without it (messages rejected when RAM full if overflow disabled)"
- [ ] "All messages in S3 are age-encrypted (encryption happens before upload, not via S3 server-side encryption)"
- [ ] "Phase integration test validates: queue-on-offline, auto-delivery-on-reconnect, and queue-disabled-bounce behaviors"

## Files

- `cloud-relay/relay/queue/overflow.go`
- `cloud-relay/relay/queue/overflow_test.go`
- `cloud-relay/relay/queue/queue.go`
- `cloud-relay/relay/config/config.go`
- `cloud-relay/relay/config/config_test.go`
- `cloud-relay/cmd/relay/main.go`
- `cloud-relay/docker-compose.yml`
- `cloud-relay/tests/integration_test.go`
- `go.mod`
- `go.sum`
