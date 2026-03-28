# Tavern URL

> A free, open-source, privacy-first URL shortener built for nonprofits and service organizations.

[![CI](https://github.com/taverns-red/tavern-url/actions/workflows/ci.yml/badge.svg)](https://github.com/taverns-red/tavern-url/actions/workflows/ci.yml)

## Features

- **Short links** with auto-generated Base62 slugs or custom vanity URLs
- **Privacy-first analytics** — country, device, referrer (no PII, no cookies, no fingerprinting)
- **QR code generation** — downloadable PNG at any size
- **Multi-user** with organizations, roles (owner/admin/member), and invites
- **Dual authentication** — session cookies and API keys (`tvn_` prefix, SHA-256 hashed)
- **Dark mode** — auto-detects system preference with manual toggle
- **Bulk link creation** — paste up to 100 URLs at once
- **Dashboard search** — filter links by URL or slug
- **Link editing** — update destination URL or slug
- **Browser extension** — Chrome MV3 for one-click shortening
- **CSV export** — download analytics for board reports
- **Custom domains** — DNS TXT verification for branded links
- **Webhooks** — HMAC-SHA256 signed event delivery
- **Rate limiting** — configurable per-IP token bucket (60 req/min default)
- **Server-rendered UI** — HTMX + Templ templates, zero JavaScript frameworks

## Tech Stack

| Layer | Technology |
|-------|------------|
| Language | Go 1.25 |
| Router | chi/v5 |
| Database | PostgreSQL (pgx/v5) |
| Frontend | Templ + HTMX 2.0 |
| QR Codes | skip2/go-qrcode |
| Auth | bcrypt + Google OAuth 2.0 |
| CSS | Vanilla CSS with custom properties (dark mode) |
| CI | GitHub Actions |
| Deploy | Docker + Fly.io |

## Quickstart

### Prerequisites

- Go 1.25+
- Docker (for PostgreSQL)

### Local Development

```bash
# Clone
git clone https://github.com/taverns-red/tavern-url.git
cd tavern-url

# Start PostgreSQL only
docker compose up -d db

# Set environment variables
export DATABASE_URL="postgres://tavern:tavern_dev@localhost:5432/tavern?sslmode=disable"
export PORT=8080
export BASE_URL="http://localhost:8080"
export SESSION_SECRET="change-me-to-something-random-32-chars"

# Run the server (applies migrations automatically)
go run ./cmd/tavern
# → http://localhost:8080
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | (required) | PostgreSQL connection string |
| `PORT` | `8080` | HTTP server port |
| `BASE_URL` | `http://localhost:8080` | Public URL for short links |
| `SESSION_SECRET` | (required) | 32+ char secret for session cookies |
| `GOOGLE_CLIENT_ID` | (optional) | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | (optional) | Google OAuth client secret |

## API

All endpoints return JSON. Authentication via session cookie or `Authorization: Bearer tvn_...` header.

### Links

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/links` | Create a short link |
| `POST` | `/api/v1/links/bulk` | Bulk create (1–100 URLs) |
| `GET` | `/api/v1/links` | List all links (`?q=` search) |
| `PUT` | `/api/v1/links/{id}` | Edit a link |
| `DELETE` | `/api/v1/links/{id}` | Delete a link |
| `GET` | `/api/v1/links/{id}/analytics` | Get link analytics |
| `GET` | `/api/v1/links/{id}/analytics/export` | CSV export |
| `GET` | `/api/v1/links/{id}/qr` | Generate QR code (`?size=512`) |

### Auth

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/auth/register` | Create account |
| `POST` | `/api/v1/auth/login` | Log in |
| `POST` | `/api/v1/auth/logout` | Log out |
| `GET` | `/api/v1/auth/me` | Current user |

### API Keys

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/keys` | Create an API key |
| `GET` | `/api/v1/keys` | List your API keys |
| `DELETE` | `/api/v1/keys/{id}` | Delete an API key |

### Organizations

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/orgs` | Create an org |
| `GET` | `/api/v1/orgs` | List your orgs |
| `POST` | `/api/v1/orgs/{slug}/invite` | Invite a member |
| `PUT` | `/api/v1/orgs/{slug}/members/{id}/role` | Change member role |

## Browser Extension

A Chrome MV3 extension is included in `/extension/`. To install:

1. Open `chrome://extensions`
2. Enable "Developer mode"
3. Click "Load unpacked" and select the `extension/` directory
4. Click the extension icon on any page to shorten the current URL

Requires an API key — generate one at `/dashboard` or via `POST /api/v1/keys`.

## Deploy to Fly.io

```bash
fly launch
fly secrets set DATABASE_URL="..." SESSION_SECRET="..." GOOGLE_CLIENT_ID="..." GOOGLE_CLIENT_SECRET="..."
fly deploy
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE).
