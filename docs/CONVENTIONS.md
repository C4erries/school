# Принципы и конвенции проекта

> Этот документ — **свод правил** для всех, кто пишет код (включая AI-агентов).
> Если ИИ-агент не знает как поступить — он смотрит сюда.

---

## 🏗️ Общие принципы

### 12-Factor (разумно)
- **Конфигурация** через переменные окружения. Файл `.env` для локальной разработки, **никогда не коммитим** `.env` с секретами.
- **Логи** в stdout/stderr в JSON формате.
- **Stateless** сервисы — состояние в БД/Redis/MinIO.
- **Один кодбейз** — один репозиторий (monorepo).
- **Зависимости** явно объявлены (go.mod, package.json).
- **Dev/prod parity** — Docker Compose для всего.

### Монорепо
```
school/
├── backend/          # Go приложение
├── frontend/         # React приложение
├── deploy/           # Docker, nginx, compose
├── docs/             # Документация
├── scripts/          # Утилиты, хелперы
└── ...
```

### Git
- Ветка `main` — стабильная.
- Feature ветки: `feature/auth`, `feature/courses`, etc.
- Коммиты: conventional commits (`feat:`, `fix:`, `docs:`, `refactor:`, `chore:`).
- Мёрж через squash (один коммит на фичу).

---

## 🐹 Go Backend

### Архитектура: Hexagonal / DDD

> Подробности: [ADR-002](decisions/002-go-architecture.md), [ADR-003](decisions/003-go-tooling.md)

Три слоя с чёткими границами:

| Слой | Путь | Импортирует | Ответственность |
|------|------|-------------|-----------------|
| Domain | `internal/domain/` | Ничего | Бизнес-сущности, value objects |
| Application | `internal/application/` | Domain | Use cases, бизнес-логика, интерфейсы зависимостей |
| Infrastructure | `internal/infrastructure/` | Domain + Application | Адаптеры, DI, конфиг, маппинг типов |

### Структура проекта

```
backend/
├── cmd/
│   ├── api/                           # main.go — HTTP API сервер
│   └── worker/                        # main.go — фоновые задачи (будущее)
│
├── internal/
│   ├── domain/                        # Бизнес-сущности (общий пакет)
│   │   ├── user.go
│   │   ├── course.go
│   │   └── scheduling/               # Изолированные поддомены — в подпакеты
│   │
│   ├── application/                   # Use cases, разбиты по сервисам
│   │   ├── auth/
│   │   │   └── service.go            # Определяет нужные интерфейсы тут же
│   │   ├── courses/
│   │   └── assignments/
│   │
│   └── infrastructure/                # Per-service инфраструктура
│       ├── api/                       # Инфраструктура для cmd/api
│       │   ├── config/               # Viper конфиг
│       │   ├── di/                   # Ручная сборка: NewContainer(cfg)
│       │   ├── adapters/
│       │   │   ├── httpserver/       # IN: oapi-generated + handlers
│       │   │   ├── postgres/         # OUT: реализация репозиториев (squirrel)
│       │   │   ├── s3/               # OUT: файлы (MinIO)
│       │   │   └── valkey/           # OUT: кэш/сессии (valkey-go)
│       │   └── metrics/
│       │
│       ├── contracts/                 # Общие инфраструктурные контракты
│       ├── runtime/                   # Runtime-утилиты
│       └── transporterrors/           # Общие ошибки транспорта
│
├── api/
│   └── openapi/
│       └── api.yaml                   # OpenAPI контракт
│
├── migrations/                        # SQL миграции (golang-migrate)
├── go.mod
└── go.sum
```

### Правила слоёв

**Domain**:
- Чистые структуры и методы, **нет импортов** из application/infrastructure
- Общий плоский пакет; изолированные поддомены — подпакеты

**Application**:
- Определяет **интерфейсы в пакете-потребителе** (idiomatic Go)
- Работает только с domain-типами
- Не знает про HTTP, SQL, S3

**Infrastructure**:
- Реализует интерфейсы из application
- Маппит external типы ↔ domain типы (DTOs живут здесь)
- Каждый сервис (`cmd/<svc>`) имеет свою папку `infrastructure/<svc>/`
- IN-адаптеры (HTTP handlers) вызывают application services
- OUT-адаптеры (postgres, s3) реализуют интерфейсы application services

### Конвенции Go
- **Версия Go**: 1.23+
- **Именование**: стандартное Go — camelCase для приватного, PascalCase для экспортируемого
- **Ошибки**: оборачиваем с контекстом (`fmt.Errorf("create user: %w", err)`)
- **Контекст**: `context.Context` — первый параметр во всех функциях с I/O
- **Интерфейсы**: определяем **по месту использования** (в пакете-потребителе)
- **Транзакции**: прокидываются через `context.Context` (паттерн `Transactor`), бизнес-логика изолирована от деталей СУБД
- **Тесты**: см. раздел «Тестирование» ниже
- **Линтер**: `golangci-lint` с конфигом в репозитории

### Тестирование
- **Фреймворк**: `testify` (assert, require, suite)
- **Табличные тесты** — предпочтительный формат, используем по возможности
- **Моки**: генерируем через `mockery v3+` — это основной способ
- Моки кладём в `<пакет>/mocks/`, конфиг в `.mockery.yaml`
- **Фейки / стабы**: допустимы когда генерируемый мок неудобен (сложное поведение, stateful-зависимости), но это исключение, а не правило
- **Именование**: `TestXxx_Method_Scenario` или табличный `TestXxx_Method`

### Инструментарий

| Инструмент | Назначение |
|------------|------------|
| `net/http` (stdlib) | HTTP router (может быть заменён позже) |
| `squirrel` | SQL query builder (не ORM!) |
| `golang-migrate v4` | Миграции БД (через Docker Compose) |
| `oapi-codegen` | Генерация server interface + types из OpenAPI |
| `mockery v3+` | Генерация моков (в `<pkg>/mocks/`) |
| `viper` | Конфигурация (env + yaml) |
| `slog` (stdlib) | Логирование |
| `valkey-go` | Высокопроизводительный клиент для Valkey (кэш/сессии) |
| `testify` | Тестирование (assert, mock) |
| `golangci-lint` | Линтер |

### Работа с БД
- **PostgreSQL** — основная БД
- **squirrel** для запросов, raw SQL допустим для статических запросов
- **golang-migrate** — миграции, запуск через Docker Compose
- **Никакого ORM** (GORM, ent, etc.)
- **Транзакции (Transactor через контекст)**:
  - Интерфейс объявляется на стороне потребителя (в `application`):
    ```go
    type Transactor interface {
        WithinTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
    }
    ```
  - Реализация в `infrastructure` начинает транзакцию и помещает её в `context.Context` (`txCtx`).
  - Репозитории проверяют наличие транзакции в контексте: если есть — выполняют запрос в транзакции, если нет — через обычный пул соединений.
  - Слой Use Case **никогда не импортирует** низкоуровневые типы транзакций (`*sql.Tx`, `pgx.Tx`).

### API контракт
- OpenAPI спеки: `api/openapi/<service>.yaml`
- Генерация: `oapi-codegen` → `infrastructure/<svc>/adapters/httpserver/generated/`
- Handlers имплементят сгенерированный `ServerInterface`
- Версионирование: `/api/v1/...`
- Формат ошибок: единый JSON
- Пагинация: cursor-based (для лент), offset-based (для админки)
- Аутентификация: Bearer JWT

### Обработка ошибок (API)
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Email is required",
    "details": [...]
  }
}
```

---

## ⚛️ React Frontend

### Стек
- **React 18+** с TypeScript (strict mode)
- **Vite** для сборки
- **Routing**: React Router v6+
- **State management**: TBD (Zustand / Redux Toolkit — определим когда понадобится)
- **HTTP клиент**: axios / ky с типизированным API layer
- **UI**: TBD (может Shadcn/UI, может MUI — определим на Фазе 1)

### Структура
```
frontend/
├── src/
│   ├── app/              # App shell, providers, routing
│   ├── pages/            # Page components
│   ├── features/         # Feature modules
│   ├── shared/           # Shared UI components, utils, hooks
│   ├── api/              # API layer (typed endpoints)
│   └── types/            # Shared TypeScript types
├── public/
├── index.html
├── vite.config.ts
├── tsconfig.json
└── package.json
```

### Конвенции
- **Компоненты**: functional components + hooks
- **Именование**: PascalCase для компонентов, camelCase для утилит
- **Стили**: TBD (CSS Modules / Tailwind / styled-components)
- **Типы**: строгая типизация, минимум `any`

---

## 🐳 Docker / Infrastructure

### Docker Compose сервисы
| Сервис | Образ | Порт |
|--------|-------|------|
| `postgres` | postgres:16 | 5432 |
| `valkey` | valkey/valkey:8 | 6379 |
| `minio` | minio/minio | 9000/9001 |
| `backend` | Собираем сами | 8080 |
| `frontend` | Node (dev) / nginx (prod) | 3000 |
| `nginx` | nginx:alpine | 80/443 |

### Nginx
- `/api/*` → backend
- `/*` → frontend
- Статика с кэшированием
- В будущем: SSL (Let's Encrypt / mkcert для dev)

### Env файлы
- `.env.example` — пример, коммитим
- `.env` — локальные значения, в `.gitignore`
- Секреты (JWT secret, DB password) — через env, **не хардкодим**

---

## 📋 Работа с задачами

### Где что искать
| Документ | Для чего |
|----------|----------|
| `docs/PROJECT_VISION.md` | Что строим и зачем |
| `docs/PRODUCT_SPEC.md` | Бизнес-логика, сценарии ролей, правила расписания и денег |
| `docs/ROADMAP.md` | Глобальный план по фазам |
| `docs/CURRENT_SPRINT.md` | Что делаем прямо сейчас |
| `docs/CONVENTIONS.md` | Как пишем код (этот файл) |
| `docs/decisions/` | ADR — Architecture Decision Records |

### Инструкция для AI-агентов
1. **Перед началом работы**: прочитай `CURRENT_SPRINT.md` чтобы понять контекст
2. **Сверяйся с конвенциями**: этот файл — закон
3. **Не ломай существующее**: прогони тесты перед коммитом
4. **Документируй решения**: если принимаешь архитектурное решение — создай ADR
5. **Обновляй статусы**: после завершения задачи обнови `CURRENT_SPRINT.md`
6. **Спрашивай при неясности**: лучше спросить, чем нагородить

---

*Последнее обновление: 2026-10-04*
