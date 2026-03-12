---
estimated_steps: 5
estimated_files: 3
---

# T01: Fix tunnel IP mismatch and create validation script skeleton with dry-run mode

**Slice:** S01 — Infrastructure Validation — DNS, TLS & Tunnel
**Milestone:** M005

## Description

The Caddyfile hardcodes `10.0.0.2` as the home device IP but WireGuard templates default to `10.8.0.2`. This mismatch will break Caddy's reverse proxy to the home device. Fix this by making the Caddyfile use an env var `{$HOME_DEVICE_IP}` with the correct default.

Then create the validation script skeleton (`scripts/validate-infrastructure.sh`) with the output format, argument parsing, section runner framework, and dry-run mode. This skeleton is the foundation that T02 and T03 build on.

## Steps

1. Fix `cloud-relay/caddy/Caddyfile` — replace all `10.0.0.2` references with `{$HOME_DEVICE_IP}` (Caddy env var syntax). Note the default must be set in `.env` or compose, not in Caddyfile syntax.
2. Update `cloud-relay/docker-compose.yml` — add `HOME_DEVICE_IP` env var to Caddy service with default `10.8.0.2` matching WireGuard subnet.
3. Create `scripts/validate-infrastructure.sh` — argument parsing for `--json`, `--verbose`, `--dry-run`, `--help`. Source config from env vars. Define JSON output schema with `overall_status` and `sections` map (dns, tls, tunnel, ports, stability). Each section has `status`, `checks` array, and `timestamp`.
4. Implement section runner framework — function that takes section name, runs section script from `scripts/lib/`, captures output, updates JSON. Dry-run mode returns mock pass results for each section.
5. Create `scripts/lib/` directory with empty section stubs that return mock results.

## Must-Haves

- [ ] All `10.0.0.2` in Caddyfile replaced with `{$HOME_DEVICE_IP}`
- [ ] Caddy compose service passes `HOME_DEVICE_IP` with default `10.8.0.2`
- [ ] Validation script has `--json`, `--verbose`, `--dry-run`, `--help` flags
- [ ] Dry-run produces valid JSON with all 5 section placeholders
- [ ] Script is executable (`chmod +x`)

## Verification

- `grep '10.0.0.2' cloud-relay/caddy/Caddyfile` returns empty (no hardcoded IPs)
- `grep 'HOME_DEVICE_IP' cloud-relay/caddy/Caddyfile` finds env var references
- `bash scripts/validate-infrastructure.sh --dry-run --json | jq '.overall_status'` returns `"pass"`
- `bash scripts/validate-infrastructure.sh --dry-run --json | jq '.sections | keys | sort'` returns `["dns","ports","stability","tls","tunnel"]`

## Observability Impact

- Signals added/changed: JSON output schema established — all future sections conform to `{status, checks[], timestamp}` structure
- How a future agent inspects this: `--dry-run --json` validates script structure without live infra
- Failure state exposed: each section reports status with error detail; overall_status is "fail" if any section fails

## Inputs

- `cloud-relay/caddy/Caddyfile` — current file with hardcoded `10.0.0.2`
- `cloud-relay/docker-compose.yml` — Caddy service environment
- `transport/wireguard/config/templates.go` — confirms correct subnet is `10.8.0.0/24`

## Expected Output

- `cloud-relay/caddy/Caddyfile` — updated with `{$HOME_DEVICE_IP}` env vars
- `cloud-relay/docker-compose.yml` — updated with `HOME_DEVICE_IP` env var
- `scripts/validate-infrastructure.sh` — new, executable, with working dry-run
- `scripts/lib/` — directory with stub section scripts
