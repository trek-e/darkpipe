---
estimated_steps: 5
estimated_files: 5
---

# T03: Implement TLS, tunnel, port, and stability validation sections

**Slice:** S01 — Infrastructure Validation — DNS, TLS & Tunnel
**Milestone:** M005

## Description

Complete all remaining validation sections: TLS certificate checks on Caddy endpoints, tunnel health via existing `tunnel-test.sh`, port reachability via TCP connect, and stability via `outage-sim.sh`. Each section follows the same pattern — a lib script with dry-run support and JSON output conforming to the T01 schema.

## Steps

1. Create `scripts/lib/validate-tls.sh` — for each HTTPS domain (webmail, autoconfig, autodiscover), run `openssl s_client -connect $RELAY_IP:443 -servername $DOMAIN` to validate certificate chain, check expiry with `openssl x509 -checkend`, verify subject matches domain. In dry-run mode, return mock pass results. Report days until expiry.
2. Create `scripts/lib/validate-tunnel.sh` — invoke `deploy/test/tunnel-test.sh $TRANSPORT_TYPE` (default: wireguard), capture exit code and output. Parse for pass/fail lines. Allow 30-second warm-up for fresh tunnels. In dry-run mode, mock the tunnel-test output.
3. Create `scripts/lib/validate-ports.sh` — TCP connect test using `nc -z -w5` or bash `/dev/tcp/` to cloud relay on ports 25 and 443. For ports 587 and 993, test via tunnel IP (10.8.0.2 or configured `HOME_DEVICE_IP`). Report per-port pass/fail with timeout details. In dry-run mode, mock all ports as reachable.
4. Create `scripts/lib/validate-stability.sh` — check if running as root; if yes, invoke `deploy/test/outage-sim.sh $TRANSPORT_TYPE` and capture results. If not root, skip with status "skip" and reason "requires root". In dry-run mode, mock pass.
5. Integrate all sections into main script — source each lib file, wire into section runners, verify complete JSON output with all 5 sections populated.

## Must-Haves

- [ ] TLS section checks certificate chain, expiry, and domain match for all configured HTTPS domains
- [ ] Tunnel section invokes existing `tunnel-test.sh` and parses results
- [ ] Port section tests 25, 443 on relay and 587, 993 through tunnel
- [ ] Stability section invokes `outage-sim.sh` when root, skips gracefully when not
- [ ] All sections have dry-run mode and JSON output
- [ ] Script overall_status reflects worst section status (any fail → overall fail)

## Verification

- `bash scripts/validate-infrastructure.sh --dry-run --json | jq '.sections | keys | length'` returns 5
- `bash scripts/validate-infrastructure.sh --dry-run --json | jq '.sections.tls.checks | length'` returns ≥ 1
- `bash scripts/validate-infrastructure.sh --dry-run --json | jq '.sections.ports.checks | length'` returns 4 (25, 443, 587, 993)
- `bash scripts/validate-infrastructure.sh --dry-run --json | jq '.overall_status'` returns `"pass"`
- Dry-run completes in < 2 seconds (no network calls)

## Observability Impact

- Signals added/changed: TLS expiry days, tunnel handshake age, port response times, stability recovery duration
- How a future agent inspects this: `--json | jq '.sections.tls.checks[] | select(.status=="fail")'` pinpoints cert issues; same pattern for all sections
- Failure state exposed: TLS shows days to expiry and chain errors; tunnel shows handshake age and connectivity; ports show which specific port/host failed; stability shows recovery time or timeout

## Inputs

- `scripts/validate-infrastructure.sh` — skeleton with section runner from T01
- `scripts/lib/validate-dns.sh` — pattern reference from T02
- `deploy/test/tunnel-test.sh` — existing tunnel health test
- `deploy/test/outage-sim.sh` — existing outage simulation

## Expected Output

- `scripts/lib/validate-tls.sh` — new, TLS validation section
- `scripts/lib/validate-tunnel.sh` — new, tunnel health section
- `scripts/lib/validate-ports.sh` — new, port reachability section
- `scripts/lib/validate-stability.sh` — new, stability section
- `scripts/validate-infrastructure.sh` — updated with all sections wired in
