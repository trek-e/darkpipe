# S01: Infrastructure Validation — DNS, TLS & Tunnel — Research

**Date:** 2026-03-12

## Summary

This slice validates the foundational infrastructure that all subsequent M005 slices depend on: DNS records resolving correctly from external resolvers, TLS certificates trusted by clients, WireGuard/mTLS tunnel stable between cloud relay and home device, and required ports (25, 587, 993, 443) responding from the public internet.

The codebase already has comprehensive tooling for every validation area. The `dns/validator/` package validates SPF, DKIM, DMARC, MX, SRV, and autodiscover CNAMEs against external resolvers (Google, Cloudflare, OpenDNS). The `deploy/test/tunnel-test.sh` script validates WireGuard handshake age and mTLS handshake success. The `deploy/test/outage-sim.sh` script simulates 60-second outages and verifies auto-recovery. Caddy handles Let's Encrypt auto-HTTPS on the cloud relay. The `deploy/setup/pkg/validate/` package provides port reachability checks. The primary work is orchestrating these existing tools into a coherent validation sequence against a live deployment, documenting results, and fixing any issues found.

The key risk is that this slice requires a live environment — real VPS, real domain, real home device on a residential network. No amount of CI or mock testing substitutes for the actual NAT traversal, DNS propagation, and TLS challenge that happen in production. The validation script/checklist must be designed to run against the user's actual infrastructure, report pass/fail clearly, and document any fixes needed.

## Recommendation

Build a single orchestration validation script (`scripts/validate-infrastructure.sh`) that runs all existing validation tools in sequence against a live deployment. The script should:

1. **DNS validation** — invoke `dns-setup --validate-only` (or use `dns/validator.Checker.CheckAll`) against external resolvers for all record types (MX, A, SPF, DKIM, DMARC, SRV, autoconfig/autodiscover CNAMEs)
2. **TLS validation** — check Let's Encrypt certificates on the cloud relay (Caddy endpoints on 443), verify chain validity and expiry with `openssl s_client`
3. **Tunnel validation** — run `deploy/test/tunnel-test.sh` for the configured transport (WireGuard or mTLS), verify handshake freshness and connectivity
4. **Port reachability** — check ports 25, 587, 993, 443 from an external vantage point (use the cloud relay's public IP)
5. **Tunnel stability** — run `deploy/test/outage-sim.sh` for a brief interruption test and verify auto-recovery

Do NOT hand-roll DNS queries, TLS checks, or tunnel health checks — the existing codebase already has all of this. The new work is glue and orchestration.

## Don't Hand-Roll

| Problem | Existing Solution | Why Use It |
|---------|------------------|------------|
| DNS record validation (SPF/DKIM/DMARC/MX/SRV/CNAME) | `dns/validator/checker.go` + `srv.go` | Full validation suite with external resolver support, structured results, multi-server failover |
| DNS setup and dry-run | `dns/cmd/dns-setup/main.go` (`--validate-only` flag) | Complete CLI tool for DNS validation with colored output and JSON mode |
| Port reachability | `deploy/setup/pkg/validate/ports.go` | TCP dial with timeout, already tested |
| WireGuard tunnel health | `transport/wireguard/monitor/health.go` + `deploy/test/tunnel-test.sh` | Handshake age check, connectivity test, peer enumeration |
| mTLS tunnel health | `deploy/test/tunnel-test.sh` (mtls mode) | Certificate path check, TLS handshake verification, expiry check |
| Tunnel outage recovery | `deploy/test/outage-sim.sh` | 60-second disconnect + auto-recovery verification (matches Phase 1 success criteria) |
| WireGuard reconnection | `transport/wireguard/monitor/reconnect.sh` | systemd-based restart on stale handshakes, post-restart verification |
| TLS certificate monitoring | `cloud-relay/relay/tls/monitor.go` | Postfix log TLS event detection and notification |
| Let's Encrypt renewal | `cloud-relay/certbot/docker-compose.certbot.yml` | Standalone certbot sidecar with 12-hour renewal loop |
| WireGuard config generation | `transport/wireguard/config/templates.go` | Cloud and home config templates with sensible defaults (10.8.0.0/24, keepalive=25) |
| WireGuard key generation | `transport/wireguard/keygen/keygen.go` | Secure key pair generation |

## Existing Code and Patterns

- `dns/validator/checker.go` — Core DNS validation with `CheckAll()` method that validates SPF, DKIM, DMARC, MX in one call. Returns structured `ValidationReport` with pass/fail per record. Uses external resolvers (8.8.8.8, 1.1.1.1).
- `dns/validator/srv.go` — SRV record validation (`_imaps._tcp`, `_submission._tcp`) and autodiscover CNAME validation (`autoconfig.*`, `autodiscover.*`). These are NOT called by `CheckAll()` — must be invoked separately.
- `dns/cmd/dns-setup/main.go` — Full CLI with `--validate-only` mode. Uses env vars `DARKPIPE_DOMAIN`, `RELAY_HOSTNAME`, `RELAY_IP`. Also supports `--json` for structured output.
- `deploy/setup/pkg/validate/dns.go` — Simpler DNS validation (MX + A/AAAA + SPF + DKIM + DMARC). Separate from `dns/validator/` — this is in the setup wizard path. Uses same external resolvers.
- `deploy/setup/pkg/validate/ports.go` — `ValidatePort(hostname, port, timeout)` for TCP reachability checks.
- `deploy/test/tunnel-test.sh` — Accepts `wireguard` or `mtls` as first arg. WireGuard mode checks: interface exists, handshake freshness (5min max), ping through tunnel. mTLS mode checks: server listening, cert files exist, TLS handshake with mutual auth, cert expiry.
- `deploy/test/outage-sim.sh` — Requires root. Simulates home internet outage for 60 seconds, then verifies tunnel auto-recovery. Uses `tunnel-test.sh` for pre/post verification.
- `transport/wireguard/monitor/reconnect.sh` — Cron/systemd safety net that restarts `wg-quick@wg0` if handshake age exceeds threshold. Already deployed via `deploy/wireguard/` setup scripts.
- `cloud-relay/caddy/Caddyfile` — Caddy handles auto-HTTPS for webmail, autoconfig, and autodiscover subdomains. Routes to home device (10.0.0.2) through tunnel. Uses env vars for domain configuration.
- `cloud-relay/docker-compose.yml` — Cloud relay: Caddy (80, 443) + relay daemon (25). Caddy uses `caddy-data` volume for Let's Encrypt certs.
- `home-device/docker-compose.yml` — Home device: mail server (25, 587, 993) + webmail (8080) + profile server (8090) + rspamd + redis. Profile-based selection of mail server variant.
- `deploy/wireguard/cloud-setup.sh` — Installs WireGuard, enables systemd service with auto-restart (30s delay), opens UDP 51820, enables IP forwarding.
- `deploy/wireguard/home-setup.sh` — Installs WireGuard, verifies PersistentKeepalive in config, enables systemd auto-restart. No firewall changes needed (NAT traversal is outbound-only).
- `scripts/check-runtime.sh` — Pre-flight check for container runtime, version, compose tool, SELinux, port 25. Has `--ci` flag to skip network checks.

## Constraints

- **Live infrastructure required** — all validation requires a real VPS with public IP, real domain with DNS control, and real home device on a residential network. Cannot be simulated in CI.
- **WireGuard is default transport** — decision recorded; mTLS is secondary. Validate whichever the user has configured.
- **Caddy manages TLS** — Let's Encrypt certificates are obtained and renewed by Caddy (auto-HTTPS) on the cloud relay. The certbot sidecar is an alternative for edge cases. Do not duplicate TLS management.
- **Port 80 must be open for HTTP-01 challenge** — Caddy's auto-HTTPS uses HTTP-01. If port 80 is blocked, DNS-01 challenge is needed (not yet automated).
- **Home device uses tunnel IP (10.0.0.2)** — Caddy reverse_proxies to `10.0.0.2:*` through the WireGuard tunnel. This IP is hardcoded in the Caddyfile.
- **DNS propagation delay** — records may not be visible from all resolvers immediately. TTL and caching affect validation timing. The validator already retries multiple servers but doesn't wait for propagation.
- **Root required for outage simulation** — `outage-sim.sh` needs root to stop/start systemd services.
- **SRV and CNAME checks are separate from CheckAll()** — `dns/validator/checker.go` `CheckAll()` only covers SPF, DKIM, DMARC, MX. SRV (`CheckSRV`) and autodiscover CNAMEs (`CheckAutodiscoverCNAMEs`) must be called separately. The orchestration script must invoke all validators.
- **Port 25 on VPS** — ISP/cloud provider must allow outbound AND inbound port 25. `check-runtime.sh` validates this but only locally. External reachability needs a separate check.
- **Self-signed TLS for IMAP/submission** — Per decision: "Self-signed TLS certificates acceptable for IMAP/submission (traffic within WireGuard tunnel)." External clients connect through the tunnel, not directly. Only the Caddy HTTPS endpoint needs a publicly trusted cert.

## Common Pitfalls

- **DNS propagation vs validation timing** — Running DNS validation immediately after record creation will fail. The validator uses external resolvers (Google, Cloudflare) which cache. Include a propagation wait or retry-with-delay strategy in the orchestration script. Typical propagation: 5-60 minutes for most records.
- **Caddy auto-HTTPS requires correct domain env vars** — If `WEBMAIL_DOMAINS`, `AUTOCONFIG_DOMAINS`, `AUTODISCOVER_DOMAINS` aren't set correctly in the cloud relay compose, Caddy won't request certs for the right domains. Validate env vars before checking TLS.
- **Multiple SPF records** — The validator correctly detects this (RFC 7208 violation), but it's a common mistake when adding records. The orchestration should flag this prominently.
- **WireGuard handshake age on fresh tunnel** — A newly started tunnel may not have a handshake yet. The tunnel test treats "no handshake recorded" as a failure. Allow a warm-up period (30s) after starting the tunnel before validating.
- **Port 587/993 not reachable from outside** — These ports are on the HOME device, not the cloud relay. External clients reach them through Caddy (HTTPS proxying for webmail) or directly through the tunnel IP. Port reachability checks must target the correct host — the cloud relay's public IP for ports 25/80/443, but 587/993 may need to be tested differently depending on whether Caddy is proxying them or if they're exposed through the tunnel.
- **Caddy reverse proxy hardcodes 10.0.0.2** — If the tunnel uses different addressing (default is 10.8.0.0/24 in WireGuard templates, but Caddyfile uses 10.0.0.2), there's a mismatch. Verify tunnel addressing matches Caddyfile expectations. This is a potential configuration gap.
- **Certbot sidecar port conflict with Caddy** — Both `docker-compose.certbot.yml` and main `docker-compose.yml` bind port 80. They should not run simultaneously. Caddy handles its own ACME challenges, so the certbot sidecar is only for non-Caddy setups.

## Open Risks

- **Tunnel IP address mismatch** — WireGuard templates use `10.8.0.0/24` (cloud=10.8.0.1, home=10.8.0.2), but the Caddyfile uses `10.0.0.2` for reverse proxy targets. This is a configuration inconsistency that will cause Caddy to fail to reach home device services. Must be reconciled during validation.
- **IMAP/SMTP external access path unclear** — Ports 587 (SMTP submission) and 993 (IMAPS) are bound on the home device, not the cloud relay. External clients need to reach these somehow. The Caddyfile only proxies HTTP/HTTPS, not raw TCP. This means either: (a) IMAP/SMTP clients connect directly to the home device (breaks when home IP changes / NAT), (b) there's a TCP proxy on the cloud relay not yet configured, or (c) clients use the tunnel. This needs investigation — it's a potential gap in the architecture for external device access.
- **DNS-01 challenge fallback not automated** — If port 80 is blocked (e.g., ISP blocking), Caddy's HTTP-01 challenge fails. DNS-01 via Cloudflare API is possible but not configured. May need a `caddy-dns/cloudflare` plugin.
- **No external port scan capability in codebase** — `ValidatePort()` does a TCP dial, but if run from the home network, it tests internal connectivity, not external reachability. True external validation requires running from a different network or using an external service.
- **Home router NAT/firewall may block WireGuard UDP** — The home setup script expects outbound UDP 51820 to work (NAT traversal via PersistentKeepalive). Most residential NATs allow this, but some restrictive firewalls or ISP-level NAT (CGNAT) could interfere. No automated detection for CGNAT.

## Skills Discovered

| Technology | Skill | Status |
|------------|-------|--------|
| WireGuard tunnel | — | none found |
| Let's Encrypt / TLS | — | none found (generic results only) |
| DNS email auth (SPF/DKIM/DMARC) | `sickn33/antigravity-awesome-skills@email-systems` | available (281 installs) — covers email systems broadly |
| Cloudflare DNS | `cloudflare/skills@sandbox-sdk` | available (1K installs) — Cloudflare Workers SDK, not DNS management |
| Cloudflare troubleshooting | `cloudflare-troubleshooting` | installed (local skill) |

No skills are directly relevant to this infrastructure validation work. The `cloudflare-troubleshooting` skill (already installed) may be useful if DNS issues arise during validation, but the primary work is orchestrating existing DarkPipe validation tools against a live deployment.

## Sources

- WireGuard config defaults (10.8.0.0/24, keepalive=25, listen 51820) from `transport/wireguard/config/templates.go`
- DNS validation capabilities from `dns/validator/checker.go` and `dns/validator/srv.go`
- Caddy reverse proxy targets (10.0.0.2) from `cloud-relay/caddy/Caddyfile`
- Let's Encrypt via Caddy auto-HTTPS from `cloud-relay/docker-compose.yml`
- Tunnel test capabilities from `deploy/test/tunnel-test.sh` and `deploy/test/outage-sim.sh`
- Port validation from `deploy/setup/pkg/validate/ports.go`
- Decision: "Self-signed TLS certificates acceptable for IMAP/submission" from `.gsd/DECISIONS.md`
- Decision: "Health check threshold: 5 minutes max handshake age" from `.gsd/DECISIONS.md`
