# S02: Log Hygiene & PII Redaction — Research

**Date:** 2026-03-11

## Summary

The codebase uses Go's standard `log` package exclusively (~116 log call sites across all Go files). There is no structured logging (no `slog`), no log-level gating, and no redaction layer. PII leaks are concentrated in two areas: (1) the cloud-relay SMTP session (`session.go`) which logs full email addresses in `MAIL FROM`, `RCPT TO`, `DATA`, and `SUCCESS` messages at every verbosity level, and (2) the profile-server handlers (`handlers.go`, `webui.go`) which log email addresses, partial tokens, and app password storage operations.

The project runs Go 1.25.7, which includes `log/slog` (available since Go 1.21). The codebase is small enough (~116 log sites) that a targeted approach — redacting PII in the specific hot-path files rather than a full slog migration — is the right call for this hardening slice. A debug flag (`RELAY_DEBUG` / `LOG_DEBUG`) can gate verbose envelope logging behind an opt-in mechanism.

The mailmigrate package also logs email addresses in warning messages, but these are CLI-interactive migration tools (not daemon logs), so they're lower priority. The monitoring/alert package formats alert emails with admin addresses but those are configuration values, not user PII.

## Recommendation

**Approach: Targeted redaction + debug flag, no logging framework migration.**

1. Create a small `internal/logutil` (or similar) package with an `EmailRedact(addr string) string` function that masks the local part of email addresses (e.g., `s***r@example.com`).
2. Add a `RELAY_DEBUG` env var (bool, default false) to `config.Config`.
3. In `smtp/session.go`: at default verbosity, log redacted addresses. When `RELAY_DEBUG=true`, log full addresses.
4. In `profile-server/handlers.go`: redact email in "Profile downloaded" and "QR code generated" log lines. The token prefix logging (`token[:8]`) is acceptable — tokens are single-use and 8 chars is not enough to reconstruct.
5. In `profile-server/webui.go`: redact email in "Failed to list app passwords" log line.
6. Audit `LogRequest` middleware — it logs `r.URL.RawQuery` which can contain `?emailaddress=` and `?token=`. Redact query params containing PII.
7. Skip mailmigrate package (CLI tool, not daemon, out of scope for container log hygiene).

**Why not slog migration?** The codebase is small and consistent with `log.Printf`. A full slog migration would touch every file, is better suited to a separate slice, and doesn't address the core problem (PII in log output). Redaction is orthogonal to the logging framework.

## Don't Hand-Roll

| Problem | Existing Solution | Why Use It |
|---------|------------------|------------|
| Email redaction | Simple string split on `@` + mask | Too simple for a library — 5 lines of code. Keep it in-repo. |
| Debug flag | `config.LoadFromEnv()` pattern | Already loads all env vars with defaults. Add `RELAY_DEBUG` the same way. |
| Structured logging | `log/slog` (stdlib) | Available but not needed for this slice. Future consideration. |

## Existing Code and Patterns

- `cloud-relay/relay/smtp/session.go` — **Primary PII source.** 6 log lines that print full `from` and `to` email addresses. Every SMTP transaction logs sender and all recipients in plaintext.
- `cloud-relay/relay/smtp/server.go` — Sets `s.ErrorLog = log.Default()`. The go-smtp library's internal error logging goes through the standard logger.
- `cloud-relay/relay/config/config.go` — Pattern for adding new env vars: `getEnv()` / `getEnvBool()` helpers. Add `RELAY_DEBUG` here.
- `cloud-relay/cmd/relay/main.go` — Startup log at line 37 prints `cfg.HomeDeviceAddr` (not PII, just network config). Line 43 prints webhook URL (operational, acceptable). No credentials logged.
- `cloud-relay/relay/queue/processor.go` — Logs message IDs (opaque hex strings, not PII). Clean.
- `home-device/profiles/cmd/profile-server/handlers.go:109` — `log.Printf("Profile downloaded for %s (token: %s, device: %s)", email, token[:8], deviceName)` — Logs full email + token prefix.
- `home-device/profiles/cmd/profile-server/handlers.go:212` — `log.Printf("QR code generated for %s", email)` — Logs full email.
- `home-device/profiles/cmd/profile-server/webui.go:97` — `log.Printf("Failed to list app passwords for %s: %v", email, err)` — Logs full email on error path.
- `home-device/profiles/cmd/profile-server/handlers.go:295-309` — `LogRequest` middleware logs `r.URL.RawQuery` which can contain `emailaddress=user@example.com` and `token=<full-token>`.
- `home-device/profiles/cmd/profile-server/webui.go:220-260` — Instructions HTML includes `plainPassword` in the response body (intentional — shown to user once). Not a log issue, but worth noting for the template.HTML audit in S03.

## Constraints

- **Go 1.25.7** — `log/slog` available but unused. No need to introduce it in this slice.
- **No external dependencies** — Project minimizes deps (decision: "cenkalti/backoff/v4 as only external Go dependency"). Redaction must be stdlib-only.
- **Two separate Go modules** — `cloud-relay` (root `go.mod`) and `home-device/profiles` (own `go.mod`). A shared `internal/logutil` package can't span modules without making it a proper package. Options: (a) duplicate the 5-line redaction function, (b) put it in root module and import from profile-server, (c) keep redaction inline. Option (a) is simplest given the function is trivial.
- **go-smtp library** logs via `s.ErrorLog` — its internal log output is not controllable beyond swapping the logger. It doesn't log PII (only connection errors).
- **Container log output** — These are Docker containers; logs go to `docker logs`. No syslog or log aggregation in scope.

## Common Pitfalls

- **Redacting too aggressively** — Redacting email domain makes logs useless for multi-domain debugging. Only mask the local part (before `@`), preserve domain.
- **Incomplete redaction in format strings** — The `DATA` log line includes both `from` and `to` as `%s` and `%v`. The `%v` for `[]string` needs the entire slice redacted, not just individual elements.
- **Query string PII in LogRequest middleware** — Easy to miss. The `r.URL.RawQuery` field will contain `emailaddress=user@example.com` on autoconfig requests and `token=<full-token>` on profile download requests. Must redact or omit sensitive query params.
- **Token logging** — `token[:8]` in handlers.go is actually fine — tokens are 32 bytes base64url (43 chars), single-use, and 8 chars is not reconstructible. Leave as-is.
- **Debug flag scope** — The profile-server is a separate binary with its own config. It needs its own debug flag (`PROFILE_DEBUG` or `LOG_DEBUG`), not `RELAY_DEBUG`.

## Open Risks

- **go-smtp library internal logging** — If the go-smtp library ever logs connection details including MAIL FROM/RCPT TO in error paths, that's outside our control. Low risk — reviewed the library and it only logs connection-level errors, not SMTP command content.
- **Postfix's own logs** — Postfix running in the same container logs full email addresses to syslog. This slice only covers Go application logs. Postfix log redaction would require `postconf -e header_checks` or `smtp_header_checks` — out of scope for S02 but worth noting.
- **Future log aggregation** — If logs are ever shipped to a centralized system, the redaction approach here (local-part masking) is sufficient for GDPR/privacy, but the decision should be documented.

## Skills Discovered

| Technology | Skill | Status |
|------------|-------|--------|
| Go (general) | saisudhir14/golang-agent-skill@golang | available (289 installs) — general Go skill, not specific to logging |
| Go idioms | rohitg00/awesome-claude-code-toolkit@golang-idioms | available (17 installs) — general patterns |
| Python logging | smithery.ai@logging | available (2 installs) — wrong language |
| Go observability | python-observability (installed) | installed but Python-specific, not applicable |

No directly relevant skills for Go log redaction. The work is straightforward stdlib code — no skill needed.

## Sources

- Codebase exploration: `rg` search across all `.go` files for `log.` call sites (116 total)
- `cloud-relay/relay/smtp/session.go` — 6 PII-leaking log lines identified
- `home-device/profiles/cmd/profile-server/handlers.go` — 3 PII-leaking log lines identified
- `home-device/profiles/cmd/profile-server/webui.go` — 1 PII-leaking log line identified
- `cloud-relay/relay/config/config.go` — env var loading pattern for debug flag
- Go `log/slog` availability confirmed via `go.mod` (Go 1.25.7 ≥ 1.21)
