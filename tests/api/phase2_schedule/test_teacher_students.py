import pytest
import httpx


@pytest.mark.schedule
class TestTeacherStudents:
    """Интеграционные тесты прикрепления учеников к преподавателям."""

    def test_assign_student_by_admin_success(
        self,
        client: httpx.Client,
        admin_user,
        teacher_user,
        student_user,
    ):
        """Прикрепление ученика к преподавателю администратором (201 Created)."""
        payload = {
            "teacher_id": teacher_user["id"],
            "student_id": student_user["id"],
        }

        response = client.post(
            "/api/v1/teachers/students",
            json=payload,
            headers=admin_user["headers"],
        )
        assert response.status_code == 201
        data = response.json()

        assert "id" in data
        assert data["teacher_id"] == teacher_user["id"]
        assert data["student_id"] == student_user["id"]
        assert "created_at" in data

    def test_teacher_lists_assigned_students(
        self,
        client: httpx.Client,
        admin_user,
        teacher_user,
        student_user,
    ):
        """Преподаватель запрашивает список закрепленных за ним учеников (200 OK)."""
        # Сначала привязываем ученика через администратора
        assign_res = client.post(
            "/api/v1/teachers/students",
            json={"teacher_id": teacher_user["id"], "student_id": student_user["id"]},
            headers=admin_user["headers"],
        )
        assert assign_res.status_code == 201

        # Преподаватель запрашивает своих учеников
        list_res = client.get(
            "/api/v1/teachers/students",
            headers=teacher_user["headers"],
        )
        assert list_res.status_code == 200
        students = list_res.json()

        assert isinstance(students, list)
        matching = [s for s in students if s["id"] == student_user["id"]]
        assert len(matching) == 1
        assert matching[0]["email"] == student_user["payload"]["email"]
        assert matching[0]["full_name"] == student_user["payload"]["full_name"]
        assert matching[0]["role"] == "student"

    def test_student_cannot_assign_student_forbidden(
        self,
        client: httpx.Client,
        student_user,
        teacher_user,
        registered_user,
    ):
        """Защита: студент пытается привязать ученика -> 403 Forbidden."""
        _, other_student = registered_user(role="student")
        other_student_id = other_student["user"]["id"]

        payload = {
            "teacher_id": teacher_user["id"],
            "student_id": other_student_id,
        }

        response = client.post(
            "/api/v1/teachers/students",
            json=payload,
            headers=student_user["headers"],
        )
        assert response.status_code == 403
        data = response.json()
        assert data["error"]["code"] == "FORBIDDEN"

    def test_teacher_cannot_assign_student_forbidden(
        self,
        client: httpx.Client,
        teacher_user,
        student_user,
    ):
        """Защита: преподаватель также не имеет прав назначать привязку (только admin/owner) -> 403 Forbidden."""
        payload = {
            "teacher_id": teacher_user["id"],
            "student_id": student_user["id"],
        }
        response = client.post(
            "/api/v1/teachers/students",
            json=payload,
            headers=teacher_user["headers"],
        )
        assert response.status_code == 403

    def test_student_cannot_list_teacher_students_forbidden(
        self,
        client: httpx.Client,
        student_user,
    ):
        """Защита: ученик не может просматривать список закрепленных учеников -> 403 Forbidden."""
        response = client.get(
            "/api/v1/teachers/students",
            headers=student_user["headers"],
        )
        assert response.status_code == 403

    def test_assign_duplicate_student_conflict(
        self,
        client: httpx.Client,
        admin_user,
        teacher_user,
        student_user,
    ):
        """Повторное прикрепление того же ученика возвращает 409 Conflict."""
        payload = {
            "teacher_id": teacher_user["id"],
            "student_id": student_user["id"],
        }
        res1 = client.post("/api/v1/teachers/students", json=payload, headers=admin_user["headers"])
        assert res1.status_code == 201

        res2 = client.post("/api/v1/teachers/students", json=payload, headers=admin_user["headers"])
        assert res2.status_code == 409
        data = res2.json()
        assert data["error"]["code"] == "ALREADY_ASSIGNED"
