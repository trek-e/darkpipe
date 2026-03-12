---
estimated_steps: 4
estimated_files: 6
---

# T04: Update platform guides with Podman notes

**Slice:** S02 — Runtime-Agnostic Documentation & Tooling
**Milestone:** M003

## Description

Add Podman context to all 6 existing platform guides. Raspberry Pi and Proxmox LXC get "Using Podman" subsections (Podman is a viable alternative on these platforms). Synology, Unraid, and TrueNAS get brief notes that Podman is not applicable (platform uses native Docker integration). Mac Silicon gets an Apple Containers forward-reference pointing to the S03 guide.

Changes are additive — no rewriting of existing guide content. Each addition is a clearly delineated subsection.

## Steps

1. Update `deploy/platform-guides/raspberry-pi.md`:
   - Add "## Using Podman" section after the Docker deployment instructions
   - Cover: installing Podman on Raspberry Pi OS, rootful vs rootless for cloud relay / home device, override file usage, link to Podman platform guide for full details
2. Update `deploy/platform-guides/proxmox-lxc.md`:
   - Add "## Using Podman" section
   - Cover: Podman installation inside LXC container (alternative to Docker), note that Podman may be simpler in unprivileged LXC, link to Podman platform guide
3. Update `deploy/platform-guides/synology-nas.md`, `deploy/platform-guides/unraid.md`, `deploy/platform-guides/truenas-scale.md`:
   - Add brief note (3-5 lines) in a "Container Runtime" or "Alternative Runtimes" section explaining that the platform uses native Docker integration and Podman is not applicable
4. Update `deploy/platform-guides/mac-silicon.md`:
   - Add a note in prerequisites or near the Docker Desktop section pointing to the Apple Containers guide (coming in S03, link placeholder: `apple-containers.md`)
   - Do NOT document Apple Containers setup (that's S03 scope)

## Must-Haves

- [ ] Raspberry Pi guide has "Using Podman" section with override file reference
- [ ] Proxmox LXC guide has "Using Podman" section
- [ ] Synology, Unraid, TrueNAS guides note Podman is not applicable
- [ ] Mac Silicon guide has Apple Containers forward-reference
- [ ] No existing guide content is removed or rewritten
- [ ] All additions link to the Podman platform guide where relevant

## Verification

- `grep -qi "podman" deploy/platform-guides/raspberry-pi.md` confirms Podman section
- `grep -qi "podman" deploy/platform-guides/proxmox-lxc.md` confirms Podman section
- `grep -qi "apple containers\|apple-containers" deploy/platform-guides/mac-silicon.md` confirms forward reference
- `bash scripts/verify-s02-docs.sh` passes all checks with 0 failures (final green run)

## Observability Impact

- Signals added/changed: None (documentation only)
- How a future agent inspects this: grep for Podman mentions; run verify-s02-docs.sh
- Failure state exposed: None

## Inputs

- `deploy/platform-guides/raspberry-pi.md`, `deploy/platform-guides/proxmox-lxc.md`, `deploy/platform-guides/unraid.md`, `deploy/platform-guides/synology-nas.md`, `deploy/platform-guides/truenas-scale.md`, `deploy/platform-guides/mac-silicon.md` — current platform guides
- `deploy/platform-guides/podman.md` — link target (created in T02)
- S02 research — which platforms should have Podman subsections vs not-applicable notes

## Expected Output

- 6 updated platform guide files, each with appropriate Podman context
- `bash scripts/verify-s02-docs.sh` exits 0 — all slice acceptance criteria pass
