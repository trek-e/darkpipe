# Phase 5: Queue & Offline Handling - Research

**Researched:** 2026-02-14
**Domain:** Encrypted message queuing with S3-compatible overflow storage
**Confidence:** HIGH

## Summary

Phase 5 adds encrypted message queuing to the cloud relay when the home device is offline. The solution uses age encryption for at-rest message protection, an in-memory RAM queue with optional S3-compatible overflow via MinIO Go SDK (supporting Storj, AWS S3, MinIO), and Postfix content filter integration for queue control. The existing relay architecture (emersion/go-smtp forwarder abstraction) requires minimal modification—a new QueuedForwarder wrapper that attempts immediate delivery, falls back to encrypted queue on transport failure, and delivers queued messages when the home device reconnects.

**Primary recommendation:** Use filippo.io/age for encryption (stdlib dependencies only, streaming API, battle-tested), github.com/minio/minio-go/v7 for S3 operations, and stdlib hash/crc32 for corruption detection. Queue state persists in memory with periodic disk snapshots (JSON) for crash recovery. Avoid custom encryption or queue persistence—use proven libraries.

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| filippo.io/age | v1.2.0+ | Encryption for queued messages | Simple API, stdlib-only deps, ChaCha20-Poly1305 AEAD, streaming support, battle-tested, X25519 key exchange |
| github.com/minio/minio-go/v7 | v7.0.80+ | S3-compatible object storage client | Official MinIO SDK, works with Storj/AWS S3/MinIO, server-side encryption support, streaming uploads, 288 code examples in Context7 |
| github.com/cenkalti/backoff/v4 | v4.3.0+ | Exponential backoff for retries | Already used in Phase 1 mTLS reconnection, proven in production, context-aware, configurable |
| github.com/emersion/go-message | Latest | RFC 5322 message parsing for Message-ID dedup | Same author as go-smtp (consistency), streaming API, RFC-compliant header parsing, handles Message-ID extraction |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| hash/crc32 | stdlib | Fast corruption detection for queue entries | Non-cryptographic integrity check, extremely fast, sufficient for accidental corruption (not malicious tampering) |
| encoding/json | stdlib | Queue metadata persistence for crash recovery | Periodic snapshots of queue state (not messages—those are encrypted files), human-readable for debugging |
| net/mail | stdlib | Message-ID parsing fallback | If go-message unavailable, stdlib provides basic RFC 5322 parsing (less feature-complete) |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| age | NaCl (golang.org/x/crypto/nacl) | NaCl box requires managing nonces, age handles this automatically; age has simpler key format (AGE-SECRET-KEY-...) |
| age | ChaCha20-Poly1305 directly | Would need to implement key derivation, nonce management, and streaming; age provides this for free |
| In-memory queue | LevelDB-backed queue (goque, dque) | On-disk queues add complexity, require cleanup on crash, conflict with 256MB RAM limit; ephemeral RAM queue matches relay design |
| minio-go | AWS SDK v2 | MinIO SDK supports all S3-compatible providers, simpler API, better Storj documentation |
| Periodic JSON snapshots | SQLite queue metadata | SQLite adds dependency, overkill for simple queue state; JSON is human-readable for debugging |

**Installation:**
```bash
go get filippo.io/age@latest
go get github.com/minio/minio-go/v7@latest
go get github.com/emersion/go-message@latest
# cenkalti/backoff/v4 already in go.mod from Phase 1
```

## Architecture Patterns

### Recommended Project Structure

```
cloud-relay/
├── relay/
│   ├── forward/
│   │   ├── forwarder.go      # Existing Forwarder interface
│   │   ├── queued.go          # NEW: QueuedForwarder wrapper
│   │   └── queued_test.go     # NEW: Tests for queue behavior
│   ├── queue/
│   │   ├── queue.go           # NEW: In-memory message queue
│   │   ├── queue_test.go      # NEW: Queue operations tests
│   │   ├── encrypt.go         # NEW: age encryption/decryption
│   │   ├── encrypt_test.go    # NEW: Encryption tests
│   │   ├── overflow.go        # NEW: S3 overflow storage
│   │   └── overflow_test.go   # NEW: S3 operations tests
│   └── config/
│       └── config.go          # MODIFY: Add queue config fields
```

### Pattern 1: Queued Forwarder Wrapper

**What:** A Forwarder implementation that wraps the underlying transport forwarder (WireGuardForwarder or MTLSForwarder) and adds queue-on-failure logic.

**When to use:** Always—this becomes the default forwarder when queuing is enabled. Maintains the existing Forwarder interface contract.

**Example:**
```go
// Source: Derived from existing forward/forwarder.go pattern
package forward

import (
    "context"
    "io"
    "github.com/darkpipe/darkpipe/cloud-relay/relay/queue"
)

// QueuedForwarder wraps a transport forwarder with queue-on-failure logic.
type QueuedForwarder struct {
    transport Forwarder          // Underlying WireGuard or mTLS forwarder
    queue     *queue.MessageQueue // Encrypted in-memory queue
    enabled   bool                // User can disable queuing (QUEUE-03)
}

func NewQueuedForwarder(transport Forwarder, q *queue.MessageQueue, enabled bool) *QueuedForwarder {
    return &QueuedForwarder{
        transport: transport,
        queue:     q,
        enabled:   enabled,
    }
}

func (f *QueuedForwarder) Forward(ctx context.Context, from string, to []string, data io.Reader) error {
    // Attempt immediate forwarding
    err := f.transport.Forward(ctx, from, to, data)
    if err == nil {
        return nil // Success, no queuing needed
    }

    // If queuing disabled, return error immediately (sender retries)
    if !f.enabled {
        return err
    }

    // Transport failed, queue encrypted message
    return f.queue.Enqueue(ctx, from, to, data)
}

func (f *QueuedForwarder) Close() error {
    // Close transport and queue
    if err := f.transport.Close(); err != nil {
        return err
    }
    return f.queue.Close()
}
```

### Pattern 2: Encrypted Message Queue with Overflow

**What:** In-memory queue storing encrypted messages with automatic S3 overflow when RAM threshold exceeded.

**When to use:** Core queue implementation. All queued messages encrypted at rest, even in RAM.

**Example:**
```go
// Source: Inspired by Context7 minio-go and age examples
package queue

import (
    "bytes"
    "context"
    "crypto/rand"
    "encoding/json"
    "filippo.io/age"
    "io"
    "sync"
    "time"
)

type QueuedMessage struct {
    ID          string    // Message-ID from email header (dedup key)
    From        string    // SMTP envelope sender
    To          []string  // SMTP envelope recipients
    EnqueuedAt  time.Time // When message entered queue
    EncryptedData []byte  // age-encrypted message body
    Checksum    uint32    // CRC32 for corruption detection
    InOverflow  bool      // True if stored in S3, false if in RAM
    OverflowKey string    // S3 object key if InOverflow=true
}

type MessageQueue struct {
    mu            sync.RWMutex
    messages      map[string]*QueuedMessage // Key: Message-ID
    recipient     age.Recipient             // Public key for encryption
    identity      age.Identity              // Private key for decryption
    maxRAMBytes   int64                     // Trigger overflow to S3
    currentRAM    int64                     // Current RAM usage estimate
    overflow      *OverflowStorage          // S3-compatible storage (nil if disabled)
}

func (q *MessageQueue) Enqueue(ctx context.Context, from string, to []string, data io.Reader) error {
    q.mu.Lock()
    defer q.mu.Unlock()

    // Read and parse message to extract Message-ID for dedup
    msgData, msgID, err := q.readAndParseMessage(data)
    if err != nil {
        return err
    }

    // Check for duplicate
    if _, exists := q.messages[msgID]; exists {
        return nil // Deduplicated (sender may retry same message)
    }

    // Encrypt message
    encrypted, checksum, err := q.encryptMessage(msgData)
    if err != nil {
        return err
    }

    msg := &QueuedMessage{
        ID:            msgID,
        From:          from,
        To:            to,
        EnqueuedAt:    time.Now(),
        EncryptedData: encrypted,
        Checksum:      checksum,
        InOverflow:    false,
    }

    // Check if we need to overflow to S3
    msgSize := int64(len(encrypted))
    if q.overflow != nil && (q.currentRAM + msgSize) > q.maxRAMBytes {
        // Move oldest messages to S3 until we have space
        if err := q.overflowOldestMessages(ctx, msgSize); err != nil {
            return err
        }
    }

    q.messages[msgID] = msg
    q.currentRAM += msgSize
    return nil
}

func (q *MessageQueue) encryptMessage(data []byte) (encrypted []byte, checksum uint32, err error) {
    // Encrypt with age (streaming)
    var buf bytes.Buffer
    w, err := age.Encrypt(&buf, q.recipient)
    if err != nil {
        return nil, 0, err
    }
    if _, err := w.Write(data); err != nil {
        return nil, 0, err
    }
    if err := w.Close(); err != nil {
        return nil, 0, err
    }

    encrypted = buf.Bytes()
    checksum = crc32.ChecksumIEEE(encrypted) // Fast corruption detection
    return encrypted, checksum, nil
}
```

### Pattern 3: Background Queue Processor

**What:** Goroutine that periodically attempts delivery of queued messages when home device is reachable.

**When to use:** Started when relay daemon initializes, runs until shutdown.

**Example:**
```go
// Source: Adapted from Phase 1 reconnection pattern with cenkalti/backoff
package queue

import (
    "context"
    "log"
    "time"
    "github.com/cenkalti/backoff/v4"
)

func (q *MessageQueue) StartProcessor(ctx context.Context, forwarder forward.Forwarder) {
    ticker := time.NewTicker(30 * time.Second) // Check every 30s
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            q.processQueue(ctx, forwarder)
        }
    }
}

func (q *MessageQueue) processQueue(ctx context.Context, forwarder forward.Forwarder) {
    q.mu.Lock()
    pending := make([]*QueuedMessage, 0, len(q.messages))
    for _, msg := range q.messages {
        pending = append(pending, msg)
    }
    q.mu.Unlock()

    if len(pending) == 0 {
        return
    }

    log.Printf("Queue processor: attempting delivery of %d messages", len(pending))

    for _, msg := range pending {
        // Decrypt message
        plaintext, err := q.decryptMessage(msg.EncryptedData, msg.Checksum)
        if err != nil {
            log.Printf("ERROR: decrypt failed for %s: %v", msg.ID, err)
            continue
        }

        // Attempt delivery with exponential backoff
        bo := backoff.NewExponentialBackOff()
        bo.MaxElapsedTime = 5 * time.Minute
        err = backoff.Retry(func() error {
            return forwarder.Forward(ctx, msg.From, msg.To, bytes.NewReader(plaintext))
        }, backoff.WithContext(bo, ctx))

        if err == nil {
            // Success, remove from queue
            q.mu.Lock()
            delete(q.messages, msg.ID)
            q.currentRAM -= int64(len(msg.EncryptedData))
            q.mu.Unlock()
            log.Printf("Queue processor: delivered %s", msg.ID)
        } else {
            log.Printf("Queue processor: delivery failed for %s: %v", msg.ID, err)
        }
    }
}
```

### Pattern 4: S3 Overflow Storage

**What:** Store messages in S3-compatible object storage when RAM queue exceeds threshold.

**When to use:** Optional feature (QUEUE-02). Configured via environment variables.

**Example:**
```go
// Source: Context7 minio-go examples + Storj best practices
package queue

import (
    "bytes"
    "context"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

type OverflowStorage struct {
    client     *minio.Client
    bucketName string
}

func NewOverflowStorage(endpoint, accessKey, secretKey, bucket string) (*OverflowStorage, error) {
    client, err := minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: true, // Always use HTTPS for Storj/S3
    })
    if err != nil {
        return nil, err
    }

    return &OverflowStorage{
        client:     client,
        bucketName: bucket,
    }, nil
}

func (s *OverflowStorage) Upload(ctx context.Context, key string, data []byte) error {
    // Upload already-encrypted message to S3
    // Note: Message is already age-encrypted, S3 server-side encryption is optional additional layer
    _, err := s.client.PutObject(ctx, s.bucketName, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
        ContentType: "application/octet-stream",
    })
    return err
}

func (s *OverflowStorage) Download(ctx context.Context, key string) ([]byte, error) {
    obj, err := s.client.GetObject(ctx, s.bucketName, key, minio.GetObjectOptions{})
    if err != nil {
        return nil, err
    }
    defer obj.Close()

    return io.ReadAll(obj)
}

func (s *OverflowStorage) Delete(ctx context.Context, key string) error {
    return s.client.RemoveObject(ctx, s.bucketName, key, minio.RemoveObjectOptions{})
}
```

### Pattern 5: Message-ID Extraction for Deduplication

**What:** Parse email headers to extract Message-ID for deduplication tracking.

**When to use:** Every queued message. Prevents duplicate delivery if sender retries.

**Example:**
```go
// Source: emersion/go-message documentation + RFC 5322 spec
package queue

import (
    "bytes"
    "io"
    "github.com/emersion/go-message/mail"
)

func (q *MessageQueue) readAndParseMessage(data io.Reader) (msgData []byte, msgID string, err error) {
    // Read full message into buffer
    buf := &bytes.Buffer{}
    if _, err := io.Copy(buf, data); err != nil {
        return nil, "", err
    }
    msgData = buf.Bytes()

    // Parse message headers to extract Message-ID
    mr, err := mail.CreateReader(bytes.NewReader(msgData))
    if err != nil {
        // If parsing fails, generate fallback ID (hash of message)
        return msgData, q.generateFallbackID(msgData), nil
    }

    // Extract Message-ID from headers
    msgID = mr.Header.Get("Message-ID")
    if msgID == "" {
        // No Message-ID header, generate fallback
        msgID = q.generateFallbackID(msgData)
    }

    return msgData, msgID, nil
}

func (q *MessageQueue) generateFallbackID(msgData []byte) string {
    // Generate deterministic ID from message content
    // Use SHA-256 to avoid collisions (RFC 5322 recommends unique Message-IDs)
    h := sha256.Sum256(msgData)
    return fmt.Sprintf("<%x@darkpipe.local>", h[:16]) // 128-bit hex digest
}
```

### Anti-Patterns to Avoid

- **Custom encryption:** Do NOT implement your own encryption. Use age (battle-tested, simple API, handles nonces/keys correctly).
- **Persistent queue on disk:** Conflicts with ephemeral relay design and 256MB RAM constraint. Use in-memory queue with periodic JSON snapshots for crash recovery only.
- **Blocking queue operations:** Use goroutines and context for async processing. Queue operations must not block SMTP sessions.
- **MD5 for deduplication:** RFC 5322 Message-IDs are already unique. If hashing needed, use SHA-256 (MD5 collisions are trivial).
- **Ignoring overflow threshold:** Cloud relay has 256MB memory limit (docker-compose constraint). Must overflow to S3 or reject new messages when limit reached.
- **Storing plaintext in queue:** All queued messages MUST be encrypted at rest, even in RAM (QUEUE-01 requirement).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Encryption | Custom XSalsa20-Poly1305 wrapper | filippo.io/age | Nonce reuse vulnerabilities, key derivation mistakes, streaming complexity; age handles all of this correctly |
| S3 client | HTTP requests to S3 API | github.com/minio/minio-go/v7 | S3 signature v4 auth is complex, multipart upload logic, retry handling, connection pooling; MinIO SDK is battle-tested |
| Exponential backoff | Custom sleep loop with multiplier | cenkalti/backoff/v4 (already in project) | Jitter for thundering herd, context cancellation, max elapsed time tracking; already proven in Phase 1 |
| Message parsing | Regex on email headers | emersion/go-message | RFC 5322 edge cases (folded headers, quoted-printable, MIME parts); same author as go-smtp ensures consistency |
| Queue persistence | Binary format or custom serialization | encoding/json (stdlib) for metadata only | JSON is human-readable for debugging, stdlib handles edge cases, sufficient for periodic snapshots (not hot path) |
| Corruption detection | Custom checksum | hash/crc32 (stdlib) | CRC32 is proven for accidental corruption, extremely fast, stdlib implementation optimized |

**Key insight:** Email is a complex protocol with decades of edge cases. The encryption and storage domains are security-critical. Use proven libraries to avoid vulnerabilities and subtle bugs that take months to surface.

## Common Pitfalls

### Pitfall 1: Message-ID Collisions from Poorly Configured Senders

**What goes wrong:** Some mail clients generate non-unique Message-IDs (missing FQDN, weak timestamp resolution). Deduplication may incorrectly merge distinct messages.

**Why it happens:** RFC 5322 recommends uniqueness but doesn't enforce it. Misconfigured clients may reuse IDs.

**How to avoid:** Use Message-ID as primary dedup key, but add fallback: if Message-ID missing, hash (from, to, subject, date) to generate synthetic ID. Log when fallback used for debugging.

**Warning signs:** Users report "missing" messages that were actually deduplicated. Check logs for fallback ID generation frequency.

### Pitfall 2: Memory Exhaustion from Unreachable Home Device

**What goes wrong:** If home device offline for extended period (days), queue grows unbounded until OOM killer terminates relay container.

**Why it happens:** No overflow configured, or overflow upload fails silently, RAM keeps accumulating messages.

**How to avoid:** Enforce hard RAM limit in code (not just docker-compose). When limit reached: (1) overflow to S3 if configured, (2) return 4xx SMTP code if overflow disabled/failed, causing sender to retry later.

**Warning signs:** Docker stats show relay approaching 256MB limit. Implement /metrics endpoint exposing queue depth and RAM usage.

### Pitfall 3: Thundering Herd on Reconnection

**What goes wrong:** When home device reconnects after long outage, queue processor attempts to deliver 1000+ messages simultaneously, overwhelming home device.

**Why it happens:** No rate limiting on queue processing. All messages tried in parallel.

**How to avoid:** Process queue in batches (e.g., 10 messages/minute). Use semaphore to limit concurrent deliveries. Respect SMTP 4xx responses from home device.

**Warning signs:** Home device SMTP logs show connection refused or timeouts when queue drains. Cloud relay logs show delivery failures immediately after reconnection.

### Pitfall 4: S3 Credentials Exposed in Environment Variables

**What goes wrong:** Storj/S3 credentials logged in docker-compose.yml or container inspect output, visible to anyone with Docker access.

**Why it happens:** Environment variables are not secrets. Docker stores them in plaintext.

**How to avoid:** Use Docker secrets or external secret manager (HashiCorp Vault, AWS Secrets Manager). Never commit credentials to git. Document secure credential injection in deployment guide.

**Warning signs:** `docker inspect` shows STORJ_ACCESS_KEY in Env array. Git history contains test credentials (even if later removed).

### Pitfall 5: Age Key Loss Causes Permanent Message Loss

**What goes wrong:** Cloud relay crashes, loses age private key (not persisted), queued messages unrecoverable.

**Why it happens:** Age identity generated at startup and kept in memory only.

**How to avoid:** Generate age identity on first startup, persist to /data/queue-keys/identity (mounted volume). Load on subsequent startups. Backup this key—losing it means losing queued messages permanently.

**Warning signs:** Container restart causes "decryption failed" errors for existing queue. No persistent volume mounted at /data.

### Pitfall 6: Postfix Reinjection Loop

**What goes wrong:** After queue delivers message back through Postfix (via sendmail or SMTP localhost:10026), Postfix re-queues it to cloud relay, creating infinite loop.

**Why it happens:** Postfix content_filter applies to reinjected mail. No mechanism to mark "already processed."

**How to avoid:** Reinject on separate SMTP port (e.g., localhost:10026) with receive_override_options = no_header_body_checks,no_milters. This prevents re-filtering. Document in Postfix config comments.

**Warning signs:** Postfix queue grows rapidly with same Message-ID cycling. Relay logs show repeated deliveries of identical messages.

### Pitfall 7: CRC32 Collision Causes Silent Corruption

**What goes wrong:** Encrypted message data corrupted (disk error, memory bit flip), but CRC32 passes, queue delivers corrupted message.

**Why it happens:** CRC32 has 1 in 2^32 collision probability. For large queues or cosmic ray scenarios, collisions possible.

**How to avoid:** CRC32 is for fast detection, not security. Age encryption includes Poly1305 MAC (cryptographic authentication). If CRC32 passes but age.Decrypt fails, log error and quarantine message (don't retry). Users can manually inspect quarantined messages.

**Warning signs:** Age decryption failures with "authentication failed" (MAC mismatch). CRC32 passed but message unreadable.

### Pitfall 8: Queue Snapshot Corruption During Crash

**What goes wrong:** Relay crashes mid-write of JSON snapshot, corrupts queue metadata file, all queue state lost on restart.

**Why it happens:** JSON written directly to file without atomic write pattern.

**How to avoid:** Write snapshot to temp file, fsync, then atomic rename. On startup, if corruption detected, log error but don't crash—start with empty queue (messages already delivered or lost, better than crash loop).

**Warning signs:** Relay fails to start after crash with "JSON unmarshal error." Snapshot file is zero bytes or truncated.

## Code Examples

Verified patterns from official sources and derived from existing project code:

### Age Encryption/Decryption Streaming

```go
// Source: Context7 filippo.io/age examples + existing relay/forward pattern
package queue

import (
    "bytes"
    "filippo.io/age"
    "io"
)

func encryptStream(plaintext io.Reader, recipient age.Recipient) (io.Reader, error) {
    // Create encrypting pipe
    pr, pw := io.Pipe()

    go func() {
        w, err := age.Encrypt(pw, recipient)
        if err != nil {
            pw.CloseWithError(err)
            return
        }

        if _, err := io.Copy(w, plaintext); err != nil {
            pw.CloseWithError(err)
            return
        }

        if err := w.Close(); err != nil {
            pw.CloseWithError(err)
            return
        }

        pw.Close()
    }()

    return pr, nil
}

func decryptStream(encrypted io.Reader, identity age.Identity) (io.Reader, error) {
    // Decrypt returns a Reader directly
    return age.Decrypt(encrypted, identity)
}
```

### MinIO S3 Upload with Encryption

```go
// Source: Context7 minio-go examples + Storj documentation
package queue

import (
    "bytes"
    "context"
    "github.com/minio/minio-go/v7"
)

func uploadToStorj(ctx context.Context, client *minio.Client, bucket, key string, data []byte) error {
    // Data is already age-encrypted, no additional S3 encryption needed
    // (Storj provides end-to-end encryption by default)
    _, err := client.PutObject(ctx, bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
        ContentType: "application/octet-stream",
    })
    return err
}
```

### Exponential Backoff with Context

```go
// Source: Existing Phase 1 mTLS reconnection code + cenkalti/backoff docs
package queue

import (
    "context"
    "github.com/cenkalti/backoff/v4"
    "time"
)

func retryDelivery(ctx context.Context, fn func() error) error {
    bo := backoff.NewExponentialBackOff()
    bo.InitialInterval = 1 * time.Second
    bo.MaxInterval = 30 * time.Second
    bo.MaxElapsedTime = 5 * time.Minute

    return backoff.Retry(fn, backoff.WithContext(bo, ctx))
}
```

### Atomic File Write for Queue Snapshots

```go
// Source: Standard Go pattern for atomic writes
package queue

import (
    "encoding/json"
    "os"
    "path/filepath"
)

func (q *MessageQueue) writeSnapshotAtomic(path string) error {
    q.mu.RLock()
    defer q.mu.RUnlock()

    // Marshal queue metadata (not message data—already encrypted files)
    data, err := json.MarshalIndent(q.messages, "", "  ")
    if err != nil {
        return err
    }

    // Write to temp file
    tmpPath := path + ".tmp"
    f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
    if err != nil {
        return err
    }

    if _, err := f.Write(data); err != nil {
        f.Close()
        return err
    }

    // Fsync to ensure data on disk
    if err := f.Sync(); err != nil {
        f.Close()
        return err
    }
    f.Close()

    // Atomic rename
    return os.Rename(tmpPath, path)
}
```

### CRC32 Integrity Check

```go
// Source: Go stdlib hash/crc32 documentation
package queue

import (
    "hash/crc32"
)

func computeChecksum(data []byte) uint32 {
    return crc32.ChecksumIEEE(data) // IEEE polynomial (most common)
}

func verifyChecksum(data []byte, expected uint32) bool {
    actual := crc32.ChecksumIEEE(data)
    return actual == expected
}
```

### Message-ID Extraction with Fallback

```go
// Source: emersion/go-message documentation + RFC 5322 best practices
package queue

import (
    "bytes"
    "crypto/sha256"
    "fmt"
    "github.com/emersion/go-message/mail"
)

func extractMessageID(msgData []byte) string {
    mr, err := mail.CreateReader(bytes.NewReader(msgData))
    if err != nil {
        // Parse failure, use hash fallback
        return generateFallbackID(msgData)
    }

    msgID := mr.Header.Get("Message-ID")
    if msgID == "" {
        // No Message-ID header
        return generateFallbackID(msgData)
    }

    return msgID
}

func generateFallbackID(msgData []byte) string {
    h := sha256.Sum256(msgData)
    // Use first 128 bits (16 bytes) for ID
    return fmt.Sprintf("<fallback-%x@darkpipe.local>", h[:16])
}
```

## Integration with Existing Relay Architecture

### Modification Points

Based on existing codebase exploration:

1. **cmd/relay/main.go**: Wrap forwarder in QueuedForwarder before passing to NewServer
   ```go
   // BEFORE (Phase 2):
   forwarder := forward.NewWireGuardForwarder(cfg.HomeDeviceAddr)
   server := smtp.NewServer(forwarder, cfg)

   // AFTER (Phase 5):
   transport := forward.NewWireGuardForwarder(cfg.HomeDeviceAddr)
   queue := queue.NewMessageQueue(cfg.QueueConfig)
   forwarder := forward.NewQueuedForwarder(transport, queue, cfg.QueueEnabled)
   go queue.StartProcessor(context.Background(), transport) // Background delivery
   server := smtp.NewServer(forwarder, cfg)
   ```

2. **relay/config/config.go**: Add queue configuration fields
   ```go
   type Config struct {
       // ... existing fields ...

       // Queue configuration (QUEUE-01, QUEUE-02, QUEUE-03)
       QueueEnabled    bool
       QueueMaxRAMBytes int64
       QueueKeyPath     string // age identity file path

       // S3 overflow configuration (optional)
       OverflowEnabled  bool
       OverflowEndpoint string
       OverflowBucket   string
       OverflowAccessKey string
       OverflowSecretKey string
   }
   ```

3. **relay/forward/forwarder.go**: No changes needed—QueuedForwarder implements existing interface

4. **Postfix configuration**: No changes needed for queuing behavior. If user disables queuing (QUEUE-03), relay returns SMTP error, Postfix handles retry/bounce automatically per RFC 5321.

### No Postfix Integration Required

The existing relay architecture uses Postfix as a passive MTA—it receives mail on port 25, forwards to relay daemon on localhost:10025, and relay daemon handles delivery to home device. Queue logic sits entirely within the relay daemon (Go code). Postfix behavior unchanged:

- **Queue enabled**: Relay daemon queues on failure, Postfix accepts message immediately (250 OK)
- **Queue disabled**: Relay daemon returns 4xx error, Postfix queues and retries per normal SMTP semantics

This design avoids Postfix content_filter complexity and keeps queue logic in Go (easier to test, no shell scripts).

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| NaCl secretbox for encryption | Age for file/message encryption | 2019 (age release) | Simpler API, no nonce management, streaming support, better key format |
| AWS SDK for S3 | MinIO SDK for S3-compatible storage | 2015+ (MinIO maturity) | Works with Storj/MinIO/S3, simpler API, better multi-provider support |
| Custom persistent queues | In-memory queues with optional overflow | 2020s (container-native design) | Matches ephemeral relay design, simpler crash recovery, avoids disk I/O |
| MD5 for deduplication | SHA-256 for message hashing | 2010s (MD5 collision attacks) | Prevents collision-based dedup bypass |
| Postfix hold queue control | Application-layer queuing in Go | 2020s (microservice patterns) | Easier testing, no shell scripts, container-friendly |

**Deprecated/outdated:**
- **NaCl/secretbox**: Still secure but age provides better UX (key format, streaming, documentation)
- **defer_transports in Postfix**: Complex to manage, requires shell scripting; application-layer queuing is cleaner for container deployments
- **MD5 checksums**: Use SHA-256 or CRC32 (depending on use case); MD5 collisions are trivial to generate

## Open Questions

1. **Queue retention policy**
   - What we know: Messages queue when home device offline
   - What's unclear: Max time before purging old messages (24 hours? 7 days?)
   - Recommendation: Start with 7-day TTL, make configurable via env var (QUEUE_TTL_HOURS). Log warnings at 24h, 72h milestones.

2. **Overflow cost management**
   - What we know: Storj charges per GB stored and bandwidth
   - What's unclear: How to prevent runaway costs if home device offline for weeks
   - Recommendation: Max queue size limit (e.g., 1000 messages or 10GB), return 4xx when exceeded, document in user guide

3. **Age key rotation**
   - What we know: Age identity persists across restarts
   - What's unclear: Should keys rotate periodically? How to handle queued messages during rotation?
   - Recommendation: No rotation for Phase 5—static key is simpler. Document manual rotation procedure (decrypt queue with old key, re-encrypt with new key) for Phase 9 (operations).

4. **Postfix defer vs reject when queue disabled**
   - What we know: Queuing can be disabled (QUEUE-03)
   - What's unclear: Should relay return 4xx (sender retries) or 5xx (immediate bounce)?
   - Recommendation: Return 451 (temporary failure)—gives sender's MTA chance to retry. Document that bounce behavior is sender-dependent (not under user's control).

## Sources

### Primary (HIGH confidence)

- [filippo.io/age Context7 library](https://context7.com/filosottile/age/llms.txt) - Encryption API examples, key generation, streaming patterns
- [github.com/minio/minio-go Context7 library](https://context7.com/minio/minio-go/llms.txt) - S3-compatible storage operations, encryption support
- [Postfix After-Queue Content Filter](http://www.postfix.org/FILTER_README.html) - Official documentation on message deferral and reinjection
- [Postfix postsuper(1) manual](https://www.postfix.org/postsuper.1.html) - Queue management commands (hold, release, delete)
- [RFC 5322: Internet Message Format](https://datatracker.ietf.org/doc/html/rfc5322) - Message-ID uniqueness requirements
- [Go stdlib hash/crc32 package](https://pkg.go.dev/hash/crc32) - CRC32 checksum API
- [cenkalti/backoff/v4 package](https://pkg.go.dev/github.com/cenkalti/backoff/v4) - Exponential backoff with context support

### Secondary (MEDIUM confidence)

- [Storj S3 Compatibility](https://www.storj.io/blog/what-is-s3-compatibility) - S3 API compatibility confirmation, encryption at rest
- [Storj S3 Gateway Docs](https://storj.dev/dcs/api/s3/s3-compatible-gateway) - Authentication with access grants, encryption architecture
- [emersion/go-message GitHub](https://github.com/emersion/go-message) - RFC 5322 parsing library, Message-ID extraction
- [Postfix Queue Management](https://easyengine.io/tutorials/mail/postfix-queue/) - Queue operations tutorial (hold, defer, release)
- [ChaCha20-Poly1305 Wikipedia](https://en.wikipedia.org/wiki/ChaCha20-Poly1305) - AEAD algorithm background (used by age)
- [How to Implement Retry Logic in Go with Exponential Backoff](https://oneuptime.com/blog/post/2026-01-07-go-retry-exponential-backoff/view) - Recent (Jan 2026) practical patterns

### Tertiary (LOW confidence, flagged for validation)

- [EDRM Message ID Hash](https://www.relativity.com/blog/introducing-the-edrm-message-id-hash-simplify-cross-platform-email-duplicate-identification/) - Email deduplication patterns (legal discovery context, not SMTP)
- [Postfix defer_transports discussion](https://list.postfix.users.narkive.com/zBK3m13G/defer-transports) - Mailing list thread on queue control (older, may not reflect current Postfix)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries verified via Context7 or official docs, existing project already uses cenkalti/backoff and emersion libraries
- Architecture: HIGH - Patterns derived from existing relay code (forward/forwarder.go, smtp/session.go) and official library examples
- Pitfalls: MEDIUM-HIGH - Based on official Postfix docs (HIGH) and general Go/email best practices (MEDIUM), some pitfalls inferred from project constraints (256MB limit)

**Research date:** 2026-02-14
**Valid until:** 2026-04-14 (60 days—stable domain, libraries mature, Postfix changes infrequently)

**Key uncertainties resolved:**
- ✓ Age vs NaCl: Age chosen (simpler API, streaming, better docs)
- ✓ Queue persistence: In-memory with periodic JSON snapshots (matches ephemeral relay design)
- ✓ Postfix integration: None needed (application-layer queue in Go)
- ✓ S3 client: MinIO SDK (supports Storj, AWS S3, MinIO with one API)
- ✓ Deduplication: Message-ID header with SHA-256 fallback

**Research scope coverage:**
- ✓ Encryption approach for queued messages
- ✓ Go libraries for S3-compatible storage
- ✓ Queue management patterns (ordering, dedup, corruption)
- ✓ Integration with existing relay forwarding code
- ✓ Postfix queue control mechanisms
- ✓ Memory constraints on 256MB VPS