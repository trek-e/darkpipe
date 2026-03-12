# M002: Post-Launch Hardening — Context

**Gathered:** 2026-03-11
**Status:** Ready for planning

## Project Description

DarkPipe is a shipped v1.0 self-hosted email sovereignty stack. This milestone addresses security hardening, operational quality, and code consistency findings from the post-launch product analysis.

## Why This Milestone

Product analysis identified 4 P1 (high-priority) and 7 P2 (medium-priority) issues across container security, TLS configuration, logging hygiene, error handling, and operational tooling. None are launch blockers, but they reduce the security posture and operational maturity of a product whose core promise is privacy and user control. Addressing them now prevents technical debt accumulation before v2 feature work begins.

## User-Visible Outcome

### When this milestone is complete, the user can:

- Deploy all DarkPipe containers knowing none run as root unnecessarily
- Trust that no PII (email addresses, tokens) leaks into container logs at default verbosity
- Reference `.env.example` files to understand all configuration options without reading source code
- Build Docker images with the same Go version declared in `go.mod`

### Entry point / environment

- Entry point: `docker compose up` / `darkpipe-setup` CLI
- Environment: Docker on Linux (amd64/arm64), Raspberry Pi 4+, NAS platforms
- Live dependencies involved: Postfix, Stalwart/Maddy/Dovecot, Rspamd, Redis, Caddy

## Completion Class

- Contract complete means: `go vet`, `go build`, all existing tests pass, new tests for hardened paths pass
- Integration complete means: Docker images build and start with non-root users, health checks pass
- Operational complete means: containers survive restart cycles, logs are clean at default verbosity

## Final Integrated Acceptance

To call this milestone complete, we must prove:

- All 5 Docker images build and run with explicit non-root users (or documented justification for root)
- `docker compose up` starts all services with health checks passing and no PII in default logs
- Provider IMAP connections use explicit TLS 1.2+ minimum
- SMTP relay rejects messages exceeding configurable size limit

## Risks and Unknowns

- Mail server base images (Stalwart, Maddy) may require root for port binding — need to test `cap_add: [NET_BIND_SERVICE]` approach
- Postfix traditionally runs as root — may need to keep root but drop capabilities

## Existing Codebase / Prior Art

- `cloud-relay/Dockerfile` — needs USER directive and size limit on SMTP DATA
- `cloud-relay/relay/smtp/session.go` — SMTP session with unbounded io.Copy
- `cloud-relay/relay/config/config.go` — env var config without .env.example
- `home-device/profiles/Dockerfile` — already has USER nonroot (reference pattern)
- `deploy/setup/pkg/providers/*.go` — 7 provider files with default tls.Config
- `transport/mtls/server/listener.go` — reference for explicit MinVersion: tls.VersionTLS12

> See `.gsd/DECISIONS.md` for all architectural and pattern decisions.

## Relevant Requirements

- No new requirements — this milestone hardens existing validated requirements

## Scope

### In Scope

- Container security hardening (USER directives, capability drops)
- TLS config explicitness in provider IMAP clients
- Log hygiene (PII redaction, structured logging)
- SMTP DATA size limit
- `.env.example` generation
- Go version alignment
- Error response consistency
- template.HTML audit
- Docker Compose security directives
- deploy/setup test coverage improvement

### Out of Scope / Non-Goals

- New features (v2 requirements)
- Architectural changes
- Performance optimization
- Documentation rewrites

## Technical Constraints

- Must not break existing Docker Compose deployments
- Must maintain compatibility with all 3 mail server options
- Changes to Dockerfiles must preserve multi-arch builds

## Integration Points

- Docker build pipeline (GitHub Actions) — Go version must match across go.mod and Dockerfiles
- Postfix — root requirement investigation
- Stalwart/Maddy base images — USER compatibility
