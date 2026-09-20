-- Drop domain schema in dependency-safe order.

DROP TABLE IF EXISTS village_officials;
DROP TABLE IF EXISTS village_profile;
DROP TABLE IF EXISTS site_settings;
DROP TABLE IF EXISTS gallery_items;
DROP TABLE IF EXISTS cash_reminders;
DROP TABLE IF EXISTS dues_payments;
DROP TABLE IF EXISTS dues_periods;
DROP TABLE IF EXISTS cash_transactions;
DROP TABLE IF EXISTS complaints;
DROP TABLE IF EXISTS activities;
DROP TABLE IF EXISTS announcement_recipients;
DROP TABLE IF EXISTS announcements;
DROP TABLE IF EXISTS residents;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS cash_transaction_type;
DROP TYPE IF EXISTS complaint_urgency;
DROP TYPE IF EXISTS complaint_status;
DROP TYPE IF EXISTS announcement_visibility;
DROP TYPE IF EXISTS announcement_status;
DROP TYPE IF EXISTS gender;
DROP TYPE IF EXISTS user_role;

DROP FUNCTION IF EXISTS set_updated_at();
