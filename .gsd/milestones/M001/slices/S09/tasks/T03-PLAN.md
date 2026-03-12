# T03: 09-monitoring-observability 03

**Slice:** S09 — **Milestone:** M001

## Description

Build the status aggregation layer, CLI command, web dashboard, push-based external monitoring, Docker health checks, and phase integration test suite. This plan wires all monitoring packages together into user-facing interfaces.

Purpose: This is the integration plan that makes monitoring visible to users via CLI (power users) and web dashboard (household members). It also adds Docker health check integration, Caddy health endpoint with Basic Auth, push-based external monitoring, and the comprehensive phase test suite.

Output: Status aggregator package, CLI command, web dashboard on profile server, Docker health checks, Caddy route, push pinger, and integration test suite.

## Must-Haves

- [ ] "User can see mail queue depth and stuck/deferred message count at a glance"
- [ ] "User can check delivery status of recent outbound messages (delivered, deferred, bounced)"
- [ ] "Cloud relay container exposes health check endpoints that return pass/fail status"
- [ ] "Certificate rotation is configurable and rotations happen automatically without service interruption"
- [ ] "User receives an alert at least 14 days before any certificate expires, and again at 7 days if not renewed"
- [ ] "darkpipe status CLI shows all four metric categories with JSON output option"
- [ ] "Web dashboard on profile server shows glanceable system health for household members"
- [ ] "Push-based pings to external uptime services work without exposing inbound ports"

## Files

- `monitoring/status/aggregator.go`
- `monitoring/status/cli.go`
- `monitoring/status/dashboard.go`
- `monitoring/status/push.go`
- `monitoring/status/aggregator_test.go`
- `monitoring/status/cli_test.go`
- `monitoring/status/push_test.go`
- `home-device/profiles/cmd/profile-server/main.go`
- `home-device/profiles/cmd/profile-server/templates/status.html`
- `home-device/docker-compose.yml`
- `cloud-relay/docker-compose.yml`
- `cloud-relay/caddy/Caddyfile`
- `tests/test-monitoring.sh`
