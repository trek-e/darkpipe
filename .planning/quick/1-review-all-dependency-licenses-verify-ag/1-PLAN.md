---
phase: quick-1
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - LICENSE
  - THIRD-PARTY-LICENSES.md
  - "transport/**/*.go"
  - "cloud-relay/**/*.go"
  - "home-device/**/*.go"
  - "dns/**/*.go"
  - "monitoring/**/*.go"
  - "deploy/setup/**/*.go"
autonomous: true
must_haves:
  truths:
    - "Repository has AGPLv3 license file at root"
    - "All 178 Go files have SPDX copyright header as first two lines"
    - "Third-party license document exists listing all dependencies by license type"
  artifacts:
    - path: "LICENSE"
      provides: "Full AGPLv3 license text"
      contains: "GNU AFFERO GENERAL PUBLIC LICENSE"
    - path: "THIRD-PARTY-LICENSES.md"
      provides: "All dependency licenses grouped by type"
      contains: "Apache-2.0"
  key_links: []
---

<objective>
Add AGPLv3 licensing infrastructure to the DarkPipe repository.

Purpose: Establish legal licensing before the repository goes public at beta. All Go dependency licenses (MIT, BSD-3-Clause, Apache-2.0) have been verified AGPLv3-compatible. Service software runs as separate Docker processes (mere aggregation).
Output: LICENSE file, copyright headers on all Go source files, THIRD-PARTY-LICENSES.md
</objective>

<execution_context>
@/Users/trekkie/.claude/get-shit-done/workflows/execute-plan.md
@/Users/trekkie/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/STATE.md
</context>

<tasks>

<task type="auto">
  <name>Task 1: Create LICENSE file and add SPDX copyright headers to all Go files</name>
  <files>LICENSE, all 178 .go files across transport/, cloud-relay/, home-device/, dns/, monitoring/, deploy/setup/</files>
  <action>
1. Create `LICENSE` at repo root with the FULL GNU Affero General Public License v3.0 text.
   - Fetch from https://www.gnu.org/licenses/agpl-3.0.txt or use the canonical AGPLv3 text
   - Prepend a copyright notice block before the license body:
     ```
     DarkPipe - Self-hosted encrypted email infrastructure
     Copyright (C) 2026 The Artificer of Ciphers, LLC. North Carolina, USA

     This program is free software: you can redistribute it and/or modify
     it under the terms of the GNU Affero General Public License as published by
     the Free Software Foundation, either version 3 of the License, or
     (at your option) any later version.

     This program is distributed in the hope that it will be useful,
     but WITHOUT ANY WARRANTY; without even the implied warranty of
     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
     GNU Affero General Public License for more details.

     You should have received a copy of the GNU Affero General Public License
     along with this program. If not, see <https://www.gnu.org/licenses/>.
     ```
   - Then the full AGPLv3 text below

2. Add SPDX copyright header to ALL .go files using find+sed (NOT manual editing).
   - The header to prepend (exactly these two lines followed by a blank line):
     ```
     // Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
     // SPDX-License-Identifier: AGPL-3.0-or-later
     ```
   - Use a command like:
     ```bash
     find . -name "*.go" -not -path "*/vendor/*" -exec sed -i '' '1s/^/\/\/ Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.\n\/\/ SPDX-License-Identifier: AGPL-3.0-or-later\n\n/' {} \;
     ```
   - IMPORTANT: Verify no file gets double-headers if re-run. Check a sample of files after.
   - IMPORTANT: The blank line between header and `package` declaration must exist.
  </action>
  <verify>
- `test -f LICENSE` — file exists
- `head -1 LICENSE` contains "DarkPipe" or copyright notice
- `grep -c "AFFERO GENERAL PUBLIC LICENSE" LICENSE` returns at least 1
- `find . -name "*.go" -not -path "*/vendor/*" | xargs head -1 | grep -c "Copyright"` equals 178
- `find . -name "*.go" -not -path "*/vendor/*" | xargs head -2 | grep -c "SPDX-License-Identifier"` equals 178
- Spot-check 3-4 files from different directories to confirm header + blank line + package line
- `go build ./...` still compiles (headers don't break anything)
  </verify>
  <done>LICENSE file contains full AGPLv3 text with copyright preamble. All 178 .go files have the two-line SPDX copyright header prepended with a blank line separator before the package declaration.</done>
</task>

<task type="auto">
  <name>Task 2: Create THIRD-PARTY-LICENSES.md documenting all dependency licenses</name>
  <files>THIRD-PARTY-LICENSES.md</files>
  <action>
Create `THIRD-PARTY-LICENSES.md` at repo root documenting all Go dependencies across all 3 modules (root, deploy/setup, home-device/profiles).

Structure:
```markdown
# Third-Party Licenses

DarkPipe uses the following third-party Go libraries. All licenses are compatible with AGPLv3.

## Apache-2.0

| Module | Version |
|--------|---------|
| ... | ... |

## BSD-3-Clause

| Module | Version |
|--------|---------|
| ... | ... |

## MIT

| Module | Version |
|--------|---------|
| ... | ... |

---

## Service Software (Docker containers — mere aggregation)

The following software runs as separate processes in Docker containers and is NOT linked with DarkPipe code:

| Software | License | Notes |
|----------|---------|-------|
| Postfix | Eclipse Public License 2.0 / IBM Public License | SMTP MTA |
| Stalwart | AGPLv3 | IMAP/JMAP server |
| Rspamd | Apache-2.0 | Spam filtering |
| Redis | BSD-3-Clause | Caching for Rspamd |
| Caddy | Apache-2.0 | Reverse proxy/webmail |
| Roundcube | GPLv3 | Webmail client |
| SnappyMail | AGPLv3 | Webmail client |
| Radicale | GPLv3 | CalDAV/CardDAV |
| WireGuard | GPLv2 | VPN tunnel (kernel module) |
| step-ca | Apache-2.0 | Internal CA |
| Certbot | Apache-2.0 | ACME client |
```

Dependency license assignments (from completed research — all verified):

**Apache-2.0:**
- github.com/aws/aws-sdk-go-v2 (and all aws/ sub-packages)
- github.com/aws/smithy-go
- github.com/cloudflare/cloudflare-go/v6
- github.com/dustin/go-humanize
- github.com/go-ini/ini
- github.com/google/btree
- github.com/google/go-cmp
- github.com/google/uuid
- github.com/minio/minio-go/v7 (and minio/ sub-packages)
- github.com/klauspost/compress
- github.com/klauspost/cpuid/v2
- github.com/klauspost/crc32
- github.com/philhofer/fwd
- github.com/tinylib/msgp
- github.com/spf13/cobra
- github.com/spf13/pflag
- github.com/inconshreveable/mousetrap
- cloud.google.com/go/compute/metadata
- golang.org/x/* (all — crypto, mod, net, sync, sys, term, text, tools, time, oauth2, telemetry, exp, xerrors)
- go.yaml.in/yaml/v3

**BSD-3-Clause:**
- filippo.io/age
- filippo.io/edwards25519
- filippo.io/hpke
- filippo.io/nistec
- github.com/emersion/go-smtp
- github.com/emersion/go-sasl
- github.com/emersion/go-msgauth
- github.com/emersion/go-message
- github.com/emersion/go-milter
- github.com/emersion/go-imap/v2
- github.com/emersion/go-ical
- github.com/emersion/go-vcard
- github.com/emersion/go-webdav
- github.com/google/btree (also Apache — dual-licensed, list under Apache)
- github.com/miekg/dns
- github.com/cenkalti/backoff/v4
- github.com/rs/xid
- golang.zx2c4.com/wireguard
- golang.zx2c4.com/wireguard/wgctrl
- golang.zx2c4.com/wintun
- gvisor.dev/gvisor
- github.com/josharian/native
- github.com/mdlayher/genetlink
- github.com/mdlayher/netlink
- github.com/mdlayher/socket
- github.com/mikioh/ipaddr
- github.com/rogpeppe/go-internal
- gopkg.in/check.v1
- gopkg.in/yaml.v3
- c2sp.org/CCTV/age

**MIT:**
- github.com/fatih/color
- github.com/mattn/go-colorable
- github.com/mattn/go-isatty
- github.com/mattn/go-runewidth
- github.com/micromdm/plist
- github.com/skip2/go-qrcode
- github.com/tidwall/gjson
- github.com/tidwall/match
- github.com/tidwall/pretty
- github.com/tidwall/sjson
- github.com/stretchr/testify
- github.com/yuin/goldmark
- github.com/AlecAivazis/survey/v2
- github.com/pterm/pterm
- github.com/containerd/console
- github.com/lithammer/fuzzysearch
- github.com/rivo/uniseg
- github.com/teambition/rrule-go
- github.com/xo/terminfo
- github.com/gookit/color
- github.com/mgutz/ansi
- github.com/kballard/go-shellquote
- atomicgo.dev/cursor
- atomicgo.dev/keyboard
- atomicgo.dev/schedule

Deduplicate entries that appear in multiple modules. List highest version used. Verify by checking a few module repos on GitHub if uncertain about any license classification. Use `go-licenses` or manual checks for any edge cases.
  </action>
  <verify>
- `test -f THIRD-PARTY-LICENSES.md` — file exists
- Document has sections for Apache-2.0, BSD-3-Clause, MIT
- Document has Service Software section
- All dependencies from `go list -m all` across all 3 modules are accounted for (minus the darkpipe module itself)
- No duplicate entries within a section
  </verify>
  <done>THIRD-PARTY-LICENSES.md exists at repo root with all Go dependencies grouped by license type (Apache-2.0, BSD-3-Clause, MIT) and a separate service software section. All deps from all 3 Go modules are accounted for with no duplicates.</done>
</task>

</tasks>

<verification>
- `LICENSE` file exists with full AGPLv3 text and copyright preamble
- All 178 `.go` files have SPDX copyright header (first two lines)
- `THIRD-PARTY-LICENSES.md` covers all dependencies across all 3 Go modules
- `go build ./...` passes in all module directories (headers did not break compilation)
- `grep -r "SPDX-License-Identifier" --include="*.go" | wc -l` equals 178
</verification>

<success_criteria>
- LICENSE file with full AGPLv3 text at repo root
- 178/178 Go files have copyright + SPDX header
- THIRD-PARTY-LICENSES.md documents all dependencies grouped by license type
- All Go code still compiles
</success_criteria>

<output>
After completion, create `.planning/quick/1-review-all-dependency-licenses-verify-ag/1-SUMMARY.md`
</output>
