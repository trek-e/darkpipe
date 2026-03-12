---
estimated_steps: 4
estimated_files: 4
---

# T04: Add documentation, env template updates, and end-to-end dry-run verification

**Slice:** S01 — Infrastructure Validation — DNS, TLS & Tunnel
**Milestone:** M005

## Description

Make the validation script discoverable and self-documenting. Update env templates with the new `HOME_DEVICE_IP` variable. Update deploy documentation to reference the validation workflow. Run final end-to-end dry-run to verify JSON schema consistency and complete output.

## Steps

1. Add comprehensive header documentation to `scripts/validate-infrastructure.sh` — purpose, prerequisites (env vars, tools), usage examples, environment variable reference table, exit codes. Ensure `--help` flag outputs this cleanly.
2. Update `cloud-relay/.env.example` — add `HOME_DEVICE_IP=10.8.0.2` with comment explaining it must match WireGuard tunnel addressing. Review and ensure all env vars used by the validation script are documented.
3. Update `deploy/README.md` — add "Infrastructure Validation" section explaining when and how to run the validation script, what each section checks, and how to interpret results. Link to the script.
4. Run full dry-run verification — execute `scripts/validate-infrastructure.sh --dry-run --json` and validate: all 5 sections present, each section has `status` and `checks` array, overall_status is "pass", all check names are present, JSON is valid. Fix any inconsistencies found.

## Must-Haves

- [ ] `--help` flag works and shows usage, env vars, and examples
- [ ] `HOME_DEVICE_IP` in `cloud-relay/.env.example` with documentation comment
- [ ] `deploy/README.md` has infrastructure validation section
- [ ] Dry-run produces valid JSON with all sections and checks populated

## Verification

- `bash scripts/validate-infrastructure.sh --help` exits 0 with usage text including env var table
- `grep HOME_DEVICE_IP cloud-relay/.env.example` finds the variable with comment
- `grep -l validate-infrastructure deploy/README.md` finds reference
- `bash scripts/validate-infrastructure.sh --dry-run --json | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['overall_status']=='pass'; assert len(d['sections'])==5; [assert len(s['checks'])>0 for s in d['sections'].values()]"` passes

## Observability Impact

- Signals added/changed: None (documentation task)
- How a future agent inspects this: `--help` provides self-contained usage reference; env.example documents all required variables
- Failure state exposed: None

## Inputs

- `scripts/validate-infrastructure.sh` — completed script from T01-T03
- `scripts/lib/validate-*.sh` — all section scripts from T02-T03
- `cloud-relay/.env.example` — existing env template
- `deploy/README.md` — existing deployment docs

## Expected Output

- `scripts/validate-infrastructure.sh` — updated with header docs and --help
- `cloud-relay/.env.example` — updated with HOME_DEVICE_IP
- `deploy/README.md` — updated with validation section
