# Текущий спринт / ближайшие задачи

> Этот файл — **единственная точка правды** о том, что делаем прямо сейчас.
> Обновляется перед началом каждого спринта и после завершения задач.
> Завершенные спринты архивируются в каталоге `docs/sprints/`.

## Текущая фаза: 2.3.1 — Регулярные занятия (Recurring Lessons & RRULE), Google Calendar Pattern и Двусторонняя совместимость с Google Календарем 🔵 (В РАБОТЕ)

---

## 🧠 Аналитика задач спринта (Роль: CEO, PM & Lead Architect)

### 1. Бизнес-обоснование (CEO View)
* **Проблема**: 90% расписания репетитора — это постоянная сетка на учебный год («Вторник и Четверг в 17:00»). Отсутствие серий заставляло преподавателя каждую неделю тратить время на рутинное прокликивание 30–40 одинаковых уроков.
* **Решение**: Репетитор один раз настраивает циклический слот — система автономно держит расписание на весь учебный год.
* **Бизнес-эффект**: Резкий рост удержания (Retention) преподавателей, максимальная автоматизация финансового прогноза (`/analytics/forecast`) и закрытие главного барьера перехода с Google Календаря.

---

### 2. Архитектура хранения: Подход Б (Виртуальные слоты + Материализация исключений, [ADR-012](decisions/0012-recurring-lessons-rrule-and-calendar-sync.md))
* **Проблема**: Генерация сотен физических строк в таблице `lessons` раздувает базу данных, ломает бессрочные серии и делает массовые переносы ресурсоемкими и транзакционно уязвимыми.
* **Решение**:
  - Таблица `lesson_series` хранит правило RFC 5545 RRULE (`FREQ=WEEKLY;BYDAY=TU,TH`), время дня, длительность, формат и горизонт дат.
  - Таблица `lessons` хранит только разовые уроки, фактически проведенные занятия (`status = completed`) и точечные исключения (`series_id` + `original_start_time` / Recurrence-ID).
  - При запросе расписания за неделю/месяц (`GET /schedule/lessons?from=...&to=...`) сервис динамически генерирует виртуальные слоты по RRULE и накладывает их на физические исключения из БД.

---

### 3. Полный паттерн управления изменениями (The Google Calendar Pattern)
При редактировании или отмене занятия из серии преподаватель выбирает область действия:
1. **«Только этот урок» (`scope: this_only`)**:
   - Точечная отмена: создается запись в `lessons` со статусом `cancelled`, `series_id` и `original_start_time`. Виртуальный слот на эту дату подавляется.
   - Точечный перенос/изменение: создается запись в `lessons` с новым временем/кабинетом, связкой с `series_id` и `original_start_time`.
2. **«Этот и все последующие» (`scope: this_and_following`)**:
   - Исходная серия обрезается: `until_date = occurrence_date - 1 день`.
   - Начиная с даты выбранного урока создается новая `lesson_series` с обновленными параметрами (время, кабинет, формат).
3. **«Все уроки серии» (`scope: all_in_series`)**:
   - Обновляются глобальные поля исходной записи `lesson_series`.

---

### 4. Взаимодействие с Абонементами и Финансовым Прогнозом
* **Абонементы**: Списание баланса часов ученика происходит строго по факту проведения (`CompleteLesson`). Будущие виртуальные слоты баланс не уменьшают.
* **Финансовый прогноз (`/analytics/forecast`)**: Виртуальные слоты в диапазоне дат автоматически включаются в расчет планируемой нагрузки и ожидаемого дохода по тарифам клиентов (за исключением отмененных дат).

---

### 5. Совместимость с Google Calendar: iCal Feed и Импорт `.ics`
1. **Экспорт и Live-подписка (`feed.ics` и `export.ics`)**:
   - Выгрузка серий по стандарту RFC 5545 с директивами `RRULE`, `EXDATE` (отмененные даты) и кастомными VEVENT с `RECURRENCE-ID` (перенесенные вхождения).
   - Google Calendar и Apple Calendar нативно понимают эти правила и показывают аккуратную циклическую сетку.
2. **Импорт расписания из Google Calendar (`POST /api/v1/integrations/calendar/import`)**:
   - Загрузка стандартного файла `.ics`, выгруженного из Google Календаря.
   - Автоматический парсинг VEVENT (разовые события и серии RRULE) и создание уроков и серий в расписании репетитора.

---

## 📋 Таблица задач спринта 2.3.1

| # | Задача | Статус | Приоритет | Ответственный / Субагент | Заметки |
|---|--------|--------|-----------|---------------------------|---------|
| 1 | Миграция БД 000010 и Доменная модель `LessonSeries` | ✅ Done | Критический | `backend_developer` | Таблица `lesson_series`, поля `lessons.series_id`, `lessons.original_start_time`, доменные типы. |
| 2 | OpenAPI контракт: схемы серий, scope изменений и импорт `.ics` | ✅ Done | Критический | `backend_developer` | Эндпоинты `/schedule/series`, scope в `/schedule/lessons`, эндпоинт `/integrations/calendar/import`. |
| 3 | Backend: Сервис серий, генератор слотов RRULE и Google Calendar Pattern | ✅ Done | Высокий | `backend_developer` | Алгоритм генерации слотов, слияние в `ListLessons`, обработка `this_only`, `this_and_following`, `all_in_series`. |
| 4 | Backend: Интеграция с Абонементами, Прогнозом и Google Calendar (iCal Feed + Import) | ✅ Done | Высокий | `backend_developer` | Проведение виртуального урока со списанием абонемента, учет в `/analytics/forecast`, RFC 5545 RRULE в `feed.ics`, парсер `.ics`. |
| 5 | Frontend: UI создания серий, диалог Google Calendar Pattern и импорт `.ics` | ✅ Done | Высокий | `frontend_developer` | Тумблер повторений в модалке урока, иконка повторений в сетке, модалка выбора scope, кнопка импорта в `CalendarSyncModal`. |
| 6 | E2E сценарии в Docker и регрессионная верификация | ✅ Done | Критический | `qa_engineer` | Полный цикл автотестов: создание серии, отображение в неделе, исключения, разделение серии, импорт `.ics` (63/63 passed). |

---

## 📋 Детальное Микро-ТЗ спринта 2.3.1

### Задача 1: Миграция БД 000010 (`backend/migrations/000010_lesson_series.up.sql`)
* Создание таблицы `lesson_series`:
  - `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
  - `teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
  - `client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE`
  - `classroom_id UUID REFERENCES classrooms(id) ON DELETE SET NULL`
  - `title VARCHAR(255) NOT NULL DEFAULT 'Занятие'`
  - `rrule VARCHAR(255) NOT NULL` (например, `FREQ=WEEKLY;BYDAY=TU,TH`)
  - `start_time_of_day TIME NOT NULL`
  - `duration_minutes INT NOT NULL`
  - `format VARCHAR(50) NOT NULL` (`individual`, `pair`, `group`)
  - `location_or_url TEXT`
  - `notes TEXT`
  - `start_date DATE NOT NULL`
  - `until_date DATE` (nullable)
  - `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
  - `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
* Добавление полей в `lessons`:
  - `series_id UUID REFERENCES lesson_series(id) ON DELETE CASCADE`
  - `original_start_time TIMESTAMPTZ` (метка исходного вхождения по RFC 5545 Recurrence-ID)
* Доменные сущности в `backend/internal/domain/lesson_series.go`.

---

### Задача 2: Контракт OpenAPI (`backend/api/openapi/api.yaml`)
* **Новые эндпоинты**:
  - `POST /api/v1/schedule/series`: создание регулярной серии.
  - `GET /api/v1/schedule/series`: получение активных серий преподавателя.
  - `GET /api/v1/schedule/series/{id}`: детали серии.
  - `PUT /api/v1/schedule/series/{id}`: обновление всей серии.
  - `DELETE /api/v1/schedule/series/{id}`: удаление всей серии.
  - `POST /api/v1/integrations/calendar/import`: загрузка `.ics` файла (multipart/form-data).
* **Расширение существующих эндпоинтов**:
  - `PATCH /api/v1/schedule/lessons/{id}`: добавление query/body параметра `scope` (`this_only`, `this_and_following`, `all_in_series`).
  - `POST /api/v1/schedule/lessons/{id}/cancel`: параметр `scope` для отмены только текущего урока или всей цепочки.
  - `LessonResponse`: флаги `is_recurring: boolean`, `series_id?: uuid`.
* Выполнить `make oapi`.

---

### Задача 3: Бизнес-логика серий и алгоритм виртуальных слотов
* Репозиторий `LessonSeriesRepository` в `adapters/postgres/schedule/series_repository.go`.
* Модуль вычисления дат повторений (RRULE weekly engine):
  - По входным дням недели и `start_date`/`until_date` генерирует даты вхождения в интервале `[from, to]`.
* Метод `ListLessons` сервиса расписания:
  - Выборка физических уроков за интервал.
  - Генерация виртуальных уроков серий.
  - Слияние и фильтрация исключений (подавление отмененных, подмена перенесенных).
* Методы модификации:
  - `UpdateRecurringLesson(ctx, input)` с обработкой `this_only`, `this_and_following`, `all_in_series`.

---

### Задача 4: Абонементы, Прогноз и Google Calendar (iCal Feed & Import)
* Метод `CompleteLesson`: если урок виртуальный — материализовать в БД и списать часы с абонемента ученика.
* Сервис аналитики: учет виртуальных уроков серий при расчете прогноза выручки `/analytics/forecast`.
* Генератор `feed.ics`: формирование `RRULE`, параметров `EXDATE` и VEVENT с `RECURRENCE-ID`.
* Импорт `import.ics`: парсинг входящего iCalendar файла и сохранение серий/уроков.

---

### Задача 5: Frontend UI (React + Apple Liquid Glass)
* Модалка `CreateLessonModal.tsx`:
  - Тумблер `[✓] Повторять еженедельно`.
  - Кнопки-чипы выбора дней недели: `[Пн] [Вт] [Ср] [Чт] [Пт] [Сб] [Вс]`.
  - Дата окончания (до определенного числа или бессрочно).
* Отображение в расписании (`ScheduleWeekView.tsx`, `ScheduleDayView.tsx`):
  - Значок серии (иконка `RefreshCw` / `Repeat`) на карточках уроков.
* Модалка выбора области действия (`RecurrenceScopeModal.tsx`):
  - Диалог выбора: «Только этот урок», «Этот и все последующие», «Все уроки серии».
* Модалка синхронизации (`CalendarSyncModal.tsx`):
  - Блок «Импорт расписания из Google Календаря» (загрузка файла `.ics`).

---

### Задача 6: Тестирование и верификация полного цикла
* Модульные тесты генератора RRULE и исключений.
* E2E автотесты в `tests/api/phase2_schedule/test_recurring_lessons.py`.
* 100% зеленые юнит-тесты (`make test`) и E2E тесты в Docker (`make test-e2e`).

---

## 🎯 Definition of Done (DoD) Спринта 2.3.1
- [x] Применена миграция 000010 для серий и исключений.
- [x] Контракт OpenAPI обновлен эндпоинтами серий и импорта `.ics`, код сгенерирован (`make oapi`).
- [x] Реализовано создание еженедельных серий уроков по выбранным дням недели.
- [x] Реализован динамический расчет виртуальных слотов в расписании без раздувания базы данных.
- [x] Реализован Google Calendar Pattern при редактировании и отмене: «Только этот», «Этот и последующие», «Вся серия».
- [x] Проведение виртуального урока корректно материализует его и списывает баланс абонемента.
- [x] Виртуальные уроки серий учитываются в финансовом прогнозе (`/analytics/forecast`).
- [x] Календарный фид (`feed.ics`) генерирует стандартный RFC 5545 RRULE с поддержкой EXDATE/RECURRENCE-ID.
- [x] Реализован импорт расписания из внешнего `.ics` файла (Google Calendar).
- [x] Интерфейс расписания на фронтенде поддерживает настройку серий, индикацию и модалку выбора области действия.
- [x] Все юнит-тесты (`make test`) и E2E тесты в Docker (`make test-e2e`) успешно проходят (63/63 passed).
- [x] Сборка фронтенда чистая (`npm run lint && npm run build`, 0 ошибок).

---

## Лог изменений

| Дата | Что изменилось |
|------|---------------|
| 2026-10-06 | **Сформирован Спринт 2.3.1 (Фаза 2.3)**: Регулярные занятия (Recurring Lessons & RFC 5545 RRULE), Google Calendar Pattern (Только этот / Этот и последующие / Вся серия), двусторонняя совместимость с Google Календарем (Live iCal feed с RRULE/EXDATE и импорт `.ics` файлов). Зафиксирован [ADR-012](decisions/0012-recurring-lessons-rrule-and-calendar-sync.md). Спринт 2.2.4 заархивирован в `docs/sprints/sprint-2.2.4.md`. |

---

*Последнее обновление: 2026-10-06*
