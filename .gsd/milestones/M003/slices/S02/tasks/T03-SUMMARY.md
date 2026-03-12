---
id: T03
parent: S02
milestone: M003
provides:
  - Runtime-agnostic language in all five core docs with Podman callouts
key_files:
  - docs/quickstart.md
  - docs/configuration.md
  - docs/contributing.md
  - docs/security.md
  - docs/faq.md
key_decisions:
  - Added "Container Runtime" note block at top of quickstart.md and configuration.md rather than inline disclaimers — keeps callouts scannable without cluttering command examples
  - FAQ Podman answer links to platform guide and lists key differences (rootful, overrides, SELinux) rather than duplicating setup instructions
patterns_established:
  - "> **Podman users:**" callout block pattern for runtime-specific notes in docs
  - Genericize nouns ("container environment", "container HEALTHCHECK") but keep commands copy-pasteable (`docker compose`)
observability_surfaces:
  - none (documentation only)
duration: ~10 min
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T03: Make core docs runtime-agnostic and update FAQ

**Updated five core docs to use runtime-agnostic language with Podman callouts, and rewrote FAQ to declare Podman "fully supported" with link to platform guide.**

## What Happened

Updated all five core documentation files following the pattern: keep `docker compose` commands copy-pasteable, genericize "Docker" to "container" where the statement applies to any runtime, and add Podman callout blocks where behavior differs.

- **quickstart.md**: Added "Container Runtime" note at top, changed "home device running Docker" to "running containers", added Podman callouts at Docker install step, cloud relay deploy, and home device deploy sections.
- **configuration.md**: Added runtime note at top, added Podman callout at compose profiles section, genericized "Docker Compose" section heading to "Compose Profiles" and "All Docker Compose services" to "All compose services".
- **contributing.md**: Updated prerequisites from "Docker 27+" to "Docker 24+ or Podman 5.3+", added note about `scripts/check-runtime.sh` for environment validation.
- **security.md**: Genericized "Docker HEALTHCHECK" to "container HEALTHCHECK" in two places, added note about Podman rootless mode providing additional security isolation.
- **faq.md**: Rewrote "Can I use Podman?" from "Probably, but not officially supported" to "Yes, fully supported" with link to Podman platform guide, key differences list (rootful, overrides, SELinux), and check-runtime.sh reference. Removed all "Not tested" and "not officially supported" language. Also genericized "Basic Docker knowledge" to "Basic container knowledge".

## Verification

All task-level checks pass:
- `grep -q "container runtime" docs/quickstart.md` — PASS
- `grep -qi "podman" docs/quickstart.md` — PASS
- `grep -qi "fully supported" docs/faq.md` — PASS
- `! grep -qi "not tested\|not officially supported" docs/faq.md` — PASS
- `grep -qi "podman 5.3" docs/contributing.md` — PASS
- `docker compose` commands remain in quickstart.md (16 occurrences, unchanged)

Slice-level verification (`bash scripts/verify-s02-docs.sh`): 10/12 pass, 2 fail.
The 2 failures are platform guide Podman notes (raspberry-pi.md, proxmox-lxc.md) — those are scope of T04/T05, not T03.

## Diagnostics

- Grep for key phrases in each doc to verify language changes
- Run `bash scripts/verify-s02-docs.sh` to check all slice acceptance criteria

## Deviations

None.

## Known Issues

None.

## Files Created/Modified

- `docs/quickstart.md` — Added Container Runtime note, Podman callouts, genericized "home device running Docker"
- `docs/configuration.md` — Added runtime note, Podman callout at profiles section, genericized headings
- `docs/contributing.md` — Updated prerequisites to include Podman 5.3+, added check-runtime.sh note
- `docs/security.md` — Genericized HEALTHCHECK references, added Podman rootless security note
- `docs/faq.md` — Rewrote Podman answer to "fully supported" with platform guide link
