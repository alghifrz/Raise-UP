# RAISE UP Database Design

Phase 2 schema for the Go + PostgreSQL modular monolith. Redesigned from the previous Prisma model; not a 1:1 conversion.

## Tables

| Table | Purpose |
|---|---|
| `users` | Admin accounts (SUPER_ADMIN, ADMIN_RW) |
| `residents` | Warga registry |
| `announcements` | Public/private announcements |
| `announcement_recipients` | Private announcement targeting (resident ↔ announcement) |
| `activities` | Scheduled activities + reminder metadata (`date` timestamptz, `reminder_schedule_version`) |
| `complaints` | Aduan with submission snapshots |
| `cash_transactions` | Kas income/expense ledger |
| `dues_periods` | Iuran period definitions |
| `dues_payments` | Per-resident payment per period |
| `cash_reminders` | Future kas reminder automation rows |
| `gallery_items` | Gallery media |
| `site_settings` | Singleton portal settings |
| `village_profile` | Singleton RW profile content |
| `village_officials` | Officials under a village profile |

## Important Decisions

### Announcement recipients → relational table

Rejected Prisma-style `recipient_ids TEXT[]` / UUID arrays.

`announcement_recipients (announcement_id, resident_id)` supports:

- targeting without array mutation
- FK integrity to `residents`
- efficient “announcements for resident X” queries

Empty recipient set means “no explicit targets” (PRIVATE announcements may still require recipients at the application layer later).

### Complaints: snapshot fields kept

`resident_name` and `phone` are **required snapshots** of what was submitted.

`resident_id` is **nullable** and links to the registry when known.

Rationale: complaint history must remain intact if the resident is renamed, phone changes, or the resident row is deleted (`ON DELETE SET NULL`).

### Money as `BIGINT` (IDR)

`amount` columns store whole Indonesian Rupiah. No `NUMERIC`/`FLOAT` for money in this schema.

### Enums vs free text

PostgreSQL enums are used only for stable value sets (roles, gender, announcement status/visibility, complaint status/urgency, cash transaction type).

Flexible fields such as `category`, `block`, and activity free-text stay as `TEXT`.

### Singletons

`site_settings` and `village_profile` are application singletons. Each is seeded with one row in the migration. Enforcement is by convention (one row); avoiding awkward boolean PK tricks keeps UUID consistency.

### Dues uniqueness

- `dues_periods`: `UNIQUE (year, month, half)` — same as the previous model.
- `dues_payments`: `UNIQUE (period_id, resident_id)` — one payment record per resident per period for now (no partial payments).

### Timestamps

- Prefer `TIMESTAMPTZ`.
- `updated_at` is maintained by a shared `set_updated_at` trigger where the column exists.
- `dues_periods` and some automation tables only need `created_at`.

## Relationships

```text
users 1──* announcements (author_id)
announcements 1──* announcement_recipients *──1 residents
residents 1──* complaints (optional)
residents 1──* dues_payments
dues_periods 1──* dues_payments
village_profile 1──* village_officials
```

Standalone (no FKs): `activities`, `cash_transactions`, `cash_reminders`, `gallery_items`, `site_settings`.

## ON DELETE Behavior

| FK | Behavior | Why |
|---|---|---|
| `announcements.author_id` → `users` | `RESTRICT` | Keep authorship integrity; delete users intentionally |
| `announcement_recipients.announcement_id` | `CASCADE` | Recipients go with announcement |
| `announcement_recipients.resident_id` | `CASCADE` | Drop targeting if resident removed |
| `complaints.resident_id` | `SET NULL` | Preserve complaint + snapshots |
| `dues_payments.period_id` | `CASCADE` | Payments belong to period |
| `dues_payments.resident_id` | `RESTRICT` | Do not silently erase payment history |
| `village_officials.profile_id` | `CASCADE` | Officials belong to profile |

## Indexes (beyond PK/UNIQUE)

- `residents (phone)`, `residents (name)` — lookup / search
- `announcements (status, published_at DESC)`, `announcements (author_id)`, `announcements (visibility)`
- `announcement_recipients (resident_id)` — reverse lookup
- `complaints (status)`, `complaints (received_at DESC)`, `complaints (resident_id)`
- `cash_transactions (type, created_at DESC)`, `cash_transactions (category)`, `cash_transactions (created_at DESC)`
- `dues_payments (resident_id)`, `dues_payments (paid_at DESC)`
- `activities (date)`, `cash_reminders (due_date)`, `gallery_items (sort_order)`, `village_officials (profile_id, sort_order)`

## Out of Scope

Auth flows, HTTP handlers, n8n execution, WhatsApp, AI, Redis, and frontend storage backends.
