# Спринт 2: SaaS-Ассистент Репетитора (MVP v1 & v1.5)

> Статус: ✅ **ЗАВЕРШЕН** (2026-10-05)

## Цель спринта
Выполнить пивот продукта от Платформы к **Tutor Assistant** (ADR 005). Отвязать учеников от пользователей (заменить на `clients`), добавить логику абонементов (`client_subscriptions`), адаптировать расписание (нахлесты) и создать финансовый дашборд с расчетом метрик (потенциальная выручка, реальный доход с учетом тегов/комиссий, средняя ставка в час).

## Выполненные задачи

| # | Задача | Статус | Ответственный | Заметки |
|---|--------|--------|---------------|---------|
| 1 | Миграция БД: `clients`, `client_subscriptions`, обновление `lessons` | ✅ Done | `backend-dev` | Миграция `000004_tutor_assistant_pivot.up.sql` применена. `clients`, `client_subscriptions` созданы. `teacher_students` удалена. |
| 2 | Доменные модели и Use Cases для Clients и Subscriptions | ✅ Done | `backend-dev` | Созданы модели `Client`, `ClientSubscription`, сервисы `crm` и списание баланса. |
| 3 | Рефакторинг модуля Lessons | ✅ Done | `backend-dev` | Сущность `Lesson` и `schedule.Service` переписаны на `ClientID`, убраны аппрувы, поддержаны нахлесты. |
| 4 | Финансовый движок (Дашборд) на бэкенде | ✅ Done | `backend-dev` | Создан `dashboard.Service`: расчет Gross Potential Revenue, Net Income и средней ставки. |
| 5 | OpenAPI (`api.yaml`) и HTTP Handlers | ✅ Done | `backend-dev` | `make oapi` сгенерирован, подключены `client_handler.go`, `dashboard_handler.go`, `schedule_handler.go`. |
| 6 | Frontend: Удаление роли `student` и старых экранов | ✅ Done | `frontend-dev` | Удален `StudentLessonsPage.tsx`, убрана роль `student` из регистрации и типов. |
| 7 | Frontend: CRM репетитора (Список учеников) | ✅ Done | `frontend-dev` | Создан `TeacherClientsPage.tsx`: карточки клиентов, теги комиссий, модалка пополнения абонемента. |
| 8 | Frontend: Расписание (Календарь) репетитора | ✅ Done | `frontend-dev` | `TeacherSchedulePage.tsx` обновлен на `Client`, прямая постановка в `scheduled`, поддержка нахлестов. |
| 9 | Frontend: Финансовый Дашборд | ✅ Done | `frontend-dev` | `DashboardPage.tsx` расширен виджетом: Gross Revenue, Net Income, Средняя ставка в час. |
| 10| Обновление QA E2E тестов | ✅ Done | `qa-e2e` | 37 тестов успешно проходят в Docker (`make test-e2e` 100% PASS). |

## Результаты
- Архитектура полностью переведена на `clients`.
- E2E тесты зелёные (37 passed).
- Фронтенд собирается без ошибок линтера и TS.
