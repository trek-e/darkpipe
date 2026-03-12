---
estimated_steps: 5
estimated_files: 3
---

# T01: Create verification script and runtime compatibility check script

**Slice:** S02 — Runtime-Agnostic Documentation & Tooling
**Milestone:** M003

## Description

Verification-first: create `scripts/verify-s02-docs.sh` with all slice acceptance checks (most will fail initially since docs aren't updated yet), and create `scripts/check-runtime.sh` — the primary tooling deliverable that detects the user's container runtime, validates version requirements, checks compose tool availability, and reports SELinux state.

The check-runtime.sh script reuses the pass/fail/skip shell function pattern established by `scripts/verify-podman-compat.sh` in S01. It answers: "Is my system ready to run DarkPipe?" for any supported runtime.

## Steps

1. Read `scripts/verify-podman-compat.sh` to extract the pass/fail/skip function pattern and output format
2. Create `scripts/check-runtime.sh` with these check categories:
   - Runtime detection: find docker or podman, report which is available
   - Version validation: Docker 24+ or Podman 5.3+ (permissive regex for distribution-specific version strings)
   - Compose tool: `docker compose` or `podman-compose` available
   - SELinux state: detect enforcement (graceful skip if `getenforce` not found)
   - Network basics: port 25 not already bound (cloud relay prerequisite)
   - Summary: total pass/fail/skip counts with exit code 0 (all pass) or 1 (any fail)
3. Create `scripts/verify-s02-docs.sh` with slice acceptance checks:
   - `check-runtime.sh` exists and is executable
   - `deploy/platform-guides/podman.md` exists and has >50 lines
   - FAQ no longer contains "Not tested" or "not officially supported" in Podman section
   - `docs/quickstart.md` contains "container runtime" and mentions Podman
   - Platform guides (raspberry-pi, proxmox-lxc) mention Podman
   - All checks use the same pass/fail pattern
4. Make both scripts executable
5. Validate: `bash -n` both scripts, run `check-runtime.sh`, run `verify-s02-docs.sh` (expect partial failures)

## Must-Haves

- [ ] `check-runtime.sh` detects Docker and Podman with correct version thresholds
- [ ] `check-runtime.sh` handles missing commands gracefully (skip, not crash)
- [ ] `check-runtime.sh` follows the pass/fail/skip output pattern from verify-podman-compat.sh
- [ ] `check-runtime.sh` supports `--help` flag
- [ ] `verify-s02-docs.sh` checks all slice acceptance criteria
- [ ] Both scripts pass `bash -n` syntax validation

## Verification

- `bash -n scripts/check-runtime.sh` exits 0
- `bash -n scripts/verify-s02-docs.sh` exits 0
- `bash scripts/check-runtime.sh` runs and prints structured PASS/FAIL output
- `bash scripts/check-runtime.sh --help` prints usage and exits 0
- `bash scripts/verify-s02-docs.sh` runs (some checks expected to fail at this stage)

## Observability Impact

- Signals added/changed: `check-runtime.sh` provides structured PASS/FAIL/SKIP output for each prerequisite check
- How a future agent inspects this: Run `bash scripts/check-runtime.sh` to see runtime environment status
- Failure state exposed: Each failed check prints the detected value, required value, and suggested fix

## Inputs

- `scripts/verify-podman-compat.sh` — pass/fail/skip function pattern to reuse
- `scripts/verify-container-security.sh` — additional pattern reference for compose file iteration
- S02 research — version thresholds (Docker 24+, Podman 5.3+), SELinux detection approach

## Expected Output

- `scripts/check-runtime.sh` — executable runtime compatibility check script with structured output
- `scripts/verify-s02-docs.sh` — executable slice verification script checking all S02 acceptance criteria
