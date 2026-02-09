#!/bin/sh
# Post-renewal hook for Let's Encrypt certificates
# This script runs after certbot successfully renews certificates

echo "Certificate renewed at $(date)"
echo "Certificate files updated in /etc/letsencrypt/live/${RENEWED_DOMAINS}"

# The actual Postfix reload is handled by the entrypoint.sh cert watcher
# which monitors the certificate mtime and triggers reload when it changes.
# This keeps the certbot container decoupled from the relay container.

echo "Postfix will detect the certificate change and reload automatically"
