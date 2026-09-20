-- name: CreateDuesPeriod :one
INSERT INTO dues_periods (year, month, half, amount)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetDuesPeriodByID :one
SELECT * FROM dues_periods
WHERE id = $1;

-- name: ListDuesPeriodsFiltered :many
SELECT *
FROM dues_periods
WHERE
    (sqlc.narg(year)::integer IS NULL OR year = sqlc.narg(year))
    AND (sqlc.narg(month)::integer IS NULL OR month = sqlc.narg(month))
    AND (sqlc.narg(half)::integer IS NULL OR half = sqlc.narg(half))
ORDER BY year DESC, month DESC, half DESC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountDuesPeriodsFiltered :one
SELECT COUNT(*)::bigint AS count
FROM dues_periods
WHERE
    (sqlc.narg(year)::integer IS NULL OR year = sqlc.narg(year))
    AND (sqlc.narg(month)::integer IS NULL OR month = sqlc.narg(month))
    AND (sqlc.narg(half)::integer IS NULL OR half = sqlc.narg(half));

-- name: UpdateDuesPeriod :one
UPDATE dues_periods
SET
    year = COALESCE(sqlc.narg(year), year),
    month = COALESCE(sqlc.narg(month), month),
    half = COALESCE(sqlc.narg(half), half),
    amount = COALESCE(sqlc.narg(amount), amount)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteDuesPeriod :execrows
DELETE FROM dues_periods
WHERE id = $1;

-- name: CreateDuesPayment :one
INSERT INTO dues_payments (period_id, resident_id, amount, paid_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetDuesPaymentByID :one
SELECT
    dp.id,
    dp.period_id,
    dp.resident_id,
    r.name AS resident_name,
    dp.amount,
    dp.paid_at,
    dp.created_at
FROM dues_payments dp
INNER JOIN residents r ON r.id = dp.resident_id
WHERE dp.id = $1;

-- name: ListDuesPaymentsFiltered :many
SELECT
    dp.id,
    dp.period_id,
    dp.resident_id,
    r.name AS resident_name,
    dp.amount,
    dp.paid_at,
    dp.created_at
FROM dues_payments dp
INNER JOIN residents r ON r.id = dp.resident_id
WHERE
    (sqlc.narg(period_id)::uuid IS NULL OR dp.period_id = sqlc.narg(period_id))
    AND (sqlc.narg(resident_id)::uuid IS NULL OR dp.resident_id = sqlc.narg(resident_id))
    AND (
        sqlc.arg(search)::text = ''
        OR r.name ILIKE '%' || sqlc.arg(search) || '%'
    )
    AND (sqlc.narg(from_at)::timestamptz IS NULL OR dp.paid_at >= sqlc.narg(from_at))
    AND (sqlc.narg(to_at)::timestamptz IS NULL OR dp.paid_at < sqlc.narg(to_at))
ORDER BY dp.paid_at DESC, dp.created_at DESC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountDuesPaymentsFiltered :one
SELECT COUNT(*)::bigint AS count
FROM dues_payments dp
INNER JOIN residents r ON r.id = dp.resident_id
WHERE
    (sqlc.narg(period_id)::uuid IS NULL OR dp.period_id = sqlc.narg(period_id))
    AND (sqlc.narg(resident_id)::uuid IS NULL OR dp.resident_id = sqlc.narg(resident_id))
    AND (
        sqlc.arg(search)::text = ''
        OR r.name ILIKE '%' || sqlc.arg(search) || '%'
    )
    AND (sqlc.narg(from_at)::timestamptz IS NULL OR dp.paid_at >= sqlc.narg(from_at))
    AND (sqlc.narg(to_at)::timestamptz IS NULL OR dp.paid_at < sqlc.narg(to_at));

-- name: UpdateDuesPayment :one
UPDATE dues_payments
SET
    period_id = COALESCE(sqlc.narg(period_id), period_id),
    resident_id = COALESCE(sqlc.narg(resident_id), resident_id),
    amount = COALESCE(sqlc.narg(amount), amount),
    paid_at = COALESCE(sqlc.narg(paid_at), paid_at)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteDuesPayment :execrows
DELETE FROM dues_payments
WHERE id = $1;

-- name: GetDuesPeriodSummary :one
SELECT
    p.id AS period_id,
    p.year,
    p.month,
    p.half,
    p.amount AS expected_amount,
    COALESCE(s.resident_count, 0)::bigint AS resident_count,
    COALESCE(s.paid_count, 0)::bigint AS paid_count,
    (COALESCE(s.resident_count, 0) - COALESCE(s.paid_count, 0))::bigint AS unpaid_count,
    (COALESCE(s.resident_count, 0) * p.amount)::bigint AS expected_total,
    COALESCE(s.collected_total, 0)::bigint AS collected_total,
    (
        (COALESCE(s.resident_count, 0) * p.amount) - COALESCE(s.collected_total, 0)
    )::bigint AS outstanding_total
FROM dues_periods p
LEFT JOIN LATERAL (
    SELECT
        COUNT(r.id)::bigint AS resident_count,
        COUNT(dp.id)::bigint AS paid_count,
        COALESCE(SUM(dp.amount), 0)::bigint AS collected_total
    FROM residents r
    LEFT JOIN dues_payments dp
        ON dp.resident_id = r.id
       AND dp.period_id = p.id
) s ON true
WHERE p.id = sqlc.arg(period_id);

-- name: ListDuesResidentPaymentStatus :many
SELECT
    r.id AS resident_id,
    r.name AS resident_name,
    r.phone,
    CASE WHEN dp.id IS NULL THEN 'UNPAID' ELSE 'PAID' END::text AS status,
    dp.id AS payment_id,
    dp.amount,
    dp.paid_at
FROM residents r
LEFT JOIN dues_payments dp
    ON dp.resident_id = r.id
   AND dp.period_id = sqlc.arg(period_id)
WHERE
    (
        sqlc.arg(search)::text = ''
        OR r.name ILIKE '%' || sqlc.arg(search) || '%'
    )
    AND (
        sqlc.arg(status_filter)::text = ''
        OR (sqlc.arg(status_filter) = 'PAID' AND dp.id IS NOT NULL)
        OR (sqlc.arg(status_filter) = 'UNPAID' AND dp.id IS NULL)
    )
ORDER BY r.name ASC, r.created_at ASC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountDuesResidentPaymentStatus :one
SELECT COUNT(*)::bigint AS count
FROM residents r
LEFT JOIN dues_payments dp
    ON dp.resident_id = r.id
   AND dp.period_id = sqlc.arg(period_id)
WHERE
    (
        sqlc.arg(search)::text = ''
        OR r.name ILIKE '%' || sqlc.arg(search) || '%'
    )
    AND (
        sqlc.arg(status_filter)::text = ''
        OR (sqlc.arg(status_filter) = 'PAID' AND dp.id IS NOT NULL)
        OR (sqlc.arg(status_filter) = 'UNPAID' AND dp.id IS NULL)
    );
