---
id: T01
parent: S02
milestone: M003
provides:
  - scripts/check-runtime.sh — runtime compatibility check script
  - scripts/verify-s02-docs.sh — slice acceptance verification script
key_files:
  - scripts/check-runtime.sh
  - scripts/verify-s02-docs.sh
key_decisions:
  - Reused pass/fail/skip output pattern from verify-podman-compat.sh with extended fail() supporting detected/required/fix fields
  - Added --quiet flag for CI usage (summary-only output)
  - Compose detection checks docker compose plugin first, then podman-compose, then legacy docker-compose standalone
  - Port check uses ss > netstat > lsof fallback chain for cross-platform compatibility
patterns_established:
  - Extended fail() helper with optional detected/required/fix parameters for structured failure diagnostics
  - --help and --quiet flags as standard for DarkPipe check scripts
observability_surfaces:
  - "bash scripts/check-runtime.sh — structured PASS/FAIL/SKIP output with environment summary"
  - "bash scripts/verify-s02-docs.sh — slice acceptance criteria status"
duration: 15m
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T01: Create verification script and runtime compatibility check script

**Created `scripts/check-runtime.sh` (runtime prereq validator) and `scripts/verify-s02-docs.sh` (slice acceptance checker), both passing syntax validation and producing structured output.**

## What Happened

Built two scripts following the pass/fail/skip pattern from `verify-podman-compat.sh`:

1. **`scripts/check-runtime.sh`** — Detects container runtime (Docker/Podman), validates version minimums (Docker 24+ / Podman 5.3+), checks compose tool availability (docker compose / podman-compose / docker-compose), reports SELinux enforcement state, and verifies port 25 availability. Each failed check prints detected value, required value, and suggested fix. Supports `--help` and `--quiet` flags. Handles missing commands gracefully with skip.

2. **`scripts/verify-s02-docs.sh`** — Checks all S02 slice acceptance criteria: runtime script existence/executability/syntax, Podman platform guide existence (>50 lines), FAQ free of "Not tested"/"not officially supported", quickstart contains "container runtime" and mentions Podman, platform guides (raspberry-pi, proxmox-lxc) mention Podman.

## Verification

- `bash -n scripts/check-runtime.sh` — exits 0 ✅
- `bash -n scripts/verify-s02-docs.sh` — exits 0 ✅
- `bash scripts/check-runtime.sh --help` — prints usage, exits 0 ✅
- `bash scripts/check-runtime.sh` — detected Podman 5.8.0, docker-compose 5.0.0, SELinux skipped (macOS), port 25 available. 4 pass, 0 fail, 1 skip ✅
- `bash scripts/verify-s02-docs.sh` — 4 pass (runtime script checks), 7 fail (expected: docs not yet updated) ✅

### Slice-level verification status (T01)

| Check | Status | Notes |
|-------|--------|-------|
| `check-runtime.sh --help` exits 0 | ✅ PASS | |
| `bash -n check-runtime.sh` | ✅ PASS | |
| `check-runtime.sh` exists and executable | ✅ PASS | |
| `podman.md` exists >50 lines | ❌ FAIL | Expected — T02 deliverable |
| FAQ no "Not tested" | ❌ FAIL | Expected — T03 deliverable |
| quickstart "container runtime" | ❌ FAIL | Expected — T03 deliverable |
| quickstart mentions Podman | ❌ FAIL | Expected — T03 deliverable |
| Platform guides mention Podman | ❌ FAIL | Expected — T04 deliverable |

## Diagnostics

- Run `bash scripts/check-runtime.sh` to see full runtime environment status with structured PASS/FAIL/SKIP output
- Run `bash scripts/check-runtime.sh --quiet` for summary-only output (CI-friendly)
- Run `bash scripts/verify-s02-docs.sh` to check all slice acceptance criteria

## Deviations

- Added `--quiet` flag to `check-runtime.sh` (not in plan but aligns with CI usage pattern)
- Added legacy `docker-compose` standalone detection as third compose fallback (plan only mentioned `docker compose` and `podman-compose`)

## Known Issues

None.

## Files Created/Modified

- `scripts/check-runtime.sh` — Runtime compatibility check script (new)
- `scripts/verify-s02-docs.sh` — S02 slice acceptance verification script (new)
