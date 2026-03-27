# ADR-001: Go + HTMX + Templ Stack

**Status**: Accepted
**Date**: 2026-03-26
**Context**: Choosing a tech stack for a privacy-first URL shortener targeting NFP organizations. Need server-rendered web UI, minimal ops burden, and accessibility for NFP tech volunteers.
**Decision**: Use Go for the backend, HTMX for frontend interactivity, and Templ for type-safe HTML templating. PostgreSQL for persistence.
**Consequences**:
- ✅ Single binary deployment, no Node/JS build step
- ✅ Compile-time template safety via Templ
- ✅ Low contributor barrier (Go is widely known)
- ✅ Progressive enhancement (works without JS)
- ⚠️ Smaller ecosystem for pre-built UI components vs React
- ⚠️ Tighter frontend-backend coupling than an SPA
**Alternatives Considered**:
- Rust + Askama: Better perf ceiling, but slower dev velocity and higher contributor barrier
- Node + Svelte: Rich component ecosystem, but adds Node runtime and build complexity
- Node + React: Most popular, but heaviest; contradicts "extreme simplicity" goal
