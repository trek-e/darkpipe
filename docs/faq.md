# DarkPipe Frequently Asked Questions

Common questions about DarkPipe, self-hosted email, setup, security, and operations.

## General

### What is DarkPipe?

DarkPipe is a complete self-hosted email infrastructure that separates internet-facing mail services (running on a VPS) from persistent mail storage (running on hardware you control at home). Mail arrives at the cloud relay, is immediately forwarded over an encrypted tunnel to your home device, and is never stored on the cloud. This gives you the privacy of self-hosted email without the deliverability challenges of running SMTP directly from a residential IP.

### Why not just use Gmail / Outlook / ProtonMail?

**Privacy:** Gmail and Outlook read your mail for advertising and analysis. ProtonMail is better for privacy, but you're still trusting a third party with your data. DarkPipe puts mail storage entirely under your physical control.

**Control:** With DarkPipe, you own your email infrastructure. No service can lock you out, change terms of service, or read your mail. Your data stays on hardware you control.

**Sovereignty:** Email is critical infrastructure. DarkPipe gives you true email sovereignty - independence from any provider.

**Trade-off:** DarkPipe requires more technical skill than clicking "sign up" on Gmail. You need to manage your own server, troubleshoot issues, and keep software updated. It's a trade-off: convenience vs sovereignty.

### How is this different from Mail-in-a-Box / Mailu / docker-mailserver?

Those are excellent projects, but they run everything on a single server (typically a VPS). This means:

- All your mail is stored on a VPS you don't physically control
- If the VPS provider has a data breach, your mail is exposed
- VPS providers can be compelled to hand over data

**DarkPipe's key difference:** Split architecture. The cloud relay is a pass-through gateway with no persistent storage. Your mail lives only on your home device, which you physically control.

**Similarity:** Like those projects, DarkPipe automates the hard parts of email (DNS, TLS, deliverability).

### Is self-hosted email hard?

**Honestly, yes - but DarkPipe automates the hardest parts.**

Email is one of the most complex internet protocols. Running a mail server traditionally required:
- Deep understanding of SMTP, DNS, TLS, SPF, DKIM, DMARC
- Manual configuration of Postfix, Dovecot, spam filters
- Constant vigilance for security updates
- Dealing with deliverability issues (spam folders, blocklists)

**What DarkPipe automates:**
- DNS configuration (SPF, DKIM, DMARC records)
- TLS certificate management (Let's Encrypt)
- Transport encryption (WireGuard or mTLS)
- Spam filtering (Rspamd with greylisting)
- Device onboarding (QR codes, profiles)
- Migration from existing providers

**What you still need:**
- Basic container knowledge (docker compose up, logs, restart)
- Understanding of DNS (or willingness to learn)
- A VPS with port 25 access (see VPS provider guide)
- A home device to run Docker (Raspberry Pi, NAS, or Linux PC)
- Patience during IP warmup (4-6 weeks for new IP to build reputation)

**Bottom line:** If you're comfortable with command-line tools and willing to learn, DarkPipe makes self-hosted email achievable. If "docker compose up" sounds intimidating, stick with Gmail.

### Will my emails go to spam?

**Initially, yes - but this improves over time.**

New IP addresses have no sending reputation. Receiving mail servers treat mail from unknown IPs with suspicion. This is normal and happens with ANY new mail server (not just DarkPipe).

**IP warmup process (4-6 weeks):**
- Week 1: Send 10-20 emails/day to engaged recipients (people you know)
- Week 2: 50-100 emails/day
- Week 3: 200-500 emails/day
- Week 4-6: Gradually increase to normal volume

**What helps deliverability:**
- Proper DNS configuration (SPF, DKIM, DMARC) - DarkPipe automates this
- Clean IP (not on blocklists) - check with MXToolbox before deployment
- Reverse DNS (PTR record) matching your hostname - set via VPS provider
- Consistent sending patterns (don't spike from 0 to 1000 emails overnight)
- Engaged recipients (people who open and respond to your mail)

**What hurts deliverability:**
- Sending to purchased email lists or cold contacts
- High spam complaint rate
- Sudden volume spikes
- Poor email content (spammy subject lines, too many links)

**Realistic expectations:**
- Gmail/Outlook will be cautious for first 2-4 weeks
- Smaller providers (self-hosted) are usually more lenient
- After warmup, deliverability should match traditional providers

**Monitoring:** Use mail-tester.com to check your deliverability score. Aim for 9/10 or higher.

## Setup and Deployment

### What VPS provider should I use?

See [docs/vps-providers.md](vps-providers.md) for full compatibility matrix.

**Recommended providers:**
- **Hetzner** (best price/performance, EU-based, SMTP-friendly)
- **Vultr** (port 25 after account verification, multiple regions)
- **OVH** (SMTP-friendly, good EU presence)

**Avoid:**
- **DigitalOcean** (blocks port 25, no exceptions)
- **Google Cloud** (blocks port 25, no exceptions)
- **AWS EC2** (port 25 throttled by default, requires support ticket)

**Key requirement:** Port 25 must be open for SMTP. Many providers block it to prevent spam. Check provider documentation before purchasing.

### Can I run everything on one server?

**No - and this is intentional.**

DarkPipe's security model requires separating the cloud relay from mail storage. Running everything on one server defeats the privacy purpose:

- If everything is on a VPS, your mail is stored on hardware you don't control
- VPS provider can access your data (legally or via breach)
- This is the problem DarkPipe solves

**You MUST have:**
1. Cloud relay on a VPS (for internet-facing SMTP and deliverability)
2. Home device on your hardware (for mail storage and privacy)

**Why this architecture:**
- Residential IPs are often blocklisted for SMTP (ISPs block port 25)
- Cloud relay ensures deliverability while preserving privacy
- Your mail lives only on hardware you physically control

### What home hardware do I need?

**Minimum:**
- Raspberry Pi 4 (4GB RAM recommended, 2GB possible with optimization)
- 20GB storage (SSD strongly recommended over SD card)
- Ethernet connection (Wi-Fi works but not recommended)

**Recommended:**
- Any x64 or arm64 system with 4GB+ RAM
- SSD or NVMe storage (mail servers do lots of small random I/O)
- Wired gigabit Ethernet
- Uninterruptible Power Supply (UPS) for uptime

**Supported platforms:**
- Raspberry Pi 4+ (arm64)
- NAS: TrueNAS Scale, Unraid, Synology
- Proxmox LXC containers
- Mac Silicon (M-series Apple computers)
- Any Docker-capable Linux (Ubuntu, Debian, Fedora, etc.)

See [deploy/platform-guides/](../deploy/platform-guides/) for platform-specific instructions.

### Can I use Podman instead of Docker?

**Yes, fully supported.** DarkPipe works with Podman 5.3+ and podman-compose using the provided override files.

See the [Podman platform guide](../deploy/platform-guides/podman.md) for complete setup instructions. Key differences from Docker:

- **Cloud relay requires rootful Podman** — port 25 binding needs root privileges
- **Home device can run rootless** — additional security isolation over Docker
- **Override files required** — use `docker-compose.podman.yml` alongside the base compose file to adjust volume mounts and security options
- **SELinux systems** — override files include `:Z` volume labels for proper SELinux context

Run `bash scripts/check-runtime.sh` to validate your Podman installation meets all prerequisites.

### How much does it cost?

**Monthly costs:**
- VPS: $3-6/month (1 vCPU, 1GB RAM)
- Domain: ~$10/year (~$1/month)
- Home hardware: varies (if you already own it, $0; Raspberry Pi 4 starter kit ~$100 one-time)
- Electricity: negligible (Raspberry Pi uses ~5-10W, ~$1/month)

**Total ongoing cost: $4-7/month**

Compare to:
- Gmail: Free (but you're the product)
- ProtonMail Plus: $4/month
- Fastmail: $5/month
- Microsoft 365: $6/month

**DarkPipe is competitive on cost, but requires your time to manage.**

## Mail Servers

### Which mail server should I choose?

Depends on your priorities:

**Stalwart (Default, recommended for most users):**
- Modern, written in Rust (memory-safe)
- All-in-one: IMAP, SMTP, CalDAV, CardDAV, web admin UI
- JMAP support (next-gen mail protocol)
- Best if: You want a modern, feature-rich server with minimal configuration

**Maddy (Minimal, for low-resource systems):**
- Single Go binary, minimal dependencies
- Suitable for 2GB RAM systems
- No built-in CalDAV/CardDAV (use Radicale)
- Best if: You want minimal resource usage or prefer Go-based software

**Postfix + Dovecot (Battle-tested, conservative):**
- Decades of production use
- Extensive documentation and community knowledge
- Separate MTA (Postfix) and IMAP (Dovecot)
- No built-in CalDAV/CardDAV (use Radicale)
- Best if: You want proven, traditional software with maximum community support

**Can't decide?** Start with Stalwart. It's the most complete and easiest to manage.

### Can I switch mail servers later?

**Yes, but it requires migrating mail data.**

Mail server formats are not identical. Switching requires:

1. Re-run darkpipe-setup wizard with new mail server choice
2. Migrate mail from old server to new server (via IMAP or maildir export/import)
3. Update DNS (DKIM keys change with mail server)
4. Test thoroughly before decommissioning old server

**Easier path:** Choose carefully at the start. Switching is possible but not trivial.

### Does DarkPipe support POP3?

**No. POP3 is explicitly out of scope.**

**Why:**
- POP3 downloads mail to client and deletes from server (by default)
- This is a security liability: mail exists only on endpoint (phone/laptop)
- If device is lost/stolen/broken, mail is lost
- POP3 is a 1990s protocol; IMAP is the modern standard (synchronizes mail across devices)

**Use IMAP instead.** All DarkPipe mail servers support IMAP (port 993).

## Security and Privacy

### Is my email encrypted end-to-end?

**Transport is encrypted. At-rest encryption depends on your home device setup.**

**What's encrypted:**
- Internet to cloud relay: TLS (Let's Encrypt)
- Cloud relay to home device: WireGuard or mTLS
- Home device to mail clients: TLS (IMAPS/SMTPS)
- Offline queue on cloud relay: age encryption

**Additional security hardening:**
- All containers run with dropped Linux capabilities and read-only filesystems
- Logs never contain full email addresses or credentials at default verbosity
- SMTP relay enforces configurable message size limits (50MB default)

**What's NOT encrypted (by default):**
- Mail storage on home device (plaintext on disk unless you enable filesystem encryption)

**For end-to-end encryption (E2EE) like PGP or S/MIME:**
- DarkPipe does not prevent you from using PGP/S/MIME
- Configure your mail client (Thunderbird, K-9 Mail, etc.) for PGP
- DarkPipe is about storage sovereignty, not content encryption

**Recommendation:** Enable full-disk encryption (LUKS, FileVault) on your home device for at-rest encryption.

### What happens if my VPS is compromised?

**An attacker could read mail in transit, but cannot access stored mail.**

**What an attacker CAN do if cloud relay is compromised:**
- Read mail passing through the relay (before encryption or after decryption)
- Decrypt offline queue (age private key is on cloud relay)
- Redirect mail to different destinations

**What an attacker CANNOT do:**
- Access stored mail on your home device (none is stored on cloud relay)
- Read historical mail (only current mail in transit)
- Decrypt mail on your home device (keys are on home device)

**Mitigations:**
- Harden your VPS (firewall, automatic updates, SSH key-only)
- Monitor cloud relay logs for suspicious activity
- Use end-to-end encryption (PGP/S/MIME) if you're a high-risk target
- Choose a reputable VPS provider

### What happens if my home device is offline?

**Cloud relay queues mail encrypted until your home device reconnects.**

**Offline queue behavior (configurable):**
- Default: Queue mail encrypted on cloud relay disk (or S3 overflow if enabled)
- Alternative: Bounce mail immediately if home device is unreachable

**How it works:**
1. Cloud relay detects home device is offline
2. Mail encrypted with age (filippo.io/age) using recipient key
3. Encrypted mail stored on disk or S3-compatible storage
4. When home device reconnects, queue drains automatically
5. Encrypted mail deleted after successful delivery

**Queue timeout:** Default 7 days (configurable). After timeout, mail is bounced.

**S3 overflow:** If local queue exceeds threshold (default 100MB), overflow to S3-compatible storage (Storj, AWS S3, MinIO).

**Recommendation:** Configure S3 overflow for reliability if your home device has frequent outages.

## Operations

### How do I add users and domains?

Depends on your mail server:

**Stalwart:**
- Access web UI: http://HOME_DEVICE_IP:8080
- Log in with admin credentials
- Go to "Users" or "Domains" section
- Add via web interface

**Maddy:**
- Edit `home-device/maddy/setup-users.sh`
- Add line: `maddyctl creds create newuser@example.com`
- Restart Maddy container: `docker compose restart maddy`

**Postfix + Dovecot:**
- Edit `home-device/postfix-dovecot/postfix/vmailbox` (add user)
- Edit `home-device/postfix-dovecot/dovecot/users` (add password)
- Restart container: `docker compose restart postfix-dovecot`

**Aliases:**
- Stalwart: Web UI
- Maddy: Edit config file
- Postfix+Dovecot: Edit `home-device/postfix-dovecot/postfix/virtual`

### How do I back up my mail?

**Back up the mail data directory on your home device.** Cloud relay has no mail to back up.

**Backup locations:**
- Stalwart: `/var/lib/docker/volumes/mail-data/` (or wherever Docker volumes are stored)
- Maddy: `/var/lib/docker/volumes/mail-data/`
- Postfix+Dovecot: `/var/lib/docker/volumes/mail-data/`

**Backup methods:**
1. **Docker volume backup:**
   ```bash
   docker run --rm -v mail-data:/data -v $(pwd):/backup alpine tar czf /backup/mail-backup.tar.gz /data
   ```

2. **Filesystem snapshot (if using ZFS/Btrfs):**
   ```bash
   zfs snapshot tank/mail@backup-2026-02-15
   ```

3. **rsync to external drive:**
   ```bash
   rsync -avz /var/lib/docker/volumes/mail-data/ /mnt/external/darkpipe-backup/
   ```

**Backup frequency:** Daily or weekly, depending on mail volume.

**Restore:** Reverse the backup process (extract tar.gz, rsync back, or restore snapshot).

### How do I update DarkPipe?

**Update Docker images when new versions are released.**

**Process:**
1. Pull new images:
   ```bash
   docker pull ghcr.io/trek-e/darkpipe/cloud-relay:vX.Y.Z
   docker pull ghcr.io/trek-e/darkpipe/home-stalwart:vX.Y.Z-default
   ```

2. Update docker-compose.yml to reference new version tags

3. Restart services:
   ```bash
   docker compose up -d
   ```

4. Verify services are healthy:
   ```bash
   docker compose ps
   ```

**Configuration migrations:** Setup wizard handles config migrations if needed. Re-run wizard with `--migrate-config` flag if prompted in release notes.

**Downtime:** Typically < 30 seconds for image pull and container restart.

### How do I migrate from another email provider?

See [docs/migration.md](migration.md) for detailed guide.

**Quick summary:**
1. Download darkpipe-setup tool
2. Run `darkpipe-setup migrate` subcommand
3. Select source provider (Gmail, Outlook, iCloud, etc.)
4. Authenticate (OAuth2 for Gmail/Outlook, credentials for others)
5. Run dry-run to preview migration
6. Run with `--apply` to execute migration

**Supported providers:** Gmail, Outlook, iCloud, MailCow, Mailu, docker-mailserver, generic IMAP

**Migration time:** 30 minutes to 24+ hours depending on mailbox size.

## Troubleshooting

### Mail isn't being delivered (inbound or outbound)

**Check DNS records:**
```bash
dns-setup --domain example.com --validate-only
```

Ensure MX, SPF, DKIM, DMARC records are present and correct.

**Check port 25 on cloud relay:**
```bash
telnet gmail-smtp-in.l.google.com 25
```

If this times out, port 25 is blocked by your VPS provider.

**Check cloud relay logs:**
```bash
cd cloud-relay
docker compose logs -f relay
```

Look for connection errors, authentication failures, or relay issues.

**Check transport tunnel:**
```bash
# On cloud relay
ping 10.8.0.2  # WireGuard
# OR
nc -zv mail.example.com 25  # mTLS
```

If ping/nc fails, transport tunnel is down.

**Check IP reputation:**
- Go to [MXToolbox Blacklist Check](https://mxtoolbox.com/blacklists.aspx)
- Enter your cloud relay IP
- If blocklisted, follow delisting procedures (usually automatic after 24-48 hours of good behavior)

### Webmail isn't loading

**Check Caddy reverse proxy:**
```bash
docker compose logs -f caddy
```

Look for TLS certificate errors or proxy errors.

**Check webmail container:**
```bash
docker compose ps
```

Ensure webmail container (roundcube or snappymail) is "Up" and healthy.

**Check port access:**
- Try http://HOME_DEVICE_IP:8080 (direct to webmail, bypassing Caddy)
- If this works but HTTPS doesn't, Caddy TLS issue
- If this doesn't work, webmail container issue

**Check DNS:**
- Ensure mail.example.com points to your home device IP (or cloud relay if using reverse proxy)

### Transport tunnel is down (WireGuard or mTLS)

**For WireGuard:**
```bash
# Check WireGuard status
sudo wg show

# Restart WireGuard
sudo wg-quick down wg0
sudo wg-quick up wg0

# Check firewall
sudo iptables -L | grep 51820  # UDP port 51820 must be allowed
```

**For mTLS:**
```bash
# Check certificate validity
openssl x509 -in /certs/relay.crt -noout -dates

# Check TCP connectivity
nc -zv HOME_DEVICE_IP 25

# Check TLS handshake
openssl s_client -connect HOME_DEVICE_IP:25 -cert /certs/relay.crt -key /certs/relay.key
```

**Check monitoring dashboard:**
- Navigate to http://HOME_DEVICE_IP:8090/monitoring
- Check "Transport Status" section for connectivity alerts

### Services using too much memory

**Check resource usage:**
```bash
docker stats
```

**Common memory hogs:**
- Roundcube (256MB limit) - switch to SnappyMail (128MB)
- Stalwart (512MB limit) - switch to Maddy (256MB) or Postfix+Dovecot
- Rspamd (256MB limit) - reduce Redis memory or disable some modules

**Optimization for 2GB RAM systems:**
- Use Maddy instead of Stalwart
- Use SnappyMail instead of Roundcube
- Reduce Docker memory limits in docker-compose.yml
- See platform guides for memory-constrained setups

### Need help?

**Community support:**
- [GitHub Discussions](https://github.com/trek-e/darkpipe/discussions) - Ask questions, share experiences
- [GitHub Issues](https://github.com/trek-e/darkpipe/issues) - Report bugs

**Documentation:**
- [Quick Start Guide](quickstart.md)
- [Configuration Reference](configuration.md)
- [Architecture Overview](architecture.md)
- [Security Model](security.md)

**Before asking for help:**
- Search existing Discussions and Issues
- Check the FAQ (you're reading it!)
- Include relevant logs, environment details, and error messages in your post

---

Last Updated: 2026-03-12

License: AGPLv3 - See [LICENSE](../LICENSE)
