# Architecture Overview

## System Architecture

Tavern URL follows a **server-rendered monolith** architecture. A single Go binary serves both the API and the web UI.

```
┌──────────────────────────────────────────────────┐
│                   Cloudflare CDN                  │
│              (free tier, DDoS + cache)            │
└──────────────────────┬───────────────────────────┘
                       │
┌──────────────────────▼───────────────────────────┐
│                Go HTTP Server                     │
│  ┌─────────┐ ┌──────────┐ ┌───────────────────┐ │
│  │Redirect │ │ Web UI   │ │ REST API          │ │
│  │Handler  │ │ (HTMX +  │ │ /api/v1/...       │ │
│  │GET /:id │ │  Templ)  │ │                   │ │
│  └────┬────┘ └────┬─────┘ └────────┬──────────┘ │
│       │           │                │             │
│  ┌────▼───────────▼────────────────▼──────────┐  │
│  │              Service Layer                  │  │
│  │  LinkService · AuthService · OrgService     │  │
│  │  AnalyticsService · QRService               │  │
│  └────────────────────┬───────────────────────┘  │
│                       │                          │
│  ┌────────────────────▼───────────────────────┐  │
│  │            Repository Layer                 │  │
│  │         (PostgreSQL via pgx)               │  │
│  └────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────┘
```

## Key Design Decisions

- **Server-rendered HTML** — HTMX handles interactivity via HTML-over-the-wire. No separate frontend build.
- **Layered architecture** — Handler → Service → Repository. Dependencies flow inward.
- **Interfaces at boundaries** — Services depend on repository interfaces, enabling easy testing.
- **Privacy by design** — PII (IP, User-Agent) is processed ephemerally and never persisted.

## Module Dependency Rules

```
handler → service → repository → model
   ↓         ↓
  auth    middleware
```

- `model/` has zero dependencies (pure types)
- `repository/` depends only on `model/`
- `service/` depends on repository interfaces + `model/`
- `handler/` depends on `service/` + `auth/` + `middleware/`
- No circular dependencies allowed

## ADRs

Architecture Decision Records are stored in `docs/architecture-decisions/`.
