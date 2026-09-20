-- name: CreateCashTransaction :one
INSERT INTO cash_transactions (type, title, amount, category, note)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetCashTransactionByID :one
SELECT * FROM cash_transactions
WHERE id = $1;

-- name: ListCashTransactionsFiltered :many
SELECT *
FROM cash_transactions
WHERE
    (
        sqlc.narg(type)::cash_transaction_type IS NULL
        OR type = sqlc.narg(type)
    )
    AND (
        sqlc.arg(category)::text = ''
        OR category = sqlc.arg(category)
    )
    AND (
        sqlc.arg(search)::text = ''
        OR title ILIKE '%' || sqlc.arg(search) || '%'
        OR note ILIKE '%' || sqlc.arg(search) || '%'
        OR category ILIKE '%' || sqlc.arg(search) || '%'
    )
    AND (
        sqlc.narg(from_at)::timestamptz IS NULL
        OR created_at >= sqlc.narg(from_at)
    )
    AND (
        sqlc.narg(to_exclusive)::timestamptz IS NULL
        OR created_at < sqlc.narg(to_exclusive)
    )
ORDER BY created_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountCashTransactionsFiltered :one
SELECT count(*)::bigint
FROM cash_transactions
WHERE
    (
        sqlc.narg(type)::cash_transaction_type IS NULL
        OR type = sqlc.narg(type)
    )
    AND (
        sqlc.arg(category)::text = ''
        OR category = sqlc.arg(category)
    )
    AND (
        sqlc.arg(search)::text = ''
        OR title ILIKE '%' || sqlc.arg(search) || '%'
        OR note ILIKE '%' || sqlc.arg(search) || '%'
        OR category ILIKE '%' || sqlc.arg(search) || '%'
    )
    AND (
        sqlc.narg(from_at)::timestamptz IS NULL
        OR created_at >= sqlc.narg(from_at)
    )
    AND (
        sqlc.narg(to_exclusive)::timestamptz IS NULL
        OR created_at < sqlc.narg(to_exclusive)
    );

-- name: UpdateCashTransaction :one
UPDATE cash_transactions
SET
    type = $2,
    title = $3,
    amount = $4,
    category = $5,
    note = $6
WHERE id = $1
RETURNING *;

-- name: DeleteCashTransaction :exec
DELETE FROM cash_transactions
WHERE id = $1;

-- name: GetCashTransactionSummary :one
SELECT
    COALESCE(SUM(CASE WHEN type = 'INCOME' THEN amount ELSE 0 END), 0)::bigint AS total_income,
    COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN amount ELSE 0 END), 0)::bigint AS total_expense
FROM cash_transactions
WHERE
    (
        sqlc.narg(from_at)::timestamptz IS NULL
        OR created_at >= sqlc.narg(from_at)
    )
    AND (
        sqlc.narg(to_exclusive)::timestamptz IS NULL
        OR created_at < sqlc.narg(to_exclusive)
    );
