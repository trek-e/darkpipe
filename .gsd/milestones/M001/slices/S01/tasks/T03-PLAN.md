# T03: 01-transport-layer 03

**Slice:** S01 — **Milestone:** M001

## Description

Harden auto-reconnection, implement transport health monitoring, create integration test scripts for outage simulation, and document VPS provider selection for port 25 SMTP.

Purpose: The transport layer must survive real-world internet interruptions without manual intervention. This plan adds the monitoring and self-healing that turns a working tunnel into a resilient one. VPS provider documentation ensures users can deploy on providers that allow SMTP traffic.

Output: WireGuard health monitor (handshake-based), unified transport health checker, outage simulation script, tunnel connectivity test, VPS provider guide.

## Must-Haves

- [ ] "After simulating a home internet outage (disconnect for 60 seconds), the WireGuard tunnel automatically re-establishes and data flows resume without manual intervention"
- [ ] "WireGuard tunnel health is monitored by checking handshake timestamp age, alerting if older than 5 minutes"
- [ ] "A unified health checker reports status of whichever transport is active (WireGuard or mTLS)"

## Files

- `transport/wireguard/monitor/health.go`
- `transport/wireguard/monitor/health_test.go`
- `transport/wireguard/monitor/reconnect.sh`
- `transport/health/checker.go`
- `transport/health/checker_test.go`
- `deploy/test/tunnel-test.sh`
- `deploy/test/outage-sim.sh`
- `docs/vps-providers.md`
