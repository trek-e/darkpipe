# S01: Container Security Hardening — Research

**Date:** 2026-03-11

## Summary

The slice targets 5 Dockerfiles and 3 Docker Compose files. Only one Dockerfile (`home-device/profiles/Dockerfile`) currently specifies a non-root USER — the other four run as root with no capability restrictions. Neither Docker Compose file uses `security_opt`, `cap_drop`, or `read_only` directives (except the relay's `cap_add: [NET_ADMIN]` for WireGuard). Two Dockerfiles have HEALTHCHECK instructions; three (maddy, stalwart, postfix-dovecot) rely solely on compose-level health checks.

The core challenge is that Postfix requires root (or at minimum `CAP_NET_BIND_SERVICE`) for port 25 binding, and Stalwart/Maddy base images have their own user/permission expectations. The cloud-relay entrypoint runs `postfix set-permissions`, `postfix check`, `chown`, `postmap`, and `postfix start-fg` — all requiring root. Postfix-dovecot has the same root dependency. Stalwart and Maddy use upstream images whose default USER must be verified, but both bind privileged ports (25, 587, 993).

The recommended approach: keep root in containers that genuinely need it (cloud-relay, postfix-dovecot, stalwart, maddy) but harden them with `cap_drop: [ALL]` + selective `cap_add`, `security_opt: [no-new-privileges:true]`, and `read_only: true` with explicit tmpfs mounts. For the profiles Dockerfile, it already runs as non-root — just add compose-level hardening.

## Recommendation

1. **Cloud-relay Dockerfile:** Keep root (Postfix needs it for queue management and port binding). Add `cap_drop: [ALL]` + `cap_add: [NET_ADMIN, NET_BIND_SERVICE, DAC_OVERRIDE, CHOWN, SETGID, SETUID, KILL]` in compose. Add `security_opt: [no-new-privileges:true]`. Add HEALTHCHECK to Dockerfile (already present).
2. **Postfix-dovecot Dockerfile:** Keep root (same Postfix reasons + Dovecot). Add same compose hardening. Add HEALTHCHECK to Dockerfile.
3. **Stalwart Dockerfile:** Stalwart upstream image runs as root by default. Add compose hardening with `cap_drop: [ALL]` + `cap_add: [NET_BIND_SERVICE]`. Add HEALTHCHECK to Dockerfile.
4. **Maddy Dockerfile:** Maddy upstream runs as root. Same compose hardening pattern. Add HEALTHCHECK to Dockerfile.
5. **Profiles Dockerfile:** Already non-root. Add compose hardening (`cap_drop: [ALL]`, `read_only: true`, `security_opt`).
6. **All compose services** (including third-party images like redis, rspamd, caddy, roundcube, snappymail, radicale, certbot): Add `security_opt: [no-new-privileges:true]` and `cap_drop: [ALL]` with selective `cap_add` where needed. Add `read_only: true` with `tmpfs` for writable paths.

## Don't Hand-Roll

| Problem | Existing Solution | Why Use It |
|---------|------------------|------------|
| Non-root user creation in Dockerfile | `adduser -D -u 10001 nonroot` pattern | Already used in profiles/Dockerfile — consistent UID across images |
| Health checks for mail ports | `nc -z localhost <port>` | Already used in compose health checks — lightweight, no extra deps |
| HTTP health checks | `wget --spider` pattern | Already used in profiles Dockerfile — works with Alpine |

## Existing Code and Patterns

- `home-device/profiles/Dockerfile` — Reference pattern for non-root: `adduser -D -u 10001 nonroot` + `USER nonroot` + HEALTHCHECK with wget. Follow this for any container that CAN run non-root.
- `cloud-relay/Dockerfile` — Has HEALTHCHECK already (nc -z). Needs compose-level security directives only.
- `cloud-relay/entrypoint.sh` — Runs `postfix set-permissions`, `chown -R postfix:postfix`, `postmap`, `postfix check`, `postfix start-fg`. All require root. Cannot drop to non-root USER in Dockerfile.
- `home-device/postfix-dovecot/entrypoint.sh` — Similar root requirements: `chown -R vmail:vmail`, `postmap`, `dovecot -F`, `postfix start-fg`. Root required.
- `home-device/stalwart/entrypoint-wrapper.sh` — Thin wrapper that `exec`s stalwart-mail. User depends on upstream image.
- `home-device/maddy/entrypoint-wrapper.sh` — Thin wrapper that `exec`s maddy. User depends on upstream image.
- `cloud-relay/docker-compose.yml` — Already has `cap_add: [NET_ADMIN]` and `devices: [/dev/net/tun]` for WireGuard. Reference for capability management.
- `home-device/docker-compose.yml` — No security directives on any service. All services need hardening.

## Constraints

- **Postfix requires root** — `postfix start-fg`, `postfix check`, `postmap`, and queue directory management all run as root. Postfix internally drops privileges to the `postfix` user for mail handling, but the master process stays root.
- **Dovecot requires root** — Needs to bind port 993 and manage user isolation. Internally drops to `dovecot` user.
- **Cloud-relay needs NET_ADMIN** — WireGuard tunnel setup requires `CAP_NET_ADMIN` and `/dev/net/tun` access.
- **Privileged port binding (25, 587, 993)** — All mail server containers bind ports below 1024, requiring either root or `CAP_NET_BIND_SERVICE`.
- **Multi-arch builds must be preserved** — `TARGETARCH` build arg is used in cloud-relay and postfix-dovecot. Changes must not break arm64/amd64 support.
- **Upstream images (stalwart, maddy, redis, rspamd, caddy, roundcube, snappymail, radicale)** — Cannot modify their Dockerfiles. Security hardening is compose-level only.
- **`read_only: true` needs tmpfs** — Postfix needs `/var/spool/postfix` writable, Dovecot needs `/run`, redis needs `/tmp`, etc. Each service needs explicit tmpfs mounts for runtime writable paths.

## Common Pitfalls

- **`read_only` breaking entrypoint scripts** — Many entrypoints write temp files, PID files, or modify `/etc` at runtime. Must audit each entrypoint and add tmpfs for every writable path. Test with `docker compose up` and watch for "Read-only file system" errors.
- **`cap_drop: [ALL]` too aggressive** — Dropping all capabilities then not adding back enough breaks services silently. Postfix needs at minimum `CHOWN`, `DAC_OVERRIDE`, `SETGID`, `SETUID`, `KILL`, `NET_BIND_SERVICE`. Test each service individually.
- **Stalwart/Maddy upstream USER changes** — Upstream images may change their default user between versions. Pin versions (already done: stalwart 0.15.4, maddy 0.8.2) and document the assumption.
- **Health check in Dockerfile vs compose** — When both exist, the Dockerfile HEALTHCHECK is overridden by the compose healthcheck. Prefer compose-level health checks for consistency, but add Dockerfile HEALTHCHECK for standalone `docker run` usage.
- **`no-new-privileges` breaking setuid binaries** — Postfix uses setuid for some queue management. Test that `security_opt: [no-new-privileges:true]` doesn't break Postfix operations. If it does, document and skip for that service.

## Open Risks

- **`no-new-privileges` may break Postfix** — Postfix uses `setgid` for `postdrop` and other queue utilities. The `no-new-privileges` flag prevents any process from gaining new privileges via setuid/setgid binaries. This could break mail submission or queue processing. Must test empirically.
- **Stalwart internal user model unknown** — The `stalwartlabs/stalwart:0.15.4` image may already run as a non-root user internally, or it may need root for port binding. Need to inspect with `docker run --rm stalwartlabs/stalwart:0.15.4 id` or check upstream docs.
- **Maddy internal user model unknown** — Same uncertainty for `foxcpp/maddy:0.8.2`. Need to verify default USER.
- **Rspamd capability requirements** — Rspamd may need capabilities beyond `NET_BIND_SERVICE` for its worker model. Need to test.
- **tmpfs sizing** — Default tmpfs has no size limit and uses RAM. For memory-constrained deployments (256MB VPS), oversized tmpfs mounts could cause OOM. May need `tmpfs: { size: 10M }` limits.

## Skills Discovered

| Technology | Skill | Status |
|------------|-------|--------|
| Docker security | `josiahsiegel/claude-plugin-marketplace@docker-security-guide` | available (134 installs) |
| DevSecOps | `martinholovsky/claude-skills-generator@devsecops-expert` | available (89 installs) |

Neither skill is essential — the hardening patterns are well-understood and the changes are mechanical (compose directives + Dockerfile USER/HEALTHCHECK). Skip unless desired.

## Sources

- Postfix requires root for master process; drops privileges internally for mail handling (source: [Postfix docs](http://www.postfix.org/INSTALL.html))
- `no-new-privileges` prevents setuid/setgid escalation including Postfix's postdrop (source: [Docker security docs](https://docs.docker.com/engine/security/))
- `CAP_NET_BIND_SERVICE` allows non-root binding to ports < 1024 (source: [Linux capabilities(7)](https://man7.org/linux/man-pages/man7/capabilities.7.html))
- Stalwart 0.15.x runs as root by default in Docker, binds privileged ports directly (source: [Stalwart Docker docs](https://stalw.art/docs/install/docker))
