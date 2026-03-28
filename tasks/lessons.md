# Tavern URL — Lessons Learned

## Sprint 1: Vertical Slice
- Base62 slug generation needs collision retry logic — 5 retries is sufficient for 56B combinations.
- `curl -I` sends HEAD, not GET — chi's `r.Get()` correctly rejects it (405). Not a bug.
- Putting `/{slug}` route last in chi avoids catching API and page routes.

## Sprint 2: Authentication
- bcrypt at cost 12 takes ~250ms — acceptable for auth but explains slow test times.
- Email normalization (lowercase + trim) should happen at the service layer, not repository.
- `json:"-"` on PasswordHash prevents accidental serialization — defense in depth.

## Sprint 3: Orgs + OAuth
- Transactional org creation (org + owner membership in one tx) prevents orphaned orgs.
- Google OAuth users need a sentinel password hash ("oauth:google") — not a valid bcrypt hash.
- Slug validation should lowercase input before pattern matching — uppercase slugs are valid after normalization.

## Sprint 4: UI Foundation
- `embed.FS` can only embed from subdirectories — can't use `../../static` from handler package.
- Serving static files from disk in dev, embed at production build stage is the practical approach.
- HTMX CDN is 14KB gzipped — acceptable for a server-rendered app, no npm needed.

## Sprint 5: Link Management UI
- Expanding a repository interface requires updating all mocks — interface-first design has this cost.
- HTMX `hx-target` + `hx-swap` enables surgical DOM updates without full page reloads.
- When expanding an interface, update mocks in ALL test files (service + handler) simultaneously.

## Sprints 6-8: Analytics + QR
- Fire-and-forget goroutines for click recording keep redirect latency <6ms.
- Cloudflare's CF-IPCountry header is the cheapest way to get country-level geo data.
- `skip2/go-qrcode` produces PNG directly — no SVG support, but PNG covers most use cases.
- Device detection from User-Agent is intentionally simple (mobile/desktop/tablet/bot) to match privacy goals.

## Sprints 9-11: API Keys + Deploy
- API key format `tvn_<hex>` makes keys easily identifiable (e.g., in logs, git scans).
- SHA-256 hash storage means even a DB breach doesn't expose raw API keys.
- Multi-stage Docker builds produce ~30MB images — Alpine is king for Go binaries.
- Fly.io health checks at `/health` enable zero-downtime deploys with auto-start/stop.

## Sprints 12-21: MVP Completion + Post-MVP
- Dual-auth middleware (session OR API key) keeps handler code auth-agnostic.
- Token bucket rate limiting is simpler than sliding window and sufficient for most use cases.
- `embed.FS` can only embed from same package directory — use `fs.FS` interface to decouple.
- In-memory rate limiters need periodic cleanup goroutines to prevent memory leaks.

## Sprints 22-31: Growth + Platform
- CSS custom properties make dark mode trivial — override variables, not selectors.
- `prefers-color-scheme` + `[data-theme]` gives both auto-detect and manual toggle.
- When changing a service method signature (e.g., `UpdateLink`), update all callers AND mocks simultaneously.
- Browser extensions need `chrome.storage` not `localStorage` for data persistence.
- HMAC-SHA256 webhook signatures prevent replay attacks — include event type in headers.
- DNS TXT verification via `net.LookupTXT` works without any external dependencies.
- `containsCI` (case-insensitive search) is better than adding DB-level ILIKE for small datasets.

## Sprint 32: Documentation & Stability
- Product spec must be updated as features ship — "Shipped: None" after 31 sprints is a red flag.
- HTMX form error display requires HTML responses, not JSON — use `writeFormError` for form submissions.
- `strings.HasPrefix` for Content-Type detection handles charset suffixes (`; charset=utf-8`).
- External JS with `data-*` attributes is CSP-safe and eliminates all inline `onclick` handlers.
- Go version must be consistent across `go.mod`, `ci.yml`, `Dockerfile`, and `README.md`.

## Sprint 33: UI Feature Parity
- HTMX `hx-trigger="input changed delay:300ms"` gives debounced search without custom JS.
- `hx-push-url="true"` keeps search queries in browser history — shareable/bookmarkable.
- Edit modals need `htmx.process(form)` after dynamically setting `hx-put` via JS.
- Templ components for HTMX partials (LinkList, APIKeyList) eliminate inline HTML in handlers.
- When handler returns HTML for HTMX, don't set Content-Type — templ sets it automatically.

## Sprint 34-37 (Phase 1: Hardening)

- **Service signature sprawl**: Adding optional params to `CreateLink` (expiresAt, maxClicks, password) required updating all callers including tests. Consider an options struct pattern (`CreateLinkOpts`) for services with >4 params to avoid cascading changes.
- **Templ generate timing**: The LSP shows `undefined: templates.X` errors until `templ generate` runs. This is expected — always run `templ generate` before `go vet`/`go test`.
- **Form vs JSON dual paths**: Each handler that supports both HTMX forms and JSON needs `isForm` branching. Extract a `parseRequest` helper to reduce boilerplate.
- **Password gate POST route**: The `/{slug}` catch-all needed both GET and POST for the password gate. Chi requires explicit `r.Post` alongside `r.Get` — `r.HandleFunc` would also work but is less explicit.
- **Modal proliferation**: Each new feature adds a modal → backdrop click + auto-close lists grow. Consider a generic modal manager in JS rather than enumerating IDs.

## Sprints 38-63 (Phases 2-6)

- **Batch template creation**: Creating multiple `.templ` files in one go is efficient — templ generate processes all at once. Group related templates (e.g., integrations.templ combines CLI + browser ext + Zapier + embed widget) to reduce file count.
- **Admin consolidation pattern**: Instead of separate pages per enterprise feature, consolidating SSO/audit/roles/billing/SLA into a single `admin.templ` with card sections reduces routing complexity and cognitive overhead.
- **Public vs authenticated routes**: `Docs()` and `Apply()` don't require auth — important for onboarding flow. Consistent `h.isAuthenticated(r)` guard + redirect pattern in all other handlers.
- **Templ curly braces**: `{slug}` in template text is interpreted as a Go expression. Use string literals like `YOUR-SLUG` or `fmt.Sprintf` instead.
- **Phase scaling**: As routes proliferate, consider grouping into chi route groups (e.g., `r.Route("/settings", ...)` and `r.Route("/admin", ...)`) to maintain organization.
