#!/usr/bin/env bash
# verify-s02-docs.sh — Verify all S02 (Runtime-Agnostic Documentation & Tooling)
# slice acceptance criteria. Exit non-zero if any check fails.
#
# Checks:
#   1. check-runtime.sh exists and is executable
#   2. Podman platform guide exists and has >50 lines
#   3. FAQ no longer contains "Not tested" / "not officially supported" for Podman
#   4. quickstart.md contains runtime-agnostic language ("container runtime")
#   5. quickstart.md mentions Podman
#   6. Platform guides (raspberry-pi, proxmox-lxc) mention Podman
#
# Output format matches verify-podman-compat.sh (PASS/FAIL/SKIP pattern).

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

PASS=0
FAIL=0
SKIP=0
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

skip() {
  SKIP=$((SKIP + 1))
  printf "  ⏭️  SKIP: %s\n" "$1"
}

# ---------------------------------------------------------------------------
# 1. check-runtime.sh exists and is executable
# ---------------------------------------------------------------------------
check_runtime_script() {
  printf "\n🔧 Runtime Check Script\n"

  local script="${REPO_ROOT}/scripts/check-runtime.sh"

  if [[ ! -f "$script" ]]; then
    fail "scripts/check-runtime.sh does not exist"
    return
  fi

  pass "scripts/check-runtime.sh exists"

  if [[ -x "$script" ]]; then
    pass "scripts/check-runtime.sh is executable"
  else
    fail "scripts/check-runtime.sh is not executable (missing +x)"
  fi

  # Syntax check
  if bash -n "$script" 2>/dev/null; then
    pass "scripts/check-runtime.sh passes bash -n syntax check"
  else
    fail "scripts/check-runtime.sh has syntax errors"
  fi

  # --help flag
  if bash "$script" --help &>/dev/null; then
    pass "scripts/check-runtime.sh --help exits 0"
  else
    fail "scripts/check-runtime.sh --help does not exit 0"
  fi
}

# ---------------------------------------------------------------------------
# 2. Podman platform guide exists and has >50 lines
# ---------------------------------------------------------------------------
check_podman_guide() {
  printf "\n📖 Podman Platform Guide\n"

  local guide="${REPO_ROOT}/deploy/platform-guides/podman.md"

  if [[ ! -f "$guide" ]]; then
    fail "deploy/platform-guides/podman.md does not exist"
    return
  fi

  pass "deploy/platform-guides/podman.md exists"

  local line_count
  line_count=$(wc -l < "$guide" | tr -d ' ')

  if (( line_count > 50 )); then
    pass "deploy/platform-guides/podman.md has ${line_count} lines (>50)"
  else
    fail "deploy/platform-guides/podman.md has only ${line_count} lines (need >50)"
  fi
}

# ---------------------------------------------------------------------------
# 3. FAQ: no "Not tested" / "not officially supported" for Podman
# ---------------------------------------------------------------------------
check_faq_podman() {
  printf "\n❓ FAQ Podman Status\n"

  local faq="${REPO_ROOT}/docs/faq.md"

  if [[ ! -f "$faq" ]]; then
    fail "docs/faq.md does not exist"
    return
  fi

  # Check for problematic phrases (case-insensitive)
  if grep -qi "not tested" "$faq"; then
    fail "docs/faq.md still contains 'Not tested'"
  else
    pass "docs/faq.md does not contain 'Not tested'"
  fi

  if grep -qi "not officially supported" "$faq"; then
    fail "docs/faq.md still contains 'not officially supported'"
  else
    pass "docs/faq.md does not contain 'not officially supported'"
  fi
}

# ---------------------------------------------------------------------------
# 4. quickstart.md contains "container runtime"
# ---------------------------------------------------------------------------
check_quickstart_agnostic() {
  printf "\n📝 Quickstart Runtime-Agnostic Language\n"

  local qs="${REPO_ROOT}/docs/quickstart.md"

  if [[ ! -f "$qs" ]]; then
    fail "docs/quickstart.md does not exist"
    return
  fi

  if grep -qi "container runtime" "$qs"; then
    pass "docs/quickstart.md contains 'container runtime'"
  else
    fail "docs/quickstart.md does not contain 'container runtime'"
  fi
}

# ---------------------------------------------------------------------------
# 5. quickstart.md mentions Podman
# ---------------------------------------------------------------------------
check_quickstart_podman() {
  printf "\n🦭 Quickstart Podman Mention\n"

  local qs="${REPO_ROOT}/docs/quickstart.md"

  if [[ ! -f "$qs" ]]; then
    fail "docs/quickstart.md does not exist"
    return
  fi

  if grep -qi "podman" "$qs"; then
    pass "docs/quickstart.md mentions Podman"
  else
    fail "docs/quickstart.md does not mention Podman"
  fi
}

# ---------------------------------------------------------------------------
# 6. Platform guides mention Podman
# ---------------------------------------------------------------------------
check_platform_guides_podman() {
  printf "\n🗺️  Platform Guide Podman Notes\n"

  local guides_with_podman=(
    "deploy/platform-guides/raspberry-pi.md"
    "deploy/platform-guides/proxmox-lxc.md"
  )

  for guide in "${guides_with_podman[@]}"; do
    local filepath="${REPO_ROOT}/${guide}"

    if [[ ! -f "$filepath" ]]; then
      fail "${guide} does not exist"
      continue
    fi

    if grep -qi "podman" "$filepath"; then
      pass "${guide} mentions Podman"
    else
      fail "${guide} does not mention Podman"
    fi
  done
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
printf "=== S02 Slice Verification: Runtime-Agnostic Documentation & Tooling ===\n"

check_runtime_script
check_podman_guide
check_faq_podman
check_quickstart_agnostic
check_quickstart_podman
check_platform_guides_podman

printf "\n=== Summary ===\n"
printf "Total: %d | Pass: %d | Fail: %d | Skip: %d\n" "$TOTAL" "$PASS" "$FAIL" "$SKIP"

if [[ $FAIL -gt 0 ]]; then
  printf "\n⚠️  %d check(s) failed. See details above.\n" "$FAIL"
  exit 1
else
  printf "\n✅ All S02 acceptance criteria met.\n"
  exit 0
fi
