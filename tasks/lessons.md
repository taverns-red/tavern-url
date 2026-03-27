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
