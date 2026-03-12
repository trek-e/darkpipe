# S05: Queue Offline Handling

**Goal:** Build the encrypted in-memory message queue and QueuedForwarder wrapper that intercepts forwarding failures and queues messages encrypted with age when the home device is offline.
**Demo:** Build the encrypted in-memory message queue and QueuedForwarder wrapper that intercepts forwarding failures and queues messages encrypted with age when the home device is offline.

## Must-Haves


## Tasks

- [x] **T01: 05-queue-offline-handling 01** `est:9min 27s`
  - Build the encrypted in-memory message queue and QueuedForwarder wrapper that intercepts forwarding failures and queues messages encrypted with age when the home device is offline.

Purpose: Fulfills QUEUE-01 (encrypted queue on cloud relay) and QUEUE-03 (disable queuing to bounce). This is the core offline handling capability that lets users receive mail even when their home device has a temporary internet outage.

Output: queue package (encrypt, queue, processor), QueuedForwarder in forward package, config extensions, main.go integration. All queued messages encrypted at rest with age. Background processor auto-delivers when home device reconnects.
- [x] **T02: 05-queue-offline-handling 02**
  - Add S3-compatible overflow storage for the message queue and create the phase integration test suite.

Purpose: Fulfills QUEUE-02 (overflow to Storj/S3 when RAM queue full). Also creates the end-of-phase test suite per project workflow rules ("Create test suite at end of each phase"). The integration tests validate all three phase success criteria.

Output: overflow.go (MinIO SDK S3 client), queue.go modified with overflow integration, updated config and main.go, docker-compose volume for queue data, integration tests proving queue-on-offline + auto-delivery + disabled-bounce behaviors.

## Files Likely Touched

- `cloud-relay/relay/queue/encrypt.go`
- `cloud-relay/relay/queue/encrypt_test.go`
- `cloud-relay/relay/queue/queue.go`
- `cloud-relay/relay/queue/queue_test.go`
- `cloud-relay/relay/queue/processor.go`
- `cloud-relay/relay/queue/processor_test.go`
- `cloud-relay/relay/forward/queued.go`
- `cloud-relay/relay/forward/queued_test.go`
- `cloud-relay/relay/config/config.go`
- `cloud-relay/relay/config/config_test.go`
- `cloud-relay/cmd/relay/main.go`
- `go.mod`
- `go.sum`
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
