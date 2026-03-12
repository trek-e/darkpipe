# S02: Features, Quickstart & FAQ Pages

**Goal:** All 4 pages complete — features page with detailed capability breakdowns, quickstart page adapted from repo docs, FAQ page with accurate answers. All pages share the S01 layout and are navigable.
**Demo:** Navigate between all 4 pages at localhost:4321 — landing, /features, /quickstart, /faq. All content renders, all internal links work.

## Must-Haves

- Features page with detailed DarkPipe capability breakdowns
- Quickstart page adapted from docs/quickstart.md
- FAQ page adapted from docs/faq.md with accordion UI
- All pages use shared layout from S01
- Cross-page navigation works (nav links, CTAs)
- Content accuracy matches repo docs

## Proof Level

- This slice proves: contract
- Real runtime required: yes (local dev server)
- Human/UAT required: yes (content accuracy review)

## Verification

- `cd website && npm run build` — builds without errors, 4 pages output
- All 4 routes respond 200: /, /features, /quickstart, /faq
- Internal links resolve (no 404s within site)
- Content spot-checks: quickstart prerequisites match docs, FAQ answers match docs

## Observability / Diagnostics

- Runtime signals: Astro build output (page count, errors)
- Inspection surfaces: localhost:4321 in browser, curl for HTML checks
- Failure visibility: Build errors in terminal
- Redaction constraints: none

## Integration Closure

- Upstream surfaces consumed: S01 layout/components, docs/quickstart.md, docs/faq.md, docs/configuration.md, README.md
- New wiring introduced in this slice: 3 new page routes
- What remains before the milestone is truly usable end-to-end: S03 (polish, a11y audit, deployment)

## Tasks

- [x] **T01: Build Features page** `est:30m`
  - Why: Detailed feature breakdowns beyond the 6 highlight cards on the landing page
  - Files: `website/src/pages/features.astro`
  - Do: Create features page with sections covering mail server options (Stalwart/Maddy/Postfix+Dovecot comparison), transport encryption, offline queuing, spam filtering, device onboarding, migration, monitoring, DNS automation, multi-user/multi-domain, and container runtime support. Use consistent card/section design from S01.
  - Verify: `/features` returns 200, content matches README and configuration docs
  - Done when: Features page renders with all major DarkPipe capabilities documented

- [x] **T02: Build Quickstart page** `est:30m`
  - Why: Website quickstart adapted from docs/quickstart.md — the primary onboarding path for new users
  - Files: `website/src/pages/quickstart.astro`
  - Do: Adapt docs/quickstart.md into a web-native format. Structured steps with code blocks, numbered sections. Keep runtime-agnostic language with Podman callouts. Include all 8 steps from the guide.
  - Verify: `/quickstart` returns 200, step count matches docs, code blocks render
  - Done when: Quickstart page renders all 8 steps with code blocks and proper formatting

- [x] **T03: Build FAQ page** `est:25m`
  - Why: Common questions with expandable answers — reduces friction for evaluators
  - Files: `website/src/pages/faq.astro`, `website/src/components/FaqItem.astro`
  - Do: Adapt docs/faq.md into accordion-style FAQ. Group by category (General, Setup, Mail Servers, Security, Operations, Troubleshooting). Minimal JS for expand/collapse using details/summary HTML elements (no JS dependency for core content).
  - Verify: `/faq` returns 200, all FAQ categories present, expand/collapse works
  - Done when: FAQ page renders all categories with expandable questions and accurate answers

## Files Likely Touched

- `website/src/pages/features.astro`
- `website/src/pages/quickstart.astro`
- `website/src/pages/faq.astro`
- `website/src/components/FaqItem.astro`
