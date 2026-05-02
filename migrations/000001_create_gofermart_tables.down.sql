-- Удаление таблиц
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS orders;


-- Удаление индексов
DROP INDEX IF EXISTS idx_users_login;
DROP INDEX IF EXISTS idx_orders_user_id
DROP INDEX IF EXISTS idx_orders_number
DROP INDEX IF EXISTS idx_orders_status
DROP INDEX IF EXISTS idx_orders_uploaded_at