---
slice: S01
milestone: M004
status: complete
started: 2026-03-12
completed: 2026-03-12
tasks_completed: 3
tasks_total: 3
---

# S01: Astro Project & Hero Landing Page — Summary

## What Was Built

Astro 6 + Tailwind CSS 4 website in `website/` directory with a complete hero landing page:

- **Project scaffold**: Astro with static output, Tailwind via `@tailwindcss/vite`, TypeScript strict mode
- **Design system**: Dark theme with "pipe green" (#22d3a7) accent, Space Mono display font, IBM Plex Sans body, custom CSS variables for full palette
- **Shared layout**: Base layout with meta tags, Open Graph, skip-to-content link, Nav component (desktop + mobile menu), Footer
- **Hero section**: "Your email lives on your hardware" headline, tagline badge, two CTAs (Get Started + GitHub), stats row (38K LOC, 3 mail servers, 7 migration providers, arm64+x64)
- **Architecture diagram**: SVG/CSS diagram showing Internet → Cloud Relay → WireGuard/mTLS tunnel → Home Device flow with animated dashed lines, component labels, and "no storage" badge on relay
- **Feature highlights**: 6 cards in responsive grid — mail server choice, encrypted transport, offline queuing, spam filtering, device onboarding, migration tools. Each with custom SVG icon.

## Key Decisions

- Website lives in `website/` directory within the main repo (not a separate repo)
- Tailwind 4 via Vite plugin (not the older @astrojs/tailwind integration)
- Google Fonts for typography (IBM Plex Sans, Space Mono, JetBrains Mono) — loaded via stylesheet link
- Architecture diagram is pure SVG/CSS with CSS animation (no JS dependency)
- Mobile nav uses minimal vanilla JS for toggle (progressive enhancement)

## Verification

- `npm run build` succeeds (552ms build time, 1 page)
- Dev server at localhost:4321
- HTML structure: 1 H1, 2 H2s, 6 H3s (proper hierarchy)
- Accessibility: skip link, ARIA labels, landmark roles, list roles
- All 6 feature cards render with accurate content matching README

## What's Next (S02)

S02 adds the remaining 3 pages: Features (detailed), Quickstart, and FAQ — all consuming S01's layout and design system.
