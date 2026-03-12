# DarkPipe Deployment

This directory contains deployment resources for DarkPipe: platform guides,
WireGuard configuration, PKI setup, and infrastructure validation tooling.

## Directory Structure

```
deploy/
├── platform-guides/   # Per-platform setup guides (Raspberry Pi, Synology, etc.)
├── pki/               # Certificate authority and mTLS tooling
├── setup/             # Deployment setup CLI
├── templates/         # Configuration templates
├── test/              # Integration and tunnel test scripts
└── wireguard/         # WireGuard configuration helpers
```

## Infrastructure Validation

After deploying the cloud relay and home device, run the infrastructure
validation script to verify the full stack is working:

```bash
# Quick smoke test (no live infrastructure needed)
./scripts/validate-infrastructure.sh --dry-run

# Full validation against a live deployment
RELAY_DOMAIN=darkpipe.email ./scripts/validate-infrastructure.sh --json | jq .

# Human-readable output with verbose diagnostics
RELAY_DOMAIN=darkpipe.email ./scripts/validate-infrastructure.sh --verbose
```

### What Each Section Checks

| Section     | Description |
|-------------|-------------|
| **dns**     | MX, A, SPF, DKIM, DMARC, SRV, and CNAME records resolve correctly from external resolvers (Google 8.8.8.8, Cloudflare 1.1.1.1). Detects propagation mismatches. |
| **tls**     | Certificate chain validity, expiry (warns at <30 days), and domain match for `relay.`, `mail.`, and `autoconfig.` subdomains. Connects to RELAY_IP with SNI. |
| **tunnel**  | WireGuard tunnel connectivity between cloud relay and home device via `deploy/test/tunnel-test.sh`. |
| **ports**   | TCP reachability for SMTP (25) and HTTPS (443) on the relay, plus submission (587) and IMAP (993) through the tunnel. |
| **stability** | Service recovery timing after restart. Requires root; skips with reason when run unprivileged. |

### Environment Variables

The validation script reads these environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `RELAY_DOMAIN` | `example.com` | Primary mail domain to validate |
| `HOME_DEVICE_IP` | `10.8.0.2` | Home device IP on WireGuard tunnel |
| `RELAY_IP` | *(auto-detect)* | Cloud relay public IPv4 address |

### Interpreting Results

**JSON output** (`--json`) returns a structured object:

```json
{
  "overall_status": "pass",
  "timestamp": "2026-03-12T12:00:00Z",
  "config": { "relay_domain": "...", "home_device_ip": "...", "dry_run": false },
  "sections": {
    "dns":       { "status": "pass", "checks": [...], "timestamp": "..." },
    "tls":       { "status": "pass", "checks": [...], "timestamp": "..." },
    "tunnel":    { "status": "pass", "checks": [...], "timestamp": "..." },
    "ports":     { "status": "pass", "checks": [...], "timestamp": "..." },
    "stability": { "status": "skip", "checks": [...], "timestamp": "..." }
  }
}
```

Each check in the `checks` array has: `name`, `status` (pass/fail/skip),
`detail`, and `suggested_fix`.

**Exit codes:** 0 = all pass, 1 = one or more failures, 2 = script error.

**Filtering failures:**

```bash
# Show only failed DNS checks
./scripts/validate-infrastructure.sh --json | jq '.sections.dns.checks[] | select(.status=="fail")'

# Show all failures across all sections
./scripts/validate-infrastructure.sh --json | jq '[.sections[].checks[] | select(.status=="fail")]'
```

### When to Run

- **After initial deployment** — verify DNS propagation, TLS provisioning, and tunnel connectivity.
- **After configuration changes** — confirm nothing broke (domain, IP, or certificate changes).
- **In CI/CD** — use `--dry-run` to validate script logic without live infrastructure.
- **Troubleshooting** — use `--json` output to pinpoint which specific check is failing.
