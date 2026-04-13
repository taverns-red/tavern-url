# Changelog

## [1.1.0](https://github.com/taverns-red/tavern-url/compare/v1.0.0...v1.1.0) (2026-03-30)


### Features

* add ErrorPage template and graceful OAuth failure handling ([#61](https://github.com/taverns-red/tavern-url/issues/61)) ([22cd263](https://github.com/taverns-red/tavern-url/commit/22cd263f64df3144fc4a891bcdb31030d3cddd0c))
* add expiration and click limit status to link detail page ([d519082](https://github.com/taverns-red/tavern-url/commit/d519082b327d8c45a67d30bef9d618fb1032f823))
* add Fly.io deployment infrastructure ([#64](https://github.com/taverns-red/tavern-url/issues/64)) ([806124a](https://github.com/taverns-red/tavern-url/commit/806124a31f6fb0346c67db5ccae897d333fa383b))
* migrate to structured logging with log/slog ([#60](https://github.com/taverns-red/tavern-url/issues/60)) ([48a1479](https://github.com/taverns-red/tavern-url/commit/48a1479592ce2ada9f0295bb9e322e9bde4b08aa))
* **security:** harden configuration and implement CSRF ([#1](https://github.com/taverns-red/tavern-url/issues/1), [#2](https://github.com/taverns-red/tavern-url/issues/2), [#3](https://github.com/taverns-red/tavern-url/issues/3), [#4](https://github.com/taverns-red/tavern-url/issues/4)) ([69b94bf](https://github.com/taverns-red/tavern-url/commit/69b94bfcdd10f4eba2615e56c9dfba1495a0045d))


### Bug Fixes

* add TrustedOrigins to CSRF config for TLS termination ([#64](https://github.com/taverns-red/tavern-url/issues/64)) ([fdc360c](https://github.com/taverns-red/tavern-url/commit/fdc360ccd5fd54f6b7749c864cdb550dbe99b291))
* **ci:** correct coverage gate to 39% (average includes 0% packages) ([72697d1](https://github.com/taverns-red/tavern-url/commit/72697d105dc168bcbe27d1ac306ac9b493bc88b8))
* compile removed inline script to satisfy CSP & bust app.js cache ([#64](https://github.com/taverns-red/tavern-url/issues/64)) ([a3693a6](https://github.com/taverns-red/tavern-url/commit/a3693a65f8b88382184fd9039bc41b672fbb34ad))
* coverage gate threshold and calculation ([#38](https://github.com/taverns-red/tavern-url/issues/38)) ([0eb46e8](https://github.com/taverns-red/tavern-url/commit/0eb46e8e5d5f03274a3b96e0d29e6740fa06b799))
* move HTMX CSRF config to static JS to respect CSP ([#64](https://github.com/taverns-red/tavern-url/issues/64)) ([3d1d0bb](https://github.com/taverns-red/tavern-url/commit/3d1d0bb123cb1e2a92a994852a20effaa055d090))
* parse URL host for CSRF TrustedOrigins ([#64](https://github.com/taverns-red/tavern-url/issues/64)) ([5f26439](https://github.com/taverns-red/tavern-url/commit/5f2643973299e82e2949efbc9c0164cbe7e80794))
* replace panic in GenerateSlug with error return ([#59](https://github.com/taverns-red/tavern-url/issues/59)) ([c8ee4fa](https://github.com/taverns-red/tavern-url/commit/c8ee4faaa717e045c165edfe602bb6fbb1dc3220))
* use CMD instead of ENTRYPOINT for Fly release_command ([#64](https://github.com/taverns-red/tavern-url/issues/64)) ([3840888](https://github.com/taverns-red/tavern-url/commit/3840888da77b3a2314223128f97090cf93a77a43))
* use go install for golangci-lint in CI ([#37](https://github.com/taverns-red/tavern-url/issues/37)) ([2d720fa](https://github.com/taverns-red/tavern-url/commit/2d720fa490a92b44559cb42631b095c814f054ad))

## 1.0.0 (2026-03-28)


### Features

* add analytics engine, QR codes, and analytics API (Sprints 6-8) ([#7](https://github.com/taverns-red/tavern-url/issues/7) [#8](https://github.com/taverns-red/tavern-url/issues/8) [#9](https://github.com/taverns-red/tavern-url/issues/9)) ([a836e79](https://github.com/taverns-red/tavern-url/commit/a836e79ba8149c291a9f9c6e39be4882a619a314))
* add API key auth middleware and management handlers (Sprint 12) ([#15](https://github.com/taverns-red/tavern-url/issues/15)) ([0b9da9a](https://github.com/taverns-red/tavern-url/commit/0b9da9a233a2e04b924ff7ff6263ff1a818348f2))
* add API keys, org management, CI/CD, and deploy config (Sprints 9-11) ([#10](https://github.com/taverns-red/tavern-url/issues/10) [#11](https://github.com/taverns-red/tavern-url/issues/11) [#12](https://github.com/taverns-red/tavern-url/issues/12)) ([8ad53b8](https://github.com/taverns-red/tavern-url/commit/8ad53b8f6e99850bc8e352eb8f1684fa4eab9da3))
* add domain model, migration, and slug generation ([#2](https://github.com/taverns-red/tavern-url/issues/2)) ([8b7e3dd](https://github.com/taverns-red/tavern-url/commit/8b7e3dd28caf0f56233f1fe0112b0e88befab83a))
* add initial product specification document for Tavern URL ([7a48dd7](https://github.com/taverns-red/tavern-url/commit/7a48dd7e3c0a41813fa20adfb29867830c1c7ac4))
* add link detail page with analytics charts (Sprint 13) ([#16](https://github.com/taverns-red/tavern-url/issues/16)) ([3bee15e](https://github.com/taverns-red/tavern-url/commit/3bee15e39237383703308cac6d83b85c0f41cbc2))
* add link management UI and CRUD API (Sprint 5) ([#6](https://github.com/taverns-red/tavern-url/issues/6)) ([39aa0ea](https://github.com/taverns-red/tavern-url/commit/39aa0ea117b2c1dbdacd35a33de97a492da9a4d3))
* add main.go application wiring with graceful shutdown ([#2](https://github.com/taverns-red/tavern-url/issues/2)) ([6a26cbb](https://github.com/taverns-red/tavern-url/commit/6a26cbbe20e48670b9bb2c713c47239eb3ca7a6a))
* add organizations, memberships, and Google OAuth ([#4](https://github.com/taverns-red/tavern-url/issues/4)) ([d3cd7cd](https://github.com/taverns-red/tavern-url/commit/d3cd7cdf18ac43da5f52ff6592a5d8c67a5c7e22))
* add repository, service, and HTTP handlers ([#2](https://github.com/taverns-red/tavern-url/issues/2)) ([0b1f918](https://github.com/taverns-red/tavern-url/commit/0b1f918fa93b4595438500d0b4047e4f07c96df4))
* add user authentication with sessions ([#3](https://github.com/taverns-red/tavern-url/issues/3)) ([96abe35](https://github.com/taverns-red/tavern-url/commit/96abe35f81ef4da4e6bbb883546883d0787798e8))
* add web UI foundation with Templ + HTMX (Sprint 4) ([#5](https://github.com/taverns-red/tavern-url/issues/5)) ([58d7ba0](https://github.com/taverns-red/tavern-url/commit/58d7ba0d530da6f01b15bf06672f9e6148d5f547))
* dark mode CSS and static asset embedding support (Sprint 22) ([#26](https://github.com/taverns-red/tavern-url/issues/26)) ([328bf21](https://github.com/taverns-red/tavern-url/commit/328bf2122fca1855ebd69190ccf36cbe622b9780))
* link expiration, CSV export, custom domains, security headers (Sprints 17-21) ([#20](https://github.com/taverns-red/tavern-url/issues/20) [#21](https://github.com/taverns-red/tavern-url/issues/21) [#22](https://github.com/taverns-red/tavern-url/issues/22) [#23](https://github.com/taverns-red/tavern-url/issues/23) [#24](https://github.com/taverns-red/tavern-url/issues/24)) ([5aea46e](https://github.com/taverns-red/tavern-url/commit/5aea46e53fa3a42cad860e4f409a3fd36462b74e))
* org invite/role management, rate limiting, README, docs (Sprints 14-16) ([#17](https://github.com/taverns-red/tavern-url/issues/17) [#18](https://github.com/taverns-red/tavern-url/issues/18) [#19](https://github.com/taverns-red/tavern-url/issues/19)) ([c1a95e3](https://github.com/taverns-red/tavern-url/commit/c1a95e3a5bbd6b19eb8bd0c93fad4c78ad70a5e0))
* password links, dynamic redirects, extension, webhooks, onboarding (Sprints 26-31) ([#30](https://github.com/taverns-red/tavern-url/issues/30) [#31](https://github.com/taverns-red/tavern-url/issues/31) [#32](https://github.com/taverns-red/tavern-url/issues/32) [#33](https://github.com/taverns-red/tavern-url/issues/33) [#34](https://github.com/taverns-red/tavern-url/issues/34) [#35](https://github.com/taverns-red/tavern-url/issues/35)) ([196a789](https://github.com/taverns-red/tavern-url/commit/196a78924962e6f7ad2b9edb28bb81e47b8005e8))
* search, editing, bulk create, email, password reset, DNS verification (Sprints 23-25) ([#27](https://github.com/taverns-red/tavern-url/issues/27) [#28](https://github.com/taverns-red/tavern-url/issues/28) [#29](https://github.com/taverns-red/tavern-url/issues/29)) ([cb3d21b](https://github.com/taverns-red/tavern-url/commit/cb3d21bdbf2e1fe34d6dd6b82c6c72f8316eac79))
* Sprint 33 — UI feature parity ([b1f7ffe](https://github.com/taverns-red/tavern-url/commit/b1f7ffe140f160b9fc6a1eab1357d2590f48773f))
* Sprint 34 — link expiration & click limits UI ([#35](https://github.com/taverns-red/tavern-url/issues/35)) ([d29f4fc](https://github.com/taverns-red/tavern-url/commit/d29f4fc0b74839fd290d6c2b438a24a4a2b8dea1))
* Sprint 35 — org management UI ([#36](https://github.com/taverns-red/tavern-url/issues/36)) ([24fd0e2](https://github.com/taverns-red/tavern-url/commit/24fd0e27aaa5edb13a150981bab7d60371f97c93))
* Sprint 36 — password-protected links UI ([#37](https://github.com/taverns-red/tavern-url/issues/37)) ([346b202](https://github.com/taverns-red/tavern-url/commit/346b20272245fbad96e517bdfab8621e5f21a3d4))
* Sprint 37 — redirect rules UI ([#38](https://github.com/taverns-red/tavern-url/issues/38)) ([84d00a7](https://github.com/taverns-red/tavern-url/commit/84d00a7af84c6d68b8f307c2303a52c58b38d2f4))
* Sprint 38 — UTM builder & campaign tags ([#39](https://github.com/taverns-red/tavern-url/issues/39)) ([599c2df](https://github.com/taverns-red/tavern-url/commit/599c2dfe794852fcb6de6ece897af7bef9ee5365))
* Sprints 39-43 — Phase 2 Growth features ([#40](https://github.com/taverns-red/tavern-url/issues/40)-[#44](https://github.com/taverns-red/tavern-url/issues/44)) ([c65a0fb](https://github.com/taverns-red/tavern-url/commit/c65a0fb74ec5e85c6c97cb99e6171eae78e4b112))
* Sprints 44-48 — Phase 3 Platform features ([#45](https://github.com/taverns-red/tavern-url/issues/45)-[#49](https://github.com/taverns-red/tavern-url/issues/49)) ([f3fd197](https://github.com/taverns-red/tavern-url/commit/f3fd197353ec91176ff105d181a08a344692fa9b))
* Sprints 49-63 — Phases 4-6 Enterprise, Scale, Ecosystem ([#50](https://github.com/taverns-red/tavern-url/issues/50)-[#64](https://github.com/taverns-red/tavern-url/issues/64)) ([dd7bcdc](https://github.com/taverns-red/tavern-url/commit/dd7bcdc7b5ce547ba877cea5eb65e6f15567c752))


### Bug Fixes

* **ci:** upgrade Dockerfile Go from 1.23 to 1.25 ([b183cf5](https://github.com/taverns-red/tavern-url/commit/b183cf5de1a1b953181189cd9735214c9422e7c6))
* **ci:** use go.mod Go version and remove coverprofile ([27d9874](https://github.com/taverns-red/tavern-url/commit/27d98740e7b0d603e54c8bb197f92feefae26fdb))
* HTMX compatibility for all handlers (CSP, forms, redirects) ([f5cd482](https://github.com/taverns-red/tavern-url/commit/f5cd4824dd3e794ad1489e2dcf3dfb16ab1b6fbf))
* invite member form not submitting ([#36](https://github.com/taverns-red/tavern-url/issues/36)) ([06bd706](https://github.com/taverns-red/tavern-url/commit/06bd706059dc4ab95cd512027d8b1749a698feaf))
* register/login forms accept form-encoded data from HTMX ([#26](https://github.com/taverns-red/tavern-url/issues/26)) ([bfae9f7](https://github.com/taverns-red/tavern-url/commit/bfae9f77e10948264628a23bf55702451519ff2c))
* replace inline onclick with external app.js for CSP compliance ([0c80f68](https://github.com/taverns-red/tavern-url/commit/0c80f68492e62a5c32c298f6931dca95f8c40d2c))
* use go-version-file to match go.mod, remove coverprofile flag. ([27d9874](https://github.com/taverns-red/tavern-url/commit/27d98740e7b0d603e54c8bb197f92feefae26fdb))
