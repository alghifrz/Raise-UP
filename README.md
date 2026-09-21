# RAISE UP

Digital portal and automation system for RW (Rukun Warga) administration.

## Stack

- **Backend:** Go, Gin, PostgreSQL, sqlc, golang-migrate, JWT, RBAC
- **Frontend:** React, Vite, TypeScript, Tailwind CSS, React Router, TanStack Query
- **Infra:** Docker / Docker Compose

## Project Structure

```text
/
├── backend/
├── frontend/
├── docker-compose.yml
├── README.md
└── docs/
```

### Backend

```text
backend/
├── cmd/api/           # Application entrypoint
├── config/            # Environment-based configuration
├── middleware/        # HTTP middleware (CORS, auth, RBAC, logging)
├── pkg/               # Shared packages (JWT, response, database, …)
├── internal/          # Domain modules
└── db/
    ├── migrations/
    ├── queries/
    └── generated/
```

### Frontend 

```text
frontend/
└── src/
    ├── app/           # Providers and app shell
    ├── components/    # Reusable UI primitives
    ├── features/
    │   ├── auth/      # Login + auth state
    │   └── dashboard/ # Dashboard summary
    ├── layouts/       # Admin layout, sidebar, topbar
    ├── lib/           # API client, storage, utils
    └── routes/        # Route config + protected routes
```

Architecture style: **modular monolith** on the backend. The frontend mirrors domain features under `features/`.

API versioning: `/api/v1/`.

Health endpoint (unversioned): `GET /health`.

## Prerequisites

- Go 1.25+
- Node.js 20+ (for the frontend)
- Docker & Docker Compose
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI (for local migrations)
- [sqlc](https://docs.sqlc.dev/) CLI (for query code generation)

## Environment

Copy the example env file and adjust values:

```bash
cp .env.example .env
```

Required variables:

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string |
| `PORT` | HTTP listen port (default `8080`) |
| `JWT_SECRET` | Secret for signing JWT access tokens (min 16 chars) |

Optional:

| Variable | Description |
|---|---|
| `JWT_EXPIRES_IN` | Access token lifetime (`1h`, `3600`, etc.; default `1h`) |
| `CORS_ALLOWED_ORIGINS` | Comma-separated origins (default `http://localhost:5173`) |
| `SHUTDOWN_TIMEOUT_SECONDS` | Graceful shutdown timeout (default `10`) |
| `API_HOST_PORT` | Host port mapped to the API container (default `8080`) |
| `ADMIN_EMAIL` / `ADMIN_PASSWORD` / `ADMIN_NAME` / `ADMIN_ROLE` | Used by `cmd/seed` only |
| `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` | Used by Docker Compose |
| `POSTGRES_PORT` | Host port for Postgres (default `5433` to avoid clashing with local installs on `5432`) |
| `DOCKER_DATABASE_URL` | Optional external PostgreSQL URL used by the Compose API and migration containers |
| `SUPABASE_URL` | Supabase project URL used by the backend Storage client |
| `SUPABASE_SECRET_KEY` | Preferred server-only Storage credential (`sb_secret_...`); never expose it to the frontend |
| `SUPABASE_SERVICE_ROLE_KEY` | Legacy server-key fallback when a secret key is unavailable |
| `SUPABASE_GALLERY_BUCKET` | Public gallery bucket name (default `gallery`) |
| `SUPABASE_GALLERY_MAX_UPLOAD_BYTES` | Maximum gallery upload size (default `10485760`) |
| `WHATSAPP_ENABLED` | Enable Meta WhatsApp Cloud API (`true`/`false`; also requires credentials below) |
| `WHATSAPP_API_VERSION` | Graph API version (default `v21.0`) |
| `WHATSAPP_PHONE_NUMBER_ID` | Meta Phone number ID |
| `WHATSAPP_BUSINESS_ACCOUNT_ID` | WhatsApp Business Account ID (optional metadata) |
| `WHATSAPP_ACCESS_TOKEN` | Permanent / system-user access token |
| `WHATSAPP_VERIFY_TOKEN` | Shared secret for webhook verification challenge |
| `WHATSAPP_APP_SECRET` | Meta App Secret (validates `X-Hub-Signature-256`) |

## Frontend

Admin UI for authentication and the dashboard summary.

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

Frontend env:

| Variable | Description |
|---|---|
| `VITE_API_BASE_URL` | Backend origin, e.g. `http://localhost:8080` |

Routes:

- `/login` — admin login
- `/dashboard` — protected dashboard (real API data)
- `/residents` — residents list + create/edit/delete
- `/complaints` — complaints list + create
- `/complaints/:id` — complaint detail, edit, status actions, delete
- `/announcements` — announcements list + create
- `/announcements/:id` — announcement detail, edit, publish, delete
- `/finance` — finance transactions + summary
- `/finance/:id` — transaction detail
- `/dues` — dues periods
- `/dues/periods/:id` — period detail, summary, payments
- `/activities` — activities list + create
- `/activities/:id` — activity detail, edit, delete
- other sidebar modules — coming soon placeholders

Auth flow:

1. `POST /api/v1/auth/login` stores JWT in `localStorage`
2. On startup, if a token exists, `GET /api/v1/auth/me` restores the session
3. Invalid tokens (401) are cleared; user is treated as logged out

CORS: backend default `CORS_ALLOWED_ORIGINS` already includes `http://localhost:5173`.

## Authentication

Endpoints:

- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me` (requires `Authorization: Bearer <token>`)

Roles reuse the PostgreSQL enum values via sqlc: `SUPER_ADMIN`, `ADMIN_RW`.

RBAC helpers:

- `middleware.RequireAuth(tokens)`
- `middleware.RequireRole(db.UserRoleSUPERADMIN, db.UserRoleADMINRW)`

### Seed an administrator

There is no public registration endpoint. Create the first admin with:

```bash
cd backend
go run ./cmd/seed
```

Requires `ADMIN_EMAIL` and `ADMIN_PASSWORD` in `.env` (see `.env.example`).  
If the email already exists, the seed command updates name, password, and role.

### Example login

```bash
curl -s http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"change-me-admin-password"}'
```

## Residents API

All resident endpoints require JWT auth and role `SUPER_ADMIN` or `ADMIN_RW`.

- `GET /api/v1/residents?page=1&page_size=20&search=&gender=`
- `GET /api/v1/residents/:id`
- `POST /api/v1/residents`
- `PATCH /api/v1/residents/:id`
- `DELETE /api/v1/residents/:id`

## Complaints API

All complaint endpoints require JWT auth and role `SUPER_ADMIN` or `ADMIN_RW`.

- `GET /api/v1/complaints?page=1&page_size=20&search=&status=&urgency=&category=`
- `GET /api/v1/complaints/:id`
- `POST /api/v1/complaints`
- `PATCH /api/v1/complaints/:id`
- `PATCH /api/v1/complaints/:id/status`
- `DELETE /api/v1/complaints/:id`

### Create example

```bash
curl -s http://localhost:8080/api/v1/complaints \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "resident_name": "Budi",
    "phone": "08123456789",
    "block": "A1",
    "category": "Kebersihan",
    "urgency": "NORMAL",
    "message": "Sampah menumpuk"
  }'
```

If `resident_id` is provided, it is **authoritative**: the service loads that resident and stores its current name/phone as the complaint snapshot (request `resident_name`/`phone` are ignored).

### Status transitions

`PATCH /api/v1/complaints/:id/status` with `{"status":"DIPROSES"}`.

Allowed:

- `BARU` → `DIPROSES` | `DITOLAK`
- `DIPROSES` → `SELESAI` | `DITOLAK`

Terminal states `SELESAI` and `DITOLAK` cannot change (HTTP 409 `INVALID_STATUS_TRANSITION`).

Status cannot be changed via the general PATCH endpoint.

## Announcements API

All announcement endpoints require JWT auth and role `SUPER_ADMIN` or `ADMIN_RW`.

- `GET /api/v1/announcements?page=1&page_size=20&search=&status=&visibility=&category=`
- `GET /api/v1/announcements/:id`
- `POST /api/v1/announcements`
- `PATCH /api/v1/announcements/:id`
- `PATCH /api/v1/announcements/:id/status`
- `DELETE /api/v1/announcements/:id`

### PUBLIC vs PRIVATE

- **PUBLIC**: `recipient_ids` must be empty; GET returns `recipient_ids: []`
- **PRIVATE**: requires at least one existing resident ID
- Changing PRIVATE → PUBLIC removes recipient rows
- Changing PUBLIC → PRIVATE requires a non-empty `recipient_ids` set
- Updating PRIVATE with `recipient_ids` **replaces** the full recipient set (transactional)

`author_id` is taken from the JWT principal (not from the request body). New announcements start as `DRAFT`.

### Publish

`PATCH /api/v1/announcements/:id/status` with `{"status":"PUBLISHED"}`.

Only `DRAFT → PUBLISHED` is allowed. Sets `published_at`. Republishing returns 409.

## Finance API

All finance endpoints require JWT auth and role `SUPER_ADMIN` or `ADMIN_RW`.

- `GET /api/v1/finance/transactions?page=1&page_size=20&type=&category=&search=&from=&to=`
- `GET /api/v1/finance/transactions/:id`
- `POST /api/v1/finance/transactions`
- `PATCH /api/v1/finance/transactions/:id`
- `DELETE /api/v1/finance/transactions/:id`
- `GET /api/v1/finance/summary?from=&to=`

Amounts are IDR integers (`int64` / JSON number without decimals), e.g. `"amount": 500000`.

### Date filters (`from` / `to`)

Use `YYYY-MM-DD` interpreted as **Asia/Jakarta** calendar days:

- `from` → `created_at >= from 00:00 Asia/Jakarta`
- `to` → `created_at < (to + 1 day) 00:00 Asia/Jakarta` (inclusive full day)

`from > to` returns `400 INVALID_REQUEST`.

### Summary

Returns PostgreSQL aggregates:

```json
{"data":{"total_income":5000000,"total_expense":3200000,"balance":1800000}}
```

Empty result sets return zeros (not null).

## Dues API

All dues endpoints require JWT auth and role `SUPER_ADMIN` or `ADMIN_RW`.

### Periods

- `GET /api/v1/dues/periods?page=1&page_size=20&year=&month=&half=`
- `GET /api/v1/dues/periods/:id`
- `POST /api/v1/dues/periods`
- `PATCH /api/v1/dues/periods/:id`
- `DELETE /api/v1/dues/periods/:id`
- `GET /api/v1/dues/periods/:id/summary`
- `GET /api/v1/dues/periods/:id/status?page=1&page_size=20&status=&search=`

A period is `year + month + half` (half `1` = first half, `2` = second half). Amount is the expected dues per resident for that period (`int64`, no floats).

Duplicate `year + month + half` returns `409 DUES_PERIOD_ALREADY_EXISTS`.

**Cascade warning:** deleting a period permanently deletes all associated payments (`ON DELETE CASCADE`). There is no soft delete.

### Payments

- `GET /api/v1/dues/payments?page=1&page_size=20&period_id=&resident_id=&search=&from=&to=`
- `GET /api/v1/dues/payments/:id`
- `POST /api/v1/dues/payments`
- `PATCH /api/v1/dues/payments/:id`
- `DELETE /api/v1/dues/payments/:id`

Payment amount **must equal** the period amount. Mismatch returns `400 INVALID_PAYMENT_AMOUNT`. Partial payments are not supported.

One payment per resident per period (`UNIQUE(period_id, resident_id)`). Duplicates return `409 DUES_PAYMENT_ALREADY_EXISTS`.

`paid_at` is RFC3339. List filters `from` / `to` use `YYYY-MM-DD` Asia/Jakarta calendar-day bounds on `paid_at` (same semantics as finance).

### Payment status (derived)

Status is **not stored**. Existence of a payment row means `PAID`; absence means `UNPAID`.

`GET /api/v1/dues/periods/:id/status` returns each resident with derived status via `LEFT JOIN` (supports `status=PAID|UNPAID` and `search` on resident name).

### Period summary

`GET /api/v1/dues/periods/:id/summary` aggregates in PostgreSQL:

```json
{
  "data": {
    "period_id": "...",
    "year": 2026,
    "month": 9,
    "half": 1,
    "expected_amount": 50000,
    "resident_count": 100,
    "paid_count": 70,
    "unpaid_count": 30,
    "expected_total": 5000000,
    "collected_total": 3500000,
    "outstanding_total": 1500000
  }
}
```

## Activities API

All activity endpoints require JWT auth and role `SUPER_ADMIN` or `ADMIN_RW`.

- `GET /api/v1/activities?page=1&page_size=20&search=&from=&to=`
- `GET /api/v1/activities/:id`
- `POST /api/v1/activities`
- `PATCH /api/v1/activities/:id`
- `DELETE /api/v1/activities/:id`

### Create example

```json
{
  "name": "Kerja Bakti RT",
  "description": "Kerja bakti membersihkan lingkungan RT",
  "date": "2026-10-10T07:00:00+07:00",
  "reminder_days_before": 2,
  "reminder_time": "19:00",
  "reminder_message": "Pengingat: kerja bakti RT akan dilaksanakan 2 hari lagi."
}
```

`date` is RFC3339 (instant preserved). List filters `from` / `to` use `YYYY-MM-DD` Asia/Jakarta calendar-day bounds on `date`. Ordering: `date ASC`, then `created_at DESC`.

### Reminder configuration (client-writable)

- `reminder_days_before` (default `0`, must be `>= 0`)
- `reminder_time` (`HH:MM` 24-hour, Asia/Jakarta wall clock; optional)
- `reminder_message` (optional)

### Reminder operational state (server-managed)

- `reminder_scheduled_at`
- `reminder_n8n_execution_id`
- `reminder_sent_at`
- `reminder_schedule_version`

Clients cannot set operational fields. On create: version starts at `1`, operational fields are `null`.

When reminder configuration changes on update (`date`, `reminder_days_before`, `reminder_time`, or `reminder_message`):

- `reminder_schedule_version` increments
- operational fields are cleared to `null`

Name/description-only updates leave version and operational state unchanged.

**Scheduling / n8n integration is not implemented in this phase.** These fields prepare for a future scheduler without coupling Activity CRUD to n8n.

## Gallery API

All gallery endpoints require JWT auth and role `SUPER_ADMIN` or `ADMIN_RW`.

- `GET /api/v1/gallery?page=1&page_size=20&search=`
- `GET /api/v1/gallery/:id`
- `POST /api/v1/gallery`
- `POST /api/v1/gallery/upload` — multipart upload (`image`, `caption`, `sort_order`)
- `PATCH /api/v1/gallery/:id`
- `PUT /api/v1/gallery/:id/image` — multipart image replacement with optional metadata
- `DELETE /api/v1/gallery/:id`

```json
{
  "image_url": "https://example.com/image.jpg",
  "storage_path": "gallery/2026/image.jpg",
  "caption": "Kerja bakti warga",
  "sort_order": 1
}
```

The JSON endpoint remains available for existing externally hosted images. Managed uploads accept JPEG, PNG, or WebP and default to a 10 MB limit. Supabase object paths are server-generated; the service role key never reaches the browser. Default order: `sort_order ASC`, `created_at DESC`.

### Supabase setup and migration

1. Create a Supabase project and copy the direct or session-pooler PostgreSQL URL. Use `sslmode=require`.
2. Copy the project URL and service-role key from Supabase project settings into the backend `.env`.
3. Create the public bucket:

```powershell
.\scripts\setup-supabase-gallery.ps1 `
  -SupabaseUrl $env:SUPABASE_URL `
  -ServerKey $env:SUPABASE_SECRET_KEY
```

4. Back up the local database, apply repository migrations, replace migration seed rows with all local data, and print source/target row counts:

```powershell
.\scripts\migrate-to-supabase.ps1 `
  -SourceDatabaseUrl "postgres://raiseup:raiseup@127.0.0.1:5433/raiseup?sslmode=disable" `
  -TargetDatabaseUrl "postgresql://postgres.PROJECT:PASSWORD@POOLER:5432/postgres?sslmode=require"
```

The script writes ignored rollback artifacts to `tmp/supabase-migration/`. Keep `source-full.dump` until login, CRUD, and gallery upload/replace/delete smoke tests pass. For a local Go process set `DATABASE_URL` to the target URL; for Docker Compose set `DOCKER_DATABASE_URL`.

## Site Settings API

Singleton resource (seeded in migrations). Protected with `SUPER_ADMIN` / `ADMIN_RW`.

- `GET /api/v1/site-settings`
- `PATCH /api/v1/site-settings`

No `POST`. All PATCH fields optional; strings are trimmed. Nullable URL-like fields: `chairman_photo_url`, `chairman_photo_storage_path`, `maps_url`, `embed_url`, `whatsapp_url`. Missing singleton → `404 SITE_SETTINGS_NOT_FOUND`.

## Village Profile & Officials API

Protected with `SUPER_ADMIN` / `ADMIN_RW`.

Profile (singleton, seeded):

- `GET /api/v1/village-profile`
- `PATCH /api/v1/village-profile`

```json
{ "history": "...", "vision": "...", "mission": "..." }
```

Officials (belong to the singleton profile; client does not send `profile_id`):

- `GET /api/v1/village-profile/officials`
- `POST /api/v1/village-profile/officials`
- `PATCH /api/v1/village-profile/officials/:id`
- `DELETE /api/v1/village-profile/officials/:id`

```json
{
  "name": "Budi Santoso",
  "position": "Ketua RT",
  "photo_url": "https://example.com/budi.jpg",
  "sort_order": 1
}
```

`name` and `position` required. Order: `sort_order ASC`, then `id ASC`.

## Dashboard API

Protected with JWT auth and role `SUPER_ADMIN` or `ADMIN_RW`.

- `GET /api/v1/dashboard/summary`

Returns pre-aggregated admin dashboard metrics (no query parameters). Empty domains return zeros; a missing dues period returns `"period": null` with zero dues totals.

### Metrics

| Section | Fields | Meaning |
|---------|--------|---------|
| `residents` | `total`, `male`, `female` | Counts from `residents` (`LAKI_LAKI` / `PEREMPUAN`) |
| `complaints` | `total`, `baru`, `diproses`, `selesai`, `ditolak` | Counts by complaint status |
| `finance` | `income`, `expense`, `balance` | IDR `int64` sums; `balance = income - expense` |
| `dues` | `period`, counts, money totals | Current period = latest `year DESC, month DESC, half DESC` |
| `announcements` | `draft`, `published` | Counts by announcement status |
| `activities` | `upcoming`, `total` | `upcoming` = `date >= NOW()` |

Dues paid/unpaid is derived from payment row existence for the current period (no status column).

Example:

```json
{
  "data": {
    "residents": {"total": 125, "male": 64, "female": 61},
    "complaints": {"total": 18, "baru": 3, "diproses": 5, "selesai": 8, "ditolak": 2},
    "finance": {"income": 12500000, "expense": 4300000, "balance": 8200000},
    "dues": {
      "period": {"year": 2026, "month": 9, "half": 2},
      "resident_count": 125,
      "paid_count": 100,
      "unpaid_count": 25,
      "expected_total": 6250000,
      "collected_total": 5000000,
      "outstanding_total": 1250000
    },
    "announcements": {"draft": 2, "published": 8},
    "activities": {"upcoming": 4, "total": 12}
  }
}
```

## Run with Docker Compose

From the repository root:

```bash
cp .env.example .env
docker compose up --build
```

This starts:

1. PostgreSQL
2. Database migrations (`migrate` service)
3. Go API (container port `8080`, host port via `API_HOST_PORT`, default `8080`)

Verify:

```bash
curl http://localhost:8080/health
```

If host port `8080` is already in use, set `API_HOST_PORT=8081` in `.env` and curl that port instead.

Expected response:

```json
{"status":"ok"}
```

Stop:

```bash
docker compose down
```

## Run Locally (without Docker for the API)

1. Start PostgreSQL (Docker is fine for the database only):

```bash
docker compose up postgres -d
```

2. Copy env and point `DATABASE_URL` at localhost:

```bash
cp .env.example .env
```

3. Run migrations (see below).

4. Start the API:

```bash
cd backend
go run ./cmd/api
```

## Database Design

Schema decisions for Phase 2 are documented in [`docs/database-design.md`](docs/database-design.md).

Domain migration: `backend/db/migrations/000002_domain_schema.{up,down}.sql`.

## Database Migrations

Migrations live in `backend/db/migrations` and are applied with golang-migrate.

**Via Docker Compose** (automatic on `docker compose up`):

The `migrate` service runs `up` against Postgres before the API starts.

**Via Compose one-off:**

```bash
docker compose run --rm migrate
```

**Locally (if migrate CLI is installed):**

```bash
migrate -path backend/db/migrations -database "postgres://raiseup:raiseup@localhost:5432/raiseup?sslmode=disable" up
```

Useful commands:

```bash
# Apply all up migrations
migrate -path backend/db/migrations -database "$DATABASE_URL" up

# Roll back one step
migrate -path backend/db/migrations -database "$DATABASE_URL" down 1

# Check version
migrate -path backend/db/migrations -database "$DATABASE_URL" version
```

## sqlc Generation

Query definitions live under `backend/db/queries` (by domain). Generated code is written to `backend/db/generated`.

From the `backend` directory:

```bash
cd backend
sqlc generate
```

If the sqlc CLI is not installed locally, use Docker:

```bash
cd backend
docker run --rm -v "${PWD}:/src" -w /src sqlc/sqlc:1.29.0 generate
```

Configuration: `backend/sqlc.yaml`.

## Backend Development Commands

```bash
cd backend

# Format
gofmt -w .

# Vet
go vet ./...

# Run API
go run ./cmd/api

# Build binary
go build -o bin/api ./cmd/api
```

## Notes

- Domain modules through dashboard summary are implemented on the backend.
- Frontend Phase 12A provides auth, admin shell, and dashboard.
- Frontend Phase 12B adds residents and complaints admin UI.
- Frontend Phase 12C adds announcements and activities admin UI.
- Frontend Phase 12D adds finance and dues admin UI.
- Frontend Phase 12E adds gallery, site settings, village profile, and officials admin UI.
- Frontend Phase 12F hardens admin UX: error boundary, 404, auth session cleanup, and consistent polish.
- Admin chat supports DIRECT/GROUP between admins and WHATSAPP threads with residents via Meta Cloud API.
- WhatsApp webhook: `GET/POST /api/v1/webhooks/whatsapp` (public). Notifications test: `POST /api/v1/whatsapp/notifications`.
- Redis, n8n, and AI remain out of scope. Activity reminder scheduling can later call the WhatsApp notification service.
