# Tavern URL — Product Specification

> A free, open-source URL shortener built for nonprofits and service organizations.

## Product Vision

**Extreme simplicity combined with deep control.** Tavern URL provides nonprofit and service organizations with a powerful, privacy-first URL shortener that is free to use, easy to operate, and gives operators full control over their data and configuration.

### Who We Serve

- **Primary persona: "The NFP Comms Lead"** — A nonprofit communications director who needs branded short links for newsletters, social media, and fundraising campaigns. They want analytics without compromising donor/visitor privacy. They are not deeply technical but can follow simple setup instructions.
- **Secondary persona: "The NFP Tech Volunteer"** — A technical volunteer or small IT team at a nonprofit who deploys and maintains the instance. They value a simple, single-binary deployment with minimal ops burden.

### Core Value Proposition

| What we do | What we don't do |
|---|---|
| Free, privacy-first link shortening | Track individual visitors |
| Simple, beautiful web UI | Require complex infrastructure |
| Multi-user with team management | Charge for advanced features |
| Aggregate click analytics | Store PII or use cookies |
| QR codes for print campaigns | Require database expertise to run |
| API keys for automation | Lock features behind paywalls |

---

## Shipped Features

*None yet — project inception.*

---

## Planned: v1 (MVP)

### Core Link Management

- **Create short URLs** with auto-generated Base62 slugs (6 chars → ~56B combinations)
- **Custom slugs** — user-specified vanity URLs (e.g., `/spring-gala`)
- **Bulk create** — paste multiple URLs to shorten at once
- **Edit / delete** existing short links
- **302 redirects** (to enable aggregate analytics; configurable to 301)

### Privacy-First Analytics

- **Aggregate click counts** — total clicks per link, per day/week/month
- **Geo (country-level)** — derived from IP at request time, IP discarded immediately
- **Device category** — mobile / desktop / tablet (from User-Agent, not stored raw)
- **Referrer domain** — grouped (e.g., "twitter.com", "direct"), not full referrer URL
- **No cookies, no fingerprinting, no PII storage**

### QR Code Generation

- Auto-generated QR code for every short link
- Downloadable as PNG / SVG
- Customizable foreground/background colors

### Multi-User & Teams

- **Registration** with email + password
- **Third-party auth** via Google (OAuth 2.0 / OIDC), extensible to other providers
- **Organizations** — group users into orgs; links belong to an org
- **Roles**: Owner, Admin, Member (view-only analytics, can create links)
- **API keys** — per-user, scoped to their org, for programmatic link creation

### Web UI

- Dashboard with link list, search, and filtering
- Link detail page with analytics charts
- Org settings (members, API keys)
- Responsive design (mobile-friendly for on-the-go campaign checks)

---

## Backlog (Post-MVP)

| Feature | Priority | Notes |
|---|---|---|
| Link expiration / max-click limits | High | Useful for time-limited campaigns |
| Custom domains | High | Branded links (e.g., `go.habitat.org/gala`) |
| Password-protected links | Medium | For internal/restricted content |
| Dynamic redirects (device/geo) | Medium | A/B testing for campaigns |
| Browser extension | Medium | Quick shortening from any page |
| Slack integration | Low | Shorten links in Slack messages |
| Raycast / Alfred plugin | Low | Power-user convenience |
| CSV export of analytics | Medium | For board reports |
| Webhook notifications | Low | Notify on link creation or click milestones |
| Self-serve onboarding for NFPs | High | Application form, verification, auto-provisioning |

---

## Technical Decisions

### Stack

| Layer | Choice | Rationale |
|---|---|---|
| **Language** | **Go** | Single binary, fast compile, excellent HTTP stdlib, lower barrier for NFP contributors than Rust. Meets "extreme simplicity" goal. |
| **HTTP Router** | `chi` or `echo` | Lightweight, idiomatic Go, middleware-friendly |
| **Frontend** | **HTMX + Templ** | Server-rendered HTML with interactive sprinkles. No Node build step, no JS framework. Stays in the Go ecosystem. Aligns with "lighter but powerful." |
| **Database** | **PostgreSQL** | Production-grade, excellent with Go (`pgx`), strong NFP cloud support |
| **Auth** | `oauth2` + session cookies | Google OIDC for third-party; bcrypt for local passwords |
| **QR** | `go-qrcode` or `skip2/go-qrcode` | Pure Go, no CGo dependencies |
| **Migrations** | `goose` or `atlas` | SQL-based, version-controlled schema |
| **CSS** | Vanilla CSS with custom properties | No build step, themeable, accessible |

### Why Go over Rust

Both are excellent. For this project, Go wins on:

1. **Contributor accessibility** — NFP tech volunteers are more likely to know Go or pick it up quickly
2. **Development velocity** — faster iteration from portfolio → production
3. **Ecosystem maturity** — richer HTTP/web middleware ecosystem
4. **Deployment simplicity** — single static binary, no runtime dependencies
5. **Good enough performance** — hundreds of links/day is trivially handled; Go's HTTP server can handle tens of thousands of concurrent connections

### Why HTMX + Templ over a JS Framework

1. **No build step** — `go build` produces everything, including embedded static assets
2. **Extreme simplicity** — HTML is the API; HTMX adds interactivity via attributes
3. **Type safety** — Templ compiles to Go; template errors caught at compile time
4. **Performance** — server-rendered HTML is faster than client-side hydration
5. **Accessibility** — progressive enhancement; works without JS for basic flows

### Deployment Recommendation

> [!TIP]
> **Recommended: Fly.io** (~$5/mo for a small instance) **+ Neon Postgres** (free tier: 0.5GB storage, 190 compute hours/mo). Total cost: ~$5/mo or less.

**NFP-friendly hosting options:**

| Provider | Program | Benefit |
|---|---|---|
| **Google Cloud** | Google for Nonprofits | Free GCP credits (substantial annual), free Workspace |
| **AWS** | AWS Nonprofit Credit Program | $1,000 in credits (12 months), Imagine Grant up to $100K credits |
| **Cloudflare** | Project Galileo / Startups | Up to $250K in credits; free DDoS/WAF for qualifying orgs |
| **Fly.io** | No NFP program | ~$5/mo minimum; simple Docker/binary deploy |
| **Render** | Free tier | Free for static sites; ~$7/mo for web services |

**Recommendation for production:**
- **Start with Fly.io + Neon** (simple, cheap, fast to deploy)
- **Apply for Google Cloud for Nonprofits** once the NFP entity is established (free and most generous long-term)
- **Cloudflare free tier** in front for CDN/DDoS protection regardless of backend host

### Monorepo Structure

```
tavern-url/
├── cmd/
│   └── tavern/          # main entrypoint
├── internal/
│   ├── handler/         # HTTP handlers
│   ├── service/         # business logic
│   ├── repository/      # database access
│   ├── model/           # domain types
│   ├── auth/            # authentication (local + OAuth)
│   └── middleware/      # rate limiting, logging, etc.
├── templates/           # Templ components
├── static/              # CSS, images, JS (minimal)
├── migrations/          # SQL migration files
├── docs/
│   └── product-spec.md  # this file
├── tasks/
│   └── lessons.md
├── Dockerfile
├── docker-compose.yml   # local dev (app + postgres)
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Privacy Policy (Technical)

1. **No cookies** for analytics. Session cookie only for authenticated users.
2. **IP addresses** are used ephemerally to derive country-level geo, then discarded. Never stored.
3. **User-Agent** is parsed to device category (mobile/desktop/tablet), then discarded. Never stored raw.
4. **Referrer** is truncated to domain only. Full URL never stored.
5. **No third-party tracking scripts** (no Google Analytics, no pixels).
6. **GDPR-friendly by design** — no personal data collected from link visitors.

---

## Business Rules

| Rule | Why | Where |
|---|---|---|
| Slugs are case-sensitive, alphanumeric + hyphens only | Avoid confusion, URL-safe | `service/link.go` |
| Custom slugs: 3–64 chars | Prevent squatting on 1-2 char slugs; reasonable max | `service/link.go` |
| Auto slugs: 6 chars Base62 | 56B combinations; short enough for print | `service/slug.go` |
| Rate limit: 60 creates/min per user | Prevent abuse | `middleware/rate.go` |
| Rate limit: none on redirects | Redirects must be fast and unrestricted | — |
| Org limit: 10,000 links per org (configurable) | Prevent runaway usage | `service/org.go` |
| API keys: 256-bit random, hashed with SHA-256 | Security best practice | `auth/apikey.go` |

---

## Non-Functional Requirements

| Requirement | Target |
|---|---|
| Redirect latency (p99) | < 50ms |
| Availability | 99.9% (single-instance Fly.io SLA) |
| Max concurrent connections | 1,000+ (Go stdlib handles this trivially) |
| Binary size | < 20MB |
| Docker image size | < 30MB (distroless or alpine) |
| Startup time | < 2 seconds |
| Zero-downtime deploys | Blue-green via Fly.io |

---

## Definition of Done (for v1)

- [ ] User can register, log in (email + Google OAuth), and manage their profile
- [ ] User can create, edit, delete short links with auto or custom slugs
- [ ] Short links redirect correctly (302)
- [ ] Dashboard shows link list with click counts
- [ ] Link detail page shows aggregate analytics (clicks over time, geo, device, referrer)
- [ ] QR codes generated and downloadable for each link
- [ ] Org management: create org, invite members, assign roles
- [ ] API key generation and link creation via API
- [ ] Privacy: no PII stored, no cookies on redirect
- [ ] Deployed and accessible on a public URL
- [ ] README with setup instructions
- [ ] All tests passing, CI green
