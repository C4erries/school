import uuid
from datetime import datetime, timedelta, timezone
import pytest
import httpx


def to_rfc3339(dt: datetime) -> str:
    return dt.strftime("%Y-%m-%dT%H:%M:%SZ")


@pytest.mark.schedule
class TestFinanceAndExport:
    """E2E тесты для Спринта 2.2.1: Финансы, платежи, взаиморасчеты со школами и экспорт."""

    def test_create_payment_and_auto_credit_balance(self, client: httpx.Client, teacher_user):
        """Создание платежа автоматически начисляет часы в абонемент клиента."""
        teacher = teacher_user

        # 1. Создаем клиента
        c_res = client.post(
            "/api/v1/clients",
            json={
                "name": f"Плательщик {uuid.uuid4().hex[:4]}",
                "rate_individual": 2000.0,
            },
            headers=teacher["headers"],
        )
        assert c_res.status_code == 201
        client_data = c_res.json()
        client_id = client_data["id"]

        # Исходный баланс 0
        get_c = client.get(f"/api/v1/clients/{client_id}", headers=teacher["headers"])
        assert get_c.status_code == 200
        assert get_c.json()["balances"]["individual_hours"] == 0.0

        # 2. Регистрируем оплату 8 000 руб за 4 часа
        pay_res = client.post(
            "/api/v1/finance/payments",
            json={
                "client_id": client_id,
                "amount": 8000.0,
                "hours": 4.0,
                "format": "individual",
                "payment_method": "transfer",
                "notes": "Оплата за октябрь через СБП",
            },
            headers=teacher["headers"],
        )
        assert pay_res.status_code == 201, f"Failed to create payment: {pay_res.text}"
        payment = pay_res.json()
        assert payment["amount"] == 8000.0
        assert payment["hours"] == 4.0
        assert payment["format"] == "individual"
        assert payment["payment_method"] == "transfer"
        assert payment["notes"] == "Оплата за октябрь через СБП"

        # 3. Баланс клиента автоматически пополнился на 4 часа!
        updated_c = client.get(f"/api/v1/clients/{client_id}", headers=teacher["headers"])
        assert updated_c.status_code == 200
        assert updated_c.json()["balances"]["individual_hours"] == 4.0

        # 4. Проверяем журнал платежей
        list_res = client.get(
            f"/api/v1/finance/payments?client_id={client_id}",
            headers=teacher["headers"],
        )
        assert list_res.status_code == 200
        items = list_res.json()
        assert len(items) >= 1
        assert any(p["id"] == payment["id"] for p in items)

    def test_finance_summary(self, client: httpx.Client, teacher_user):
        """Проверка расчета общей финансовой сводки за месяц."""
        teacher = teacher_user
        current_month = datetime.now(timezone.utc).strftime("%Y-%m")

        # 1. Запрос сводки
        res = client.get(
            f"/api/v1/finance/summary?month={current_month}",
            headers=teacher["headers"],
        )
        assert res.status_code == 200, f"Summary failed: {res.text}"
        data = res.json()
        assert data["month"] == current_month
        assert "total_payments" in data
        assert "total_earned" in data
        assert "total_debts" in data
        assert "total_commissions" in data
        assert "active_subscriptions_count" in data
        assert "debtors_count" in data

    def test_partner_settlements_and_payout(self, client: httpx.Client, teacher_user):
        """Расчет комиссий партнерских школ по тегам и фиксация выплаты."""
        teacher = teacher_user
        current_month = datetime.now(timezone.utc).strftime("%Y-%m")

        # 1. Создаем партнерский тег со ставкой 25% комиссии
        tag_res = client.post(
            "/api/v1/tags",
            json={
                "name": f"Партнер-{uuid.uuid4().hex[:4]}",
                "school_percent": 25,
                "color": "indigo",
            },
            headers=teacher["headers"],
        )
        assert tag_res.status_code == 201
        tag = tag_res.json()
        tag_id = tag["id"]

        # 2. Создаем клиента с этим тегом
        c_res = client.post(
            "/api/v1/clients",
            json={
                "name": f"Ученик Школы {uuid.uuid4().hex[:4]}",
                "rate_individual": 2000.0,
                "tag_ids": [tag_id],
            },
            headers=teacher["headers"],
        )
        assert c_res.status_code == 201
        client_id = c_res.json()["id"]

        # 3. Назначаем и проводим урок
        now = datetime.now(timezone.utc).replace(microsecond=0)
        l_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client_id,
                "title": "Урок от партнера",
                "format": "individual",
                "start_time": to_rfc3339(now),
                "end_time": to_rfc3339(now + timedelta(hours=1)),
            },
            headers=teacher["headers"],
        )
        assert l_res.status_code == 201
        lesson_id = l_res.json()["id"]

        comp_res = client.post(f"/api/v1/lessons/{lesson_id}/complete", headers=teacher["headers"])
        assert comp_res.status_code == 200

        # 4. Проверяем расчет комиссий по партнерам
        settle_res = client.get(
            f"/api/v1/finance/partner-settlements?month={current_month}",
            headers=teacher["headers"],
        )
        assert settle_res.status_code == 200
        settlements = settle_res.json()
        matching = next((s for s in settlements if s["tag_id"] == tag_id), None)
        assert matching is not None, "Tag settlement not found in report"
        assert matching["school_percent"] == 25
        assert matching["lessons_count"] >= 1
        assert matching["is_paid"] is False

        # 5. Фиксируем выплату комиссии партнерской школе
        payout_res = client.post(
            "/api/v1/finance/partner-payouts",
            json={
                "tag_id": tag_id,
                "period_month": current_month,
                "gross_amount": matching["gross_amount"],
                "commission_amount": matching["commission_amount"],
                "notes": "Выплата за расчетный месяц по реквизитам",
            },
            headers=teacher["headers"],
        )
        assert payout_res.status_code == 201
        payout = payout_res.json()
        assert payout["tag_id"] == tag_id
        assert payout["period_month"] == current_month

        # 6. Защита от дубликата: повторная выплата возвращает конфликт 409
        dup_res = client.post(
            "/api/v1/finance/partner-payouts",
            json={
                "tag_id": tag_id,
                "period_month": current_month,
                "gross_amount": matching["gross_amount"],
                "commission_amount": matching["commission_amount"],
            },
            headers=teacher["headers"],
        )
        assert dup_res.status_code == 409

    def test_csv_exports_with_utf8_bom(self, client: httpx.Client, teacher_user):
        """Экспорт клиентов, уроков и платежей в формате CSV с UTF-8 BOM для Excel."""
        teacher = teacher_user

        # 1. Экспорт клиентов
        c_exp = client.get("/api/v1/export/clients", headers=teacher["headers"])
        assert c_exp.status_code == 200
        assert "text/csv" in c_exp.headers.get("content-type", "")
        # Проверка UTF-8 BOM (\xef\xbb\xbf)
        assert c_exp.content.startswith(b"\xef\xbb\xbf"), "Clients CSV must start with UTF-8 BOM"

        # 2. Экспорт уроков
        l_exp = client.get("/api/v1/export/lessons", headers=teacher["headers"])
        assert l_exp.status_code == 200
        assert "text/csv" in l_exp.headers.get("content-type", "")
        assert l_exp.content.startswith(b"\xef\xbb\xbf"), "Lessons CSV must start with UTF-8 BOM"

        # 3. Экспорт платежей
        p_exp = client.get("/api/v1/export/payments", headers=teacher["headers"])
        assert p_exp.status_code == 200
        assert "text/csv" in p_exp.headers.get("content-type", "")
        assert p_exp.content.startswith(b"\xef\xbb\xbf"), "Payments CSV must start with UTF-8 BOM"

    def test_finance_security_rbac(self, client: httpx.Client, student_user):
        """Ученик не имеет доступа к финансовым эндпоинтам (403 Forbidden)."""
        student = student_user

        res1 = client.get("/api/v1/finance/summary", headers=student["headers"])
        assert res1.status_code == 403

        res2 = client.get("/api/v1/finance/payments", headers=student["headers"])
        assert res2.status_code == 403

        res3 = client.get("/api/v1/export/clients", headers=student["headers"])
        assert res3.status_code == 403
