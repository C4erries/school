-- 000001_init.up.sql
-- Инициализирующая миграция

-- Включаем расширение для генерации UUID v4 (если понадобится в будущем)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица версионирования или базовая таблица проверки
CREATE TABLE IF NOT EXISTS schema_initialization_check (
    id SERIAL PRIMARY KEY,
    initialized_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
