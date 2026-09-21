-- Make dues synchronization triggers safe when callers use an empty search_path,
-- including pg_dump data restores.
ALTER FUNCTION sync_dues_payment_to_finance()
    SET search_path = public, pg_catalog;

ALTER FUNCTION delete_dues_payment_finance()
    SET search_path = public, pg_catalog;
