#!/usr/bin/env bash
# verify-container-security.sh — Audit all compose files and Dockerfiles for required security directives.
# Exit non-zero if any check fails. Designed for CI and agent-driven verification.
#
# Checks per compose service:
#   - security_opt includes no-new-privileges:true
#   - cap_drop includes ALL
#   - read_only: true
#
# Checks per Dockerfile:
#   - HEALTHCHECK instruction present

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

COMPOSE_FILES=(
  "cloud-relay/docker-compose.yml"
  "home-device/docker-compose.yml"
  "cloud-relay/certbot/docker-compose.certbot.yml"
)

DOCKERFILES=(
  "cloud-relay/Dockerfile"
  "home-device/maddy/Dockerfile"
  "home-device/postfix-dovecot/Dockerfile"
  "home-device/profiles/Dockerfile"
  "home-device/stalwart/Dockerfile"
)

PASS=0
FAIL=0
TOTAL=0

pass() {
  PASS=$((PASS + 1))
  TOTAL=$((TOTAL + 1))
  printf "  ✅ PASS: %s\n" "$1"
}

fail() {
  FAIL=$((FAIL + 1))
  TOTAL=$((TOTAL + 1))
  printf "  ❌ FAIL: %s\n" "$1"
}

# ---------------------------------------------------------------------------
# Compose file checks
# ---------------------------------------------------------------------------
check_compose_file() {
  local file="$1"
  local filepath="${REPO_ROOT}/${file}"

  if [[ ! -f "$filepath" ]]; then
    printf "\n📄 %s — FILE NOT FOUND\n" "$file"
    fail "${file}: file missing"
    return
  fi

  printf "\n📄 %s\n" "$file"

  # Extract service names (top-level keys under 'services:')
  local in_services=false
  local services=()
  while IFS= read -r line; do
    # Detect the 'services:' top-level key
    if [[ "$line" =~ ^services: ]]; then
      in_services=true
      continue
    fi
    # If we're in services block, capture service names (2-space indent, ending with colon)
    if $in_services; then
      # Stop at next top-level key (no leading whitespace)
      if [[ "$line" =~ ^[a-z] ]] && [[ ! "$line" =~ ^[[:space:]] ]]; then
        break
      fi
      # Match service name: exactly 2 spaces, then a word, then colon
      if [[ "$line" =~ ^[[:space:]][[:space:]][a-zA-Z_-]+: ]] && [[ ! "$line" =~ ^[[:space:]][[:space:]][[:space:]] ]]; then
        local svc
        svc=$(echo "$line" | sed 's/^[[:space:]]*//' | cut -d: -f1)
        # Skip comment lines
        [[ "$svc" =~ ^# ]] && continue
        services+=("$svc")
      fi
    fi
  done < "$filepath"

  if [[ ${#services[@]} -eq 0 ]]; then
    fail "${file}: no services found"
    return
  fi

  for svc in "${services[@]}"; do
    printf "  🔍 Service: %s\n" "$svc"

    # Extract the service block — from "  <svc>:" until next service or top-level key
    local svc_block
    svc_block=$(awk -v svc="  ${svc}:" '
      BEGIN { found=0 }
      $0 ~ "^"svc"$" || $0 ~ "^"svc" " { found=1; next }
      found && /^  [a-zA-Z_-]+:/ && !/^    / { found=0 }
      found && /^[a-z]/ { found=0 }
      found { print }
    ' "$filepath")

    # Check security_opt: no-new-privileges
    if echo "$svc_block" | grep -q 'no-new-privileges'; then
      pass "${svc}: security_opt no-new-privileges"
    else
      fail "${svc}: missing security_opt no-new-privileges"
    fi

    # Check cap_drop: ALL
    if echo "$svc_block" | grep -q 'cap_drop' && echo "$svc_block" | grep -A5 'cap_drop' | grep -q 'ALL'; then
      pass "${svc}: cap_drop ALL"
    else
      fail "${svc}: missing cap_drop ALL"
    fi

    # Check read_only: true
    if echo "$svc_block" | grep -qE 'read_only:\s*true'; then
      pass "${svc}: read_only true"
    else
      fail "${svc}: missing read_only true"
    fi
  done
}

# ---------------------------------------------------------------------------
# Dockerfile HEALTHCHECK check
# ---------------------------------------------------------------------------
check_dockerfile() {
  local file="$1"
  local filepath="${REPO_ROOT}/${file}"

  if [[ ! -f "$filepath" ]]; then
    printf "\n🐳 %s — FILE NOT FOUND\n" "$file"
    fail "${file}: file missing"
    return
  fi

  printf "\n🐳 %s\n" "$file"

  if grep -q '^HEALTHCHECK' "$filepath"; then
    pass "${file}: HEALTHCHECK present"
  else
    fail "${file}: missing HEALTHCHECK instruction"
  fi
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
printf "=== DarkPipe Container Security Verification ===\n"
printf "Checking %d compose files and %d Dockerfiles\n" "${#COMPOSE_FILES[@]}" "${#DOCKERFILES[@]}"

for f in "${COMPOSE_FILES[@]}"; do
  check_compose_file "$f"
done

for f in "${DOCKERFILES[@]}"; do
  check_dockerfile "$f"
done

printf "\n=== Summary ===\n"
printf "Total: %d | Pass: %d | Fail: %d\n" "$TOTAL" "$PASS" "$FAIL"

if [[ $FAIL -gt 0 ]]; then
  printf "\n⚠️  %d check(s) failed. See details above.\n" "$FAIL"
  exit 1
else
  printf "\n✅ All checks passed.\n"
  exit 0
fi
