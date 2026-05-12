-- Удаление таблиц
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS balance_operations;


-- Удаление индексов
DROP INDEX IF EXISTS idx_users_login;
DROP INDEX IF EXISTS idx_orders_user_id;

DROP INDEX IF EXISTS idx_orders_number;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_uploaded_at;

DROP INDEX IF EXISTS idx_balance_operations_user_id;
DROP INDEX IF EXISTS idx_balance_operations_order_number;
DROP INDEX IF EXISTS idx_balance_operations_processed_at;