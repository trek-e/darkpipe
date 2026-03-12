---
id: T02
parent: S01
milestone: M001
provides:
  - "mTLS server with RequireAndVerifyClientCert for cloud relay"
  - "mTLS client with persistent connection and exponential backoff for home device"
  - "step-ca PKI setup helpers (InitCA, IssueCertificate, CheckCertExpiry)"
  - "Certificate renewal automation via systemd timers with jitter"
  - "step-ca deployment script for full CA installation"
  - "Shared test cert generator (ECDSA P-256) for mTLS tests"
requires: []
affects: []
key_files: []
key_decisions: []
patterns_established: []
observability_surfaces: []
drill_down_paths: []
duration: 7min
verification_result: passed
completed_at: 2026-02-09
blocker_discovered: false
---
# T02: 01-transport-layer 02

**# Phase 1 Plan 2: mTLS Transport and Internal PKI Summary**

## What Happened

# Phase 1 Plan 2: mTLS Transport and Internal PKI Summary

**mTLS server/client with mutual cert verification, step-ca PKI helpers, and systemd-based certificate renewal with randomized jitter**

## Performance

- **Duration:** 7 min
- **Started:** 2026-02-09T01:51:14Z
- **Completed:** 2026-02-09T01:58:16Z
- **Tasks:** 2
- **Files modified:** 13

## Accomplishments
- mTLS server enforces RequireAndVerifyClientCert -- rejects connections without valid client certificate
- mTLS client maintains persistent connection with exponential backoff (1s initial, 5min max, never gives up)
- step-ca PKI setup with Go helpers for CA initialization, cert issuance, and expiry checking
- Certificate renewal automated via systemd timer with RandomizedDelaySec jitter
- 18 tests cover full mTLS handshake, cert rejection, reconnection with retries, and context cancellation
- Only one external Go dependency added: cenkalti/backoff/v4 v4.3.0

## Task Commits

Each task was committed atomically:

1. **Task 1: step-ca PKI setup and certificate renewal automation** - `be8e8a1` (feat)
2. **Task 2: mTLS server/client with persistent connection and backoff** - `7782092` (feat)

## Files Created/Modified
- `transport/pki/ca/setup.go` - CAConfig, InitCA, IssueCertificate, CheckCertExpiry helpers wrapping step CLI
- `transport/pki/ca/setup_test.go` - Tests for cert expiry parsing, missing binary detection, config defaults
- `transport/mtls/server/listener.go` - mTLS server with RequireAndVerifyClientCert, Listen, Serve, Close
- `transport/mtls/server/listener_test.go` - Tests for valid client accept, cert-less rejection, Serve loop
- `transport/mtls/client/connector.go` - mTLS client with Connect and MaintainConnection (backoff retry)
- `transport/mtls/client/connector_test.go` - Tests for connection, retry on failure, context cancellation
- `transport/mtls/testutil/certs.go` - Shared test helper generating CA, server, client certs (ECDSA P-256)
- `transport/pki/renewal/systemd/cert-renewer@.service` - Systemd service with ExecCondition needs-renewal
- `transport/pki/renewal/systemd/cert-renewer@.timer` - Timer with RandomizedDelaySec=5min jitter
- `transport/pki/renewal/hooks/reload.sh` - Post-renewal hook reloading services via systemctl
- `deploy/pki/step-ca-setup.sh` - Full step-ca installation, CA init, systemd service setup
- `go.mod` - Added cenkalti/backoff/v4 dependency
- `go.sum` - Dependency checksums

## Decisions Made
- Used cenkalti/backoff/v4 as the only external Go dependency for reconnection (proven library, prevents thundering herd, supports context cancellation)
- Created shared testutil package for cert generation to avoid duplication between server and client test files
- Relied on Go's default TLS cipher suite selection rather than explicit configuration, allowing TLS 1.3 with post-quantum key exchange (X25519MLKEM768) in Go 1.24+
- Server tests explicitly call Handshake() on accepted connections because TLS 1.3 performs certificate verification post-handshake

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed TLS handshake deadlock in server acceptance test**
- **Found during:** Task 2 (mTLS server tests)
- **Issue:** TestServer_AcceptsValidClient hung indefinitely because tls.Listen returns connections before completing the TLS handshake. The server goroutine put the raw connection on a channel without calling Handshake(), causing the client-side tls.Dial to wait forever for the server's handshake response.
- **Fix:** Added explicit tc.Handshake() call in the server Accept goroutine before signaling the result channel
- **Files modified:** transport/mtls/server/listener_test.go
- **Verification:** Test completes in <1s instead of timing out
- **Committed in:** 7782092 (Task 2 commit)

**2. [Rule 1 - Bug] Fixed TLS 1.3 cert rejection test for post-handshake authentication**
- **Found during:** Task 2 (mTLS server tests)
- **Issue:** TestServer_RejectsClientWithoutCert falsely passed because TLS 1.3 sends client certificate requests post-handshake. tls.Dial succeeds, and the rejection only surfaces during data exchange. The original test checked only for Dial/Handshake failure.
- **Fix:** Restructured test to attempt data exchange (Write+Read) after Dial, verifying that the connection fails at the data level even if the initial handshake succeeds
- **Files modified:** transport/mtls/server/listener_test.go
- **Verification:** Test correctly passes by detecting write/read failure on no-cert connection
- **Committed in:** 7782092 (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (2 bugs in test code)
**Impact on plan:** Both fixes were necessary for test correctness. TLS 1.3 post-handshake behavior is a known subtlety. No scope creep.

## Issues Encountered
None beyond the test fixes documented above.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- mTLS transport complete and tested, ready for Plan 01-03 (auto-reconnection hardening)
- PKI foundation available for Phase 2 (cloud relay) certificate needs
- cenkalti/backoff/v4 available for any future retry logic
- Shared test cert generator available in transport/mtls/testutil for future tests

## Self-Check: PASSED

- All 12 created files verified present on disk
- Commit be8e8a1 (Task 1) verified in git log
- Commit 7782092 (Task 2) verified in git log
- 18 tests passing across transport/pki and transport/mtls packages
- go vet reports no issues

---
*Phase: 01-transport-layer*
*Completed: 2026-02-09*
