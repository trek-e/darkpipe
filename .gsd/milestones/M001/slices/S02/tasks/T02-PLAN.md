# T02: 02-cloud-relay 02

**Slice:** S02 — **Milestone:** M001

## Description

Add TLS enforcement, strict mode, notification system, and Let's Encrypt certificate automation to the cloud relay.

Purpose: This plan fulfills RELAY-04 (TLS enforced on all connections), RELAY-05 (optional strict mode to refuse plaintext-only peers), RELAY-06 (user notified when remote server lacks TLS), and CERT-01 (Let's Encrypt certificates for public-facing TLS). Without TLS, the relay cannot participate in modern email delivery where major providers (Gmail, Outlook) expect encrypted connections.

Output: TLS-enabled Postfix with automated Let's Encrypt certificates, configurable strict mode, and real-time TLS quality monitoring with webhook notifications.

## Must-Haves

- [ ] "Postfix offers STARTTLS on port 25 using Let's Encrypt certificates"
- [ ] "When strict mode is enabled and a remote server cannot negotiate TLS, the connection is refused"
- [ ] "When a remote server connects without TLS support, a notification is emitted with the offending domain"
- [ ] "Let's Encrypt certificates are obtained automatically via certbot and renewed on a 12-hour cycle"
- [ ] "Postfix reloads TLS certificates after certbot renewal without service restart"

## Files

- `cloud-relay/relay/tls/monitor.go`
- `cloud-relay/relay/tls/monitor_test.go`
- `cloud-relay/relay/tls/strict.go`
- `cloud-relay/relay/tls/strict_test.go`
- `cloud-relay/relay/notify/notifier.go`
- `cloud-relay/relay/notify/notifier_test.go`
- `cloud-relay/relay/notify/webhook.go`
- `cloud-relay/certbot/docker-compose.certbot.yml`
- `cloud-relay/certbot/renew-hook.sh`
- `cloud-relay/postfix-config/main.cf`
- `cloud-relay/docker-compose.yml`
- `cloud-relay/relay/config/config.go`
- `cloud-relay/cmd/relay/main.go`
