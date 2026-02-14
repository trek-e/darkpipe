#!/bin/bash
# Postfix+Dovecot Entrypoint Script
# This script:
# 1. Detects setup configuration and loads environment
# 2. Supports Docker secrets via _FILE suffix convention
# 3. Generates self-signed TLS certificate if none exists
# 4. Creates initial empty map files for Postfix
# 5. Creates default admin user in Dovecot users file
# 6. Starts Dovecot in background
# 7. Starts Postfix in foreground

set -e

# ============================================================================
# Setup Detection
# ============================================================================

# Check for DarkPipe setup configuration
if [ -f "/config/.darkpipe-configured" ]; then
  # Load generated environment from setup script
  [ -f "/config/home.env" ] && . /config/home.env
fi

# ============================================================================
# Docker Secrets Support (_FILE suffix convention)
# ============================================================================

# Read secret from file if _FILE variant is set
for var in ADMIN_PASSWORD DKIM_PRIVATE_KEY; do
  file_var="${var}_FILE"
  eval file_path="\$$file_var"
  if [ -n "$file_path" ] && [ -f "$file_path" ]; then
    eval export "$var=\$(cat \"\$file_path\" | tr -d '\n')"
  fi
done

# ============================================================================
# Environment Variables with Defaults
# ============================================================================

# Environment variables with defaults
MAIL_DOMAIN="${MAIL_DOMAIN:-example.com}"
MAIL_HOSTNAME="${MAIL_HOSTNAME:-mail.example.com}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@example.com}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-changeme}"

# ============================================================================
# Validate Required Configuration
# ============================================================================

if [ -z "$MAIL_DOMAIN" ]; then
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "⚠️  DarkPipe setup has not been run"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""
  echo "Please run the setup script first or set required environment variables."
  echo ""
  echo "If you have the darkpipe-setup binary:"
  echo "  ./darkpipe-setup"
  echo ""
  exit 1
fi

echo "==> Starting Postfix+Dovecot mail server"
echo "    Domain: ${MAIL_DOMAIN}"
echo "    Hostname: ${MAIL_HOSTNAME}"
echo "    Admin email: ${ADMIN_EMAIL}"

# ============================================================================
# Generate Self-Signed TLS Certificate
# ============================================================================

if [ ! -f /etc/ssl/mail/cert.pem ] || [ ! -f /etc/ssl/mail/key.pem ]; then
  echo "==> Generating self-signed TLS certificate"
  openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
    -keyout /etc/ssl/mail/key.pem \
    -out /etc/ssl/mail/cert.pem \
    -subj "/C=US/ST=State/L=City/O=DarkPipe/CN=${MAIL_HOSTNAME}" \
    2>/dev/null

  chmod 0600 /etc/ssl/mail/key.pem
  chmod 0644 /etc/ssl/mail/cert.pem
  echo "    Generated certificate for ${MAIL_HOSTNAME}"
else
  echo "==> Using existing TLS certificate"
fi

# ============================================================================
# Substitute Environment Variables in Postfix Configuration
# ============================================================================

echo "==> Configuring Postfix"

# Substitute myhostname
sed -i "s/^myhostname = .*/myhostname = ${MAIL_HOSTNAME}/" /etc/postfix/main.cf

# Substitute mydomain
sed -i "s/^mydomain = .*/mydomain = ${MAIL_DOMAIN}/" /etc/postfix/main.cf

# Substitute virtual_mailbox_domains (supports multiple domains comma-separated)
# Default: example.com, example.org
VIRTUAL_DOMAINS="${VIRTUAL_DOMAINS:-example.com, example.org}"
sed -i "s/^virtual_mailbox_domains = .*/virtual_mailbox_domains = ${VIRTUAL_DOMAINS}/" /etc/postfix/main.cf

# Substitute postmaster_address in Dovecot config
sed -i "s/postmaster_address = .*/postmaster_address = postmaster@${MAIL_DOMAIN}/" /etc/dovecot/dovecot.conf

# ============================================================================
# Create Initial Postfix Map Files
# ============================================================================

echo "==> Initializing Postfix map files"

# Check if vmailbox file exists in config directory
if [ -f /etc/postfix/vmailbox ]; then
  echo "    Using existing vmailbox file"
else
  echo "    ERROR: vmailbox file not found at /etc/postfix/vmailbox"
  exit 1
fi

# Check if virtual alias map exists
if [ -f /etc/postfix/virtual ]; then
  echo "    Using existing virtual alias map"
else
  echo "    ERROR: virtual alias map not found at /etc/postfix/virtual"
  exit 1
fi

# Convert text map files to LMDB database format
postmap lmdb:/etc/postfix/vmailbox
postmap lmdb:/etc/postfix/virtual
echo "    Postfix maps compiled to LMDB format"

# ============================================================================
# Create Initial Dovecot Users File
# ============================================================================

echo "==> Initializing Dovecot users"

if [ -f /etc/dovecot/users ]; then
  chmod 0600 /etc/dovecot/users
  echo "    Using existing users file"
else
  echo "    ERROR: users file not found at /etc/dovecot/users"
  exit 1
fi

# ============================================================================
# Create Maildirs for All Users in vmailbox
# ============================================================================

echo "==> Creating maildirs for all users"

# Read vmailbox file and create directories for each user
while IFS= read -r line; do
  # Skip comments and empty lines
  [[ "$line" =~ ^#.*$ ]] && continue
  [[ -z "$line" ]] && continue

  # Extract email and path from vmailbox entry
  # Format: email@domain  domain/user/
  email=$(echo "$line" | awk '{print $1}')
  maildir_path=$(echo "$line" | awk '{print $2}')

  # Full maildir path
  full_maildir="/var/mail/vhosts/${maildir_path}Maildir"

  if [ ! -d "${full_maildir}" ]; then
    echo "    Creating maildir for ${email}"
    mkdir -p "${full_maildir}"/{cur,new,tmp}
  fi
done < /etc/postfix/vmailbox

# Set ownership and permissions for all vhosts
chown -R vmail:vmail /var/mail/vhosts
chmod -R 0700 /var/mail/vhosts

echo "    Maildirs created and permissions set"

# ============================================================================
# Fix Postfix Queue Directory Permissions
# ============================================================================

# Ensure Postfix queue directory exists and has correct permissions
if [ ! -d /var/spool/postfix/private ]; then
  mkdir -p /var/spool/postfix/private
fi

# Create pid directory for Postfix
if [ ! -d /var/spool/postfix/pid ]; then
  mkdir -p /var/spool/postfix/pid
fi

# Set correct permissions for Postfix directories
chown -R postfix:postfix /var/spool/postfix
chmod 0755 /var/spool/postfix
chmod 0700 /var/spool/postfix/private

# ============================================================================
# Start Dovecot in Background
# ============================================================================

echo "==> Starting Dovecot"
dovecot -F &
DOVECOT_PID=$!

# Wait for Dovecot to create LMTP socket
sleep 2

if [ ! -S /var/spool/postfix/private/dovecot-lmtp ]; then
  echo "ERROR: Dovecot LMTP socket not created"
  kill $DOVECOT_PID 2>/dev/null || true
  exit 1
fi

echo "    Dovecot started (PID: ${DOVECOT_PID})"

# ============================================================================
# Start Postfix in Foreground
# ============================================================================

echo "==> Starting Postfix"

# Initialize Postfix if needed
if [ ! -f /var/spool/postfix/pid/master.pid ]; then
  postfix check || true
fi

# Trap SIGTERM for graceful shutdown
trap "echo 'Shutting down...'; postfix stop; kill ${DOVECOT_PID} 2>/dev/null || true; exit 0" SIGTERM SIGINT

# Start Postfix in foreground mode
exec postfix start-fg
