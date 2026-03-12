# M004: DarkPipe Website (darkpipe.org) — Context

**Gathered:** 2026-03-12
**Status:** Ready for planning

## Project Description

Build the public-facing website for darkpipe.org — an Astro + Tailwind CSS static site deployed to Cloudflare Pages. The site announces DarkPipe to the world with a striking hero landing page, feature breakdowns, quickstart guide, and FAQ. It should convey the project's core value proposition (email sovereignty on your own hardware) with a distinctive, privacy-focused aesthetic that reflects the project's philosophy.

## Why This Milestone

DarkPipe has shipped v1.0, completed post-launch hardening (M002), and expanded runtime compatibility (M003). There is no public website — the project exists only as a GitHub repository. A dedicated website is needed to:

1. **Announce DarkPipe** to privacy-conscious users, self-hosters, and the broader open-source community
2. **Explain the value proposition** in a way a README cannot — visual storytelling, architecture diagrams, feature highlights
3. **Lower the barrier to entry** — quickstart on the website guides new users without requiring them to navigate the repo first
4. **Establish credibility** — a polished website signals a serious, maintained project and builds trust for something as sensitive as email infrastructure

## User-Visible Outcome

### When this milestone is complete, the user can:

- Visit darkpipe.org and understand what DarkPipe is, why it exists, and how it works within 30 seconds
- Browse feature details with visual explanations of the architecture (cloud relay ↔ WireGuard/mTLS ↔ home device)
- Follow a quickstart guide directly on the website to begin a deployment
- Find answers to common questions (Podman support, port 25 requirements, mail server choices, etc.) in the FAQ
- Navigate a fast, accessible, mobile-responsive site that loads in under 2 seconds

### Entry point / environment

- Entry point: https://darkpipe.org
- Environment: Browser (desktop + mobile), deployed via Cloudflare Pages
- Live dependencies involved: Cloudflare Pages, Cloudflare DNS, GitHub (deploy source)

## Completion Class

- Contract complete means: all pages render, Lighthouse scores ≥90 (performance, accessibility, best practices, SEO), all links resolve, HTML validates
- Integration complete means: site deploys to Cloudflare Pages from GitHub, custom domain resolves, HTTPS works
- Operational complete means: none (static site, no runtime operations)

## Final Integrated Acceptance

To call this milestone complete, we must prove:

- darkpipe.org loads in a browser with all 4 pages rendering correctly
- Lighthouse audit scores ≥90 across all 4 categories
- Mobile viewport is responsive and usable
- Quickstart content matches current repo documentation
- FAQ answers are accurate and match project state post-M003

## Risks and Unknowns

- **Design quality** — the site must feel distinctive and trustworthy, not like generic AI-generated template; the frontend-design skill mitigates this
- **Cloudflare Pages DNS configuration** — domain setup may require specific DNS records; straightforward but needs verification
- **Content accuracy** — quickstart and FAQ must align with current project state (post-M003 runtime compatibility)
- **Architecture diagram** — visual explanation of the relay/transport/home stack needs to be clear and accurate

## Existing Codebase / Prior Art

- `docs/quickstart.md` — existing quickstart guide (source of truth for quickstart page content)
- `docs/faq.md` — existing FAQ (source of truth for FAQ page content)
- `docs/configuration.md` — feature details and configuration options
- `deploy/platform-guides/` — platform-specific deployment guides (linked from quickstart)
- `README.md` — project overview and feature list

> See `.gsd/DECISIONS.md` for all architectural and pattern decisions.

## Relevant Requirements

- No formal requirements — this is a new initiative. Website requirements are defined within this milestone.

## Scope

### In Scope

- Astro project setup with Tailwind CSS
- Hero landing page with value proposition, architecture overview, and CTAs
- Features page with detailed breakdowns of DarkPipe capabilities
- Quickstart page adapted from docs/quickstart.md
- FAQ page adapted from docs/faq.md
- Responsive design (mobile, tablet, desktop)
- Dark theme aligned with privacy/sovereignty brand
- Cloudflare Pages deployment configuration
- SEO basics (meta tags, Open Graph, structured data)
- Accessibility (WCAG AA)
- Architecture diagram/illustration

### Out of Scope / Non-Goals

- Full documentation portal (docs stay in repo for now)
- Blog (future milestone)
- User accounts, authentication, or any dynamic backend
- Custom CMS or content management
- Internationalization / multi-language
- Analytics beyond Cloudflare's built-in
- Email signup / newsletter
- Community forum integration

## Technical Constraints

- Astro with static output (no SSR)
- Tailwind CSS for styling
- Deploys to Cloudflare Pages
- Must work without JavaScript for core content (progressive enhancement)
- Images optimized (WebP/AVIF with fallbacks)
- No third-party tracking scripts or cookies

## Integration Points

- **Cloudflare Pages** — build and deploy from GitHub repo
- **Cloudflare DNS** — darkpipe.org domain configuration
- **GitHub** — source repository, deploy trigger on push to main
- **Existing docs** — quickstart and FAQ content sourced from repo docs

## Open Questions

- Should the website live in the main darkpipe repo or a separate repo? — Current thinking: separate `website/` directory in the main repo or a `darkpipe-website` repo. Separate repo keeps concerns clean and deploy triggers independent.
- Architecture diagram format — SVG illustration, animated diagram, or simple styled boxes? — Current thinking: SVG with CSS animations for the relay→tunnel→home flow.
