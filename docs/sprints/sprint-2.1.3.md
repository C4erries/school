# Спринт 2.1.3: Продвинутая CRM, Valkey CSC, Декомпозиция фронтенда и Багфикс онлайн-уроков

> Статус: ✅ **ЗАВЕРШЕН** (2026-10-05)

## Цель спринта
1. **Багфикс онлайн-уроков и 405 PUT**: устранить ошибочный `PUT` в `updateLesson`, обеспечить сброс кабинета в `nil` при переключении на онлайн и чтение `online_link` из `location_or_url`.
2. **CRM и ученики**:
   - Настраиваемый дефолтный прайс репетитора (`users.default_rate_*`) с автоподстановкой при создании клиента.
   - Продвинутый реестр: поиск по имени/телефону, табы статуса («Активные», «Архив», «Все»), фильтры баланса и тегов, архивация/восстановление учеников.
   - Ручная корректировка баланса абонементов репетитором с обязательной причиной в аудит-лог (`client_balance_adjustments`).
3. **Архитектура бэкенда и производительность**:
   - Применение миграции `000006_sprint_2_1_3_features.up.sql`.
   - Client-Side Caching (CSC) через Valkey для справочников (`/classrooms`, `/tags`) с RESP3-инвалидацией при мутациях.
4. **Декомпозиция монолитного фронтенда (ликвидация файлов 1000+ строк)**:
   - `TeacherSchedulePage.tsx` (сокращен с 1443 до 218 строк, модули вынесены в `features/schedule/`).
   - `TeacherClientsPage.tsx` (сокращен с 945 до 193 строк, модули вынесены в `features/clients/`).
5. **Оптимизация тестов**:
   - Удаление мертвого артефакта `test_teacher_students.py`.
   - Написание E2E тестов для дефолтных ставок, архивации, фильтрации и ручной корректировки баланса.

---

## Выполненные задачи

| # | Задача | Статус | Ответственный | Заметки |
|---|--------|--------|---------------|---------|
| 0 | Фикс редактирования онлайн-занятия и 405 PUT | ✅ Done | `frontend-dev` / `backend-dev` | Устранен 405 PUT (прямой PATCH). Очистка кабинета в `nil` при переводе в онлайн. Корректное чтение `online_link`. |
| 1 | Дефолтный прайс репетитора (Бэкенд + UI) | ✅ Done | `backend-dev` + `frontend-dev` | Миграция 000006, `GET`/`PUT` `/users/me/rates`, модалка `DefaultRatesModal` в шапке CRM, автозаполнение при создании клиента. |
| 2 | Продвинутый реестр и архивация учеников | ✅ Done | `backend-dev` + `frontend-dev` | Поле `is_archived`, живой поиск по имени/телефону, табы статуса, фильтры баланса и тегов, кнопки «В архив» / «Восстановить». |
| 3 | Ручная корректировка баланса с аудитом | ✅ Done | `backend-dev` + `frontend-dev` | Таблица `client_balance_adjustments`, `POST /clients/{id}/adjust-balance`, модалка `AdjustBalanceModal` с обязательной причиной. |
| 4 | Client-Side Caching (CSC) через Valkey | ✅ Done | `backend-dev` | Кэширование справочников (`/classrooms`, `/tags`) через `valkey-go` `DoCache` с RESP3-инвалидацией при мутациях. |
| 5 | Декомпозиция `TeacherSchedulePage.tsx` | ✅ Done | `frontend-dev` | Файл сокращен с 1443 до 218 строк. Модули: `ScheduleToolbar`, `ScheduleWeekView`, `ScheduleDayView`, `ScheduleListView`, `CreateLessonModal`, `EditLessonModal`, `useSchedulePositioning`. |
| 6 | Декомпозиция `TeacherClientsPage.tsx` | ✅ Done | `frontend-dev` | Файл сокращен с 945 до 193 строк. Модули: `ClientCard`, `ClientFilters`, `CreateClientModal`, `EditClientModal`, `AddSubscriptionModal`, `AdjustBalanceModal`, `DefaultRatesModal`, `useClientFilters`. |
| 7 | Оптимизация тестового сьюта и новые E2E тесты | ✅ Done | `qa-e2e` | Удален рудимент `test_teacher_students.py`. Добавлены тесты дефолтных ставок, архивации, корректировки баланса. 43 passed (100%). |

---

## Результаты
- 43 E2E автотеста в Docker: 100% PASS.
- 0 ошибок ESLint и TypeScript при сборке фронтенда.
- Все файлы фронтенда приведены к размеру < 300 строк.
