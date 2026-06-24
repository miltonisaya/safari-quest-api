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
source env/.env && goose up

# Goose migration commands (after sourcing env/.env)
goose status
goose up
goose down
goose create <name> sql
```

## Architecture

- `main.go` — entry point; creates the Gin router and registers routes
- `database/postgres.go` — opens the GORM + raw `*sql.DB` connection; exposes `GORMDB`, `SQLDB`, and `DB_MIGRATOR` package-level vars; reads `GOOSE_DBSTRING` from the environment
- `env/goose.env` — shell exports for Goose (`GOOSE_DRIVER`, `GOOSE_DBSTRING`, `GOOSE_MIGRATION_DIR`); source this before running goose commands

## Git Commits

- Commit each file separately — never batch multiple files into one commit
- Write detailed commit messages that explain what changed and why (e.g. "Add UUID primary key to Role model to avoid exposing sequential IDs")
- Do not add "Co-Authored-By" lines to any commit message

## Database

- PostgreSQL, accessed via `gorm.io/driver/postgres`
- Migrations managed by [Goose](https://github.com/pressly/goose); migration files go in `./migrations/` (per `GOOSE_MIGRATION_DIR`)
- Connection string is read from `GOOSE_DBSTRING` env var (also used by the app via `database.ConnectToDatabase()`)
