CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE balances (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    current DECIMAL(20, 2) NOT NULL DEFAULT 0,
    withdrawn DECIMAL(20, 2) NOT NULL DEFAULT 0
);

create TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    number TEXT NOT NULL UNIQUE CHECK (number ~ '^\d+$'),
    status order_status NOT NULL DEFAULT 'NEW',
    accrual NUMERIC(20, 2),
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX orders_uploaded_at_desc_idx ON orders(user_id, uploaded_at DESC);

CREATE TABLE withdrawals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    order_number TEXT NOT NULL UNIQUE CHECK (order_number ~ '^\d+$'),
    amount DECIMAL(20, 2) NOT NULL CHECK (amount > 0),
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX withdrawals_processed_at_desc_idx ON withdrawals(user_id, processed_at DESC);