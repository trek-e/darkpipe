# VPS Provider Selection for DarkPipe

## Overview

DarkPipe requires a VPS with unrestricted **port 25 access** to send and receive SMTP traffic. Many cloud providers block port 25 by default to prevent spam, making proper provider selection critical for deployment.

This guide documents known compatible and restricted providers, validation procedures, and minimum VPS specifications.

## Port 25 Requirement

### Why Port 25 is Critical

Port 25 is the standard SMTP port used for server-to-server email delivery:

- **Inbound**: Other mail servers connect to your relay on port 25 to deliver mail
- **Outbound**: Your relay connects to other mail servers on port 25 to send mail
- **No alternatives**: While port 587 (submission) and 465 (SMTPS) exist, they're for client-to-server only -- server-to-server email MUST use port 25 per RFC 5321

Without port 25 access, your DarkPipe relay cannot receive mail from the internet or deliver mail to external recipients.

### How Providers Block Port 25

Providers typically block port 25 in one of several ways:

1. **Firewall block**: Traffic to/from port 25 is silently dropped at the network edge
2. **Rate limiting**: Port 25 traffic is throttled to near-zero (AWS "throttle removal")
3. **Policy restriction**: Account-level restriction that requires support ticket to lift
4. **Permanent block**: No option to unblock (e.g., Google Cloud Compute)

## Known Compatible Providers

These providers allow port 25 access, though some require account verification or support requests for new accounts.

| Provider | Port 25 | Notes | Tested |
|----------|---------|-------|--------|
| **Vultr** | Yes (after account verification) | New accounts may be restricted initially. Contact support with legitimate use case. Typically approved within 24 hours. | No |
| **Hetzner** | Yes | Available on all plans. Strong EU privacy laws (GDPR). Excellent price/performance. Popular for self-hosted mail. | No |
| **OVH/OVHcloud** | Yes | Must manually configure reverse DNS (PTR record) via control panel. Available in EU, US, Canada, Asia. | No |
| **Linode/Akamai** | Yes (after account age) | New accounts may have restrictions. Typically opens after 7-14 days or via support ticket. Good documentation for mail server setup. | No |
| **BuyVM** | Yes | Explicitly SMTP-friendly. Known for allowing services other hosts block. Smaller provider but reliable. Limited regions (US, Luxembourg). | No |
| **Ramnode** | Yes | Port 25 open by default. Multiple regions. Good reputation for privacy-focused hosting. | No |
| **Scaleway** | Yes | EU-based (France). Port 25 open on Dedibox and some Instance types. Check product specs before purchase. | No |
| **Contabo** | Yes | Very low cost. EU and US regions. Port 25 open but IP reputation may vary (shared IP ranges). | No |

### Selection Criteria Beyond Port 25

When choosing a compatible provider, also consider:

- **IP Reputation**: Shared hosting IPs may be on spam blocklists. Check with [MXToolbox](https://mxtoolbox.com/blacklists.aspx) before purchase if possible.
- **Reverse DNS**: Provider must allow you to set PTR records for your IP address (required for deliverability).
- **Static IPv4**: You need a dedicated, static IPv4 address (not dynamic, not shared).
- **Network Quality**: Low-latency, reliable connectivity matters for real-time email.
- **Data Center Location**: Choose based on your primary communication partners (US vs EU vs Asia).

## Known Restricted Providers

These providers block port 25 by default. Some allow unblocking via support ticket, but approval is not guaranteed.

| Provider | Port 25 | Notes |
|----------|---------|-------|
| **DigitalOcean** | Blocked by default | Must request via support ticket. Often denied for new accounts. Approval rates vary by account age and history. Not recommended for new deployments. |
| **AWS EC2** | Throttled by default | Must submit "Request to Remove Email Sending Limitations" via AWS Support. Can take 24-48 hours. Throttle (not full block) means some mail gets through slowly. |
| **Google Cloud (GCP)** | **Permanently blocked** | Cannot send on port 25, period. No exceptions. Google explicitly prohibits SMTP servers on Compute Engine. |
| **Microsoft Azure** | Blocked | Enterprise/MSDN accounts may request unblock. Pay-as-you-go accounts typically cannot. Not reliable for self-hosted mail. |
| **Oracle Cloud** | Blocked on free tier | Paid accounts can request unblock. Free tier (Always Free) does not allow port 25. |
| **Hostinger VPS** | Blocked | No option to unblock. Shared hosting only supports their mail service. |

### Why These Providers Block Port 25

Cloud providers block port 25 primarily to:

1. **Prevent spam**: Compromised instances are often used to send spam
2. **Protect IP reputation**: Spam from their networks damages deliverability for all customers
3. **Regulatory compliance**: Some jurisdictions have strict anti-spam laws

This is reasonable for general-purpose hosting, but problematic for legitimate mail servers.

## How to Validate Your Provider

Before committing to a VPS for production use, validate that port 25 actually works.

### Validation Procedure

1. **Provision smallest VPS** (typically $3-6/month)
2. **Test inbound port 25**:
   ```bash
   # On your VPS
   nc -l 25

   # From external machine (not on the same provider network)
   telnet YOUR_VPS_IP 25
   ```

   Expected: Connection succeeds, you can type and see it echo on the VPS.

3. **Test outbound port 25**:
   ```bash
   # From your VPS
   telnet gmail-smtp-in.l.google.com 25
   ```

   Expected: Connection succeeds, you see an SMTP greeting like `220 mx.google.com ESMTP`.

4. **Check reverse DNS (PTR record)**:
   ```bash
   # From anywhere
   dig -x YOUR_VPS_IP +short
   ```

   Expected: Returns a hostname. If empty, you need to configure it in your provider's control panel.

   Most providers require you to:
   - Set a hostname for your VPS (e.g., `relay.yourdomain.com`)
   - Configure PTR record to point from your IP back to that hostname

   Without proper reverse DNS, many mail servers will reject your mail as potential spam.

5. **Check IP reputation**:
   ```bash
   # Visit these services with your VPS IP
   # https://mxtoolbox.com/blacklists.aspx
   # https://www.spamhaus.org/lookup/
   ```

   If your IP is on blocklists, you can request delisting, but a provider with many blocklisted IPs is a red flag.

### Automated Validation Script

You can use the `deploy/test/vps-validate.sh` script (if available) to automate these checks:

```bash
# On your VPS
curl -sSL https://raw.githubusercontent.com/yourusername/darkpipe/master/deploy/test/vps-validate.sh | bash
```

(Note: This script will be added in a future phase.)

## Minimum VPS Specifications

DarkPipe's relay component is lightweight. Minimum recommended specs:

| Resource | Minimum | Recommended | Notes |
|----------|---------|-------------|-------|
| **CPU** | 1 vCPU | 2 vCPU | Email is I/O-bound, not CPU-bound |
| **RAM** | 512 MB | 1 GB | Postfix + WireGuard/mTLS + monitoring |
| **Storage** | 10 GB SSD | 20 GB SSD | Mail queue, logs, system |
| **Network** | 1 TB/month | Unlimited | Email traffic is minimal |
| **IPv4** | Static, dedicated | Static, dedicated | **Required** -- no dynamic IPs |
| **IPv6** | Optional | Recommended | Some providers deliver via IPv6 |
| **Reverse DNS** | **Required** | **Required** | Must be able to set PTR record |

### Storage Considerations

- **Mail queue**: During outages or delivery delays, mail queues on disk. Budget 1-5 GB for queue.
- **Logs**: Postfix, system logs, monitoring logs. Rotate frequently.
- **Backups**: You should backup your config and PKI certificates (but not mail -- mail is on your home device).

### Bandwidth Considerations

Email is extremely low bandwidth:

- Average email: 50-100 KB
- 1000 emails/day: ~50-100 MB/day
- 30,000 emails/month: ~1.5-3 GB/month

Even the smallest VPS bandwidth allocations (500 GB - 1 TB/month) are more than sufficient.

## Recommended Starting Configuration

For most users, we recommend:

**Provider**: Hetzner or Vultr
**Plan**: Smallest VPS (1 vCPU, 1 GB RAM, 20 GB SSD)
**Cost**: ~$3-6/month
**Location**: Closest to your primary email contacts (US East for US-based, EU for European)

### Example: Hetzner CX11

- **Cost**: €3.79/month (~$4.50/month)
- **Specs**: 1 vCPU (AMD), 2 GB RAM, 20 GB SSD, 20 TB traffic
- **Location**: Germany, Finland, US (Virginia, Oregon)
- **Port 25**: Open by default
- **Reverse DNS**: Configurable via web console
- **IPv6**: Included

This is overkill for DarkPipe's relay component but provides headroom for monitoring, logging, and future features.

## Next Steps

After selecting a provider and provisioning a VPS:

1. **Set hostname**: `hostnamectl set-hostname relay.yourdomain.com`
2. **Configure reverse DNS**: Set PTR record in provider's control panel
3. **Verify port 25**: Run validation procedure above
4. **Install DarkPipe relay**: Follow `docs/deployment/relay-setup.md`
5. **Configure monitoring**: See Phase 9 for health monitoring setup

## Frequently Asked Questions

### Can I use a VPS from a restricted provider if I get port 25 unblocked?

Yes, if you can get written confirmation from support that port 25 is unblocked for your account. However:

- Approval is not guaranteed
- May be revoked if abuse is detected on their network
- Often slower support response for mail-related issues

**Recommendation**: Start with a provider that allows port 25 by default. Less friction, less risk.

### Can I use a home internet connection instead of a VPS?

Most residential ISPs block port 25 outbound (and sometimes inbound) to prevent spam from infected home PCs. Additionally:

- Residential IPs are often on spam blocklists
- Most ISPs don't allow reverse DNS configuration
- Dynamic IPs break email (SPF, DKIM, reputation)

**Recommendation**: Always use a VPS for the relay component. DarkPipe is designed for hybrid cloud + home architecture.

### What if my chosen provider's IP is on a blocklist?

1. **Request delisting**: Most blocklists allow IP owners to request removal
2. **Check reason**: Some blocklists only list IPs with active spam (temporary), others list entire ranges (problematic)
3. **Warm up IP**: Send low volume legitimate mail for 2-4 weeks before increasing volume
4. **Switch providers**: If the range is permanently blocklisted, choose a different provider

IP reputation is critical for deliverability. An IP with bad history can take months to rehabilitate.

### Do I need multiple VPSs for redundancy?

Not initially. Start with a single VPS. You can add a backup MX later:

- Primary MX (your DarkPipe relay)
- Secondary MX (backup relay on different VPS/provider)

Phase 7 (High Availability) covers multi-relay setups.

## Testing Status

All providers listed as "Tested: No" have not been validated by the DarkPipe project. The information is based on public documentation and community reports.

If you successfully deploy on a listed provider, please contribute your results:

```bash
# After successful deployment
./deploy/test/tunnel-test.sh wireguard
./deploy/test/vps-report.sh > my-provider-report.txt
# Submit report via GitHub issue or PR
```

This helps future users make informed decisions.

## Additional Resources

- [RFC 5321: Simple Mail Transfer Protocol](https://tools.ietf.org/html/rfc5321)
- [MXToolbox Blacklist Check](https://mxtoolbox.com/blacklists.aspx)
- [Spamhaus IP Reputation](https://www.spamhaus.org/lookup/)
- [How to Check Reverse DNS](https://mxtoolbox.com/ReverseLookup.aspx)
- [Email Deliverability Best Practices](https://www.validity.com/everest/resource/email-deliverability-guide/)

---

**Last Updated**: 2026-02-09
**Maintainer**: DarkPipe Project
**License**: AGPLv3
