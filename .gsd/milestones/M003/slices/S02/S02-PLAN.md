# S02: Runtime-Agnostic Documentation & Tooling

**Goal:** All core docs use runtime-agnostic language, a runtime compatibility check script validates supported runtimes, a Podman platform guide is published, and the FAQ declares Podman "fully supported."
**Demo:** Running `bash scripts/check-runtime.sh` on a system with Docker or Podman prints PASS/FAIL status for each prerequisite. Docs no longer assume Docker — they use generic terminology with callout blocks for runtime-specific differences. The Podman platform guide consolidates deployment instructions into a single reference. FAQ says "Yes, fully supported."

## Must-Haves

- Runtime compatibility check script (`scripts/check-runtime.sh`) that detects Docker/Podman, validates versions, checks compose availability, and reports SELinux state
- Podman platform guide (`deploy/platform-guides/podman.md`) consolidating cloud-relay and home-device PODMAN.md content
- Core docs (`docs/quickstart.md`, `docs/configuration.md`, `docs/contributing.md`, `docs/security.md`) use runtime-agnostic language with Podman callout blocks
- FAQ updated with official "fully supported" Podman answer and links to guide
- Platform guides updated with Podman subsections where relevant (raspberry-pi, proxmox-lxc) and "not applicable" notes where not (synology, unraid, truenas)
- All existing Docker command examples remain copy-pasteable (no `$RUNTIME` variables)

## Proof Level

- This slice proves: contract
- Real runtime required: no (script tested with mocked/real `--version` output; doc changes are textual)
- Human/UAT required: no

## Verification

- `bash scripts/check-runtime.sh --help` exits 0 and prints usage
- `bash -n scripts/check-runtime.sh` passes syntax check
- `bash scripts/verify-s02-docs.sh` — custom verification script that:
  - Confirms `scripts/check-runtime.sh` exists and is executable
  - Confirms `deploy/platform-guides/podman.md` exists and has >50 lines
  - Confirms FAQ no longer contains "Not tested" or "not officially supported" in the Podman section
  - Confirms `docs/quickstart.md` contains "container runtime" (runtime-agnostic language present)
  - Confirms `docs/quickstart.md` contains "podman" or "Podman" (Podman callout present)
  - Confirms each platform guide that should have a Podman note does (raspberry-pi, proxmox-lxc)
  - All checks pass with 0 failures

## Observability / Diagnostics

- Runtime signals: `check-runtime.sh` outputs structured PASS/FAIL/SKIP lines with check names, matching `verify-podman-compat.sh` pattern
- Inspection surfaces: `check-runtime.sh` prints detected runtime, version, compose tool, and SELinux state
- Failure visibility: Each failed check prints the reason and suggested fix
- Redaction constraints: None (no secrets in docs or check script)

## Integration Closure

- Upstream surfaces consumed: S01's `cloud-relay/PODMAN.md`, `home-device/PODMAN.md`, `docker-compose.podman.yml` override files, `scripts/verify-podman-compat.sh` patterns
- New wiring introduced in this slice: `scripts/check-runtime.sh` (new entrypoint for runtime validation), `deploy/platform-guides/podman.md` (new guide), `scripts/verify-s02-docs.sh` (verification)
- What remains before the milestone is truly usable end-to-end: S03 (Apple Containers guide), S04 (CI integration of check-runtime.sh and Podman build job)

## Tasks

- [x] **T01: Create verification script and runtime compatibility check script** `est:45m`
  - Why: Verification-first — create the S02 verification script (initially failing) and the runtime compatibility check script that is the primary tooling deliverable of this slice
  - Files: `scripts/verify-s02-docs.sh`, `scripts/check-runtime.sh`
  - Do: Build `verify-s02-docs.sh` with all slice acceptance checks (will fail until docs are updated). Build `check-runtime.sh` reusing pass/fail/skip pattern from `verify-podman-compat.sh` — detect runtime (docker/podman), check version minimums (Docker 24+, Podman 5.3+), check compose tool, detect SELinux state, validate network prerequisites. Handle missing commands gracefully with skip.
  - Verify: `bash -n scripts/check-runtime.sh && bash -n scripts/verify-s02-docs.sh` (syntax valid); `bash scripts/check-runtime.sh` runs and prints structured output; `bash scripts/verify-s02-docs.sh` runs (expected: some checks fail since docs not yet updated)
  - Done when: Both scripts exist, are executable, pass syntax check, and `check-runtime.sh` correctly detects the local runtime environment

- [x] **T02: Publish Podman platform guide** `est:30m`
  - Why: The consolidated Podman deployment guide is a primary deliverable — it serves as the single source of truth for Podman users, referenced by FAQ and other docs
  - Files: `deploy/platform-guides/podman.md`
  - Do: Create comprehensive guide following existing platform guide structure (Prerequisites, Quick Start, Detailed Steps, Troubleshooting, Resources). Consolidate content from `cloud-relay/PODMAN.md` and `home-device/PODMAN.md`. Cover: rootful requirement for cloud relay, rootless option for home device, override file usage, SELinux labels, firewalld config, version requirements, host-gateway difference. Include the `check-runtime.sh` in the prerequisites section.
  - Verify: File exists, >50 lines, covers rootful/rootless, references override files, includes troubleshooting section
  - Done when: `deploy/platform-guides/podman.md` is a complete, self-contained Podman deployment guide

- [x] **T03: Make core docs runtime-agnostic and update FAQ** `est:45m`
  - Why: The core docs (quickstart, configuration, contributing, security, faq) are the most-read files and need runtime-agnostic language to make Podman a first-class option
  - Files: `docs/quickstart.md`, `docs/configuration.md`, `docs/contributing.md`, `docs/security.md`, `docs/faq.md`
  - Do: Add a "Container Runtime" note at the top of quickstart/configuration explaining that examples use `docker compose` but `podman-compose` works identically with override files. Replace "Docker" with "container runtime" where the statement is runtime-generic (e.g., "Docker containers" → "containers"). Keep `docker compose` in command examples — do NOT replace with variables. Add Podman callout blocks ("> **Podman users:**") where behavior differs. Update FAQ Podman section: replace "Not tested" with "Yes, fully supported" answer, link to platform guide, list key differences. Update contributing.md prerequisites to list Podman as alternative.
  - Verify: `bash scripts/verify-s02-docs.sh` — FAQ check passes, quickstart contains "container runtime" and "Podman" callout
  - Done when: All 5 core docs updated, FAQ declares Podman fully supported with link to guide

- [x] **T04: Update platform guides with Podman notes** `est:30m`
  - Why: Platform guides are deployment-specific — users on Raspberry Pi or Proxmox need to know Podman is an option; users on Synology/Unraid need to know it's not practical
  - Files: `deploy/platform-guides/raspberry-pi.md`, `deploy/platform-guides/proxmox-lxc.md`, `deploy/platform-guides/unraid.md`, `deploy/platform-guides/synology-nas.md`, `deploy/platform-guides/truenas-scale.md`, `deploy/platform-guides/mac-silicon.md`
  - Do: Add "Using Podman" subsection to raspberry-pi and proxmox-lxc guides (these platforms commonly run Podman). Add brief "Podman not applicable" notes to synology, unraid, truenas (platform uses native Docker). Add Apple Containers forward-reference to mac-silicon guide (link to future S03 guide). Keep changes minimal — don't rewrite guides, just add the relevant Podman context.
  - Verify: `bash scripts/verify-s02-docs.sh` passes all checks including platform guide Podman notes; `grep -l -i podman deploy/platform-guides/raspberry-pi.md deploy/platform-guides/proxmox-lxc.md` returns both files
  - Done when: All 6 platform guides updated, verification script passes all checks with 0 failures

## Files Likely Touched

- `scripts/check-runtime.sh` (new)
- `scripts/verify-s02-docs.sh` (new)
- `deploy/platform-guides/podman.md` (new)
- `docs/quickstart.md`
- `docs/configuration.md`
- `docs/contributing.md`
- `docs/security.md`
- `docs/faq.md`
- `deploy/platform-guides/raspberry-pi.md`
- `deploy/platform-guides/proxmox-lxc.md`
- `deploy/platform-guides/unraid.md`
- `deploy/platform-guides/synology-nas.md`
- `deploy/platform-guides/truenas-scale.md`
- `deploy/platform-guides/mac-silicon.md`
