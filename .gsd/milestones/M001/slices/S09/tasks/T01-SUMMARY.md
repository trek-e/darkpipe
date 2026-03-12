---
id: T01
parent: S09
milestone: M001
provides: []
requires: []
affects: []
key_files: []
key_decisions: []
patterns_established: []
observability_surfaces: []
drill_down_paths: []
duration: 
verification_result: passed
completed_at: 
blocker_discovered: false
---
# T01: 09-monitoring-observability 01

**# Phase 09 Plan 01: Core Monitoring Data Collection Summary**

## What Happened

# Phase 09 Plan 01: Core Monitoring Data Collection Summary

Core monitoring packages (health checks, queue stats, delivery tracking) implemented with full test coverage and thread safety.

## What Was Built

### Health Check Framework (monitoring/health/)

**Unified health checker with liveness/readiness separation:**
- `Checker` struct with extensible `RegisterCheck()` for adding health checks
- `Liveness()` always returns "up" (cheap check, process is alive)
- `Readiness()` runs deep checks and returns "down" if any fail
- HTTP handlers for `/health/live` (200 OK) and `/health/ready` (200/503)
- Content-Type: application/health+json for Kubernetes compatibility

**Service-specific health checks:**
- `CheckPostfix()` - TCP dial to localhost:25 with 2-second timeout
- `CheckIMAP()` - TCP dial to localhost:993 with 2-second timeout
- `CheckTunnel()` - leverages transport/health.Check() for WireGuard/mTLS
- TRANSPORT_TYPE env var determines tunnel type (wireguard or mtls)

**JSON-serializable responses:**
- `HealthStatus` with status, checks array, timestamp
- `CheckResult` with name, status, message, duration (milliseconds)
- Custom JSON marshaling for duration fields

### Queue Monitoring (monitoring/queue/)

**Postfix queue parser via postqueue -j:**
- `QueueMessage` struct with queue_id, queue_name, arrival_time (Unix epoch), sender, recipients
- `GetQueueStats()` returns aggregate statistics (depth, active, deferred, hold, stuck)
- `GetDetailedQueue()` returns full queue snapshot with individual messages
- Stuck message detection: messages older than 24-hour threshold
- Empty queue handling (postqueue returns empty output)

**Mock executor pattern for testing:**
- `PostqueueExecutor` interface allows injecting test data
- `RealPostqueueExecutor` runs actual postqueue command
- `SetPostqueueExecutor()` for test injection

**NDJSON parsing:**
- Handles Postfix postqueue -j output (one JSON object per line)
- Skips malformed lines gracefully (continues parsing valid entries)
- Classifies by queue_name: active, deferred, hold

### Delivery Tracking (monitoring/delivery/)

**Postfix mail.log parser:**
- `Parser` with regex patterns for status=sent/deferred/bounced/expired
- Extracts to=, from=, relay=, delay=, dsn=, status detail from log lines
- Timestamp parsing from syslog format (MMM DD HH:MM:SS)
- Queue ID extraction with alphanumeric validation
- Returns nil for non-delivery log lines (connection, cleanup, etc.)

**Ring buffer tracker:**
- `DeliveryTracker` with configurable capacity (default 1000, env MONITOR_DELIVERY_HISTORY)
- Thread-safe Record() and read operations (sync.RWMutex)
- `GetRecent(n)` returns last N entries, newest first
- `GetByQueueID(id)` returns all entries for a queue ID (multi-recipient tracking)
- `GetStats()` aggregates delivered/deferred/bounced/expired counts with time period

**Wraparound behavior:**
- Ring buffer wraps at capacity (oldest entries overwritten)
- Handles both pre-wrap (partial buffer) and post-wrap (full buffer) states
- Index calculation for newest-first retrieval

## Verification

All tests passing with race detector:
```
go test ./monitoring/health/... -v -race
go test ./monitoring/queue/... -v -race
go test ./monitoring/delivery/... -v -race
go vet ./monitoring/health/... ./monitoring/queue/... ./monitoring/delivery/...
```

**Test coverage:**
- 47 tests across 3 packages
- Health check: 9 tests (liveness, readiness, HTTP handlers, JSON marshaling)
- Queue monitor: 8 tests (JSON parsing, stuck detection, queue classification)
- Delivery tracker: 14 tests (ring buffer, thread safety, queue ID lookup)
- Delivery parser: 16 tests (status extraction, timestamp parsing, malformed lines)

**Race detector clean:**
- Thread safety verified on DeliveryTracker (10 writers + 5 readers concurrently)
- No data races in RWMutex usage

## Deviations from Plan

None - plan executed exactly as written.

## Key Decisions

1. **Liveness vs Readiness:** Liveness is cheap (always "up"), readiness is deep (actual service checks). This follows Kubernetes patterns and avoids unnecessary restarts.

2. **application/health+json content type:** Future-proofs for Kubernetes health probes (RFC draft-inadarei-api-health-check).

3. **postqueue -j JSON output:** More reliable than text parsing, handles empty queue gracefully.

4. **24-hour stuck threshold:** Configurable via QueueStats.StuckThreshold field, can be adjusted per deployment.

5. **Ring buffer over database:** Avoids I/O overhead for high-throughput mail servers. 1000 entries covers ~30 minutes of typical traffic.

6. **Thread-safe tracker:** RWMutex allows concurrent log parsing (writer) and query access (readers) without blocking.

7. **Both inbound and outbound tracking:** Parser extracts all delivery statuses regardless of direction (relay->home or home->world).

8. **Status normalization:** Maps Postfix status=sent to "delivered" for clearer user-facing language.

## Success Criteria Met

- MON-01 (partial): Queue depth and stuck message count available via GetQueueStats()
- MON-02 (partial): Delivery status tracking with sent/deferred/bounced classification
- MON-03 (partial): Health check endpoints with liveness/readiness separation

These three packages form the data layer for Plan 09-02 (CLI) and Plan 09-03 (aggregation/alerts).

## Next Steps

Plan 09-02 will build the CLI that consumes these packages (darkpipe-ctl monitor queue, darkpipe-ctl monitor delivery, darkpipe-ctl monitor health).

Plan 09-03 will aggregate all monitoring data and implement alert triggers.

## Self-Check

Verifying all claimed files and commits exist:

```bash
# Check created files
[ -f "monitoring/health/checker.go" ] && echo "FOUND" || echo "MISSING"
[ -f "monitoring/health/postfix.go" ] && echo "FOUND" || echo "MISSING"
[ -f "monitoring/health/imap.go" ] && echo "FOUND" || echo "MISSING"
[ -f "monitoring/health/tunnel.go" ] && echo "FOUND" || echo "MISSING"
[ -f "monitoring/health/server.go" ] && echo "FOUND" || echo "MISSING"
[ -f "monitoring/health/checker_test.go" ] && echo "FOUND" || echo "MISSING"
[ -f "monitoring/queue/mailq.go" ] && echo "FOUND" || echo "MISSING"
[ -f "monitoring/queue/stats.go" ] && echo "FOUND" || echo "MISSING"
[ -f "monitoring/queue/mailq_test.go" ] && echo "FOUND" || echo "MISSING"
[ -f "monitoring/delivery/parser.go" ] && echo "FOUND" || echo "MISSING"
[ -f "monitoring/delivery/tracker.go" ] && echo "FOUND" || echo "MISSING"
[ -f "monitoring/delivery/status.go" ] && echo "FOUND" || echo "MISSING"
[ -f "monitoring/delivery/parser_test.go" ] && echo "FOUND" || echo "MISSING"
[ -f "monitoring/delivery/tracker_test.go" ] && echo "FOUND" || echo "MISSING"

# Check commits
git log --oneline --all | grep "519a9cb"
git log --oneline --all | grep "9fbafdf"
```

## Self-Check: PASSED

All 14 files verified present.
Both commits verified in git log:
- 519a9cb: Task 1 (health + queue)
- 9fbafdf: Task 2 (delivery tracker)
