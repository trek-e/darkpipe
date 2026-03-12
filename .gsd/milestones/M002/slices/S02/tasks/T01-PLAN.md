---
estimated_steps: 5
estimated_files: 4
---

# T01: Add email redaction and debug-gated logging to cloud-relay

**Slice:** S02 — Log Hygiene & PII Redaction
**Milestone:** M002

## Description

Create the `RedactEmail` utility in the cloud-relay module, add a `RELAY_DEBUG` env var to the config, and apply redaction to all 6 PII-leaking log lines in `smtp/session.go`. At default verbosity, SMTP session logs will show `s***r@example.com` instead of `sender@example.com`. When `RELAY_DEBUG=true`, full addresses are logged for operational debugging.

## Steps

1. Create `cloud-relay/relay/logutil/redact.go` with `RedactEmail(addr string) string` — split on `@`, keep first and last char of local-part with `***` between, preserve domain. Handle edge cases: missing `@` (return as-is), single-char local-part (return `*@domain`), two-char local-part (`s*@domain`), empty string (return empty). Add `RedactEmails(addrs []string) []string` helper for recipient slices.
2. Create `cloud-relay/relay/logutil/redact_test.go` with table-driven tests covering: normal address, short local-parts (1, 2, 3 chars), no `@` sign, empty string, multiple addresses via `RedactEmails`.
3. Add `RelayDebug bool` field to `config.Config` struct in `cloud-relay/relay/config/config.go`, loaded via `getEnvBool("RELAY_DEBUG", false)`.
4. Thread debug flag to session: add a `debug bool` field to the `Session` struct in `smtp/session.go`. Set it from config when creating sessions in `server.go` or wherever sessions are instantiated. In each of the 6 log lines that print `from` or `to` addresses, use `logutil.RedactEmail(from)` / `logutil.RedactEmails(to)` when `!s.debug`, and raw values when `s.debug`.
5. Run `go vet ./...` and `go test ./...` from `cloud-relay/` to confirm everything compiles and passes.

## Must-Haves

- [ ] `RedactEmail` masks local-part, preserves domain
- [ ] `RedactEmails` handles `[]string` slices
- [ ] Edge cases handled: empty, no-@, 1-char, 2-char local-parts
- [ ] `RELAY_DEBUG` env var added to config (default false)
- [ ] All 6 session.go log lines use redacted addresses at default verbosity
- [ ] Full addresses logged when `RELAY_DEBUG=true`
- [ ] Table-driven unit tests for redaction function

## Verification

- `cd cloud-relay && go vet ./...` exits 0
- `cd cloud-relay && go test ./...` exits 0 with new redaction tests passing
- Manual review: no raw `%s` or `%v` for email variables in session.go log lines without a debug guard

## Observability Impact

- Signals added/changed: SMTP session log lines now emit redacted email addresses by default; `RELAY_DEBUG=true` restores full verbosity
- How a future agent inspects this: set `RELAY_DEBUG=true` in environment, restart relay, check `docker logs` for full addresses
- Failure state exposed: if redaction function returns empty or malformed output, unit tests fail; log lines would show obviously broken addresses (e.g., `***@` or empty) making the problem self-evident

## Inputs

- `cloud-relay/relay/smtp/session.go` — 6 log lines with raw email addresses (identified in research)
- `cloud-relay/relay/config/config.go` — `getEnvBool()` pattern for adding new env vars
- S02-RESEARCH.md — exact line locations and format string patterns

## Expected Output

- `cloud-relay/relay/logutil/redact.go` — new file with `RedactEmail` and `RedactEmails` functions
- `cloud-relay/relay/logutil/redact_test.go` — new file with table-driven tests
- `cloud-relay/relay/config/config.go` — modified with `RelayDebug` field
- `cloud-relay/relay/smtp/session.go` — modified: all 6 log lines use conditional redaction
