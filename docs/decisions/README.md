# Architecture Decision Records

Здесь хранятся записи о ключевых архитектурных решениях.

## Формат

Каждый ADR — отдельный файл: `NNN-название.md`

```markdown
# ADR-NNN: Название решения

**Дата**: YYYY-MM-DD
**Статус**: Proposed / Accepted / Deprecated / Superseded

## Контекст
Почему встал вопрос?

## Решение
Что решили?

## Альтернативы
Что ещё рассматривали?

## Последствия
Что это означает для проекта?
```

## Список ADR

| # | Название | Статус | Дата |
|---|----------|--------|------|
| 001 | [Технологический стек](001-tech-stack.md) | Accepted | 2026-10-04 |
| 002 | [Go архитектура (DDD layout)](002-go-architecture.md) | Accepted | 2026-10-04 |
| 003 | [Инструментарий Go](003-go-tooling.md) | Accepted | 2026-10-04 |
| 004 | [E2E Testing Framework (Python + pytest + Docker)](004-e2e-testing-framework.md) | Accepted | 2026-10-04 |
| 0005 | [Tutor Assistant Pivot (Отделение Client от User)](0005-tutor-assistant-pivot.md) | Accepted | 2026-10-05 |
| 0006 | [Единый AppLayout и правила верстки Apple Liquid Glass](0006-app-layout-and-glass-conventions.md) | Accepted | 2026-10-05 |
| 0007 | [Учет в часах, тарифная сетка ставок и форматные абонементы](0007-hourly-rates-and-format-subscriptions.md) | Accepted | 2026-10-05 |
| 0008 | [Декомпозиция монолитного фронтенда и Client-Side Caching (CSC) через Valkey](0008-frontend-decomposition-and-valkey-csc.md) | Accepted | 2026-10-05 |
| 0009 | [Журнал платежей (Payments Ledger), Взаиморасчеты с партнерами и Экспорт данных](0009-financial-ledger-and-partner-settlements.md) | Accepted | 2026-10-06 |
| 0010 | [Регламент оркестрации субагентов и передачи контекста](0010-subagent-orchestration-and-prompting.md) | Accepted | 2026-10-06 |
| 0011 | [Архитектурный рефакторинг бэкенда, переход на Echo v4 и декомпозиция адаптеров](0011-backend-refactoring-and-echo-migration.md) | Accepted | 2026-10-06 |
| 0012 | [Регулярные занятия (Recurring Lessons), стандарт RFC 5545 RRULE и совместимость с Google Calendar](0012-recurring-lessons-rrule-and-calendar-sync.md) | Accepted | 2026-10-06 |
| 0013 | [Дневник занятий (Lesson Journal) и Управление домашними заданиями (Homework Management)](0013-lesson-journal-and-homework-management.md) | Accepted | 2026-10-08 |


