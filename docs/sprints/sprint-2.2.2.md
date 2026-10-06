# Спринт 2.2.2: Аналитика и Статистика репетитора + Календарная интеграция (Google Calendar & iCal / Webcal) ✅

> **Период**: 2026-10-06  
> **Статус**: ✅ Завершен (DoD выполнен на 100%, 52/52 E2E тестов зеленые)  
> **Ветка**: `dev`

---

## 🎯 Цели спринта

1. **Стратегическая аналитика эффективности преподавателя**:
   - Переход от операционной бухгалтерии к комплексному анализу продуктивности.
   - Сводные показатели (KPI): фактически отработанные часы, доходимость занятий (Completion Rate), чистый доход репетитора с учетом партнерских комиссий, средняя реальная почасовая ставка.
   - Динамика ключевых метрик по неделям и месяцам на чистом адаптивном SVG-графике в стиле Apple Liquid Glass.
   - Распределение по форматам занятий (индивидуальные, парные, групповые) и долям в выручке.
   - Рейтинг учеников с сортировкой по выручке, часам и числу отмен.
2. **Интеграция с внешними календарями (Google Calendar, Apple Calendar, Outlook)**:
   - Live-подписка iCal / Webcal по стандарту RFC 5545 (`feed.ics?token=UUID`), обеспечивающая автоматическую синхронизацию и нативные push-уведомления на смартфонах без необходимости сложной верификации OAuth2 в Google Cloud Console.
   - Генерация уникального `calendar_token` для каждого преподавателя с возможностью ротации в 1 клик.
   - Разовый экспорт расписания в файл `.ics`.
3. **Apple Liquid Glass UI**:
   - Страница «Статистика» (`/teacher/analytics`) со стеклянными карточками KPI, селектором периодов и графиками.
   - Модальное окно «Синхронизация с календарем» (`CalendarSyncModal`) с кнопками быстрого перехода в Google Calendar и вызова системного Apple Calendar.
4. **E2E верификация**:
   - 52 E2E теста в Docker (`make test-e2e`) пройдены успешно, сборка фронтенда чистая (`npm run lint && npm run build`, 0 ошибок).

---

## 📦 Реализованные компоненты

### 1. Серверная часть (Go & PostgreSQL)
* **Миграция 000009**:
  - `backend/migrations/000009_calendar_token.up.sql` (добавление `calendar_token UUID NOT NULL DEFAULT gen_random_uuid()` в таблицу `users`).
* **Модели и Репозитории**:
  - Обновление `domain.User` (`CalendarToken uuid.UUID`).
  - Методы в `UserRepository`: `GetByCalendarToken`, `RotateCalendarToken`.
* **Сервис аналитики `AnalyticsService`** (`backend/internal/application/analytics/service.go`):
  - `GetOverview`: подсчет отработанных часов, процента доходимости, валовой выручки, вычета партнерских % и чистой почасовой ставки.
  - `GetDynamics`: группировка по неделям или месяцам за заданный интервал дат.
  - `GetFormatStats`: доли часов и выручки по форматам `individual`, `pair`, `group`.
  - `GetClientStats`: рейтинг учеников с ранжированием по часам, выручке и отменам.
* **Сервис календаря `CalendarService`** (`backend/internal/application/calendar/service.go`):
  - Полноценная генерация iCalendar потока (RFC 5545): `BEGIN:VCALENDAR`, `PRODID:-//School Tutor Assistant//RU`, `VERSION:2.0`, `BEGIN:VEVENT`, форматирование `DTSTART`/`DTEND` в UTC, описание локаций/онлайн-ссылок и контактов учеников.
  - Выдача потока `GET /api/v1/integrations/calendar/feed.ics` по персональному токену без JWT-авторизации (для внешних календарных клиентов).
  - Эндпоинты настроек `/settings`, ротации токена `/rotate-token` и экспорта `/export`.
* **Тестирование**:
  - Написаны подробные юнит-тесты `backend/internal/application/analytics/service_test.go` и `backend/internal/application/calendar/service_test.go`.

### 2. Клиентская часть (React, TypeScript, Tailwind)
* **Навигация**: добавлен пункт «Статистика» с иконкой `TrendingUp` в `AppNavbar.tsx`.
* **Страница `/teacher/analytics`** (`frontend/src/pages/teacher/TeacherAnalyticsPage.tsx`):
  - `AnalyticsPeriodSelector.tsx`: быстрое переключение периодов («Этот месяц», «3 месяца», «Полгода», «Год») и кастомный интервал дат.
  - `AnalyticsKPICards.tsx`: 4 стеклянные карточки с градиентными бейджами.
  - `AnalyticsDynamicsChart.tsx`: интерактивный SVG-график (линии и столбцы с тултипами) без тяжелых зависимостей.
  - `AnalyticsFormatsChart.tsx`: диаграмма долей форматов занятий.
  - `AnalyticsClientsTable.tsx`: таблица рейтинга учеников с сортировкой.
* **Модальное окно синхронизации `CalendarSyncModal.tsx`**:
  - Быстрое добавление в Google Calendar (`calendar.google.com/calendar/render?cid=webcal://...`).
  - Быстрое добавление в Apple Calendar (`webcal://...`).
  - Копирование персональной ссылки фида, скачивание `.ics`, перевыпуск скомпрометированного токена.
  - Кнопка вызова добавлена в шапку расписания (`TeacherSchedulePage.tsx`) и на страницу аналитики.

---

## 🎯 Результаты и приемка
- Миграция 000009 применена в Docker.
- Все эндпоинты `/analytics/*` и `/integrations/calendar/*` работают в соответствии с контрактом OpenAPI.
- E2E тесты в Docker: 52/52 пройдены (добавлены тесты в `tests/api/phase2_schedule/test_analytics_and_calendar.py`).
- Сборка фронтенда: 0 ошибок TypeScript и ESLint.

