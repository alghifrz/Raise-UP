ALTER TABLE activities
    ALTER COLUMN date TYPE DATE
    USING ((date AT TIME ZONE 'Asia/Jakarta')::date);

ALTER TABLE activities
    DROP COLUMN IF EXISTS reminder_schedule_version;
