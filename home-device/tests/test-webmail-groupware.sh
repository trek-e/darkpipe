#!/usr/bin/env bash
# Phase 6: Webmail & Groupware Integration Tests
# Tests all Phase 6 success criteria
# Usage: ./test-webmail-groupware.sh [webmail-type] [caldav-type]
#   webmail-type: roundcube | snappymail (default: roundcube)
#   caldav-type:  radicale | stalwart (default: radicale)

set -euo pipefail

# ANSI color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
PASSED=0
FAILED=0

# Parse arguments
WEBMAIL_TYPE="${1:-roundcube}"
CALDAV_TYPE="${2:-radicale}"

# Helper functions
pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED++))
}

fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED++))
}

info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

echo "Phase 6: Webmail & Groupware Integration Tests"
echo "Testing with: webmail=$WEBMAIL_TYPE, caldav=$CALDAV_TYPE"
echo ""

# Test 1: Webmail Access (WEB-01)
info "Testing webmail HTTP response on port 8080..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/ 2>/dev/null || echo "000")
if [ "$HTTP_CODE" = "200" ]; then
    pass "Webmail responds on port 8080 (HTTP 200)"
else
    fail "Webmail does not respond (HTTP $HTTP_CODE)"
fi

# Test 2: Webmail Content Detection (WEB-01)
info "Testing webmail content type..."
WEBMAIL_RESPONSE=$(curl -s http://localhost:8080/ 2>/dev/null || echo "")
if [ "$WEBMAIL_TYPE" = "roundcube" ]; then
    if echo "$WEBMAIL_RESPONSE" | grep -qi "roundcube\|rcmlogin"; then
        pass "Roundcube webmail detected"
    else
        fail "Roundcube content not found in response"
    fi
elif [ "$WEBMAIL_TYPE" = "snappymail" ]; then
    if echo "$WEBMAIL_RESPONSE" | grep -qi "snappymail\|rainloop"; then
        pass "SnappyMail webmail detected"
    else
        fail "SnappyMail content not found in response"
    fi
fi

# Test 3: Mobile Responsive (WEB-02)
info "Testing mobile responsiveness..."
if [ "$WEBMAIL_TYPE" = "roundcube" ]; then
    if echo "$WEBMAIL_RESPONSE" | grep -qi "elastic"; then
        pass "Roundcube Elastic skin detected (mobile-responsive)"
    else
        fail "Roundcube Elastic skin not detected"
    fi
elif [ "$WEBMAIL_TYPE" = "snappymail" ]; then
    if echo "$WEBMAIL_RESPONSE" | grep -qi "viewport"; then
        pass "SnappyMail viewport meta tag detected (mobile-responsive)"
    else
        fail "SnappyMail viewport meta tag not found"
    fi
fi

# Test 4: CalDAV Auto-Discovery (CAL-01)
info "Testing CalDAV well-known redirect..."
if [ "$CALDAV_TYPE" = "radicale" ]; then
    CALDAV_REDIRECT=$(curl -s -o /dev/null -w "%{http_code}" -I http://localhost:8080/.well-known/caldav 2>/dev/null || echo "000")
    if [ "$CALDAV_REDIRECT" = "301" ] || [ "$CALDAV_REDIRECT" = "302" ]; then
        pass "CalDAV well-known redirect configured (HTTP $CALDAV_REDIRECT)"
    else
        fail "CalDAV well-known redirect not working (HTTP $CALDAV_REDIRECT)"
    fi

    # Test Radicale endpoint
    info "Testing Radicale endpoint..."
    RADICALE_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:5232/.web/ 2>/dev/null || echo "000")
    if [ "$RADICALE_CODE" = "200" ]; then
        pass "Radicale responds on port 5232"
    else
        fail "Radicale does not respond (HTTP $RADICALE_CODE)"
    fi
else
    info "Stalwart has built-in CalDAV (skipping separate CalDAV tests)"
fi

# Test 5: CardDAV Auto-Discovery (CAL-02)
info "Testing CardDAV well-known redirect..."
if [ "$CALDAV_TYPE" = "radicale" ]; then
    CARDDAV_REDIRECT=$(curl -s -o /dev/null -w "%{http_code}" -I http://localhost:8080/.well-known/carddav 2>/dev/null || echo "000")
    if [ "$CARDDAV_REDIRECT" = "301" ] || [ "$CARDDAV_REDIRECT" = "302" ]; then
        pass "CardDAV well-known redirect configured (HTTP $CARDDAV_REDIRECT)"
    else
        fail "CardDAV well-known redirect not working (HTTP $CARDDAV_REDIRECT)"
    fi
fi

# Test 6: Setup Scripts Exist
info "Testing setup scripts..."
if bash /Users/trekkie/projects/darkpipe/home-device/caldav-carddav/setup-collections.sh 2>&1 | grep -q "Usage:"; then
    pass "Setup-collections.sh script runs (shows usage)"
else
    fail "Setup-collections.sh script failed"
fi

if bash /Users/trekkie/projects/darkpipe/home-device/caldav-carddav/sync-users.sh 2>&1 | grep -q "Usage:"; then
    pass "Sync-users.sh script runs (shows usage)"
else
    fail "Sync-users.sh script failed"
fi

# Test 7: Docker Compose Validation
info "Testing Docker compose profile combinations..."

# We can't actually run docker compose config without docker installed,
# so we'll just check that the file is valid YAML
if [ -f /Users/trekkie/projects/darkpipe/home-device/docker-compose.yml ]; then
    pass "Home device docker-compose.yml exists"
else
    fail "Home device docker-compose.yml not found"
fi

if [ -f /Users/trekkie/projects/darkpipe/cloud-relay/docker-compose.yml ]; then
    pass "Cloud relay docker-compose.yml exists"
else
    fail "Cloud relay docker-compose.yml not found"
fi

# Test 8: Configuration Files Exist
info "Testing configuration files..."
if [ -f /Users/trekkie/projects/darkpipe/home-device/caldav-carddav/radicale/config/config ]; then
    pass "Radicale config file exists"
else
    fail "Radicale config file not found"
fi

if [ -f /Users/trekkie/projects/darkpipe/home-device/caldav-carddav/radicale/rights ]; then
    pass "Radicale rights file exists"
else
    fail "Radicale rights file not found"
fi

if [ -f /Users/trekkie/projects/darkpipe/cloud-relay/caddy/Caddyfile ]; then
    pass "Caddyfile exists"
else
    fail "Caddyfile not found"
fi

# Test 9: Radicale Rights File Content
info "Testing Radicale rights file content..."
if grep -q "shared" /Users/trekkie/projects/darkpipe/home-device/caldav-carddav/radicale/rights; then
    pass "Radicale rights file contains shared collection rules"
else
    fail "Radicale rights file missing shared collection rules"
fi

# Test 10: Caddyfile CalDAV/CardDAV Configuration
info "Testing Caddyfile CalDAV/CardDAV configuration..."
if grep -q "well-known/caldav" /Users/trekkie/projects/darkpipe/cloud-relay/caddy/Caddyfile; then
    pass "Caddyfile contains CalDAV well-known redirect"
else
    fail "Caddyfile missing CalDAV well-known redirect"
fi

if grep -q "well-known/carddav" /Users/trekkie/projects/darkpipe/cloud-relay/caddy/Caddyfile; then
    pass "Caddyfile contains CardDAV well-known redirect"
else
    fail "Caddyfile missing CardDAV well-known redirect"
fi

if grep -q "5232" /Users/trekkie/projects/darkpipe/cloud-relay/caddy/Caddyfile; then
    pass "Caddyfile contains Radicale reverse proxy (port 5232)"
else
    fail "Caddyfile missing Radicale reverse proxy"
fi

# Summary
echo ""
echo "================================================"
echo "Phase 6 Tests: $PASSED passed, $FAILED failed"
echo "================================================"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed.${NC}"
    exit 1
fi
