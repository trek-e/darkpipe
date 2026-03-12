# S01: Astro Project & Hero Landing Page

**Goal:** Astro project with Tailwind CSS is set up, a shared layout with navigation and footer is in place, and the hero landing page renders locally with full visual treatment — value proposition, architecture diagram, feature highlights, and CTAs.
**Demo:** Run `npm run dev` in `website/` and see the hero landing page at localhost:4321 with navigation, architecture diagram, feature highlights, and footer.

## Must-Haves

- Astro project with static output and Tailwind CSS
- Shared layout component (header nav + footer + meta tags)
- Hero section with DarkPipe value proposition and CTAs
- Architecture diagram (SVG/CSS) showing cloud relay → tunnel → home device flow
- Feature highlights section (6 key capabilities)
- Dark theme with distinctive privacy-focused aesthetic
- Responsive at 375px, 768px, and 1280px
- No JavaScript required for core content

## Proof Level

- This slice proves: contract
- Real runtime required: yes (local dev server)
- Human/UAT required: yes (visual design review)

## Verification

- `cd website && npm run build` — builds without errors
- `cd website && npm run dev` — serves at localhost:4321
- Browser verification: hero page renders with nav, hero, features, architecture diagram, footer
- Responsive check: content usable at 375px viewport width

## Observability / Diagnostics

- Runtime signals: Astro dev server console (build errors, warnings)
- Inspection surfaces: localhost:4321 in browser
- Failure visibility: Build errors in terminal, browser console errors
- Redaction constraints: none

## Integration Closure

- Upstream surfaces consumed: README.md (feature list, architecture), docs/quickstart.md (CTA targets)
- New wiring introduced in this slice: `website/` directory with Astro project, package.json scripts
- What remains before the milestone is truly usable end-to-end: S02 (remaining pages), S03 (polish, a11y, deployment)

## Tasks

- [x] **T01: Scaffold Astro project with Tailwind CSS** `est:20m`
  - Why: Foundation for the entire website — need Astro + Tailwind configured with static output
  - Files: `website/package.json`, `website/astro.config.mjs`, `website/tailwind.config.mjs`, `website/src/layouts/Base.astro`
  - Do: Create Astro project in `website/` dir, add Tailwind integration, configure static output, set up base layout with meta tags, create shared nav and footer components, establish color palette and typography as CSS variables. Dark theme with distinctive aesthetic per frontend-design skill.
  - Verify: `cd website && npm run build` succeeds, dev server starts at localhost:4321
  - Done when: Base layout renders with nav and footer, dark theme applied, responsive shell works

- [x] **T02: Build hero section and architecture diagram** `est:45m`
  - Why: The hero is the first thing visitors see — must convey "your email, your hardware" instantly with a compelling architecture visualization
  - Files: `website/src/pages/index.astro`, `website/src/components/Hero.astro`, `website/src/components/ArchitectureDiagram.astro`
  - Do: Build hero section with headline, subheadline, two CTAs (GitHub repo + Quickstart). Create CSS/SVG architecture diagram showing Internet → Cloud Relay → WireGuard/mTLS tunnel → Home Device flow with animation. Bold typography, atmospheric background treatment.
  - Verify: Hero renders at localhost:4321 with headline, CTAs, and animated architecture diagram
  - Done when: Hero section is visually striking, architecture diagram clearly communicates the data flow, responsive at all breakpoints

- [x] **T03: Build feature highlights section** `est:30m`
  - Why: Below the hero, visitors need to see DarkPipe's key capabilities at a glance before deciding to explore further
  - Files: `website/src/components/Features.astro`, `website/src/pages/index.astro`
  - Do: Create 6 feature cards covering: mail server choice, encrypted transport, offline queuing, spam filtering, device onboarding, and migration tools. Each card with icon/illustration, title, and short description. Grid layout that reflows on mobile.
  - Verify: Feature cards render below hero, grid is responsive, content is accurate
  - Done when: 6 feature cards visible, grid reflows to single column on mobile, content matches README capabilities

## Files Likely Touched

- `website/package.json`
- `website/astro.config.mjs`
- `website/tailwind.config.mjs`
- `website/src/layouts/Base.astro`
- `website/src/components/Nav.astro`
- `website/src/components/Footer.astro`
- `website/src/components/Hero.astro`
- `website/src/components/ArchitectureDiagram.astro`
- `website/src/components/Features.astro`
- `website/src/pages/index.astro`
