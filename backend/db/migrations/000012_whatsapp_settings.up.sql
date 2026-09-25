-- Runtime chatbot toggle for the admin Chat page.
-- Messages still ingest when this is off; only auto-replies stop.

CREATE TABLE whatsapp_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bot_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER whatsapp_settings_set_updated_at
    BEFORE UPDATE ON whatsapp_settings
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

INSERT INTO whatsapp_settings (bot_enabled) VALUES (TRUE);
