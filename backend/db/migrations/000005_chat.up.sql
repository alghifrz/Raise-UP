-- Chat module: WhatsApp-style conversations between admin users.

CREATE TYPE conversation_type AS ENUM ('DIRECT', 'GROUP');

CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type conversation_type NOT NULL,
    title TEXT,
    created_by UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    last_message_at TIMESTAMPTZ,
    last_message_preview TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT conversations_group_title_chk CHECK (
        (type = 'DIRECT' AND title IS NULL)
        OR (type = 'GROUP' AND title IS NOT NULL AND length(btrim(title)) > 0)
    )
);

CREATE INDEX conversations_last_message_at_idx
    ON conversations (last_message_at DESC NULLS LAST);

CREATE TRIGGER conversations_set_updated_at
    BEFORE UPDATE ON conversations
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE conversation_participants (
    conversation_id UUID NOT NULL REFERENCES conversations (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_read_at TIMESTAMPTZ,
    PRIMARY KEY (conversation_id, user_id)
);

CREATE INDEX conversation_participants_user_id_idx
    ON conversation_participants (user_id);

CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations (id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT messages_body_not_empty CHECK (length(btrim(body)) > 0)
);

CREATE INDEX messages_conversation_created_at_idx
    ON messages (conversation_id, created_at DESC);
