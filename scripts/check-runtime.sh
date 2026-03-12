#!/usr/bin/env bash
# check-runtime.sh — Detect the user's container runtime environment and
# validate prerequisites for running DarkPipe.
#
# Checks:
#   1. Runtime detection: Docker or Podman available
#   2. Version validation: Docker 24+ or Podman 5.3+
#   3. Compose tool: docker compose or podman-compose available
#   4. SELinux state: enforcing / permissive / disabled / not installed
#   5. Network basics: port 25 not already bound (cloud relay prerequisite)
#
# Exit codes:
#   0 — all checks passed
#   1 — one or more checks failed
#
# Output format matches verify-podman-compat.sh (PASS/FAIL/SKIP pattern).

set -euo pipefail

# ---------------------------------------------------------------------------
# Usage
# ---------------------------------------------------------------------------
usage() {
  cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Detect the container runtime environment and validate DarkPipe prerequisites.

Options:
  --help    Show this help message and exit
  --quiet   Suppress per-check output; print only the summary line

Checks performed:
  • Container runtime detection (Docker or Podman)
  • Runtime version validation (Docker 24+ / Podman 5.3+)
  • Compose tool availability (docker compose / podman-compose)
  • SELinux enforcement state
  • Port 25 availability (cloud relay prerequisite)

Exit codes:
  0  All checks passed
  1  One or more checks failed
EOF
  exit 0
}

# ---------------------------------------------------------------------------
# Parse arguments
# ---------------------------------------------------------------------------
QUIET=false

for arg in "$@"; do
  case "$arg" in
    --help) usage ;;
    --quiet) QUIET=true ;;
    *) printf "Unknown option: %s\n" "$arg" >&2; exit 2 ;;
  esac
done

# ---------------------------------------------------------------------------
# Counters & helpers (same pattern as verify-podman-compat.sh)
# ---------------------------------------------------------------------------
PASS=0
FAIL=0
SKIP=0
TOTAL=0

pass() {
  PASS=$((PASS + 1))
  TOTAL=$((TOTAL + 1))
  [[ "$QUIET" == true ]] && return
  printf "  ✅ PASS: %s\n" "$1"
}

fail() {
  FAIL=$((FAIL + 1))
  TOTAL=$((TOTAL + 1))
  [[ "$QUIET" == true ]] && return
  printf "  ❌ FAIL: %s\n" "$1"
  if [[ -n "${2:-}" ]]; then
    printf "          → Detected: %s\n" "$2"
  fi
  if [[ -n "${3:-}" ]]; then
    printf "          → Required: %s\n" "$3"
  fi
  if [[ -n "${4:-}" ]]; then
    printf "          → Fix: %s\n" "$4"
  fi
}

skip() {
  SKIP=$((SKIP + 1))
  [[ "$QUIET" == true ]] && return
  printf "  ⏭️  SKIP: %s\n" "$1"
}

# ---------------------------------------------------------------------------
# State variables populated during checks
# ---------------------------------------------------------------------------
DETECTED_RUNTIME=""       # "docker" | "podman" | ""
DETECTED_VERSION=""       # raw version string
DETECTED_COMPOSE=""       # "docker compose" | "podman-compose" | ""
DETECTED_SELINUX=""       # "Enforcing" | "Permissive" | "Disabled" | "not installed"

# ---------------------------------------------------------------------------
# 1. Runtime detection
# ---------------------------------------------------------------------------
check_runtime_detection() {
  [[ "$QUIET" == false ]] && printf "\n🐳 Container Runtime Detection\n"

  if command -v docker &>/dev/null; then
    DETECTED_RUNTIME="docker"
    pass "Docker found at $(command -v docker)"
  elif command -v podman &>/dev/null; then
    DETECTED_RUNTIME="podman"
    pass "Podman found at $(command -v podman)"
  else
    DETECTED_RUNTIME=""
    fail "No container runtime found" \
         "neither docker nor podman in PATH" \
         "Docker 24+ or Podman 5.3+" \
         "Install Docker (https://docs.docker.com/get-docker/) or Podman (https://podman.io/getting-started/installation)"
  fi

  # Also note if both are available
  if [[ "$DETECTED_RUNTIME" == "docker" ]] && command -v podman &>/dev/null; then
    [[ "$QUIET" == false ]] && printf "  ℹ️  INFO: Podman also available (Docker will be used as primary)\n"
  fi
}

# ---------------------------------------------------------------------------
# 2. Version validation
# ---------------------------------------------------------------------------
# Extract a semver-ish major.minor from a version string.
# Handles formats like:
#   Docker version 24.0.7, build afdd53b
#   podman version 5.3.1
#   Docker version 27.1.1-rd, build cc0ee3e
#   podman version 5.4.0-dev
parse_version() {
  local raw="$1"
  # Grab first occurrence of digits.digits (optionally .digits)
  echo "$raw" | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -1
}

# Compare "major.minor" >= "required_major.required_minor"
version_ge() {
  local ver="$1" req="$2"
  local ver_major ver_minor req_major req_minor
  ver_major="${ver%%.*}"
  ver_minor="${ver#*.}"; ver_minor="${ver_minor%%.*}"
  req_major="${req%%.*}"
  req_minor="${req#*.}"; req_minor="${req_minor%%.*}"

  if (( ver_major > req_major )); then
    return 0
  elif (( ver_major == req_major && ver_minor >= req_minor )); then
    return 0
  else
    return 1
  fi
}

check_runtime_version() {
  [[ "$QUIET" == false ]] && printf "\n📐 Runtime Version Validation\n"

  if [[ -z "$DETECTED_RUNTIME" ]]; then
    skip "No runtime detected — skipping version check"
    return
  fi

  local raw_version
  raw_version=$("$DETECTED_RUNTIME" --version 2>/dev/null || echo "")

  if [[ -z "$raw_version" ]]; then
    fail "${DETECTED_RUNTIME} --version returned no output" \
         "(empty)" \
         "A working ${DETECTED_RUNTIME} installation" \
         "Reinstall or check that the ${DETECTED_RUNTIME} daemon is running"
    return
  fi

  DETECTED_VERSION=$(parse_version "$raw_version")

  if [[ -z "$DETECTED_VERSION" ]]; then
    fail "Could not parse version from ${DETECTED_RUNTIME} --version" \
         "$raw_version" \
         "Parseable version string" \
         "Check ${DETECTED_RUNTIME} installation"
    return
  fi

  local required
  if [[ "$DETECTED_RUNTIME" == "docker" ]]; then
    required="24.0"
    if version_ge "$DETECTED_VERSION" "$required"; then
      pass "Docker version ${DETECTED_VERSION} >= ${required}"
    else
      fail "Docker version too old" \
           "$DETECTED_VERSION" \
           "${required}+" \
           "Upgrade Docker: https://docs.docker.com/engine/install/"
    fi
  elif [[ "$DETECTED_RUNTIME" == "podman" ]]; then
    required="5.3"
    if version_ge "$DETECTED_VERSION" "$required"; then
      pass "Podman version ${DETECTED_VERSION} >= ${required}"
    else
      fail "Podman version too old" \
           "$DETECTED_VERSION" \
           "${required}+" \
           "Upgrade Podman: https://podman.io/getting-started/installation"
    fi
  fi
}

# ---------------------------------------------------------------------------
# 3. Compose tool
# ---------------------------------------------------------------------------
check_compose_tool() {
  [[ "$QUIET" == false ]] && printf "\n🔧 Compose Tool Availability\n"

  # Try docker compose (v2 plugin) first
  if command -v docker &>/dev/null && docker compose version &>/dev/null; then
    local compose_ver
    compose_ver=$(docker compose version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")
    DETECTED_COMPOSE="docker compose"
    pass "docker compose available (v${compose_ver})"
    return
  fi

  # Try podman-compose
  if command -v podman-compose &>/dev/null; then
    local pc_ver
    pc_ver=$(podman-compose --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -1 || echo "unknown")
    DETECTED_COMPOSE="podman-compose"
    pass "podman-compose available (v${pc_ver})"
    return
  fi

  # Try docker-compose (legacy standalone v1/v2)
  if command -v docker-compose &>/dev/null; then
    local dc_ver
    dc_ver=$(docker-compose --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")
    DETECTED_COMPOSE="docker-compose"
    pass "docker-compose (standalone) available (v${dc_ver})"
    return
  fi

  DETECTED_COMPOSE=""
  fail "No compose tool found" \
       "none of: docker compose, podman-compose, docker-compose" \
       "A compose tool in PATH" \
       "Install docker compose plugin or podman-compose (pip install podman-compose)"
}

# ---------------------------------------------------------------------------
# 4. SELinux state
# ---------------------------------------------------------------------------
check_selinux() {
  [[ "$QUIET" == false ]] && printf "\n🔒 SELinux State\n"

  if ! command -v getenforce &>/dev/null; then
    DETECTED_SELINUX="not installed"
    skip "getenforce not found — SELinux not available on this system"
    return
  fi

  DETECTED_SELINUX=$(getenforce 2>/dev/null || echo "unknown")

  case "$DETECTED_SELINUX" in
    Enforcing)
      pass "SELinux is Enforcing — use docker-compose.podman-selinux.yml override for :z labels"
      ;;
    Permissive)
      pass "SELinux is Permissive — :z labels optional but recommended"
      ;;
    Disabled)
      pass "SELinux is Disabled — no volume label overrides needed"
      ;;
    *)
      skip "SELinux state unknown: ${DETECTED_SELINUX}"
      ;;
  esac
}

# ---------------------------------------------------------------------------
# 5. Network basics: port 25
# ---------------------------------------------------------------------------
check_port_25() {
  [[ "$QUIET" == false ]] && printf "\n🌐 Network Prerequisites\n"

  # Try multiple approaches to check port binding
  local port_in_use=false

  if command -v ss &>/dev/null; then
    if ss -tlnp 2>/dev/null | grep -qE ':25\b'; then
      port_in_use=true
    fi
  elif command -v netstat &>/dev/null; then
    if netstat -tlnp 2>/dev/null | grep -qE ':25\b'; then
      port_in_use=true
    fi
  elif command -v lsof &>/dev/null; then
    if lsof -iTCP:25 -sTCP:LISTEN &>/dev/null; then
      port_in_use=true
    fi
  else
    skip "No port-check tool available (ss, netstat, lsof) — cannot verify port 25"
    return
  fi

  if [[ "$port_in_use" == true ]]; then
    fail "Port 25 is already in use" \
         "another process is listening on port 25" \
         "port 25 available for the cloud relay SMTP server" \
         "Stop the existing SMTP service (e.g., sudo systemctl stop postfix) or configure a different port"
  else
    pass "Port 25 is available"
  fi
}

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
print_summary() {
  if [[ "$QUIET" == false ]]; then
    printf "\n=== Environment Summary ===\n"
    printf "  Runtime:  %s\n" "${DETECTED_RUNTIME:-none}"
    printf "  Version:  %s\n" "${DETECTED_VERSION:-n/a}"
    printf "  Compose:  %s\n" "${DETECTED_COMPOSE:-none}"
    printf "  SELinux:  %s\n" "${DETECTED_SELINUX:-unknown}"
  fi

  printf "\n=== Results ===\n"
  printf "Total: %d | Pass: %d | Fail: %d | Skip: %d\n" "$TOTAL" "$PASS" "$FAIL" "$SKIP"

  if [[ $FAIL -gt 0 ]]; then
    printf "\n⚠️  %d check(s) failed. See details above.\n" "$FAIL"
    exit 1
  else
    printf "\n✅ System is ready for DarkPipe.\n"
    exit 0
  fi
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
[[ "$QUIET" == false ]] && printf "=== DarkPipe Runtime Compatibility Check ===\n"

check_runtime_detection
check_runtime_version
check_compose_tool
check_selinux
check_port_25
print_summary
