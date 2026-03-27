# Contributing to Tavern URL

Thank you for your interest in contributing! Tavern URL is built for nonprofits, and we welcome contributions from NFP tech volunteers and the broader open-source community.

## Development Setup

1. **Prerequisites**: Go 1.23+, Docker, [goose](https://github.com/pressly/goose), [templ](https://templ.guide/)
2. Clone the repo and start Postgres: `docker compose up -d`
3. Run migrations: `make migrate`
4. Start dev server: `make dev`

## Workflow

1. Pick an issue (or create one)
2. Create a branch: `git checkout -b feat/my-feature`
3. Write tests first (TDD)
4. Implement the feature
5. Run `make test` and `make vet`
6. Commit with conventional commits: `feat: add feature (#42)`
7. Open a PR

## Code Style

- `go vet` and `go test -race` must pass
- Follow existing patterns in `internal/handler`, `internal/service`, `internal/repository`
- Use the repository pattern for database access
- Keep handlers thin — business logic belongs in services

## Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` new features
- `fix:` bug fixes
- `docs:` documentation
- `test:` adding tests
- `chore:` maintenance
- `refactor:` code changes that don't fix bugs or add features
