CREATE TABLE IF NOT EXISTS accounts (
    account    VARCHAR(50) PRIMARY KEY,
    full_name  VARCHAR(255) NOT NULL,
    balance    NUMERIC(18,2) NOT NULL DEFAULT 0,
    is_active  BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS payments (
    txn_id      VARCHAR(50) PRIMARY KEY,
    account     VARCHAR(50) NOT NULL REFERENCES accounts(account),
    amount      NUMERIC(18,2) NOT NULL,
    prv_txn     VARCHAR(50),
    status      VARCHAR(20) NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT now()
);

-- тестовые данные для ручной проверки (check/pay через Postman)
INSERT INTO accounts (account, full_name, balance, is_active) VALUES
    ('992918400400', 'Usmonalizoda Suhrob Usmonali', 1000.00, true),
    ('992900000001', 'Test Client Two', 500.00, true),
    ('992900000002', 'Blocked Client', 0.00, false)
ON CONFLICT (account) DO NOTHING;