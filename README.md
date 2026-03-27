# Tavern URL

> A free, open-source, privacy-first URL shortener built for nonprofits and service organizations.

[![CI](https://github.com/taverns-red/tavern-url/actions/workflows/ci.yml/badge.svg)](https://github.com/taverns-red/tavern-url/actions/workflows/ci.yml)

## Features

- **Short links** with auto-generated Base62 slugs or custom vanity URLs
- **Privacy-first analytics** — country, device, referrer (no PII, no cookies, no fingerprinting)
- **QR code generation** — PNG with custom foreground/background colors
- **Multi-user** with organizations, roles (owner/admin/member), and invites
- **Dual authentication** — session cookies and API keys (`tvn_` prefix, SHA-256 hashed)
- **Rate limiting** — configurable per-IP token bucket (60 req/min default)
- **Server-rendered UI** — HTMX + Templ templates, zero JavaScript frameworks

## Tech Stack

| Layer | Technology |
|-------|------------|
| Language | Go 1.23 |
| Router | chi/v5 |
| Database | PostgreSQL (pgx/v5) |
| Frontend | Templ + HTMX |
| QR Codes | skip2/go-qrcode |
| Auth | bcrypt + Google OAuth 2.0 |
| Deploy | Docker + Fly.io |

## Quickstart

```bash
# Clone
git clone https://github.com/taverns-red/tavern-url.git
cd tavern-url

# Start PostgreSQL
docker compose up -d

# Run migrations
goose -dir migrations postgres "postgres://tavern:tavern_dev@localhost:5432/tavern?sslmode=disable" up

# Start the server
make dev
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
| `GET` | `/api/v1/links` | List all links |
| `DELETE` | `/api/v1/links/{id}` | Delete a link |
| `GET` | `/api/v1/links/{id}/analytics` | Get link analytics |
| `GET` | `/api/v1/links/{id}/qr` | Generate QR code (PNG) |

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
