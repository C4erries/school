# Текущий спринт / ближайшие задачи

> Этот файл — **единственная точка правды** о том, что делаем прямо сейчас.
> Обновляется перед началом каждого спринта и после завершения задач.
> Завершенные спринты архивируются в каталоге `docs/sprints/`.

## Текущая фаза: 2.4.1 — Дневник занятий (Lesson Journal) и Домашние задания (Homework Management) 🔵 (В РАБОТЕ)

---

## 🧠 Аналитика задач спринта (Роль: CEO, PM & Lead Architect)

### 1. Бизнес-обоснование (CEO View)
* **Проблема**: Репетитор ведет занятия по 10–25 ученикам параллельно. Вся академическая память («что проходили на прошлом уроке», «какое ДЗ задали», «почему у Вани пробелы в тригонометрии») хранится в разрозненных блокнотах, переписках в Telegram или в голове. Это приводит к потере контекста перед уроком, смазанным результатам и неуверенности в общении с родителями.
* **Решение**: Встроенный, компактный «Дневник занятий»:
  - Мгновенная фиксация темы урока, замечаний репетитора и экспресс-оценки понимания (1–5).
  - Привязка и учет домашних заданий со статусами сдачи (`assigned` / `completed` / `not_done`).
  - Быстрый доступ к контексту перед уроком в расписании + полный таймлайн обучения в карточке ученика.
* **Бизнес-эффект**: Резкий рост профессиональной ценности платформы для репетитора (Stickiness & LTV), переход от чистого расписания к полноценной системе управления обучением (LMS/CRM-гибрид), подготовка базы к Фазе 4 (личный кабинет ученика) и Фазе 5 (проверка фото тетрадей).

---

### 2. Архитектура хранения: Двухуровневый академический хаб ([ADR-013](decisions/0013-lesson-journal-and-homework-management.md))
* **Сущность `lesson_journals` (1:1 к `lessons`)**:
  - Хранит дидактический отчет по уроку: пройденная тема (`topic`), приватные заметки преподавателя (`notes`), оценка усвоения (`performance_score` от 1 до 5).
  - Привязана внешним ключом `lesson_id UNIQUE REFERENCES lessons(id) ON DELETE CASCADE`.
* **Сущность `homework_assignments` (1:N к `clients` и `lessons`)**:
  - Самостоятельная сущность учебного задания: заголовок (`title`), текст/описание (`description`), срок выполнения (`due_date`), статус (`assigned`, `completed`, `not_done`), заметки по проверке (`review_notes`).
  - Опционально ссылается на `assigned_lesson_id` (урок, на котором задание было выдано).
* **Взаимодействие с виртуальными слотами серий ([ADR-012](decisions/0012-recurring-lessons-rrule-and-calendar-sync.md))**:
  - Если дневник или ДЗ создаются для виртуального слота циклической серии (еще не существующего в БД), сервис на лету находит слот, вычисляет детерминированный `VirtualLessonID(series.ID, slot.OriginalStartTime)` и **материализует** физическую запись в `lessons`, после чего привязывает отчет.

---

### 3. Продуктовый UX: Ненавязчивый флоу и две точки входа
1. **Сохранение скорости работы в календаре**:
   - Клик по зеленой галочке `✓` **НЕ открывает** принудительных модалок и сохраняет моментальное завершение занятия в 1 клик со списанием абонемента.
   - На карточке урока в расписании появляется **отдельная кнопка «Дневник / ДЗ»** (иконка `BookOpen`), доступная **всегда** — до занятия, во время него или после завершения.
2. **Модальное окно `LessonJournalModal` (в расписании)**:
   - Ввод темы урока (быстрый input).
   - Легкая шкала оценки понимания (чипы `1 2 3 4 5` без утяжеления интерфейса).
   - Заметки преподавателя (текстовое поле).
   - Блок «Домашнее задание»: быстрое назначение ДЗ к следующему занятию.
   - Экспресс-проверка: если к этому уроку было задано ДЗ, его можно в 1 клик перевести в статус «Выполнено» / «Не сделано».
3. **Хронологический архив в карточке ученика (`ClientCard.tsx`)**:
   - Новая кнопка «Дневник & ДЗ» в карточке ученика на странице `/teacher/clients`.
   - Модальное окно `ClientJournalModal`:
     - Вкладка «История занятий»: упорядоченный по датам список тем, оценок и заметок.
     - Вкладка «Домашние задания»: карточки заданий с бейджами статусов, фильтрами («Все», «В работе», «Выполнено», «Не сдано») и возможностью смены статуса.

---

## 📋 Таблица задач спринта 2.4.1

| # | Задача | Статус | Приоритет | Ответственный / Субагент | Заметки |
|---|--------|--------|-----------|---------------------------|---------|
| 1 | Миграция БД 000011 и Доменные модели `LessonJournal` и `HomeworkAssignment` | 🔲 To Do | Критический | `backend_developer` | Таблицы `lesson_journals` и `homework_assignments`, доменные типы, валидация. |
| 2 | OpenAPI контракт: схемы дневника, отчетов и домашних заданий | 🔲 To Do | Критический | `backend_developer` | Эндпоинты `/schedule/lessons/{id}/journal`, `/crm/clients/{id}/journal`, `/crm/clients/{id}/homework`, `/homework/{id}`, запуск `make oapi`. |
| 3 | Backend: Сервисы дневника и ДЗ, PostgreSQL адаптеры, материализация слотов | 🔲 To Do | Высокий | `backend_developer` | Репозитории в `adapters/postgres/`, слой use-cases в `application/`, материализация виртуальных уроков. |
| 4 | Frontend: Модальное окно `LessonJournalModal`, интеграция с карточками расписания | 🔲 To Do | Высокий | `frontend_developer` | Кнопка `BookOpen` в расписании (неделя, день, список), модалка темы, оценки (1–5), заметок и ДЗ в стиле Liquid Glass. |
| 5 | Frontend: Модальное окно `ClientJournalModal` и лента в карточке ученика | 🔲 To Do | Высокий | `frontend_developer` | Кнопка «Дневник & ДЗ» в `ClientCard.tsx`, таймлайн тем и история домашних заданий с фильтрами. |
| 6 | E2E сценарии в Docker и регрессионная верификация | 🔲 To Do | Критический | `qa_engineer` | Автотесты полного цикла: заполнение отчета, материализация виртуального урока, смена статуса ДЗ, таймлайн клиента (63+ passed). |

---

## 📋 Детальное Микро-ТЗ спринта 2.4.1

### Задача 1: Миграция БД 000011 (`backend/migrations/000011_lesson_journal_and_homework.up.sql`)
* Создание таблицы `lesson_journals`:
  - `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
  - `lesson_id UUID NOT NULL UNIQUE REFERENCES lessons(id) ON DELETE CASCADE`
  - `client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE`
  - `teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
  - `topic VARCHAR(255) NOT NULL`
  - `notes TEXT`
  - `performance_score INT CHECK (performance_score BETWEEN 1 AND 5)`
  - `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
  - `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
  - Индексы: `idx_lesson_journals_lesson_id`, `idx_lesson_journals_client_id`, `idx_lesson_journals_teacher_id`.
* Создание таблицы `homework_assignments`:
  - `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
  - `client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE`
  - `teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
  - `assigned_lesson_id UUID REFERENCES lessons(id) ON DELETE SET NULL`
  - `title VARCHAR(255) NOT NULL`
  - `description TEXT`
  - `due_date DATE`
  - `status VARCHAR(50) NOT NULL DEFAULT 'assigned' CHECK (status IN ('assigned', 'completed', 'not_done'))`
  - `review_notes TEXT`
  - `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
  - `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
  - Индексы: `idx_homework_assignments_client_id`, `idx_homework_assignments_teacher_id`, `idx_homework_assignments_status`.
* Доменные сущности в `backend/internal/domain/journal.go` и `backend/internal/domain/homework.go`.

---

### Задача 2: Контракт OpenAPI (`backend/api/openapi/api.yaml`)
* **Схемы DTO**:
  - `LessonJournalResponse`: id, lesson_id, client_id, topic, notes, performance_score, created_at, updated_at.
  - `UpsertLessonJournalRequest`: topic (string), notes (optional string), performance_score (optional int 1..5).
  - `HomeworkAssignmentResponse`: id, client_id, assigned_lesson_id, title, description, due_date, status, review_notes, created_at.
  - `CreateHomeworkRequest`: title (string), description (optional), due_date (optional date), assigned_lesson_id (optional uuid).
  - `UpdateHomeworkStatusRequest`: status (`assigned`, `completed`, `not_done`), review_notes (optional).
  - `LessonJournalBundleResponse`: journal (`LessonJournalResponse` nullable), assigned_homeworks (`[]HomeworkAssignmentResponse`), due_homeworks (`[]HomeworkAssignmentResponse` - ДЗ, выданные ранее к этой дате/уроку).
* **Эндпоинты**:
  - `GET /api/v1/schedule/lessons/{id}/journal`: получение отчета по уроку и связанных ДЗ.
  - `PUT /api/v1/schedule/lessons/{id}/journal`: сохранение/обновление отчета по уроку.
  - `GET /api/v1/crm/clients/{id}/journal`: хронологический список отчетов по урокам ученика.
  - `GET /api/v1/crm/clients/{id}/homework`: список всех ДЗ ученика с query-фильтром `status`.
  - `POST /api/v1/crm/clients/{id}/homework`: создание нового домашнего задания.
  - `PATCH /api/v1/homework/{id}`: обновление статуса и рецензии на ДЗ.
  - `DELETE /api/v1/homework/{id}`: удаление ДЗ.
* Запуск `make oapi`.

---

### Задача 3: Backend бизнес-логика и репозитории
* **Репозитории PostgreSQL**:
  - `backend/internal/infrastructure/api/adapters/postgres/schedule/journal_repository.go`
  - `backend/internal/infrastructure/api/adapters/postgres/crm/homework_repository.go`
* **Application Services**:
  - Пакет `internal/application/journal/` (или расширение сервиса расписания):
    - Получение бандла дневника для урока (журнал + выданные ДЗ + ДЗ к проверке).
    - Upsert журнала урока: если урок виртуальный — вызов материализации из `schedule.Service` с детерминированным `VirtualLessonID`, затем сохранение `LessonJournal`.
  - Пакет `internal/application/homework/`:
    - Выдача ДЗ ученику.
    - Смена статуса (`assigned` -> `completed` / `not_done`).
* **HTTP Handlers**:
  - Обработчики на Echo v4 в `infrastructure/api/adapters/http/schedule/journal_handler.go` и `infrastructure/api/adapters/http/crm/homework_handler.go`.
  - Регистрация в DI контейнере `di.Container`.

---

### Задача 4: Frontend Расписание — `LessonJournalModal`
* Кнопка «Дневник / ДЗ» на карточках уроков:
  - Иконка `BookOpen` в `ScheduleWeekView.tsx`, `ScheduleDayView.tsx`, `ScheduleListView.tsx`.
  - Индикатор заполненности (например, синяя точка или мягкая подсветка, если тема урока уже записана).
* Компонент `LessonJournalModal.tsx`:
  - Быстрый ввод темы занятия.
  - Шкала оценки (1–5) в стиле Apple Liquid Glass: мягкие чипы с интерактивным выбором.
  - Заметки преподавателя (textarea).
  - Блок создания ДЗ к следующему уроку (заголовок, описание, дедлайн).
  - Блок проверки предыдущего ДЗ (если к текущему уроку было невыполненное задание — кнопки `[✓ Сдано]` и `[✗ Не сдано]`).
* Вызовы API через типизированный модуль `frontend/src/api/journal.ts`.

---

### Задача 5: Frontend CRM — `ClientJournalModal`
* В карточке ученика `ClientCard.tsx`:
  - Кнопка «Дневник & ДЗ» (рядом с кнопками баланса и абонементов).
  - Бейдж количества несданных ДЗ (если есть статус `assigned`).
* Компонент `ClientJournalModal.tsx`:
  - Две вкладки: «История уроков» и «Домашние задания».
  - Вкладка «История уроков»: таймлайн занятий (дата, тема, оценка усвоения, заметки преподавателя).
  - Вкладка «Домашние задания»: список карточек заданий, фильтрация по статусам (`Все`, `В работе`, `Сдано`, `Не выполнено`), возможность отметить сдачу прямо из списка.

---

### Задача 6: Комплексное тестирование и верификация полного цикла
* Модульные тесты Go (`make test`):
  - Проверка валидации оценок 1–5.
  - Проверка материализации виртуального урока при сохранении дневника.
  - Проверка изоляции репетиторов (RBAC/мультиарендность: репетитор не может видеть/править чужие отчеты).
* E2E тесты в Docker (`tests/api/phase2_schedule/test_journal_and_homework.py`):
  - Создание отчета по обычному уроку.
  - Создание отчета по виртуальному слоту регулярной серии (проверка материализации в `lessons`).
  - Выдача ДЗ, смена статуса на `completed`.
  - Получение ленты дневника ученика в CRM.
* 100% зеленые юнит-тесты и E2E тесты в Docker (`make test-e2e`).
* Чистая сборка фронтенда (`npm run lint && npm run build`).

---

## 🎯 Definition of Done (DoD) Спринта 2.4.1
- [ ] Применена миграция `000011_lesson_journal_and_homework.up.sql`.
- [ ] Контракт OpenAPI обновлен схемами и эндпоинтами дневника и ДЗ, код сгенерирован (`make oapi`).
- [ ] Реализовано сохранение и чтение отчета по уроку (тема, оценка 1–5, заметки).
- [ ] Реализовано управление домашними заданиями (выдача, дедлайн, статусы `assigned`, `completed`, `not_done`).
- [ ] При сохранении отчета по виртуальному слоту серии урок корректно материализуется в БД.
- [ ] Проведение урока по галочке `✓` в расписании остается мгновенным в 1 клик.
- [ ] Карточка урока в расписании имеет отдельную кнопку «Дневник / ДЗ», открывающую `LessonJournalModal`.
- [ ] В карточке ученика (`ClientCard.tsx`) доступен просмотр истории занятий и статусов ДЗ (`ClientJournalModal.tsx`).
- [ ] Строго соблюдены правила Apple Liquid Glass (полупрозрачные карточки, отсутствие `bg-white`).
- [ ] Все созданные и измененные файлы кода не превышают 300–400 строк.
- [ ] Все юнит-тесты (`make test`) и E2E тесты в Docker (`make test-e2e`) зеленые (100% PASS).
- [ ] Сборка фронтенда чистая (`npm run lint && npm run build`, 0 ошибок и предупреждений).

---

## Лог изменений

| Дата | Что изменилось |
|------|---------------|
| 2026-10-08 | **Сформирован Спринт 2.4.1 (Фаза 2.4)**: Дневник занятий (Lesson Journal) и Управление домашними заданиями (Homework Management). Зафиксирован [ADR-013](decisions/0013-lesson-journal-and-homework-management.md) (Двухуровневый академический хаб: `lesson_journals` 1:1 + `homework_assignments` 1:N). Разработан ненавязчивый UX (отдельная кнопка отчета, независимое проведение в 1 клик, шкала понимания 1–5, таймлайн в карточке ученика). |
| 2026-10-06 | Спринт 2.3.1 (Регулярные занятия, RRULE, Google Calendar Pattern) успешно завершен и заархивирован в `docs/sprints/sprint-2.3.1.md`. |

---

*Последнее обновление: 2026-10-08*
