# T03: 08-device-profiles-client-setup 03

**Slice:** S08 — **Milestone:** M001

## Description

Integrate the profile server into Docker Compose, build the webmail-accessible device management UI, add the CLI QR code command, and create the phase integration test suite.

Purpose: This is the user-facing integration plan that ties together all Phase 8 components into a complete device onboarding experience. Users interact through webmail ("Add Device" button) or CLI (`darkpipe qr`), and the infrastructure (Docker containers, Caddy routes, profile server) delivers the sub-2-minute device setup promise. The phase test suite validates all five PROF requirements.

Output: Complete device onboarding system with Docker deployment, web UI, CLI, and comprehensive test suite.

## Must-Haves

- [ ] "Users can access 'Add Device' page from webmail to download profiles and see QR codes"
- [ ] "Users can manage app passwords (list devices, revoke) from the webmail-accessible device management page"
- [ ] "CLI command 'darkpipe qr user@domain' generates and displays a QR code in the terminal"
- [ ] "Profile server runs as a Docker container alongside the mail server on the home device"
- [ ] "Phase integration test validates all five PROF requirements end-to-end"

## Files

- `home-device/profiles/Dockerfile`
- `home-device/docker-compose.yml`
- `cloud-relay/docker-compose.yml`
- `home-device/profiles/cmd/profile-server/cli.go`
- `home-device/profiles/cmd/profile-server/webui.go`
- `home-device/profiles/cmd/profile-server/templates/add_device.html`
- `home-device/profiles/cmd/profile-server/templates/device_list.html`
- `home-device/profiles/cmd/profile-server/static/style.css`
- `tests/test-device-profiles.sh`
