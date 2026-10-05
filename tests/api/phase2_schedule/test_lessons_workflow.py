import uuid
from datetime import datetime, timedelta, timezone
import pytest
import httpx


def to_rfc3339(dt: datetime) -> str:
    """Форматирует datetime в ISO/RFC3339 строку с суффиксом Z."""
    return dt.strftime("%Y-%m-%dT%H:%M:%SZ")


@pytest.mark.schedule
class TestLessonsWorkflow:
    """Интеграционные тесты жизненного цикла уроков, абонементов по форматам, редактирования (PATCH), нахлёстов и коллизий кабинетов."""

    @pytest.fixture
    def setup_classroom(self, client: httpx.Client, admin_user):
        """Фикстура создания тестового кабинета администратором."""
        unique_name = f"Кабинет-Тест {uuid.uuid4().hex[:6]}"
        res = client.post(
            "/api/v1/classrooms",
            json={"name": unique_name, "capacity": 4, "color": "#3B82F6"},
            headers=admin_user["headers"],
        )
        assert res.status_code == 201, f"Failed to create classroom: {res.text}"
        return res.json()

    @pytest.fixture
    def teacher_with_client(self, client: httpx.Client, teacher_user):
        """Фикстура создания клиента для преподавателя с тарифной сеткой."""
        unique_name = f"Клиент-{uuid.uuid4().hex[:6]}"
        res = client.post(
            "/api/v1/clients",
            json={
                "name": unique_name,
                "phone": "+79997654321",
                "rate_individual": 1600.0,
                "rate_pair": 1100.0,
                "rate_group": 800.0,
                "school_percent_tag": 20,
            },
            headers=teacher_user["headers"],
        )
        assert res.status_code == 201, f"Failed to create client: {res.text}"
        return teacher_user, res.json()

    def test_create_lesson_immediately_scheduled(
        self,
        client: httpx.Client,
        teacher_with_client,
        setup_classroom,
    ):
        """Создание урока преподавателем с форматом 'individual': статус сразу scheduled."""
        teacher, client_data = teacher_with_client
        classroom = setup_classroom

        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=2)
        start_time = to_rfc3339(now.replace(hour=11, minute=0, second=0))
        end_time = to_rfc3339(now.replace(hour=12, minute=0, second=0))

        create_payload = {
            "client_id": client_data["id"],
            "classroom_id": classroom["id"],
            "start_time": start_time,
            "end_time": end_time,
            "format": "individual",
            "notes": "Урок физики: механика",
        }
        res = client.post("/api/v1/lessons", json=create_payload, headers=teacher["headers"])
        assert res.status_code == 201, f"Failed to schedule lesson: {res.text}"
        lesson = res.json()

        assert "id" in lesson
        assert lesson["teacher_id"] == teacher["id"]
        assert lesson["client_id"] == client_data["id"]
        assert lesson["classroom_id"] == classroom["id"]
        assert lesson["status"] == "scheduled"
        assert lesson["format"] == "individual"

    def test_lesson_complete_deducts_exact_hours_by_format(
        self,
        client: httpx.Client,
        teacher_with_client,
    ):
        """Завершение урока списывает точную длительность (1.5ч) из абонемента соответствующего формата (individual и pair)."""
        teacher, client_data = teacher_with_client

        # 1. Пополняем абонементы: individual = 10.0 ч, pair = 4.0 ч
        sub1_res = client.post(
            f"/api/v1/clients/{client_data['id']}/subscriptions",
            json={"format": "individual", "balance": 10.0},
            headers=teacher["headers"],
        )
        assert sub1_res.status_code == 201
        sub1_id = sub1_res.json()["id"]

        sub2_res = client.post(
            f"/api/v1/clients/{client_data['id']}/subscriptions",
            json={"format": "pair", "balance": 4.0},
            headers=teacher["headers"],
        )
        assert sub2_res.status_code == 201
        sub2_id = sub2_res.json()["id"]

        # 2. Создаем индивидуальный урок длительностью 1.5 часа (10:00 - 11:30)
        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=3)
        start_time = to_rfc3339(now.replace(hour=10, minute=0, second=0))
        end_time = to_rfc3339(now.replace(hour=11, minute=30, second=0))

        lesson_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client_data["id"],
                "start_time": start_time,
                "end_time": end_time,
                "format": "individual",
                "notes": "Индивидуальный урок",
            },
            headers=teacher["headers"],
        )
        assert lesson_res.status_code == 201
        lesson_id = lesson_res.json()["id"]

        # 3. Преподаватель завершает урок
        comp_res = client.post(
            f"/api/v1/lessons/{lesson_id}/complete",
            headers=teacher["headers"],
        )
        assert comp_res.status_code == 200, f"Failed to complete lesson: {comp_res.text}"
        assert comp_res.json()["status"] == "completed"

        # 4. Проверяем баланс абонементов:
        # individual должен уменьшиться с 10.0 до 8.5
        # pair должен остаться неизменным (4.0)
        list_sub = client.get(
            f"/api/v1/clients/{client_data['id']}/subscriptions",
            headers=teacher["headers"],
        ).json()

        target_ind = next((s for s in list_sub if s["id"] == sub1_id), None)
        assert target_ind is not None
        assert abs(target_ind["balance"] - 8.5) < 1e-4

        target_pair = next((s for s in list_sub if s["id"] == sub2_id), None)
        assert target_pair is not None
        assert abs(target_pair["balance"] - 4.0) < 1e-4

        # 5. Проверяем агрегацию балансов на клиенте
        c_info = client.get("/api/v1/clients", headers=teacher["headers"]).json()
        target_c = next((c for c in c_info if c["id"] == client_data["id"]), None)
        assert target_c is not None
        assert abs(target_c["balances"]["individual_hours"] - 8.5) < 1e-4
        assert abs(target_c["balances"]["pair_hours"] - 4.0) < 1e-4

        # 6. Проводим парное занятие на 1.0 час (14:00 - 15:00)
        p_start = to_rfc3339(now.replace(hour=14, minute=0, second=0))
        p_end = to_rfc3339(now.replace(hour=15, minute=0, second=0))
        p_lesson_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client_data["id"],
                "start_time": p_start,
                "end_time": p_end,
                "format": "pair",
            },
            headers=teacher["headers"],
        )
        assert p_lesson_res.status_code == 201
        p_lesson_id = p_lesson_res.json()["id"]

        comp_p = client.post(f"/api/v1/lessons/{p_lesson_id}/complete", headers=teacher["headers"])
        assert comp_p.status_code == 200

        # Баланс pair стал 4.0 - 1.0 = 3.0, individual остался 8.5
        list_sub_after = client.get(
            f"/api/v1/clients/{client_data['id']}/subscriptions",
            headers=teacher["headers"],
        ).json()
        target_pair_after = next((s for s in list_sub_after if s["id"] == sub2_id), None)
        assert abs(target_pair_after["balance"] - 3.0) < 1e-4

    def test_patch_lesson_reschedule_and_edit(
        self,
        client: httpx.Client,
        teacher_with_client,
        setup_classroom,
    ):
        """Редактирование урока (PATCH /api/v1/lessons/{id}): перенос времени, смена кабинета и заметок."""
        teacher, client_data = teacher_with_client
        classroom = setup_classroom

        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=5)
        orig_start = to_rfc3339(now.replace(hour=10, minute=0, second=0))
        orig_end = to_rfc3339(now.replace(hour=11, minute=0, second=0))

        create_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client_data["id"],
                "title": "Геометрия: Треугольники",
                "start_time": orig_start,
                "end_time": orig_end,
                "format": "individual",
                "notes": "Исходная заметка",
            },
            headers=teacher["headers"],
        )
        assert create_res.status_code == 201
        created_lesson = create_res.json()
        assert created_lesson["title"] == "Геометрия: Треугольники"
        assert created_lesson["notes"] == "Исходная заметка"
        lesson_id = created_lesson["id"]

        # 1. Редактируем урок: переносим на 14:00 - 15:30, прикрепляем кабинет и обновляем заметку и тему раздельно
        new_start = to_rfc3339(now.replace(hour=14, minute=0, second=0))
        new_end = to_rfc3339(now.replace(hour=15, minute=30, second=0))

        patch_res = client.patch(
            f"/api/v1/lessons/{lesson_id}",
            json={
                "title": "Стереометрия 11 класс",
                "start_time": new_start,
                "end_time": new_end,
                "classroom_id": classroom["id"],
                "notes": "Перенесенный урок по стереометрии",
            },
            headers=teacher["headers"],
        )
        assert patch_res.status_code == 200, f"PATCH lesson failed: {patch_res.text}"
        updated = patch_res.json()

        assert updated["id"] == lesson_id
        assert updated["title"] == "Стереометрия 11 класс"
        assert updated["start_time"] == new_start
        assert updated["end_time"] == new_end
        assert updated["classroom_id"] == classroom["id"]
        assert updated["notes"] == "Перенесенный урок по стереометрии"
        assert updated["status"] == "scheduled"

        # 2. Переключаем урок в онлайн (указываем ссылку без classroom_id) -> кабинет должен очиститься
        patch_online = client.patch(
            f"/api/v1/lessons/{lesson_id}",
            json={
                "location_or_url": "https://meet.google.com/xyz-abc",
            },
            headers=teacher["headers"],
        )
        assert patch_online.status_code == 200, f"PATCH online failed: {patch_online.text}"
        online_data = patch_online.json()
        assert online_data.get("classroom_id") is None
        assert online_data["location_or_url"] == "https://meet.google.com/xyz-abc"

        # 3. Переключаем обратно в оффлайн (привязываем кабинет, location_or_url сбрасывается)
        patch_back = client.patch(
            f"/api/v1/lessons/{lesson_id}",
            json={
                "classroom_id": classroom["id"],
                "location_or_url": "",
            },
            headers=teacher["headers"],
        )
        assert patch_back.status_code == 200
        back_data = patch_back.json()
        assert back_data["classroom_id"] == classroom["id"]
        assert back_data.get("location_or_url") is None or back_data["location_or_url"] == ""

        # 4. Сбрасываем кабинет через classroom_id: null
        patch_null = client.patch(
            f"/api/v1/lessons/{lesson_id}",
            json={"classroom_id": None},
            headers=teacher["headers"],
        )
        assert patch_null.status_code == 200
        assert patch_null.json().get("classroom_id") is None


    def test_lesson_overlap_allowed_for_same_teacher(
        self,
        client: httpx.Client,
        teacher_with_client,
    ):
        """Нахлёст занятий: один и тот же репетитор успешно назначает два урока с частичным пересечением времени."""
        teacher, client1 = teacher_with_client

        # Создаем второго клиента для того же репетитора
        c2_res = client.post(
            "/api/v1/clients",
            json={"name": f"Клиент-2 {uuid.uuid4().hex[:4]}", "rate_individual": 1500.0},
            headers=teacher["headers"],
        )
        assert c2_res.status_code == 201
        client2 = c2_res.json()

        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=4)
        # Урок 1: 15:00 - 16:00 (individual)
        start1 = to_rfc3339(now.replace(hour=15, minute=0, second=0))
        end1 = to_rfc3339(now.replace(hour=16, minute=0, second=0))

        # Урок 2: 15:45 - 16:45 (нахлёст 15:45 - 16:00, pair)
        start2 = to_rfc3339(now.replace(hour=15, minute=45, second=0))
        end2 = to_rfc3339(now.replace(hour=16, minute=45, second=0))

        res1 = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client1["id"],
                "start_time": start1,
                "end_time": end1,
                "format": "individual",
            },
            headers=teacher["headers"],
        )
        assert res1.status_code == 201, f"Failed lesson 1: {res1.text}"

        res2 = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client2["id"],
                "start_time": start2,
                "end_time": end2,
                "format": "pair",
            },
            headers=teacher["headers"],
        )
        assert res2.status_code == 201, f"Failed lesson 2 overlap: {res2.text}"

    def test_classroom_collision_conflict_409(
        self,
        client: httpx.Client,
        teacher_with_client,
        registered_user,
        setup_classroom,
    ):
        """Коллизия кабинета: попытка второго преподавателя занять тот же кабинет в то же время -> 409 Conflict."""
        teacher1, client1 = teacher_with_client
        classroom = setup_classroom

        # Второй преподаватель и его клиент
        _, t2_reg = registered_user(role="teacher")
        t2_headers = {"Authorization": f"Bearer {t2_reg['tokens']['access_token']}"}
        c2_res = client.post(
            "/api/v1/clients",
            json={"name": "Клиент Учителя 2", "rate_individual": 1800.0},
            headers=t2_headers,
        )
        assert c2_res.status_code == 201
        client2 = c2_res.json()

        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=6)
        # Учитель 1 бронирует кабинет на 17:00 - 18:00
        start1 = to_rfc3339(now.replace(hour=17, minute=0, second=0))
        end1 = to_rfc3339(now.replace(hour=18, minute=0, second=0))

        book1_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client1["id"],
                "classroom_id": classroom["id"],
                "start_time": start1,
                "end_time": end1,
                "format": "individual",
            },
            headers=teacher1["headers"],
        )
        assert book1_res.status_code == 201

        # Учитель 2 пытается занять этот же кабинет на 17:30 - 18:30 (коллизия)
        start2 = to_rfc3339(now.replace(hour=17, minute=30, second=0))
        end2 = to_rfc3339(now.replace(hour=18, minute=30, second=0))

        collision_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client2["id"],
                "classroom_id": classroom["id"],
                "start_time": start2,
                "end_time": end2,
                "format": "individual",
            },
            headers=t2_headers,
        )
        assert collision_res.status_code == 409, f"Expected 409 collision, got {collision_res.status_code}: {collision_res.text}"

    def test_lesson_cancel_workflow(
        self,
        client: httpx.Client,
        teacher_with_client,
    ):
        """Отмена запланированного урока преподавателем с указанием причины через POST /api/v1/lessons/{id}/cancel."""
        teacher, client_data = teacher_with_client

        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=7)
        start_time = to_rfc3339(now.replace(hour=12, minute=0, second=0))
        end_time = to_rfc3339(now.replace(hour=13, minute=0, second=0))

        create_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client_data["id"],
                "start_time": start_time,
                "end_time": end_time,
                "format": "individual",
            },
            headers=teacher["headers"],
        )
        assert create_res.status_code == 201
        lesson_id = create_res.json()["id"]

        # Отмена урока
        cancel_res = client.post(
            f"/api/v1/lessons/{lesson_id}/cancel",
            json={"reason": "Болезнь ученика"},
            headers=teacher["headers"],
        )
        assert cancel_res.status_code == 200
        cancelled_lesson = cancel_res.json()
        assert cancelled_lesson["status"] == "cancelled"
        assert cancelled_lesson["cancel_reason"] == "Болезнь ученика"
