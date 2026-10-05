-- 000005_hourly_rates_and_format_subscriptions.down.sql

DROP TABLE IF EXISTS client_tags;
DROP TABLE IF EXISTS tags;

ALTER TABLE lessons DROP CONSTRAINT IF EXISTS lessons_format_check;
UPDATE lessons SET format = 'offline' WHERE format NOT IN ('online', 'offline');
ALTER TABLE lessons ADD CONSTRAINT lessons_format_check CHECK (format IN ('online', 'offline'));

ALTER TABLE client_subscriptions ADD COLUMN IF NOT EXISTS type VARCHAR(50) NOT NULL DEFAULT 'hours' CHECK (type IN ('lessons', 'hours'));
ALTER TABLE client_subscriptions DROP COLUMN IF EXISTS format;

ALTER TABLE clients DROP COLUMN IF EXISTS rate_individual;
ALTER TABLE clients DROP COLUMN IF EXISTS rate_pair;
ALTER TABLE clients DROP COLUMN IF EXISTS rate_group;
