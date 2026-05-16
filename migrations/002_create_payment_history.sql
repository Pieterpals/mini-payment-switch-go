-- 002_create_payment_history.sql
-- Audit trail for all payment state changes

CREATE TABLE IF NOT EXISTS payment_history (
    id          BIGSERIAL       PRIMARY KEY,
    trx_id      VARCHAR(64)     NOT NULL,
    status      VARCHAR(16)     NOT NULL,
    action      VARCHAR(32)     NOT NULL,
    detail      TEXT,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Fast lookup by transaction ID
CREATE INDEX IF NOT EXISTS idx_payment_history_trx_id ON payment_history(trx_id);
