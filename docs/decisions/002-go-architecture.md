# ADR-002: Go архитектура (DDD layout)

**Дата**: 2026-10-04
**Статус**: Accepted

## Контекст

Нужно определить структуру Go-кода: как организовать слои DDD, где живут сервисы,
как разделять несколько сервисов в monorepo.

## Решение

### Архитектурный стиль

Гексагональная архитектура (Ports & Adapters) с DDD-слоями.

### Три слоя

| Слой | Путь | Знает про | Ответственность |
|------|------|-----------|-----------------|
| **Domain** | `internal/domain/` | Ничего | Бизнес-сущности, value objects |
| **Application** | `internal/application/` | Domain | Use cases, бизнес-логика, определяет нужные интерфейсы |
| **Infrastructure** | `internal/infrastructure/` | Domain + Application | Адаптеры (HTTP, DB, S3), DI, конфиг, маппинг типов |

### Направление зависимостей

```
domain  ←──  application  ←──  infrastructure  ←──  cmd/
(чистый)     (use cases)       (адаптеры)           (main)
```

- `domain` — не импортирует ничего из проекта
- `application` — импортирует только `domain`
- `infrastructure` — импортирует `domain` + `application`
- `cmd/` — точка входа, импортирует `infrastructure` (DI container)

### Структура `internal/domain/`

Общий пакет для бизнес-сущностей. По умолчанию flat-структура.
Изолированные поддомены — в подпакетах.

```
domain/
├── user.go              # User, Role, ...
├── course.go            # Course, Module, Lesson
├── assignment.go        # Assignment, Submission
├── errors.go            # Доменные ошибки
└── scheduling/          # Изолированный поддомен (если не нужен всем)
    └── event.go
```

### Структура `internal/application/`

Разбит по бизнес-сервисам (bounded contexts). Каждый сервис определяет
свои интерфейсы **по месту использования** (idiomatic Go).

```
application/
├── auth/
│   └── service.go       # type UserRepo interface {...}
│                        # type Service struct { repo UserRepo }
├── courses/
│   └── service.go
└── assignments/
    └── service.go
```

- Интерфейсы определяются **в пакете, который их потребляет**
- Application работает с domain-типами
- Маппинг external → domain делает infrastructure

### Структура `internal/infrastructure/`

Организована **per-service** (каждый сервис из `cmd/` имеет свою инфраструктуру).
Плюс общие пакеты на уровне infrastructure.

```
infrastructure/
├── api/                          # === Сервис "api" ===
│   ├── config/                  #   Конфиг этого сервиса (Viper)
│   ├── di/                      #   DI container (NewContainer)
│   ├── adapters/
│   │   ├── httpserver/          #   IN-адаптер: HTTP (oapi-generated + handlers)
│   │   ├── postgres/            #   OUT-адаптер: репозитории
│   │   ├── s3/                  #   OUT-адаптер: файлы (MinIO)
│   │   └── redis/               #   OUT-адаптер: кэш/сессии
│   └── metrics/                 #   Метрики
│
├── worker/                       # === Сервис "worker" (будущее) ===
│   ├── config/
│   ├── di/
│   └── adapters/
│       └── ...
│
├── contracts/                    # Общие инфраструктурные контракты
├── runtime/                      # Runtime-утилиты
└── transporterrors/              # Общие ошибки транспорта
```

### DI Container

Ручная сборка. Иерархическая структура:
- Каждый адаптер (postgres, redis, s3) может иметь свой "мини-контейнер"
- Сервисный контейнер (`infrastructure/api/di/`) собирает адаптеры + application services
- Берёт всё из конфигурации, включает что нужно с нужными настройками

```go
// infrastructure/api/di/container.go
type Container struct {
    AuthService    *auth.Service
    CoursesService *courses.Service
    HTTPServer     *httpserver.Server
    // ...
}

func NewContainer(cfg *config.Config) (*Container, error) {
    db := postgres.NewConnection(cfg.Postgres)
    userRepo := postgres.NewUserRepo(db)
    authService := auth.NewService(userRepo)
    // ...
}
```

### Несколько бинарников

Физическое разделение сервисов:

```
cmd/
├── api/
│   └── main.go          # HTTP API сервер
├── worker/
│   └── main.go          # Background worker (будущее)
└── ...
```

Каждый `cmd/<service>/main.go` создаёт свой `infrastructure/<service>/di.Container`.

## Альтернативы

- **Feature-based** (internal/user/, internal/course/) — отвергнуто, сложнее шарить domain
- **Классический DDD** (domain/application/infrastructure без per-service split) — не масштабируется при нескольких бинарниках
- **Wire/fx** для DI — отвергнуто в пользу простоты ручной сборки

## Последствия

- Чёткие границы между слоями, легко видеть зависимости
- При добавлении нового сервиса — создаём `cmd/<svc>/` + `infrastructure/<svc>/`
- Application services переиспользуются между сервисами
- Domain — единая точка правды для бизнес-сущностей
