# M001: MVP (Phases 1-10) — SHIPPED 2026-02-15

**Vision:** DarkPipe is a complete self-hosted email sovereignty stack.

## Success Criteria


## Slices

- [x] **S01: Transport Layer** `risk:medium` `depends:[]`
  > After this: Establish the WireGuard tunnel foundation: Go module initialization, WireGuard keypair generation, config file generation for hub (cloud) and spoke (home) topologies, and deployment scripts with systemd auto-restart.
- [x] **S02: Cloud Relay** `risk:medium` `depends:[S01]`
  > After this: Build the core cloud relay: a Postfix relay-only container with a Go SMTP relay daemon that bridges internet-facing SMTP to the home device via WireGuard or mTLS transport.
- [x] **S03: Home Mail Server** `risk:medium` `depends:[S02]`
  > After this: Set up the home device mail server foundation with all three selectable options (Stalwart, Maddy, Postfix+Dovecot), each providing inbound SMTP (port 25), IMAP (port 993), and SMTP submission (port 587).
- [x] **S04: Dns Email Auth** `risk:medium` `depends:[S03]`
  > After this: Generate DKIM keys, build SPF/DKIM/DMARC/MX DNS records, implement DKIM message signing, and produce human-readable DNS setup guides for manual configuration.
- [x] **S05: Queue Offline Handling** `risk:medium` `depends:[S04]`
  > After this: Build the encrypted in-memory message queue and QueuedForwarder wrapper that intercepts forwarding failures and queues messages encrypted with age when the home device is offline.
- [x] **S06: Webmail Groupware** `risk:medium` `depends:[S05]`
  > After this: Deploy Caddy reverse proxy on cloud relay and user-selectable webmail (Roundcube or SnappyMail) on home device, enabling browser-based email access remotely via HTTPS tunnel.
- [x] **S07: Build System Deployment** `risk:medium` `depends:[S06]`
  > After this: Update all Dockerfiles for multi-architecture builds with size optimization and setup detection, then create GitHub Actions workflows for custom component selection builds, pre-built default stack publishing, and semantic version releases.
- [x] **S08: Device Profiles Client Setup** `risk:medium` `depends:[S07]`
  > After this: Build the app password system and profile generation core libraries for DarkPipe device onboarding.
- [x] **S09: Monitoring Observability** `risk:medium` `depends:[S08]`
  > After this: Build the core monitoring data collection packages: health check framework with deep readiness probes, Postfix mail queue parser, and delivery status tracker with log parsing.
- [x] **S10: Mail Migration** `risk:medium` `depends:[S09]`
  > After this: Build the IMAP mail migration core engine: resumable state tracking, IMAP sync with date/flag preservation, and configurable folder mapping.
