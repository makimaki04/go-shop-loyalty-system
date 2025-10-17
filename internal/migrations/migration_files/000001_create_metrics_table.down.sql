DROP INDEX IF EXISTS withdrawals_processed_at_desc_idx;
DROP TABLE IF EXISTS withdrawals;

DROP INDEX IF EXISTS orders_uploaded_at_desc_idx;
DROP TABLE IF EXISTS orders;

DROP TABLE IF EXISTS balances;

DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS order_status;