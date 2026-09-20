-- name: CreateActivity :one
INSERT INTO activities (
    name,
    description,
    date,
    reminder_days_before,
    reminder_time,
    reminder_message,
    reminder_schedule_version,
    reminder_scheduled_at,
    reminder_n8n_execution_id,
    reminder_sent_at
)
VALUES (
    $1, $2, $3, $4, $5, $6,
    1,
    NULL,
    NULL,
    NULL
)
RETURNING *;

-- name: GetActivityByID :one
SELECT * FROM activities
WHERE id = $1;

-- name: ListActivitiesFiltered :many
SELECT *
FROM activities
WHERE
    (
        sqlc.arg(search)::text = ''
        OR name ILIKE '%' || sqlc.arg(search) || '%'
        OR description ILIKE '%' || sqlc.arg(search) || '%'
    )
    AND (sqlc.narg(from_at)::timestamptz IS NULL OR date >= sqlc.narg(from_at))
    AND (sqlc.narg(to_at)::timestamptz IS NULL OR date < sqlc.narg(to_at))
ORDER BY date ASC, created_at DESC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountActivitiesFiltered :one
SELECT COUNT(*)::bigint AS count
FROM activities
WHERE
    (
        sqlc.arg(search)::text = ''
        OR name ILIKE '%' || sqlc.arg(search) || '%'
        OR description ILIKE '%' || sqlc.arg(search) || '%'
    )
    AND (sqlc.narg(from_at)::timestamptz IS NULL OR date >= sqlc.narg(from_at))
    AND (sqlc.narg(to_at)::timestamptz IS NULL OR date < sqlc.narg(to_at));

-- name: UpdateActivity :one
UPDATE activities
SET
    name = COALESCE(sqlc.narg(name), name),
    description = COALESCE(sqlc.narg(description), description),
    date = COALESCE(sqlc.narg(date), date),
    reminder_days_before = COALESCE(sqlc.narg(reminder_days_before), reminder_days_before),
    reminder_time = COALESCE(sqlc.narg(reminder_time), reminder_time),
    reminder_message = COALESCE(sqlc.narg(reminder_message), reminder_message),
    reminder_schedule_version = CASE
        WHEN sqlc.arg(reminder_config_changed)::boolean THEN reminder_schedule_version + 1
        ELSE reminder_schedule_version
    END,
    reminder_scheduled_at = CASE
        WHEN sqlc.arg(reminder_config_changed)::boolean THEN NULL
        ELSE reminder_scheduled_at
    END,
    reminder_n8n_execution_id = CASE
        WHEN sqlc.arg(reminder_config_changed)::boolean THEN NULL
        ELSE reminder_n8n_execution_id
    END,
    reminder_sent_at = CASE
        WHEN sqlc.arg(reminder_config_changed)::boolean THEN NULL
        ELSE reminder_sent_at
    END
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteActivity :execrows
DELETE FROM activities
WHERE id = $1;
