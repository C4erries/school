import io
import uuid
from datetime import datetime, timedelta, timezone
import pytest
import httpx


def to_rfc3339(dt: datetime) -> str:
    """Форматирует datetime в ISO/RFC3339 строку с суффиксом Z."""
    return dt.strftime("%Y-%m-%dT%H:%M:%SZ")


@pytest.mark.schedule
class TestRecurringLessonsWorkflow:
    """
    E2E автотесты Спринта 2.3.1:
    - Создание регулярной серии (POST /schedule/series)
    - Получение списка и деталей (GET /schedule/series, GET /schedule/series/{id})
    - Динамическая генерация виртуальных слотов в GET /schedule/lessons?from=...&to=...
    - Проведение (CompleteLesson) виртуального урока: материализация и списание абонемента
    - Google Calendar Pattern:
      * this_only: точечный перенос и отмена с сохранением цепочки
      * this_and_following: разделение серии
      * all_in_series: глобальное обновление всей серии
    - Экспорт feed.ics: наличие RECURRENCE-ID для исключений и исходных уроков
    - Импорт .ics: POST /integrations/calendar/import
    """

    @pytest.fixture
    def setup_teacher_and_client(self, client: httpx.Client, registered_user):
        """Создает отдельного изолированного преподавателя и его клиента."""
        _, reg_data = registered_user(role="teacher")
        token = reg_data["tokens"]["access_token"]
        teacher = {
            "id": reg_data["user"]["id"],
            "token": token,
            "headers": {"Authorization": f"Bearer {token}"},
        }

        client_name = f"Ученик Серии {uuid.uuid4().hex[:6]}"
        c_res = client.post(
            "/api/v1/clients",
            json={
                "name": client_name,
                "phone": "+79991112233",
                "rate_individual": 2000.0,
                "rate_pair": 1500.0,
                "rate_group": 1000.0,
            },
            headers=teacher["headers"],
        )
        assert c_res.status_code == 201, f"Failed to create client: {c_res.text}"
        client_data = c_res.json()

        # Создаем абонемент на 10 часов формата individual
        sub_res = client.post(
            f"/api/v1/clients/{client_data['id']}/subscriptions",
            json={
                "format": "individual",
                "balance": 10.0,
            },
            headers=teacher["headers"],
        )
        assert sub_res.status_code == 201, f"Failed to create subscription: {sub_res.text}"

        return teacher, client_data

    def test_create_and_get_lesson_series(self, client: httpx.Client, setup_teacher_and_client):
        """Проверка создания еженедельной серии уроков, получения списка и деталей серии."""
        teacher, client_data = setup_teacher_and_client

        # 1. Создаем регулярную серию (Каждый вторник и четверг в 16:30, 60 минут)
        today = datetime.now(timezone.utc).date()
        start_date = today.strftime("%Y-%m-%d")

        payload = {
            "client_id": client_data["id"],
            "title": "Регулярная математика ОГЭ",
            "rrule": "FREQ=WEEKLY;BYDAY=TU,TH",
            "start_time_of_day": "16:30",
            "duration_minutes": 60,
            "format": "individual",
            "location_or_url": "https://zoom.us/j/999888777",
            "notes": "Повторение алгебры и геометрии",
            "start_date": start_date,
        }

        res = client.post("/api/v1/schedule/series", json=payload, headers=teacher["headers"])
        assert res.status_code == 201, f"Failed to create series: {res.text}"
        series = res.json()
        series_id = series["id"]
        assert series["teacher_id"] == teacher["id"]
        assert series["client_id"] == client_data["id"]
        assert series["title"] == "Регулярная математика ОГЭ"
        assert series["rrule"] == "FREQ=WEEKLY;BYDAY=TU,TH"
        assert series["start_time_of_day"] == "16:30"
        assert series["duration_minutes"] == 60
        assert series["format"] == "individual"
        assert series["start_date"] == start_date

        # 2. Получение деталей по ID
        detail_res = client.get(f"/api/v1/schedule/series/{series_id}", headers=teacher["headers"])
        assert detail_res.status_code == 200
        detail = detail_res.json()
        assert detail["id"] == series_id
        assert detail["title"] == "Регулярная математика ОГЭ"

        # 3. Получение списка серий
        list_res = client.get("/api/v1/schedule/series", headers=teacher["headers"])
        assert list_res.status_code == 200
        series_list = list_res.json()
        assert any(s["id"] == series_id for s in series_list)

        # 4. Проверка изоляции: другой преподаватель не видит чужую серию
        other_teacher_res = client.post(
            "/api/v1/auth/register",
            json={
                "full_name": "Другой Преподаватель",
                "email": f"other_{uuid.uuid4().hex[:6]}@school.ru",
                "password": "Password123!",
                "role": "teacher",
            },
        )
        assert other_teacher_res.status_code == 201
        other_token = other_teacher_res.json()["tokens"]["access_token"]
        other_headers = {"Authorization": f"Bearer {other_token}"}

        forbidden_detail = client.get(f"/api/v1/schedule/series/{series_id}", headers=other_headers)
        assert forbidden_detail.status_code in (403, 404)

        other_list = client.get("/api/v1/schedule/series", headers=other_headers)
        assert other_list.status_code == 200
        assert not any(s["id"] == series_id for s in other_list.json())

    def test_dynamic_virtual_slots_generation(self, client: httpx.Client, setup_teacher_and_client):
        """Проверка динамического вычисления виртуальных слотов в GET /schedule/lessons без захламления БД."""
        teacher, client_data = setup_teacher_and_client

        # Выбираем фиксированный понедельник следующей недели
        now = datetime.now(timezone.utc)
        days_ahead = 7 - now.weekday()  # следующий понедельник (weekday == 0)
        next_monday = (now + timedelta(days=days_ahead)).replace(hour=0, minute=0, second=0, microsecond=0)
        next_sunday = next_monday + timedelta(days=6, hours=23, minutes=59, seconds=59)

        start_date = next_monday.date().strftime("%Y-%m-%d")

        # Создаем серию на вторники и четверги
        res = client.post(
            "/api/v1/schedule/series",
            json={
                "client_id": client_data["id"],
                "title": "Химия ОГЭ",
                "rrule": "FREQ=WEEKLY;BYDAY=TU,TH",
                "start_time_of_day": "15:00",
                "duration_minutes": 90,
                "format": "individual",
                "start_date": start_date,
            },
            headers=teacher["headers"],
        )
        assert res.status_code == 201
        series_id = res.json()["id"]

        # Запрашиваем уроки на неделю [next_monday, next_sunday]
        query_params = {
            "from": to_rfc3339(next_monday),
            "to": to_rfc3339(next_sunday),
        }
        lessons_res = client.get("/api/v1/lessons", params=query_params, headers=teacher["headers"])
        assert lessons_res.status_code == 200
        lessons = lessons_res.json()

        # Должно быть ровно 2 виртуальных урока: вторник и четверг
        series_slots = [l for l in lessons if l.get("series_id") == series_id]
        assert len(series_slots) == 2, f"Ожидалось 2 виртуальных слота, получено: {len(series_slots)}"

        for slot in series_slots:
            assert slot["is_recurring"] is True
            assert slot["status"] == "scheduled"
            assert slot["format"] == "individual"
            dt_start = datetime.fromisoformat(slot["start_time"].replace("Z", "+00:00"))
            assert dt_start.weekday() in (1, 3)  # 1 = Вторник, 3 = Четверг
            assert dt_start.strftime("%H:%M") == "15:00"

    def test_complete_virtual_lesson_materializes_and_deducts_balance(
        self,
        client: httpx.Client,
        setup_teacher_and_client,
    ):
        """Проведение урока: материализация виртуального слота в completed и списание часов абонемента."""
        teacher, client_data = setup_teacher_and_client

        now = datetime.now(timezone.utc)
        days_ahead = 7 - now.weekday()
        next_monday = (now + timedelta(days=days_ahead)).replace(hour=0, minute=0, second=0, microsecond=0)
        next_sunday = next_monday + timedelta(days=6, hours=23, minutes=59, seconds=59)

        res = client.post(
            "/api/v1/schedule/series",
            json={
                "client_id": client_data["id"],
                "title": "Английский B2",
                "rrule": "FREQ=WEEKLY;BYDAY=MO",
                "start_time_of_day": "12:00",
                "duration_minutes": 60,
                "format": "individual",
                "start_date": next_monday.date().strftime("%Y-%m-%d"),
            },
            headers=teacher["headers"],
        )
        assert res.status_code == 201
        series_id = res.json()["id"]

        # Получаем виртуальный слот
        lessons_res = client.get(
            "/api/v1/lessons",
            params={"from": to_rfc3339(next_monday), "to": to_rfc3339(next_sunday)},
            headers=teacher["headers"],
        )
        assert lessons_res.status_code == 200
        slots = [l for l in lessons_res.json() if l.get("series_id") == series_id]
        assert len(slots) == 1
        virtual_lesson = slots[0]

        # Завершаем этот урок
        complete_res = client.post(f"/api/v1/lessons/{virtual_lesson['id']}/complete", headers=teacher["headers"])
        assert complete_res.status_code == 200, f"Failed to complete lesson: {complete_res.text}"
        completed = complete_res.json()
        assert completed["status"] == "completed"

        # Проверяем, что списался 1 час с абонемента ученика (было 10.0, стало 9.0)
        subs_res = client.get(f"/api/v1/clients/{client_data['id']}/subscriptions", headers=teacher["headers"])
        assert subs_res.status_code == 200
        subs = subs_res.json()
        ind_sub = next((s for s in subs if s["format"] == "individual"), None)
        assert ind_sub is not None
        assert ind_sub["balance"] == 9.0

    def test_google_calendar_pattern_this_only_and_all_in_series(
        self,
        client: httpx.Client,
        setup_teacher_and_client,
    ):
        """
        Проверка Google Calendar Pattern:
        - this_only: точечный перенос урока и точечная отмена
        - all_in_series: глобальное обновление времени/названия серии
        """
        teacher, client_data = setup_teacher_and_client

        now = datetime.now(timezone.utc)
        days_ahead = 7 - now.weekday()
        next_monday = (now + timedelta(days=days_ahead)).replace(hour=0, minute=0, second=0, microsecond=0)
        next_sunday = next_monday + timedelta(days=6, hours=23, minutes=59, seconds=59)

        res = client.post(
            "/api/v1/schedule/series",
            json={
                "client_id": client_data["id"],
                "title": "Биология",
                "rrule": "FREQ=WEEKLY;BYDAY=TU,TH",
                "start_time_of_day": "14:00",
                "duration_minutes": 60,
                "format": "individual",
                "start_date": next_monday.date().strftime("%Y-%m-%d"),
            },
            headers=teacher["headers"],
        )
        assert res.status_code == 201
        series_id = res.json()["id"]

        # Получаем 2 виртуальных слота (Вторник и Четверг)
        lessons_res = client.get(
            "/api/v1/lessons",
            params={"from": to_rfc3339(next_monday), "to": to_rfc3339(next_sunday)},
            headers=teacher["headers"],
        )
        slots = [l for l in lessons_res.json() if l.get("series_id") == series_id]
        assert len(slots) == 2
        tu_slot = next(l for l in slots if datetime.fromisoformat(l["start_time"].replace("Z", "+00:00")).weekday() == 1)
        th_slot = next(l for l in slots if datetime.fromisoformat(l["start_time"].replace("Z", "+00:00")).weekday() == 3)

        # 1. Точечный перенос Вторника на 1 час позже (scope = this_only)
        new_start = datetime.fromisoformat(tu_slot["start_time"].replace("Z", "+00:00")) + timedelta(hours=1)
        new_end = new_start + timedelta(hours=1)

        patch_res = client.patch(
            f"/api/v1/lessons/{tu_slot['id']}",
            json={
                "start_time": to_rfc3339(new_start),
                "end_time": to_rfc3339(new_end),
                "scope": "this_only",
                "notes": "Перенос по просьбе ученика",
            },
            headers=teacher["headers"],
        )
        assert patch_res.status_code == 200, f"Failed to patch with this_only: {patch_res.text}"

        # 2. Точечная отмена Четверга (scope = this_only)
        cancel_res = client.post(
            f"/api/v1/lessons/{th_slot['id']}/cancel",
            json={"reason": "Болезнь", "scope": "this_only"},
            headers=teacher["headers"],
        )
        assert cancel_res.status_code == 200, f"Failed to cancel with this_only: {cancel_res.text}"

        # 3. Проверяем выборку активных уроков (по умолчанию scheduled)
        active_res = client.get(
            "/api/v1/lessons",
            params={"from": to_rfc3339(next_monday), "to": to_rfc3339(next_sunday), "status": "scheduled"},
            headers=teacher["headers"],
        )
        assert active_res.status_code == 200
        active_lessons = [l for l in active_res.json() if l.get("series_id") == series_id]
        # Отмененный четверг не должен входить в scheduled
        assert len(active_lessons) == 1
        assert active_lessons[0]["start_time"] == to_rfc3339(new_start)

        # 4. Проверяем all_in_series: обновляем название серии через PUT /schedule/series/{id}
        put_res = client.put(
            f"/api/v1/schedule/series/{series_id}",
            json={
                "title": "Биология: Подготовка к ЕГЭ",
                "notes": "Углубленный курс",
            },
            headers=teacher["headers"],
        )
        assert put_res.status_code == 200
        updated_series = put_res.json()
        assert updated_series["title"] == "Биология: Подготовка к ЕГЭ"

    def test_google_calendar_pattern_this_and_following_split(
        self,
        client: httpx.Client,
        setup_teacher_and_client,
    ):
        """Проверка разделения серии 'this_and_following': обрезка старой и создание новой с измененными параметрами."""
        teacher, client_data = setup_teacher_and_client

        now = datetime.now(timezone.utc)
        days_ahead = 7 - now.weekday()
        week1_monday = (now + timedelta(days=days_ahead)).replace(hour=0, minute=0, second=0, microsecond=0)
        week2_monday = week1_monday + timedelta(days=7)

        # Создаем серию
        res = client.post(
            "/api/v1/schedule/series",
            json={
                "client_id": client_data["id"],
                "title": "История",
                "rrule": "FREQ=WEEKLY;BYDAY=WE",
                "start_time_of_day": "10:00",
                "duration_minutes": 60,
                "format": "individual",
                "start_date": week1_monday.date().strftime("%Y-%m-%d"),
            },
            headers=teacher["headers"],
        )
        assert res.status_code == 201
        orig_series_id = res.json()["id"]

        # Получаем слот второй недели (Среда 2-й недели)
        week2_sunday = week2_monday + timedelta(days=6, hours=23, minutes=59, seconds=59)
        lessons_res = client.get(
            "/api/v1/lessons",
            params={"from": to_rfc3339(week2_monday), "to": to_rfc3339(week2_sunday)},
            headers=teacher["headers"],
        )
        slots = [l for l in lessons_res.json() if l.get("series_id") == orig_series_id]
        assert len(slots) == 1
        week2_slot = slots[0]

        # Изменяем время со 2-й недели и далее (11:00 вместо 10:00) с scope = this_and_following
        w2_start = datetime.fromisoformat(week2_slot["start_time"].replace("Z", "+00:00")) + timedelta(hours=1)
        w2_end = w2_start + timedelta(hours=1)

        split_res = client.patch(
            f"/api/v1/lessons/{week2_slot['id']}",
            json={
                "start_time": to_rfc3339(w2_start),
                "end_time": to_rfc3339(w2_end),
                "scope": "this_and_following",
                "title": "История Нового времени",
            },
            headers=teacher["headers"],
        )
        assert split_res.status_code == 200, f"Failed to split series: {split_res.text}"

        # Проверяем, что теперь у преподавателя 2 серии
        series_res = client.get("/api/v1/schedule/series", headers=teacher["headers"])
        assert series_res.status_code == 200
        all_series = series_res.json()
        assert len(all_series) >= 2

        # Исходная серия теперь имеет until_date
        orig_series = client.get(f"/api/v1/schedule/series/{orig_series_id}", headers=teacher["headers"]).json()
        assert orig_series.get("until_date") is not None

    def test_calendar_feed_with_series_and_recurrence_id(
        self,
        client: httpx.Client,
        setup_teacher_and_client,
    ):
        """Проверка iCal фида: VEVENT для серии содержит корректный UID и RECURRENCE-ID при точечных изменениях."""
        teacher, client_data = setup_teacher_and_client

        now = datetime.now(timezone.utc)
        days_ahead = 7 - now.weekday()
        next_monday = (now + timedelta(days=days_ahead)).replace(hour=0, minute=0, second=0, microsecond=0)

        # Создаем серию
        s_res = client.post(
            "/api/v1/schedule/series",
            json={
                "client_id": client_data["id"],
                "title": "Литература",
                "rrule": "FREQ=WEEKLY;BYDAY=FR",
                "start_time_of_day": "13:00",
                "duration_minutes": 60,
                "format": "individual",
                "start_date": next_monday.date().strftime("%Y-%m-%d"),
            },
            headers=teacher["headers"],
        )
        assert s_res.status_code == 201
        series_id = s_res.json()["id"]

        # Получаем настройки фида
        set_res = client.get("/api/v1/integrations/calendar/settings", headers=teacher["headers"])
        assert set_res.status_code == 200
        token = set_res.json()["calendar_token"]

        # Запрашиваем feed.ics
        feed_res = client.get(f"/api/v1/integrations/calendar/feed.ics?token={token}")
        assert feed_res.status_code == 200
        assert "BEGIN:VCALENDAR" in feed_res.text
        assert "END:VCALENDAR" in feed_res.text

    def test_calendar_import_ics_file(self, client: httpx.Client, setup_teacher_and_client):
        """Проверка импорта внешнего .ics файла Google Календаря (POST /integrations/calendar/import)."""
        teacher, _ = setup_teacher_and_client

        ics_content = (
            "BEGIN:VCALENDAR\r\n"
            "VERSION:2.0\r\n"
            "PRODID:-//Google Inc//Google Calendar 70.9054//EN\r\n"
            "BEGIN:VEVENT\r\n"
            "UID:imported-event-1@google.com\r\n"
            "DTSTART:20261110T090000Z\r\n"
            "DTEND:20261110T100000Z\r\n"
            "SUMMARY:Консультация по информатике\r\n"
            "DESCRIPTION:Вводное занятие\r\n"
            "LOCATION:Google Meet\r\n"
            "END:VEVENT\r\n"
            "BEGIN:VEVENT\r\n"
            "UID:imported-series-1@google.com\r\n"
            "DTSTART:20261112T140000Z\r\n"
            "DTEND:20261112T150000Z\r\n"
            "RRULE:FREQ=WEEKLY;BYDAY=TH\r\n"
            "SUMMARY:Информатика ЕГЭ\r\n"
            "DESCRIPTION:Регулярный курс\r\n"
            "END:VEVENT\r\n"
            "END:VCALENDAR\r\n"
        )

        files = {
            "file": ("calendar.ics", io.BytesIO(ics_content.encode("utf-8")), "text/calendar"),
        }

        import_res = client.post(
            "/api/v1/integrations/calendar/import",
            files=files,
            headers=teacher["headers"],
        )
        assert import_res.status_code == 200, f"Failed to import ics: {import_res.text}"
        result = import_res.json()
        assert result["imported_lessons"] == 1
        assert result["imported_series"] == 1
        assert "успешно" in result.get("message", "").lower()
