---
id: T04
parent: S02
milestone: M003
provides:
  - Podman context in all 6 platform guides (subsections or not-applicable notes)
  - Apple Containers forward-reference in Mac Silicon guide
key_files:
  - deploy/platform-guides/raspberry-pi.md
  - deploy/platform-guides/proxmox-lxc.md
  - deploy/platform-guides/synology-nas.md
  - deploy/platform-guides/unraid.md
  - deploy/platform-guides/truenas-scale.md
  - deploy/platform-guides/mac-silicon.md
key_decisions:
  - Raspberry Pi and Proxmox LXC get full "Using Podman" sections (Podman is viable); Synology, Unraid, TrueNAS get "Alternative Runtimes" notes (Podman not applicable on these platforms)
  - Apple Containers note placed as a callout block before the "Purpose" section in Mac Silicon guide, pointing to apple-containers.md placeholder
patterns_established:
  - "Alternative Runtimes" section pattern for platform guides where a runtime is not applicable — brief explanation + link to the relevant guide
observability_surfaces:
  - none
duration: ~10min
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T04: Update platform guides with Podman notes

**Added Podman context to all 6 existing platform guides and Apple Containers forward-reference to Mac Silicon guide.**

## What Happened

Added additive sections to each of the 6 platform guides:

- **raspberry-pi.md** — Full "Using Podman" section covering: installation on Raspberry Pi OS, rootful vs rootless guidance (cloud relay needs rootful, home device can be rootless), override file usage, key differences table (daemon/rootless/compose/SELinux), runtime validation command, and link to Podman platform guide.
- **proxmox-lxc.md** — Full "Using Podman" section covering: installation inside LXC, why Podman may be simpler in unprivileged LXC (no daemon needed), override file usage, runtime validation, and link to Podman platform guide.
- **synology-nas.md** — "Alternative Runtimes" note: Podman not applicable, platform uses Container Manager (Docker).
- **unraid.md** — "Alternative Runtimes" note: Podman not applicable, platform uses native Docker engine.
- **truenas-scale.md** — "Alternative Runtimes" note: Podman not applicable, platform uses native Docker engine via Custom Apps.
- **mac-silicon.md** — Apple Containers callout block referencing `apple-containers.md` (coming in S03).

All 6 guides also gained a Podman Platform Guide link in their "See Also" section. No existing content was removed or rewritten.

## Verification

- `grep -qi "podman" deploy/platform-guides/raspberry-pi.md` — PASS
- `grep -qi "podman" deploy/platform-guides/proxmox-lxc.md` — PASS
- `grep -qi "apple containers\|apple-containers" deploy/platform-guides/mac-silicon.md` — PASS
- `grep -qi "podman"` on synology-nas.md, unraid.md, truenas-scale.md — all PASS
- `bash scripts/verify-s02-docs.sh` — **12/12 checks pass, 0 failures** (final green run)

## Diagnostics

- Grep for Podman mentions: `grep -rn "odman" deploy/platform-guides/`
- Run `bash scripts/verify-s02-docs.sh` to re-check all slice acceptance criteria

## Deviations

None.

## Known Issues

None.

## Files Created/Modified

- `deploy/platform-guides/raspberry-pi.md` — Added "Using Podman" section with install, rootful/rootless, overrides, key differences table
- `deploy/platform-guides/proxmox-lxc.md` — Added "Using Podman" section with install, LXC benefits, overrides
- `deploy/platform-guides/synology-nas.md` — Added "Alternative Runtimes" note (Podman not applicable)
- `deploy/platform-guides/unraid.md` — Added "Alternative Runtimes" note (Podman not applicable)
- `deploy/platform-guides/truenas-scale.md` — Added "Alternative Runtimes" note (Podman not applicable)
- `deploy/platform-guides/mac-silicon.md` — Added Apple Containers forward-reference callout
