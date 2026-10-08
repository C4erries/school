-- 000012_client_notes.up.sql
-- Создание таблицы свободных заметок репетитора по ученикам (Спринт 2.4.2, ADR-014)

CREATE TABLE IF NOT EXISTS client_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_client_notes_client_id ON client_notes (client_id);
CREATE INDEX IF NOT EXISTS idx_client_notes_teacher_id ON client_notes (teacher_id);
CREATE INDEX IF NOT EXISTS idx_client_notes_created_at ON client_notes (created_at);

