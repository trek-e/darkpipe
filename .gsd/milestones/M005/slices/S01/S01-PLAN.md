# S01: Infrastructure Validation — DNS, TLS & Tunnel

**Goal:** Prove DNS, TLS, tunnel, and port reachability work from the public internet by building a single orchestration validation script that exercises all existing validation tools against a live deployment.
**Demo:** Run `scripts/validate-infrastructure.sh` against the live deployment and see all checks pass — DNS records resolve from external resolvers, TLS certs are valid, WireGuard tunnel is stable, and required ports respond.

## Must-Haves

- DNS validation covers all record types: MX, A, SPF, DKIM, DMARC, SRV, autoconfig/autodiscover CNAMEs
- TLS certificate validation on cloud relay HTTPS endpoints (chain, expiry, trust)
- WireGuard/mTLS tunnel health verified (handshake freshness, connectivity)
- Port reachability confirmed: 25, 443 on cloud relay; 587, 993 reachable through tunnel
- Tunnel stability verified via outage simulation with auto-recovery
- Tunnel IP address mismatch between Caddyfile (10.0.0.2) and WireGuard templates (10.8.0.0/24) reconciled
- Clear pass/fail output with structured JSON results for agent consumption
- All issues found documented with fixes applied

## Proof Level

- This slice proves: contract + operational (DNS/TLS contract checks + tunnel stability operational test)
- Real runtime required: yes (live VPS, real domain, real home device on residential network)
- Human/UAT required: no (all checks are automated; human runs the script but results are machine-verifiable)

## Verification

- `bash scripts/validate-infrastructure.sh --json 2>/dev/null | jq '.overall_status'` returns `"pass"`
- `bash scripts/validate-infrastructure.sh --json 2>/dev/null | jq '.sections | keys'` returns all 5 sections: dns, tls, tunnel, ports, stability
- Each section has `.status` of `"pass"` or `"skip"` (skip only when infrastructure not reachable, with reason)
- `bash scripts/validate-infrastructure.sh --dry-run` exits 0 (validates script logic without live infrastructure)
- Tunnel IP mismatch is fixed: `grep '10.0.0.2' cloud-relay/caddy/Caddyfile` returns nothing (uses env var or 10.8.0.2)

## Observability / Diagnostics

- Runtime signals: JSON output with per-check pass/fail, timestamps, error messages, and section summaries
- Inspection surfaces: `scripts/validate-infrastructure.sh --json` produces machine-readable results; `--verbose` shows detailed check output; `--dry-run` validates script without live infra
- Failure visibility: each check reports: check name, status (pass/fail/skip), error detail, and suggested fix
- Redaction constraints: no secrets in output; DKIM key shown as truncated hash only

## Integration Closure

- Upstream surfaces consumed: `dns/validator/` (DNS checks), `deploy/test/tunnel-test.sh` (tunnel health), `deploy/test/outage-sim.sh` (stability), `deploy/setup/pkg/validate/ports.go` (port reachability), `dns/cmd/dns-setup/` (CLI validation)
- New wiring introduced in this slice: orchestration script (`scripts/validate-infrastructure.sh`) that sequences all existing validators; Caddyfile tunnel IP fix
- What remains before the milestone is truly usable end-to-end: S02 (email round-trip delivery), S03 (device connectivity and mobile onboarding)

## Tasks

- [x] **T01: Fix tunnel IP mismatch and create validation script skeleton with dry-run mode** `est:45m`
  - Why: The Caddyfile hardcodes `10.0.0.2` but WireGuard templates use `10.8.0.0/24` — this will break Caddy reverse proxy. The validation script skeleton establishes the output format and dry-run capability that all subsequent tasks build on.
  - Files: `cloud-relay/caddy/Caddyfile`, `scripts/validate-infrastructure.sh`
  - Do: Fix Caddyfile to use `{$HOME_DEVICE_IP:10.8.0.2}` env var with correct default matching WireGuard subnet. Create validation script with argument parsing (`--json`, `--verbose`, `--dry-run`), section runner framework, JSON output structure, and dry-run mode that validates script logic with mock results. Source config from env vars (`DARKPIPE_DOMAIN`, `RELAY_HOSTNAME`, `RELAY_IP`, `TRANSPORT_TYPE`).
  - Verify: `bash scripts/validate-infrastructure.sh --dry-run` exits 0 with valid JSON; `grep '10.0.0.2' cloud-relay/caddy/Caddyfile` returns empty; Caddyfile uses env var with 10.8.0.2 default.
  - Done when: dry-run produces valid JSON with all 5 section placeholders and Caddyfile tunnel IP is reconciled

- [x] **T02: Implement DNS validation section using existing dns-setup and validator packages** `est:30m`
  - Why: DNS is the foundation — if records don't resolve externally, nothing else works. This task wires existing `dns-setup --validate-only --json` and the SRV/CNAME validators into the orchestration script.
  - Files: `scripts/validate-infrastructure.sh`, `scripts/lib/validate-dns.sh`
  - Do: Create `scripts/lib/validate-dns.sh` sourced by main script. Run `dns-setup --validate-only --json` for core records (MX, A, SPF, DKIM, DMARC). Separately invoke SRV and autodiscover CNAME checks via `dig` against external resolvers (8.8.8.8, 1.1.1.1) since SRV/CNAME checks aren't in the CLI's validate-only mode. Aggregate results into the JSON section. Handle DNS propagation gracefully — report "pending" not "fail" if records exist on some resolvers but not all.
  - Verify: `bash scripts/validate-infrastructure.sh --dry-run --json | jq '.sections.dns'` shows all record types; with mocked dns-setup output, section reports correct pass/fail.
  - Done when: DNS section validates MX, A, SPF, DKIM, DMARC, SRV (_imaps._tcp, _submission._tcp), autoconfig CNAME, autodiscover CNAME

- [x] **T03: Implement TLS, tunnel, port, and stability validation sections** `est:45m`
  - Why: Completes all remaining validation sections by wiring existing tools. TLS checks Caddy's Let's Encrypt certs. Tunnel uses `tunnel-test.sh`. Ports use TCP dial. Stability uses `outage-sim.sh`.
  - Files: `scripts/validate-infrastructure.sh`, `scripts/lib/validate-tls.sh`, `scripts/lib/validate-tunnel.sh`, `scripts/lib/validate-ports.sh`, `scripts/lib/validate-stability.sh`
  - Do: TLS section — `openssl s_client` against cloud relay 443 for each domain, check chain validity, expiry (>7 days), and trust. Tunnel section — invoke `deploy/test/tunnel-test.sh` with configured transport type, parse output for pass/fail. Ports section — TCP connect test to cloud relay on 25, 443; test 587, 993 through tunnel IP. Stability section — invoke `deploy/test/outage-sim.sh` if running as root, skip with reason if not. Each section writes results into the JSON structure. All sections have dry-run stubs.
  - Verify: `bash scripts/validate-infrastructure.sh --dry-run --json | jq '.sections | keys | length'` returns 5; each section has `.status` and `.checks` array; `--dry-run` completes in <2 seconds.
  - Done when: all 5 sections produce structured results in both human-readable and JSON modes, dry-run works for all sections, script exits 0 on all-pass and non-zero on any failure

- [x] **T04: Add documentation, env template updates, and end-to-end dry-run verification** `est:30m`
  - Why: The script needs to be discoverable and usable by a future agent or human operator. Env templates need the new `HOME_DEVICE_IP` variable. The complete script needs a final integration test.
  - Files: `scripts/validate-infrastructure.sh` (header docs), `cloud-relay/.env.example`, `home-device/.env.example`, `deploy/README.md`
  - Do: Add comprehensive header documentation to the validation script (purpose, prerequisites, env vars, examples). Add `HOME_DEVICE_IP` to cloud-relay `.env.example` with comment. Update deploy README with validation section linking to the script. Run full dry-run and verify JSON schema consistency across all sections. Ensure `--help` flag works.
  - Verify: `bash scripts/validate-infrastructure.sh --help` exits 0 with usage; `bash scripts/validate-infrastructure.sh --dry-run --json | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['overall_status']=='pass'; assert len(d['sections'])==5"` passes; `grep HOME_DEVICE_IP cloud-relay/.env.example` finds the variable.
  - Done when: script is self-documenting, env templates updated, README links to validation, dry-run produces valid complete JSON

## Files Likely Touched

- `scripts/validate-infrastructure.sh` (new — main orchestration script)
- `scripts/lib/validate-dns.sh` (new — DNS validation section)
- `scripts/lib/validate-tls.sh` (new — TLS validation section)
- `scripts/lib/validate-tunnel.sh` (new — tunnel health section)
- `scripts/lib/validate-ports.sh` (new — port reachability section)
- `scripts/lib/validate-stability.sh` (new — outage/recovery section)
- `cloud-relay/caddy/Caddyfile` (fix — tunnel IP env var)
- `cloud-relay/.env.example` (update — HOME_DEVICE_IP)
- `deploy/README.md` (update — validation section)
