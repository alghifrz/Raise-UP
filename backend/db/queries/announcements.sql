-- name: CreateAnnouncement :one
INSERT INTO announcements (
    title,
    excerpt,
    body,
    category,
    visibility,
    status,
    thumbnail_url,
    author_id,
    published_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetAnnouncementByID :one
SELECT * FROM announcements
WHERE id = $1;

-- name: ListAnnouncementsFiltered :many
SELECT *
FROM announcements
WHERE
    (
        sqlc.arg(search)::text = ''
        OR title ILIKE '%' || sqlc.arg(search) || '%'
        OR excerpt ILIKE '%' || sqlc.arg(search) || '%'
        OR body ILIKE '%' || sqlc.arg(search) || '%'
        OR category ILIKE '%' || sqlc.arg(search) || '%'
    )
    AND (
        sqlc.narg(status)::announcement_status IS NULL
        OR status = sqlc.narg(status)
    )
    AND (
        sqlc.narg(visibility)::announcement_visibility IS NULL
        OR visibility = sqlc.narg(visibility)
    )
    AND (
        sqlc.arg(category)::text = ''
        OR category = sqlc.arg(category)
    )
ORDER BY updated_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountAnnouncementsFiltered :one
SELECT count(*)::bigint
FROM announcements
WHERE
    (
        sqlc.arg(search)::text = ''
        OR title ILIKE '%' || sqlc.arg(search) || '%'
        OR excerpt ILIKE '%' || sqlc.arg(search) || '%'
        OR body ILIKE '%' || sqlc.arg(search) || '%'
        OR category ILIKE '%' || sqlc.arg(search) || '%'
    )
    AND (
        sqlc.narg(status)::announcement_status IS NULL
        OR status = sqlc.narg(status)
    )
    AND (
        sqlc.narg(visibility)::announcement_visibility IS NULL
        OR visibility = sqlc.narg(visibility)
    )
    AND (
        sqlc.arg(category)::text = ''
        OR category = sqlc.arg(category)
    );

-- name: UpdateAnnouncement :one
UPDATE announcements
SET
    title = $2,
    excerpt = $3,
    body = $4,
    category = $5,
    visibility = $6,
    thumbnail_url = $7
WHERE id = $1
RETURNING *;

-- name: PublishAnnouncement :one
UPDATE announcements
SET
    status = 'PUBLISHED',
    published_at = $2
WHERE id = $1
RETURNING *;

-- name: DeleteAnnouncement :exec
DELETE FROM announcements
WHERE id = $1;

-- name: ListAnnouncementRecipientIDs :many
SELECT resident_id
FROM announcement_recipients
WHERE announcement_id = $1
ORDER BY resident_id;

-- name: DeleteAnnouncementRecipients :exec
DELETE FROM announcement_recipients
WHERE announcement_id = $1;

-- name: AddAnnouncementRecipient :exec
INSERT INTO announcement_recipients (announcement_id, resident_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;
