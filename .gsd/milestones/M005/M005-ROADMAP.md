# M005: Design Validation — External Access & Device Connectivity

**Vision:** Prove the DarkPipe architecture works end-to-end from the public internet — phones, laptops, and mail clients can reach the home mail server through the cloud relay from any network, with valid TLS, correct DNS, and stable tunnel connectivity.

## Success Criteria

- DNS records (MX, A, SPF, DKIM, DMARC, autoconfig, autodiscover) resolve correctly from external resolvers
- TLS certificates are valid and trusted by standard clients (no warnings)
- IMAP (993) and SMTP (587) accept authenticated connections from external networks
- Webmail loads over HTTPS from an external network
- Full inbound round-trip: external sender → cloud relay → tunnel → home mailbox → IMAP client
- Full outbound round-trip: mail client → SMTP → home → tunnel → cloud relay → external recipient
- Mobile device receives .mobileconfig profile and syncs email/calendar/contacts
- Monitoring dashboard shows all services healthy during external access
- WireGuard/mTLS tunnel reconnects automatically after brief interruption

## Key Risks / Unknowns

- Home network NAT/firewall may block WireGuard UDP — could require port forwarding configuration
- ISP may throttle or block VPN traffic — WireGuard uses UDP 51820 which is rarely blocked
- Let's Encrypt HTTP-01 challenge may fail if port 80 routing is incorrect
- VPS IP may be on blocklists affecting outbound delivery

## Proof Strategy

- NAT/firewall risk → retire in S01 by establishing and verifying the tunnel from a real residential network
- TLS/DNS risk → retire in S01 by validating all records and certificates from external resolvers
- Mail delivery risk → retire in S02 by completing a full inbound+outbound round-trip with external providers
- Device connectivity risk → retire in S03 by onboarding a real phone and verifying sync from cellular

## Verification Classes

- Contract verification: DNS record validation, TLS certificate chain validation, port reachability checks
- Integration verification: Full email round-trip through relay→tunnel→home→client chain
- Operational verification: Tunnel reconnection, service restart recovery, monitoring health
- UAT / human verification: Mobile device onboarding, webmail visual check, calendar sync

## Milestone Definition of Done

This milestone is complete only when all are true:

- All slices are complete with passing verification
- DNS validates from external resolvers (not just local)
- TLS certificates trusted by all tested clients
- Email round-trip proven (inbound and outbound) with real external providers
- At least one mobile device onboarded and syncing from external network
- Monitoring dashboard shows healthy status
- Any issues found are documented with fixes applied

## Requirement Coverage

- Covers: SMTP relay, encrypted transport, device onboarding, webmail, CalDAV/CardDAV, DNS auth, monitoring
- Partially covers: offline queuing (brief tunnel interruption test only)
- Leaves for later: IP warmup, migration, spam filter tuning, multi-user provisioning
- Orphan risks: none

## Slices

- [x] **S01: Infrastructure Validation — DNS, TLS & Tunnel** `risk:high` `depends:[]`
  > After this: DNS records validate from external resolvers, TLS certificates are trusted, WireGuard/mTLS tunnel is stable between cloud relay and home device, and all required ports (25, 587, 993, 443) respond from the public internet.

- [ ] **S02: Email Round-Trip — Inbound & Outbound Delivery** `risk:high` `depends:[S01]`
  > After this: An email sent from an external provider (Gmail/Outlook) arrives in the home mailbox via cloud relay, and an email sent from the home mail server is delivered to an external recipient — full bidirectional proof.

- [ ] **S03: Device Connectivity — Mobile, Desktop & Webmail** `risk:medium` `depends:[S02]`
  > After this: A phone on cellular data has a working mail account (via .mobileconfig or manual setup), can send/receive email, sync calendar/contacts, and access webmail. A desktop client (Thunderbird) connects via IMAP/SMTP from an external network. Monitoring dashboard confirms all healthy.

## Boundary Map

### S01 (independent)

Produces:
- Validated DNS records (MX, A, SPF, DKIM, DMARC, autoconfig, autodiscover, SRV)
- Valid TLS certificates (Let's Encrypt) on cloud relay and webmail endpoints
- Stable WireGuard/mTLS tunnel between cloud relay and home device
- Port reachability confirmed: 25, 587, 993, 443 from external network
- Checklist of any NAT/firewall/ISP issues encountered and resolved

Consumes:
- nothing (first slice)

### S02 (depends on S01)

Produces:
- Proven inbound delivery: external → relay → tunnel → home → mailbox
- Proven outbound delivery: home → tunnel → relay → external recipient
- SPF/DKIM/DMARC pass confirmation on outbound mail
- Delivery logs showing successful routing through all components
- IP reputation check (MXToolbox blocklist scan)

Consumes:
- S01's validated DNS, TLS, tunnel, and port reachability

### S03 (depends on S02)

Produces:
- Mobile device onboarded via .mobileconfig or manual config
- Email send/receive verified from phone on cellular data
- CalDAV/CardDAV sync verified from phone on cellular data
- Desktop mail client (Thunderbird) IMAP/SMTP verified from external network
- Webmail HTTPS access verified from external network
- Monitoring dashboard health check from external network
- Validation report documenting all tests, results, and any fixes applied

Consumes:
- S01's infrastructure (DNS, TLS, tunnel, ports)
- S02's proven email delivery (mail flow works end-to-end)
