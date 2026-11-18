# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SafariQuest API is a Go REST API built with Gin and GORM backed by PostgreSQL. It uses Goose for database migrations.

## Commands

```bash
# Run the server
go run main.go

# Build
go build -o safari-quest-api .

# Run tests
go test ./...

# Run a single test
go test ./path/to/package -run TestName

# Install dependencies
go mod tidy

# Load DB env vars and run goose migrations
source .env && goose up

# Goose migration commands (after sourcing .env)
goose status
goose up
goose down
goose create <name> sql
```

## Architecture

- `main.go` — entry point; creates the Gin router and registers routes
- `database/postgres.go` — opens the GORM + raw `*sql.DB` connection; exposes `GORMDB`, `SQLDB`, and `DB_MIGRATOR` package-level vars; reads `GOOSE_DBSTRING` from the environment
- `.env` — environment variables for both the app and Goose (`GOOSE_DRIVER`, `GOOSE_DBSTRING`, `GOOSE_MIGRATION_DIR`); source this before running goose commands
- `api/v1/routes.go` — decides which router group each feature is mounted on (public, `Auth()` only, or `Auth()` + `Authorize()`); it declares no feature paths itself

### Feature packages (`internal/<feature>/`)

Everything belonging to one struct lives in one package, one file per layer, named `<struct>_<layer>.go` in snake_case:

```
internal/role/
├── role_model.go        # GORM model
├── role_repository.go   # database queries
├── role_service.go      # business logic, Input/Response DTOs, websocket events
├── role_controller.go   # Gin handlers + Swagger annotations
└── role_routes.go       # RegisterRoutes(rg) — the feature's paths
```

Features: `auth` (no model/repository), `authority`, `role`, `user`, `menu`, `geographytype` (files prefixed `geography_type_`), `geography`. The `/ws` controller and route live in `websocket/`.

Naming inside a package (the layers share one namespace, so names must not collide):
- Service: `GetAll`, `GetByUUID`, `Create`, `Update`, `Delete`; DTOs `Input`, `Response`; `ToResponse` when another feature needs the conversion
- Repository: `FindAll`, `FindByUUID`, `FindByUUIDs`, `Insert`, `Save`, `Remove`
- Controller type: `Controller`

Dependencies point one way only: `user → role → authority`, `menu → authority`, `geography → geographytype`, `auth → user`. Go forbids import cycles, so never add a back-reference field (e.g. `Authority.Roles`) or a call from a lower feature into a higher one. Feature routes must not import `middlewares` (that would cycle through `user` → `websocket`); `api/v1/routes.go` passes in an already-protected group instead.

When adding a feature: create `internal/<feature>/` with the files above, call its `RegisterRoutes` from `api/v1/routes.go`, and let Air (or `swag init`) regenerate `docs/`.

## Git Commits

- Commit each file separately — never batch multiple files into one commit
- Write detailed commit messages that explain what changed and why (e.g. "Add UUID primary key to Role model to avoid exposing sequential IDs")
- Do not add "Co-Authored-By" lines to any commit message

## Database

- PostgreSQL, accessed via `gorm.io/driver/postgres`
- Migrations managed by [Goose](https://github.com/pressly/goose); migration files go in `./migrations/` (per `GOOSE_MIGRATION_DIR`)
- Connection string is read from `GOOSE_DBSTRING` env var (also used by the app via `database.ConnectToDatabase()`)
