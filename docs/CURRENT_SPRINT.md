# Текущий спринт / ближайшие задачи

> Этот файл — **единственная точка правды** о том, что делаем прямо сейчас.
> Обновляется перед началом каждого спринта и после завершения задач.
> Завершенные спринты архивируются в каталоге `docs/sprints/`.

## Текущая фаза: 2.4.2 — Полноэкранный Дневник-Мессенджер (Study Stream), Быстрые заметки ученика и Автосохранение черновиков 🟢 (ЗАВЕРШЕН)

---

## 🧠 Аналитика задач спринта (Роль: CEO, PM & Lead Architect)

### 1. Бизнес-обоснование (CEO View)
* **Проблема**: Модалка отчета `LessonJournalModal` решила задачу быстрого ввода отчета за 30 секунд после урока. Однако когда преподаватель хочет подготовиться к уроку, перечитать историю прогресса за прошлый месяц, продумать программу или зафиксировать неформальную мысль вне урока («позвонила мама», «забыл тетрадь») — модального окна недостаточно. Кроме того, ввод отчета со смартфона несет риск потери текста при входящем звонке или разрядке телефона.
* **Решение**: Полноэкранный рабочий центр «Дневник-Мессенджер» (`/teacher/journal`):
  - Двухколоночный интерфейс: список учеников, отсортированный по дате последнего проведенного занятия (самые актуальные сверху) + центральный хронологический поток (Study Stream).
  - Два типа сообщений в потоке: **Урочный отчет** (тема, оценка понимания 1–5, заметки, ДЗ) и **Свободная быстрая заметка** (быстрая фиксация мыслей репетитора вне урока).
  - Сквозная связка: кнопка «Дневник & ДЗ» в CRM переводит на эту страницу с уже выбранным учеником, а в модалке расписания появляется кнопка «Перейти в журнал» для вдумчивого продолжения.
  - Автосохранение черновиков (Zero Data Loss): сохранение полей отчета в `localStorage` с автоматическим восстановлением.
* **Бизнес-эффект**: Превращение платформы в полноценный "второй мозг" репетитора, повышение комфорта при глубокой работе и гарантия сохранности данных на смартфонах.

---

### 2. Архитектура решения ([ADR-014](decisions/0014-fullpage-journal-messenger-and-drafts.md))
* **Сущность свободных заметок `client_notes` (миграция 000012)**:
  - `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
  - `client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE`
  - `teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
  - `content TEXT NOT NULL`
  - `created_at TIMESTAMPTZ`, `updated_at TIMESTAMPTZ`
* **Сортировка учеников по активности**:
  - Добавление в API клиентов или вычисление поля `last_lesson_at` (дата и время самого свежего проведенного/запланированного урока) для ранжирования списка учеников в боковой панели.
* **Автосохранение черновика (Client-side Hook `useJournalDraft`)**:
  - Дебаунс 300ms для записи `topic`, `notes`, `performance_score`, `homework` в `localStorage` по ключу `journal_draft_${lessonId}`.
  - Мягкое восстановление при открытии формы и очистка при успешном `PUT /schedule/lessons/{id}/journal`.

---

## 📋 Таблица задач спринта 2.4.2

| # | Задача | Статус | Приоритет | Ответственный / Субагент | Заметки |
|---|--------|--------|-----------|---------------------------|---------|
| 1 | Миграция БД 000012 (`client_notes`) и Доменная модель `ClientNote` | ✅ Done | Критический | `backend_developer` | Таблица `client_notes`, доменная структура `ClientNote`, методы валидации и тесты. |
| 2 | OpenAPI контракт: эндпоинты свободных заметок и поле `last_lesson_at` | ✅ Done | Критический | `backend_developer` | Эндпоинты `/crm/clients/{id}/notes` (GET, POST, DELETE), кодогенерация `make oapi`. |
| 3 | Backend: Репозиторий `notes_repository.go`, use-cases и Echo v4 хэндлеры | ✅ Done | Высокий | `backend_developer` | Реализация CRUD заметок, сортировка учеников по `last_lesson_at`, юнит-тесты. |
| 4 | Frontend: Механизм автосохранения черновиков (`useJournalDraft`) и кнопка «Перейти в журнал» в `LessonJournalModal` | ✅ Done | Высокий | `frontend_developer` | LocalStorage автосохранение полей отчета, индикатор черновика, кнопка перехода в модалке расписания. |
| 5 | Frontend: Полноэкранная страница `/teacher/journal` (Liquid Glass Messenger) | ✅ Done | Высокий | `frontend_developer` | Двухколоночный layout (сайдбар учеников по датам, центральный поток уроков и заметок, поле быстрой отправки, переход из CRM). |
| 6 | E2E сценарии в Docker и регрессионная верификация | ✅ Done | Критический | `qa_engineer` | Автотесты заметок, навигации `/teacher/journal` и проверка всего сьюта в Docker (72/72 green). |

---

## 📋 Детальное Микро-ТЗ спринта 2.4.2

### Задача 1: Миграция БД 000012 (`backend/migrations/000012_client_notes.up.sql`)
* Создание таблицы `client_notes`:
  - `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
  - `client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE`
  - `teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
  - `content TEXT NOT NULL`
  - `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
  - `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
  - Индексы: `idx_client_notes_client_id`, `idx_client_notes_teacher_id`, `idx_client_notes_created_at`.
* Файл отката `000012_client_notes.down.sql`.
* Доменная модель `backend/internal/domain/client_note.go` и юнит-тесты `client_note_test.go`.

---

### Задача 2: Контракт OpenAPI (`backend/api/openapi/api.yaml`)
* **Схемы DTO**:
  - `ClientNoteResponse`: id, client_id, teacher_id, content, created_at, updated_at.
  - `CreateClientNoteRequest`: content (required string, непустой).
  - `UpcomingLessonInfo`: lesson_id, start_time, end_time, format, title, topic.
  - `StudyStreamItemResponse`: id, type (`lesson_report` | `note`), timestamp (date-time), lesson_report (optional LessonJournalBundleResponse), note (optional ClientNoteResponse).
  - `StudyStreamResponse`: client_id, upcoming_lesson (optional UpcomingLessonInfo), items (array of StudyStreamItemResponse, отсортированных хронологически как в чате).
  - `ClientResponse`: добавление опционального поля `last_lesson_at` (date-time string nullable).
* **Эндпоинты**:
  - `GET /crm/clients/{id}/stream` -> единый агрегированный поток обучения `StudyStreamResponse`.
  - `POST /crm/clients/{id}/notes` -> создание быстрой заметки (`CreateClientNoteRequest`).
  - `DELETE /crm/clients/notes/{id}` -> удаление быстрой заметки (204 No Content).
* Выполнить `make oapi`.

---

### Задача 3: Backend реализация свободных заметок, сортировки и Study Stream
* Репозиторий `backend/internal/infrastructure/api/adapters/postgres/crm/notes_repository.go`.
* Расширение выборки клиентов с подсчетом `last_lesson_at` через `MAX(lessons.start_time)`.
* Сервис `StudyStreamService` (или в `application/journal/`):
  - Агрегация проведенных уроков с отчетами из `lesson_journals` и свободных заметок из `client_notes`.
  - Поиск ближайшего запланированного урока для компактной плашки `upcoming_lesson`.
  - Сортировка ленты (старые сверху, свежие снизу).
* Echo v4 хэндлеры в `adapters/http/crm/notes_handler.go` и `stream_handler.go`.
* Регистрация в DI контейнере, юнит-тесты `make test`.


---

### Задача 4: Frontend — Автосохранение черновиков и связка с расписанием
* Хук `frontend/src/shared/hooks/useJournalDraft.ts`:
  - Сохранение `topic`, `notes`, `performance_score`, `homework` в `localStorage` с дебаунсом.
  - Очистка черновика при сохранении отчета на сервер.
* Доработка `LessonJournalModal.tsx`:
  - Подключение `useJournalDraft(lesson.id)`.
  - Кнопка в заголовке модалки: «Перейти в журнал» (`GlassButton`, иконка `ExternalLink`), осуществляющая переход на `/teacher/journal?clientId=${lesson.client_id}&lessonId=${lesson.id}`.

---

### Задача 5: Frontend — Полноэкранная страница `/teacher/journal`
* Добавление маршрута `/teacher/journal` в `frontend/src/app/App.tsx` и пункта «Журнал» в `AppNavbar.tsx`.
* Страница `frontend/src/pages/teacher/TeacherJournalPage.tsx` в стиле Apple Liquid Glass:
  - **Сайдбар учеников**: список учеников, отсортированный по `last_lesson_at`, живой поиск, бейдж несданных ДЗ, подсветка активного ученика.
  - **Центральный поток (Study Stream)**:
    - Шапка с именем ученика, телефоном и балансом.
    - Единая хронологическая лента сообщений:
      1. Блоки проведенных уроков (тема, оценка понимания 1–5 чипом, заметки, выданное ДЗ со статусом);
      2. Блоки свободных заметок преподавателя (текст, дата, кнопка удаления).
    - Нижняя панель ввода (Message Input Bar): текстовое поле для отправки быстрой заметки + кнопка «Заполнить отчет по уроку».
* В карточке ученика `ClientCard.tsx`:
  - Кнопка «Дневник & ДЗ» теперь осуществляет роутинг на `/teacher/journal?clientId=${client.id}`.

---

### Задача 6: Комплексное E2E тестирование и регрессия
* Дополнение тестов в `tests/api/phase2_schedule/test_journal_and_homework.py`:
  - Создание, получение и удаление свободных заметок ученика.
  - Проверка сортировки учеников по дате последнего занятия.
* Прогон `make test-e2e` в Docker (100% green).
* Чистая сборка фронтенда `npm run lint && npm run build`.

---

## 🎯 Definition of Done (DoD) Спринта 2.4.2
- [ ] Применена миграция `000012_client_notes.up.sql`.
- [ ] Контракт OpenAPI обновлен эндпоинтами свободных заметок и полем `last_lesson_at`, код сгенерирован (`make oapi`).
- [ ] Реализован бэкенд CRUD свободных заметок с валидацией и проверкой прав репетитора.
- [ ] Список клиентов поддерживает отображение/сортировку по дате последнего занятия.
- [x] Реализован хук автосохранения черновика `useJournalDraft` для предотвращения потери данных.
- [x] В `LessonJournalModal` добавлена кнопка перехода в полноэкранный журнал.
- [x] Реализована страница `/teacher/journal` в стиле Apple Liquid Glass (двухколоночный layout мессенджера).
- [x] Кнопка «Дневник & ДЗ» в `ClientCard.tsx` ведет на страницу журнала с открытым профилем ученика.
- [ ] Все созданные файлы кода строго меньше 300–400 строк.
- [ ] Все юнит-тесты (`make test`) и E2E тесты в Docker (`make test-e2e`) зеленые (100% PASS).
- [ ] Сборка фронтенда чистая (`npm run lint && npm run build`, 0 ошибок).

---

## 📌 Бэклог следующих спринтов / Технический долг

### 🛠️ Инженерный рефакторинг Go: Унификация тестирования (Mockery) и декомпозиция `adapters/http`
1. **Аудит и унификация юнит-тестов Go**:
   - Устранить все рукописные структуры-заглушки (`mockLessonRepo`, `mockService` и т.п.) прямо в тестовых файлах `*_test.go`.
   - Заполнить интерфейсы в `.mockery.yaml` и настроить генерацию моков через `mockery v3+` в папки `<пакет>/mocks/`.
   - Провести строгое разделение: `Mock` (генерируемый через Mockery с `testify/mock`) vs `Fake` (только сложные stateful in-memory хранилища с префиксом `Fake`).
2. **Декомпозиция HTTP-адаптеров (`adapters/http`)**:
   - Ликвидировать накопление плоских файлов в корне `backend/internal/infrastructure/api/adapters/http/`.
   - Вынести контроллеры и их тесты по изолированным предметным подпакетам (по аналогии с `adapters/postgres/`):
     * `adapters/http/journal/` (перенести `journal_homework_server.go` и `journal_handler_test.go`);
     * `adapters/http/crm/` (перенести `client_handler_test.go`, `tag_handler_test.go`);
     * `adapters/http/schedule/` (перенести `schedule_handler_test.go`);
     * `adapters/http/finance/` (перенести `finance_handler_test.go`);
     * `adapters/http/analytics/` (перенести `analytics_and_calendar_handler_test.go`);
     * `adapters/http/dashboard/` (перенести `dashboard_handler_test.go`).
   - В корне `adapters/http/` оставить только маршрутизатор/фасад `server.go` и общие инфраструктурные обвязки.

---

## Лог изменений

| Дата | Что изменилось |
|------|---------------|
| 2026-10-08 | **Сформирован Спринт 2.4.2 (Фаза 2.4)**: Полноэкранный Дневник-Мессенджер (Study Stream), Быстрые свободные заметки ученика (`client_notes`) и Автосохранение черновиков (`useJournalDraft`). Зафиксирован [ADR-014](decisions/0014-fullpage-journal-messenger-and-drafts.md). Спринт 2.4.1 успешно завершен и заархивирован в `docs/sprints/sprint-2.4.1.md`. |
| 2026-10-08 | Спринт 2.4.1 (Дневник занятий и Домашние задания) успешно завершен (68/68 E2E passed). |
| 2026-10-06 | Спринт 2.3.1 (Регулярные занятия, RRULE, Google Calendar Pattern) успешно завершен и заархивирован в `docs/sprints/sprint-2.3.1.md`. |

---

*Последнее обновление: 2026-10-08*
