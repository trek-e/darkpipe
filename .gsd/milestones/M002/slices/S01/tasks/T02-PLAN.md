---
estimated_steps: 4
estimated_files: 1
---

# T02: Harden all home-device compose services with security directives

**Slice:** S01 — Container Security Hardening
**Milestone:** M002

## Description

Add security directives to all 9 services in the home-device docker-compose.yml. This is the largest single file change — each service needs cap_drop, selective cap_add, security_opt, read_only, and tmpfs mounts tailored to its runtime requirements.

## Steps

1. Add security directives to the 3 mail server services (stalwart, maddy, postfix-dovecot):
   - All three: `security_opt: [no-new-privileges:true]`, `cap_drop: [ALL]`, `read_only: true`
   - stalwart: `cap_add: [NET_BIND_SERVICE]`, `tmpfs: [/tmp, /run]`
   - maddy: `cap_add: [NET_BIND_SERVICE]`, `tmpfs: [/tmp, /run]`
   - postfix-dovecot: `cap_add: [NET_BIND_SERVICE, DAC_OVERRIDE, CHOWN, SETGID, SETUID, KILL]`, `tmpfs: [/tmp, /run, /var/spool/postfix, /var/lib/postfix]`
   - Add root justification comments on each mail server service
2. Add security directives to webmail services (roundcube, snappymail):
   - Both: `security_opt: [no-new-privileges:true]`, `cap_drop: [ALL]`, `read_only: true`
   - roundcube: `tmpfs: [/tmp, /run, /var/tmp]` (PHP needs /tmp and /var/tmp)
   - snappymail: `tmpfs: [/tmp, /run]`
3. Add security directives to infrastructure services (rspamd, redis, radicale):
   - All three: `security_opt: [no-new-privileges:true]`, `cap_drop: [ALL]`, `read_only: true`
   - rspamd: `tmpfs: [/tmp, /run]`
   - redis: `tmpfs: [/tmp]`
   - radicale: no cap_add needed (port >1024), `tmpfs: [/tmp]`
4. Add security directives to profile-server:
   - `security_opt: [no-new-privileges:true]`, `cap_drop: [ALL]`, `read_only: true`, `tmpfs: [/tmp]`
   - No cap_add needed (non-root user, port 8090 > 1024)

## Must-Haves

- [x] All 9 services have `security_opt: [no-new-privileges:true]`
- [x] All 9 services have `cap_drop: [ALL]`
- [x] All 9 services have `read_only: true`
- [x] All 9 services have `tmpfs` mounts for their writable paths
- [x] Mail server services have appropriate `cap_add` for privileged port binding
- [x] Root justification comments on stalwart, maddy, postfix-dovecot services
- [x] Compose file validates with `docker compose config`

## Verification

- `docker compose -f home-device/docker-compose.yml config --quiet` exits 0
- `bash scripts/verify-container-security.sh` passes for home-device compose

## Observability Impact

- Signals added/changed: None
- How a future agent inspects this: Run verification script
- Failure state exposed: None

## Inputs

- `home-device/docker-compose.yml` — current compose without security directives
- `scripts/verify-container-security.sh` — from T01, validates directives
- S01-RESEARCH.md — capability requirements per service

## Expected Output

- `home-device/docker-compose.yml` — all 9 services hardened with security directives
