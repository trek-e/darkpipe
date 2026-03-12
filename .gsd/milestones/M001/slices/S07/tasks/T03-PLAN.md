# T03: 07-build-system-deployment 03

**Slice:** S07 — **Milestone:** M001

## Description

Create native app templates for TrueNAS Scale and Unraid, write platform-specific deployment guides for all target platforms, and build the Phase 7 integration test suite.

Purpose: Deliver UX-03 (runs on RPi4, x64 Docker, TrueNAS Scale, Unraid) with native platform integration and clear guides for all target platforms. The phase test suite validates all Phase 7 artifacts work correctly.

Output: TrueNAS Scale YAML app template, Unraid XML template, six platform deployment guides, and a comprehensive integration test script.

## Must-Haves

- [ ] "TrueNAS Scale users can deploy DarkPipe as a custom app using the provided YAML template with a form-based UI for configuration"
- [ ] "Unraid users can install DarkPipe from Community Applications using the provided XML template"
- [ ] "Platform-specific deployment guides exist for RPi4, TrueNAS Scale, Unraid, Proxmox LXC, Synology NAS, and Mac Silicon"
- [ ] "Phase integration tests validate that Dockerfiles build, workflows are syntactically valid, setup tool compiles, and platform templates are well-formed"

## Files

- `deploy/templates/truenas-scale/app.yaml`
- `deploy/templates/truenas-scale/questions.yaml`
- `deploy/templates/unraid/darkpipe.xml`
- `deploy/platform-guides/raspberry-pi.md`
- `deploy/platform-guides/truenas-scale.md`
- `deploy/platform-guides/unraid.md`
- `deploy/platform-guides/proxmox-lxc.md`
- `deploy/platform-guides/synology-nas.md`
- `deploy/platform-guides/mac-silicon.md`
- `tests/test-phase-07.sh`
