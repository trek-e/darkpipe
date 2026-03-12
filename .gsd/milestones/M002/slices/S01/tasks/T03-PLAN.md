---
estimated_steps: 4
estimated_files: 3
---

# T03: Add HEALTHCHECK to Dockerfiles missing them and run full verification

**Slice:** S01 — Container Security Hardening
**Milestone:** M002

## Description

Three custom Dockerfiles (postfix-dovecot, stalwart, maddy) lack HEALTHCHECK instructions. Add them for standalone `docker run` usage and defense-in-depth. Then run full verification to confirm the entire slice is complete.

## Steps

1. Add HEALTHCHECK to `home-device/postfix-dovecot/Dockerfile`:
   - `HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 CMD nc -z localhost 25 || exit 1`
   - Place before CMD/ENTRYPOINT, after EXPOSE
   - Verify `netcat-openbsd` or equivalent is already installed (bash is present, check for nc)
   - If nc not available, add `apk add --no-cache netcat-openbsd` to the existing RUN layer
2. Add HEALTHCHECK to `home-device/stalwart/Dockerfile`:
   - `HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 CMD nc -z localhost 25 || exit 1`
   - Stalwart base image is Debian-based — verify nc availability or use alternative (`curl --silent --fail http://localhost:8080/ || exit 1` if HTTP management port is available)
3. Add HEALTHCHECK to `home-device/maddy/Dockerfile`:
   - `HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 CMD nc -z localhost 25 || exit 1`
   - Maddy base image may not have nc — check and use alternative if needed (e.g. `/bin/sh -c 'echo > /dev/tcp/localhost/25'` or install nc)
4. Run full verification:
   - `bash scripts/verify-container-security.sh` — all checks must pass
   - `docker compose -f cloud-relay/docker-compose.yml config --quiet`
   - `docker compose -f home-device/docker-compose.yml config --quiet`
   - `docker compose -f cloud-relay/certbot/docker-compose.certbot.yml config --quiet`
   - Build test: `docker compose -f home-device/docker-compose.yml build postfix-dovecot` (the only custom-built image that changed)

## Must-Haves

- [ ] postfix-dovecot Dockerfile has HEALTHCHECK
- [ ] stalwart Dockerfile has HEALTHCHECK
- [ ] maddy Dockerfile has HEALTHCHECK
- [ ] All 5 Dockerfiles now have HEALTHCHECK (cloud-relay and profiles already had them)
- [ ] Full verification script passes with zero failures
- [ ] `docker compose build` succeeds for postfix-dovecot

## Verification

- `bash scripts/verify-container-security.sh` exits 0 with all checks passing
- `grep -l HEALTHCHECK home-device/postfix-dovecot/Dockerfile home-device/stalwart/Dockerfile home-device/maddy/Dockerfile` returns all 3 files
- `docker compose -f home-device/docker-compose.yml build postfix-dovecot` exits 0

## Observability Impact

- Signals added/changed: HEALTHCHECK in 3 Dockerfiles provides container health status via `docker inspect` and `docker ps`
- How a future agent inspects this: `docker inspect --format='{{.State.Health.Status}}' <container>` or `docker ps` shows health column
- Failure state exposed: Container health status transitions to `unhealthy` when port 25 is unreachable

## Inputs

- `home-device/postfix-dovecot/Dockerfile` — needs HEALTHCHECK
- `home-device/stalwart/Dockerfile` — needs HEALTHCHECK
- `home-device/maddy/Dockerfile` — needs HEALTHCHECK
- `scripts/verify-container-security.sh` — from T01

## Expected Output

- `home-device/postfix-dovecot/Dockerfile` — with HEALTHCHECK instruction
- `home-device/stalwart/Dockerfile` — with HEALTHCHECK instruction
- `home-device/maddy/Dockerfile` — with HEALTHCHECK instruction
- Full verification passing — slice complete
