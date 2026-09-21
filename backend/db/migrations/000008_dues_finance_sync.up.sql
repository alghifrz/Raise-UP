-- Keep dues payments and finance income synchronized without duplicating rows.

CREATE TABLE dues_finance_links (
    dues_payment_id UUID PRIMARY KEY
        REFERENCES dues_payments (id) ON DELETE CASCADE,
    cash_transaction_id UUID NOT NULL UNIQUE
        REFERENCES cash_transactions (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION sync_dues_payment_to_finance()
RETURNS TRIGGER AS $$
DECLARE
    linked_transaction_id UUID;
    period_year INTEGER;
    period_month INTEGER;
    period_half INTEGER;
    resident_name TEXT;
    transaction_title TEXT;
    transaction_note TEXT;
BEGIN
    SELECT year, month, half
    INTO period_year, period_month, period_half
    FROM dues_periods
    WHERE id = NEW.period_id;

    SELECT name
    INTO resident_name
    FROM residents
    WHERE id = NEW.resident_id;

    transaction_title := format(
        'Iuran %s/%s Tahap %s - %s',
        period_month,
        period_year,
        period_half,
        resident_name
    );
    transaction_note := format(
        'Sinkron otomatis pembayaran iuran warga %s',
        resident_name
    );

    SELECT cash_transaction_id
    INTO linked_transaction_id
    FROM dues_finance_links
    WHERE dues_payment_id = NEW.id;

    IF linked_transaction_id IS NULL THEN
        INSERT INTO cash_transactions (
            type,
            title,
            amount,
            category,
            note,
            created_at,
            updated_at
        )
        VALUES (
            'INCOME',
            transaction_title,
            NEW.amount,
            'IURAN',
            transaction_note,
            NEW.paid_at,
            NOW()
        )
        RETURNING id INTO linked_transaction_id;

        INSERT INTO dues_finance_links (dues_payment_id, cash_transaction_id)
        VALUES (NEW.id, linked_transaction_id);
    ELSE
        UPDATE cash_transactions
        SET
            type = 'INCOME',
            title = transaction_title,
            amount = NEW.amount,
            category = 'IURAN',
            note = transaction_note,
            created_at = NEW.paid_at
        WHERE id = linked_transaction_id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION delete_dues_payment_finance()
RETURNS TRIGGER AS $$
DECLARE
    linked_transaction_id UUID;
BEGIN
    SELECT cash_transaction_id
    INTO linked_transaction_id
    FROM dues_finance_links
    WHERE dues_payment_id = OLD.id;

    IF linked_transaction_id IS NOT NULL THEN
        DELETE FROM cash_transactions
        WHERE id = linked_transaction_id;
    END IF;

    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER dues_payments_sync_finance
    AFTER INSERT OR UPDATE OF period_id, resident_id, amount, paid_at
    ON dues_payments
    FOR EACH ROW
    EXECUTE FUNCTION sync_dues_payment_to_finance();

CREATE TRIGGER dues_payments_delete_finance
    BEFORE DELETE
    ON dues_payments
    FOR EACH ROW
    EXECUTE FUNCTION delete_dues_payment_finance();

-- Backfill existing payments. The trigger creates exactly one linked income row.
UPDATE dues_payments
SET amount = amount;
