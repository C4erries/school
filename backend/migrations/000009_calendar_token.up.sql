-- 000009_calendar_token.up.sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS calendar_token UUID NOT NULL DEFAULT gen_random_uuid();
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_calendar_token ON users (calendar_token);
