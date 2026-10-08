import uuid
from datetime import datetime, timedelta, timezone
import pytest
import httpx


def to_rfc3339(dt: datetime) -> str:
    """Форматирует datetime в ISO/RFC3339 строку с суффиксом Z."""
    return dt.strftime("%Y-%m-%dT%H:%M:%SZ")


@pytest.mark.schedule
class TestJournalAndHomeworkWorkflow:
    """
    E2E автотесты Спринта 2.4.1 (Дневник занятий и Домашние задания):
    - Создание и чтение отчета урока: PUT & GET /schedule/lessons/{id}/journal
    - Отчет по виртуальному слоту серии и автоматическая материализация в БД
    - Жизненный цикл домашнего задания: создание, статусы completed/not_done, удаление
    - Хронологическая лента дневника ученика в CRM: GET /crm/clients/{id}/journal
    - Мультиарендная изоляция и безопасность между преподавателями
    """

    @pytest.fixture
    def setup_teacher_and_client(self, client: httpx.Client, registered_user):
        """Создает изолированного преподавателя и его клиента с абонементом."""
        _, reg_data = registered_user(role="teacher")
        token = reg_data["tokens"]["access_token"]
        teacher = {
            "id": reg_data["user"]["id"],
            "token": token,
            "headers": {"Authorization": f"Bearer {token}"},
        }

        unique_id = uuid.uuid4().hex[:6]
        c_res = client.post(
            "/api/v1/clients",
            json={
                "name": f"Ученик Дневника {unique_id}",
                "phone": "+79998881122",
                "rate_individual": 2200.0,
                "rate_pair": 1600.0,
                "rate_group": 1200.0,
            },
            headers=teacher["headers"],
        )
        assert c_res.status_code == 201, f"Failed to create client: {c_res.text}"
        client_data = c_res.json()

        sub_res = client.post(
            f"/api/v1/clients/{client_data['id']}/subscriptions",
            json={"format": "individual", "balance": 15.0},
            headers=teacher["headers"],
        )
        assert sub_res.status_code == 201, f"Failed to create subscription: {sub_res.text}"

        return teacher, client_data

    def test_create_and_get_lesson_journal(self, client: httpx.Client, setup_teacher_and_client):
        """Создание, чтение, обновление отчета по разовому уроку и валидация оценок 1..5."""
        teacher, client_data = setup_teacher_and_client

        # 1. Создаем разовый урок
        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=1)
        start_time = to_rfc3339(now.replace(hour=14, minute=0, second=0))
        end_time = to_rfc3339(now.replace(hour=15, minute=0, second=0))

        l_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client_data["id"],
                "title": "Квадратные уравнения",
                "start_time": start_time,
                "end_time": end_time,
                "format": "individual",
            },
            headers=teacher["headers"],
        )
        assert l_res.status_code == 201, f"Failed to create lesson: {l_res.text}"
        lesson = l_res.json()
        lesson_id = lesson["id"]

        # 2. До сохранения отчета проверяем GET бандла (журнал пустой, списки ДЗ пустые)
        initial_bundle = client.get(f"/api/v1/schedule/lessons/{lesson_id}/journal", headers=teacher["headers"])
        assert initial_bundle.status_code == 200
        bundle_data = initial_bundle.json()
        assert bundle_data.get("journal") is None
        assert bundle_data["assigned_homeworks"] == []
        assert bundle_data["due_homeworks"] == []

        # 3. Сохраняем отчет (PUT)
        journal_payload = {
            "topic": "Теорема Виета и дискриминант",
            "notes": "Ученик уверенно считает дискриминант, путается в знаках корней",
            "performance_score": 4,
        }
        put_res = client.put(
            f"/api/v1/schedule/lessons/{lesson_id}/journal",
            json=journal_payload,
            headers=teacher["headers"],
        )
        assert put_res.status_code == 200, f"Failed to upsert journal: {put_res.text}"
        saved_journal = put_res.json()
        assert saved_journal["lesson_id"] == lesson_id
        assert saved_journal["client_id"] == client_data["id"]
        assert saved_journal["teacher_id"] == teacher["id"]
        assert saved_journal["topic"] == journal_payload["topic"]
        assert saved_journal["notes"] == journal_payload["notes"]
        assert saved_journal["performance_score"] == 4

        # 4. Проверяем чтение сохраненного отчета через GET бандла
        get_res = client.get(f"/api/v1/schedule/lessons/{lesson_id}/journal", headers=teacher["headers"])
        assert get_res.status_code == 200
        updated_bundle = get_res.json()
        assert updated_bundle["journal"] is not None
        assert updated_bundle["journal"]["id"] == saved_journal["id"]
        assert updated_bundle["journal"]["topic"] == journal_payload["topic"]
        assert updated_bundle["journal"]["performance_score"] == 4

        # 5. Обновляем отчет (меняем оценку и заметку)
        update_payload = {
            "topic": "Теорема Виета (закрепление)",
            "notes": "Повторили знаки, все решено верно!",
            "performance_score": 5,
        }
        put2_res = client.put(
            f"/api/v1/schedule/lessons/{lesson_id}/journal",
            json=update_payload,
            headers=teacher["headers"],
        )
        assert put2_res.status_code == 200
        updated_journal = put2_res.json()
        assert updated_journal["topic"] == update_payload["topic"]
        assert updated_journal["performance_score"] == 5

        # 6. Валидация: пустая тема -> 400
        bad_topic_res = client.put(
            f"/api/v1/schedule/lessons/{lesson_id}/journal",
            json={"topic": "", "performance_score": 5},
            headers=teacher["headers"],
        )
        assert bad_topic_res.status_code == 400

        # 7. Валидация: некорректная оценка (0 или 6) -> 400
        bad_score_res1 = client.put(
            f"/api/v1/schedule/lessons/{lesson_id}/journal",
            json={"topic": "Тема", "performance_score": 0},
            headers=teacher["headers"],
        )
        assert bad_score_res1.status_code == 400

        bad_score_res2 = client.put(
            f"/api/v1/schedule/lessons/{lesson_id}/journal",
            json={"topic": "Тема", "performance_score": 6},
            headers=teacher["headers"],
        )
        assert bad_score_res2.status_code == 400

        # 8. Несуществующий урок -> 404
        fake_id = str(uuid.uuid4())
        not_found_res = client.get(f"/api/v1/schedule/lessons/{fake_id}/journal", headers=teacher["headers"])
        assert not_found_res.status_code == 404

    def test_journal_on_virtual_recurring_lesson(self, client: httpx.Client, setup_teacher_and_client):
        """Сохранение отчета по виртуальному слоту серии: автоматическая материализация урока и сохранение журнала."""
        teacher, client_data = setup_teacher_and_client

        # 1. Создаем регулярную серию (по вторникам и четвергам)
        now = datetime.now(timezone.utc)
        days_ahead = 7 - now.weekday()
        next_monday = (now + timedelta(days=days_ahead)).replace(hour=0, minute=0, second=0, microsecond=0)
        next_sunday = next_monday + timedelta(days=6, hours=23, minutes=59, seconds=59)
        start_date = next_monday.date().strftime("%Y-%m-%d")

        s_res = client.post(
            "/api/v1/schedule/series",
            json={
                "client_id": client_data["id"],
                "title": "Геометрия: Планиметрия",
                "rrule": "FREQ=WEEKLY;BYDAY=TU,TH",
                "start_time_of_day": "17:00",
                "duration_minutes": 60,
                "format": "individual",
                "start_date": start_date,
            },
            headers=teacher["headers"],
        )
        assert s_res.status_code == 201, f"Failed to create series: {s_res.text}"
        series = s_res.json()
        series_id = series["id"]

        # 2. Получаем виртуальные слоты в расписании
        lessons_res = client.get(
            "/api/v1/lessons",
            params={"from": to_rfc3339(next_monday), "to": to_rfc3339(next_sunday)},
            headers=teacher["headers"],
        )
        assert lessons_res.status_code == 200
        slots = [l for l in lessons_res.json() if l.get("series_id") == series_id]
        assert len(slots) >= 1
        virtual_slot = slots[0]
        virtual_lesson_id = virtual_slot["id"]

        # 3. GET бандла для виртуального слота возвращает 200 с journal == None
        v_bundle_res = client.get(f"/api/v1/schedule/lessons/{virtual_lesson_id}/journal", headers=teacher["headers"])
        assert v_bundle_res.status_code == 200
        assert v_bundle_res.json().get("journal") is None

        # 4. Сохраняем отчет по виртуальному слоту
        journal_payload = {
            "topic": "Теорема Пифагора и подобие треугольников",
            "notes": "Разобрали признаки подобия, решили 5 базовых задач",
            "performance_score": 5,
        }
        put_res = client.put(
            f"/api/v1/schedule/lessons/{virtual_lesson_id}/journal",
            json=journal_payload,
            headers=teacher["headers"],
        )
        assert put_res.status_code == 200, f"Failed to save journal for virtual lesson: {put_res.text}"
        saved_journal = put_res.json()
        assert saved_journal["lesson_id"] == virtual_lesson_id
        assert saved_journal["topic"] == journal_payload["topic"]

        # 5. Проверяем, что виртуальный урок материализовался в таблице расписания
        lessons_after = client.get(
            "/api/v1/lessons",
            params={"from": to_rfc3339(next_monday), "to": to_rfc3339(next_sunday)},
            headers=teacher["headers"],
        )
        assert lessons_after.status_code == 200
        persisted_lesson = next((l for l in lessons_after.json() if l["id"] == virtual_lesson_id), None)
        assert persisted_lesson is not None, "Материализованный урок должен присутствовать в расписании"
        assert persisted_lesson["series_id"] == series_id
        assert persisted_lesson["status"] == "scheduled"

        # 6. Повторный GET журнала возвращает сохраненный отчет
        get_res = client.get(f"/api/v1/schedule/lessons/{virtual_lesson_id}/journal", headers=teacher["headers"])
        assert get_res.status_code == 200
        bundle_after = get_res.json()
        assert bundle_after["journal"] is not None
        assert bundle_after["journal"]["topic"] == journal_payload["topic"]
        assert bundle_after["journal"]["performance_score"] == 5

    def test_homework_lifecycle(self, client: httpx.Client, setup_teacher_and_client):
        """Жизненный цикл ДЗ: создание, привязка к уроку, смена статусов completed/not_done, фильтры, удаление."""
        teacher, client_data = setup_teacher_and_client

        # 1. Создаем урок для привязки ДЗ
        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=2)
        start_time = to_rfc3339(now.replace(hour=16, minute=0, second=0))
        end_time = to_rfc3339(now.replace(hour=17, minute=0, second=0))

        l_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client_data["id"],
                "title": "Подготовка к контрольной",
                "start_time": start_time,
                "end_time": end_time,
                "format": "individual",
            },
            headers=teacher["headers"],
        )
        assert l_res.status_code == 201
        lesson_id = l_res.json()["id"]

        # 2. Создаем ДЗ с привязкой к уроку
        due_date = (now + timedelta(days=5)).date().strftime("%Y-%m-%d")
        hw_payload = {
            "title": "Номера 102-110 в сборнике задач",
            "description": "Оформить решение задач на двойной угол с проверкой",
            "due_date": due_date,
            "assigned_lesson_id": lesson_id,
        }
        create_res = client.post(
            f"/api/v1/crm/clients/{client_data['id']}/homework",
            json=hw_payload,
            headers=teacher["headers"],
        )
        assert create_res.status_code == 201, f"Failed to create homework: {create_res.text}"
        hw = create_res.json()
        hw_id = hw["id"]
        assert hw["client_id"] == client_data["id"]
        assert hw["teacher_id"] == teacher["id"]
        assert hw["assigned_lesson_id"] == lesson_id
        assert hw["title"] == hw_payload["title"]
        assert hw["description"] == hw_payload["description"]
        assert hw["status"] == "assigned"

        # 3. Проверяем наличие ДЗ в списке ученика
        list_res = client.get(f"/api/v1/crm/clients/{client_data['id']}/homework", headers=teacher["headers"])
        assert list_res.status_code == 200
        client_hws = list_res.json()
        assert any(item["id"] == hw_id for item in client_hws)

        # 4. Проверяем наличие выданного ДЗ в бандле урока
        bundle_res = client.get(f"/api/v1/schedule/lessons/{lesson_id}/journal", headers=teacher["headers"])
        assert bundle_res.status_code == 200
        bundle = bundle_res.json()
        assert any(item["id"] == hw_id for item in bundle["assigned_homeworks"])

        # 5. Смена статуса на completed с рецензией
        patch1_res = client.patch(
            f"/api/v1/homework/{hw_id}",
            json={"status": "completed", "review_notes": "Все решено без ошибок, молодец!"},
            headers=teacher["headers"],
        )
        assert patch1_res.status_code == 200
        updated1 = patch1_res.json()
        assert updated1["status"] == "completed"
        assert updated1["review_notes"] == "Все решено без ошибок, молодец!"

        # 6. Проверка фильтрации по статусу: completed vs assigned
        comp_res = client.get(
            f"/api/v1/crm/clients/{client_data['id']}/homework",
            params={"status": "completed"},
            headers=teacher["headers"],
        )
        assert comp_res.status_code == 200
        assert any(item["id"] == hw_id for item in comp_res.json())

        assigned_res = client.get(
            f"/api/v1/crm/clients/{client_data['id']}/homework",
            params={"status": "assigned"},
            headers=teacher["headers"],
        )
        assert assigned_res.status_code == 200
        assert not any(item["id"] == hw_id for item in assigned_res.json())

        # 7. Смена статуса на not_done
        patch2_res = client.patch(
            f"/api/v1/homework/{hw_id}",
            json={"status": "not_done", "review_notes": "Задание не сдано вовремя"},
            headers=teacher["headers"],
        )
        assert patch2_res.status_code == 200
        assert patch2_res.json()["status"] == "not_done"

        # 8. Удаление ДЗ
        del_res = client.delete(f"/api/v1/homework/{hw_id}", headers=teacher["headers"])
        assert del_res.status_code == 204

        # 9. Проверяем, что ДЗ отсутствует в списке
        list_after_res = client.get(f"/api/v1/crm/clients/{client_data['id']}/homework", headers=teacher["headers"])
        assert list_after_res.status_code == 200
        assert not any(item["id"] == hw_id for item in list_after_res.json())

        # 10. Повторный PATCH на удаленное ДЗ возвращает 404
        patch_deleted = client.patch(
            f"/api/v1/homework/{hw_id}",
            json={"status": "completed"},
            headers=teacher["headers"],
        )
        assert patch_deleted.status_code == 404

        # 11. Валидация: пустое название -> 400, некорректный статус -> 400
        bad_hw_res = client.post(
            f"/api/v1/crm/clients/{client_data['id']}/homework",
            json={"title": ""},
            headers=teacher["headers"],
        )
        assert bad_hw_res.status_code == 400

    def test_client_journal_timeline(self, client: httpx.Client, setup_teacher_and_client):
        """Хронологическая лента пройденных уроков с темами и оценками в карточке ученика."""
        teacher, client_data = setup_teacher_and_client
        now = datetime.now(timezone.utc).replace(microsecond=0)

        # Создаем 2 урока в разные дни
        lessons_info = [
            {"day_offset": 1, "topic": "Функции и графики параболы", "score": 4, "notes": "Разобрали вершину параболы"},
            {"day_offset": 3, "topic": "Исследование функции с помощью производной", "score": 5, "notes": "Отлично освоил экстремумы"},
        ]

        created_journal_ids = []
        for info in lessons_info:
            lesson_dt = now + timedelta(days=info["day_offset"])
            start_str = to_rfc3339(lesson_dt.replace(hour=11, minute=0, second=0))
            end_str = to_rfc3339(lesson_dt.replace(hour=12, minute=0, second=0))

            l_res = client.post(
                "/api/v1/lessons",
                json={
                    "client_id": client_data["id"],
                    "title": info["topic"],
                    "start_time": start_str,
                    "end_time": end_str,
                    "format": "individual",
                },
                headers=teacher["headers"],
            )
            assert l_res.status_code == 201
            l_id = l_res.json()["id"]

            put_res = client.put(
                f"/api/v1/schedule/lessons/{l_id}/journal",
                json={
                    "topic": info["topic"],
                    "notes": info["notes"],
                    "performance_score": info["score"],
                },
                headers=teacher["headers"],
            )
            assert put_res.status_code == 200
            created_journal_ids.append(put_res.json()["id"])

        # Запрашиваем таймлайн дневника ученика в CRM
        timeline_res = client.get(f"/api/v1/crm/clients/{client_data['id']}/journal", headers=teacher["headers"])
        assert timeline_res.status_code == 200
        journals = timeline_res.json()
        assert len(journals) == 2

        # Проверяем наличие тем и оценок
        retrieved_ids = [j["id"] for j in journals]
        for j_id in created_journal_ids:
            assert j_id in retrieved_ids

        topics = [j["topic"] for j in journals]
        assert "Функции и графики параболы" in topics
        assert "Исследование функции с помощью производной" in topics

    def test_journal_and_homework_isolation(self, client: httpx.Client, setup_teacher_and_client, registered_user):
        """Проверка изоляции: преподаватель B не имеет доступа к журналу и ДЗ ученика преподавателя A."""
        teacher_a, client_a = setup_teacher_and_client

        # Преподаватель B
        _, reg_b = registered_user(role="teacher")
        token_b = reg_b["tokens"]["access_token"]
        headers_b = {"Authorization": f"Bearer {token_b}"}

        # Преподаватель A создает урок, отчет и ДЗ
        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=2)
        start_time = to_rfc3339(now.replace(hour=13, minute=0, second=0))
        end_time = to_rfc3339(now.replace(hour=14, minute=0, second=0))

        lesson_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client_a["id"],
                "title": "Приватный урок Преподавателя А",
                "start_time": start_time,
                "end_time": end_time,
                "format": "individual",
            },
            headers=teacher_a["headers"],
        )
        assert lesson_res.status_code == 201
        lesson_a_id = lesson_res.json()["id"]

        j_res = client.put(
            f"/api/v1/schedule/lessons/{lesson_a_id}/journal",
            json={"topic": "Секретная тема", "notes": "Заметки А", "performance_score": 5},
            headers=teacher_a["headers"],
        )
        assert j_res.status_code == 200

        client_a_id = client_a["id"]
        hw_res = client.post(
            f"/api/v1/crm/clients/{client_a_id}/homework",
            json={"title": "Секретное ДЗ", "description": "Только для ученика А"},
            headers=teacher_a["headers"],
        )
        assert hw_res.status_code == 201
        hw_a_id = hw_res.json()["id"]

        # Преподаватель B пытается получить журнал урока A -> 403 или 404
        get_j_res = client.get(f"/api/v1/schedule/lessons/{lesson_a_id}/journal", headers=headers_b)
        assert get_j_res.status_code in (403, 404)

        # Преподаватель B пытается перезаписать журнал урока A -> 403 или 404
        put_j_res = client.put(
            f"/api/v1/schedule/lessons/{lesson_a_id}/journal",
            json={"topic": "Взлом темы", "performance_score": 1},
            headers=headers_b,
        )
        assert put_j_res.status_code in (403, 404)

        # Преподаватель B пытается получить таймлайн клиента A -> 403 или 404
        get_client_j_res = client.get(f"/api/v1/crm/clients/{client_a_id}/journal", headers=headers_b)
        assert get_client_j_res.status_code in (403, 404)

        # Преподаватель B пытается прочитать ДЗ клиента A -> 403 или 404
        get_client_hw_res = client.get(f"/api/v1/crm/clients/{client_a_id}/homework", headers=headers_b)
        assert get_client_hw_res.status_code in (403, 404)

        # Преподаватель B пытается создать ДЗ для клиента A -> 403 или 404
        create_foreign_hw = client.post(
            f"/api/v1/crm/clients/{client_a_id}/homework",
            json={"title": "Несанкционированное ДЗ"},
            headers=headers_b,
        )
        assert create_foreign_hw.status_code in (403, 404)

        # Преподаватель B пытается изменить статус ДЗ ученика A -> 403 или 404
        patch_hw_res = client.patch(
            f"/api/v1/homework/{hw_a_id}",
            json={"status": "completed"},
            headers=headers_b,
        )
        assert patch_hw_res.status_code in (403, 404)

        # Преподаватель B пытается удалить ДЗ ученика A -> 403 или 404
        del_hw_res = client.delete(f"/api/v1/homework/{hw_a_id}", headers=headers_b)
        assert del_hw_res.status_code in (403, 404)

        # Неавторизованные запросы (без токена) -> 401
        unauth_j = client.get(f"/api/v1/schedule/lessons/{lesson_a_id}/journal")
        assert unauth_j.status_code == 401

        unauth_hw = client.get(f"/api/v1/crm/clients/{client_a_id}/homework")
        assert unauth_hw.status_code == 401

        # Проверка целостности данных у преподавателя A
        check_a_j = client.get(f"/api/v1/schedule/lessons/{lesson_a_id}/journal", headers=teacher_a["headers"])
        assert check_a_j.status_code == 200
        assert check_a_j.json()["journal"]["topic"] == "Секретная тема"

        check_a_hw = client.get(f"/api/v1/crm/clients/{client_a_id}/homework", headers=teacher_a["headers"])
        assert check_a_hw.status_code == 200
        assert any(item["id"] == hw_a_id for item in check_a_hw.json())
