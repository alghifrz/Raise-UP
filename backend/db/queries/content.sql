-- Gallery

-- name: CreateGalleryItem :one
INSERT INTO gallery_items (image_url, storage_path, caption, sort_order)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetGalleryItemByID :one
SELECT * FROM gallery_items
WHERE id = $1;

-- name: ListGalleryItemsFiltered :many
SELECT *
FROM gallery_items
WHERE
    (
        sqlc.arg(search)::text = ''
        OR caption ILIKE '%' || sqlc.arg(search) || '%'
        OR image_url ILIKE '%' || sqlc.arg(search) || '%'
    )
ORDER BY sort_order ASC, created_at DESC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountGalleryItemsFiltered :one
SELECT COUNT(*)::bigint AS count
FROM gallery_items
WHERE
    (
        sqlc.arg(search)::text = ''
        OR caption ILIKE '%' || sqlc.arg(search) || '%'
        OR image_url ILIKE '%' || sqlc.arg(search) || '%'
    );

-- name: UpdateGalleryItem :one
UPDATE gallery_items
SET
    image_url = COALESCE(sqlc.narg(image_url), image_url),
    storage_path = COALESCE(sqlc.narg(storage_path), storage_path),
    caption = COALESCE(sqlc.narg(caption), caption),
    sort_order = COALESCE(sqlc.narg(sort_order), sort_order)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteGalleryItem :execrows
DELETE FROM gallery_items
WHERE id = $1;

-- Cash reminders (kept for future phases; unused by Phase 10)

-- name: CreateCashReminder :one
INSERT INTO cash_reminders (title, due_date, sent_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetCashReminderByID :one
SELECT * FROM cash_reminders
WHERE id = $1;

-- name: ListCashReminders :many
SELECT * FROM cash_reminders
ORDER BY due_date ASC;

-- name: DeleteCashReminder :exec
DELETE FROM cash_reminders
WHERE id = $1;

-- Site settings (singleton)

-- name: GetSiteSettings :one
SELECT * FROM site_settings
ORDER BY updated_at DESC
LIMIT 1;

-- name: UpdateSiteSettings :one
UPDATE site_settings
SET
    site_name = COALESCE(sqlc.narg(site_name), site_name),
    tagline = COALESCE(sqlc.narg(tagline), tagline),
    chairman_name = COALESCE(sqlc.narg(chairman_name), chairman_name),
    chairman_role = COALESCE(sqlc.narg(chairman_role), chairman_role),
    chairman_quote = COALESCE(sqlc.narg(chairman_quote), chairman_quote),
    chairman_photo_url = COALESCE(sqlc.narg(chairman_photo_url), chairman_photo_url),
    chairman_photo_storage_path = COALESCE(sqlc.narg(chairman_photo_storage_path), chairman_photo_storage_path),
    map_title = COALESCE(sqlc.narg(map_title), map_title),
    map_description = COALESCE(sqlc.narg(map_description), map_description),
    maps_url = COALESCE(sqlc.narg(maps_url), maps_url),
    embed_url = COALESCE(sqlc.narg(embed_url), embed_url),
    address = COALESCE(sqlc.narg(address), address),
    phone = COALESCE(sqlc.narg(phone), phone),
    whatsapp_url = COALESCE(sqlc.narg(whatsapp_url), whatsapp_url),
    footer_blurb = COALESCE(sqlc.narg(footer_blurb), footer_blurb)
WHERE id = sqlc.arg(id)
RETURNING *;

-- Village profile (singleton)

-- name: GetVillageProfile :one
SELECT * FROM village_profile
ORDER BY updated_at DESC
LIMIT 1;

-- name: UpdateVillageProfile :one
UPDATE village_profile
SET
    history = COALESCE(sqlc.narg(history), history),
    vision = COALESCE(sqlc.narg(vision), vision),
    mission = COALESCE(sqlc.narg(mission), mission)
WHERE id = sqlc.arg(id)
RETURNING *;

-- Village officials

-- name: CreateVillageOfficial :one
INSERT INTO village_officials (profile_id, name, position, photo_url, sort_order)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetVillageOfficialByID :one
SELECT * FROM village_officials
WHERE id = $1;

-- name: ListVillageOfficials :many
SELECT *
FROM village_officials
WHERE profile_id = $1
ORDER BY sort_order ASC, id ASC;

-- name: UpdateVillageOfficial :one
UPDATE village_officials
SET
    name = COALESCE(sqlc.narg(name), name),
    position = COALESCE(sqlc.narg(position), position),
    photo_url = COALESCE(sqlc.narg(photo_url), photo_url),
    sort_order = COALESCE(sqlc.narg(sort_order), sort_order)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteVillageOfficial :execrows
DELETE FROM village_officials
WHERE id = $1;
