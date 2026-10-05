import uuid
from datetime import datetime, timedelta, timezone
import pytest
import httpx


def to_rfc3339(dt: datetime) -> str:
    """Форматирует datetime в ISO/RFC3339 строку с суффиксом Z."""
    return dt.strftime("%Y-%m-%dT%H:%M:%SZ")


@pytest.mark.schedule
class TestDashboard:
    """Интеграционные тесты финансового дашборда репетитора."""

    def test_dashboard_metrics_empty(self, client: httpx.Client, teacher_user):
        """Дашборд для репетитора без уроков возвращает нулевые метрики."""
        now = datetime.now(timezone.utc).replace(microsecond=0)
        from_time = to_rfc3339(now - timedelta(days=7))
        to_time = to_rfc3339(now + timedelta(days=7))

        res = client.get(
            f"/api/v1/dashboard/metrics?from={from_time}&to={to_time}",
            headers=teacher_user["headers"],
        )
        assert res.status_code == 200, f"Failed to get dashboard metrics: {res.text}"
        data = res.json()

        assert data["gross_potential_revenue"] == 0.0
        assert data["net_income"] == 0.0
        assert data["average_rate"] == 0.0

    def test_dashboard_metrics_calculation(self, client: httpx.Client, teacher_user):
        """Проверка корректности расчета Gross Potential Revenue, Net Income и Average Rate."""
        teacher = teacher_user

        # 1. Создаем двух клиентов с разными ставками и процентами комиссии
        # Клиент 1: ставка 2000 руб/час, комиссия школы 20%
        c1_res = client.post(
            "/api/v1/clients",
            json={
                "name": f"Клиент-1 {uuid.uuid4().hex[:4]}",
                "base_rate": 2000.0,
                "school_percent_tag": 20,
            },
            headers=teacher["headers"],
        )
        assert c1_res.status_code == 201
        client1 = c1_res.json()

        # Клиент 2: ставка 1000 руб/час, комиссия школы 0%
        c2_res = client.post(
            "/api/v1/clients",
            json={
                "name": f"Клиент-2 {uuid.uuid4().hex[:4]}",
                "base_rate": 1000.0,
                "school_percent_tag": 0,
            },
            headers=teacher["headers"],
        )
        assert c2_res.status_code == 201
        client2 = c2_res.json()

        # 2. Назначаем уроки в тестовый день
        test_day = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=10)
        from_time = to_rfc3339(test_day.replace(hour=0, minute=0, second=0))
        to_time = to_rfc3339(test_day.replace(hour=23, minute=59, second=59))

        # Урок 1: Клиент 1, 1 час (10:00 - 11:00)
        # Gross = 2000 * 1 = 2000
        # Net = 2000 - (2000 * 0.20) = 1600
        l1_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client1["id"],
                "start_time": to_rfc3339(test_day.replace(hour=10, minute=0, second=0)),
                "end_time": to_rfc3339(test_day.replace(hour=11, minute=0, second=0)),
                "format": "online",
            },
            headers=teacher["headers"],
        )
        assert l1_res.status_code == 201

        # Урок 2: Клиент 2, 2 часа (12:00 - 14:00)
        # Gross = 1000 * 2 = 2000
        # Net = 2000 - (2000 * 0.0) = 2000
        l2_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client2["id"],
                "start_time": to_rfc3339(test_day.replace(hour=12, minute=0, second=0)),
                "end_time": to_rfc3339(test_day.replace(hour=14, minute=0, second=0)),
                "format": "online",
            },
            headers=teacher["headers"],
        )
        assert l2_res.status_code == 201

        # Урок 3: Отмененный урок (не должен учитываться)
        l3_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client1["id"],
                "start_time": to_rfc3339(test_day.replace(hour=15, minute=0, second=0)),
                "end_time": to_rfc3339(test_day.replace(hour=16, minute=0, second=0)),
                "format": "online",
            },
            headers=teacher["headers"],
        )
        assert l3_res.status_code == 201
        l3_id = l3_res.json()["id"]

        cancel_res = client.post(
            f"/api/v1/lessons/{l3_id}/cancel",
            json={"reason": "Отменен для теста"},
            headers=teacher["headers"],
        )
        assert cancel_res.status_code == 200

        # 3. Запрашиваем метрики дашборда
        res = client.get(
            f"/api/v1/dashboard/metrics?from={from_time}&to={to_time}",
            headers=teacher["headers"],
        )
        assert res.status_code == 200, f"Dashboard metrics error: {res.text}"
        metrics = res.json()

        # Ожидаемые значения:
        # Gross: 2000 + 2000 = 4000
        # Net: 1600 + 2000 = 3600
        # Average rate: (2000 + 1000) / 2 = 1500
        assert abs(metrics["gross_potential_revenue"] - 4000.0) < 1e-4
        assert abs(metrics["net_income"] - 3600.0) < 1e-4
        assert abs(metrics["average_rate"] - 1500.0) < 1e-4

    def test_dashboard_teacher_isolation(
        self, client: httpx.Client, teacher_user, registered_user
    ):
        """Изоляция дашборда: репетитор видит только свои финансовые показатели."""
        _, other_teacher_reg = registered_user(role="teacher")
        other_headers = {"Authorization": f"Bearer {other_teacher_reg['tokens']['access_token']}"}

        now = datetime.now(timezone.utc).replace(microsecond=0)
        from_time = to_rfc3339(now)
        to_time = to_rfc3339(now + timedelta(days=30))

        # У второго преподавателя нет уроков в этом периоде -> все метрики 0
        res = client.get(
            f"/api/v1/dashboard/metrics?from={from_time}&to={to_time}",
            headers=other_headers,
        )
        assert res.status_code == 200
        data = res.json()
        assert data["gross_potential_revenue"] == 0.0
        assert data["net_income"] == 0.0
        assert data["average_rate"] == 0.0

    def test_dashboard_unauthorized(self, client: httpx.Client):
        """Запрос метрик дашборда без токена возвращает 401 Unauthorized."""
        now = datetime.now(timezone.utc).replace(microsecond=0)
        from_time = to_rfc3339(now)
        to_time = to_rfc3339(now + timedelta(days=1))

        res = client.get(f"/api/v1/dashboard/metrics?from={from_time}&to={to_time}")
        assert res.status_code == 401
