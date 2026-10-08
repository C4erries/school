# Спринт 2.4.2: Полноэкранный Дневник-Мессенджер (Study Stream), Быстрые заметки ученика и Автосохранение черновиков

**Даты**: 2026-10-08  
**Статус**: ✅ Завершен (100% DoD)  
**Архитектурное решение**: [ADR-014](../decisions/0014-fullpage-journal-messenger-and-drafts.md) (Полноэкранный рабочий центр репетитора, полиморфный поток обучения и Zero Data Loss)

---

## 🎯 Цели и Результаты спринта

1. **Модель данных и миграция 000012**:
   - Миграция `000012_client_notes.up.sql` и `down.sql`:
     - Таблица `client_notes`: быстрая фиксация свободных заметок репетитора вне уроков (`id`, `client_id`, `teacher_id`, `content`, `created_at`, `updated_at`).
     - Внешние ключи с каскадным удалением `ON DELETE CASCADE` и индексы по `client_id`, `teacher_id`, `created_at`.
   - Доменная сущность `ClientNote` (`backend/internal/domain/client_note.go`) со строгой валидацией контента и юнит-тестами.

2. **OpenAPI Контракт и кодогенерация**:
   - Новые эндпоинты в `backend/api/openapi/api.yaml`:
     - `GET /crm/clients/{id}/stream` — объединенный хронологический поток обучения (`StudyStreamResponse`).
     - `POST /crm/clients/{id}/notes` — создание свободной заметки ученика (`CreateClientNoteRequest`).
     - `DELETE /crm/clients/notes/{id}` — удаление свободной заметки (204 No Content).
     - Поле `last_lesson_at` в схеме `ClientResponse` для сортировки учеников по актуальности проведенных занятий.
   - Выполнена кодогенерация через `make oapi`.

3. **Backend бизнес-логика и агрегация потока**:
   - Репозиторий `notes_repository.go` (`backend/internal/infrastructure/api/adapters/postgres/crm/`).
   - Подсчет `last_lesson_at` в `client_repo_queries.go` через `MAX(lessons.start_time)`.
   - В `schedule.Service` реализован метод `GetUpcomingLesson(ctx, clientID, teacherID)`.
   - В `journal.Service`:
     - Реализованы CRUD-методы заметок с multi-tenant RBAC изоляцией.
     - Метод `GetClientStudyStream`: полиморфная агрегация отчетов уроков (`lesson_report`) и заметок (`note`), хронологическая сортировка `Timestamp ASC` (мессенджер-поток: старые сверху, новые снизу) и компактная плашка ближайшего запланированного урока `upcoming_lesson`.
   - Echo v4 хэндлеры в `journal_homework_server.go` и регистрация в DI-контейнере `container.go`.

4. **Frontend интерфейс в стиле Apple Liquid Glass**:
   - **Хук `useJournalDraft` (Zero Data Loss)**:
     - Дебаунс 300ms для сохранения полей отчета и ДЗ в `localStorage`.
     - Надежный метод `getDraft()` без циклических ре-рендеров и паразитных сбросов каретки ввода.
     - Индикатор «Восстановлен черновик» в шапке модалки и очистка черновика при успешном сохранении.
     - Автоподтягивание темы занятия из карточки урока расписания (`fallbackTopic`).
   - **Полноэкранная страница `/teacher/journal` (Study Stream Messenger)**:
     - Сайдбар учеников с поиском и сортировкой по дате последнего проведенного урока (`last_lesson_at` DESC).
     - Центральный хронологический поток сообщений с автоскроллом вниз.
     - Карточки урочных отчетов с баджами оценок 1–5 и списками домашних заданий.
     - Карточки заметок преподавателя с возможностью удаления.
     - Компактный баннер ближайшего запланированного урока с кнопкой «Задать план».
     - Нижний Input Bar для мгновенной отправки быстрых заметок по ученику.
   - **Сквозная навигация**:
     - В `LessonJournalModal` добавлена кнопка «Открыть в журнале ↗».
     - В карточке ученика CRM (`ClientCard.tsx`) кнопка «Дневник & ДЗ» ведет на `/teacher/journal?clientId=...`.
     - Добавлен пункт меню «Журнал» в `AppNavbar.tsx`.

---

## 📊 Метрики качества и Тестирование

- **Docker E2E автотесты**: **72 из 72 тестов пройдены успешно (100% PASS)**:
  - 68 существующих тестов расписания, финансов, CRM, серий, дневника и аналитики.
  - 4 комплексных сценария в `tests/api/phase2_schedule/test_study_stream.py`:
    * `test_client_notes_lifecycle` (создание, валидация пустого ввода, удаление и 404 при повторном удалении);
    * `test_study_stream_aggregation_and_ordering` (объединение отчетов и заметок, сортировка ASC, данные upcoming_lesson);
    * `test_client_last_lesson_at_sorting` (расчет времени последнего урока и сортировка в CRM);
    * `test_multi_tenant_isolation` (изоляция между разными преподавателями).
- **Go юнит-тесты (`make test`)**: 100% PASS без data races.
- **Frontend линтер и сборка (`npm run lint && npm run build`)**: 0 ошибок, 0 ворнингов.

