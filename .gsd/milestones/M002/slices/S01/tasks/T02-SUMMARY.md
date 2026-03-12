---
id: T02
parent: S01
milestone: M002
provides:
  - All 9 home-device compose services hardened with security directives
key_files:
  - home-device/docker-compose.yml
key_decisions:
  - postfix-dovecot gets 6 cap_add entries (NET_BIND_SERVICE + mail delivery caps) vs 1 for stalwart/maddy
  - Roundcube gets /var/tmp tmpfs in addition to /tmp and /run (PHP session files)
  - Services on ports >1024 (radicale, profile-server, redis) get no cap_add
patterns_established:
  - Root justification comments on all mail server services requiring privileged ports
observability_surfaces:
  - Run `bash scripts/verify-container-security.sh` to audit all compose services
duration: 10m
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T02: Harden all home-device compose services with security directives

**Added security_opt, cap_drop, cap_add, read_only, and tmpfs to all 9 home-device services**

## What Happened

Applied security hardening directives to all 9 services in `home-device/docker-compose.yml`:

- **Mail servers** (stalwart, maddy, postfix-dovecot): security_opt no-new-privileges, cap_drop ALL, cap_add NET_BIND_SERVICE (+ DAC_OVERRIDE/CHOWN/SETGID/SETUID/KILL for postfix-dovecot), read_only true, tmpfs for /tmp and /run (+ /var/spool/postfix and /var/lib/postfix for postfix-dovecot). Root justification comments added.
- **Webmail** (roundcube, snappymail): security_opt, cap_drop ALL, read_only true, tmpfs (roundcube adds /var/tmp for PHP).
- **Infrastructure** (rspamd, redis, radicale): security_opt, cap_drop ALL, read_only true, tmpfs. No cap_add needed (rspamd/redis don't bind privileged ports in this config, radicale on 5232).
- **profile-server**: security_opt, cap_drop ALL, read_only true, tmpfs /tmp. No cap_add (non-root, port 8090).

Directive block order follows T01 pattern: security_opt → cap_drop → cap_add → read_only → tmpfs, placed before ports.

## Verification

- `bash scripts/verify-container-security.sh` — all 9 home-device services pass (27/27 compose checks). 3 Dockerfile HEALTHCHECK failures are pre-existing and scoped to T03.
- `docker compose -f home-device/docker-compose.yml config --quiet` — docker not installed on dev machine; YAML structure validated by verification script's parser.

### Slice-level verification status:
- ✅ `bash scripts/verify-container-security.sh` — all compose checks pass (38/38); 3 Dockerfile checks fail (T03 scope)
- ⏳ `docker compose -f home-device/docker-compose.yml config --quiet` — docker not available on this machine
- ✅ `docker compose -f cloud-relay/docker-compose.yml config --quiet` — passed in T01
- ⏳ `docker compose -f cloud-relay/certbot/docker-compose.certbot.yml config --quiet` — docker not available

## Diagnostics

Run `bash scripts/verify-container-security.sh` at any time to audit security posture across all compose files.

## Deviations

None.

## Known Issues

- `docker compose config` cannot be validated on this machine (no docker installed). YAML structure is sound per the verification script parser.

## Files Created/Modified

- `home-device/docker-compose.yml` — added security directives to all 9 services
