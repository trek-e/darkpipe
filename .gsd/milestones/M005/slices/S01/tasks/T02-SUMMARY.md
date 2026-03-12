---
id: T02
parent: S01
milestone: M005
provides:
  - DNS validation section with 9 record-type checks (MX, A, SPF, DKIM, DMARC, SRV×2, CNAME×2)
  - Sourced-function integration pattern for section scripts
key_files:
  - scripts/lib/validate-dns.sh
  - scripts/validate-infrastructure.sh
key_decisions:
  - Used dig-based checks for all record types instead of depending on dns-setup binary; makes validation work without Go toolchain
  - Introduced sourced-function pattern: main script sources lib scripts that define run_<section>_validation() functions, falling back to generic dry-run/subprocess for stubs
patterns_established:
  - Section scripts define run_<section>_validation() function; main script sources and calls it
  - Each check queries both resolvers (8.8.8.8, 1.1.1.1) and detects propagation mismatches
  - DKIM key output truncated to 60 chars for redaction
observability_surfaces:
  - "jq '.sections.dns.checks[] | select(.status==\"fail\")' shows which records failed with error detail and suggested fix"
  - "Each check includes resolver-specific propagation mismatch detection"
duration: 15min
verification_result: passed
completed_at: 2026-03-12
blocker_discovered: false
---

# T02: Implement DNS validation section using existing dns-setup and validator packages

**Built DNS validation section with 9 record-type checks using external resolvers, with dry-run mock support and propagation mismatch detection.**

## What Happened

Created `scripts/lib/validate-dns.sh` with `run_dns_validation()` function that validates all required DNS record types: MX, A, SPF, DKIM, DMARC, SRV (_imaps._tcp, _submission._tcp), autoconfig CNAME, and autodiscover CNAME. Each check queries both 8.8.8.8 and 1.1.1.1 and detects propagation mismatches when resolvers disagree.

Updated the main orchestration script to source section lib scripts that define `run_<section>_validation()` functions. The source loop uses `grep` to only source scripts containing function definitions, avoiding the stubs' top-level echo statements. When a sourced function exists, `run_section()` delegates to it for both dry-run and live modes; otherwise it falls back to the existing generic handler.

Dry-run mode returns realistic mock results for all 9 record types without making network calls.

## Verification

- `bash scripts/validate-infrastructure.sh --dry-run --json | jq '.sections.dns.checks | length'` → **9** ✓
- `bash scripts/validate-infrastructure.sh --dry-run --json | jq -r '.sections.dns.checks[].name'` → all 9 names present (mx, a, spf, dkim, dmarc, srv_imaps, srv_submission, autoconfig_cname, autodiscover_cname) ✓
- `bash scripts/validate-infrastructure.sh --dry-run` exits 0 ✓
- Dry-run completes without network calls ✓
- Slice-level: `--json | jq '.overall_status'` → "pass" ✓
- Slice-level: `--json | jq '.sections | keys'` → all 5 sections present ✓
- Slice-level: `grep '10.0.0.2' cloud-relay/caddy/Caddyfile` → none found ✓

## Diagnostics

- `scripts/validate-infrastructure.sh --dry-run --json | jq '.sections.dns.checks[] | select(.status=="fail")'` — shows failed DNS checks with error detail and suggested fix
- Each check includes which resolvers returned what, detecting propagation delays
- DKIM key display truncated to prevent secrets leaking

## Deviations

Used dig-based checks for all record types rather than invoking `dns-setup --validate-only --json` binary. The dig approach is self-contained (no Go build dependency), queries the same external resolvers, and checks the same records. The dns-setup binary can be integrated later if deeper validation is needed.

## Known Issues

None.

## Files Created/Modified

- `scripts/lib/validate-dns.sh` — new, DNS validation section with run_dns_validation() function
- `scripts/validate-infrastructure.sh` — updated to source section libs and delegate to sourced functions
