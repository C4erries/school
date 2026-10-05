-- 000004_tutor_assistant_pivot.down.sql

-- Recreate teacher_students
CREATE TABLE IF NOT EXISTS teacher_students (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_teacher_student UNIQUE (teacher_id, student_id)
);

CREATE INDEX IF NOT EXISTS idx_teacher_students_teacher_id ON teacher_students (teacher_id);
CREATE INDEX IF NOT EXISTS idx_teacher_students_student_id ON teacher_students (student_id);

-- Revert lessons
DELETE FROM lessons;

ALTER TABLE lessons DROP CONSTRAINT IF EXISTS lessons_client_id_fkey;
DROP INDEX IF EXISTS idx_lessons_client_id;
ALTER TABLE lessons DROP COLUMN IF EXISTS client_id;

ALTER TABLE lessons ADD COLUMN student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_lessons_student_id ON lessons (student_id);

ALTER TABLE lessons DROP CONSTRAINT IF EXISTS lessons_status_check;
ALTER TABLE lessons ADD CONSTRAINT lessons_status_check CHECK (
    status IN ('pending_confirmation', 'confirmed', 'completed', 'cancelled_by_teacher', 'cancelled_by_student', 'declined', 'no_show')
);
ALTER TABLE lessons ALTER COLUMN status SET DEFAULT 'pending_confirmation';

-- Drop client_subscriptions
DROP TABLE IF EXISTS client_subscriptions;

-- Drop clients
DROP TABLE IF EXISTS clients;
