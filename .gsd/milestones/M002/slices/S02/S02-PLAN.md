# S02: Log Hygiene & PII Redaction

**Goal:** SMTP session logs redact email addresses at default verbosity, token logging is behind a debug flag, and no credentials appear in any log output.
**Demo:** Run the cloud-relay and profile-server, trigger SMTP transactions and profile downloads, inspect `docker logs` output — no full email addresses or tokens visible. Set `RELAY_DEBUG=true` / `PROFILE_DEBUG=true` and re-check — full addresses appear.

## Must-Haves

- Email local-part redaction function that masks everything before `@` (e.g., `s***r@example.com`), preserving domain
- `RELAY_DEBUG` env var in cloud-relay config (bool, default false) gates verbose envelope logging
- `PROFILE_DEBUG` env var in profile-server gates verbose user-identity logging
- All 6 PII-leaking log lines in `smtp/session.go` use redacted addresses at default verbosity
- All 3 PII-leaking log lines in profile-server (`handlers.go`, `webui.go`) use redacted addresses at default verbosity
- `LogRequest` middleware redacts `emailaddress=` and `token=` query params from logged URLs
- Unit tests for the redaction function in both modules
- `go vet` and `go build` pass cleanly for both modules

## Proof Level

- This slice proves: operational
- Real runtime required: no (unit tests + build verification; runtime log inspection is optional UAT)
- Human/UAT required: no

## Verification

- `cd cloud-relay && go vet ./... && go test ./...` — all pass, including new redaction tests
- `cd home-device/profiles && go vet ./... && go test ./...` — all pass, including new redaction tests
- `bash scripts/verify-log-redaction.sh` — static grep script confirms no unredacted `log.Printf` calls with raw email format strings in hot-path files

## Observability / Diagnostics

- Runtime signals: log output itself is the observable surface — redacted by default, full-detail when debug flag is set
- Inspection surfaces: `RELAY_DEBUG=true` / `PROFILE_DEBUG=true` env vars restore verbose logging for operational debugging
- Failure visibility: if redaction breaks, log output will contain `<redacted>` or empty strings instead of masked addresses — unit tests catch this
- Redaction constraints: email local-parts and full tokens are PII; domains and token prefixes (8 chars) are not

## Integration Closure

- Upstream surfaces consumed: `cloud-relay/relay/config/config.go` (env var loading pattern), `cloud-relay/relay/smtp/session.go` (log call sites), `home-device/profiles/cmd/profile-server/handlers.go` and `webui.go` (log call sites)
- New wiring introduced in this slice: redaction functions called inline at log sites; debug flags read from config and passed to session/handlers
- What remains before the milestone is truly usable end-to-end: S03 (TLS hardening), S04 (operational quality) — independent of this slice

## Tasks

- [x] **T01: Add email redaction and debug-gated logging to cloud-relay** `est:45m`
  - Why: The cloud-relay SMTP session is the primary PII source — 6 log lines print full sender/recipient addresses on every transaction
  - Files: `cloud-relay/relay/logutil/redact.go`, `cloud-relay/relay/logutil/redact_test.go`, `cloud-relay/relay/config/config.go`, `cloud-relay/relay/smtp/session.go`
  - Do: Create `relay/logutil` package with `RedactEmail` function (split on `@`, mask local-part keeping first+last char). Add `RelayDebug` bool to config loaded from `RELAY_DEBUG` env var. Thread debug flag into Session struct. Replace all 6 `log.Printf` calls in session.go to use redacted addresses when debug is off, full addresses when on. Handle `[]string` recipient slices. Write table-driven unit tests for `RedactEmail` covering normal addresses, short local-parts, missing `@`, empty strings.
  - Verify: `cd cloud-relay && go vet ./... && go test ./...`
  - Done when: All tests pass, `go vet` clean, no raw `%s` or `%v` for email addresses in session.go log lines at default verbosity

- [x] **T02: Add email redaction to profile-server and harden LogRequest middleware** `est:45m`
  - Why: Profile-server logs full email addresses in 3 places and the LogRequest middleware leaks email+token via query strings — completing coverage for all Go application log PII
  - Files: `home-device/profiles/internal/logutil/redact.go`, `home-device/profiles/internal/logutil/redact_test.go`, `home-device/profiles/cmd/profile-server/handlers.go`, `home-device/profiles/cmd/profile-server/webui.go`, `scripts/verify-log-redaction.sh`
  - Do: Duplicate `RedactEmail` function into profile-server's own `internal/logutil` package (separate Go module — can't share). Add `PROFILE_DEBUG` env var support (simple `os.Getenv` check, profile-server has no config struct). Redact email in "Profile downloaded", "QR code generated", and "Failed to list app passwords" log lines. In `LogRequest` middleware, redact query params: strip `emailaddress` and `token` values from `r.URL.RawQuery` before logging (replace values with `[REDACTED]`). Leave `token[:8]` logging as-is (acceptable per research). Write unit tests. Create `scripts/verify-log-redaction.sh` that greps session.go, handlers.go, and webui.go for unredacted email format patterns as a CI-friendly static check.
  - Verify: `cd home-device/profiles && go vet ./... && go test ./...` and `bash scripts/verify-log-redaction.sh`
  - Done when: All tests pass, `go vet` clean, verification script exits 0, LogRequest no longer logs raw email or token query params

## Files Likely Touched

- `cloud-relay/relay/logutil/redact.go` (new)
- `cloud-relay/relay/logutil/redact_test.go` (new)
- `cloud-relay/relay/config/config.go`
- `cloud-relay/relay/smtp/session.go`
- `home-device/profiles/internal/logutil/redact.go` (new)
- `home-device/profiles/internal/logutil/redact_test.go` (new)
- `home-device/profiles/cmd/profile-server/handlers.go`
- `home-device/profiles/cmd/profile-server/webui.go`
- `scripts/verify-log-redaction.sh` (new)
