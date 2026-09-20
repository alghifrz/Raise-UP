-- name: CreateComplaint :one
INSERT INTO complaints (
    ref,
    resident_id,
    resident_name,
    phone,
    block,
    category,
    urgency,
    status,
    message
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetComplaintByID :one
SELECT * FROM complaints
WHERE id = $1;

-- name: GetComplaintByRef :one
SELECT * FROM complaints
WHERE ref = $1;

-- name: ListComplaintsFiltered :many
SELECT *
FROM complaints
WHERE
    (
        sqlc.arg(search)::text = ''
        OR ref ILIKE '%' || sqlc.arg(search) || '%'
        OR resident_name ILIKE '%' || sqlc.arg(search) || '%'
        OR phone ILIKE '%' || sqlc.arg(search) || '%'
        OR message ILIKE '%' || sqlc.arg(search) || '%'
    )
    AND (
        sqlc.narg(status)::complaint_status IS NULL
        OR status = sqlc.narg(status)
    )
    AND (
        sqlc.narg(urgency)::complaint_urgency IS NULL
        OR urgency = sqlc.narg(urgency)
    )
    AND (
        sqlc.arg(category)::text = ''
        OR category = sqlc.arg(category)
    )
ORDER BY received_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountComplaintsFiltered :one
SELECT count(*)::bigint
FROM complaints
WHERE
    (
        sqlc.arg(search)::text = ''
        OR ref ILIKE '%' || sqlc.arg(search) || '%'
        OR resident_name ILIKE '%' || sqlc.arg(search) || '%'
        OR phone ILIKE '%' || sqlc.arg(search) || '%'
        OR message ILIKE '%' || sqlc.arg(search) || '%'
    )
    AND (
        sqlc.narg(status)::complaint_status IS NULL
        OR status = sqlc.narg(status)
    )
    AND (
        sqlc.narg(urgency)::complaint_urgency IS NULL
        OR urgency = sqlc.narg(urgency)
    )
    AND (
        sqlc.arg(category)::text = ''
        OR category = sqlc.arg(category)
    );

-- name: UpdateComplaint :one
UPDATE complaints
SET
    resident_id = $2,
    resident_name = $3,
    phone = $4,
    block = $5,
    category = $6,
    urgency = $7,
    message = $8
WHERE id = $1
RETURNING *;

-- name: UpdateComplaintStatus :one
UPDATE complaints
SET status = $2
WHERE id = $1
RETURNING *;

-- name: DeleteComplaint :exec
DELETE FROM complaints
WHERE id = $1;
