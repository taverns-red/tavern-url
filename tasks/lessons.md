# Lessons Learned

> Insights discovered during development. Newest entries at the top.

---

## 2026-03-26 — Interface-first repository design enables 100% unit test coverage without a database
**Context**: Building the first vertical slice (Sprint 1). Needed to test service and handler layers without Postgres.
**Insight**: Defining `LinkRepository` as an interface with sentinel errors (`ErrSlugExists`, `ErrLinkNotFound`) allowed creating lightweight mock implementations. The mock repo in service tests is ~30 lines and reusable. Handler tests use `httptest.NewRecorder()` with the same mock pattern.
**Impact**: Future layers (analytics, auth, orgs) should follow the same pattern: interface at the boundary, mock in tests. Integration tests with Postgres are a separate concern and should use a `_test` database.

## 2026-03-26 — Base62 slug collision retry is essential even at low scale
**Context**: Implementing auto-generated slugs with 6 chars (62^6 ≈ 56B combinations).
**Insight**: Even though collisions are astronomically unlikely at low scale, the collision retry loop (max 5 attempts) is cheap insurance. The test `TestCreateLink_AutoSlugRetry` verifies this path with a mock that forces 3 consecutive collisions before succeeding.
**Impact**: Always test the retry/fallback path, not just the happy path. Property-based tests for slug uniqueness would be a good Sprint 2 addition.

---

<!-- Template:
## YYYY-MM-DD — [Title]
**Context**: What were you doing?
**Insight**: What did you learn?
**Impact**: How does this change future work?
-->
