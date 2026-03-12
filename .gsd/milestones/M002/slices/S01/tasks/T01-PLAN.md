---
estimated_steps: 5
estimated_files: 3
---

# T01: Create security verification script and harden cloud-relay compose files

**Slice:** S01 — Container Security Hardening
**Milestone:** M002

## Description

Create the verification script that validates all compose files and Dockerfiles for required security directives. Then apply security hardening to the cloud-relay docker-compose.yml (relay + caddy services) and the certbot docker-compose.certbot.yml. The relay service is the internet-facing attack surface, making it the highest-priority target.

## Steps

1. Create `scripts/verify-container-security.sh` that:
   - Parses all 3 compose files and checks each service block for `security_opt`, `cap_drop`, and `read_only`
   - Parses all 5 Dockerfiles and checks for HEALTHCHECK instruction
   - Outputs per-service PASS/FAIL with specific missing directives
   - Exits non-zero if any check fails
2. Add security directives to the `relay` service in `cloud-relay/docker-compose.yml`:
   - `security_opt: [no-new-privileges:true]`
   - `cap_drop: [ALL]`
   - `cap_add: [NET_ADMIN, NET_BIND_SERVICE, DAC_OVERRIDE, CHOWN, SETGID, SETUID, KILL]` (keep existing NET_ADMIN, add others for Postfix)
   - `read_only: true`
   - `tmpfs: [/var/spool/postfix, /var/lib/postfix, /tmp, /run, /var/run]`
   - Add comment: `# Root required: Postfix master process needs root for queue management, port 25 binding, and setuid helpers`
   - Remove the old standalone `cap_add: [NET_ADMIN]` block (now consolidated)
3. Add security directives to the `caddy` service:
   - `security_opt: [no-new-privileges:true]`
   - `cap_drop: [ALL]`
   - `cap_add: [NET_BIND_SERVICE]`
   - `read_only: true`
   - `tmpfs: [/tmp, /run]`
4. Add security directives to the `certbot` service in `cloud-relay/certbot/docker-compose.certbot.yml`:
   - `security_opt: [no-new-privileges:true]`
   - `cap_drop: [ALL]`
   - `cap_add: [NET_BIND_SERVICE]`
   - `read_only: true`
   - `tmpfs: [/tmp, /run]`
5. Validate: run `docker compose -f cloud-relay/docker-compose.yml config --quiet` and `bash scripts/verify-container-security.sh` (expecting partial pass — home-device not yet hardened)

## Must-Haves

- [ ] Verification script checks all 3 compose files and all 5 Dockerfiles
- [ ] Verification script exits non-zero on any failure with clear per-service output
- [ ] Relay service has cap_drop ALL + selective cap_add including NET_ADMIN for WireGuard
- [ ] Caddy and certbot services have security directives
- [ ] All modified compose files pass `docker compose config --quiet`
- [ ] Root justification comment on relay service

## Verification

- `bash scripts/verify-container-security.sh` passes for cloud-relay and certbot files (home-device expected to fail)
- `docker compose -f cloud-relay/docker-compose.yml config --quiet` exits 0
- `docker compose -f cloud-relay/certbot/docker-compose.certbot.yml config --quiet` exits 0

## Observability Impact

- Signals added/changed: None (static config changes)
- How a future agent inspects this: Run `bash scripts/verify-container-security.sh` to audit security posture
- Failure state exposed: Script prints per-service PASS/FAIL with specific missing directives

## Inputs

- `cloud-relay/docker-compose.yml` — current compose without security directives (except cap_add NET_ADMIN)
- `cloud-relay/certbot/docker-compose.certbot.yml` — current certbot compose without security directives
- S01-RESEARCH.md — capability requirements per service

## Expected Output

- `scripts/verify-container-security.sh` — reusable security audit script
- `cloud-relay/docker-compose.yml` — hardened with security directives on relay and caddy
- `cloud-relay/certbot/docker-compose.certbot.yml` — hardened with security directives on certbot
