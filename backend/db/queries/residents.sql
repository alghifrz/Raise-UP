-- name: CreateResident :one
INSERT INTO residents (name, phone, gender)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetResidentByID :one
SELECT * FROM residents
WHERE id = $1;

-- name: ListResidentsFiltered :many
SELECT *
FROM residents
WHERE
    (
        sqlc.arg(search)::text = ''
        OR name ILIKE '%' || sqlc.arg(search) || '%'
        OR phone ILIKE '%' || sqlc.arg(search) || '%'
    )
    AND (
        sqlc.narg(gender)::gender IS NULL
        OR gender = sqlc.narg(gender)
    )
ORDER BY name ASC, created_at ASC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountResidentsFiltered :one
SELECT count(*)::bigint
FROM residents
WHERE
    (
        sqlc.arg(search)::text = ''
        OR name ILIKE '%' || sqlc.arg(search) || '%'
        OR phone ILIKE '%' || sqlc.arg(search) || '%'
    )
    AND (
        sqlc.narg(gender)::gender IS NULL
        OR gender = sqlc.narg(gender)
    );

-- name: GetResidentByPhone :one
SELECT * FROM residents
WHERE phone = $1;

-- name: UpdateResident :one
UPDATE residents
SET
    name = $2,
    phone = $3,
    gender = $4
WHERE id = $1
RETURNING *;

-- name: DeleteResident :exec
DELETE FROM residents
WHERE id = $1;
