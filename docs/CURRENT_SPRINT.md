# Текущий спринт / ближайшие задачи

> Этот файл — **единственная точка правды** о том, что делаем прямо сейчас.
> Обновляется перед началом каждого спринта и после завершения задач.
> Завершенные спринты архивируются в каталоге `docs/sprints/`.

## Текущая фаза: 2.2.4 — Архитектурный рефакторинг бэкенда, переход на веб-фреймворк Echo v4 и декомпозиция адаптеров 🔵 (В РАБОТЕ)

---

## 🧠 Аналитика задач спринта (Роль: Tech Lead & Software Architect)

### 1. Переход на веб-фреймворк Echo v4 ([ADR-011](decisions/0011-backend-refactoring-and-echo-migration.md))
* **Проблема**:
  Текущий HTTP-транспорт построен на стандартном `net/http.ServeMux`. Это приводит к избыточному бойлерплейту:
  - Ручная обвязка контекстов запроса и извлечение параметров из путей URL;
  - Ручной парсинг Bearer-токенов в хэндлерах и middleware;
  - Ручная сериализация ошибок `writeError(w, status, code, msg)`;
  - Сложности с добавлением стандартных кросс-функциональных middleware (CORS, Recover, RequestID).
* **Решение**:
  - Подключение `github.com/labstack/echo/v4`.
  - Переключение генератора `oapi-codegen` на генерацию Echo-сервера (`echo-server: true` вместо `std-http-server: true`).
  - Все хэндлеры реализуют идиоматичный интерфейс Echo: `func(c echo.Context) error`.
  - Использование встроенных и производительных Echo middleware: Recover, CORS, централизованный `HTTPErrorHandler`.

---

### 2. Декомпозиция плоского пакета `adapters/postgres` по доменам
* **Проблема**:
  Каталог `backend/internal/infrastructure/api/adapters/postgres` содержит 12 файлов в одном общем пространстве имен (`client_repo.go`, `lesson_repository.go`, `user_repository.go`, `tag_repo.go`, `payment_repository.go` и др.). Отсутствуют доменные границы, что затрудняет навигацию и повышает связность кода.
* **Решение**:
  - Структурирование `adapters/postgres/` по изолированным предметным пакетам:
    - `adapters/postgres/` (базовые утилиты пула соединений `connection.go` и контекстных транзакций `transactor.go`).
    - `adapters/postgres/auth/` (`user_repository.go`).
    - `adapters/postgres/crm/` (`client_repo.go`, `subscription_repo.go`, `balance_adjustment_repo.go`, `tag_repo.go`).
    - `adapters/postgres/schedule/` (`lesson_repository.go`, `classroom_repository.go`, `teacher_student_repository.go`).
    - `adapters/postgres/finance/` (`payment_repository.go`, `partner_payout_repository.go`).

---

### 3. Модульная декомпозиция Application-сервисов (< 300–400 строк)
* **Проблема**:
  Файлы бизнес-логики разрослись до критических объемов:
  - `application/finance/service.go` — **878 строк**;
  - `application/analytics/service.go` — **802 строки**;
  - `application/crm/service.go` — **476 строк**;
  - `application/schedule/service.go` — **455 строк**.
  Это нарушает стандарт конвенций проекта (< 300–400 строк) и затрудняет чтение и модификацию кода AI-агентами.
* **Решение**:
  Разбиение логики внутри каждого пакета по нескольким файлам без изменения контракта структуры сервиса:
  - `application/finance/`:
    - `service.go`: интерфейсы зависимостей, конструктор `NewService`, метод `GetFinanceSummary` (< 200 строк);
    - `payments.go`: проведение платежей `RecordPayment`, история оплат, расчет задолженностей (< 250 строк);
    - `settlements.go`: расчет партнерских обязательств `GetPartnerSettlements`, фиксация выплат `RecordPartnerPayout` (< 250 строк);
    - `export.go`: генерация CSV отчетов с UTF-8 BOM (`ExportClients`, `ExportSchedule`, `ExportPayments`) (< 200 строк).
  - `application/analytics/`:
    - `service.go`: интерфейсы, конструктор, сводные KPI `GetOverview` (< 200 строк);
    - `forecast.go`: расчет прогнозируемой нагрузки и доходов `GetForecast` (< 250 строк);
    - `tags.go`: агрегация учеников и маржинальности по тегам `GetTagStats` (< 200 строк);
    - `dynamics.go`: временные ряды `GetDynamics`, доли форматов `GetFormatStats`, рейтинг учеников `GetClientStats` (< 250 строк).
  - `application/crm/`:
    - `service.go`: CRUD операции над клиентами (< 250 строк);
    - `subscriptions.go`: покупка, балансы и списание абонементов (< 200 строк);
    - `balance.go`: ручные корректировки баланса с аудитом (< 150 строк).
  - `application/schedule/`:
    - `service.go`: жизненный цикл уроков (создание, проведение, отмена) (< 250 строк);
    - `classrooms.go`: управление кабинетами и валидация нахлёстов/коллизий (< 250 строк).

---

### 4. Создание нового HTTP-слоя на базе Echo v4 (`adapters/http/`)
* **Проблема**:
  Каталог `adapters/httpserver` содержал монолитную структуру `APIHandler` со свалкой хэндлеров всех подсистем (`client_handler.go` 573 строки, `schedule_handler.go` 392 строки, `finance_handler.go` 321 строка).
* **Решение**:
  - Создание нового чистого пакета `adapters/http/`:
    - `adapters/http/middleware/`: Echo-совместимые middleware авторизации `AuthMiddleware` (извлечение Bearer, валидация JWT, сохранение в `echo.Context`), RBAC `RequireRoles`, структурированный `slog` логгер.
    - `adapters/http/response/`: формат стандартных ошибок и маппинг.
    - Предметные хэндлеры (`echo.Context`):
      - `adapters/http/auth/`: регистрация, логин, refresh, получение текущего профиля `/auth/me`, дефолтные ставки.
      - `adapters/http/crm/`: `client_handler.go`, `subscription_handler.go`, `tag_handler.go` (каждый < 250 строк).
      - `adapters/http/schedule/`: `lesson_handler.go`, `classroom_handler.go`, `calendar_handler.go`.
      - `adapters/http/finance/`: `finance_handler.go`, `export_handler.go`.
      - `adapters/http/analytics/`: `analytics_handler.go`.
      - `adapters/http/dashboard/`: `dashboard_handler.go`.
    - `adapters/http/server.go`: композитный фасад `Server`, реализующий сгенерированный `generated.ServerInterface`, настройка Echo роутера, регистрация маршрутов `/health` и `/api/v1/*`, graceful shutdown.

---

### 5. Обновление DI контейнера и верификация
* Обновление `backend/internal/infrastructure/api/di/container.go` под новые пакеты репозиториев и Echo сервер.
* Обновление `backend/cmd/api/main.go` для запуска и graceful shutdown Echo сервера.
* Адаптация unit-тестов хэндлеров под вызовы с `echo.Context`.
* Полное удаление устаревшего каталога `backend/internal/infrastructure/api/adapters/httpserver/`.
* Сквозная проверка: `make test` (Go юнит-тесты), `docker compose up -d --build` и `make test-e2e` (все 56 E2E тестов в Docker).

---

## 📋 Таблица задач спринта 2.2.4

| # | Задача | Статус | Приоритет | Ответственный / Субагент | Заметки |
|---|--------|--------|-----------|---------------------------|---------|
| 1 | Добавление Echo v4 и переключение генератора `oapi-codegen` на `echo-server` | 🔲 To Do | Критический | `backend_developer` | Добавить `github.com/labstack/echo/v4`, обновить `oapi-codegen.yaml`, сгенерировать Echo `ServerInterface`. |
| 2 | Декомпозиция `adapters/postgres` по предметным доменам | 🔲 To Do | Высокий | `backend_developer` | Создать пакеты `auth`, `crm`, `schedule`, `finance` внутри `adapters/postgres/`, декомпозировать `client_repo.go` < 400 строк. |
| 3 | Модульная декомпозиция Application-сервисов (< 300–400 строк) | 🔲 To Do | Высокий | `backend_developer` | Разбить монолитные `finance/service.go`, `analytics/service.go`, `crm/service.go`, `schedule/service.go` на логические модули. |
| 4 | Реализация HTTP-слоя на базе Echo v4 (`adapters/http/`) | 🔲 To Do | Высокий | `backend_developer` | Реализовать Echo middleware, response, доменные хэндлеры (`auth`, `crm`, `schedule`, `finance`, `analytics`, `dashboard`) и root Echo Server. |
| 5 | Обновление DI (`container.go`), `main.go`, перевод unit-тестов и удаление legacy `httpserver` | 🔲 To Do | Высокий | `backend_developer` | Подключение новых компонентов в DI, перевод юнит-тестов на Echo Context, удаление старого каталога `httpserver/`. |
| 6 | Сквозная верификация: Unit-тесты (`make test`) и E2E тесты в Docker (`make test-e2e`) | 🔲 To Do | Критический | `qa_engineer` | Прогон `make test` (100% pass без race conditions), сборка Docker и прогон всех 56/56 E2E тестов (`make test-e2e`). |

---

## 📋 Детальное Микро-ТЗ спринта 2.2.4

### Задача 1: Зависимости и OpenAPI кодогенерация
* **Действия**:
  1. Выполнить `go get github.com/labstack/echo/v4` в каталоге `backend/`.
  2. Обновить `backend/api/openapi/oapi-codegen.yaml`:
     ```yaml
     package: generated
     generate:
       echo-server: true
       models: true
     output: internal/infrastructure/api/adapters/http/generated/api.gen.go
     ```
  3. Запустить `make oapi`.
  4. Проверить создание файла `backend/internal/infrastructure/api/adapters/http/generated/api.gen.go` и интерфейса `ServerInterface` с методами вида `(ctx echo.Context) error`.

---

### Задача 2: Декомпозиция `adapters/postgres`
* **Действия**:
  1. Оставить в `internal/infrastructure/api/adapters/postgres/`:
     - `connection.go` (подключение к БД)
     - `transactor.go` (транзакционный менеджер)
  2. Создать подпакеты:
     - `adapters/postgres/auth/`: `user_repository.go`
     - `adapters/postgres/crm/`: `client_repo.go`, `client_repo_queries.go` (вынос фильтров и сборки SQL для сохранения лимита < 300 строк), `subscription_repo.go`, `balance_adjustment_repo.go`, `tag_repo.go`
     - `adapters/postgres/schedule/`: `lesson_repository.go`, `classroom_repository.go`, `teacher_student_repository.go`
     - `adapters/postgres/finance/`: `payment_repository.go`, `partner_payout_repository.go`
  3. Проверить, что все экспортируемые конструкторы (`NewUserRepository`, `NewClientRepository` и т.д.) доступны в своих пакетах.

---

### Задача 3: Модульная декомпозиция Application-сервисов
* **Действия**:
  1. `backend/internal/application/finance/`:
     - `service.go`: структуры, конструктор `NewService`, метод `GetFinanceSummary`
     - `payments.go`: `RecordPayment`, `ListPayments`, расчет задолженностей
     - `settlements.go`: `GetPartnerSettlements`, `RecordPartnerPayout`, `ListPartnerPayouts`
     - `export.go`: `ExportClientsCSV`, `ExportScheduleCSV`, `ExportPaymentsCSV`
  2. `backend/internal/application/analytics/`:
     - `service.go`: структуры, конструктор `NewService`, `GetOverview`
     - `forecast.go`: `GetForecast`
     - `tags.go`: `GetTagStats`
     - `dynamics.go`: `GetDynamics`, `GetFormatStats`, `GetClientStats`
  3. `backend/internal/application/crm/`:
     - `service.go`: CRUD клиентов, поиск, архивация
     - `subscriptions.go`: `AddSubscription`, списание
     - `balance.go`: `AdjustBalance`
  4. `backend/internal/application/schedule/`:
     - `service.go`: создание, изменение, проведение, отмена уроков
     - `classrooms.go`: управление кабинетами, проверка коллизий
  5. Убедиться, что каждый файл строго < 300–400 строк кода.

---

### Задача 4: Реализация нового HTTP-транспорта на Echo v4 (`adapters/http/`)
* **Действия**:
  1. `adapters/http/middleware/`:
     - `auth.go`: Echo middleware `AuthMiddleware(validator TokenValidator) echo.MiddlewareFunc`. Извлекает Bearer token, валидирует, помещает `UserClaims` в `c.Set("user_claims", claims)`. Хелпер `UserFromContext(c echo.Context) (*security.UserClaims, bool)`.
     - `roles.go`: Echo middleware `RequireRoles(roles ...domain.Role) echo.MiddlewareFunc`.
     - `logger.go`: интеграция `slog` через `middleware.RequestLoggerWithConfig`.
  2. `adapters/http/response/`:
     - Хелпер `Error(c echo.Context, status int, code, message string) error` возвращающий `generated.ErrorResponse`.
  3. Доменные хэндлеры:
     - `adapters/http/auth/`: регистрация, логин, refresh, me, rates.
     - `adapters/http/crm/`: `client_handler.go`, `subscription_handler.go`, `tag_handler.go`.
     - `adapters/http/schedule/`: `lesson_handler.go`, `classroom_handler.go`, `calendar_handler.go`.
     - `adapters/http/finance/`: `finance_handler.go`, `export_handler.go`.
     - `adapters/http/analytics/`: `analytics_handler.go`.
     - `adapters/http/dashboard/`: `dashboard_handler.go`.
  4. `adapters/http/server.go`:
     - Root facade `Server` объединяет хэндлеры и реализует `generated.ServerInterface`.
     - Регистрация эндпоинтов через `generated.RegisterHandlers(e, server)` и `generated.RegisterHandlersWithBaseURL(e, server, "/api/v1")`.
     - Graceful stop через `e.Shutdown(ctx)`.

---

### Задача 5 & 6: DI сборка, тесты и верификация полного цикла
* **Действия**:
  1. Обновить `di/container.go`: внедрение Echo Server, новых пакетов адаптеров postgres и http.
  2. Обновить `cmd/api/main.go` под Echo Server.
  3. Перевести unit-тесты хэндлеров на Echo context:
     - `req := httptest.NewRequest(...)`, `rec := httptest.NewRecorder()`, `c := echo.New().NewContext(req, rec)`.
  4. Удалить устаревший каталог `adapters/httpserver/`.
  5. Прогон `make test` (все тесты Go зеленые, без гонок).
  6. Прогон `docker compose up -d --build` и `make test-e2e` (все 56/56 тестов зеленые).

---

## 🎯 Definition of Done (DoD) Спринта 2.2.4
- [ ] Веб-фреймворк Echo v4 подключен и настроен в качестве основного HTTP-транспорта.
- [ ] OpenAPI кодогенерация переведена на `echo-server`, генерируется `ServerInterface` для Echo.
- [ ] Пакет `adapters/postgres` декомпозирован на изолированные доменные подпакеты (`auth`, `crm`, `schedule`, `finance`).
- [ ] Монолитные файлы `finance/service.go`, `analytics/service.go`, `crm/service.go`, `schedule/service.go` и `client_repo.go` декомпозированы по стандарту < 300–400 строк.
- [ ] Создан новый модульный пакет `adapters/http/` с Echo middleware, централизованной обработкой ошибок и доменными хэндлерами.
- [ ] Каталог устаревшего `adapters/httpserver` полностью удален.
- [ ] DI контейнер и `cmd/api/main.go` переведены на Echo с сохранением graceful shutdown.
- [ ] Unit-тесты бэкенда успешно адаптированы под Echo и проходят без ошибок (`make test`).
- [ ] Контракт API полностью сохранен: 56/56 E2E автотестов в Docker проходят успешно (`make test-e2e`).

---

## 🗄️ Оставшийся бэклог на следующие спринты

1. **Фаза 2.3: Регулярные занятия (Recurring Lessons & Series)**:
   - Поддержка стандартов RFC 5545 RRULE (повторения еженедельно, с интервалами, по дням недели).
   - Генерация виртуальных вхождений без раздувания базы данных.
   - Гранулярное редактирование: «Только этот урок», «Этот и последующие», «Вся серия».
   - Интеграция с расчетом прогноза и абонементами.
2. **Анализ оптимизации и ресурсоемкости фронтенда (GPU/CPU профилирование, Lightweight / Power Save Mode)**:
   - Оптимизация `<LiquidBackground />` (canvas FPS limit, pause on blur / tab hidden).
   - Тумблер Lite Mode для отключения тяжелого `backdrop-blur` и канваса для максимальной разгрузки видеокарты на слабых устройствах.
3. **Фаза 3: Двусторонняя интеграция с Google Calendar (OAuth 2.0)**:
   - Прямая запись событий в Google Calendar через Google API.
   - Двусторонняя блокировка слотов в расписании школы при занятости в личном календаре Google.
4. **Фаза 4: Платформа "Школа" (Multi-player)**:
   - Личные кабинеты учеников и владельца школы.

---

## Лог изменений

| Дата | Что изменилось |
|------|---------------|
| 2026-10-06 | **Сформирован Спринт 2.2.4 (Технический долг)**: Архитектурный рефакторинг бэкенда перед переходом к Фазе 2.3. Миграция на веб-фреймворк Echo v4, генерация Echo-сервера через `oapi-codegen`, декомпозиция плоских пакетов `adapters/postgres` и `adapters/http` по доменным контекстам (`auth`, `crm`, `schedule`, `finance`, `analytics`, `dashboard`), декомпозиция крупных сервисов на модули < 300–400 строк. Спринт 2.2.3 заархивирован в `docs/sprints/sprint-2.2.3.md`. Зафиксирован [ADR-011](decisions/0011-backend-refactoring-and-echo-migration.md). |
| 2026-10-06 | **Спринт 2.2.3 успешно завершен**: Умная аналитика (Факт/Прогноз), статистика по тегам, редизайн Дашборда (виджеты расписания на сегодня, финансов месяца, быстрых действий), разделение тегов на бизнес-партнеров (`school_percent > 0`) и информационные, индикатор времени в календаре. 56/56 E2E тестов в Docker пройдены успешно. |

---

*Последнее обновление: 2026-10-06*
