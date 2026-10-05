-- 1. Default rates for users (tutors)
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS default_rate_individual NUMERIC(10, 2) DEFAULT 1500,
    ADD COLUMN IF NOT EXISTS default_rate_pair NUMERIC(10, 2) DEFAULT 1000,
    ADD COLUMN IF NOT EXISTS default_rate_group NUMERIC(10, 2) DEFAULT 700;

-- 2. Archived status for clients
ALTER TABLE clients
    ADD COLUMN IF NOT EXISTS is_archived BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_clients_is_archived ON clients (is_archived);

-- 3. Balance adjustments audit log
CREATE TABLE IF NOT EXISTS client_balance_adjustments (
    id UUID PRIMARY KEY,
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    format VARCHAR(50) NOT NULL CHECK (format IN ('individual', 'pair', 'group')),
    delta_hours NUMERIC(10, 2) NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_client_balance_adjustments_client_id ON client_balance_adjustments (client_id);
CREATE INDEX IF NOT EXISTS idx_client_balance_adjustments_teacher_id ON client_balance_adjustments (teacher_id);
