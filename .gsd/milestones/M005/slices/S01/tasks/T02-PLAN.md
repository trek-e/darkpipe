---
estimated_steps: 4
estimated_files: 2
---

# T02: Implement DNS validation section using existing dns-setup and validator packages

**Slice:** S01 — Infrastructure Validation — DNS, TLS & Tunnel
**Milestone:** M005

## Description

Wire the existing DNS validation tools into the orchestration script. The `dns-setup --validate-only --json` CLI already validates MX, A, SPF, DKIM, and DMARC against external resolvers. SRV records and autodiscover CNAMEs need separate `dig` queries since they're not covered by the CLI's validate-only mode. The DNS section must aggregate all results and handle propagation delays gracefully.

## Steps

1. Create `scripts/lib/validate-dns.sh` with `run_dns_validation()` function. In dry-run mode, return mock results for all record types.
2. For live mode: build and invoke `dns-setup --validate-only --json` (binary at `dns/cmd/dns-setup/`), capture JSON output, extract per-record results. Handle missing binary gracefully (suggest `go build`).
3. Add SRV record checks via `dig @8.8.8.8 _imaps._tcp.$DOMAIN SRV` and `dig @8.8.8.8 _submission._tcp.$DOMAIN SRV`. Add autodiscover CNAME checks via `dig @8.8.8.8 autoconfig.$DOMAIN CNAME` and `dig @8.8.8.8 autodiscover.$DOMAIN CNAME`. Parse results and add to checks array.
4. Integrate into main script — source `scripts/lib/validate-dns.sh`, call it from the dns section runner, merge results into JSON output.

## Must-Haves

- [ ] DNS section validates: MX, A, SPF, DKIM, DMARC, SRV (_imaps._tcp, _submission._tcp), autoconfig CNAME, autodiscover CNAME
- [ ] Uses external resolvers (8.8.8.8, 1.1.1.1) not local resolver
- [ ] Handles missing `dns-setup` binary with clear error message
- [ ] Dry-run mode returns mock results without network calls
- [ ] JSON output conforms to section schema from T01

## Verification

- `bash scripts/validate-infrastructure.sh --dry-run --json | jq '.sections.dns.checks | length'` returns ≥ 9 (one per record type)
- `bash scripts/validate-infrastructure.sh --dry-run --json | jq '.sections.dns.checks[].name'` includes "mx", "a", "spf", "dkim", "dmarc", "srv_imaps", "srv_submission", "autoconfig_cname", "autodiscover_cname"
- Dry-run completes without network calls

## Observability Impact

- Signals added/changed: per-record pass/fail/pending status with resolver responses
- How a future agent inspects this: `--json | jq '.sections.dns.checks[] | select(.status=="fail")'` shows which records failed
- Failure state exposed: each check includes error detail and which resolvers disagree (propagation detection)

## Inputs

- `scripts/validate-infrastructure.sh` — skeleton from T01 with section runner framework
- `dns/cmd/dns-setup/main.go` — existing CLI with `--validate-only --json`
- `dns/validator/srv.go` — reference for SRV record names and expected values

## Expected Output

- `scripts/lib/validate-dns.sh` — new, DNS validation section implementation
- `scripts/validate-infrastructure.sh` — updated to source and invoke DNS section
