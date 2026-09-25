-- wa_contact_name must stay the WhatsApp profile name.
-- Older ingest copied residents.name into that column, which produced
-- labels like "Tes Warga (Tes Warga)" instead of "Smartfren (Tes Warga)".
UPDATE conversations c
SET wa_contact_name = NULL
FROM residents r
WHERE c.type = 'WHATSAPP'
  AND c.resident_id = r.id
  AND c.wa_contact_name IS NOT NULL
  AND lower(btrim(c.wa_contact_name)) = lower(btrim(r.name));
