CREATE TABLE whatsapp_bot_sessions (
    phone TEXT PRIMARY KEY,
    state TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT whatsapp_bot_sessions_state_valid
        CHECK (state IN ('AWAITING_COMPLAINT'))
);

CREATE INDEX whatsapp_bot_sessions_updated_at_idx
    ON whatsapp_bot_sessions (updated_at);
