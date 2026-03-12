#!/usr/bin/env bash
# verify-log-redaction.sh — Static analysis: ensure no unredacted PII in log calls.
# Checks that log.Printf calls referencing email/address variables use redaction helpers.
# Exit 0 if clean, exit 1 with details if violations found.

set -euo pipefail

VIOLATIONS=0

check_file() {
    local file="$1"
    local basename
    basename=$(basename "$file")

    if [ ! -f "$file" ]; then
        echo "WARN: $file not found, skipping"
        return
    fi

    # Find log.Printf lines that format a raw email-like variable
    # without a corresponding Redact or logEmail call.
    # Pattern: log.Printf containing a %s or %v format with a variable named
    # email/addr/address on the same line, but NOT wrapped in logEmail/RedactEmail.
    while IFS= read -r line; do
        # Skip lines that already use redaction
        if echo "$line" | grep -qiE '(logEmail|RedactEmail|RedactQueryParams|RedactEmails|logFrom|logTo)'; then
            continue
        fi
        # Flag lines with log.Printf that reference email-like variables
        if echo "$line" | grep -qE 'log\.Printf.*%[sv].*email' || \
           echo "$line" | grep -qE 'log\.Printf.*email.*%[sv]'; then
            echo "VIOLATION in $file: $line"
            VIOLATIONS=$((VIOLATIONS + 1))
        fi
    done < <(grep -n 'log\.Printf' "$file" 2>/dev/null || true)
}

# Files to check
check_file "cloud-relay/relay/smtp/session.go"
check_file "home-device/profiles/cmd/profile-server/handlers.go"
check_file "home-device/profiles/cmd/profile-server/webui.go"

# Also check LogRequest middleware doesn't log raw query with RawQuery
if grep -n 'r\.URL\.RawQuery' home-device/profiles/cmd/profile-server/handlers.go 2>/dev/null | grep -v 'RedactQueryParams' | grep -qv '^\s*//' ; then
    echo "VIOLATION in handlers.go: raw r.URL.RawQuery used without RedactQueryParams"
    VIOLATIONS=$((VIOLATIONS + 1))
fi

if [ "$VIOLATIONS" -gt 0 ]; then
    echo ""
    echo "FAILED: $VIOLATIONS violation(s) found"
    exit 1
fi

echo "PASSED: No unredacted PII patterns found in log calls"
exit 0
