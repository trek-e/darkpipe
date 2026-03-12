# S03: Polish, Accessibility & Deployment

**Goal:** Site passes HTML validation, has complete SEO meta, is responsive across breakpoints, accessible (semantic HTML, focus states, color contrast), and has Cloudflare Pages deployment config ready. Deployed to Cloudflare Pages at darkpipe.org.
**Demo:** Site builds clean, HTML validates, responsive at 375/768/1280px, Cloudflare Pages deploys from repo.

## Must-Haves

- HTML validation (no errors)
- SEO: sitemap, robots.txt, structured data
- Responsive verified at 375px, 768px, 1280px
- Focus states visible on all interactive elements
- Color contrast meets WCAG AA (4.5:1 text, 3:1 UI)
- Cloudflare Pages deployment configuration
- OG image placeholder

## Proof Level

- This slice proves: integration (deployment to Cloudflare Pages)
- Real runtime required: yes (Cloudflare Pages)
- Human/UAT required: yes (domain DNS, visual review)

## Verification

- `cd website && npm run build` — zero warnings beyond Vite internal
- HTML output validates (no unclosed tags, proper nesting)
- sitemap.xml and robots.txt present in build output
- All pages have unique title and description meta
- Cloudflare Pages config present (wrangler.toml or build settings)

## Observability / Diagnostics

- Runtime signals: Build output, Cloudflare Pages deploy logs
- Inspection surfaces: Built HTML in dist/, Cloudflare Pages dashboard
- Failure visibility: Build errors, deploy errors
- Redaction constraints: Cloudflare API token (if used)

## Integration Closure

- Upstream surfaces consumed: S01 layout, S02 pages, all components
- New wiring introduced: sitemap integration, robots.txt, Cloudflare Pages config
- What remains: DNS configuration for darkpipe.org (user action)

## Tasks

- [x] **T01: SEO, sitemap, robots.txt, and structured data** `est:15m`
  - Why: Search engines need sitemap and robots.txt; structured data improves rich results
  - Files: `website/astro.config.mjs`, `website/public/robots.txt`, `website/src/layouts/Base.astro`
  - Do: Add @astrojs/sitemap integration, create robots.txt, add JSON-LD structured data to Base layout. Verify unique titles/descriptions on all pages.
  - Verify: sitemap.xml in build output, robots.txt serves, structured data in HTML
  - Done when: Build outputs sitemap.xml, robots.txt exists, all pages have unique meta

- [x] **T02: Accessibility and responsive polish** `est:20m`
  - Why: WCAG AA compliance and responsive usability are milestone success criteria
  - Files: `website/src/styles/global.css`, various components
  - Do: Add visible focus ring styles, verify heading hierarchy on all pages, check color contrast of key text/bg combinations, verify responsive layout at 375px via HTML structure analysis. Fix any issues found.
  - Verify: Focus styles visible, heading hierarchy correct (already verified), contrast ratios pass
  - Done when: All interactive elements have visible focus rings, no contrast failures in key combinations

- [x] **T03: Cloudflare Pages deployment config** `est:10m`
  - Why: Site needs to deploy to Cloudflare Pages from the GitHub repo
  - Files: `website/wrangler.toml`, `website/public/_headers`, `website/public/_redirects`
  - Do: Create Cloudflare Pages config with build command, output dir, security headers (CSP, X-Frame-Options, etc.), and cache headers for static assets. Create _redirects for any needed redirects.
  - Verify: Build still succeeds, _headers and _redirects in dist output
  - Done when: Cloudflare Pages config ready for deployment

## Files Likely Touched

- `website/astro.config.mjs`
- `website/public/robots.txt`
- `website/public/_headers`
- `website/public/_redirects`
- `website/src/layouts/Base.astro`
- `website/src/styles/global.css`
