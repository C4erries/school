import uuid
import pytest
import httpx


@pytest.mark.schedule
class TestClassrooms:
    """Интеграционные тесты управления учебными кабинетами (Classrooms)."""

    def test_create_classroom_by_admin_success(self, client: httpx.Client, admin_user):
        """Создание кабинета администратором (201 Created)."""
        unique_name = f"Кабинет №{uuid.uuid4().hex[:6]}"
        payload = {
            "name": unique_name,
            "capacity": 5,
            "color": "#10B981",
            "description": "Оснащен интерактивной доской",
        }

        response = client.post("/api/v1/classrooms", json=payload, headers=admin_user["headers"])
        assert response.status_code == 201
        data = response.json()

        assert "id" in data
        assert data["name"] == payload["name"]
        assert data["capacity"] == payload["capacity"]
        assert data["color"] == payload["color"]
        assert data["description"] == payload["description"]
        assert "created_at" in data

    @pytest.mark.parametrize("user_fixture", ["teacher_user", "student_user"])
    def test_create_classroom_forbidden_for_non_admin(self, client: httpx.Client, request, user_fixture):
        """Защита: не-администратор (преподаватель или ученик) не может создавать кабинеты (403 Forbidden)."""
        user = request.getfixturevalue(user_fixture)
        payload = {
            "name": f"Неавторизованный кабинет {uuid.uuid4().hex[:4]}",
            "capacity": 3,
            "color": "#EF4444",
        }

        response = client.post("/api/v1/classrooms", json=payload, headers=user["headers"])
        assert response.status_code == 403
        data = response.json()
        assert "error" in data
        assert data["error"]["code"] == "FORBIDDEN"

    def test_create_classroom_unauthorized_without_token(self, client: httpx.Client):
        """Запрос на создание кабинета без токена возвращает 401 Unauthorized."""
        payload = {
            "name": "Анонимный кабинет",
            "capacity": 2,
        }
        response = client.post("/api/v1/classrooms", json=payload)
        assert response.status_code == 401

    def test_create_classroom_validation_errors(self, client: httpx.Client, admin_user):
        """Валидация при создании кабинета: недопустимая вместимость и пустое имя -> 400 Bad Request."""
        # 1. Вместимость <= 0
        res1 = client.post(
            "/api/v1/classrooms",
            json={"name": "Кабинет с 0 мест", "capacity": 0},
            headers=admin_user["headers"],
        )
        assert res1.status_code == 400

        # 2. Слишком короткое или пустое имя
        res2 = client.post(
            "/api/v1/classrooms",
            json={"name": "A", "capacity": 4},
            headers=admin_user["headers"],
        )
        assert res2.status_code == 400

    def test_get_classrooms_list_success(self, client: httpx.Client, admin_user, student_user):
        """Получение списка кабинетов (200 OK) авторизованными пользователями."""
        # Создаем тестовый кабинет
        unique_name = f"Кабинет-Тест {uuid.uuid4().hex[:6]}"
        create_res = client.post(
            "/api/v1/classrooms",
            json={"name": unique_name, "capacity": 6, "color": "#6366F1"},
            headers=admin_user["headers"],
        )
        assert create_res.status_code == 201
        created_id = create_res.json()["id"]

        # Запрашиваем список кабинетов от имени студента
        list_res = client.get("/api/v1/classrooms", headers=student_user["headers"])
        assert list_res.status_code == 200
        classrooms = list_res.json()

        assert isinstance(classrooms, list)
        matching = [c for c in classrooms if c["id"] == created_id]
        assert len(matching) == 1
        assert matching[0]["name"] == unique_name
        assert matching[0]["capacity"] == 6
