# M002: Post-Launch Hardening

**Vision:** All DarkPipe containers run with minimal privileges, logs never leak PII, and operational configuration is self-documenting — matching the security posture users expect from a privacy-first email stack.

## Success Criteria

- No container runs as root without documented justification and explicit capability restrictions
- Default log verbosity contains zero email addresses, tokens, or credentials
- Every environment variable has a documented default in `.env.example`
- All TLS connections specify explicit minimum version
- SMTP relay enforces a configurable message size limit

## Key Risks / Unknowns

- Postfix/Stalwart/Maddy may require root for privileged port binding — investigate cap_add approach

## Proof Strategy

- Root requirement → retire in S01 by proving containers start and pass health checks as non-root (or with minimal caps)

## Verification Classes

- Contract verification: `go vet`, `go build`, unit tests, `docker build`
- Integration verification: `docker compose up` with health check validation
- Operational verification: container restart cycles, log output inspection
- UAT / human verification: none

## Milestone Definition of Done

This milestone is complete only when all are true:

- All slices are complete with passing verification
- Docker images build and run with hardened security
- `go vet ./...` and `go build ./...` pass cleanly
- All existing tests continue to pass
- Success criteria are verified against running containers

## Slices

- [x] **S01: Container Security Hardening** `risk:medium` `depends:[]`
  > After this: All 5 Dockerfiles specify non-root USER (or document root justification with cap_drop/security_opt), Docker Compose files include read_only, no-new-privileges, and cap_drop directives.
- [x] **S02: Log Hygiene & PII Redaction** `risk:low` `depends:[]`
  > After this: SMTP session logs redact email addresses at default verbosity, token logging is behind a debug flag, and no credentials appear in any log output.
- [ ] **S03: TLS & Input Hardening** `risk:low` `depends:[]`
  > After this: All 7 provider IMAP clients specify MinVersion: tls.VersionTLS12 explicitly, SMTP DATA handler enforces a configurable message size limit, and template.HTML usage is audited safe.
- [ ] **S04: Operational Quality** `risk:low` `depends:[]`
  > After this: .env.example files exist for cloud-relay and home-device, Go version is consistent across go.mod files and Dockerfiles, web API error responses are structured JSON, and deploy/setup test coverage reaches project average.

## Boundary Map

### S01 (independent)

Produces:
- Hardened Dockerfiles with USER directives and HEALTHCHECK
- Docker Compose security directives (read_only, no-new-privileges, cap_drop)

Consumes:
- nothing (independent)

### S02 (independent)

Produces:
- Redacted log output at default verbosity
- Debug-gated verbose logging

Consumes:
- nothing (independent)

### S03 (independent)

Produces:
- Explicit tls.Config with MinVersion on all provider IMAP clients
- Configurable SMTP message size limit
- Audited template.HTML usage

Consumes:
- nothing (independent)

### S04 (independent)

Produces:
- `.env.example` files for cloud-relay and home-device
- Aligned Go versions across go.mod and Dockerfiles
- Structured JSON error responses on API endpoints
- Additional test files for deploy/setup packages

Consumes:
- nothing (independent)
