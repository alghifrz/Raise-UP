-- Domain schema for RAISE UP v2.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ---------------------------------------------------------------------------
-- Shared updated_at trigger
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ---------------------------------------------------------------------------
-- Enums
-- ---------------------------------------------------------------------------
CREATE TYPE user_role AS ENUM ('SUPER_ADMIN', 'ADMIN_RW');
CREATE TYPE gender AS ENUM ('LAKI_LAKI', 'PEREMPUAN');
CREATE TYPE announcement_status AS ENUM ('DRAFT', 'PUBLISHED');
CREATE TYPE announcement_visibility AS ENUM ('PUBLIC', 'PRIVATE');
CREATE TYPE complaint_status AS ENUM ('BARU', 'DIPROSES', 'SELESAI', 'DITOLAK');
CREATE TYPE complaint_urgency AS ENUM ('PRIORITY', 'MEDIUM', 'NORMAL');
CREATE TYPE cash_transaction_type AS ENUM ('INCOME', 'EXPENSE');

-- ---------------------------------------------------------------------------
-- Users
-- ---------------------------------------------------------------------------
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role user_role NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_email_unique UNIQUE (email)
);

CREATE TRIGGER users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- Residents
-- ---------------------------------------------------------------------------
CREATE TABLE residents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    gender gender NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX residents_phone_idx ON residents (phone);
CREATE INDEX residents_name_idx ON residents (name);

CREATE TRIGGER residents_set_updated_at
    BEFORE UPDATE ON residents
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- Announcements
-- ---------------------------------------------------------------------------
CREATE TABLE announcements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    excerpt TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT '',
    visibility announcement_visibility NOT NULL DEFAULT 'PUBLIC',
    status announcement_status NOT NULL DEFAULT 'DRAFT',
    thumbnail_url TEXT,
    author_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX announcements_author_id_idx ON announcements (author_id);
CREATE INDEX announcements_visibility_idx ON announcements (visibility);
CREATE INDEX announcements_status_published_at_idx ON announcements (status, published_at DESC);

CREATE TRIGGER announcements_set_updated_at
    BEFORE UPDATE ON announcements
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE announcement_recipients (
    announcement_id UUID NOT NULL REFERENCES announcements (id) ON DELETE CASCADE,
    resident_id UUID NOT NULL REFERENCES residents (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (announcement_id, resident_id)
);

CREATE INDEX announcement_recipients_resident_id_idx ON announcement_recipients (resident_id);

-- ---------------------------------------------------------------------------
-- Activities
-- ---------------------------------------------------------------------------
CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    date DATE NOT NULL,
    reminder_days_before INTEGER NOT NULL DEFAULT 0,
    reminder_time TIME,
    reminder_message TEXT NOT NULL DEFAULT '',
    reminder_scheduled_at TIMESTAMPTZ,
    reminder_n8n_execution_id TEXT,
    reminder_sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT activities_reminder_days_before_non_negative CHECK (reminder_days_before >= 0)
);

CREATE INDEX activities_date_idx ON activities (date);

CREATE TRIGGER activities_set_updated_at
    BEFORE UPDATE ON activities
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- Complaints
-- Snapshot fields (resident_name, phone) preserve the submitted report even if
-- the linked resident changes or is removed later.
-- ---------------------------------------------------------------------------
CREATE TABLE complaints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ref TEXT NOT NULL,
    resident_id UUID REFERENCES residents (id) ON DELETE SET NULL,
    resident_name TEXT NOT NULL,
    phone TEXT NOT NULL,
    block TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT '',
    urgency complaint_urgency NOT NULL DEFAULT 'NORMAL',
    status complaint_status NOT NULL DEFAULT 'BARU',
    message TEXT NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT complaints_ref_unique UNIQUE (ref)
);

CREATE INDEX complaints_status_idx ON complaints (status);
CREATE INDEX complaints_received_at_idx ON complaints (received_at DESC);
CREATE INDEX complaints_resident_id_idx ON complaints (resident_id);

CREATE TRIGGER complaints_set_updated_at
    BEFORE UPDATE ON complaints
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- Finance
-- ---------------------------------------------------------------------------
CREATE TABLE cash_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type cash_transaction_type NOT NULL,
    title TEXT NOT NULL,
    amount BIGINT NOT NULL,
    category TEXT NOT NULL DEFAULT '',
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT cash_transactions_amount_positive CHECK (amount > 0)
);

CREATE INDEX cash_transactions_type_created_at_idx ON cash_transactions (type, created_at DESC);
CREATE INDEX cash_transactions_category_idx ON cash_transactions (category);
CREATE INDEX cash_transactions_created_at_idx ON cash_transactions (created_at DESC);

CREATE TRIGGER cash_transactions_set_updated_at
    BEFORE UPDATE ON cash_transactions
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- Dues
-- ---------------------------------------------------------------------------
CREATE TABLE dues_periods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,
    half INTEGER NOT NULL,
    amount BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT dues_periods_year_month_half_unique UNIQUE (year, month, half),
    CONSTRAINT dues_periods_month_valid CHECK (month BETWEEN 1 AND 12),
    CONSTRAINT dues_periods_half_valid CHECK (half IN (1, 2)),
    CONSTRAINT dues_periods_amount_positive CHECK (amount > 0)
);

CREATE TABLE dues_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    period_id UUID NOT NULL REFERENCES dues_periods (id) ON DELETE CASCADE,
    resident_id UUID NOT NULL REFERENCES residents (id) ON DELETE RESTRICT,
    amount BIGINT NOT NULL,
    paid_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT dues_payments_period_resident_unique UNIQUE (period_id, resident_id),
    CONSTRAINT dues_payments_amount_positive CHECK (amount > 0)
);

CREATE INDEX dues_payments_resident_id_idx ON dues_payments (resident_id);
CREATE INDEX dues_payments_paid_at_idx ON dues_payments (paid_at DESC);

-- ---------------------------------------------------------------------------
-- Cash reminders
-- ---------------------------------------------------------------------------
CREATE TABLE cash_reminders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    due_date DATE NOT NULL,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX cash_reminders_due_date_idx ON cash_reminders (due_date);

-- ---------------------------------------------------------------------------
-- Content management
-- ---------------------------------------------------------------------------
CREATE TABLE gallery_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    image_url TEXT NOT NULL,
    storage_path TEXT NOT NULL DEFAULT '',
    caption TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX gallery_items_sort_order_idx ON gallery_items (sort_order);

CREATE TABLE site_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_name TEXT NOT NULL DEFAULT '',
    tagline TEXT NOT NULL DEFAULT '',
    chairman_name TEXT NOT NULL DEFAULT '',
    chairman_role TEXT NOT NULL DEFAULT '',
    chairman_quote TEXT NOT NULL DEFAULT '',
    chairman_photo_url TEXT,
    chairman_photo_storage_path TEXT,
    map_title TEXT NOT NULL DEFAULT '',
    map_description TEXT NOT NULL DEFAULT '',
    maps_url TEXT,
    embed_url TEXT,
    address TEXT NOT NULL DEFAULT '',
    phone TEXT NOT NULL DEFAULT '',
    whatsapp_url TEXT,
    footer_blurb TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER site_settings_set_updated_at
    BEFORE UPDATE ON site_settings
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE village_profile (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    history TEXT NOT NULL DEFAULT '',
    vision TEXT NOT NULL DEFAULT '',
    mission TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER village_profile_set_updated_at
    BEFORE UPDATE ON village_profile
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE village_officials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES village_profile (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    position TEXT NOT NULL,
    photo_url TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX village_officials_profile_sort_idx ON village_officials (profile_id, sort_order);

-- ---------------------------------------------------------------------------
-- Singleton seeds
-- ---------------------------------------------------------------------------
INSERT INTO site_settings (site_name, tagline)
VALUES ('Lebak Asri', 'Portal Digital RT / RW Lebak Asri');

INSERT INTO village_profile DEFAULT VALUES;
