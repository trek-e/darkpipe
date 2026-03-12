---
id: T01
parent: S01
milestone: M005
provides:
  - Caddyfile uses HOME_DEVICE_IP env var instead of hardcoded 10.0.0.2
  - Validation script skeleton with dry-run, JSON output, and section runner framework
key_files:
  - cloud-relay/caddy/Caddyfile
  - cloud-relay/docker-compose.yml
  - scripts/validate-infrastructure.sh
  - scripts/lib/validate-dns.sh
  - scripts/lib/validate-tls.sh
  - scripts/lib/validate-tunnel.sh
  - scripts/lib/validate-ports.sh
  - scripts/lib/validate-stability.sh
key_decisions:
  - Used file-based section result storage instead of bash associative arrays for macOS bash 3 compatibility
  - Section scripts in scripts/lib/ follow validate-{section}.sh naming convention
patterns_established:
  - JSON output schema: {overall_status, timestamp, config, sections: {name: {status, checks[], timestamp}}}
  - Section runner pattern: each section is an independent script in scripts/lib/ that outputs JSON checks
  - Dry-run mode returns mock pass results for all sections
observability_surfaces:
  - scripts/validate-infrastructure.sh --dry-run --json validates script structure without live infra
  - JSON output with per-section status, checks array, and timestamps
duration: 15min
verification_result: passed
completed_at: 2026-03-12
blocker_discovered: false
---

# T01: Fixed tunnel IP mismatch and created validation script skeleton with dry-run mode

**Replaced all hardcoded 10.0.0.2 IPs in Caddyfile with {$HOME_DEVICE_IP} env var and built the validation script framework with working dry-run JSON output.**

## What Happened

The Caddyfile had 11 instances of hardcoded `10.0.0.2` which didn't match the WireGuard subnet default of `10.8.0.2`. Replaced all with Caddy's `{$HOME_DEVICE_IP}` env var syntax and added the env var to docker-compose.yml with default `10.8.0.2`.

Created `scripts/validate-infrastructure.sh` with full argument parsing (--json, --verbose, --dry-run, --help), a section runner framework that executes scripts from `scripts/lib/`, and JSON output conforming to the defined schema. Used file-based temp storage for section results to maintain bash 3 compatibility on macOS (which lacks associative arrays).

Created 5 stub section scripts in `scripts/lib/` that return skip/placeholder results for non-dry-run mode.

## Verification

- `grep '10.0.0.2' cloud-relay/caddy/Caddyfile` — returns empty (no hardcoded IPs) ✓
- `grep 'HOME_DEVICE_IP' cloud-relay/caddy/Caddyfile` — finds env var references ✓
- `grep 'HOME_DEVICE_IP' cloud-relay/docker-compose.yml` — shows `${HOME_DEVICE_IP:-10.8.0.2}` ✓
- `bash scripts/validate-infrastructure.sh --dry-run --json | jq '.overall_status'` — returns `"pass"` ✓
- `bash scripts/validate-infrastructure.sh --dry-run --json | jq '.sections | keys | sort'` — returns all 5 sections ✓
- `bash scripts/validate-infrastructure.sh --dry-run` — exits 0 ✓
- Script is executable (chmod +x) ✓

### Slice-level verification (partial — T01 of 3):
- [x] `--dry-run` exits 0
- [x] `--dry-run --json` returns all 5 sections with correct keys
- [x] No hardcoded 10.0.0.2 in Caddyfile
- [ ] Full `--json` returns pass for all sections (needs T02/T03 real implementations)
- [ ] Each section has pass/skip status with reason (needs T02/T03)

## Diagnostics

- Run `scripts/validate-infrastructure.sh --dry-run --json | jq .` to inspect full output schema
- Run `scripts/validate-infrastructure.sh --help` for usage
- Section stubs in `scripts/lib/` return skip status when run outside dry-run mode

## Deviations

Rewrote script to avoid bash associative arrays (`declare -A`) which require bash 4+. macOS ships bash 3.2. Used file-based temp directory storage instead — functionally equivalent, fully portable.

## Known Issues

None.

## Files Created/Modified

- `cloud-relay/caddy/Caddyfile` — replaced 11 hardcoded `10.0.0.2` with `{$HOME_DEVICE_IP}`
- `cloud-relay/docker-compose.yml` — added `HOME_DEVICE_IP: ${HOME_DEVICE_IP:-10.8.0.2}` to Caddy service
- `scripts/validate-infrastructure.sh` — new validation orchestrator with dry-run, JSON, verbose, help
- `scripts/lib/validate-dns.sh` — stub section script
- `scripts/lib/validate-tls.sh` — stub section script
- `scripts/lib/validate-tunnel.sh` — stub section script
- `scripts/lib/validate-ports.sh` — stub section script
- `scripts/lib/validate-stability.sh` — stub section script
