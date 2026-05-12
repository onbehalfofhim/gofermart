-- migrations/000001_create_gofermart_tables.up.sql


-- Создание таблицы для хранения информации о пользователе
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    login VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    balance NUMERIC(10,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Создание таблицы для хранения информации о заказах
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY,
    number VARCHAR(255) UNIQUE NOT NULL,
    user_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'NEW',
    accrual DECIMAL(10,2) DEFAULT 0,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT fk_orders_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Создание таблицы операций с балансом пользователя
CREATE TABLE IF NOT EXISTS balance_operations (
    id UUID PRIMARY KEY,
    order_number VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL,
    operation_type VARCHAR(20) NOT NULL, -- 'CREDIT' или 'DEBIT'
    amount DECIMAL(10,2) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT fk_balance_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Создание индексов для оптимизации запросов
CREATE INDEX idx_users_login ON users(login);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_number ON orders(number);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_uploaded_at ON orders(uploaded_at);

CREATE INDEX idx_balance_operations_user_id ON balance_operations(user_id);
CREATE INDEX idx_balance_operations_order_number ON balance_operations(order_number);
CREATE INDEX idx_balance_operations_processed_at ON balance_operations(processed_at);