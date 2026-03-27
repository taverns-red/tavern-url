# 🏠 Tavern URL

A free, open-source, privacy-first URL shortener for nonprofits and service organizations.

## Vision

**Extreme simplicity combined with deep control.** Tavern URL gives nonprofit and service organizations the power of enterprise link management — custom slugs, aggregate analytics, QR codes, team management — without surveillance, complexity, or cost.

## Features (Planned — v1)

- **Custom & auto-generated short links** — branded slugs or 6-char Base62 codes
- **Privacy-first analytics** — aggregate clicks, geo (country), device, referrer domain. No cookies, no PII.
- **QR code generation** — PNG/SVG download with customizable colors
- **Multi-user & teams** — orgs, roles (Owner/Admin/Member), API keys
- **Third-party auth** — Google OAuth 2.0 alongside email/password
- **Beautiful web UI** — responsive, accessible, fast

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go |
| Frontend | HTMX + Templ |
| Database | PostgreSQL |
| Auth | OAuth 2.0 (Google) + bcrypt |
| CSS | Vanilla CSS (custom properties) |

## Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.22+
- [Templ](https://templ.guide/quick-start/installation) CLI
- [PostgreSQL](https://www.postgresql.org/download/) 15+
- [Docker](https://docs.docker.com/get-docker/) (optional, for containerized dev)

### Local Development

```bash
# Clone
git clone https://github.com/taverns-red/tavern-url.git
cd tavern-url

# Start Postgres (via Docker)
docker compose up -d db

# Copy env file and configure
cp .env.example .env

# Run migrations
make migrate

# Generate Templ files and run
make dev
```

### Docker

```bash
docker compose up
```

## Project Structure

```
tavern-url/
├── cmd/tavern/          # Application entrypoint
├── internal/
│   ├── handler/         # HTTP handlers
│   ├── service/         # Business logic
│   ├── repository/      # Database access
│   ├── model/           # Domain types
│   ├── auth/            # Authentication
│   └── middleware/       # Rate limiting, logging, etc.
├── templates/           # Templ components
├── static/              # CSS, images
├── migrations/          # SQL migrations
├── tasks/               # Engineering process docs
└── docs/                # Product & architecture docs
```

## Contributing

We welcome contributions! This is a nonprofit project — every PR helps organizations that serve their communities.

1. Fork the repo
2. Create a feature branch (`git checkout -b feat/my-feature`)
3. Follow the [engineering workflow](docs/product-spec.md) — TDD, small commits, issue-driven
4. Open a PR referencing the issue number

## License

[MIT](LICENSE) — free to use, modify, and distribute.

## About

Built by [Taverns Red](https://github.com/taverns-red) — a not-for-profit offering free and low-cost technology for service organizations.
