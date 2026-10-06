-- 000009_calendar_token.down.sql
DROP INDEX IF EXISTS idx_users_calendar_token;
ALTER TABLE users DROP COLUMN IF EXISTS calendar_token;
