# DarkPipe Security

This document describes DarkPipe's security model, threat model, encryption mechanisms, and vulnerability reporting process.

## Security Model

DarkPipe's core security promise is simple: **your mail is never stored on infrastructure you don't control.**

The cloud relay is intentionally designed as a pass-through gateway. Mail arrives, is immediately forwarded to your home device, and is not persisted on the cloud relay's disk. This architectural decision eliminates an entire class of security concerns related to third-party mail storage.

### Core Principles

1. **No persistent mail storage on cloud relay**
   - Mail is relayed immediately to home device
   - Ephemeral storage verification runs every 60 seconds
   - Offline queue is encrypted at rest (see Offline Queue section)

2. **Encrypted transport at all times**
   - Internet to cloud: TLS (Let's Encrypt certificates)
   - Cloud to home: WireGuard full tunnel OR mTLS
   - Home to clients: TLS (IMAPS port 993, SMTPS port 465/587)

3. **Internal PKI for certificate management**
   - step-ca provides internal certificate authority
   - Automatic certificate rotation (configurable: 30/60/90 days)
   - No reliance on external CAs for internal communication

4. **Email authentication standards**
   - SPF (Sender Policy Framework) - authorized senders
   - DKIM (DomainKeys Identified Mail) - cryptographic signing
   - DMARC (Domain-based Message Authentication, Reporting & Conformance) - policy enforcement
   - All configured automatically by dns-setup tool

5. **Container hardening**
   - All containers run with `no-new-privileges` security option
   - Linux capabilities dropped (`cap_drop: ALL`) with selective re-add only where needed
   - Read-only root filesystems (`read_only: true`) with explicit tmpfs mounts for writable paths
   - Container HEALTHCHECK instructions in all custom Dockerfiles
   - Podman's rootless mode provides additional security isolation by running containers without root privileges (see [Podman platform guide](../deploy/platform-guides/podman.md))

6. **PII-safe logging**
   - Email addresses are redacted in logs at default verbosity (e.g., `s***r@example.com`)
   - Tokens and credentials never appear in log output
   - Debug-level logging (with full PII) available via `RELAY_DEBUG` and `PROFILE_DEBUG` env vars

7. **TLS version enforcement**
   - All IMAP provider connections enforce TLS 1.2 minimum (`MinVersion: tls.VersionTLS12`)
   - SMTP relay enforces configurable maximum message size (default 50MB)

## Container Security Hardening

All DarkPipe Docker containers are hardened with defense-in-depth security measures.

### Linux Capabilities

All containers drop all Linux capabilities (`cap_drop: ALL`) and selectively re-add only what is required:

| Service | Capabilities Added | Justification |
|---------|-------------------|---------------|
| relay (Postfix) | NET_ADMIN, NET_BIND_SERVICE, DAC_OVERRIDE, CHOWN, SETGID, SETUID, KILL | Root required for Postfix MTA (privileged port 25, mail delivery) |
| caddy | NET_BIND_SERVICE | Binds ports 80/443 |
| certbot | NET_BIND_SERVICE | Binds port 80 for ACME challenges |
| stalwart | NET_BIND_SERVICE | Binds ports 25, 587, 993 |
| maddy | NET_BIND_SERVICE | Binds ports 25, 587, 993 |
| postfix-dovecot | NET_BIND_SERVICE, DAC_OVERRIDE, CHOWN, SETGID, SETUID, KILL | Root required for Postfix mail delivery |
| roundcube, snappymail, rspamd, redis, radicale, profile-server | (none) | No privileged operations needed |

### Read-Only Filesystems

All containers use `read_only: true` root filesystems. Writable paths are provided as explicit tmpfs mounts:

```yaml
read_only: true
tmpfs:
  - /tmp
  - /run
```

Mail servers and Postfix have additional tmpfs mounts for spool directories.

### Security Options

All containers include:
```yaml
security_opt:
  - no-new-privileges:true
```

This prevents privilege escalation via setuid/setgid binaries inside containers.

### Health Checks

All 5 custom Dockerfiles include container HEALTHCHECK instructions:
- **cloud-relay, postfix-dovecot, maddy:** `nc -z localhost 25` (SMTP port check)
- **stalwart:** `curl --silent --fail http://localhost:8080/` (management API check)
- **profile-server:** HTTP health endpoint check

All use consistent timing: `--interval=30s --timeout=10s --start-period=10s --retries=3`

### Verification

Run the security audit script to verify all container hardening directives:

```bash
bash scripts/verify-container-security.sh
```

This checks all compose files for `security_opt`, `cap_drop`, and `read_only` on every service, and all Dockerfiles for `HEALTHCHECK`. Returns exit code 0 only if all checks pass.

## Log Hygiene and PII Protection

DarkPipe redacts personally identifiable information (PII) from logs at default verbosity.

### What's Redacted

- **Email addresses:** Local-part is masked, domain preserved (e.g., `sender@example.com` → `s***r@example.com`). Domain is preserved because it's needed for multi-domain debugging.
- **Query string parameters:** Email and token parameters in HTTP access logs are replaced with `[REDACTED]`.
- **Tokens:** Only the first 8 characters of tokens are logged (256-bit tokens, not reconstructible).
- **Credentials:** Never logged at any verbosity level.

### Debug Mode

For troubleshooting, full PII can be restored per-binary:

| Variable | Binary | Default |
|----------|--------|---------|
| `RELAY_DEBUG=true` | Cloud relay SMTP service | `false` |
| `PROFILE_DEBUG=true` | Profile server | `false` |

**Warning:** Debug mode logs full email addresses. Use only for active troubleshooting, not in production.

### Verification

Run the static analysis script to check for unredacted PII patterns in log call sites:

```bash
bash scripts/verify-log-redaction.sh
```

## Encryption in Transit

### Internet to Cloud Relay

**Protocol:** TLS 1.2+ (STARTTLS on port 25, implicit TLS on port 465)

**Certificates:** Let's Encrypt via Certbot (automatic renewal)

**Configuration:**
- TLS required on all inbound SMTP connections
- Optional strict mode: reject connections that don't support TLS
- Cipher suites: Modern cipher suite selection (no SSLv3, TLS 1.0, or weak ciphers)

**Monitoring:**
- Certbot monitors certificate expiry
- Alerts sent when certificates are within 7 days of expiry
- TLS negotiation failures logged and optionally sent to webhook

### Cloud Relay to Home Device

DarkPipe supports two transport security options: WireGuard (recommended) or mTLS.

#### WireGuard Transport

**Protocol:** WireGuard (ChaCha20-Poly1305 encryption, Curve25519 key exchange)

**Architecture:**
- Full tunnel VPN between cloud and home
- Network: 10.8.0.0/24 (cloud: 10.8.0.1, home: 10.8.0.2)
- All traffic between cloud and home flows through encrypted tunnel
- Kernel-level encryption (not userspace)

**Key Management:**
- Each endpoint has a private key and public key
- Keys generated during setup, not shared
- No PKI needed (public key-based authentication)

**Security Properties:**
- Forward secrecy (ephemeral key exchange)
- Replay protection
- Identity hiding (encrypted handshake)
- NAT traversal built-in

#### mTLS Transport

**Protocol:** Mutual TLS 1.3

**Architecture:**
- TCP connection with certificate-based mutual authentication
- Both cloud relay and home device present certificates
- Certificates issued by internal CA (step-ca)

**Certificate Management:**
- Internal CA (step-ca) runs on home device
- Cloud relay certificate issued by internal CA
- Home device certificate issued by internal CA
- Automatic rotation (default: 30 days, configurable: 60/90 days)

**Security Properties:**
- Forward secrecy (via TLS 1.3 key exchange)
- Mutual authentication (both endpoints verify each other)
- No trusted third party (internal CA)
- Certificate-based identity

### Home Device to Mail Clients

**Protocol:** TLS 1.2+ (IMAPS port 993, SMTPS port 465/587)

**Certificates:** Let's Encrypt via Caddy (automatic renewal)

**Configuration:**
- TLS mandatory on IMAP (no plaintext IMAP on port 143)
- SMTP submission requires STARTTLS or implicit TLS
- Modern cipher suites only

**Autodiscovery:**
- Thunderbird autoconfig.xml served over HTTPS
- Outlook autodiscover.xml served over HTTPS
- Apple .mobileconfig signed and served over HTTPS

## Encryption at Rest

### Offline Queue Encryption

When the home device is offline, the cloud relay queues mail encrypted on local disk or S3-compatible storage.

**Encryption:** age (filippo.io/age) with X25519 recipient key

**Process:**
1. Mail arrives at cloud relay
2. Home device is unreachable
3. Mail encrypted with age using recipient public key
4. Encrypted message written to disk (or S3 if overflow enabled)
5. Original plaintext mail deleted from Postfix queue
6. When home device reconnects, encrypted message is decrypted and delivered
7. Encrypted message deleted after successful delivery

**Key Management:**
- age identity key generated during setup
- Private key stored on cloud relay (in /data/queue-keys/identity)
- Public key derived from identity
- Keys never transmitted (generated locally on cloud relay)

**Security Properties:**
- Encrypted messages are safe to store on disk or S3
- Encryption is resistant to quantum attacks (X25519)
- No key escrow (only cloud relay can decrypt)

**Note:** Offline queue encryption does NOT provide end-to-end encryption. The cloud relay can decrypt messages (it must, to deliver them). For E2EE, use PGP or S/MIME at the application layer.

### Mail Storage on Home Device

Mail storage encryption depends on your mail server configuration and underlying filesystem.

**DarkPipe does NOT enforce encryption at rest for mail storage.** This is the user's responsibility.

**Recommendations:**
- Use full-disk encryption (LUKS on Linux, FileVault on macOS)
- Enable filesystem encryption (ext4 encryption, ZFS encryption, Btrfs encryption)
- Encrypt Docker volumes (via LUKS or dm-crypt)

**Mail server specifics:**
- **Stalwart:** Stores mail in data directory, supports database encryption (configure in config.toml)
- **Maddy:** Stores mail in maildir format, no built-in encryption (use filesystem encryption)
- **Postfix+Dovecot:** Stores mail in maildir format, no built-in encryption (use filesystem encryption)

## Email Authentication

DarkPipe implements industry-standard email authentication mechanisms to prevent spoofing and improve deliverability.

### SPF (Sender Policy Framework)

**Purpose:** Authorize which mail servers can send mail for your domain

**Record example:**
```
example.com. IN TXT "v=spf1 mx -all"
```

**Interpretation:**
- `v=spf1` - SPF version 1
- `mx` - Mail can be sent from servers listed in MX records
- `-all` - Reject mail from any other servers (hard fail)

**Generated by:** dns-setup tool

### DKIM (DomainKeys Identified Mail)

**Purpose:** Cryptographically sign outbound mail to prove authenticity

**Process:**
1. Outbound mail is signed with private key (stored on cloud relay)
2. Public key is published in DNS
3. Receiving server verifies signature using public key from DNS

**Key size:** 2048-bit RSA (industry standard)

**Rotation:** dns-setup tool supports key rotation (generate new key, update DNS, switch signing key)

**Record example:**
```
default._domainkey.example.com. IN TXT "v=DKIM1; k=rsa; p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC..."
```

### DMARC (Domain-based Message Authentication, Reporting & Conformance)

**Purpose:** Define policy for handling mail that fails SPF/DKIM checks

**Record example:**
```
_dmarc.example.com. IN TXT "v=DMARC1; p=quarantine; rua=mailto:postmaster@example.com"
```

**Interpretation:**
- `p=quarantine` - Quarantine mail that fails SPF/DKIM (soft fail)
- `rua=mailto:postmaster@example.com` - Send aggregate reports to this address

**Policy options:**
- `p=none` - Monitor only (no action taken)
- `p=quarantine` - Quarantine suspicious mail (recommended for new deployments)
- `p=reject` - Reject mail that fails SPF/DKIM (strict, use after IP warmup)

**Generated by:** dns-setup tool

## Spam Filtering

DarkPipe uses Rspamd for spam filtering with greylisting.

**Rspamd:**
- Modern spam filtering engine
- Bayesian classifier (learns from spam and ham samples)
- DNS blocklists (RBL checks)
- Header and body analysis
- Configurable thresholds

**Greylisting:**
- Temporary rejection of mail from unknown senders
- Legitimate mail servers retry, spammers typically don't
- Redis backend for greylist state
- Reduces spam by 70-90% without false positives

**Configuration:**
- Default thresholds: reject score 15+, greylist score 6-15, allow score < 6
- Customize via home-device/spam-filter/rspamd/local.d/
- Web UI for statistics and management (port 11334)

## Threat Model

DarkPipe is designed to protect against specific threats. Understanding what DarkPipe does and does NOT protect against is critical for proper security evaluation.

### What DarkPipe Protects Against

**1. Cloud provider reading your mail**
- Traditional email services (Gmail, Outlook, ProtonMail) store your mail on their servers
- Providers can read your mail (for ads, analysis, legal requests, or data breaches)
- DarkPipe eliminates this: mail is stored only on your hardware

**2. Mass surveillance of stored mail**
- Intelligence agencies or attackers targeting mail storage at providers
- DarkPipe separates relay from storage - no mail stored on cloud relay
- Attackers must target individual home devices (not scalable for mass surveillance)

**3. Third-party data breaches of mail storage**
- Data breaches at email providers expose millions of mailboxes
- DarkPipe: no mail stored on cloud relay means no mail to breach
- Your mail is only vulnerable if your home device is compromised (under your control)

**4. Unauthorized access to historical email**
- Cloud providers retain mail indefinitely (for legal/compliance reasons)
- DarkPipe: you control retention policy on your home device
- Delete mail permanently when you want, no provider retention

### What DarkPipe Does NOT Protect Against

**1. Compromised home device**
- If your home device is compromised (malware, unauthorized physical access), your mail is accessible
- Mitigation: Use strong passwords, keep software updated, enable disk encryption, physical security

**2. Nation-state targeting of your specific VPS**
- Advanced attackers can intercept mail in transit at the cloud relay
- While DarkPipe encrypts transport, a compromised cloud relay can decrypt mail before forwarding
- Mitigation: Use end-to-end encryption (PGP, S/MIME) if you're a high-risk target

**3. Malware on your endpoints (mail clients)**
- Compromised phone, laptop, or desktop can read mail via IMAP
- DarkPipe does not protect endpoint security
- Mitigation: Keep devices updated, use antivirus, avoid untrusted software

**4. Social engineering**
- Phishing, impersonation, credential theft, or manipulation
- DarkPipe cannot prevent users from being tricked into revealing credentials or clicking malicious links
- Mitigation: User education, strong passwords, 2FA where possible

**5. Man-in-the-middle attacks on endpoints**
- If an attacker compromises your network or device to intercept TLS
- DarkPipe uses TLS, but compromised endpoints can't verify certificates
- Mitigation: Secure your network (WPA3, VPN), verify TLS certificates

**6. Traffic analysis**
- Timing, volume, and metadata analysis can reveal communication patterns
- DarkPipe does not hide metadata (sender, recipient, timestamp, size)
- Mitigation: Use Tor or VPN for relay connections (not officially supported)

### Cloud Relay Trust Assumptions

**The cloud relay is a trusted component.**

If the cloud relay VPS is compromised, an attacker could:
- Read mail in transit (before encryption or after decryption)
- Modify mail in transit
- Redirect mail to different destinations
- Decrypt offline queue (age private key is on cloud relay)

**What an attacker CANNOT do even if cloud relay is compromised:**
- Access stored mail on home device (none is stored on cloud relay)
- Decrypt home device mail storage (keys are on home device)
- Impersonate your domain for outbound mail (requires DKIM private key rotation)

**Mitigations:**
- Choose a reputable VPS provider
- Harden cloud relay OS (firewall, automatic updates, minimal attack surface)
- Monitor cloud relay logs for suspicious activity
- Rotate DKIM keys periodically
- Use end-to-end encryption (PGP/S/MIME) for sensitive mail

## Reporting Vulnerabilities

DarkPipe takes security vulnerabilities seriously.

### How to Report

**DO NOT open a public GitHub issue for security vulnerabilities.**

Instead, use one of these private reporting methods:

**Option 1: Email**
- Send to: security@darkpipe.org (monitored by maintainers)
- Include: Detailed description, reproduction steps, impact assessment, suggested fix (if any)
- PGP key available at: https://darkpipe.org/.well-known/pgp-key.txt (TODO: publish key)

**Option 2: GitHub Security Advisories**
- Go to: https://github.com/trek-e/darkpipe/security/advisories
- Click "Report a vulnerability"
- Fill out private vulnerability report form
- This is preferred for GitHub-hosted projects

### What to Include

**Good vulnerability reports contain:**
- Vulnerability type (e.g., SQL injection, XSS, privilege escalation, authentication bypass)
- Affected component (cloud relay, home device, specific service)
- Affected versions (e.g., "v1.0.0 and earlier")
- Steps to reproduce (detailed, unambiguous)
- Proof of concept (code, curl command, screenshots)
- Impact assessment (what can an attacker do?)
- Suggested fix (if you have one)

### Response Timeline

**Acknowledgment:** 48 hours for initial response

**Triage:** 7 days to confirm and assess severity

**Fix:** Timeframe depends on severity:
- Critical (remote code execution, authentication bypass): 7-14 days
- High (privilege escalation, data exposure): 14-30 days
- Medium (DoS, minor information disclosure): 30-60 days
- Low (non-exploitable issues): 60-90 days

**Disclosure:** 90 days after initial report (or when fix is released, whichever is sooner)

### Coordinated Disclosure

We follow coordinated disclosure practices:

- Vulnerabilities are fixed privately before public disclosure
- Reporter is credited (if desired) in release notes and SECURITY.md
- Public disclosure happens AFTER fix is released and users have time to update (minimum 7 days)
- If you publicly disclose before coordinated disclosure timeline, we may fast-track public disclosure

### Bounty Program

DarkPipe does not currently offer a bug bounty program. This may change in the future if the project secures funding.

Reporters will be credited in release notes and SECURITY.md.

## Security Best Practices

**For Cloud Relay:**
- Use a reputable VPS provider with good security track record
- Enable automatic security updates (unattended-upgrades on Ubuntu/Debian)
- Configure firewall to allow only necessary ports (25, 80, 443, 51820 for WireGuard)
- Disable SSH password authentication (use SSH keys only)
- Monitor logs regularly for suspicious activity (logs redact PII by default — use `RELAY_DEBUG=true` only when actively troubleshooting)
- Use strong passwords/keys for all services
- Run `bash scripts/verify-container-security.sh` periodically to audit container hardening

**For Home Device:**
- Enable full-disk encryption (LUKS, FileVault, etc.)
- Use strong passwords for admin accounts
- Keep Docker and host OS updated
- Configure firewall (allow only necessary ports from local network)
- Physical security: lock server room/closet where home device is stored
- Regular backups (offline or encrypted cloud backups)

**For Users:**
- Use strong, unique passwords for mail accounts
- Enable app-specific passwords for mail clients (not main password)
- Keep mail clients updated (Thunderbird, Outlook, phone apps)
- Be cautious of phishing emails (verify sender before clicking links)
- Use 2FA for admin interfaces where available

## Dependencies

DarkPipe uses many open source dependencies. Security of the overall system depends on security of these dependencies.

**Key dependencies and their security status:**

- **Go standard library:** Regular security updates via Go releases
- **emersion/go-smtp, go-imap, go-msgauth:** Maintained, security-conscious author
- **Postfix, Dovecot:** Decades of security hardening, active maintenance
- **Stalwart:** Pre-v1.0 (monitor for security updates, v1.0 expected Q2 2026)
- **Maddy:** Stable, Go-based (memory-safe language)
- **Rspamd:** Active security maintenance
- **WireGuard:** Audited, secure-by-design
- **age encryption:** Audited, modern design by Filippo Valsorda
- **Caddy, Certbot:** Active security maintenance

**Dependency management:**
- All Go dependencies listed in THIRD-PARTY-LICENSES.md
- Compatible licenses only (MIT, BSD, Apache 2.0)
- Regular updates via `go get -u` and manual testing

**Known dependency risks:**

**go-imap v2 is beta (v2.0.0-beta.8)**
- Risk: Beta software may have undiscovered bugs or security issues
- Mitigation: Monitor for updates, plan to update when v2 stable is released
- Used only in migration tool (not core mail flow)

**Stalwart is pre-v1.0 (0.15.4)**
- Risk: Pre-v1.0 software may have schema changes or security updates
- Mitigation: Monitor Stalwart releases, provide update path for users
- Alternative: Use Postfix+Dovecot for production-critical deployments

## Security Updates

Security updates are released as patch versions (e.g., v1.0.1, v1.0.2).

**Users are strongly encouraged to:**
- Watch the GitHub repository for security advisories
- Update to latest patch version within 7 days of release
- Subscribe to release notifications (GitHub "Watch" > "Releases only")

**Update process:**
1. Pull new Docker images: `docker pull ghcr.io/trek-e/darkpipe/cloud-relay:vX.Y.Z`
2. Update docker-compose.yml to reference new version
3. Restart services: `docker compose up -d`
4. Verify services are healthy: `docker compose ps`

---

Last Updated: 2026-03-12

License: AGPLv3 - See [LICENSE](../LICENSE)
