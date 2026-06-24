# SafariQuest API

A production-ready REST API built with Go, Gin, GORM, and PostgreSQL. Features JWT authentication, role-based access control, automatic database migrations, and Swagger documentation.

## Tech Stack

- **Go 1.25** with **Gin** (HTTP framework)
- **GORM** + **PostgreSQL** (database)
- **Goose** (database migrations)
- **JWT** (authentication)
- **Swaggo** (API documentation)
- **Air** (live reload for development)

## Prerequisites

- Go 1.25+
- PostgreSQL
- [Air](https://github.com/air-verse/air) — `go install github.com/air-verse/air@latest`
- [Swag](https://github.com/swaggo/swag) — `go install github.com/swaggo/swag/cmd/swag@latest`
- [Goose](https://github.com/pressly/goose) — `go install github.com/pressly/goose/v3/cmd/goose@latest`

## Getting Started

### 1. Clone and install dependencies

```bash
git clone https://github.com/miltonisaya/safari-quest-api.git
cd safari-quest-api
go mod tidy
```

### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env` with your values:

```env
SERVER_PORT=3000
GIN_MODE=debug
DB_STRING=postgres://user:password@localhost:5432/safari_quest?sslmode=disable
JWT_SECRET=your-long-random-secret
SWAGGER_HOST=localhost:3000

SEED=true
ADMIN_EMAIL=super@safariquest.com
ADMIN_DEFAULT_PASSWORD=Admin@1234

GOOSE_DRIVER=postgres
GOOSE_DBSTRING=postgres://user:password@localhost:5432/safari_quest?sslmode=disable
GOOSE_MIGRATION_DIR=./migrations
```

### 3. Create the database

```bash
psql -U postgres -c "CREATE DATABASE safari_quest;"
```

### 4. Run the server

```bash
air
```

On startup the server will:
1. Run all pending database migrations automatically
2. Seed authorities, the Administrator role, and a super admin user (when `SEED=true`)

## API Documentation

Swagger UI is available at:

```
http://localhost:3000/swagger/index.html
```

To authenticate in Swagger UI, log in via `POST /api/v1/auth/login`, copy the returned token, and click **Authorize** — enter `Bearer {token}`.

## Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `POST` | `/api/v1/auth/login` | Login and receive a JWT | No |
| `GET` | `/api/v1/roles` | List all roles | Yes |
| `POST` | `/api/v1/roles` | Create a role | Yes |
| `GET` | `/api/v1/roles/:uuid` | Get a role | Yes |
| `PUT` | `/api/v1/roles/:uuid` | Update a role | Yes |
| `DELETE` | `/api/v1/roles/:uuid` | Delete a role | Yes |
| `GET` | `/api/v1/users` | List all users | Yes |
| `POST` | `/api/v1/users` | Create a user | Yes |
| `GET` | `/api/v1/users/:uuid` | Get a user | Yes |
| `PUT` | `/api/v1/users/:uuid` | Update a user | Yes |
| `DELETE` | `/api/v1/users/:uuid` | Delete a user | Yes |

## Database Migrations

Migrations run automatically on startup. To run them manually:

```bash
source .env && goose up      # apply all pending migrations
source .env && goose down    # roll back the last migration
source .env && goose status  # show migration status
```

To create a new migration:

```bash
source .env && goose create <name> sql
```

## Project Structure

```
.
├── api/v1/          # Route registration
├── config/          # Environment variable loading
├── controllers/     # HTTP handlers (thin — delegate to services)
├── database/        # GORM connection and migration runner
├── docs/            # Generated Swagger docs (do not edit)
├── middlewares/     # Recovery, Logger, Auth, Authorize
├── migrations/      # Goose SQL migration files
├── models/          # GORM models (all fields json:"-")
├── pkg/
│   ├── authority/   # Authority code derivation shared by middleware and seeder
│   └── response/    # JSend-style response wrapper (CustomApiResponse)
├── repositories/    # Database queries
├── seeders/         # Auto-discovers routes and seeds authorities, roles, and admin user
└── services/        # Business logic and DTOs
```

## Development

Air watches for file changes, regenerates Swagger docs, and restarts the server automatically:

```bash
air
```

After adding a new endpoint, add Swagger annotation comments above the controller method — Air will run `swag init` and pick them up on the next save.

## Authentication

All protected endpoints require a `Bearer` token in the `Authorization` header:

```
Authorization: Bearer <token>
```

Tokens are signed HS256 JWTs with a 24-hour expiry. The `sub` claim holds the user UUID. Permissions are checked per-request via the `Authorize` middleware, which derives the required authority code from the route and verifies the user holds it through their assigned roles.
