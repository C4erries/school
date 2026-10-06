-- 000010_lesson_series.up.sql
-- Создание таблицы регулярных серий занятий и расширение таблицы уроков

CREATE TABLE IF NOT EXISTS lesson_series (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    classroom_id UUID REFERENCES classrooms(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL DEFAULT 'Занятие',
    rrule VARCHAR(255) NOT NULL,
    start_time_of_day TIME NOT NULL,
    duration_minutes INT NOT NULL,
    format VARCHAR(50) NOT NULL,
    location_or_url TEXT,
    notes TEXT,
    start_date DATE NOT NULL,
    until_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lesson_series_teacher_id ON lesson_series (teacher_id);
CREATE INDEX IF NOT EXISTS idx_lesson_series_client_id ON lesson_series (client_id);

ALTER TABLE lessons
    ADD COLUMN IF NOT EXISTS series_id UUID REFERENCES lesson_series(id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS original_start_time TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_lessons_series_id ON lessons (series_id);
CREATE INDEX IF NOT EXISTS idx_lessons_series_original ON lessons (series_id, original_start_time);

