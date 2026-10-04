# ADR-003: Инструментарий Go

**Дата**: 2026-10-04
**Статус**: Accepted

## Контекст

Выбор инструментов и библиотек для Go backend.

## Решения

| Категория | Выбор | Обоснование |
|-----------|-------|-------------|
| HTTP router | **net/http** (stdlib) | Минимум зависимостей, Go 1.22+ routing patterns; можно заменить позже |
| SQL | **squirrel** | Query builder без ORM, удобен для динамических запросов |
| Миграции | **golang-migrate v4** | Знакомый инструмент, запуск через Docker Compose service |
| API контракт | **oapi-codegen** | Генерация server interface + types из OpenAPI, target: net/http |
| Моки | **mockery v3+** | Генерация моков для testify; моки в `mocks/` рядом с интерфейсом |
| Конфигурация | **Viper** | Чтение env vars, yaml defaults |
| Логирование | **slog** (stdlib) | Достаточно для начала, структурированные логи |
| Линтер | **golangci-lint** | Стандарт индустрии |
| Тестирование | **testify** | assert + mock + suite |
| Кэш / In-Memory | **valkey-go** | Высокопроизводительный официальный клиент для Valkey |
| Транзакции | **Transactor-in-context** | Проброс БД-транзакции через context.Context без утечки SQL в use cases |

## Детали

### Транзакции (Transactor-in-context)
- В `internal/application/` объявляется интерфейс транзактора:
  ```go
  type Transactor interface {
      WithinTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
  }
  ```
- Реализация в `infrastructure/.../postgres` начинает транзакцию, оборачивает её в `context.WithValue` и передаёт `txCtx` в колбэк `fn`.
- Репозитории в слое инфраструктуры проверяют наличие транзакции в контексте:
  - Если есть — выполняют запрос через `tx`.
  - Если нет — выполняют запрос через обычный пул соединений (`db`).
- **Слой Use Case никогда не импортирует `*sql.Tx` / `pgx.Tx`** — бизнес-логика чиста от деталей СУБД.

### Valkey (`valkey-go`)
- Замена Redis на Valkey (`valkey/valkey:8` в Docker).
- В Go коде используем официальную легковесную библиотеку `github.com/valkey-io/valkey-go`.
- Адаптер располагается в `internal/infrastructure/<service>/adapters/valkey/`.

### oapi-codegen
- Спеки: `api/openapi/<service>.yaml`
- Генерируемый код: `internal/infrastructure/<service>/adapters/httpserver/generated/`
- Наши handlers имплементят сгенерированный ServerInterface
- Target: `std` (net/http)

### mockery
- Конфиг: `.mockery.yaml` в корне
- Моки создаются в `<package>/mocks/` рядом с интерфейсом
- `pkgname: mocks`, `structname: InterfaceName`, `filename: snake_case.go`
- Интерфейсы перечисляются явно в конфиге

### golang-migrate
- Файлы миграций: `migrations/`
- Запуск: отдельный сервис в Docker Compose
- Формат: `NNNNNN_description.up.sql` / `NNNNNN_description.down.sql`

### squirrel
- Используется в OUT-адаптерах (`infrastructure/<service>/adapters/postgres/`)
- Для статических запросов допустим и raw SQL
- **Никакого ORM** (GORM, ent, etc.)

## Альтернативы

- **chi/echo** — не нужны на старте, stdlib достаточно
- **sqlc** — хорош для статических запросов, но squirrel гибче для динамических фильтров
- **GORM** — принципиально отвергнут: скрывает SQL, магия, плохо для понимания
- **zerolog** — мощнее slog, но stdlib достаточно
- **goose** — альтернатива golang-migrate, но нет опыта

