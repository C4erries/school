# Спринт 2.4.1: Дневник занятий (Lesson Journal) и Домашние задания (Homework Management)

**Даты**: 2026-10-08  
**Статус**: ✅ Завершен (100% DoD)  
**Архитектурное решение**: [ADR-013](../decisions/0013-lesson-journal-and-homework-management.md) (Двухуровневый академический хаб: `lesson_journals` 1:1 + `homework_assignments` 1:N)

---

## 🎯 Цели и Результаты спринта

1. **Модель данных и миграция 000011**:
   - Миграция `000011_lesson_journal_and_homework.up.sql`:
     - Таблица `lesson_journals` (1:1 к `lessons`): тема занятия (`topic`), приватные заметки (`notes`), оценка усвоения (`performance_score` 1..5). Внешний ключ `lesson_id UNIQUE REFERENCES lessons(id) ON DELETE CASCADE`.
     - Таблица `homework_assignments` (1:N к `clients` и `lessons`): заголовок (`title`), описание (`description`), дедлайн (`due_date`), статус (`assigned`, `completed`, `not_done`), рецензия (`review_notes`).
   - Доменные сущности `LessonJournal` и `HomeworkAssignment` со строгой валидацией и 100% юнит-тестами.

2. **OpenAPI Контракт и Echo v4 генерация**:
   - 7 новых эндпоинтов в `backend/api/openapi/api.yaml`:
     - `GET /schedule/lessons/{id}/journal` — получение бандла отчета и заданий по уроку.
     - `PUT /schedule/lessons/{id}/journal` — создание/обновление отчета по уроку.
     - `GET /crm/clients/{id}/journal` — хронологический таймлайн уроков ученика.
     - `GET /crm/clients/{id}/homework` — список ДЗ ученика с фильтром по статусу.
     - `POST /crm/clients/{id}/homework` — выдача домашнего задания ученику.
     - `PATCH /homework/{id}` — обновление статуса и заметок проверки ДЗ.
     - `DELETE /homework/{id}` — удаление задания.
   - Кодогенерация через `make oapi`.

3. **Backend бизнес-логика и бесшовная материализация**:
   - PostgreSQL репозитории `journal_repository.go` и `homework_repository.go`.
   - `journal.Service`: поддержка бандлов, привязка заданий к датам и управление статусами.
   - В `occurrence.go` реализован `EnsurePhysicalLesson`: если репетитор сохраняет отчет по виртуальному слоту регулярной серии (ADR-012), урок на лету материализуется в таблице `lessons` со статусом `scheduled`, обеспечивая строгую целостность внешнего ключа.

4. **Frontend UI в стиле Apple Liquid Glass**:
   - **Расписание**:
     - На карточки занятий (Week, Day, List) добавлена отдельная кнопка «Дневник / ДЗ» (`BookOpen`).
     - Быстрое проведение по галочке `✓` моментально завершает урок в 1 клик и не блокируется.
     - `LessonJournalModal.tsx`: быстрый ввод темы, шкала понимания 1–5 (`LessonScoreChips.tsx`), заметки преподавателя, экспресс-проверка заданного ранее ДЗ (`LessonDueHomeworks.tsx`) и блок назначения нового ДЗ (`LessonAssignHomework.tsx`).
   - **CRM ученика**:
     - В карточку ученика (`ClientCard.tsx`) добавлена кнопка «Дневник & ДЗ».
     - `ClientJournalModal.tsx`: таймлайн уроков (`ClientLessonsTimeline.tsx`) и интерактивный список ДЗ с фильтрами и сменой статуса (`ClientHomeworkList.tsx`).
   - Все компоненты строго меньше 300 строк, стили полупрозрачного стекла с `backdrop-blur-md`.

5. **Регламенты и стандарты**:
   - Зафиксирован **Contract-First & Agent Continuity протокол** в `SUBAGENTS_GUIDE.md`, `AGENTS.md` и ADR-010.
   - Зафиксирован запрет на любые мутирующие команды Git для AI-агентов без прямого указания пользователя.
   - Зафиксирован запрет на рукописные моки (только `mockery v3+`) и требование декомпозиции `adapters/http`.

---

## 📊 Метрики качества и Тестирование

- **Docker E2E автотесты**: **68 из 68 тестов пройдены успешно (100% PASS)**:
  - 63 существующих теста расписания, финансов, CRM, серий и аналитики.
  - 5 комплексных тестов в `tests/api/phase2_schedule/test_journal_and_homework.py`:
    * `test_create_and_get_lesson_journal` (отчет по разовому уроку);
    * `test_journal_on_virtual_recurring_lesson` (авто-материализация слота серии при заполнении отчета);
    * `test_homework_lifecycle` (полный цикл ДЗ от создания до сдачи и удаления);
    * `test_client_journal_timeline` (хронологическая лента обучения в CRM);
    * `test_journal_and_homework_isolation` (мультиарендность и изоляция репетиторов).
- **Go юнит-тесты (`make test`)**: 100% PASS с детектором гонок `-race`.
- **Фронтенд линтер и сборка (`npm run lint && npm run build`)**: 0 ошибок, 0 предупреждений.

