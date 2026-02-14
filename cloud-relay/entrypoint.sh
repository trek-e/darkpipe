#!/bin/sh
set -e

echo "DarkPipe cloud relay entrypoint starting..."

# ============================================================================
# Setup Detection
# ============================================================================

# Check for DarkPipe setup configuration
if [ -f "/config/.darkpipe-configured" ]; then
  # Load generated environment from setup script
  [ -f "/config/relay.env" ] && . /config/relay.env
fi

# ============================================================================
# Docker Secrets Support (_FILE suffix convention)
# ============================================================================

# Read secret from file if _FILE variant is set
for var in RELAY_WEBHOOK_URL CERTBOT_EMAIL RELAY_OVERFLOW_ACCESS_KEY RELAY_OVERFLOW_SECRET_KEY; do
  file_var="${var}_FILE"
  eval file_path="\$$file_var"
  if [ -n "$file_path" ] && [ -f "$file_path" ]; then
    eval export "$var=\$(cat \"\$file_path\" | tr -d '\n')"
  fi
done

# ============================================================================
# Validate Required Configuration
# ============================================================================

# Substitute environment variables in Postfix main.cf
if [ -z "$RELAY_HOSTNAME" ] || [ -z "$RELAY_DOMAIN" ]; then
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "⚠️  DarkPipe setup has not been run"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""
  echo "Please run the setup script first or set RELAY_HOSTNAME and RELAY_DOMAIN environment variables."
  echo ""
  echo "If you have the darkpipe-setup binary:"
  echo "  ./darkpipe-setup"
  echo ""
  exit 1
fi

echo "Configuring Postfix for hostname=$RELAY_HOSTNAME domain=$RELAY_DOMAIN"

# Use envsubst to replace placeholders in main.cf
envsubst < /etc/postfix/main.cf.template > /etc/postfix/main.cf

# Hash the transport map (LMDB format)
echo "Creating transport map..."
postmap lmdb:/etc/postfix/transport

# Validate Postfix configuration
echo "Validating Postfix configuration..."
postfix check

# Ensure queue directory permissions
mkdir -p /var/spool/postfix
chown -R postfix:postfix /var/spool/postfix

# Check if TLS certificates exist
CERT_PATH="/etc/letsencrypt/live/${RELAY_HOSTNAME}/fullchain.pem"
if [ ! -f "$CERT_PATH" ]; then
  echo "WARNING: TLS certificates not found at $CERT_PATH"
  echo "Postfix will start without TLS (smtpd_tls_security_level=none)"
  echo "Waiting for certbot to obtain certificates..."
  # Temporarily disable TLS until certs are available
  postconf -e "smtpd_tls_security_level=none"
  postconf -e "smtp_tls_security_level=may"
else
  echo "TLS certificates found at $CERT_PATH"
fi

# Start certificate watcher in background
# This monitors cert mtime and reloads Postfix when certificates are renewed
echo "Starting certificate watcher..."
(
  LAST_MTIME=""
  while true; do
    sleep 300  # Check every 5 minutes

    if [ -f "$CERT_PATH" ]; then
      CURRENT_MTIME=$(stat -c %Y "$CERT_PATH" 2>/dev/null || stat -f %m "$CERT_PATH" 2>/dev/null || echo "0")

      if [ -n "$LAST_MTIME" ] && [ "$CURRENT_MTIME" != "$LAST_MTIME" ]; then
        echo "Certificate change detected (mtime: $LAST_MTIME -> $CURRENT_MTIME)"
        echo "Reloading Postfix to pick up new certificates..."
        postfix reload
        if [ $? -eq 0 ]; then
          echo "Postfix reloaded successfully"
        else
          echo "ERROR: Postfix reload failed"
        fi
      fi

      LAST_MTIME="$CURRENT_MTIME"

      # If certs just became available and TLS was disabled, re-enable it
      if postconf smtpd_tls_security_level | grep -q "none"; then
        echo "Certificates now available, enabling TLS..."
        postconf -e "smtpd_tls_security_level=may"
        postfix reload
      fi
    fi
  done
) &
CERT_WATCHER_PID=$!

# Start relay daemon in background
echo "Starting relay daemon..."
/usr/local/bin/relay-daemon &
RELAY_PID=$!

# Trap SIGTERM to gracefully stop all processes
trap 'echo "Stopping services..."; postfix stop; kill $RELAY_PID $CERT_WATCHER_PID 2>/dev/null; wait $RELAY_PID' SIGTERM SIGINT

# Start Postfix in foreground
echo "Starting Postfix..."
postfix start-fg &
POSTFIX_PID=$!

# Wait for Postfix to exit
wait $POSTFIX_PID
