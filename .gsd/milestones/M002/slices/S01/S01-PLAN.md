# S01: Container Security Hardening

**Goal:** All containers run with minimal privileges — non-root USER where possible, cap_drop/cap_add/security_opt/read_only in all compose services, HEALTHCHECK in all custom Dockerfiles.
**Demo:** `bash scripts/verify-container-security.sh` passes all checks against the compose files and Dockerfiles. `docker compose build` succeeds for all custom images.

## Must-Haves

- Every compose service has `security_opt: [no-new-privileges:true]` (except where empirically proven to break the service, documented)
- Every compose service has `cap_drop: [ALL]` with selective `cap_add` where needed
- Every compose service has `read_only: true` with `tmpfs` mounts for writable paths
- All 5 custom Dockerfiles have a HEALTHCHECK instruction
- Root-requiring containers (cloud-relay, postfix-dovecot, stalwart, maddy) have documented justification as code comments
- profiles Dockerfile already runs non-root — no regression

## Proof Level

- This slice proves: contract (static analysis of Dockerfiles and compose files) + build (docker build succeeds)
- Real runtime required: no (runtime validation of tmpfs/caps is deferred to integration testing in milestone verification)
- Human/UAT required: no

## Verification

- `bash scripts/verify-container-security.sh` — checks all compose files for required security directives and all Dockerfiles for HEALTHCHECK
- `docker compose -f cloud-relay/docker-compose.yml config --quiet` — validates cloud-relay compose syntax
- `docker compose -f home-device/docker-compose.yml config --quiet` — validates home-device compose syntax
- `docker compose -f cloud-relay/certbot/docker-compose.certbot.yml config --quiet` — validates certbot compose syntax

## Observability / Diagnostics

- Runtime signals: None (static hardening, no new runtime code)
- Inspection surfaces: `scripts/verify-container-security.sh` can be re-run at any time to audit security posture
- Failure visibility: Verification script prints per-service PASS/FAIL with specific missing directives
- Redaction constraints: None

## Integration Closure

- Upstream surfaces consumed: existing Dockerfiles and docker-compose files
- New wiring introduced in this slice: security directives added to existing compose services, HEALTHCHECK added to Dockerfiles
- What remains before the milestone is truly usable end-to-end: S02 (log hygiene), S03 (TLS hardening), S04 (operational quality)

## Tasks

- [x] **T01: Create security verification script and harden cloud-relay compose files** `est:45m`
  - Why: Establishes the verification harness and hardens the cloud-relay stack (relay, caddy, certbot) — the internet-facing attack surface
  - Files: `scripts/verify-container-security.sh`, `cloud-relay/docker-compose.yml`, `cloud-relay/certbot/docker-compose.certbot.yml`
  - Do: Write a bash script that greps all compose files for `security_opt`, `cap_drop`, `read_only` on every service and checks all Dockerfiles for HEALTHCHECK. Then add `security_opt: [no-new-privileges:true]`, `cap_drop: [ALL]`, selective `cap_add`, and `read_only: true` with `tmpfs` to relay, caddy, and certbot services. Relay needs `cap_add: [NET_ADMIN, NET_BIND_SERVICE, DAC_OVERRIDE, CHOWN, SETGID, SETUID, KILL]` (Postfix + WireGuard). Caddy needs `cap_add: [NET_BIND_SERVICE]`. Certbot needs `cap_add: [NET_BIND_SERVICE]`. Add comments documenting root justification for relay.
  - Verify: `bash scripts/verify-container-security.sh` passes for cloud-relay and certbot compose files; `docker compose -f cloud-relay/docker-compose.yml config --quiet` exits 0
  - Done when: All cloud-relay compose services have full security directives and verification script reports PASS for them

- [x] **T02: Harden all home-device compose services with security directives** `est:45m`
  - Why: The home-device compose has 9 services (3 mail servers, 2 webmail, radicale, rspamd, redis, profile-server) — all need security hardening
  - Files: `home-device/docker-compose.yml`
  - Do: Add `security_opt: [no-new-privileges:true]`, `cap_drop: [ALL]`, selective `cap_add`, `read_only: true`, and `tmpfs` mounts to all 9 services. Mail servers (stalwart, maddy, postfix-dovecot) need `cap_add: [NET_BIND_SERVICE]` at minimum; postfix-dovecot also needs `[CHOWN, DAC_OVERRIDE, SETGID, SETUID, KILL]`. Redis needs `tmpfs: [/tmp]`. Rspamd needs `tmpfs: [/tmp, /run]`. Webmail services need `tmpfs: [/tmp, /run]`. Profile-server needs no cap_add (non-root, port >1024). Add root justification comments on mail server services.
  - Verify: `bash scripts/verify-container-security.sh` passes for home-device compose; `docker compose -f home-device/docker-compose.yml config --quiet` exits 0
  - Done when: All 9 home-device services have full security directives and verification script reports PASS

- [x] **T03: Add HEALTHCHECK to Dockerfiles missing them and run full verification** `est:30m`
  - Why: Three Dockerfiles (postfix-dovecot, stalwart, maddy) lack HEALTHCHECK instructions — needed for standalone `docker run` usage and defense-in-depth
  - Files: `home-device/postfix-dovecot/Dockerfile`, `home-device/stalwart/Dockerfile`, `home-device/maddy/Dockerfile`
  - Do: Add `HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 CMD nc -z localhost 25 || exit 1` to postfix-dovecot and maddy Dockerfiles (both have netcat or can use it). For stalwart, add HEALTHCHECK using `nc -z localhost 25` (stalwart base image includes common tools). Confirm profiles Dockerfile already has HEALTHCHECK (no change). Confirm cloud-relay Dockerfile already has HEALTHCHECK (no change). Run full verification.
  - Verify: `bash scripts/verify-container-security.sh` passes all checks; `docker compose -f home-device/docker-compose.yml build postfix-dovecot` succeeds; all Dockerfiles contain HEALTHCHECK
  - Done when: All 5 Dockerfiles have HEALTHCHECK, full verification script passes, docker build succeeds for custom images

## Files Likely Touched

- `scripts/verify-container-security.sh` (new)
- `cloud-relay/docker-compose.yml`
- `cloud-relay/certbot/docker-compose.certbot.yml`
- `home-device/docker-compose.yml`
- `home-device/postfix-dovecot/Dockerfile`
- `home-device/stalwart/Dockerfile`
- `home-device/maddy/Dockerfile`
