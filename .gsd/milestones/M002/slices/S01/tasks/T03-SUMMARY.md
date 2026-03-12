---
id: T03
parent: S01
milestone: M002
provides:
  - HEALTHCHECK instructions in all 5 custom Dockerfiles
key_files:
  - home-device/postfix-dovecot/Dockerfile
  - home-device/stalwart/Dockerfile
  - home-device/maddy/Dockerfile
key_decisions:
  - stalwart uses curl HTTP healthcheck on :8080 (management API) instead of nc on :25, with curl installed via apt-get
  - postfix-dovecot and maddy use nc -z localhost 25 (nc available via busybox on Alpine)
patterns_established:
  - HEALTHCHECK format matches existing cloud-relay/profiles pattern: --interval=30s --timeout=10s --start-period=10s --retries=3
observability_surfaces:
  - Container health status via `docker inspect --format='{{.State.Health.Status}}' <container>` or `docker ps` health column
duration: 10m
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T03: Added HEALTHCHECK to postfix-dovecot, stalwart, and maddy Dockerfiles

**All 5 custom Dockerfiles now have HEALTHCHECK instructions; full security verification passes 41/41 checks.**

## What Happened

Added HEALTHCHECK instructions to the three Dockerfiles that were missing them:

1. **postfix-dovecot** (Alpine 3.21): `nc -z localhost 25` — busybox nc is available, checks SMTP port
2. **stalwart** (Debian-based): Installed `curl` via apt-get and uses `curl --silent --fail http://localhost:8080/` — checks the HTTP management interface rather than raw SMTP, since curl is more reliable on Debian and stalwart exposes an HTTP API
3. **maddy** (foxcpp/maddy, Alpine-based): `nc -z localhost 25` — checks SMTP port

All three use the same timing parameters as the existing cloud-relay and profiles HEALTHCHECKs: `--interval=30s --timeout=10s --start-period=10s --retries=3`.

## Verification

- `grep -l HEALTHCHECK` on all 3 Dockerfiles: returns all 3 ✅
- `bash scripts/verify-container-security.sh`: 41/41 checks pass, exit 0 ✅
- `docker compose config --quiet` and `docker compose build`: Docker not available on this dev machine — compose syntax and build verification deferred to CI/deployment environment
- All 5 Dockerfiles confirmed to have HEALTHCHECK (cloud-relay and profiles already had them)

## Diagnostics

- Run `bash scripts/verify-container-security.sh` to re-audit all security directives and HEALTHCHECKs
- Run `docker inspect --format='{{.State.Health.Status}}' <container>` to check container health at runtime
- `docker ps` shows health column for all containers with HEALTHCHECK

## Deviations

- stalwart uses `curl` HTTP healthcheck on `:8080` instead of `nc -z localhost 25` as originally planned — curl is more appropriate for Debian-based images and tests the actual management API
- Installed `curl` package in stalwart Dockerfile (small addition to support healthcheck)
- Docker compose validation and build tests could not run — Docker not installed on this machine

## Known Issues

None.

## Files Created/Modified

- `home-device/postfix-dovecot/Dockerfile` — added HEALTHCHECK (nc -z localhost 25)
- `home-device/stalwart/Dockerfile` — installed curl, added HEALTHCHECK (curl on :8080)
- `home-device/maddy/Dockerfile` — added HEALTHCHECK (nc -z localhost 25)
