-- 000003_create_schedule_and_classrooms.up.sql
-- Создание таблиц кабинетов, привязок преподаватель-ученик и уроков

-- 1. Таблица кабинетов (classrooms)
CREATE TABLE IF NOT EXISTS classrooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    capacity INT NOT NULL DEFAULT 1,
    color VARCHAR(50) NOT NULL DEFAULT '#3B82F6',
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Таблица привязок учеников к преподавателям (teacher_students)
CREATE TABLE IF NOT EXISTS teacher_students (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_teacher_student UNIQUE (teacher_id, student_id)
);

CREATE INDEX IF NOT EXISTS idx_teacher_students_teacher_id ON teacher_students (teacher_id);
CREATE INDEX IF NOT EXISTS idx_teacher_students_student_id ON teacher_students (student_id);

-- 3. Таблица уроков (lessons)
CREATE TABLE IF NOT EXISTS lessons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    classroom_id UUID REFERENCES classrooms(id) ON DELETE SET NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    format VARCHAR(50) NOT NULL CHECK (format IN ('online', 'offline')),
    location_or_url TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending_confirmation' CHECK (
        status IN ('pending_confirmation', 'confirmed', 'completed', 'cancelled_by_teacher', 'cancelled_by_student', 'declined', 'no_show')
    ),
    notes TEXT,
    cancel_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lessons_teacher_id ON lessons (teacher_id);
CREATE INDEX IF NOT EXISTS idx_lessons_student_id ON lessons (student_id);
CREATE INDEX IF NOT EXISTS idx_lessons_classroom_id ON lessons (classroom_id);
CREATE INDEX IF NOT EXISTS idx_lessons_start_end ON lessons (start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_lessons_status ON lessons (status);
