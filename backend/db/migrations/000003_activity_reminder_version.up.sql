-- Phase 9: activity reminder versioning + preserve activity date as timestamptz.
-- reminder_schedule_version is required for future scheduler invalidation.
-- date was DATE; API accepts RFC3339 and must preserve the instant/offset.

ALTER TABLE activities
    ADD COLUMN reminder_schedule_version INTEGER NOT NULL DEFAULT 1;

ALTER TABLE activities
    ALTER COLUMN date TYPE TIMESTAMPTZ
    USING (date::timestamp AT TIME ZONE 'Asia/Jakarta');
