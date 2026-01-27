-- Create accounts table
CREATE TABLE IF NOT EXISTS accounts (
    account_id BIGINT PRIMARY KEY,
    balance DECIMAL(19, 8) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT positive_balance CHECK (balance >= 0)
);

-- Create index on updated_at for potential cleanup queries
CREATE INDEX IF NOT EXISTS idx_accounts_updated_at ON accounts(updated_at);

-- Add comment for documentation
COMMENT ON TABLE accounts IS 'Stores account balances for internal fund transfers';
COMMENT ON COLUMN accounts.balance IS 'Account balance with 8 decimal places precision';
