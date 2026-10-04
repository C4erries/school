-- 000003_create_schedule_and_classrooms.down.sql
-- Откат таблиц расписания, привязок и кабинетов

DROP TABLE IF EXISTS lessons CASCADE;
DROP TABLE IF EXISTS teacher_students CASCADE;
DROP TABLE IF EXISTS classrooms CASCADE;
