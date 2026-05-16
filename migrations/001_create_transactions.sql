-- 001_create_transactions.sql
-- Mini Payment Switch: Transactions table

CREATE TABLE IF NOT EXISTS transactions (
    id            BIGSERIAL PRIMARY KEY,
    trx_id        VARCHAR(64)    NOT NULL UNIQUE,
    account_no    VARCHAR(32)    NOT NULL,
    amount        NUMERIC(18,2)  NOT NULL,
    status        VARCHAR(16)    NOT NULL DEFAULT 'PENDING',
    raw_response  JSONB,
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

-- Index for fast lookup by account
CREATE INDEX IF NOT EXISTS idx_transactions_account_no ON transactions(account_no);

-- Index for status-based queries
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
