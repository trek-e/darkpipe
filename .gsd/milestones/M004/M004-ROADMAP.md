# M004: DarkPipe Website (darkpipe.org)

**Vision:** A striking, fast, accessible website at darkpipe.org that announces DarkPipe to the world — conveying email sovereignty on your own hardware through distinctive design, clear feature breakdowns, a quickstart guide, and FAQ.

## Success Criteria

- darkpipe.org loads with hero landing page, features page, quickstart page, and FAQ page
- Lighthouse performance score ≥90
- Lighthouse accessibility score ≥90
- Lighthouse best practices score ≥90
- Lighthouse SEO score ≥90
- Mobile viewport is responsive and usable at 375px width
- Page load time under 2 seconds on broadband
- All internal links resolve (no 404s)
- Quickstart content matches current repo docs/quickstart.md
- FAQ content matches current repo docs/faq.md with runtime-agnostic language
- Site deploys to Cloudflare Pages with HTTPS on darkpipe.org

## Key Risks / Unknowns

- Design quality — must feel distinctive and trustworthy, not like a template
- Architecture diagram clarity — the relay/tunnel/home flow must be visually intuitive
- Content fidelity — quickstart and FAQ must stay accurate relative to the evolving repo

## Proof Strategy

- Design quality → retire in S01 by building the hero page with full visual treatment and reviewing in browser
- Architecture diagram → retire in S01 by building the SVG/CSS diagram and verifying it communicates the flow clearly
- Content fidelity → retire in S02 by sourcing quickstart/FAQ directly from repo docs and verifying accuracy

## Verification Classes

- Contract verification: Lighthouse audit, HTML validation, link checker, responsive viewport test
- Integration verification: Cloudflare Pages deploy, custom domain resolution, HTTPS
- Operational verification: none (static site)
- UAT / human verification: visual design review, architecture diagram clarity, mobile usability

## Milestone Definition of Done

This milestone is complete only when all are true:

- All slices are complete with passing verification
- All 4 pages render correctly in browser
- Lighthouse scores ≥90 across all categories
- Site is deployed and accessible at darkpipe.org
- Mobile responsive at 375px, 768px, and 1280px breakpoints
- No broken links
- Content accuracy verified against repo docs

## Requirement Coverage

- Covers: none (new initiative, no prior requirements)
- Partially covers: none
- Leaves for later: documentation portal, blog, i18n
- Orphan risks: none

## Slices

- [x] **S01: Astro Project & Hero Landing Page** `risk:high` `depends:[]`
  > After this: The Astro project is set up with Tailwind, the hero landing page renders locally with full visual treatment (value proposition, architecture diagram, feature highlights, CTAs), and the site has a shared layout with navigation and footer.
- [x] **S02: Features, Quickstart & FAQ Pages** `risk:medium` `depends:[S01]`
  > After this: All 4 pages are complete — features page with detailed capability breakdowns, quickstart page sourced from repo docs, FAQ page with accurate answers. All pages share the layout from S01 and are navigable.
- [x] **S03: Polish, Accessibility & Deployment** `risk:low` `depends:[S02]`
  > After this: Site passes Lighthouse ≥90 on all categories, is WCAG AA compliant, responsive across breakpoints, deployed to Cloudflare Pages at darkpipe.org with HTTPS, and all links validated.

## Boundary Map

### S01 (independent)

Produces:
- Astro project with Tailwind CSS configuration
- Shared layout component (header nav, footer, meta tags)
- Hero landing page with value proposition, architecture diagram, feature highlights, CTAs
- Design system: color palette, typography, spacing, component patterns
- Local dev server running at localhost

Consumes:
- nothing (first slice)

### S02 (depends on S01)

Produces:
- Features page with detailed DarkPipe capability breakdowns
- Quickstart page adapted from docs/quickstart.md
- FAQ page adapted from docs/faq.md
- Complete site navigation across all 4 pages

Consumes:
- S01's Astro project, shared layout, design system, and component patterns

### S03 (depends on S02)

Produces:
- Lighthouse scores ≥90 across all 4 categories
- WCAG AA accessibility compliance
- Responsive design verified at 375px, 768px, 1280px
- Cloudflare Pages deployment configuration
- Custom domain (darkpipe.org) with HTTPS
- Link validation (no 404s)
- SEO meta tags, Open Graph, structured data

Consumes:
- S02's complete 4-page site
- S01's Astro project configuration (for build/deploy setup)
