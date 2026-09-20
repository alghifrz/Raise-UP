-- name: GetDashboardResidentSummary :one
SELECT
    COUNT(*)::bigint AS total,
    COUNT(*) FILTER (WHERE gender = 'LAKI_LAKI')::bigint AS male,
    COUNT(*) FILTER (WHERE gender = 'PEREMPUAN')::bigint AS female
FROM residents;

-- name: GetDashboardComplaintSummary :one
SELECT
    COUNT(*)::bigint AS total,
    COUNT(*) FILTER (WHERE status = 'BARU')::bigint AS baru,
    COUNT(*) FILTER (WHERE status = 'DIPROSES')::bigint AS diproses,
    COUNT(*) FILTER (WHERE status = 'SELESAI')::bigint AS selesai,
    COUNT(*) FILTER (WHERE status = 'DITOLAK')::bigint AS ditolak
FROM complaints;

-- name: GetDashboardFinanceSummary :one
SELECT
    COALESCE(SUM(amount) FILTER (WHERE type = 'INCOME'), 0)::bigint AS income,
    COALESCE(SUM(amount) FILTER (WHERE type = 'EXPENSE'), 0)::bigint AS expense,
    (
        COALESCE(SUM(amount) FILTER (WHERE type = 'INCOME'), 0)
        - COALESCE(SUM(amount) FILTER (WHERE type = 'EXPENSE'), 0)
    )::bigint AS balance
FROM cash_transactions;

-- name: GetDashboardCurrentDuesSummary :one
WITH current_period AS (
    SELECT id, year, month, half, amount
    FROM dues_periods
    ORDER BY year DESC, month DESC, half DESC
    LIMIT 1
)
SELECT
    p.year,
    p.month,
    p.half,
    COALESCE(s.resident_count, 0)::bigint AS resident_count,
    COALESCE(s.paid_count, 0)::bigint AS paid_count,
    COALESCE(s.unpaid_count, 0)::bigint AS unpaid_count,
    COALESCE(s.expected_total, 0)::bigint AS expected_total,
    COALESCE(s.collected_total, 0)::bigint AS collected_total,
    COALESCE(s.outstanding_total, 0)::bigint AS outstanding_total
FROM (SELECT 1) AS anchor
LEFT JOIN current_period p ON true
LEFT JOIN LATERAL (
    SELECT
        COUNT(r.id)::bigint AS resident_count,
        COUNT(dp.id)::bigint AS paid_count,
        (COUNT(r.id) - COUNT(dp.id))::bigint AS unpaid_count,
        (COUNT(r.id) * p.amount)::bigint AS expected_total,
        COALESCE(SUM(dp.amount), 0)::bigint AS collected_total,
        ((COUNT(r.id) * p.amount) - COALESCE(SUM(dp.amount), 0))::bigint AS outstanding_total
    FROM residents r
    LEFT JOIN dues_payments dp
        ON dp.resident_id = r.id
       AND dp.period_id = p.id
    WHERE p.id IS NOT NULL
) s ON true;

-- name: GetDashboardAnnouncementSummary :one
SELECT
    COUNT(*) FILTER (WHERE status = 'DRAFT')::bigint AS draft,
    COUNT(*) FILTER (WHERE status = 'PUBLISHED')::bigint AS published
FROM announcements;

-- name: GetDashboardActivitySummary :one
SELECT
    COUNT(*) FILTER (WHERE date >= NOW())::bigint AS upcoming,
    COUNT(*)::bigint AS total
FROM activities;
