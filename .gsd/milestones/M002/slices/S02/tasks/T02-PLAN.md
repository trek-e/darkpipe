---
estimated_steps: 5
estimated_files: 5
---

# T02: Add email redaction to profile-server and harden LogRequest middleware

**Slice:** S02 — Log Hygiene & PII Redaction
**Milestone:** M002

## Description

Duplicate the `RedactEmail` utility into the profile-server module (separate Go module, can't import from cloud-relay), add `PROFILE_DEBUG` env var support, redact all 3 PII-leaking log lines in handlers.go and webui.go, and harden the `LogRequest` middleware to strip email and token values from logged query strings. Create a static verification script that greps for unredacted patterns as a CI guard.

## Steps

1. Create `home-device/profiles/internal/logutil/redact.go` with the same `RedactEmail` function as T01 (identical logic, separate module). Add `RedactQueryParams(rawQuery string) string` that parses query params and replaces values for keys `emailaddress`, `email`, and `token` with `[REDACTED]`.
2. Create `home-device/profiles/internal/logutil/redact_test.go` with table-driven tests for `RedactEmail` (same cases as T01) plus tests for `RedactQueryParams` covering: query with emailaddress param, query with token param, query with both, query with neither, empty query string.
3. Add `PROFILE_DEBUG` support: read `os.Getenv("PROFILE_DEBUG")` in handlers.go (profile-server has no config struct — keep it simple with a package-level var or init). When false/unset, use redacted addresses. When true, log full addresses.
4. Apply redaction to the 3 log lines: `handlers.go:109` ("Profile downloaded for %s"), `handlers.go:212` ("QR code generated for %s"), `webui.go:97` ("Failed to list app passwords for %s"). Leave `token[:8]` as-is. In `LogRequest` middleware (`handlers.go:295-309`), replace `r.URL.RawQuery` with `logutil.RedactQueryParams(r.URL.RawQuery)` in the log output.
5. Create `scripts/verify-log-redaction.sh` — a bash script that uses `grep`/`rg` to check that session.go, handlers.go, and webui.go don't contain unguarded `log.Printf` calls with raw email format strings (e.g., patterns like `log.Printf.*email` without a corresponding `Redact` call). Exit 0 if clean, exit 1 with details if violations found. Run full verification: `cd home-device/profiles && go vet ./... && go test ./...` and `bash scripts/verify-log-redaction.sh`.

## Must-Haves

- [ ] `RedactEmail` duplicated into profile-server module
- [ ] `RedactQueryParams` strips sensitive query parameter values
- [ ] `PROFILE_DEBUG` env var gates verbose logging (default off)
- [ ] All 3 profile-server log lines use redacted emails at default verbosity
- [ ] `LogRequest` middleware no longer logs raw email/token query values
- [ ] Unit tests for both `RedactEmail` and `RedactQueryParams`
- [ ] `scripts/verify-log-redaction.sh` static analysis script passes

## Verification

- `cd home-device/profiles && go vet ./...` exits 0
- `cd home-device/profiles && go test ./...` exits 0 with new tests passing
- `bash scripts/verify-log-redaction.sh` exits 0

## Observability Impact

- Signals added/changed: profile-server log lines emit redacted emails by default; `PROFILE_DEBUG=true` restores full verbosity; LogRequest middleware logs sanitized query strings
- How a future agent inspects this: set `PROFILE_DEBUG=true`, restart profile-server, check logs for full addresses; run verify script for static analysis
- Failure state exposed: broken redaction produces obviously malformed log output; unit tests catch function-level regressions; verify script catches source-level regressions

## Inputs

- `home-device/profiles/cmd/profile-server/handlers.go` — 3 PII log lines + LogRequest middleware (from research)
- `home-device/profiles/cmd/profile-server/webui.go` — 1 PII log line (from research)
- T01 `redact.go` — reference implementation for `RedactEmail` (duplicate, don't import)

## Expected Output

- `home-device/profiles/internal/logutil/redact.go` — new file with `RedactEmail` and `RedactQueryParams`
- `home-device/profiles/internal/logutil/redact_test.go` — new file with tests
- `home-device/profiles/cmd/profile-server/handlers.go` — modified: redacted log lines + hardened LogRequest
- `home-device/profiles/cmd/profile-server/webui.go` — modified: redacted log line
- `scripts/verify-log-redaction.sh` — new CI-friendly static check script
