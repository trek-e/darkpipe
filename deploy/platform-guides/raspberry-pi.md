# DarkPipe on Raspberry Pi 4

Deploy DarkPipe as your home mail server on Raspberry Pi 4 (arm64).

## Prerequisites

**Hardware:**
- Raspberry Pi 4 Model B: **4GB RAM recommended** (2GB possible with memory optimization)
- USB3 SSD or NVMe (via USB adapter): **Strongly recommended** over SD card for mail storage
- Power supply: Official Raspberry Pi 4 power supply (5V/3A USB-C)
- Network: Ethernet connection recommended over Wi-Fi

**Software:**
- Raspberry Pi OS 64-bit (Bookworm) or Ubuntu Server 24.04 LTS (arm64)
- Docker 27+ and Docker Compose v2
- SSH access enabled

**Why 4GB?**
Full stack (Stalwart + webmail + Rspamd + Redis + Caddy) can consume 1.5-2GB RAM under load. 2GB Pi 4 is at the absolute minimum and requires memory optimization (see below). 4GB provides comfortable headroom.

**Why SSD?**
Mail servers perform many small random I/O operations (mailbox checks, index updates, searches). SD cards degrade quickly under this workload and are 10-100x slower than SSDs. Use SD card for OS only, USB3 SSD for /var/mail.

## Quick Start (3 commands)

```bash
# 1. Install Docker (if not already installed)
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# 2. Download and run setup tool
curl -LO https://github.com/trek-e/darkpipe/releases/latest/download/darkpipe-setup-linux-arm64
chmod +x darkpipe-setup-linux-arm64
./darkpipe-setup-linux-arm64 setup

# 3. Start services
docker compose up -d
```

## Detailed Steps

### 1. OS Setup

**Option A: Raspberry Pi OS 64-bit (Recommended)**
```bash
# Verify 64-bit OS
uname -m
# Should output: aarch64

# Update system
sudo apt update && sudo apt upgrade -y

# Enable swap on SSD (NOT SD card)
# Assuming SSD is mounted at /mnt/ssd
sudo fallocate -l 2G /mnt/ssd/swapfile
sudo chmod 600 /mnt/ssd/swapfile
sudo mkswap /mnt/ssd/swapfile
sudo swapon /mnt/ssd/swapfile

# Make swap permanent
echo '/mnt/ssd/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

**Option B: Ubuntu Server 24.04 LTS**
```bash
# Ubuntu 64-bit comes pre-configured
# Verify
uname -m
# Should output: aarch64

# Update system
sudo apt update && sudo apt upgrade -y
```

### 2. Install Docker

```bash
# Install Docker via convenience script
curl -fsSL https://get.docker.com | sh

# Add user to docker group (avoid sudo)
sudo usermod -aG docker $USER

# Log out and back in, or run:
newgrp docker

# Verify Docker
docker version
docker compose version
```

### 3. Prepare Storage

```bash
# Mount SSD (example: /dev/sda1)
sudo mkdir -p /mnt/mail-storage
sudo mount /dev/sda1 /mnt/mail-storage

# Make mount permanent
echo '/dev/sda1 /mnt/mail-storage ext4 defaults,noatime 0 2' | sudo tee -a /etc/fstab

# Create DarkPipe directory
mkdir -p /mnt/mail-storage/darkpipe
cd /mnt/mail-storage/darkpipe
```

### 4. Run Interactive Setup

```bash
# Download setup tool for ARM64
curl -LO https://github.com/trek-e/darkpipe/releases/latest/download/darkpipe-setup-linux-arm64
chmod +x darkpipe-setup-linux-arm64

# Run setup wizard
./darkpipe-setup-linux-arm64 setup
```

**Setup will:**
- Detect available RAM and warn if <3GB
- Validate DNS records for your domain
- Test cloud relay SMTP port 25 connectivity
- Generate docker-compose.yml with memory-optimized defaults
- Create Docker secrets with secure permissions

**If 2GB Pi 4 detected:**
Setup will recommend:
- Mail server: Maddy (lighter than Stalwart)
- Webmail: SnappyMail (128MB) or none (use desktop client)
- Calendar: None or Radicale (if needed)

### 5. Start Services

```bash
# Start all services in background
docker compose up -d

# Check logs
docker compose logs -f

# Verify services are healthy
docker compose ps
```

### 6. Configure DNS

Your domain's DNS records must point to your cloud relay:

```
example.com.           MX  10 relay.example.com.
relay.example.com.     A   <your-vps-ip>
```

See Phase 4 DNS setup guide for full DKIM/SPF/DMARC configuration.

## Memory Optimization (for 2GB Pi 4)

If running on a 2GB Pi 4, apply these optimizations:

**Option 1: Use Maddy instead of Stalwart**
```bash
# Edit docker-compose.yml
# Use profile: maddy instead of stalwart
docker compose --profile maddy --profile snappymail up -d
```

Maddy memory footprint: ~100-150MB (vs Stalwart ~200-300MB)

**Option 2: Disable webmail, use desktop client**
```bash
# Skip webmail in docker-compose.yml
# Use Thunderbird, Apple Mail, or other IMAP client
```

**Option 3: Increase swap**
```bash
# Create 4GB swap on SSD
sudo fallocate -l 4G /mnt/ssd/swapfile
sudo mkswap /mnt/ssd/swapfile
sudo swapon /mnt/ssd/swapfile
```

**Option 4: Adjust memory limits in docker-compose.yml**
```yaml
services:
  stalwart:
    deploy:
      resources:
        limits:
          memory: 384M  # Down from 512M
  rspamd:
    deploy:
      resources:
        limits:
          memory: 192M  # Down from 256M
```

## Troubleshooting

### Container keeps restarting (OOM killed)

**Symptoms:**
```bash
docker compose ps
# Shows "Restarting" status

dmesg | grep -i oom
# Shows "Out of memory: Killed process..."
```

**Fix:**
1. Reduce memory limits in docker-compose.yml
2. Switch to Maddy mail server
3. Disable webmail
4. Add more swap on SSD

### SD card wearing out

**Symptoms:**
- Slow mail operations
- Filesystem errors in logs
- Docker errors writing to disk

**Fix:**
1. Migrate to USB3 SSD (see "Prepare Storage" above)
2. Move Docker volumes to SSD: `docker volume create --driver local --opt type=none --opt device=/mnt/ssd/darkpipe --opt o=bind mail-data`

### Thermal throttling

**Symptoms:**
```bash
vcgencmd measure_temp
# Shows >80°C

vcgencmd get_throttled
# Shows throttled=0x50000 or similar
```

**Fix:**
1. Add heatsinks and fan to Pi 4
2. Improve case ventilation
3. Reduce ambient temperature

### Port 25 blocked by ISP

**Symptoms:**
- Cloud relay can't reach home device
- Mail delivery fails

**Fix:**
Port 25 blocking affects cloud relay, not home device. Run cloud relay on a VPS with port 25 open. Home device communicates via WireGuard tunnel (any port).

### Slow IMAP performance

**Symptoms:**
- Mailbox takes >5 seconds to open
- Search is slow

**Fix:**
1. Verify mail-data volume is on SSD (not SD card)
2. Check `docker stats` for I/O wait
3. Consider Maddy (lighter IMAP implementation)

## Performance Benchmarks

**Raspberry Pi 4 4GB with USB3 SSD:**
- Mail delivery: <100ms
- IMAP FETCH: 10-50ms per message
- Webmail page load: 200-500ms
- RAM usage (Stalwart + SnappyMail): ~1.2GB
- Idle CPU: <5%

**Raspberry Pi 4 2GB with SD card:**
- Mail delivery: 200-500ms
- IMAP FETCH: 100-300ms per message
- Webmail page load: 1-3 seconds
- RAM usage (Maddy + no webmail): ~600MB
- Idle CPU: <10%

## Security Considerations

1. **Firewall**: Pi 4 sits behind NAT, only WireGuard tunnel exposed
2. **Updates**: Enable unattended-upgrades for security patches
3. **Backups**: Rsync mail-data volume to external storage weekly
4. **Physical security**: Pi 4 in your home = physical access control

## Next Steps

1. **Configure DNS**: Set up DKIM, SPF, DMARC (see Phase 4 guide)
2. **Test mail flow**: Send test email from Gmail to your@domain.com
3. **Set up backups**: Rsync or Restic to external drive/cloud
4. **Monitor logs**: `docker compose logs -f rspamd` to watch spam filtering

## Using Podman

Podman is a viable alternative to Docker on Raspberry Pi OS and Ubuntu arm64. If you prefer a daemonless, rootless-capable container runtime, you can run DarkPipe with Podman instead of Docker.

### Installing Podman on Raspberry Pi OS

```bash
# Raspberry Pi OS (Bookworm) / Debian 12
sudo apt update && sudo apt install -y podman podman-compose

# Or install podman-compose via pip if the packaged version is too old
pip3 install podman-compose

# Verify
podman --version
podman compose version   # requires podman 4.7+ with compose provider
```

### Rootful vs Rootless

- **Cloud relay** — Run rootful (`sudo podman compose up -d`). The relay needs to bind port 25, which requires root privileges.
- **Home device** — Rootless is possible if you don't need privileged ports. Run as your regular user: `podman compose up -d`.

### Override Files

DarkPipe ships Podman-specific compose overrides that adjust volume mounts (`:Z` SELinux labels) and health-check syntax:

```bash
# Use the Podman override alongside the base compose file
podman compose -f docker-compose.yml -f docker-compose.podman.yml up -d
```

### Key Differences from Docker on Pi

| Area | Docker | Podman |
|------|--------|--------|
| Daemon | dockerd (always running) | Daemonless (on-demand) |
| Rootless | Requires configuration | Built-in |
| Compose | `docker compose` (v2) | `podman compose` (v4.7+) or `podman-compose` |
| SELinux | N/A on Raspberry Pi OS | N/A on Raspberry Pi OS (relevant if using Fedora) |

### Runtime Validation

Run the DarkPipe runtime check script to verify your Podman installation meets all prerequisites:

```bash
bash scripts/check-runtime.sh
```

For full Podman deployment details — including SELinux, networking, and troubleshooting — see the [Podman Platform Guide](podman.md).

## See Also

- [Podman Platform Guide](podman.md) - Full Podman deployment reference
- [TrueNAS Scale Guide](truenas-scale.md) - Alternative home server platform
- [Unraid Guide](unraid.md) - NAS platform with Docker support
- [DarkPipe Setup Tool](../setup/) - Interactive configuration wizard
- [Phase 4 DNS Setup](../../.planning/phases/04-dns-email-auth/) - DKIM/SPF/DMARC
