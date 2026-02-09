#!/usr/bin/env bash
# step-ca-setup.sh -- Install and configure step-ca as a private certificate
# authority for DarkPipe internal transport (WireGuard cert needs and mTLS).
#
# This script:
#   1. Installs the Smallstep step CLI and step-ca if not present
#   2. Initialises a new CA with an ACME provisioner
#   3. Creates a systemd service for step-ca
#   4. Installs the cert-renewer systemd timer/service templates
#   5. Prints the CA fingerprint and URL for client bootstrapping
#
# Usage:  sudo bash deploy/pki/step-ca-setup.sh

set -euo pipefail

# ── Require root ─────────────────────────────────────────────────────────────
if [[ $EUID -ne 0 ]]; then
  echo "ERROR: This script must be run as root (sudo)." >&2
  exit 1
fi

# ── Configuration ────────────────────────────────────────────────────────────
CA_NAME="DarkPipe Internal CA"
CA_DNS="ca.darkpipe.internal"
CA_ADDRESS=":8443"
CA_PROVISIONER="darkpipe-acme"
CERT_DIR="/etc/darkpipe/certs"
STEP_USER="step"

# ── Detect architecture ─────────────────────────────────────────────────────
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  STEP_ARCH="amd64" ;;
  aarch64) STEP_ARCH="arm64" ;;
  arm64)   STEP_ARCH="arm64" ;;
  *)
    echo "ERROR: Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac
echo "Detected architecture: $ARCH -> $STEP_ARCH"

# ── Install step CLI ────────────────────────────────────────────────────────
if command -v step &>/dev/null; then
  echo "step CLI already installed: $(step version 2>/dev/null || echo 'unknown version')"
else
  echo "Installing step CLI..."
  STEP_CLI_VERSION="0.28.6"
  curl -fsSL "https://dl.smallstep.com/gh-release/cli/gh-release-header/v${STEP_CLI_VERSION}/step_linux_${STEP_CLI_VERSION}_${STEP_ARCH}.tar.gz" \
    -o /tmp/step-cli.tar.gz
  tar xzf /tmp/step-cli.tar.gz -C /tmp
  mv "/tmp/step_${STEP_CLI_VERSION}/bin/step" /usr/local/bin/step
  chmod 755 /usr/local/bin/step
  rm -rf /tmp/step-cli.tar.gz "/tmp/step_${STEP_CLI_VERSION}"
  echo "step CLI installed: $(step version)"
fi

# ── Install step-ca ─────────────────────────────────────────────────────────
if command -v step-ca &>/dev/null; then
  echo "step-ca already installed: $(step-ca version 2>/dev/null || echo 'unknown version')"
else
  echo "Installing step-ca..."
  STEP_CA_VERSION="0.29.0"
  curl -fsSL "https://dl.smallstep.com/gh-release/certificates/gh-release-header/v${STEP_CA_VERSION}/step-ca_linux_${STEP_CA_VERSION}_${STEP_ARCH}.tar.gz" \
    -o /tmp/step-ca.tar.gz
  tar xzf /tmp/step-ca.tar.gz -C /tmp
  mv "/tmp/step-ca_${STEP_CA_VERSION}/bin/step-ca" /usr/local/bin/step-ca
  chmod 755 /usr/local/bin/step-ca
  rm -rf /tmp/step-ca.tar.gz "/tmp/step-ca_${STEP_CA_VERSION}"
  echo "step-ca installed: $(step-ca version)"
fi

# ── Create step user ────────────────────────────────────────────────────────
if ! id -u "$STEP_USER" &>/dev/null; then
  useradd --system --home-dir /home/step --create-home --shell /usr/sbin/nologin "$STEP_USER"
  echo "Created system user: $STEP_USER"
fi

# ── Create cert directory ───────────────────────────────────────────────────
mkdir -p "$CERT_DIR"
chown "$STEP_USER":"$STEP_USER" "$CERT_DIR"
chmod 700 "$CERT_DIR"

# ── Initialise CA ───────────────────────────────────────────────────────────
echo "Initialising step-ca..."
sudo -u "$STEP_USER" -H step ca init \
  --name "$CA_NAME" \
  --dns "$CA_DNS" \
  --address "$CA_ADDRESS" \
  --provisioner "$CA_PROVISIONER" \
  --provisioner-type ACME

# ── Create systemd service for step-ca ──────────────────────────────────────
cat > /etc/systemd/system/step-ca.service <<EOF
[Unit]
Description=DarkPipe Internal CA (step-ca)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=$STEP_USER
Group=$STEP_USER
ExecStart=/usr/local/bin/step-ca /home/$STEP_USER/.step/config/ca.json
Restart=on-failure
RestartSec=10
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
EOF

# ── Install cert-renewer templates ──────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RENEWAL_DIR="$(dirname "$SCRIPT_DIR")/transport/pki/renewal"

# Try to locate the templates relative to the script, fall back to repo root
if [[ -f "$RENEWAL_DIR/systemd/cert-renewer@.service" ]]; then
  cp "$RENEWAL_DIR/systemd/cert-renewer@.service" /etc/systemd/system/cert-renewer@.service
  cp "$RENEWAL_DIR/systemd/cert-renewer@.timer"   /etc/systemd/system/cert-renewer@.timer
else
  echo "WARNING: cert-renewer templates not found at $RENEWAL_DIR/systemd/" >&2
  echo "         Install them manually from transport/pki/renewal/systemd/" >&2
fi

# ── Enable and start step-ca ────────────────────────────────────────────────
systemctl daemon-reload
systemctl enable step-ca.service
systemctl start step-ca.service

echo ""
echo "=========================================="
echo " DarkPipe Internal CA is running"
echo "=========================================="
echo ""
echo "  CA URL:         https://${CA_DNS}${CA_ADDRESS}"
echo "  Cert directory: $CERT_DIR"
echo ""

# Print CA fingerprint for client bootstrapping
CA_FINGERPRINT=$(sudo -u "$STEP_USER" -H step certificate fingerprint /home/$STEP_USER/.step/certs/root_ca.crt 2>/dev/null || echo "unknown")
echo "  CA fingerprint: $CA_FINGERPRINT"
echo ""
echo "  Bootstrap clients with:"
echo "    step ca bootstrap --ca-url https://${CA_DNS}${CA_ADDRESS} --fingerprint $CA_FINGERPRINT"
echo ""
echo "  Enable certificate renewal for a service:"
echo "    systemctl enable --now cert-renewer@<service-name>.timer"
echo ""
