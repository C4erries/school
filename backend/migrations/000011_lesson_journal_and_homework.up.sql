-- 000011_lesson_journal_and_homework.up.sql
-- Создание таблиц дневника занятий и домашних заданий (Спринт 2.4.1, ADR-013)

CREATE TABLE IF NOT EXISTS lesson_journals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lesson_id UUID NOT NULL UNIQUE REFERENCES lessons(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topic VARCHAR(255) NOT NULL,
    notes TEXT,
    performance_score INT CHECK (performance_score BETWEEN 1 AND 5),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lesson_journals_lesson_id ON lesson_journals (lesson_id);
CREATE INDEX IF NOT EXISTS idx_lesson_journals_client_id ON lesson_journals (client_id);
CREATE INDEX IF NOT EXISTS idx_lesson_journals_teacher_id ON lesson_journals (teacher_id);

CREATE TABLE IF NOT EXISTS homework_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assigned_lesson_id UUID REFERENCES lessons(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    due_date DATE,
    status VARCHAR(50) NOT NULL DEFAULT 'assigned' CHECK (status IN ('assigned', 'completed', 'not_done')),
    review_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_homework_assignments_client_id ON homework_assignments (client_id);
CREATE INDEX IF NOT EXISTS idx_homework_assignments_teacher_id ON homework_assignments (teacher_id);
CREATE INDEX IF NOT EXISTS idx_homework_assignments_assigned_lesson_id ON homework_assignments (assigned_lesson_id);
CREATE INDEX IF NOT EXISTS idx_homework_assignments_status ON homework_assignments (status);
