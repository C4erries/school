-- 000004_tutor_assistant_pivot.up.sql

-- 1. Create clients table
CREATE TABLE IF NOT EXISTS clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    base_rate NUMERIC(10, 2) NOT NULL DEFAULT 0,
    school_percent_tag INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_clients_teacher_id ON clients (teacher_id);

-- 2. Create client_subscriptions table
CREATE TABLE IF NOT EXISTS client_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL CHECK (type IN ('lessons', 'hours')),
    balance NUMERIC(10, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_client_subscriptions_client_id ON client_subscriptions (client_id);

-- 3. Modify lessons table
DELETE FROM lessons;

ALTER TABLE lessons DROP CONSTRAINT IF EXISTS lessons_student_id_fkey;
DROP INDEX IF EXISTS idx_lessons_student_id;
ALTER TABLE lessons DROP COLUMN IF EXISTS student_id;

ALTER TABLE lessons ADD COLUMN client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_lessons_client_id ON lessons (client_id);

ALTER TABLE lessons DROP CONSTRAINT IF EXISTS lessons_status_check;
ALTER TABLE lessons ADD CONSTRAINT lessons_status_check CHECK (
    status IN ('scheduled', 'completed', 'cancelled')
);
ALTER TABLE lessons ALTER COLUMN status SET DEFAULT 'scheduled';

-- 4. Drop teacher_students table
DROP TABLE IF EXISTS teacher_students;
