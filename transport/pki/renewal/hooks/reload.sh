#!/usr/bin/env bash
# reload.sh -- Post-renewal hook: reload a systemd service after its
# certificate has been renewed by the cert-renewer@.service unit.
#
# Usage:  reload.sh <service-name>
# Example: reload.sh darkpipe-relay

set -euo pipefail

SERVICE="${1:-}"

if [[ -z "$SERVICE" ]]; then
  echo "Usage: $0 <service-name>" >&2
  exit 1
fi

logger -t cert-renewer "Reloading $SERVICE after certificate renewal"
systemctl reload "$SERVICE"
logger -t cert-renewer "Successfully reloaded $SERVICE"
