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
| Simple, beautiful web UI with dark mode | Require complex infrastructure |
| Multi-user with team management | Charge for advanced features |
| Aggregate click analytics | Store PII or use cookies |
| QR codes for print campaigns | Require database expertise to run |
| API keys for automation | Lock features behind paywalls |

---

## Shipped Features

### v1 (Sprints 1–11)

#### Core Link Management
- **Create short URLs** with auto-generated Base62 slugs (6 chars → ~56B combinations)
- **Custom slugs** — user-specified vanity URLs (e.g., `/spring-gala`), 3–64 chars
- **Edit links** — `PUT /api/v1/links/{id}` to change destination URL or slug
- **Delete links** — with HTMX partial swap on dashboard
- **302 redirects** for all short links
- **Slug collision retry** — 5 retries on auto-generated slugs

> Implemented in: `service/link_service.go`, `service/slug.go`, `handler/link_handler.go`

#### Privacy-First Analytics
- **Aggregate click counts** — total clicks per link, per day
- **Geo (country-level)** — derived from `CF-IPCountry` header or `X-Forwarded-For`, IP discarded immediately
- **Device category** — mobile / desktop / tablet / bot (from User-Agent, not stored raw)
- **Referrer domain** — grouped by domain only, full URL never stored
- **No cookies, no fingerprinting, no PII storage**
- **CSV export** — `GET /api/v1/links/{id}/analytics/export`

> Implemented in: `service/analytics_service.go`, `handler/analytics_handler.go`, `handler/export_handler.go`

#### QR Code Generation
- Auto-generated QR code for every short link
- Downloadable as PNG at configurable size
- `GET /api/v1/links/{id}/qr?size=512`

> Implemented in: `service/qr_service.go`

#### Multi-User & Teams
- **Registration** with email + password (bcrypt cost 12)
- **Google OAuth 2.0** with auto-account creation
- **Organizations** — group users into orgs; links scoped to org
- **Roles**: Owner, Admin, Member
- **Invite members** by email

> Implemented in: `auth/`, `service/org_service.go`, `handler/org_handler.go`

#### API Keys
- `tvn_<hex>` format, SHA-256 hashed in storage
- Dual auth middleware — session cookie OR `Authorization: Bearer tvn_...`
- Create, list, delete via API

> Implemented in: `service/apikey_service.go`, `handler/apikey_handler.go`

#### Web UI
- **Server-rendered** with Templ + HTMX (zero JS frameworks)
- Dashboard with link list
- Link detail page with analytics charts (bar chart, country/device/referrer breakdowns)
- Login and register forms
- Responsive design

> Implemented in: `templates/*.templ`, `handler/page_handler.go`

#### Infrastructure
- Docker multi-stage build (~30MB image)
- CI pipeline (GitHub Actions) — vet, test, build, Docker push to GHCR
- Release Please for automated versioning
- Health check at `/health`
- Rate limiting (token bucket, 60 req/min per IP)
- Security headers (CSP, HSTS, X-Frame-Options)

---

### v2 (Sprints 22–31)

#### Dark Mode
- CSS custom properties design system
- `prefers-color-scheme` auto-detection + manual `[data-theme="dark"]` toggle
- Theme persistence via `localStorage` in `theme.js`

#### Dashboard Search
- `?q=` query parameter on `GET /api/v1/links` to filter links by URL or slug

#### Bulk Link Creation
- `POST /api/v1/links/bulk` — create 1–100 links in a single request

#### Email Infrastructure
- `email.Sender` interface with SMTP and Noop implementations
- Password reset tokens (migration 009)

#### Custom Domains
- `DomainService` for DNS TXT record verification
- `tvn-verify-<token>` TXT records for domain ownership proof
- Migration 008 for custom domain storage

#### Password-Protected Links
- Optional `password_hash` on links (migration 010)
- Visitors prompted for password before redirect

#### Dynamic Redirects
- Redirect rules engine (geo-based, device-based, weighted A/B testing)
- Migration 011 for `redirect_rules` table
- `RedirectService` evaluates rules at redirect time

#### Browser Extension
- Chrome MV3 extension in `/extension/`
- One-click URL shortening from any page
- API key authentication via `chrome.storage`

#### Webhooks
- `POST` delivery on link events with HMAC-SHA256 signatures
- Migration 012 for `webhooks` table
- `WebhookService` for registration and delivery

#### Self-Serve Onboarding
- NFP application form
- Migration 013 for `applications` table + `is_admin` flag
- Admin review workflow

---

## Backlog

| Feature | Priority | Notes |
|---|---|---|
| Link expiration / max-click limits | High | Schema exists (migration 007) but no UI |
| CSV export UI button | Medium | API exists but no dashboard integration |
| Slack integration | Low | API keys already enable programmatic access |
| Raycast / Alfred plugin | Low | Power-user convenience |

---

## Technical Decisions

### Stack

| Layer | Choice | Rationale |
|---|---|---|
| **Language** | **Go 1.25** | Single binary, fast compile, excellent HTTP stdlib |
| **HTTP Router** | `chi/v5` | Lightweight, idiomatic Go, middleware-friendly |
| **Frontend** | **HTMX 2.0 + Templ** | Server-rendered HTML, no Node build step |
| **Database** | **PostgreSQL** (`pgx/v5`) | Production-grade, 13 migrations |
| **Auth** | bcrypt + Google OAuth 2.0 | Session cookies + API keys |
| **QR** | `skip2/go-qrcode` | Pure Go, no CGo |
| **CSS** | Vanilla CSS with custom properties | Dark mode, no build step |
| **CI** | GitHub Actions | `go-version-file: 'go.mod'` for version sync |
| **Deploy** | Docker + Fly.io | ~30MB Alpine image |

### Deployment Recommendation

> [!TIP]
> **Recommended: Fly.io** (~$5/mo) **+ Neon Postgres** (free tier). Total cost: ~$5/mo or less.

---

## Privacy Policy (Technical)

1. **No cookies** for analytics. Session cookie only for authenticated users.
2. **IP addresses** used ephemerally for country-level geo, then discarded. Never stored.
3. **User-Agent** parsed to device category only. Never stored raw.
4. **Referrer** truncated to domain only. Full URL never stored.
5. **No third-party tracking scripts.**
6. **GDPR-friendly by design** — no personal data collected from link visitors.

---

## Business Rules

| Rule | Why | Where |
|---|---|---|
| Slugs: alphanumeric + hyphens, case-sensitive | URL-safe, avoid confusion | `service/link_service.go` |
| Custom slugs: 3–64 chars | Prevent squatting | `service/link_service.go` |
| Auto slugs: 6 chars Base62 | 56B combinations | `service/slug.go` |
| Rate limit: 60 creates/min per IP | Prevent abuse | `middleware/rate.go` |
| Rate limit: none on redirects | Must be fast | — |
| API key format: `tvn_<hex>` | Identifiable in logs/scans | `service/apikey_service.go` |
| API key storage: SHA-256 hash | DB breach safe | `service/apikey_service.go` |
| Email normalization: lowercase + trim | Service layer | `auth/service.go` |
| OAuth users: sentinel hash `oauth:google` | Not valid bcrypt | `auth/service.go` |
| Bulk create: 1–100 URLs max | Prevent abuse | `handler/link_handler.go` |

---

## Non-Functional Requirements

| Requirement | Target |
|---|---|
| Redirect latency (p99) | < 50ms |
| Availability | 99.9% |
| Docker image size | < 30MB |
| Startup time | < 2 seconds |
| Zero-downtime deploys | Blue-green via Fly.io |

---

## API Reference

### Auth
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/api/v1/auth/register` | — | Create account |
| `POST` | `/api/v1/auth/login` | — | Log in |
| `POST` | `/api/v1/auth/logout` | Session | Log out |
| `GET` | `/api/v1/auth/me` | Session/Key | Current user |
| `GET` | `/api/v1/auth/google/login` | — | Google OAuth |
| `GET` | `/api/v1/auth/google/callback` | — | OAuth callback |

### Links
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/api/v1/links` | Session/Key | Create link |
| `POST` | `/api/v1/links/bulk` | Session/Key | Bulk create (1–100) |
| `GET` | `/api/v1/links` | Session/Key | List links (`?q=` search) |
| `PUT` | `/api/v1/links/{id}` | Session/Key | Edit link |
| `DELETE` | `/api/v1/links/{id}` | Session/Key | Delete link |
| `GET` | `/api/v1/links/{id}/analytics` | Session/Key | Get analytics |
| `GET` | `/api/v1/links/{id}/analytics/export` | Session/Key | CSV export |
| `GET` | `/api/v1/links/{id}/qr` | Session/Key | QR code PNG |

### Organizations
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/api/v1/orgs` | Session/Key | Create org |
| `GET` | `/api/v1/orgs` | Session/Key | List orgs |
| `POST` | `/api/v1/orgs/{slug}/invite` | Session/Key | Invite member |
| `PUT` | `/api/v1/orgs/{slug}/members/{id}/role` | Session/Key | Change role |

### API Keys
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/api/v1/keys` | Session | Create key |
| `GET` | `/api/v1/keys` | Session | List keys |
| `DELETE` | `/api/v1/keys/{id}` | Session | Delete key |

### Other
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/health` | — | Health check |
| `GET` | `/{slug}` | — | Redirect |
