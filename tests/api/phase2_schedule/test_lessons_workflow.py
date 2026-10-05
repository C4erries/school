import uuid
from datetime import datetime, timedelta, timezone
import pytest
import httpx


def to_rfc3339(dt: datetime) -> str:
    """Форматирует datetime в ISO/RFC3339 строку с суффиксом Z."""
    return dt.strftime("%Y-%m-%dT%H:%M:%SZ")


@pytest.mark.schedule
class TestLessonsWorkflow:
    """Интеграционные тесты жизненного цикла уроков, абонементов, нахлестов и коллизий кабинетов."""

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
        """Фикстура создания клиента для преподавателя."""
        unique_name = f"Клиент-{uuid.uuid4().hex[:6]}"
        res = client.post(
            "/api/v1/clients",
            json={
                "name": unique_name,
                "phone": "+79997654321",
                "base_rate": 1600.0,
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
        """Создание урока преподавателем: статус сразу scheduled (ADR 005, без accept/decline)."""
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
            "format": "offline",
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
        assert lesson["format"] == "offline"

    def test_lesson_complete_deducts_lessons_subscription(
        self,
        client: httpx.Client,
        teacher_with_client,
    ):
        """Завершение урока списывает 1 занятие с абонемента ученика типа 'lessons'."""
        teacher, client_data = teacher_with_client

        # 1. Добавляем абонемент на 5 занятий
        sub_res = client.post(
            f"/api/v1/clients/{client_data['id']}/subscriptions",
            json={"type": "lessons", "balance": 5.0},
            headers=teacher["headers"],
        )
        assert sub_res.status_code == 201
        sub_id = sub_res.json()["id"]

        # 2. Создаем онлайн урок
        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=1)
        start_time = to_rfc3339(now.replace(hour=14, minute=0, second=0))
        end_time = to_rfc3339(now.replace(hour=15, minute=0, second=0))

        lesson_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client_data["id"],
                "start_time": start_time,
                "end_time": end_time,
                "format": "online",
                "notes": "Онлайн занятие",
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

        # 4. Проверяем баланс абонемента (должен уменьшиться с 5 до 4)
        list_sub = client.get(
            f"/api/v1/clients/{client_data['id']}/subscriptions",
            headers=teacher["headers"],
        ).json()
        target_sub = next((s for s in list_sub if s["id"] == sub_id), None)
        assert target_sub is not None
        assert target_sub["balance"] == 4.0

    def test_lesson_complete_deducts_hours_subscription(
        self,
        client: httpx.Client,
        teacher_with_client,
    ):
        """Завершение урока списывает точную длительность (часы) с абонемента ученика типа 'hours'."""
        teacher, client_data = teacher_with_client

        # 1. Добавляем абонемент на 10.0 часов
        sub_res = client.post(
            f"/api/v1/clients/{client_data['id']}/subscriptions",
            json={"type": "hours", "balance": 10.0},
            headers=teacher["headers"],
        )
        assert sub_res.status_code == 201
        sub_id = sub_res.json()["id"]

        # 2. Создаем онлайн урок длительностью 1.5 часа (10:00 - 11:30)
        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=3)
        start_time = to_rfc3339(now.replace(hour=10, minute=0, second=0))
        end_time = to_rfc3339(now.replace(hour=11, minute=30, second=0))

        lesson_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client_data["id"],
                "start_time": start_time,
                "end_time": end_time,
                "format": "online",
            },
            headers=teacher["headers"],
        )
        assert lesson_res.status_code == 201
        lesson_id = lesson_res.json()["id"]

        # 3. Завершаем урок
        comp_res = client.post(
            f"/api/v1/lessons/{lesson_id}/complete",
            headers=teacher["headers"],
        )
        assert comp_res.status_code == 200
        assert comp_res.json()["status"] == "completed"

        # 4. Проверяем баланс абонемента (должен стать 10.0 - 1.5 = 8.5)
        list_sub = client.get(
            f"/api/v1/clients/{client_data['id']}/subscriptions",
            headers=teacher["headers"],
        ).json()
        target_sub = next((s for s in list_sub if s["id"] == sub_id), None)
        assert target_sub is not None
        assert abs(target_sub["balance"] - 8.5) < 1e-4

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
            json={"name": f"Клиент-2 {uuid.uuid4().hex[:4]}", "base_rate": 1500.0},
            headers=teacher["headers"],
        )
        assert c2_res.status_code == 201
        client2 = c2_res.json()

        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=4)
        # Урок 1: 15:00 - 16:00
        start1 = to_rfc3339(now.replace(hour=15, minute=0, second=0))
        end1 = to_rfc3339(now.replace(hour=16, minute=0, second=0))

        # Урок 2: 15:45 - 16:45 (нахлёст 15:45 - 16:00)
        start2 = to_rfc3339(now.replace(hour=15, minute=45, second=0))
        end2 = to_rfc3339(now.replace(hour=16, minute=45, second=0))

        res1 = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client1["id"],
                "start_time": start1,
                "end_time": end1,
                "format": "online",
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
                "format": "online",
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
            json={"name": "Клиент Учителя 2", "base_rate": 1800.0},
            headers=t2_headers,
        )
        assert c2_res.status_code == 201
        client2 = c2_res.json()

        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=5)
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
                "format": "offline",
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
                "format": "offline",
            },
            headers=t2_headers,
        )
        assert collision_res.status_code == 409, f"Expected 409 collision, got {collision_res.status_code}: {collision_res.text}"

    def test_lesson_cancel_workflow(
        self,
        client: httpx.Client,
        teacher_with_client,
    ):
        """Отмена запланированного урока преподавателем с указанием причины."""
        teacher, client_data = teacher_with_client

        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=6)
        start_time = to_rfc3339(now.replace(hour=12, minute=0, second=0))
        end_time = to_rfc3339(now.replace(hour=13, minute=0, second=0))

        create_res = client.post(
            "/api/v1/lessons",
            json={
                "client_id": client_data["id"],
                "start_time": start_time,
                "end_time": end_time,
                "format": "online",
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
