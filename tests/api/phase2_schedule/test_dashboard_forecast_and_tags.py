import uuid
from datetime import datetime, timedelta, timezone
import pytest
import httpx


def to_rfc3339(dt: datetime) -> str:
    """Форматирует datetime в ISO/RFC3339 строку с суффиксом Z."""
    return dt.strftime("%Y-%m-%dT%H:%M:%SZ")


@pytest.mark.schedule
class TestDashboardForecastAndTags:
    """Комплексные E2E тесты для Спринта 2.2.3: Дашборд, Прогноз аналитики, Статистика тегов и фильтрация выплат."""

    def test_dashboard_summary_workflow(self, client: httpx.Client, teacher_user):
        """Проверка работы GET /api/v1/dashboard/summary:
        - today_lessons содержит только сегодняшние уроки с корректными полями
        - расчет month_earned (completed) и month_forecast (будущие scheduled)
        - расчет total_debts для клиентов с отрицательным балансом
        - active_clients_count и weekly_hours
        """
        teacher = teacher_user
        headers = teacher["headers"]

        # 1. Создаем клиента 1 (без тегов, ставка индивидуальная 2000)
        c1_res = client.post(
            "/api/v1/clients",
            json={
                "name": f"Дашборд Ученик 1 {uuid.uuid4().hex[:4]}",
                "rate_individual": 2000.0,
            },
            headers=headers,
        )
        assert c1_res.status_code == 201
        client1 = c1_res.json()

        # Создаем клиента 2 (для долгов и второго урока)
        c2_res = client.post(
            "/api/v1/clients",
            json={
                "name": f"Дашборд Должник 2 {uuid.uuid4().hex[:4]}",
                "rate_individual": 1500.0,
                "rate_pair": 1000.0,
            },
            headers=headers,
        )
        assert c2_res.status_code == 201
        client2 = c2_res.json()

        # Устанавливаем клиенту 2 отрицательный баланс (-2 часа individual, -1 час pair)
        # 2 * 1500 + 1 * 1000 = 3000 + 1000 = 4000 руб долга
        adj1 = client.post(
            f"/api/v1/clients/{client2['id']}/adjust-balance",
            json={"format": "individual", "delta_hours": -2.0, "reason": "Долг за индивидуальные"},
            headers=headers,
        )
        assert adj1.status_code == 200

        adj2 = client.post(
            f"/api/v1/clients/{client2['id']}/adjust-balance",
            json={"format": "pair", "delta_hours": -1.0, "reason": "Долг за пару"},
            headers=headers,
        )
        assert adj2.status_code == 200

        now = datetime.now(timezone.utc).replace(microsecond=0)

        # 2. Создаем уроки на СЕГОДНЯ:
        # Урок А: сегодня в 10:00 (1 час, individual, онлайн) -> завершим его (completed)
        lesson_today_completed_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client1["id"],
                "title": "Онлайн математика",
                "format": "individual",
                "start_time": to_rfc3339(now.replace(hour=10, minute=0, second=0)),
                "end_time": to_rfc3339(now.replace(hour=11, minute=0, second=0)),
                "location_or_url": "https://telemost.yandex.ru/j/test-dashboard",
            },
            headers=headers,
        )
        assert lesson_today_completed_res.status_code == 201
        lesson_today_completed = lesson_today_completed_res.json()

        comp_res = client.post(
            f"/api/v1/lessons/{lesson_today_completed['id']}/complete",
            headers=headers,
        )
        assert comp_res.status_code == 200

        # Урок Б: сегодня в 23:30 (1 час, individual, запланирован на вечер сегодня или будущее время сегодня)
        # Чтобы он точно считался как scheduled на сегодня:
        # Ставим start_time через 1-2 часа от текущего момента, если это сегодня,
        # либо в пределах сегодняшних суток (но > now)
        today_future_start = now + timedelta(hours=1)
        # Если today_future_start перевалил за полночь, скорректируем на 23:59
        if today_future_start.day != now.day:
            today_future_start = now.replace(hour=23, minute=0, second=0)

        lesson_today_scheduled_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client2["id"],
                "title": "Вечерний урок",
                "format": "individual",
                "start_time": to_rfc3339(today_future_start),
                "end_time": to_rfc3339(today_future_start + timedelta(hours=1)),
                "location_or_url": "https://zoom.us/j/dashboard-test",
            },
            headers=headers,
        )
        assert lesson_today_scheduled_res.status_code == 201
        lesson_today_scheduled = lesson_today_scheduled_res.json()

        # 3. Создаем урок на ЗАВТРА (scheduled) - не должен попасть в today_lessons
        tomorrow = now + timedelta(days=1)
        lesson_tomorrow_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client1["id"],
                "title": "Завтрашний урок",
                "format": "individual",
                "start_time": to_rfc3339(tomorrow.replace(hour=12, minute=0, second=0)),
                "end_time": to_rfc3339(tomorrow.replace(hour=13, minute=0, second=0)),
            },
            headers=headers,
        )
        assert lesson_tomorrow_res.status_code == 201
        lesson_tomorrow = lesson_tomorrow_res.json()

        # 4. Запрашиваем GET /api/v1/dashboard/summary
        sum_res = client.get("/api/v1/dashboard/summary", headers=headers)
        assert sum_res.status_code == 200, f"Dashboard summary failed: {sum_res.text}"
        data = sum_res.json()

        assert "today_lessons" in data
        assert "financial_snapshot" in data

        today_lessons = data["today_lessons"]
        today_ids = [l["id"] for l in today_lessons]

        # Должны быть сегодняшние уроки
        assert lesson_today_completed["id"] in today_ids
        assert lesson_today_scheduled["id"] in today_ids
        # Урок на завтра НЕ должен быть в today_lessons
        assert lesson_tomorrow["id"] not in today_ids

        # Проверяем поля элементов today_lessons
        matched_comp = next(l for l in today_lessons if l["id"] == lesson_today_completed["id"])
        assert matched_comp["client_name"] == client1["name"]
        assert matched_comp["status"] == "completed"
        assert matched_comp["format"] == "individual"
        assert matched_comp["location_type"] == "online"
        assert matched_comp["online_link"] == "https://telemost.yandex.ru/j/test-dashboard"
        assert matched_comp["start_at"]
        assert matched_comp["end_at"]

        matched_sched = next(l for l in today_lessons if l["id"] == lesson_today_scheduled["id"])
        assert matched_sched["client_name"] == client2["name"]
        assert matched_sched["status"] == "scheduled"
        assert matched_sched["online_link"] == "https://zoom.us/j/dashboard-test"

        # Проверяем финансовый снимок
        fin = data["financial_snapshot"]
        # month_earned: как минимум наш completed урок (1 час * 2000 = 2000)
        assert fin["month_earned"] >= 2000.0
        # total_debts: как минимум долг клиента 2 (4000 руб)
        assert fin["total_debts"] >= 4000.0
        # active_clients_count: как минимум 2 активных клиента
        assert fin["active_clients_count"] >= 2
        # weekly_hours: как минимум отработанные и запланированные часы текущей недели
        assert fin["weekly_hours"] >= 1.0

    def test_analytics_forecast(self, client: httpx.Client, teacher_user):
        """Проверка работы GET /api/v1/analytics/forecast:
        - Запланированные будущие уроки разных форматов (individual, pair, group)
        - Кастомные ставки клиентов
        - Вычет партнерских комиссий
        - Структура массива by_format
        """
        teacher = teacher_user
        headers = teacher["headers"]

        # Создаем партнерский тег с комиссией 20%
        tag_res = client.post(
            "/api/v1/tags",
            json={
                "name": f"Партнер-Forecast-{uuid.uuid4().hex[:4]}",
                "school_percent": 20,
                "color": "#10B981",
            },
            headers=headers,
        )
        assert tag_res.status_code == 201
        tag_id = tag_res.json()["id"]

        # Клиент 1: партнерский (20% комиссия)
        # individual: 3000, pair: 2000, group: 1500
        c1_res = client.post(
            "/api/v1/clients",
            json={
                "name": f"Forecast Клиент Партнерский {uuid.uuid4().hex[:4]}",
                "rate_individual": 3000.0,
                "rate_pair": 2000.0,
                "rate_group": 1500.0,
                "tag_ids": [tag_id],
            },
            headers=headers,
        )
        assert c1_res.status_code == 201
        client1 = c1_res.json()

        # Клиент 2: прямой клиент (0% комиссия)
        # individual: 2500, pair: 1800, group: 1200
        c2_res = client.post(
            "/api/v1/clients",
            json={
                "name": f"Forecast Клиент Прямой {uuid.uuid4().hex[:4]}",
                "rate_individual": 2500.0,
                "rate_pair": 1800.0,
                "rate_group": 1200.0,
            },
            headers=headers,
        )
        assert c2_res.status_code == 201
        client2 = c2_res.json()

        # Период планирования: следующая неделя
        now = datetime.now(timezone.utc).replace(microsecond=0)
        start_period = now + timedelta(days=20)
        end_period = start_period + timedelta(days=7)

        # Урок 1: individual, 2 часа, клиент 1 (партнерский)
        # Gross = 2 * 3000 = 6000. Comm = 6000 * 0.2 = 1200. Net = 4800.
        l1_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client1["id"],
                "title": "Урок Individual",
                "format": "individual",
                "start_time": to_rfc3339(start_period + timedelta(hours=1)),
                "end_time": to_rfc3339(start_period + timedelta(hours=3)),
            },
            headers=headers,
        )
        assert l1_res.status_code == 201

        # Урок 2: pair, 1.5 часа (90 мин), клиент 2 (прямой)
        # Gross = 1.5 * 1800 = 2700. Comm = 0. Net = 2700.
        l2_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client2["id"],
                "title": "Урок Pair",
                "format": "pair",
                "start_time": to_rfc3339(start_period + timedelta(days=1, hours=2)),
                "end_time": to_rfc3339(start_period + timedelta(days=1, hours=3, minutes=30)),
            },
            headers=headers,
        )
        assert l2_res.status_code == 201

        # Урок 3: group, 1 час, клиент 1 (партнерский)
        # Gross = 1 * 1500 = 1500. Comm = 1500 * 0.2 = 300. Net = 1200.
        l3_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client1["id"],
                "title": "Урок Group",
                "format": "group",
                "start_time": to_rfc3339(start_period + timedelta(days=2, hours=4)),
                "end_time": to_rfc3339(start_period + timedelta(days=2, hours=5)),
            },
            headers=headers,
        )
        assert l3_res.status_code == 201

        # Итого ожидаем за [start_period, end_period]:
        # Scheduled lessons: 3
        # Scheduled hours: 2 + 1.5 + 1 = 4.5 часа
        # Gross: 6000 + 2700 + 1500 = 10200 руб.
        # Partner commission expected: 1200 + 0 + 300 = 1500 руб.
        # Net potential income: 10200 - 1500 = 8700 руб.
        from_str = to_rfc3339(start_period)
        to_str = to_rfc3339(end_period)

        forecast_res = client.get(
            f"/api/v1/analytics/forecast?from={from_str}&to={to_str}",
            headers=headers,
        )
        assert forecast_res.status_code == 200, f"Forecast error: {forecast_res.text}"
        forecast = forecast_res.json()

        assert forecast["scheduled_lessons"] == 3
        assert abs(forecast["scheduled_hours"] - 4.5) < 1e-2
        assert abs(forecast["gross_potential_revenue"] - 10200.0) < 1e-2
        assert abs(forecast["partner_commission_expected"] - 1500.0) < 1e-2
        assert abs(forecast["net_potential_income"] - 8700.0) < 1e-2

        # Проверяем массив by_format
        by_format = {f["format"]: f for f in forecast["by_format"]}
        assert "individual" in by_format
        assert abs(by_format["individual"]["hours"] - 2.0) < 1e-2
        assert abs(by_format["individual"]["revenue"] - 6000.0) < 1e-2

        assert "pair" in by_format
        assert abs(by_format["pair"]["hours"] - 1.5) < 1e-2
        assert abs(by_format["pair"]["revenue"] - 2700.0) < 1e-2

        assert "group" in by_format
        assert abs(by_format["group"]["hours"] - 1.0) < 1e-2
        assert abs(by_format["group"]["revenue"] - 1500.0) < 1e-2

    def test_analytics_tags_and_partner_settlements_filtering(self, client: httpx.Client, teacher_user):
        """Проверка работы GET /api/v1/analytics/tags и GET /api/v1/finance/partner-settlements:
        - Создаем тег без комиссии (school_percent: 0) и партнерский тег (school_percent: 20)
        - Привязываем учеников и проводим уроки
        - В /api/v1/analytics/tags попадают ОБА тега со статистикой учеников, отработанных часов и выручки
        - В /api/v1/finance/partner-settlements тег с 0% исключен из взаиморасчетов
        """
        teacher = teacher_user
        headers = teacher["headers"]
        now = datetime.now(timezone.utc).replace(microsecond=0)
        month_str = now.strftime("%Y-%m")

        # 1. Создаем тег без комиссии (0%)
        zero_tag_res = client.post(
            "/api/v1/tags",
            json={
                "name": f"Инфо-Тег-{uuid.uuid4().hex[:4]}",
                "school_percent": 0,
                "color": "#3B82F6",
            },
            headers=headers,
        )
        assert zero_tag_res.status_code == 201
        zero_tag = zero_tag_res.json()
        zero_tag_id = zero_tag["id"]

        # 2. Создаем партнерский тег с комиссией (20%)
        partner_tag_res = client.post(
            "/api/v1/tags",
            json={
                "name": f"Партнер-Тег-{uuid.uuid4().hex[:4]}",
                "school_percent": 20,
                "color": "#EC4899",
            },
            headers=headers,
        )
        assert partner_tag_res.status_code == 201
        partner_tag = partner_tag_res.json()
        partner_tag_id = partner_tag["id"]

        # 3. Создаем клиента для тега 0%
        c_zero_res = client.post(
            "/api/v1/clients",
            json={
                "name": f"Ученик Инфо {uuid.uuid4().hex[:4]}",
                "rate_individual": 2000.0,
                "tag_ids": [zero_tag_id],
            },
            headers=headers,
        )
        assert c_zero_res.status_code == 201
        client_zero = c_zero_res.json()

        # 4. Создаем клиента для партнерского тега 20%
        c_partner_res = client.post(
            "/api/v1/clients",
            json={
                "name": f"Ученик Партнера {uuid.uuid4().hex[:4]}",
                "rate_individual": 3000.0,
                "tag_ids": [partner_tag_id],
            },
            headers=headers,
        )
        assert c_partner_res.status_code == 201
        client_partner = c_partner_res.json()

        # 5. Назначаем и завершаем уроки для обоих клиентов
        # Урок для первого клиента: 2 часа * 2000 = 4000 руб (Gross 4000, Net 4000)
        l_zero_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client_zero["id"],
                "title": "Урок Инфо",
                "format": "individual",
                "start_time": to_rfc3339(now.replace(hour=8, minute=0, second=0)),
                "end_time": to_rfc3339(now.replace(hour=10, minute=0, second=0)),
            },
            headers=headers,
        )
        assert l_zero_res.status_code == 201
        l_zero_id = l_zero_res.json()["id"]
        assert client.post(f"/api/v1/lessons/{l_zero_id}/complete", headers=headers).status_code == 200

        # Урок для второго клиента: 1 час * 3000 = 3000 руб (Gross 3000, Net 2400)
        l_partner_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client_partner["id"],
                "title": "Урок Партнера",
                "format": "individual",
                "start_time": to_rfc3339(now.replace(hour=11, minute=0, second=0)),
                "end_time": to_rfc3339(now.replace(hour=12, minute=0, second=0)),
            },
            headers=headers,
        )
        assert l_partner_res.status_code == 201
        l_partner_id = l_partner_res.json()["id"]
        assert client.post(f"/api/v1/lessons/{l_partner_id}/complete", headers=headers).status_code == 200

        # 6. Проверяем /api/v1/analytics/tags: ОБА тега должны присутствовать
        tags_stat_res = client.get("/api/v1/analytics/tags", headers=headers)
        assert tags_stat_res.status_code == 200, f"Tags stats failed: {tags_stat_res.text}"
        tag_stats = tags_stat_res.json()

        zero_stat = next((t for t in tag_stats if t["tag_id"] == zero_tag_id), None)
        assert zero_stat is not None, "Тег с 0% должен отображаться в аналитике тегов"
        assert zero_stat["students_count"] >= 1
        assert zero_stat["completed_hours"] >= 2.0
        assert zero_stat["gross_revenue"] >= 4000.0
        assert zero_stat["net_income"] >= 4000.0

        partner_stat = next((t for t in tag_stats if t["tag_id"] == partner_tag_id), None)
        assert partner_stat is not None, "Партнерский тег должен отображаться в аналитике тегов"
        assert partner_stat["students_count"] >= 1
        assert partner_stat["completed_hours"] >= 1.0
        assert partner_stat["gross_revenue"] >= 3000.0
        assert partner_stat["net_income"] >= 2400.0

        # 7. Проверяем /api/v1/finance/partner-settlements:
        # Тег с 0% комиссии ДОЛЖЕН быть исключен, а партнерский тег 20% присутствует!
        settle_res = client.get(
            f"/api/v1/finance/partner-settlements?month={month_str}",
            headers=headers,
        )
        assert settle_res.status_code == 200
        settlements = settle_res.json()

        assert not any(s["tag_id"] == zero_tag_id for s in settlements), (
            "Тег с 0% комиссии не должен появляться во взаиморасчетах с партнерами!"
        )
        partner_settle = next((s for s in settlements if s["tag_id"] == partner_tag_id), None)
        assert partner_settle is not None, "Партнерский тег с 20% комиссии должен быть в settlements"
        assert partner_settle["school_percent"] == 20
        assert partner_settle["gross_amount"] >= 3000.0
        assert partner_settle["commission_amount"] >= 600.0

    def test_rbac_and_data_isolation(self, client: httpx.Client, teacher_user, registered_user):
        """Проверка изоляции данных (RBAC):
        - Преподаватель А не видит дашборд и аналитику преподавателя Б.
        - Студент не имеет доступа к дашборду и прогнозу аналитики (403 Forbidden).
        """
        teacher_a = teacher_user
        headers_a = teacher_a["headers"]

        # Создаем преподавателя Б
        _, teacher_b_reg = registered_user(role="teacher")
        headers_b = {"Authorization": f"Bearer {teacher_b_reg['tokens']['access_token']}"}

        # Создаем студента
        _, student_reg = registered_user(role="student")
        headers_student = {"Authorization": f"Bearer {student_reg['tokens']['access_token']}"}

        # 1. Преподаватель А создает клиента и урок
        c_res = client.post(
            "/api/v1/clients",
            json={"name": f"Ученик А {uuid.uuid4().hex[:4]}", "rate_individual": 5000.0},
            headers=headers_a,
        )
        assert c_res.status_code == 201
        client_a_id = c_res.json()["id"]

        now = datetime.now(timezone.utc).replace(microsecond=0)
        l_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client_a_id,
                "title": "Секретный урок Преподавателя А",
                "format": "individual",
                "start_time": to_rfc3339(now),
                "end_time": to_rfc3339(now + timedelta(hours=1)),
            },
            headers=headers_a,
        )
        assert l_res.status_code == 201
        lesson_a_id = l_res.json()["id"]

        # 2. Преподаватель Б запрашивает дашборд
        b_summary_res = client.get("/api/v1/dashboard/summary", headers=headers_b)
        assert b_summary_res.status_code == 200
        b_summary = b_summary_res.json()
        b_today_ids = [l["id"] for l in b_summary["today_lessons"]]
        assert lesson_a_id not in b_today_ids
        assert b_summary["financial_snapshot"]["month_earned"] == 0.0

        # Преподаватель Б запрашивает прогноз
        from_str = to_rfc3339(now)
        to_str = to_rfc3339(now + timedelta(days=5))
        b_forecast_res = client.get(
            f"/api/v1/analytics/forecast?from={from_str}&to={to_str}",
            headers=headers_b,
        )
        assert b_forecast_res.status_code == 200
        b_forecast = b_forecast_res.json()
        assert b_forecast["scheduled_lessons"] == 0
        assert b_forecast["gross_potential_revenue"] == 0.0

        # 3. Студент пытается получить дашборд и прогноз аналитики -> 403 Forbidden
        st_summary_res = client.get("/api/v1/dashboard/summary", headers=headers_student)
        assert st_summary_res.status_code == 403

        st_forecast_res = client.get(
            f"/api/v1/analytics/forecast?from={from_str}&to={to_str}",
            headers=headers_student,
        )
        assert st_forecast_res.status_code == 403

        st_tags_res = client.get("/api/v1/analytics/tags", headers=headers_student)
        assert st_tags_res.status_code == 403

