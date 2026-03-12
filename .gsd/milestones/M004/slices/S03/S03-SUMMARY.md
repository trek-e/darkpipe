---
slice: S03
milestone: M004
status: complete
started: 2026-03-12
completed: 2026-03-12
tasks_completed: 3
tasks_total: 3
---

# S03: Polish, Accessibility & Deployment — Summary

## What Was Built

- **SEO**: @astrojs/sitemap integration (4 URLs in sitemap-index.xml), robots.txt, JSON-LD structured data (SoftwareApplication schema) in base layout. Unique title and description per page.
- **Accessibility**: Visible focus rings (`:focus-visible` with pipe-green outline), `prefers-reduced-motion` support, ghost color brightened from #4a5568 to #6b7a90 (4.59:1 on void — AA pass). All other color pairs pass AA (mist 7.81:1, cloud 13.49:1, snow 18.29:1, pipe 10.44:1).
- **Deployment config**: Cloudflare Pages `_headers` (CSP, X-Frame-Options, security headers, cache immutable for /_astro/*, TTLs for static assets), `_redirects` placeholder, OG image SVG.

## Color Contrast Audit

| Pair | Ratio | AA Normal | AA Large |
|------|-------|-----------|----------|
| cloud (#cbd5e1) on void (#07080a) | 13.49:1 | PASS | PASS |
| mist (#94a3b8) on void | 7.81:1 | PASS | PASS |
| ghost (#6b7a90) on void | 4.59:1 | PASS | PASS |
| snow (#f1f5f9) on void | 18.29:1 | PASS | PASS |
| pipe (#22d3a7) on void | 10.44:1 | PASS | PASS |

## Cloudflare Pages Deployment

To deploy, connect the GitHub repo to Cloudflare Pages with these build settings:
- **Build command**: `cd website && npm install && npm run build`
- **Build output directory**: `website/dist`
- **Root directory**: `/` (or `website/` if Cloudflare supports it)
- **Node.js version**: 20+

Then point `darkpipe.org` DNS to the Cloudflare Pages project.

## What Remains

- Connect GitHub repo to Cloudflare Pages (user action — requires Cloudflare account)
- Configure darkpipe.org DNS CNAME to Cloudflare Pages (user action)
- Replace OG image SVG with rasterized PNG (social platforms prefer PNG)
- Lighthouse audit on deployed site (requires live URL)
