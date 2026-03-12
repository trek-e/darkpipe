---
id: T03
parent: S01
milestone: M005
provides:
  - TLS validation section with certificate chain, expiry, and domain match checks for 3 HTTPS endpoints
  - Tunnel validation section wrapping deploy/test/tunnel-test.sh with parsed pass/fail output
  - Port reachability section testing TCP connectivity on ports 25, 443, 587, 993
  - Stability validation section wrapping deploy/test/outage-sim.sh with root-privilege gating
  - All 5 infrastructure validation sections now populated and producing JSON output
key_files:
  - scripts/lib/validate-tls.sh
  - scripts/lib/validate-tunnel.sh
  - scripts/lib/validate-ports.sh
  - scripts/lib/validate-stability.sh
key_decisions:
  - TLS checks connect to RELAY_IP (or resolved domain A record) with SNI for each subdomain rather than connecting to each subdomain directly — avoids DNS dependency in TLS validation
  - Port checks split relay-facing (25, 443) vs tunnel-facing (587, 993) targets to match actual traffic flow
  - Stability section returns skip (not fail) when not root — non-root runs should not show false failures
patterns_established:
  - All section scripts follow identical pattern: _<section>_check() JSON helper, _<section>_dry_run_checks(), run_<section>_validation() entry point
  - TCP connect helper _tcp_connect() tries nc -z first, falls back to bash /dev/tcp
  - Tunnel/stability sections parse [PASS]/[FAIL] lines from wrapped test scripts
observability_surfaces:
  - "TLS: days until expiry per domain, chain verification status, domain match status"
  - "Tunnel: handshake age, pass/fail count from tunnel-test.sh, connectivity detail"
  - "Ports: per-port reachable/unreachable with connection timing"
  - "Stability: recovery time in seconds, root privilege status"
duration: 15min
verification_result: passed
completed_at: 2026-03-12
blocker_discovered: false
---

# T03: Implement TLS, tunnel, port, and stability validation sections

**Built all four remaining validation sections with dry-run support, JSON output, and live infrastructure checking — completing the 5-section validation framework.**

## What Happened

Replaced the four stub scripts in `scripts/lib/` with full implementations following the sourced-function pattern established in T02.

- **TLS** (`validate-tls.sh`): For each HTTPS domain (mail, autoconfig, autodiscover), connects via `openssl s_client` with SNI, validates certificate chain, checks expiry (warns at 30 days, fails at 7), and verifies domain match via SAN/CN. Reports days until expiry.
- **Tunnel** (`validate-tunnel.sh`): Wraps `deploy/test/tunnel-test.sh`, parses `[PASS]`/`[FAIL]` output lines, extracts handshake age and latency details. Allows 60s timeout for fresh tunnels.
- **Ports** (`validate-ports.sh`): TCP connect tests using `nc -z` (with `/dev/tcp` fallback) against relay IP for ports 25/443 and tunnel IP for 587/993. Auto-resolves relay IP from DNS if not set.
- **Stability** (`validate-stability.sh`): Gates on root privilege (skip if non-root), wraps `deploy/test/outage-sim.sh`, extracts recovery time and validates against 60s threshold.

No changes needed to the main `validate-infrastructure.sh` — its sourcing loop and section runner already handle all four new sections automatically.

## Verification

All task-level checks passed:

- `jq '.sections | keys | length'` → **5** ✓
- `jq '.sections.tls.checks | length'` → **9** (3 checks × 3 domains) ✓
- `jq '.sections.ports.checks | length'` → **4** (ports 25, 443, 587, 993) ✓
- `jq '.overall_status'` → **"pass"** ✓
- Dry-run timing: **0.084s** (well under 2s limit) ✓

Slice-level checks (all passing as of this task):

- All 5 section keys present: dns, tls, tunnel, ports, stability ✓
- `--dry-run` exits 0 ✓
- No hardcoded `10.0.0.2` in Caddyfile ✓

## Diagnostics

- `scripts/validate-infrastructure.sh --dry-run --json | jq '.sections.tls.checks[] | select(.status=="fail")'` — pinpoints cert issues with chain errors and expiry days
- `scripts/validate-infrastructure.sh --dry-run --json | jq '.sections.ports.checks[] | select(.status=="fail")'` — shows which port/host failed with connection detail
- `scripts/validate-infrastructure.sh --dry-run --json | jq '.sections.tunnel.checks[] | select(.status=="fail")'` — shows tunnel health failures with parsed test output
- `scripts/validate-infrastructure.sh --dry-run --json | jq '.sections.stability'` — shows skip reason (non-root) or recovery time

## Deviations

None. All steps executed as planned. Main script required no modifications — the existing sourcing loop handled integration automatically.

## Known Issues

None.

## Files Created/Modified

- `scripts/lib/validate-tls.sh` — TLS certificate validation (chain, expiry, domain match for 3 HTTPS endpoints)
- `scripts/lib/validate-tunnel.sh` — Tunnel health validation (wraps tunnel-test.sh, parses results)
- `scripts/lib/validate-ports.sh` — Port reachability validation (TCP connect on 25, 443, 587, 993)
- `scripts/lib/validate-stability.sh` — Stability validation (wraps outage-sim.sh, requires root)
