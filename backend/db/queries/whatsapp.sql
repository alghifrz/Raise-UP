-- name: GetWhatsAppSettings :one
SELECT *
FROM whatsapp_settings
ORDER BY updated_at DESC
LIMIT 1;

-- name: UpdateWhatsAppBotEnabled :one
UPDATE whatsapp_settings
SET bot_enabled = $1
WHERE id = (
    SELECT id
    FROM whatsapp_settings
    ORDER BY updated_at DESC
    LIMIT 1
)
RETURNING *;
