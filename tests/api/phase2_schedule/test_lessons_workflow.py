import uuid
from datetime import datetime, timedelta, timezone
import pytest
import httpx


def to_rfc3339(dt: datetime) -> str:
    """Форматирует datetime в ISO/RFC3339 строку с суффиксом Z."""
    return dt.strftime("%Y-%m-%dT%H:%M:%SZ")


@pytest.mark.schedule
class TestLessonsWorkflow:
    """Интеграционные тесты жизненного цикла уроков, согласования, нахлестов и коллизий кабинетов."""

    @pytest.fixture
    def setup_classroom(self, client: httpx.Client, admin_user):
        """Фикстура создания тестового кабинета администратором."""
        unique_name = f"Кабинет-Тест {uuid.uuid4().hex[:6]}"
        res = client.post(
            "/api/v1/classrooms",
            json={"name": unique_name, "capacity": 4, "color": "#3B82F6"},
            headers=admin_user["headers"],
        )
        assert res.status_code == 201
        return res.json()

    @pytest.fixture
    def teacher_with_student(
        self,
        client: httpx.Client,
        admin_user,
        teacher_user,
        student_user,
    ):
        """Фикстура связки преподаватель <-> закрепленный ученик."""
        assign_res = client.post(
            "/api/v1/teachers/students",
            json={"teacher_id": teacher_user["id"], "student_id": student_user["id"]},
            headers=admin_user["headers"],
        )
        assert assign_res.status_code == 201
        return teacher_user, student_user

    def test_lesson_full_lifecycle_accept_and_complete(
        self,
        client: httpx.Client,
        teacher_with_student,
        setup_classroom,
    ):
        """Полный жизненный цикл урока:

        1. Назначение преподавателем -> pending_confirmation (201)
        2. Принятие учеником -> confirmed (200)
        3. Завершение преподавателем -> completed (200)
        """
        teacher, student = teacher_with_student
        classroom = setup_classroom

        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=1)
        start_time = to_rfc3339(now.replace(hour=10, minute=0, second=0))
        end_time = to_rfc3339(now.replace(hour=11, minute=0, second=0))

        # 1. Преподаватель назначает оффлайн-урок
        create_payload = {
            "student_id": student["id"],
            "classroom_id": classroom["id"],
            "start_time": start_time,
            "end_time": end_time,
            "format": "offline",
            "notes": "Урок по геометрии: теорема Пифагора",
        }
        create_res = client.post(
            "/api/v1/lessons",
            json=create_payload,
            headers=teacher["headers"],
        )
        assert create_res.status_code == 201, f"Failed to create lesson: {create_res.text}"
        lesson = create_res.json()

        assert lesson["status"] == "pending_confirmation"
        assert lesson["teacher_id"] == teacher["id"]
        assert lesson["student_id"] == student["id"]
        assert lesson["classroom_id"] == classroom["id"]
        assert lesson["format"] == "offline"
        lesson_id = lesson["id"]

        # 2. Ученик принимает урок
        accept_res = client.post(
            f"/api/v1/lessons/{lesson_id}/accept",
            headers=student["headers"],
        )
        assert accept_res.status_code == 200, f"Failed to accept lesson: {accept_res.text}"
        accepted_lesson = accept_res.json()
        assert accepted_lesson["id"] == lesson_id
        assert accepted_lesson["status"] == "confirmed"

        # 3. Преподаватель завершает урок
        complete_res = client.post(
            f"/api/v1/lessons/{lesson_id}/complete",
            headers=teacher["headers"],
        )
        assert complete_res.status_code == 200, f"Failed to complete lesson: {complete_res.text}"
        completed_lesson = complete_res.json()
        assert completed_lesson["id"] == lesson_id
        assert completed_lesson["status"] == "completed"

    def test_lesson_decline_workflow(
        self,
        client: httpx.Client,
        teacher_with_student,
    ):
        """Отклонение урока учеником с указанием причины -> declined (200 OK)."""
        teacher, student = teacher_with_student

        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=2)
        start_time = to_rfc3339(now.replace(hour=14, minute=0, second=0))
        end_time = to_rfc3339(now.replace(hour=15, minute=0, second=0))

        # Преподаватель назначает онлайн-урок
        create_payload = {
            "student_id": student["id"],
            "start_time": start_time,
            "end_time": end_time,
            "format": "online",
            "location_or_url": "https://meet.google.com/test-room",
            "notes": "Подготовка к ОГЭ",
        }
        create_res = client.post(
            "/api/v1/lessons",
            json=create_payload,
            headers=teacher["headers"],
        )
        assert create_res.status_code == 201
        lesson_id = create_res.json()["id"]

        # Ученик отклоняет урок с указанием причины
        decline_reason = "Не успеваю вернуться из школы к 14:00"
        decline_res = client.post(
            f"/api/v1/lessons/{lesson_id}/decline",
            json={"reason": decline_reason},
            headers=student["headers"],
        )
        assert decline_res.status_code == 200, f"Failed to decline lesson: {decline_res.text}"
        declined_lesson = decline_res.json()
        assert declined_lesson["id"] == lesson_id
        assert declined_lesson["status"] == "declined"
        assert declined_lesson["cancel_reason"] == decline_reason

        # Повторная попытка подтвердить отклоненный урок должна возвращать 400 Bad Request
        retry_accept = client.post(
            f"/api/v1/lessons/{lesson_id}/accept",
            headers=student["headers"],
        )
        assert retry_accept.status_code == 400

    def test_same_teacher_overlapping_lessons_allowed(
        self,
        client: httpx.Client,
        admin_user,
        teacher_user,
        registered_user,
        setup_classroom,
    ):
        """Нахлёст занятий у одного преподавателя:

        Один и тот же преподаватель успешно назначает два урока с частичным
        пересечением времени (например, 15:00-16:00 и 15:45-16:45) в одном кабинете
        (разрешено по бизнес-требованиям, например, мини-группа или совмещенный слот).
        """
        classroom = setup_classroom

        # Регистрируем двух разных студентов и прикрепляем их к одному преподавателю
        _, student1 = registered_user(role="student")
        _, student2 = registered_user(role="student")
        s1_id = student1["user"]["id"]
        s2_id = student2["user"]["id"]

        client.post(
            "/api/v1/teachers/students",
            json={"teacher_id": teacher_user["id"], "student_id": s1_id},
            headers=admin_user["headers"],
        )
        client.post(
            "/api/v1/teachers/students",
            json={"teacher_id": teacher_user["id"], "student_id": s2_id},
            headers=admin_user["headers"],
        )

        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=3)
        start_1 = to_rfc3339(now.replace(hour=15, minute=0, second=0))
        end_1 = to_rfc3339(now.replace(hour=16, minute=0, second=0))

        start_2 = to_rfc3339(now.replace(hour=15, minute=45, second=0))
        end_2 = to_rfc3339(now.replace(hour=16, minute=45, second=0))

        # Урок 1: 15:00 - 16:00
        res1 = client.post(
            "/api/v1/lessons",
            json={
                "student_id": s1_id,
                "classroom_id": classroom["id"],
                "start_time": start_1,
                "end_time": end_1,
                "format": "offline",
            },
            headers=teacher_user["headers"],
        )
        assert res1.status_code == 201, f"Failed lesson 1: {res1.text}"
        data1 = res1.json()
        assert data1["status"] == "pending_confirmation"

        # Урок 2: 15:45 - 16:45 (частичный нахлёст)
        res2 = client.post(
            "/api/v1/lessons",
            json={
                "student_id": s2_id,
                "classroom_id": classroom["id"],
                "start_time": start_2,
                "end_time": end_2,
                "format": "offline",
            },
            headers=teacher_user["headers"],
        )
        assert res2.status_code == 201, f"Failed lesson 2 (same teacher overlap): {res2.text}"
        data2 = res2.json()
        assert data2["status"] == "pending_confirmation"
        assert data2["id"] != data1["id"]

    def test_different_teachers_classroom_collision_conflict(
        self,
        client: httpx.Client,
        admin_user,
        teacher_user,
        registered_user,
        setup_classroom,
    ):
        """Коллизия кабинета:

        Второй преподаватель пытается назначить оффлайн-урок в тот же кабинет
        в то же время (пересекающийся интервал) -> получает 409 Conflict.
        """
        classroom = setup_classroom

        # Преподаватель 1 и студент 1
        _, s1 = registered_user(role="student")
        s1_id = s1["user"]["id"]
        client.post(
            "/api/v1/teachers/students",
            json={"teacher_id": teacher_user["id"], "student_id": s1_id},
            headers=admin_user["headers"],
        )

        # Преподаватель 2 и студент 2
        _, t2 = registered_user(role="teacher")
        t2_id = t2["user"]["id"]
        t2_token = t2["tokens"]["access_token"]
        t2_headers = {"Authorization": f"Bearer {t2_token}"}

        _, s2 = registered_user(role="student")
        s2_id = s2["user"]["id"]
        client.post(
            "/api/v1/teachers/students",
            json={"teacher_id": t2_id, "student_id": s2_id},
            headers=admin_user["headers"],
        )

        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=4)
        start_1 = to_rfc3339(now.replace(hour=12, minute=0, second=0))
        end_1 = to_rfc3339(now.replace(hour=13, minute=0, second=0))

        start_2 = to_rfc3339(now.replace(hour=12, minute=30, second=0))
        end_2 = to_rfc3339(now.replace(hour=13, minute=30, second=0))

        # Преподаватель 1 бронирует кабинет 12:00 - 13:00 -> 201 Created
        res1 = client.post(
            "/api/v1/lessons",
            json={
                "student_id": s1_id,
                "classroom_id": classroom["id"],
                "start_time": start_1,
                "end_time": end_1,
                "format": "offline",
            },
            headers=teacher_user["headers"],
        )
        assert res1.status_code == 201

        # Преподаватель 2 пытается занять тот же кабинет на 12:30 - 13:30 -> 409 Conflict
        res2 = client.post(
            "/api/v1/lessons",
            json={
                "student_id": s2_id,
                "classroom_id": classroom["id"],
                "start_time": start_2,
                "end_time": end_2,
                "format": "offline",
            },
            headers=t2_headers,
        )
        assert res2.status_code == 409
        error = res2.json()
        assert error["error"]["code"] == "CLASSROOM_COLLISION"

    def test_student_cannot_create_lesson_forbidden(
        self,
        client: httpx.Client,
        student_user,
        setup_classroom,
    ):
        """Защита: ученик не может назначать уроки (403 Forbidden)."""
        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=5)
        payload = {
            "student_id": student_user["id"],
            "classroom_id": setup_classroom["id"],
            "start_time": to_rfc3339(now.replace(hour=10, minute=0, second=0)),
            "end_time": to_rfc3339(now.replace(hour=11, minute=0, second=0)),
            "format": "offline",
        }
        res = client.post("/api/v1/lessons", json=payload, headers=student_user["headers"])
        assert res.status_code == 403

    def test_cannot_create_lesson_for_unassigned_student(
        self,
        client: httpx.Client,
        teacher_user,
        registered_user,
    ):
        """Попытка назначить урок не прикрепленному ученику возвращает 400 Bad Request."""
        _, other_student = registered_user(role="student")
        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=5)
        payload = {
            "student_id": other_student["user"]["id"],
            "start_time": to_rfc3339(now.replace(hour=10, minute=0, second=0)),
            "end_time": to_rfc3339(now.replace(hour=11, minute=0, second=0)),
            "format": "online",
        }
        res = client.post("/api/v1/lessons", json=payload, headers=teacher_user["headers"])
        assert res.status_code == 400
        data = res.json()
        assert data["error"]["code"] == "STUDENT_NOT_ASSIGNED"

    def test_other_student_cannot_accept_lesson_forbidden(
        self,
        client: httpx.Client,
        teacher_with_student,
        registered_user,
    ):
        """Защита: посторонний ученик не может принять чужой урок -> 403 Forbidden."""
        teacher, student = teacher_with_student
        _, other_student = registered_user(role="student")
        other_headers = {"Authorization": f"Bearer {other_student['tokens']['access_token']}"}

        now = datetime.now(timezone.utc).replace(microsecond=0) + timedelta(days=6)
        create_res = client.post(
            "/api/v1/lessons",
            json={
                "student_id": student["id"],
                "start_time": to_rfc3339(now.replace(hour=16, minute=0, second=0)),
                "end_time": to_rfc3339(now.replace(hour=17, minute=0, second=0)),
                "format": "online",
            },
            headers=teacher["headers"],
        )
        assert create_res.status_code == 201
        lesson_id = create_res.json()["id"]

        # Посторонний ученик пытается принять
        accept_res = client.post(
            f"/api/v1/lessons/{lesson_id}/accept",
            headers=other_headers,
        )
        assert accept_res.status_code == 403
