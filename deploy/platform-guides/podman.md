# DarkPipe on Podman

Run DarkPipe with Podman instead of Docker. This guide covers both the cloud relay (rootful) and home device (rootless option) deployments.

> **Validate your environment first:** Run `bash scripts/check-runtime.sh` to verify Podman version, podman-compose, cgroup v2, and SELinux state before deploying.

## Prerequisites

| Requirement | Minimum Version | Why |
|-------------|----------------|-----|
| Podman | 5.3.0+ | `extra_hosts: host-gateway` support, pasta networking |
| podman-compose | 1.x+ | Profiles, `depends_on` conditions, `deploy.resources` |
| cgroup v2 | — | Required for memory limits (`deploy.resources.limits.memory`) |

**Verify prerequisites:**

```bash
podman --version          # Must be 5.3.0+
podman-compose --version  # Must be 1.x+
mount | grep cgroup2      # Must show cgroup2 mount

# Or use the automated check:
bash scripts/check-runtime.sh
```

## Quick Start

**Cloud relay (rootful — required):**

```bash
cd cloud-relay
sudo podman-compose -f docker-compose.yml -f docker-compose.podman.yml up -d
```

**Home device (rootless — optional sysctl required):**

```bash
# Allow binding privileged ports without root
sudo sysctl net.ipv4.ip_unprivileged_port_start=0

cd home-device
podman-compose -f docker-compose.yml -f docker-compose.podman.yml \
  --profile stalwart --profile roundcube up -d
```

## Key Differences from Docker

### Override files

Every Podman deployment **must** layer the `docker-compose.podman.yml` override file. This file enables Docker Compose compatibility flags (`x-podman` extensions) so podman-compose correctly interprets `depends_on` conditions, health checks, and network naming.

Always pass it with `-f`:

```bash
podman-compose -f docker-compose.yml -f docker-compose.podman.yml [options] up -d
```

### `host-gateway` resolution

Docker resolves `extra_hosts: host-gateway` to the host IP automatically. Podman 5.3.0+ supports this via pasta networking. Additionally, Podman provides `host.containers.internal` as a built-in hostname inside every container — no `extra_hosts` entry needed.

### Networking

Podman rootful uses netavark (similar to Docker's bridge driver). Container-to-container DNS resolution works the same way as Docker when `--in-pod` is **not** used (the default).

### Security defaults

Rootless Podman runs with reduced capabilities by default. Rootful Podman's security model is similar to Docker. The compose file's `cap_drop: ALL` + explicit `cap_add` is honored correctly in both modes.

### Memory limits

`deploy.resources.limits.memory` requires cgroup v2 delegation, which is the default on modern Fedora/RHEL and Ubuntu 22.04+. If memory limits appear to have no effect, verify cgroup v2:

```bash
mount | grep cgroup2
```

## Cloud Relay Deployment

The cloud relay **must run under rootful Podman** for two reasons:

1. **Port 25 binding** — SMTP requires binding to a privileged port.
2. **`/dev/net/tun` device access** — The WireGuard container needs the TUN device for tunnel creation. Rootless Podman cannot pass through `/dev/net/tun` without `--privileged`.

### Start the cloud relay

```bash
cd cloud-relay
sudo podman-compose -f docker-compose.yml -f docker-compose.podman.yml up -d
```

### Firewall configuration (Fedora/RHEL)

Podman rootful does **not** auto-configure firewalld the way Docker does. Open ports manually:

```bash
sudo firewall-cmd --add-port=25/tcp --permanent
sudo firewall-cmd --add-port=443/tcp --permanent
sudo firewall-cmd --add-port=51820/udp --permanent   # WireGuard
sudo firewall-cmd --reload
```

## Home Device Deployment

The home device **can run rootless** — it does not require `/dev/net/tun` access. However, mail servers bind privileged ports (25, 587, 993), so you must lower the unprivileged port threshold:

```bash
# Allow rootless containers to bind ports starting from 0
sudo sysctl net.ipv4.ip_unprivileged_port_start=0

# Make persistent across reboots
echo 'net.ipv4.ip_unprivileged_port_start=0' | sudo tee /etc/sysctl.d/99-unprivileged-ports.conf
```

Alternatively, run rootful with `sudo` — no sysctl change needed.

### Profile commands

The home device uses Docker Compose profiles to select components. Always include the Podman override file.

**Stalwart + Roundcube (built-in CalDAV/CardDAV):**

```bash
podman-compose -f docker-compose.yml -f docker-compose.podman.yml \
  --profile stalwart --profile roundcube up -d
```

**Maddy + SnappyMail + Radicale:**

```bash
podman-compose -f docker-compose.yml -f docker-compose.podman.yml \
  --profile maddy --profile snappymail --profile radicale up -d
```

**Postfix+Dovecot + Roundcube + Radicale:**

```bash
podman-compose -f docker-compose.yml -f docker-compose.podman.yml \
  --profile postfix-dovecot --profile roundcube --profile radicale up -d
```

### Firewall configuration (Fedora/RHEL)

```bash
sudo firewall-cmd --add-port=25/tcp --permanent
sudo firewall-cmd --add-port=587/tcp --permanent
sudo firewall-cmd --add-port=993/tcp --permanent
sudo firewall-cmd --add-port=443/tcp --permanent    # Webmail
sudo firewall-cmd --add-port=5232/tcp --permanent   # Radicale (if used)
sudo firewall-cmd --reload
```

## SELinux

If your system runs SELinux in enforcing mode (Fedora, RHEL, CentOS), bind-mounted volumes need relabeling so containers can read host config files.

### Detect SELinux status

```bash
getenforce
# "Enforcing" = you need the SELinux override
# "Permissive" or "Disabled" = skip it
```

### Apply the SELinux override

Layer `docker-compose.podman-selinux.yml` on top of the base Podman override. This file adds `:z` (shared label) to bind-mount volumes.

**Cloud relay with SELinux:**

```bash
sudo podman-compose \
  -f docker-compose.yml \
  -f docker-compose.podman.yml \
  -f docker-compose.podman-selinux.yml \
  up -d
```

**Home device with SELinux:**

```bash
podman-compose \
  -f docker-compose.yml \
  -f docker-compose.podman.yml \
  -f docker-compose.podman-selinux.yml \
  --profile stalwart --profile roundcube up -d
```

> **Note:** `:z` (shared label) is safe on non-SELinux systems — it's a no-op. Named volumes are managed by Podman and don't need SELinux labels.

## Troubleshooting

### Pod mode breaks DNS resolution

**Symptom:** Services can't reach each other by name (e.g., rspamd → redis, webmail → mail server).

**Cause:** `podman-compose --in-pod` or `x-podman: in_pod: true` places all containers in a shared network namespace where they communicate via `localhost`, breaking service-name DNS.

**Fix:** Do **not** use `--in-pod`. DarkPipe relies on inter-container DNS resolution via the default netavark network.

### Permission denied on bind mounts (SELinux)

**Symptom:** Containers fail to read config files. Logs show `Permission denied` errors.

**Fix:** Add the SELinux override file (see [SELinux](#selinux) section above).

### Ports not reachable (firewalld)

**Symptom:** External clients can't connect to SMTP/HTTPS/WireGuard ports, even though containers are running.

**Fix:** Podman does not auto-configure firewalld. Open ports manually (see firewall commands above for [cloud relay](#firewall-configuration-fedoralrhel) or [home device](#firewall-configuration-fedoralrhel-1)).

### Memory limits not enforced

**Symptom:** Containers exceed their configured memory limits.

**Fix:** Verify cgroup v2 is active: `mount | grep cgroup2`. If not, enable it by adding `systemd.unified_cgroup_hierarchy=1` to your kernel command line.

### Rootless container can't bind port 25

**Symptom:** `Error: rootlessport listen tcp 0.0.0.0:25: bind: permission denied`

**Fix:** Lower the unprivileged port threshold:

```bash
sudo sysctl net.ipv4.ip_unprivileged_port_start=0
```

Or run rootful with `sudo`.

### WireGuard container fails (rootless)

**Symptom:** WireGuard container exits with TUN device errors.

**Fix:** The cloud relay **must** run rootful. Rootless Podman cannot pass through `/dev/net/tun`. Use `sudo podman-compose ...`.

## Resources

- **Per-component references:** [`cloud-relay/PODMAN.md`](../../cloud-relay/PODMAN.md) · [`home-device/PODMAN.md`](../../home-device/PODMAN.md)
- **Runtime validation:** [`scripts/check-runtime.sh`](../../scripts/check-runtime.sh) — checks Podman version, podman-compose, cgroup v2, SELinux state
- **Podman documentation:** [podman.io](https://podman.io/) · [podman-compose](https://github.com/containers/podman-compose)
- **SELinux guide:** [Podman and SELinux](https://docs.podman.io/en/latest/markdown/podman-run.1.html#security-label-label-disable-label-filetype-type-level-level)
