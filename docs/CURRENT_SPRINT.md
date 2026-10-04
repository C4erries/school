# Текущий спринт / ближайшие задачи

> Этот файл — **единственная точка правды** о том, что делаем прямо сейчас.
> Обновляется после завершения каждой крупной задачи.

## Текущая фаза: 0 — Фундамент

### Приоритет: Высокий (делаем сейчас)

| # | Задача | Статус | Заметки |
|---|--------|--------|---------|
| 1 | Зафиксировать Go DDD layout | ✅ Done | ADR-002, ADR-003 |
| 2 | Создать структуру Go-проекта (скелет) | ✅ Done | cmd/api, internal/ (domain, app, infra) |
| 3 | Docker Compose (Postgres, Valkey, MinIO, migrate) | ✅ Done | docker-compose.yml + healthchecks |
| 4 | Nginx конфиг | 🔲 | Следующее |
| 5 | HTTP сервер + healthcheck (net/http) | ✅ Done | /health, /api/v1/health + тесты |
| 6 | Конфигурация через env (Viper) | ✅ Done | config.Load() + тесты |
| 7 | Первая миграция БД (golang-migrate) | ✅ Done | 000001_init.up.sql / down.sql |
| 8 | Makefile / Taskfile | ✅ Done | build, test, run, compose, migrate, mock |
| 9 | React + Vite + TS проект | 🔲 | |
| 10 | golangci-lint конфиг | ✅ Done | backend/.golangci.yaml |
| 11 | oapi-codegen setup + первый контракт | 🔲 | Спека готова в api/openapi/api.yaml |
| 12 | mockery конфиг | ✅ Done | .mockery.yaml |
| 13 | README с инструкцией запуска | 🔲 | |

### Приоритет: Средний (следующее)
- Nginx reverse-proxy
- React + Vite + TS скелет
- oapi-codegen генерация кода

### Блокеры / Открытые вопросы
- Нет блокеров

---

## Лог изменений

| Дата | Что изменилось |
|------|---------------|
| 2026-10-04 | Реализован базовый скелет Go backend: структура DDD, Viper config, net/http сервер с healthcheck, Makefile, Docker Compose, первая миграция, unit-тесты (все pass) |
| 2026-10-04 | Проведен бизнес-анализ: создан PRODUCT_SPEC.md, скорректирован Roadmap (фокус на расписание и финансы репетитора) |
| 2026-10-04 | Зафиксирована Go архитектура (ADR-002, ADR-003), mockery конфиг |
| 2026-10-04 | Создан план проекта, зафиксированы стек и roadmap |

---

*Последнее обновление: 2026-10-04*
