# DarkPipe on TrueNAS Scale

Deploy DarkPipe as a Custom App on TrueNAS Scale 24.10+ (Electric Eel).

## Prerequisites

**Hardware:**
- TrueNAS Scale installation (bare metal or VM)
- 4GB+ RAM allocated to DarkPipe app
- Pool with dataset for mail storage

**Software:**
- TrueNAS Scale 24.10+ (Electric Eel) - **Required for Docker Compose support**
- Pool created with at least 20GB free space

**Why 24.10+?**
TrueNAS Scale 24.10 (Electric Eel) introduced native Docker Compose support via Custom Apps. Earlier versions (23.10 Cobia, 24.04 Dragonfish) use a Kubernetes backend which requires different templates and workflows. If you're on <24.10, **upgrade to 24.10+** for the best DarkPipe experience.

## Quick Start (UI Method)

1. **Apps** → **Discover Apps** → **Custom App**
2. **Upload** `deploy/templates/truenas-scale/app.yaml`
3. **Fill form** with your mail domain, relay hostname, admin password
4. **Install**

## Detailed Steps

### Method 1: Custom App via TrueNAS UI (Recommended)

**Step 1: Create Dataset**

Navigate to **Storage** → **Create Dataset**:
- **Name**: `apps/darkpipe`
- **Dataset Preset**: Generic
- **Share Type**: Generic
- **Path**: `/mnt/pool/apps/darkpipe` (adjust pool name)

Click **Save**.

**Step 2: Install Custom App**

Navigate to **Apps** → **Discover Apps** → **Custom App**:

1. **Application Name**: `darkpipe`
2. **Upload Configuration**: Click **Upload**, select `deploy/templates/truenas-scale/app.yaml`

The form will populate with questions from `questions.yaml`:

**Basic Configuration:**
- **Mail Domain**: Your domain (e.g., `example.com`)
- **Cloud Relay Hostname**: Your VPS hostname (e.g., `relay.example.com`)
- **Mail Server Hostname**: Your home server hostname (e.g., `mail.example.com`)
- **Admin Email**: Your admin email (e.g., `admin@example.com`)
- **Admin Password**: Create a strong password (≥16 characters)

**Component Selection:**
- **Mail Server**: Stalwart (recommended), Maddy, or Postfix+Dovecot
- **Webmail Client**: SnappyMail (recommended), Roundcube, or None
- **Calendar/Contacts**: Built-in (Stalwart only), Radicale, or None

**Transport Layer:**
- **Transport Protocol**: WireGuard (recommended)

**Storage Configuration:**
- **Storage Path**: `/mnt/pool/apps/darkpipe` (from Step 1)

**Advanced Options:**
- **Enable Message Queue**: `true` (recommended)
- **Maximum Message Size**: 52428800 (50MB)
- **Enable Webhook Notifications**: `false` (or configure if desired)

3. Click **Install**

TrueNAS will:
- Pull Docker images from GHCR
- Create volumes in your dataset
- Start all services
- Map ports (25, 587, 993, 8080)

**Step 3: Verify Deployment**

Navigate to **Apps** → **Installed**:
- **DarkPipe** should show **Active** status
- Click **Shell** to access container console
- Check logs: Click **Logs** → Select container

### Method 2: Docker Compose via SSH (Advanced)

For users who prefer command-line control:

**Step 1: Enable SSH**

Navigate to **System Settings** → **Services** → **SSH**:
- Enable SSH service
- Note the SSH port (default: 22)

**Step 2: SSH into TrueNAS**

```bash
ssh admin@truenas.local
```

**Step 3: Create Storage Directory**

```bash
# Create directory on pool
sudo mkdir -p /mnt/pool/apps/darkpipe
cd /mnt/pool/apps/darkpipe
```

**Step 4: Download Setup Tool**

```bash
# For x64 systems (most TrueNAS installations)
curl -LO https://github.com/trek-e/darkpipe/releases/latest/download/darkpipe-setup-linux-amd64
chmod +x darkpipe-setup-linux-amd64

# Run setup
sudo ./darkpipe-setup-linux-amd64 setup
```

**Step 5: Start Services**

```bash
sudo docker compose up -d
```

## Platform-Specific Notes

### Host Networking vs Bridge Networking

**Bridge networking (default)** works for most users:
- Services use container ports
- TrueNAS maps ports to host
- Easier to isolate services

**Host networking** may be required if:
- SMTP port 25 has issues with bridge networking
- You need direct access to network interfaces

To use host networking, edit `app.yaml`:
```yaml
services:
  mail-server:
    network_mode: host
```

**Trade-off**: Host networking exposes all container ports directly to the network. Less isolated but more compatible.

### Docker Secrets vs Environment Variables

**TrueNAS Scale 24.10+** supports Docker secrets natively via the Custom App UI.

**TrueNAS Scale <24.10** may not support Docker secrets in the UI. Workaround:
- Use environment variables with password masking
- Or: Use Docker Compose via SSH (Method 2) which supports secrets

In `questions.yaml`, secrets are configured as:
```yaml
- variable: admin_password
  type: password  # UI masks input
  required: true
```

This stores the value as an environment variable, not a Docker secret. For maximum security on 24.10+, use Docker Compose method with secrets files.

### Memory and CPU Limits

TrueNAS Custom Apps respect Docker Compose `deploy.resources.limits`:

```yaml
deploy:
  resources:
    limits:
      memory: 512M
      cpus: '1.0'
```

If your TrueNAS system has limited resources, adjust limits in `app.yaml` before installation.

### Storage Persistence

All DarkPipe volumes are created in your dataset path (`/mnt/pool/apps/darkpipe`):
- `postfix-queue/`: Mail queue
- `certbot-etc/`: Let's Encrypt certificates
- `mail-data/`: Mailboxes and indexes
- `rspamd-data/`: Spam filter database
- `redis-data/`: Greylisting and stats

**Backups**: Use TrueNAS Tasks → Replication Tasks to replicate the dataset to another pool or system.

### TrueNAS <24.10 (Kubernetes-Based Apps)

**Not directly supported.** DarkPipe uses Docker Compose which is only available in 24.10+.

**Options:**
1. **Upgrade to 24.10+** (recommended)
2. **Use Jailbait plugin** to run Docker inside a VM
3. **Deploy on a separate Linux VM** on the same TrueNAS host

If upgrading is not possible, see the [Proxmox LXC guide](proxmox-lxc.md) for running DarkPipe in a container on the same physical hardware.

## Troubleshooting

### Custom App install fails with "image not found"

**Cause**: GHCR rate limiting or authentication issue.

**Fix**:
1. Authenticate to GHCR via SSH:
   ```bash
   echo $GITHUB_TOKEN | docker login ghcr.io -u your-github-username --password-stdin
   ```
2. Retry installation

### Port 25 already in use

**Cause**: TrueNAS may have a mail server service running (if configured).

**Fix**:
1. Navigate to **System Settings** → **Services**
2. Disable any mail-related services (Postfix, Sendmail)
3. Retry installation

### Services show "Unhealthy" status

**Cause**: Health checks failing (common on first boot while services initialize).

**Fix**:
1. Wait 2-3 minutes for services to fully start
2. Check logs: Apps → DarkPipe → Logs
3. Verify DNS is configured correctly
4. If persists, check resource limits (may need more RAM)

### Can't access webmail on port 8080

**Cause**: Firewall or port conflict.

**Fix**:
1. Check TrueNAS firewall: System Settings → Network → Firewall
2. Verify port 8080 is not used by TrueNAS UI (default: 80, 443)
3. Try alternate port in app.yaml:
   ```yaml
   ports:
     - "8081:8080"  # Use 8081 on host
   ```

## Performance Benchmarks

**TrueNAS Scale on typical hardware (4-core CPU, 16GB RAM, SSD pool):**
- Mail delivery: <50ms
- IMAP FETCH: 5-20ms per message
- Webmail page load: 100-300ms
- RAM usage (Stalwart + SnappyMail): ~1.2GB
- Idle CPU: <2%

## Security Considerations

1. **Network isolation**: Use TrueNAS VLANs to isolate mail server
2. **Backups**: Enable ZFS snapshots on dataset (TrueNAS automated snapshots)
3. **Updates**: Update Custom App when new DarkPipe versions are released
4. **Firewall**: Configure TrueNAS firewall to only allow required ports

## Next Steps

1. **Configure DNS**: Set up DKIM, SPF, DMARC (see Phase 4 guide)
2. **Test mail flow**: Send test email from Gmail to your@domain.com
3. **Set up replication**: Replicate dataset to backup pool or remote TrueNAS
4. **Monitor logs**: Apps → DarkPipe → Logs

## Alternative Runtimes

TrueNAS Scale 24.10+ uses a native Docker engine for its Custom Apps system. Podman is **not applicable** on TrueNAS Scale — the platform does not support alternative container runtimes. Use the built-in Docker engine via Custom Apps or SSH as described in this guide.

For Podman deployment on supported platforms, see the [Podman Platform Guide](podman.md).

## See Also

- [Podman Platform Guide](podman.md) - Podman deployment (for supported platforms)
- [Raspberry Pi Guide](raspberry-pi.md) - Alternative home server platform
- [Unraid Guide](unraid.md) - Similar NAS platform
- [DarkPipe Setup Tool](../setup/) - Interactive configuration wizard
- [TrueNAS Custom Apps Docs](https://www.truenas.com/docs/truenasapps/usingcustomapp/)
