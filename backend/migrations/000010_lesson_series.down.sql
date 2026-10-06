-- 000010_lesson_series.down.sql
-- Откат таблицы регулярных серий занятий и колонок уроков

DROP INDEX IF EXISTS idx_lessons_series_original;
DROP INDEX IF EXISTS idx_lessons_series_id;

ALTER TABLE lessons
    DROP COLUMN IF EXISTS original_start_time,
    DROP COLUMN IF EXISTS series_id;

DROP TABLE IF EXISTS lesson_series;

