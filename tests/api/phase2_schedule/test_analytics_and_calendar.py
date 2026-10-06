import uuid
from datetime import datetime, timedelta, timezone
import pytest
import httpx


def to_rfc3339(dt: datetime) -> str:
    return dt.strftime("%Y-%m-%dT%H:%M:%SZ")


@pytest.mark.schedule
class TestAnalyticsAndCalendar:
    """E2E тесты для Спринта 2.2.2: Аналитика, статистика и календарная интеграция iCal/Webcal."""

    def test_calendar_settings_and_rotation(self, client: httpx.Client, teacher_user):
        """Проверка получения настроек календаря, ссылок feed/webcal и ротации токена."""
        teacher = teacher_user

        # 1. Получаем настройки календаря
        res = client.get("/api/v1/integrations/calendar/settings", headers=teacher["headers"])
        assert res.status_code == 200, f"Failed to get calendar settings: {res.text}"
        settings = res.json()
        initial_token = settings["calendar_token"]
        assert initial_token
        assert settings["feed_url"].endswith(f"token={initial_token}")
        assert settings["webcal_url"].startswith("webcal://")

        # 2. Проверяем доступность фида по первоначальному токену
        feed_res = client.get(f"/api/v1/integrations/calendar/feed.ics?token={initial_token}")
        assert feed_res.status_code == 200
        assert "text/calendar" in feed_res.headers.get("content-type", "")
        assert "BEGIN:VCALENDAR" in feed_res.text
        assert "END:VCALENDAR" in feed_res.text

        # 3. Ротируем токен календаря
        rot_res = client.post("/api/v1/integrations/calendar/rotate-token", headers=teacher["headers"])
        assert rot_res.status_code == 200
        new_settings = rot_res.json()
        new_token = new_settings["calendar_token"]
        assert new_token != initial_token
        assert new_settings["feed_url"].endswith(f"token={new_token}")

        # 4. Старый токен больше недействителен (401 Unauthorized)
        old_feed_res = client.get(f"/api/v1/integrations/calendar/feed.ics?token={initial_token}")
        assert old_feed_res.status_code == 401

        # 5. Новый токен работает
        new_feed_res = client.get(f"/api/v1/integrations/calendar/feed.ics?token={new_token}")
        assert new_feed_res.status_code == 200

    def test_calendar_feed_rfc5545_and_export(self, client: httpx.Client, teacher_user):
        """Проверка генерации событий VEVENT в стандарте RFC 5545 и экспорта .ics."""
        teacher = teacher_user

        # 1. Создаем ученика и урок
        c_res = client.post(
            "/api/v1/clients",
            json={"name": f"Календарный Ученик {uuid.uuid4().hex[:4]}", "rate_individual": 1800.0},
            headers=teacher["headers"],
        )
        assert c_res.status_code == 201
        client_id = c_res.json()["id"]

        now = datetime.now(timezone.utc).replace(microsecond=0)
        l_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client_id,
                "title": "Геометрия: Теорема Пифагора",
                "format": "individual",
                "start_time": to_rfc3339(now + timedelta(days=1)),
                "end_time": to_rfc3339(now + timedelta(days=1, hours=1)),
                "location_or_url": "https://telemost.yandex.ru/j/123456",
                "notes": "Повторить формулы прямоугольного треугольника",
            },
            headers=teacher["headers"],
        )
        assert l_res.status_code == 201
        lesson_id = l_res.json()["id"]

        # 2. Получаем настройки фида
        settings_res = client.get("/api/v1/integrations/calendar/settings", headers=teacher["headers"])
        token = settings_res.json()["calendar_token"]

        # 3. Запрашиваем iCal фид без Bearer заголовка (публичный url с токеном)
        feed = client.get(f"/api/v1/integrations/calendar/feed.ics?token={token}")
        assert feed.status_code == 200
        content = feed.text
        assert "BEGIN:VCALENDAR" in content
        assert "VERSION:2.0" in content
        assert "PRODID:-//School Tutor Assistant//RU" in content
        assert "BEGIN:VEVENT" in content
        assert f"UID:lesson-{lesson_id}@school" in content
        assert "SUMMARY:Геометрия: Теорема Пифагора" in content
        assert "LOCATION:https://telemost.yandex.ru/j/123456" in content
        assert "STATUS:CONFIRMED" in content
        assert "END:VEVENT" in content
        assert "END:VCALENDAR" in content

        # 4. Проверяем разовый экспорт .ics
        export_res = client.get("/api/v1/integrations/calendar/export", headers=teacher["headers"])
        assert export_res.status_code == 200
        assert "text/calendar" in export_res.headers.get("content-type", "")
        assert f"UID:lesson-{lesson_id}@school" in export_res.text

    def test_analytics_kpi_dynamics_and_formats(self, client: httpx.Client, teacher_user):
        """Проверка расчетов сводки аналитики, динамики, форматов и рейтинга учеников."""
        teacher = teacher_user

        # 1. Создаем тег школы со ставкой комиссии 20%
        tag_res = client.post(
            "/api/v1/tags",
            json={"name": f"Школа-Партнер {uuid.uuid4().hex[:4]}", "school_percent": 20},
            headers=teacher["headers"],
        )
        assert tag_res.status_code == 201
        tag_id = tag_res.json()["id"]

        # 2. Создаем клиента, привязанного к школе (индивидуальная ставка 2000 руб/час)
        c_res = client.post(
            "/api/v1/clients",
            json={
                "name": f"Аналитический Клиент {uuid.uuid4().hex[:4]}",
                "rate_individual": 2000.0,
                "tag_ids": [tag_id],
            },
            headers=teacher["headers"],
        )
        assert c_res.status_code == 201
        client_id = c_res.json()["id"]

        # 3. Создаем и завершаем урок (1.5 часа)
        now = datetime.now(timezone.utc).replace(microsecond=0)
        l1_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client_id,
                "title": "Урок №1 по физике",
                "format": "individual",
                "start_time": to_rfc3339(now - timedelta(days=2)),
                "end_time": to_rfc3339(now - timedelta(days=2) + timedelta(hours=1, minutes=30)),
            },
            headers=teacher["headers"],
        )
        assert l1_res.status_code == 201
        lesson1_id = l1_res.json()["id"]

        comp_res = client.post(f"/api/v1/lessons/{lesson1_id}/complete", headers=teacher["headers"])
        assert comp_res.status_code == 200

        # 4. Создаем и отменяем второй урок с причиной (start: now - 1 day, end: now - 1 day + 1 hour)
        l2_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client_id,
                "title": "Урок №2 по физике",
                "format": "individual",
                "start_time": to_rfc3339(now - timedelta(days=1)),
                "end_time": to_rfc3339(now - timedelta(days=1) + timedelta(hours=1)),
            },
            headers=teacher["headers"],
        )
        assert l2_res.status_code == 201
        lesson2_id = l2_res.json()["id"]

        canc_res = client.post(
            f"/api/v1/lessons/{lesson2_id}/cancel",
            json={"reason": "Болезнь ученика"},
            headers=teacher["headers"],
        )
        assert canc_res.status_code == 200

        # 5. Проверяем /analytics/overview
        overview_res = client.get("/api/v1/analytics/overview", headers=teacher["headers"])
        assert overview_res.status_code == 200
        overview = overview_res.json()
        assert overview["completed_lessons"] >= 1
        assert overview["cancelled_lessons"] >= 1
        assert overview["completed_hours"] >= 1.5
        # 1.5 часа * 2000 = 3000 Gross. Комиссия школы 20% = 600. Net = 2400.
        assert overview["gross_revenue"] >= 3000.0
        assert overview["net_income"] >= 2400.0
        assert overview["completion_rate"] > 0

        # 6. Проверяем /analytics/dynamics
        dyn_res = client.get("/api/v1/analytics/dynamics?interval=week", headers=teacher["headers"])
        assert dyn_res.status_code == 200
        points = dyn_res.json()
        assert isinstance(points, list)

        # 7. Проверяем /analytics/formats
        fmt_res = client.get("/api/v1/analytics/formats", headers=teacher["headers"])
        assert fmt_res.status_code == 200
        formats = fmt_res.json()
        indiv_fmt = next((f for f in formats if f["format"] == "individual"), None)
        assert indiv_fmt is not None
        assert indiv_fmt["completed_hours"] >= 1.5
        assert indiv_fmt["net_income"] >= 2400.0

        # 8. Проверяем /analytics/clients
        cls_res = client.get("/api/v1/analytics/clients?sort=hours", headers=teacher["headers"])
        assert cls_res.status_code == 200
        cls_list = cls_res.json()
        matching = next((c for c in cls_list if c["client_id"] == client_id), None)
        assert matching is not None
        assert matching["completed_hours"] == 1.5
        assert matching["net_income"] == 2400.0
        assert matching["completed_count"] == 1
        assert matching["cancelled_count"] == 1
        assert matching["attendance_rate"] == 50.0

    def test_analytics_rbac_protection(self, client: httpx.Client, student_user):
        """Ученик не имеет доступа к аналитике и настройкам календаря (403 Forbidden)."""
        student = student_user

        res1 = client.get("/api/v1/analytics/overview", headers=student["headers"])
        assert res1.status_code == 403

        res2 = client.get("/api/v1/analytics/dynamics", headers=student["headers"])
        assert res2.status_code == 403

        res3 = client.get("/api/v1/analytics/formats", headers=student["headers"])
        assert res3.status_code == 403

        res4 = client.get("/api/v1/analytics/clients", headers=student["headers"])
        assert res4.status_code == 403

        res5 = client.get("/api/v1/integrations/calendar/settings", headers=student["headers"])
        assert res5.status_code == 403
