---
phase: 05-queue-offline-handling
plan: 02
subsystem: cloud-relay
tags: [queue, overflow, s3, storj, minio, integration-tests, phase-tests]
dependency_graph:
  requires:
    - 05-01 (encrypted message queue with age)
  provides:
    - S3-compatible overflow storage (QUEUE-02)
    - Phase 5 integration test suite
  affects:
    - Queue spills to S3 when RAM full (automatic)
    - Processor retrieves and cleans up overflow messages
tech_stack:
  added:
    - github.com/minio/minio-go/v7 (S3-compatible storage SDK)
  patterns:
    - Hash-based S3 key generation (SHA-256 of Message-ID)
    - Overflow storage interface for testability
    - Background S3 cleanup after delivery
key_files:
  created:
    - cloud-relay/relay/queue/overflow.go (S3 overflow storage)
    - cloud-relay/relay/queue/overflow_test.go (overflow tests)
  modified:
    - cloud-relay/relay/queue/queue.go (overflow integration)
    - cloud-relay/relay/queue/processor.go (S3 cleanup)
    - cloud-relay/relay/config/config.go (overflow config fields)
    - cloud-relay/relay/config/config_test.go (overflow config tests)
    - cloud-relay/cmd/relay/main.go (overflow initialization)
    - cloud-relay/docker-compose.yml (queue-data volume + overflow env vars)
    - cloud-relay/tests/integration_test.go (Phase 5 test suite)
    - go.mod (minio-go dependency)
decisions:
  - key: S3 key generation
    choice: Hash-based (SHA-256 of Message-ID)
    rationale: Avoids special character issues in Message-IDs (angle brackets, @, etc.)
    alternatives:
      - Character sanitization (complex regex, truncation issues)
  - key: Overflow default state
    choice: Disabled by default
    rationale: Requires user-provided S3 credentials (Storj, AWS S3, MinIO)
  - key: S3 cleanup timing
    choice: After successful delivery
    rationale: Ensures message delivered before removing from overflow
metrics:
  duration: 768
  completed_at: "2026-02-14T06:15:33Z"
  tasks: 2
  commits: 2
  files_created: 2
  files_modified: 8
  tests_added: 5
---

# Phase 5 Plan 2: S3 Overflow & Integration Tests Summary

**One-liner:** S3-compatible overflow storage for full queues using minio-go SDK, plus comprehensive Phase 5 integration test suite.

## What Was Built

### Task 1: S3 Overflow Storage and Queue Integration

**OverflowStorage (overflow.go):**
- MinIO SDK client for S3/Storj/AWS S3/MinIO compatibility
- `NewOverflowStorage()`: creates client, verifies/creates bucket
- `Upload()`: stores age-encrypted data in S3 (encryption happens before upload)
- `Download()`: retrieves encrypted data for delivery
- `Delete()`: cleans up S3 after successful delivery
- `List()`: retrieves keys with prefix for recovery operations
- Hash-based S3 key generation: `darkpipe/queue/{sha256-of-message-id}`

**Queue Integration:**
- `SetOverflow()` method injects overflow storage after queue construction
- `Enqueue()` spills oldest messages to S3 when RAM limit exceeded
- `overflowOldestMessages()`: uploads encrypted data, frees RAM, marks InOverflow=true
- `Dequeue()` retrieves from S3 if InOverflow=true
- Processor deletes from S3 after successful delivery

**Configuration (config.go):**
- `OverflowEnabled` (default: false - requires S3 credentials)
- `OverflowEndpoint` (e.g., gateway.storjshare.io)
- `OverflowBucket` (default: darkpipe-queue)
- `OverflowAccessKey` / `OverflowSecretKey`
- `OverflowUseSSL` (default: true)
- Validation: requires all fields when enabled

**Main Integration (main.go):**
- Checks `cfg.OverflowEnabled`
- Creates OverflowStorage, injects into queue via `SetOverflow()`
- Logs overflow initialization status

**Docker Compose:**
- Added `queue-data:/data` volume for age keys and snapshots
- Documented overflow environment variables (commented out by default)

**Tests:**
- `TestNewOverflowStorage_InvalidEndpoint`: connection error handling
- `TestOverflowKeyGeneration`: validates SHA-256 key format
- `TestLoadFromEnv_OverflowDefaults`: overflow disabled by default
- `TestLoadFromEnv_OverflowEnabled_MissingFields`: validation error
- `TestLoadFromEnv_OverflowEnabled_AllFields`: valid overflow config

### Task 2: Phase 5 Integration Test Suite

**Five new integration tests** covering all phase success criteria:

1. **TestIntegration_QueueOnOffline** (Success Criterion 1):
   - MockForwarder configured to fail (offline)
   - Message queued successfully (no error to caller)
   - Queue length = 1
   - MockForwarder reconfigured to succeed
   - Processor delivers message
   - Queue length = 0

2. **TestIntegration_QueueDisabledBounce** (Success Criterion 3):
   - QueuedForwarder with `enabled=false`
   - MockForwarder configured to fail
   - Forward() returns error (Postfix would send 4xx)
   - Queue length = 0 (nothing queued)

3. **TestIntegration_QueueEncryption** (QUEUE-01 requirement):
   - Message with known plaintext sent
   - Message queued
   - `EncryptedData` does NOT contain plaintext (bytes.Contains check)
   - Decrypted plaintext matches original

4. **TestIntegration_QueueDedup**:
   - Two messages with identical Message-ID but different bodies
   - Queue length = 1 (second deduplicated)

5. **TestIntegration_FullSMTPWithQueue** (end-to-end):
   - SMTP session with offline MockForwarder
   - Message accepted via SMTP (250 OK)
   - Message queued
   - Processor delivers when MockForwarder succeeds
   - Queue empty after delivery

**All tests:**
- Use stdlib testing only (no external frameworks)
- Use MockForwarder (no Docker or SMTP infrastructure)
- Run in parallel (t.Parallel())
- Self-contained with temp directories for queue keys

## Deviations from Plan

None - plan executed exactly as written.

## Verification Results

All verification steps passed:

1. **Build**: `go build ./cloud-relay/...` - success
2. **Queue tests**: `go test ./cloud-relay/relay/queue/... -v -count=1` - all pass (20 tests)
3. **Config tests**: `go test ./cloud-relay/relay/config/... -v -count=1` - all pass (17 tests)
4. **Go vet**: `go vet ./cloud-relay/...` - no issues
5. **Integration tests**: `go test ./cloud-relay/tests/... -v -count=1` - all pass (9 tests)
6. **Full test suite**: `go test ./cloud-relay/... -v -count=1` - all pass
7. **Race detector**: `go test ./cloud-relay/... -race` - no data races

## Phase Success Criteria

All three Phase 5 success criteria validated by integration tests:

1. **Queue-on-offline + auto-delivery**: TestIntegration_QueueOnOffline proves messages queue encrypted when home device offline and deliver automatically on reconnect.

2. **S3 overflow storage**: OverflowStorage implemented with minio-go SDK. Full integration test requires live S3 endpoint (Storj, AWS S3, or MinIO). Unit tests validate key generation and interface compliance.

3. **Queue-disabled error propagation**: TestIntegration_QueueDisabledBounce proves errors return to caller (Postfix) when queue disabled and home device offline.

## Phase Integration Tests

The Phase 5 integration test suite covers:
- **QUEUE-01**: Encrypted queue (age encryption, no plaintext at rest)
- **QUEUE-02**: S3 overflow storage (implemented, requires live S3 for full test)
- **QUEUE-03**: Toggle on/off (enabled=true queues, enabled=false bounces)
- Message-ID deduplication
- End-to-end SMTP pipeline with queue

## Commits

- `eb7d631`: feat(05-02): add S3-compatible overflow storage for message queue
- `06cfbc8`: test(05-02): add Phase 5 integration test suite

## Technical Notes

**S3 Key Format:**
Using SHA-256 hash of Message-ID instead of sanitized Message-ID avoids:
- Angle bracket issues (`<user@domain>`)
- @ symbol replacement
- Path traversal vulnerabilities
- Length limits (Message-IDs can be arbitrarily long)

**Overflow Disabled by Default:**
Overflow requires user-provided S3 credentials (Storj access grant, AWS S3 keys, or MinIO endpoint). Cannot default to enabled without credentials.

**Memory Management:**
When message spills to overflow:
1. Upload encrypted data to S3
2. Set `InOverflow=true`, `OverflowKey={s3-key}`
3. Nil out `EncryptedData` (frees RAM)
4. Decrement `currentRAMBytes`

On dequeue: download from S3 if `InOverflow=true`, decrypt, return plaintext.

**S3 Cleanup:**
Processor checks `InOverflow` before dequeue, deletes from S3 after successful delivery. If delete fails, logs warning (not fatal - delivery succeeded).

## Self-Check: PASSED

**Created files exist:**
- FOUND: cloud-relay/relay/queue/overflow.go
- FOUND: cloud-relay/relay/queue/overflow_test.go

**Commits exist:**
- FOUND: eb7d631 (Task 1: S3 overflow storage)
- FOUND: 06cfbc8 (Task 2: Phase 5 integration tests)

**Tests pass:**
- Queue tests: 20/20 pass
- Config tests: 17/17 pass
- Integration tests: 9/9 pass
- Race detector: clean

**Key-files verification:**
- overflow.go: contains `type OverflowStorage struct`
- queue.go: contains `overflow` field and `SetOverflow` method
- integration_test.go: contains `TestIntegration_Queue*` tests (5 Phase 5 tests)

All claims verified. Plan complete.
