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

Features: `auth` (no model/repository), `authority`, `role`, `user`, `menu`, `geographytype` (files prefixed `geography_type_`), `geography`, `attractiontype`, `attractioncategory`, `touroperator` (files prefixed `tour_operator_`), `attraction`, `booking`, `review`, `catalog` (public, no model/repository), `account` (password reset and own profile), `setting` (system settings). The `/ws` controller and route live in `websocket/`.

Naming inside a package (the layers share one namespace, so names must not collide):
- Service: `GetAll`, `GetByUUID`, `Create`, `Update`, `Delete`; DTOs `Input`, `Response`; `ToResponse` when another feature needs the conversion
- Repository: `FindAll`, `FindByUUID`, `FindByUUIDs`, `Insert`, `Save`, `Remove`
- Controller type: `Controller`

Dependencies point one way only: `user → role → authority`, `menu → authority`, `geography → geographytype`, `auth → user`, `touroperator → user, geography`, `attraction → attractiontype, attractioncategory, geography, touroperator`, `tourpackage → attraction, touroperator, fileresource`, `booking → tourpackage, touroperator, user, companyscope`, `review → booking, tourpackage, user, companyscope`, `catalog → tourpackage, booking, review`, `account → auth, user`, `setting → user`, `booking → setting`. Go forbids import cycles, so never add a back-reference field (e.g. `Authority.Roles`) or a call from a lower feature into a higher one. Feature routes must not import `middlewares` (that would cycle through `user` → `websocket`); `api/v1/routes.go` passes in an already-protected group instead.

Authority codes are derived from the last non-parameter path segment (`pkg/authority`), so a nested route like `/tour-operators/:uuid/staff` gets `staff_*` codes. Never nest a route under a segment that already names another resource (e.g. `/users`), or its permission would also grant that resource's global endpoints.

Tour operator staff live in the `tour_operator_users` pivot. The `/tour-operators/:uuid/staff` endpoints additionally require the caller to be staff of that operator or hold the `ADMINISTRATOR` role, and staff may only grant roles they hold themselves.

Bookings (`internal/booking`): a traveler's seats on a sale package; total = package price per person × pax. Pending and confirmed bookings together may not exceed the package's pax; the check runs in a transaction that locks the package row. Status moves pending → confirmed | cancelled, confirmed → completed | cancelled; completed and cancelled are final. Company staff use `/bookings` and `/travelers` (`booking_*`, `traveler_get`, scoped by `companyscope`); travelers use `/bookable-packages` and `/my-bookings` (`bookable-package_get`, `my-booking_*`), where `DELETE /my-bookings/:uuid` cancels rather than deletes. The seeder keeps a `TRAVELER` role holding only those traveler codes plus `menu_get`.

Groups and pricing: a package runs `units` groups (1–50) on the same dates, each with `pax` seats (`tourpackage.Selling`, embedded). `booking_mode` is `shared`, `private` or `both`. Shared bookings buy seats priced by `price`; private bookings take a whole empty group priced by `private_price`. Each kind is priced `per_person` (charged for at least `*_min_travellers`, so a solo traveller can pay a minimum) or `per_group`; `max_private_units` optionally caps private groups. Bookings carry `booking_type`, `unit_number` (the group) and a copy of their pricing (`pricing`, `unit_price`, `min_travellers`, `charged_travellers`), so later price changes never alter them and edits re-price from that copy. Inside the locked transaction (`saveChecked`) shared bookings go to the partly filled group they fit most tightly, then to an empty one, never split across groups or into a private group; private bookings take the lowest empty group. `booking.Availability` (`seats_left`, `max_shared_travellers`, `private_groups_left`) replaces the old seat count everywhere. Package edits that would strand active bookings (fewer groups or seats, dropping a booking type in use) return 409 (`ErrBookedBeyond`).

Packages (`internal/tourpackage`) carry traveler-facing details: `highlights`, `inclusions` and `exclusions` are `pkg/dbtypes.StringList` (JSONB arrays), plus `difficulty` (easy/moderate/challenging), `meeting_point`, linked attractions (`package_attractions`) and an itinerary (`package_itinerary_days`). Saving a package replaces its attractions and itinerary as a whole; days are numbered in the order sent and may not outnumber the trip's days. `price` is per person. Responses include `rating` (average and count of visible reviews).

Reviews (`internal/review`, table `package_reviews`): one per completed booking, written by its traveler (`/my-reviews`, `my-review_*`). Anyone signed in reads a package's visible reviews at `/package-reviews` (`package-review_get`). Company staff list their packages' reviews at `/reviews` (`review_get`) and post one public reply at `/review-replies/:uuid` (`review-reply_put`). Only administrators hold `review-moderation_put` to hide a review; hidden reviews stay stored but leave public lists and ratings. Public reviews show the reviewer as first name and last initial.

Sign-up (`POST /auth/register`, public): travelers get the TRAVELER role; operators also get a new tour company (`is_verified = false`, unique name, generated code) linked through `tour_operator_users` and the TOUR_OPERATOR role, all in one transaction. The response signs the account in like login. Passwords need 8+ characters with a letter and a digit; emails are stored lowercase and matched case-insensitively at login. Packages of unverified or inactive companies are hidden from `/catalog` and traveler booking (`tourpackage.PublishedOperator` / `Published`) until an administrator verifies the company; staff screens are unaffected.

Booking confirmation: companies answer pending bookings with `POST /booking-confirmations/:uuid` and `POST /booking-declines/:uuid` (reason 3–500 chars, shown to the traveler); both are outside `/bookings` so their codes (`booking-confirmation_post`, `booking-decline_post`, granted to TOUR_OPERATOR) don't grant the bookings resource. `declined` and `expired` are final statuses that free seats; the general `PUT /bookings/:uuid` cannot set them. A new pending booking gets `expires_at` = now + `booking_confirm_hours` (System Settings, default 48, 0 = never); `booking.RunExpiry` (started in `main.go`) expires overdue ones at start-up and every 5 minutes. Emails (`booking_notify.go`, sent in the background) go to the traveler on received/confirmed/declined/expired/cancelled-by-company and to the company's staff on new, expired and traveler-cancelled bookings. `GET /my-bookings/:uuid` returns the traveler's booking with trip details and company contacts for the confirmation voucher.

System settings (`internal/setting`, table `system_settings`): settings are defined in code (`Definitions`: key, label, unit, default, min, max); the table only stores values. `GET`/`PUT /system-settings` (`system-setting_*`, administrators only) validate every value before saving any; `setting.Int(key)` falls back to the default.

Account (`internal/account`): `POST /auth/forgot-password` emails a one-hour reset link (`APP_URL/reset-password?token=…`) and always answers the same way; only the SHA-256 hash of the token is stored (`password_reset_tokens`), a link works once, a new link or a reset cancels older ones, and requests within a minute of the last are ignored. `POST /auth/reset-password` sets the password. `GET`/`PUT /me` and `PUT /me/password` (needs the current password) sit behind `Auth()` only, so every account manages itself without an authority; email, roles and status are not editable there. The seeder skips `/auth/` and `/me` paths. Email goes through `pkg/mailer` (`SMTP_*`, `MAIL_FROM`); with no `SMTP_HOST` it writes the message to the log.

Public catalog (`internal/catalog`): `GET /catalog/packages`, `/catalog/packages/:uuid` and `/catalog/packages/:uuid/reviews` need no sign-in and feed the public website. The list shows only packages that have not started (filters: `category`, `eco`, `max_price`, `search`, `sort`); detail also returns started packages with `bookable: false`. Responses add `image_path` (first image attachment), `duration_days` and `seats_left`, and never include attachment metadata. The seeder skips `/catalog/` paths, so they create no authority codes. Packages also carry `location`, `currency` (ISO 4217, default USD) and `eco_certified`; `package_type` is the website category (Safari, Mountain Climbing, Beach, Cultural, Eco-Tour).

`SEED_DEMO=true` loads the website's sample tour companies and packages (`seeders/demo_packages.json`) as real records starting 30+ days out; companies use `DEMO_*` codes and existing rows are skipped.

When adding a feature: create `internal/<feature>/` with the files above, call its `RegisterRoutes` from `api/v1/routes.go`, and let Air (or `swag init`) regenerate `docs/`.

## Git Commits

- Commit each file separately — never batch multiple files into one commit
- Write detailed commit messages that explain what changed and why (e.g. "Add UUID primary key to Role model to avoid exposing sequential IDs")
- Do not add "Co-Authored-By" lines to any commit message

## Database

- PostgreSQL, accessed via `gorm.io/driver/postgres`
- Migrations managed by [Goose](https://github.com/pressly/goose); migration files go in `./migrations/` (per `GOOSE_MIGRATION_DIR`)
- Connection string is read from `GOOSE_DBSTRING` env var (also used by the app via `database.ConnectToDatabase()`)
