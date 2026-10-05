import uuid
import pytest
import httpx


@pytest.mark.schedule
class TestClients:
    """Интеграционные тесты управления клиентами (Clients) и абонементами (Subscriptions)."""

    def test_create_and_get_client_success(self, client: httpx.Client, teacher_user):
        """Создание и получение клиента преподавателем (201 Created, 200 OK)."""
        unique_name = f"Ученик {uuid.uuid4().hex[:6]}"
        payload = {
            "name": unique_name,
            "phone": "+79991234567",
            "base_rate": 1500.0,
            "school_percent_tag": 15,
        }

        # Создание клиента
        response = client.post("/api/v1/clients", json=payload, headers=teacher_user["headers"])
        assert response.status_code == 201, f"Create client failed: {response.text}"
        data = response.json()

        assert "id" in data
        assert data["teacher_id"] == teacher_user["id"]
        assert data["name"] == payload["name"]
        assert data["phone"] == payload["phone"]
        assert data["base_rate"] == payload["base_rate"]
        assert data["school_percent_tag"] == payload["school_percent_tag"]
        assert "created_at" in data
        client_id = data["id"]

        # Получение списка клиентов
        list_res = client.get("/api/v1/clients", headers=teacher_user["headers"])
        assert list_res.status_code == 200, f"List clients failed: {list_res.text}"
        clients = list_res.json()
        assert isinstance(clients, list)

        created_client = next((c for c in clients if c["id"] == client_id), None)
        assert created_client is not None
        assert created_client["name"] == unique_name
        assert created_client["base_rate"] == 1500.0
        assert created_client["school_percent_tag"] == 15

    def test_clients_isolation_between_teachers(
        self, client: httpx.Client, teacher_user, registered_user
    ):
        """Проверка изоляции данных: учитель А видит только своих клиентов, учитель Б — своих."""
        _, other_teacher_reg = registered_user(role="teacher")
        other_token = other_teacher_reg["tokens"]["access_token"]
        other_headers = {"Authorization": f"Bearer {other_token}"}

        # Учитель 1 создает клиента
        res1 = client.post(
            "/api/v1/clients",
            json={"name": f"Клиент Т1 {uuid.uuid4().hex[:4]}", "base_rate": 2000.0},
            headers=teacher_user["headers"],
        )
        assert res1.status_code == 201
        client1_id = res1.json()["id"]

        # Учитель 2 создает клиента
        res2 = client.post(
            "/api/v1/clients",
            json={"name": f"Клиент Т2 {uuid.uuid4().hex[:4]}", "base_rate": 1800.0},
            headers=other_headers,
        )
        assert res2.status_code == 201
        client2_id = res2.json()["id"]

        # Учитель 1 запрашивает список: видит клиента 1, не видит клиента 2
        list1 = client.get("/api/v1/clients", headers=teacher_user["headers"]).json()
        ids1 = [c["id"] for c in list1]
        assert client1_id in ids1
        assert client2_id not in ids1

        # Учитель 2 запрашивает список: видит клиента 2, не видит клиента 1
        list2 = client.get("/api/v1/clients", headers=other_headers).json()
        ids2 = [c["id"] for c in list2]
        assert client2_id in ids2
        assert client1_id not in ids2

    def test_create_client_unauthorized(self, client: httpx.Client):
        """Запрос без токена возвращает 401 Unauthorized."""
        res = client.post(
            "/api/v1/clients",
            json={"name": "Анонимный клиент", "base_rate": 1000.0},
        )
        assert res.status_code == 401

    def test_create_client_validation_error(self, client: httpx.Client, teacher_user):
        """Невалидные данные при создании клиента возвращают 400 Bad Request."""
        # Отсутствует обязательное поле base_rate
        res = client.post(
            "/api/v1/clients",
            json={"name": "Клиент без ставки"},
            headers=teacher_user["headers"],
        )
        assert res.status_code == 400

    def test_create_and_list_subscriptions(self, client: httpx.Client, teacher_user):
        """Создание абонементов (по урокам и часам) и получение списка абонементов клиента."""
        # 1. Создаем клиента
        c_res = client.post(
            "/api/v1/clients",
            json={"name": f"Клиент с абонементом {uuid.uuid4().hex[:4]}", "base_rate": 1200.0},
            headers=teacher_user["headers"],
        )
        assert c_res.status_code == 201
        client_id = c_res.json()["id"]

        # 2. Добавляем абонемент на количество уроков (type="lessons")
        sub1_res = client.post(
            f"/api/v1/clients/{client_id}/subscriptions",
            json={"type": "lessons", "balance": 8.0},
            headers=teacher_user["headers"],
        )
        assert sub1_res.status_code == 201, f"Failed to create lessons subscription: {sub1_res.text}"
        sub1 = sub1_res.json()
        assert "id" in sub1
        assert sub1["client_id"] == client_id
        assert sub1["type"] == "lessons"
        assert sub1["balance"] == 8.0
        assert "created_at" in sub1

        # 3. Добавляем абонемент на количество часов (type="hours")
        sub2_res = client.post(
            f"/api/v1/clients/{client_id}/subscriptions",
            json={"type": "hours", "balance": 12.5},
            headers=teacher_user["headers"],
        )
        assert sub2_res.status_code == 201, f"Failed to create hours subscription: {sub2_res.text}"
        sub2 = sub2_res.json()
        assert sub2["client_id"] == client_id
        assert sub2["type"] == "hours"
        assert sub2["balance"] == 12.5

        # 4. Получаем список абонементов клиента
        list_sub_res = client.get(
            f"/api/v1/clients/{client_id}/subscriptions",
            headers=teacher_user["headers"],
        )
        assert list_sub_res.status_code == 200, f"Failed to list subscriptions: {list_sub_res.text}"
        subs = list_sub_res.json()
        assert isinstance(subs, list)
        sub_ids = [s["id"] for s in subs]
        assert sub1["id"] in sub_ids
        assert sub2["id"] in sub_ids
