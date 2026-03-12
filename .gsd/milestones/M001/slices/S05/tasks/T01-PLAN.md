# T01: 05-queue-offline-handling 01

**Slice:** S05 — **Milestone:** M001

## Description

Build the encrypted in-memory message queue and QueuedForwarder wrapper that intercepts forwarding failures and queues messages encrypted with age when the home device is offline.

Purpose: Fulfills QUEUE-01 (encrypted queue on cloud relay) and QUEUE-03 (disable queuing to bounce). This is the core offline handling capability that lets users receive mail even when their home device has a temporary internet outage.

Output: queue package (encrypt, queue, processor), QueuedForwarder in forward package, config extensions, main.go integration. All queued messages encrypted at rest with age. Background processor auto-delivers when home device reconnects.

## Must-Haves

- [ ] "When home device is offline and queuing is enabled, inbound mail is accepted (250 OK to sender) and encrypted in RAM on the cloud relay"
- [ ] "When home device reconnects, queued messages are automatically decrypted and delivered without manual intervention"
- [ ] "When queuing is disabled and home device is offline, the relay returns a 4xx temporary failure causing the sender's server to retry"
- [ ] "Queued messages are encrypted at rest using age even while held in memory"
- [ ] "Duplicate messages (same Message-ID) are deduplicated in the queue"
- [ ] "Queue processor delivers messages in batches with rate limiting to avoid overwhelming the home device on reconnection"

## Files

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
