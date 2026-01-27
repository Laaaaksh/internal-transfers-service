-- Create idempotency_keys table for handling duplicate requests
CREATE TABLE IF NOT EXISTS idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    response_status INT NOT NULL,
    response_body JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index for cleanup queries (to delete old keys)
CREATE INDEX IF NOT EXISTS idx_idempotency_created_at ON idempotency_keys(created_at);

-- Add comments for documentation
COMMENT ON TABLE idempotency_keys IS 'Stores idempotency keys to prevent duplicate transaction processing';
COMMENT ON COLUMN idempotency_keys.key IS 'Client-provided idempotency key from X-Idempotency-Key header';
