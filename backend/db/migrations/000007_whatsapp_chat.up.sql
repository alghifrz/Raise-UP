-- WhatsApp Cloud API chat columns and message sender kinds.

ALTER TABLE conversations
    ALTER COLUMN created_by DROP NOT NULL;

ALTER TABLE conversations
    ADD COLUMN IF NOT EXISTS wa_contact_phone TEXT,
    ADD COLUMN IF NOT EXISTS wa_contact_name TEXT,
    ADD COLUMN IF NOT EXISTS resident_id UUID REFERENCES residents (id) ON DELETE SET NULL;

ALTER TABLE conversations
    DROP CONSTRAINT IF EXISTS conversations_group_title_chk;

ALTER TABLE conversations
    ADD CONSTRAINT conversations_type_fields_chk CHECK (
        (
            type = 'DIRECT'
            AND title IS NULL
            AND wa_contact_phone IS NULL
        )
        OR (
            type = 'GROUP'
            AND title IS NOT NULL
            AND length(btrim(title)) > 0
            AND wa_contact_phone IS NULL
        )
        OR (
            type = 'WHATSAPP'
            AND wa_contact_phone IS NOT NULL
            AND length(btrim(wa_contact_phone)) > 0
        )
    );

CREATE UNIQUE INDEX IF NOT EXISTS conversations_whatsapp_phone_uidx
    ON conversations (wa_contact_phone)
    WHERE type = 'WHATSAPP' AND wa_contact_phone IS NOT NULL;

CREATE INDEX IF NOT EXISTS conversations_resident_id_idx
    ON conversations (resident_id)
    WHERE resident_id IS NOT NULL;

CREATE TYPE message_sender_kind AS ENUM ('USER', 'CONTACT', 'SYSTEM');

ALTER TABLE messages
    ALTER COLUMN sender_id DROP NOT NULL;

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS sender_kind message_sender_kind NOT NULL DEFAULT 'USER',
    ADD COLUMN IF NOT EXISTS wa_message_id TEXT,
    ADD COLUMN IF NOT EXISTS wa_status TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS messages_wa_message_id_uidx
    ON messages (wa_message_id)
    WHERE wa_message_id IS NOT NULL;

ALTER TABLE messages
    ADD CONSTRAINT messages_sender_kind_chk CHECK (
        (sender_kind = 'USER' AND sender_id IS NOT NULL)
        OR (sender_kind IN ('CONTACT', 'SYSTEM'))
    );
