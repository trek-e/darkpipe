---
id: T02
parent: S02
milestone: M002
provides:
  - RedactEmail utility in profile-server module (internal/logutil)
  - RedactQueryParams utility for sanitizing logged query strings
  - PROFILE_DEBUG env var for verbose logging
  - verify-log-redaction.sh static analysis CI script
key_files:
  - home-device/profiles/internal/logutil/redact.go
  - home-device/profiles/internal/logutil/redact_test.go
  - home-device/profiles/cmd/profile-server/handlers.go
  - home-device/profiles/cmd/profile-server/webui.go
  - scripts/verify-log-redaction.sh
key_decisions:
  - Used package-level var + init via os.Getenv for PROFILE_DEBUG (profile-server has no config struct, keeping it simple)
  - Created logEmail() helper in handlers.go to centralize debug-gated redaction (same pattern as T01's logFrom/logTo)
  - RedactQueryParams uses url.ParseQuery for robust parsing; falls back to redacting entire query on parse failure
patterns_established:
  - logutil package duplicated per Go module (can't share across module boundaries)
  - logEmail() helper centralizes debug-flag check + redaction in one place
  - verify-log-redaction.sh as CI-friendly static check for PII leaks
observability_surfaces:
  - PROFILE_DEBUG=true env var restores full email addresses in profile-server logs
  - LogRequest middleware logs sanitized query strings (email/token values replaced with [REDACTED])
  - verify-log-redaction.sh exits 1 with details if unredacted patterns are found
duration: 15m
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T02: Add email redaction to profile-server and harden LogRequest middleware

**Duplicated RedactEmail into profile-server module, added RedactQueryParams for query string sanitization, applied PII redaction to all 3 log lines, hardened LogRequest middleware, and created static analysis CI script.**

## What Happened

Created `internal/logutil` package in the profile-server module with `RedactEmail` (identical logic to T01's cloud-relay version) and `RedactQueryParams` (parses query string, replaces values for `emailaddress`, `email`, and `token` keys with `[REDACTED]`). Key matching is case-insensitive.

Added `PROFILE_DEBUG` env var support via a package-level `profileDebug` bool initialized from `os.Getenv("PROFILE_DEBUG")` at startup. Created `logEmail()` helper that returns redacted or full email based on the debug flag.

Applied redaction to all 3 PII-leaking log lines:
- `handlers.go`: "Profile downloaded for %s" and "QR code generated for %s"
- `webui.go`: "Failed to list app passwords for %s"

Hardened `LogRequest` middleware by replacing `r.URL.RawQuery` with `logutil.RedactQueryParams(r.URL.RawQuery)` in the JSON log entry.

Created `scripts/verify-log-redaction.sh` that checks session.go, handlers.go, and webui.go for unguarded `log.Printf` calls with raw email variables, plus verifies no raw `r.URL.RawQuery` usage without `RedactQueryParams`.

## Verification

- `cd home-device/profiles && go vet ./...` — exits 0
- `cd home-device/profiles && go test ./...` — all pass (internal/logutil + cmd/profile-server)
- `bash scripts/verify-log-redaction.sh` — exits 0, "PASSED: No unredacted PII patterns found"
- `cd cloud-relay && go vet ./... && go test ./...` — all pass (slice-level check, T01 unaffected)

All 3 slice-level verification checks pass. S02 is complete.

## Diagnostics

- Set `PROFILE_DEBUG=true`, restart profile-server, trigger profile download or QR generation — logs should show full email addresses
- Unset or set `PROFILE_DEBUG=false` — logs show redacted emails (e.g., `s***r@example.com`)
- Run `bash scripts/verify-log-redaction.sh` to check for regressions in source
- If redaction breaks: log output shows malformed addresses, unit tests fail

## Deviations

- Fixed verify script path: `cloud-relay/relay/session.go` → `cloud-relay/relay/smtp/session.go` (actual file location)

## Known Issues

None.

## Files Created/Modified

- `home-device/profiles/internal/logutil/redact.go` — new: RedactEmail + RedactQueryParams functions
- `home-device/profiles/internal/logutil/redact_test.go` — new: table-driven tests for both functions
- `home-device/profiles/cmd/profile-server/handlers.go` — modified: added logutil import, PROFILE_DEBUG support, logEmail helper, redacted 2 log lines, hardened LogRequest middleware
- `home-device/profiles/cmd/profile-server/webui.go` — modified: redacted 1 log line
- `scripts/verify-log-redaction.sh` — new: static analysis CI script for PII leak detection
