# Спринт 2.2.1: Финансовая бухгалтерия, Журнал оплат, Взаиморасчеты с партнерскими школами и Экспорт данных ✅

> **Период**: 2026-10-06  
> **Статус**: ✅ Завершен (DoD выполнен на 100%, 48/48 E2E тестов зеленые)  
> **Ветка**: `dev`

---

## 🎯 Цели спринта

1. **Полноценный финансовый учет (Payments Ledger)**:
   - Переход от простого учета часов к учету живых денег: таблица платежей `payments` (сумма в ₽, часы, формат, метод оплаты: перевод/СБП/наличные, дата, комментарий).
   - При регистрации платежа реализовано автоматическое пополнение абонемента нужного формата у клиента.
2. **Сводная таблица задолженностей и балансов**:
   - Контроль дебиторской задолженности: фильтр по должникам (`баланс < 0`), расчет суммы долга в рублях (`|часы| × ставка`), предупреждения об окончании абонемента ($\le 1$ ч).
   - Быстрое принятие оплаты прямо из карточки должника в один клик.
3. **Взаиморасчеты с партнерскими школами (% комиссии по тегам)**:
   - Автоматический расчет доли партнерских школ за расчетный месяц: общая выручка по ученикам школы, процент комиссии школы, сумма к перечислению партнеру.
   - Фиксация выплат школам в таблице `partner_payouts` (с защитой от повторной фиксации 409 Conflict).
4. **Экспорт базы данных (CSV UTF-8 BOM)**:
   - Выгрузка реестра учеников, расписания уроков и журнала оплат.
   - Формат CSV с кодировкой UTF-8 BOM (`\xef\xbb\xbf`) для корректного открытия в Excel на любых ОС.
5. **E2E верификация**:
   - Прогон `make test`, `make test-e2e` в Docker (48/48 passed) и `npm run lint && npm run build` (0 ошибок).

---

## 📦 Реализованные компоненты

### 1. Серверная часть (Go & PostgreSQL)
* **Миграция `000008_finance_and_payments.up.sql`**:
  * Таблица `payments` (`id`, `teacher_id`, `client_id`, `amount`, `hours`, `format`, `payment_method`, `paid_at`, `notes`, `created_at`).
  * Таблица `partner_payouts` (`id`, `teacher_id`, `tag_id`, `period_month`, `gross_amount`, `commission_amount`, `paid_at`, `notes`, `created_at`).
* **Доменные модели**:
  * `domain.Payment` (`backend/internal/domain/payment.go`).
  * `domain.PartnerPayout` (`backend/internal/domain/partner_payout.go`).
* **OpenAPI спецификация**:
  * Добавлены маршруты `/finance/summary`, `/finance/payments`, `/finance/partner-settlements`, `/finance/partner-payouts`, `/export/clients`, `/export/lessons`, `/export/payments`.
* **Репозитории**:
  * `PaymentRepository` (`backend/internal/infrastructure/api/adapters/postgres/payment_repository.go`).
  * `PartnerPayoutRepository` (`backend/internal/infrastructure/api/adapters/postgres/partner_payout_repository.go`).
* **Бизнес-логика**:
  * `FinanceService` (`backend/internal/application/finance/service.go`) — начисление оплат, расчет долгов по ставкам, расчет комиссий партнерских школ по завершенным урокам, генерация CSV с BOM.
* **HTTP хэндлеры**:
  * `finance_handler.go` и `export_handler.go` с маппингом ошибок (409 Conflict при дубле выплат) и проверкой прав RBAC.

### 2. Клиентская часть (React, TypeScript & Tailwind CSS)
* **Навигация**: пункт меню «Бухгалтерия» (`/teacher/finance`, иконка `Receipt`) в `AppNavbar.tsx`.
* **Компоненты (`frontend/src/features/finance/components/`)**:
  * `FinanceSummaryCards`: сводные метрики и селектор расчетного месяца.
  * `DebtsAndBalancesTable`: таблица задолженностей и балансов с быстрым приемом оплаты.
  * `PaymentsLedgerTable`: журнал платежей с фильтрацией по дате и ученику.
  * `PartnerSettlementsCard`: взаиморасчеты с партнерскими школами по тегам и фиксация выплат.
  * `AddPaymentModal`: модальное окно регистрации оплаты с авторасчетом суммы.
  * `ExportDataModal`: модальное окно выгрузки CSV файлов.
* **Страница**: `TeacherFinancePage.tsx` в стиле Apple Liquid Glass.

---

## 🧪 Результаты тестирования

* **Unit-тесты Go**: 100% пройдены (`service_test.go`, `finance_handler_test.go`).
* **E2E тесты**: 48/48 тестов пройдены в изолированном тестовом окружении Docker (`tests/api/phase2_schedule/test_finance_and_export.py`).
* **Фронтенд**: `npm run lint && npm run build` — 0 ошибок и ворнингов.
