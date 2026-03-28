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

## Sprint 64 (Quality & Observability)

- **golangci-lint Go version mismatch**: The `golangci-lint-action@v6` prebuilt binary may be compiled with an older Go version than the project targets. Use `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` instead to build with the project's Go version.
- **Coverage gate total calculation**: `go tool cover -func` outputs a `total:` line that includes ALL files. Filtering individual function lines with `grep -v` doesn't recompute the total. Instead, average coverages from `go test -cover` output per package.
- **Ratcheting thresholds**: Start coverage gates at the current floor (e.g., 8%) instead of the aspirational target (60%). Ratchet up incrementally as tests are added. This prevents blocking CI while still catching regressions.
- **HTTP write errcheck exclusions**: Excluding `(net/http.ResponseWriter).Write` and `(templ.Component).Render` from errcheck is standard Go practice — when writing to an HTTP response, there's nothing useful to do with the error.

## Sprint 65 (Test Depth)

- **Weighted redirect testing**: Test weight=100 (always match) and weight=0 (never match) as deterministic boundary cases. Avoid testing intermediate weights probabilistically.
- **httptest.Server for webhook tests**: `httptest.NewServer` gives a real HTTP server for testing webhook delivery — much better than mocking the HTTP client.
- **staticcheck SA9003 (empty branch)**: If you're checking a condition but not doing anything in the branch, either add an assertion or remove the check. The linter catches this reliably.
- **Coverage gate ratcheting**: After adding 25 new tests, the gate can be safely raised from 8% → 20%. Never raise the gate before the tests exist.

## Sprint 66 (Polish & Ship)

- **FORCE_JAVASCRIPT_ACTIONS_TO_NODE24**: Adding this env var at the workflow level forces all Node.js-based actions to use Node.js 24, eliminating the deprecation warning immediately rather than waiting for the deadline.
- **govet catches logic bugs**: The `suspect or` govet check caught `!= A || != B` which is always true — should be `!= A && != B`. Lint before commit.
- **Dashboard templates already had status UI**: Before implementing a "new" feature, always check if the templates already have the UI elements. The dashboard already had Expired/Limit Reached badges, click counts, and expiration dates — only the link detail page was missing them.
- **Redirect benchmark baseline**: Geo rule: 60ns, Device rule: 182ns, No match: 285ns. All under 1µs. This is the baseline to track performance regressions against.

## Sprint 67 (Auth & Coverage Depth)

- **Model methods are pure functions — test them first.** IsExpired, IsExhausted, IsPasswordProtected are pure functions that went from 0% to 100% in minutes. Always start with these quick wins.
- **Auth helpers are testable without mocking services.** extractBearerToken, UserIDFromContext, UserFromContext are pure functions that can be tested in the auth package without any service mocking.
- **Existing test files may already exist.** Always check `find . -name '*_test.go'` before creating new test files. I almost recreated analytics_service_test.go.
- **errcheck catches test ignoring return values.** Even in tests, golangci-lint catches unchecked errors. Use `_, _, _ = svc.Method()` to explicitly discard returns.

## Sprint 68 (Handler & Auth Coverage Push)

- **RequireAuth middleware is testable with real gorilla/sessions.** Use `NewSessionStore` with a test secret, `SetUserID` to write a cookie, then pass cookies to the middleware-wrapped handler. No HTTP server needed.
- **Mock click repos can return hardcoded summaries.** Analytics handler tests don't need real data — a mock ClickRepository returning static ClickSummary objects is sufficient for handler-level testing.
- **errcheck catches test setup ignoring returns.** Even one-line test setup like `store.SetUserID(w, req, 42)` must handle or explicitly discard the error return.

## Sprint 69 (Middleware, Email & Handler Coverage)

- **Handler validation tests don't need a real auth service.** Register/Login checks for empty fields before calling the service, so tests with `authSvc: nil` safely cover those paths.
- **Security middleware is a pure function — easiest test.** Just call SecurityHeaders(handler), check 6 response headers. Instant 64% middleware coverage.
- **Email NoopSender test requires stdout capture.** Use `os.Pipe` + `io.Copy` to capture and verify `fmt.Printf` output in NoopSender.Send.
- **Coverage gate formula averages ALL packages including 0%.** cmd (200 lines) and repository (676 lines) drag the overall average down. Real tested-package average is much higher.

## Sprint 70 (Final Handler & Auth Coverage Push)

- **Mock APIKeyRepository in auth package enables full middleware testing.** RequireAPIKey and RequireAuthOrAPIKey need a real APIKeyService, which needs a repo. Creating a mock in a `_test.go` file in the auth package makes this testable without circular dependencies.
- **Coverage gate formula averages ALL packages — be precise with thresholds.** With 4 packages at 0% and 5 tested packages averaging 72%, the overall average is ~46%. Always check the actual CI output before setting a gate.
- **Org handler tests already existed — always check `git log` before writing new tests.** Sprint 65 already had comprehensive org handler tests. Saved time by checking first.

## Sprint 71 (Repository Integration Tests)

- **Integration tests need cleanup, not transactions.** Using t.Cleanup() with DELETE statements is simpler than pgx transaction rollback and avoids interfering with the running app's data.
- **t.Skip with DATABASE_URL makes integration tests CI-compatible.** Tests skip gracefully in CI without a DB, but run locally with the existing docker-compose postgres.
- **repository_test is an external test package.** Using `repository_test` forces testing only the public API, not internal struct fields.
- **Org service and handler tests already existed from Sprint 65.** Always check file existence before writing new test files.
