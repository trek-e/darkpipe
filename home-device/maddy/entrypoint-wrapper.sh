#!/bin/sh
set -e

echo "DarkPipe Maddy entrypoint wrapper starting..."

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

# ============================================================================
# Start Maddy Mail Server
# ============================================================================

echo "Starting Maddy for domain: ${MAIL_DOMAIN}"

# Execute the original maddy binary with all arguments
exec /bin/maddy "$@"
