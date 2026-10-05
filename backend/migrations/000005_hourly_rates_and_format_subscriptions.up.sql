-- 000005_hourly_rates_and_format_subscriptions.up.sql

-- 1. Clients: add rate grid
ALTER TABLE clients ADD COLUMN IF NOT EXISTS rate_individual NUMERIC(10, 2) NOT NULL DEFAULT 0;
ALTER TABLE clients ADD COLUMN IF NOT EXISTS rate_pair NUMERIC(10, 2);
ALTER TABLE clients ADD COLUMN IF NOT EXISTS rate_group NUMERIC(10, 2);
UPDATE clients SET rate_individual = base_rate WHERE rate_individual = 0 AND base_rate > 0;

-- 2. Client subscriptions: add format and drop type
ALTER TABLE client_subscriptions ADD COLUMN IF NOT EXISTS format VARCHAR(50) NOT NULL DEFAULT 'individual';
ALTER TABLE client_subscriptions DROP CONSTRAINT IF EXISTS client_subscriptions_format_check;
ALTER TABLE client_subscriptions ADD CONSTRAINT client_subscriptions_format_check CHECK (format IN ('individual', 'pair', 'group'));
ALTER TABLE client_subscriptions DROP COLUMN IF EXISTS type;

-- 3. Create tags table
CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    school_percent INT NOT NULL DEFAULT 0,
    color VARCHAR(30) DEFAULT 'indigo',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tags_teacher_id ON tags (teacher_id);

-- 4. Create client_tags table
CREATE TABLE IF NOT EXISTS client_tags (
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (client_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_client_tags_client_id ON client_tags (client_id);
CREATE INDEX IF NOT EXISTS idx_client_tags_tag_id ON client_tags (tag_id);

-- 5. Lessons: update format check to individual, pair, group
ALTER TABLE lessons DROP CONSTRAINT IF EXISTS lessons_format_check;
UPDATE lessons SET format = 'individual' WHERE format NOT IN ('individual', 'pair', 'group');
ALTER TABLE lessons ADD CONSTRAINT lessons_format_check CHECK (format IN ('individual', 'pair', 'group'));
