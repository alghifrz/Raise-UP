ALTER TABLE messages
    DROP CONSTRAINT IF EXISTS messages_sender_kind_chk;

DROP INDEX IF EXISTS messages_wa_message_id_uidx;

ALTER TABLE messages
    DROP COLUMN IF EXISTS wa_status,
    DROP COLUMN IF EXISTS wa_message_id,
    DROP COLUMN IF EXISTS sender_kind;

DROP TYPE IF EXISTS message_sender_kind;

-- Restore NOT NULL only if no null sender_id rows remain.
UPDATE messages SET sender_id = (
    SELECT id FROM users ORDER BY created_at ASC LIMIT 1
) WHERE sender_id IS NULL;

ALTER TABLE messages
    ALTER COLUMN sender_id SET NOT NULL;

DROP INDEX IF EXISTS conversations_resident_id_idx;
DROP INDEX IF EXISTS conversations_whatsapp_phone_uidx;

ALTER TABLE conversations
    DROP CONSTRAINT IF EXISTS conversations_type_fields_chk;

ALTER TABLE conversations
    DROP COLUMN IF EXISTS resident_id,
    DROP COLUMN IF EXISTS wa_contact_name,
    DROP COLUMN IF EXISTS wa_contact_phone;

-- created_by may still be null for historical WA rows; backfill then restore.
UPDATE conversations SET created_by = (
    SELECT id FROM users ORDER BY created_at ASC LIMIT 1
) WHERE created_by IS NULL;

ALTER TABLE conversations
    ALTER COLUMN created_by SET NOT NULL;

ALTER TABLE conversations
    ADD CONSTRAINT conversations_group_title_chk CHECK (
        (type = 'DIRECT' AND title IS NULL)
        OR (type = 'GROUP' AND title IS NOT NULL AND length(btrim(title)) > 0)
    );
