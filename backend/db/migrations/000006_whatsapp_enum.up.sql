-- Must be its own migration: new enum values are not usable in the same
-- transaction that creates them on some Postgres setups.
ALTER TYPE conversation_type ADD VALUE IF NOT EXISTS 'WHATSAPP';
