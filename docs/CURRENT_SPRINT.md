# Текущий спринт / ближайшие задачи

> Этот файл — **единственная точка правды** о том, что делаем прямо сейчас.
> Обновляется после завершения каждой крупной задачи.

## Текущая фаза: 1 — Аутентификация и пользователи 🔵

### Цель спринта
Реализовать полноценную систему аутентификации и авторизации: регистрация, вход, JWT токены, сессии в Valkey, ролевая модель (student, teacher, assistant, owner), интеграция с OpenAPI контрактом, middleware защиты роутов и базовый Auth UI на фронтенде.

### Задачи спринта

| # | Задача | Статус | Приоритет | Заметки |
|---|--------|--------|-----------|---------|
| 1 | Миграция БД: таблица `users` | ✅ Done | Высокий | UUID, email, password_hash, role, phone, name |
| 2 | Доменная модель `User` и `Role` | ✅ Done | Высокий | `internal/domain/user.go`, валидация |
| 3 | Сервис хэширования паролей (`bcrypt`) | ✅ Done | Высокий | `infrastructure/security/hasher.go` |
| 4 | Сервис JWT токенов (access/refresh) | ✅ Done | Высокий | Подпись, верификация, claims |
| 5 | Хранилище refresh-сессий в Valkey | ✅ Done | Высокий | Ротация токенов, инвалидация сессий |
| 6 | PostgreSQL `UserRepository` (squirrel) | ✅ Done | Высокий | Поддержка `Transactor-in-context` |
| 7 | Application Use Cases (auth service) | ✅ Done | Высокий | Register, Login, Refresh, GetMe + unit-тесты |
| 8 | OpenAPI спецификация для `/auth/*` | ✅ Done | Высокий | Контракты в `api.yaml`, `make oapi` |
| 9 | HTTP handlers и роуты аутентификации | ✅ Done | Высокий | Реализация `ServerInterface` |
| 10 | Auth & RBAC Middleware | ✅ Done | Высокий | Bearer token validator, проверка ролей |
| 11 | Frontend: Auth Context & API клиент | ✅ Done | Средний | Токены в localStorage, user state, useAuth |
| 12 | Frontend: Страницы Login & Register | ✅ Done | Средний | Формы Apple Liquid Glass, валидация |
| 13 | Frontend: ProtectedRoute по ролям | ✅ Done | Средний | Защита роутов для студента/преподавателя |

### Definition of Done (DoD) Фазы 1
- [x] Пользователь может зарегистрироваться с ролью (по умолчанию `student` или выбор при регистрации).
- [x] Логин возвращает пару Access Token (JWT) и Refresh Token.
- [x] Refresh token валидируется через Valkey с поддержкой ротации.
- [x] Эндпоинт `GET /api/v1/auth/me` возвращает профиль авторизованного пользователя.
- [x] `AuthMiddleware` корректно валидирует токен и отклоняет неавторизованные запросы с кодом 401.
- [x] На все use cases написаны табличные юнит-тесты с моками через `testify`.
- [x] На фронтенде работают экраны входа и регистрации, сохраняется сессия после перезагрузки.

### Блокеры / Открытые вопросы
- Нет блокеров. Фаза 1 успешно выполнена и протестирована.

---

## Лог изменений

| Дата | Что изменилось |
|------|---------------|
| 2026-10-04 | **Фаза 1 полностью завершена (100%)**: реализована регистрация, логин, ротация refresh-токенов в Valkey, JWT access-токены, PostgreSQL репозиторий со squirrel и Transactor, полный Auth UI в эстетике Apple Liquid Glass с ролями `student` и `teacher`, сквозные тесты пройдены |
| 2026-10-04 | Открыта Фаза 1 (Аутентификация и пользователи): декомпозированы 13 задач, определен DoD, .agents/ добавлен в .gitignore |
| 2026-10-04 | Настроен oapi-codegen: создан oapi-codegen.yaml, сгенерирован `generated/api.gen.go`, реализован `handler.go`, добавлен таргет `make oapi` — Фаза 0 завершена на 100% |
| 2026-10-04 | Завершена Фаза 0: создан React+Vite+TS фронтенд с healthcheck polling (R1), настроен Nginx reverse proxy с WebSocket HMR (R2), объединен docker-compose.yml для всех 7 сервисов (R3), актуализирована документация |
| 2026-10-04 | Реализован базовый скелет Go backend: структура DDD, Viper config, net/http сервер с healthcheck, Makefile, Docker Compose, первая миграция, unit-тесты (все pass) |
| 2026-10-04 | Проведен бизнес-анализ: создан PRODUCT_SPEC.md, скорректирован Roadmap (фокус на расписание и финансы репетитора) |
| 2026-10-04 | Зафиксирована Go архитектура (ADR-002, ADR-003), mockery конфиг |
| 2026-10-04 | Создан план проекта, зафиксированы стек и roadmap |

---

*Последнее обновление: 2026-10-04*
