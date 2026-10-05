DROP TABLE IF EXISTS client_balance_adjustments;

DROP INDEX IF EXISTS idx_clients_is_archived;

ALTER TABLE clients
    DROP COLUMN IF EXISTS is_archived;

ALTER TABLE users
    DROP COLUMN IF EXISTS default_rate_group,
    DROP COLUMN IF EXISTS default_rate_pair,
    DROP COLUMN IF EXISTS default_rate_individual;
