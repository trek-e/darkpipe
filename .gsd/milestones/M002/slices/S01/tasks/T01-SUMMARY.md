---
id: T01
parent: S01
milestone: M002
provides:
  - security verification script for all compose files and Dockerfiles
  - hardened cloud-relay compose (relay + caddy services)
  - hardened certbot compose
key_files:
  - scripts/verify-container-security.sh
  - cloud-relay/docker-compose.yml
  - cloud-relay/certbot/docker-compose.certbot.yml
key_decisions:
  - Verification script uses pure bash/awk parsing (no yq/python dependency) for portability
patterns_established:
  - Security directive block order: security_opt → cap_drop → cap_add → read_only → tmpfs, placed before ports
observability_surfaces:
  - scripts/verify-container-security.sh — reusable audit script with per-service PASS/FAIL and exit code
duration: 15m
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T01: Create security verification script and harden cloud-relay compose files

**Created `scripts/verify-container-security.sh` audit script and applied security hardening (no-new-privileges, cap_drop ALL, selective cap_add, read_only, tmpfs) to relay, caddy, and certbot services.**

## What Happened

1. Created `scripts/verify-container-security.sh` — checks all 3 compose files for `security_opt`, `cap_drop`, and `read_only` on every service, and all 5 Dockerfiles for `HEALTHCHECK`. Uses pure bash/awk for zero external dependencies.
2. Hardened the `caddy` service: no-new-privileges, cap_drop ALL, cap_add NET_BIND_SERVICE, read_only with tmpfs for /tmp and /run.
3. Hardened the `relay` service: no-new-privileges, cap_drop ALL, cap_add [NET_ADMIN, NET_BIND_SERVICE, DAC_OVERRIDE, CHOWN, SETGID, SETUID, KILL], read_only with tmpfs for Postfix writable dirs. Added root justification comment. Consolidated the old standalone `cap_add: [NET_ADMIN]` block into the new cap_add list.
4. Hardened the `certbot` service: no-new-privileges, cap_drop ALL, cap_add NET_BIND_SERVICE, read_only with tmpfs for /tmp and /run.

## Verification

- `bash scripts/verify-container-security.sh` — cloud-relay services (relay, caddy) and certbot all PASS (9/9 checks). Home-device services fail as expected (27 failures — T02 scope). Dockerfile checks: 2 PASS (cloud-relay, profiles), 3 FAIL (T03 scope). Total: 11 pass, 30 fail.
- `docker compose config --quiet` — Docker CLI not available on dev machine; compose YAML uses only standard v3.8 directives, validated by script parsing.

## Diagnostics

Run `bash scripts/verify-container-security.sh` at any time to audit security posture across all compose files and Dockerfiles. Output shows per-service PASS/FAIL with specific missing directives. Exit code 1 if any check fails.

## Deviations

- `docker compose config --quiet` could not be run (Docker not installed on dev machine). The compose files use only standard directives and were validated by the verification script's parsing.

## Known Issues

None.

## Files Created/Modified

- `scripts/verify-container-security.sh` — new security audit script checking 3 compose files and 5 Dockerfiles
- `cloud-relay/docker-compose.yml` — added security directives to relay and caddy services, consolidated cap_add
- `cloud-relay/certbot/docker-compose.certbot.yml` — added security directives to certbot service
