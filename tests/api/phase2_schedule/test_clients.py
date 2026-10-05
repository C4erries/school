import uuid
import pytest
import httpx


@pytest.mark.schedule
class TestClients:
    """Интеграционные тесты управления клиентами (Clients), тарифной сеткой, абонементами и динамическими тегами."""

    def test_create_and_get_client_with_rate_grid(self, client: httpx.Client, teacher_user):
        """Создание клиента с тарифной сеткой (rate_individual, rate_pair, rate_group) и проверка GET /api/v1/clients."""
        unique_name = f"Ученик {uuid.uuid4().hex[:6]}"
        payload = {
            "name": unique_name,
            "phone": "+79991234567",
            "rate_individual": 2000.0,
            "rate_pair": 1400.0,
            "rate_group": 1000.0,
            "school_percent_tag": 15,
        }

        # 1. Создание клиента
        response = client.post("/api/v1/clients", json=payload, headers=teacher_user["headers"])
        assert response.status_code == 201, f"Create client failed: {response.text}"
        data = response.json()

        assert "id" in data
        assert data["teacher_id"] == teacher_user["id"]
        assert data["name"] == payload["name"]
        assert data["phone"] == payload["phone"]
        assert data["rate_individual"] == 2000.0
        assert data["rate_pair"] == 1400.0
        assert data["rate_group"] == 1000.0
        assert data["base_rate"] == 2000.0
        assert data["school_percent_tag"] == 15
        assert "balances" in data
        assert data["balances"]["individual_hours"] == 0.0
        assert data["balances"]["pair_hours"] == 0.0
        assert data["balances"]["group_hours"] == 0.0
        assert data["balances"]["total_hours"] == 0.0
        assert data["tags"] == []
        assert "created_at" in data
        client_id = data["id"]

        # 2. Получение списка клиентов
        list_res = client.get("/api/v1/clients", headers=teacher_user["headers"])
        assert list_res.status_code == 200, f"List clients failed: {list_res.text}"
        clients = list_res.json()
        assert isinstance(clients, list)

        created_client = next((c for c in clients if c["id"] == client_id), None)
        assert created_client is not None
        assert created_client["name"] == unique_name
        assert created_client["rate_individual"] == 2000.0
        assert created_client["rate_pair"] == 1400.0
        assert created_client["rate_group"] == 1000.0
        assert created_client["balances"]["total_hours"] == 0.0

    def test_client_subscriptions_by_format_and_aggregated_balances(
        self, client: httpx.Client, teacher_user
    ):
        """Покупка абонементов по форматам (individual, pair, group) и проверка агрегации балансов в GET /api/v1/clients."""
        # 1. Создаем клиента
        c_res = client.post(
            "/api/v1/clients",
            json={
                "name": f"Клиент с абонементами {uuid.uuid4().hex[:4]}",
                "rate_individual": 1800.0,
                "rate_pair": 1200.0,
                "rate_group": 900.0,
            },
            headers=teacher_user["headers"],
        )
        assert c_res.status_code == 201
        client_id = c_res.json()["id"]

        # 2. Добавляем абонемент на индивидуальные часы (10.0 ч)
        sub_ind = client.post(
            f"/api/v1/clients/{client_id}/subscriptions",
            json={"format": "individual", "balance": 10.0},
            headers=teacher_user["headers"],
        )
        assert sub_ind.status_code == 201, f"Failed individual sub: {sub_ind.text}"
        ind_data = sub_ind.json()
        assert ind_data["client_id"] == client_id
        assert ind_data["format"] == "individual"
        assert ind_data["balance"] == 10.0

        # 3. Добавляем абонемент на парные часы (5.5 ч)
        sub_pair = client.post(
            f"/api/v1/clients/{client_id}/subscriptions",
            json={"format": "pair", "balance": 5.5},
            headers=teacher_user["headers"],
        )
        assert sub_pair.status_code == 201, f"Failed pair sub: {sub_pair.text}"
        pair_data = sub_pair.json()
        assert pair_data["format"] == "pair"
        assert pair_data["balance"] == 5.5

        # 4. Добавляем абонемент на групповые часы (8.0 ч)
        sub_grp = client.post(
            f"/api/v1/clients/{client_id}/subscriptions",
            json={"format": "group", "balance": 8.0},
            headers=teacher_user["headers"],
        )
        assert sub_grp.status_code == 201, f"Failed group sub: {sub_grp.text}"
        grp_data = sub_grp.json()
        assert grp_data["format"] == "group"
        assert grp_data["balance"] == 8.0

        # 5. Проверяем детальный список абонементов
        list_subs_res = client.get(
            f"/api/v1/clients/{client_id}/subscriptions",
            headers=teacher_user["headers"],
        )
        assert list_subs_res.status_code == 200
        subs = list_subs_res.json()
        assert len(subs) == 3
        formats = [s["format"] for s in subs]
        assert "individual" in formats
        assert "pair" in formats
        assert "group" in formats

        # 6. Проверяем агрегированные балансы в GET /api/v1/clients
        clients_res = client.get("/api/v1/clients", headers=teacher_user["headers"])
        assert clients_res.status_code == 200
        clients = clients_res.json()
        target = next((c for c in clients if c["id"] == client_id), None)
        assert target is not None
        assert abs(target["balances"]["individual_hours"] - 10.0) < 1e-4
        assert abs(target["balances"]["pair_hours"] - 5.5) < 1e-4
        assert abs(target["balances"]["group_hours"] - 8.0) < 1e-4
        assert abs(target["balances"]["total_hours"] - 23.5) < 1e-4

    def test_tags_crud_and_client_binding(self, client: httpx.Client, teacher_user):
        """Создание тегов (POST /api/v1/tags), получение списка, привязка к клиенту и отображение в GET /api/v1/clients."""
        headers = teacher_user["headers"]

        # 1. Создаем два тега
        t1_res = client.post(
            "/api/v1/tags",
            json={"name": f"Школа-Математика {uuid.uuid4().hex[:4]}", "school_percent": 25, "color": "emerald"},
            headers=headers,
        )
        assert t1_res.status_code == 201, f"Failed create tag 1: {t1_res.text}"
        tag1 = t1_res.json()
        assert tag1["teacher_id"] == teacher_user["id"]
        assert tag1["school_percent"] == 25
        assert tag1["color"] == "emerald"

        t2_res = client.post(
            "/api/v1/tags",
            json={"name": f"Олимпиадники {uuid.uuid4().hex[:4]}", "school_percent": 10, "color": "purple"},
            headers=headers,
        )
        assert t2_res.status_code == 201, f"Failed create tag 2: {t2_res.text}"
        tag2 = t2_res.json()

        # 2. Получение списка тегов преподавателя
        list_tags = client.get("/api/v1/tags", headers=headers)
        assert list_tags.status_code == 200
        tag_ids = [t["id"] for t in list_tags.json()]
        assert tag1["id"] in tag_ids
        assert tag2["id"] in tag_ids

        # 3. Создаем клиента сразу с привязанным tag1
        c_res = client.post(
            "/api/v1/clients",
            json={
                "name": f"Ученик с тегом {uuid.uuid4().hex[:4]}",
                "rate_individual": 2200.0,
                "tag_ids": [tag1["id"]],
            },
            headers=headers,
        )
        assert c_res.status_code == 201
        client1 = c_res.json()
        assert len(client1["tags"]) == 1
        assert client1["tags"][0]["id"] == tag1["id"]

        # 4. Привязываем второй тег через POST /api/v1/clients/{id}/tags
        assign_res = client.post(
            f"/api/v1/clients/{client1['id']}/tags",
            json={"tag_id": tag2["id"]},
            headers=headers,
        )
        assert assign_res.status_code == 200
        client1_updated = assign_res.json()
        assert len(client1_updated["tags"]) == 2
        assigned_ids = [t["id"] for t in client1_updated["tags"]]
        assert tag1["id"] in assigned_ids
        assert tag2["id"] in assigned_ids

        # 5. Проверяем выдачу в общем списке клиентов GET /api/v1/clients
        clients_list = client.get("/api/v1/clients", headers=headers).json()
        c_found = next((c for c in clients_list if c["id"] == client1["id"]), None)
        assert c_found is not None
        assert len(c_found["tags"]) == 2

        # 6. Отвязываем tag1 через DELETE /api/v1/clients/{id}/tags/{tag_id}
        remove_res = client.delete(
            f"/api/v1/clients/{client1['id']}/tags/{tag1['id']}",
            headers=headers,
        )
        assert remove_res.status_code == 200
        c_after_remove = remove_res.json()
        assert len(c_after_remove["tags"]) == 1
        assert c_after_remove["tags"][0]["id"] == tag2["id"]

    def test_tags_protection_forbidden_for_student(self, client: httpx.Client, student_user):
        """Защита: ученик не может создавать теги (403 Forbidden)."""
        res = client.post(
            "/api/v1/tags",
            json={"name": "Хакерский тег", "school_percent": 0},
            headers=student_user["headers"],
        )
        assert res.status_code == 403

    def test_clients_isolation_between_teachers(
        self, client: httpx.Client, teacher_user, registered_user
    ):
        """Проверка изоляции данных: учитель А видит только своих клиентов и теги, учитель Б — своих."""
        _, other_teacher_reg = registered_user(role="teacher")
        other_token = other_teacher_reg["tokens"]["access_token"]
        other_headers = {"Authorization": f"Bearer {other_token}"}

        # Учитель 1 создает клиента
        res1 = client.post(
            "/api/v1/clients",
            json={"name": f"Клиент Т1 {uuid.uuid4().hex[:4]}", "rate_individual": 2000.0},
            headers=teacher_user["headers"],
        )
        assert res1.status_code == 201
        client1_id = res1.json()["id"]

        # Учитель 2 создает клиента
        res2 = client.post(
            "/api/v1/clients",
            json={"name": f"Клиент Т2 {uuid.uuid4().hex[:4]}", "rate_individual": 1800.0},
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
            json={"name": "Анонимный клиент", "rate_individual": 1000.0},
        )
        assert res.status_code == 401

    def test_create_client_validation_error(self, client: httpx.Client, teacher_user):
        """Невалидные данные при создании клиента возвращают 400 Bad Request."""
        # Отсутствует обязательное поле rate_individual / base_rate
        res = client.post(
            "/api/v1/clients",
            json={"name": "Клиент без ставки"},
            headers=teacher_user["headers"],
        )
        assert res.status_code == 400

    def test_update_client(self, client: httpx.Client, teacher_user):
        """Редактирование клиента (PATCH /api/v1/clients/{id}), проверка обновленных данных и негативные кейсы."""
        # 1. Преподаватель (teacher_user) создает клиента через POST /api/v1/clients
        initial_name = f"Клиент до {uuid.uuid4().hex[:4]}"
        create_res = client.post(
            "/api/v1/clients",
            json={
                "name": initial_name,
                "phone": "+79991112233",
                "rate_individual": 1500.0,
                "rate_pair": 1000.0,
                "rate_group": 700.0,
            },
            headers=teacher_user["headers"],
        )
        assert create_res.status_code == 201, f"Failed create client: {create_res.text}"
        client_data = create_res.json()
        client_id = client_data["id"]

        # 2. Создает тег (POST /api/v1/tags)
        tag_res = client.post(
            "/api/v1/tags",
            json={
                "name": f"Тег {uuid.uuid4().hex[:4]}",
                "school_percent": 15,
                "color": "emerald",
            },
            headers=teacher_user["headers"],
        )
        assert tag_res.status_code == 201, f"Failed create tag: {tag_res.text}"
        tag_id = tag_res.json()["id"]

        # 3. Отправляет PATCH /api/v1/clients/{id} с обновленными данными
        updated_name = f"Обновленный {uuid.uuid4().hex[:4]}"
        updated_phone = "+79998887766"
        patch_payload = {
            "name": updated_name,
            "phone": updated_phone,
            "rate_individual": 1800.0,
            "rate_pair": 1200.0,
            "rate_group": 900.0,
            "tag_ids": [tag_id],
        }
        patch_res = client.patch(
            f"/api/v1/clients/{client_id}",
            json=patch_payload,
            headers=teacher_user["headers"],
        )

        # 4. Проверяет: статус 200 OK и поля в теле ответа обновились
        assert patch_res.status_code == 200, f"Failed patch client: {patch_res.text}"
        updated_data = patch_res.json()
        assert updated_data["id"] == client_id
        assert updated_data["name"] == updated_name
        assert updated_data["phone"] == updated_phone
        assert updated_data["rate_individual"] == 1800.0
        assert updated_data["rate_pair"] == 1200.0
        assert updated_data["rate_group"] == 900.0
        assert any(t["id"] == tag_id for t in updated_data.get("tags", []))

        # 5. Делает GET /api/v1/clients и проверяет, что в списке клиентов данные также соответствуют обновленным значениям
        list_res = client.get("/api/v1/clients", headers=teacher_user["headers"])
        assert list_res.status_code == 200, f"Failed list clients: {list_res.text}"
        clients = list_res.json()
        target = next((c for c in clients if c["id"] == client_id), None)
        assert target is not None
        assert target["name"] == updated_name
        assert target["phone"] == updated_phone
        assert target["rate_individual"] == 1800.0
        assert target["rate_pair"] == 1200.0
        assert target["rate_group"] == 900.0
        assert any(t["id"] == tag_id for t in target.get("tags", []))

        # 6. Проверяет негативные кейсы:
        # PATCH несуществующего client_id -> 404
        fake_id = str(uuid.uuid4())
        res_404 = client.patch(
            f"/api/v1/clients/{fake_id}",
            json={"name": "Не существует"},
            headers=teacher_user["headers"],
        )
        assert res_404.status_code == 404, f"Expected 404, got {res_404.status_code}: {res_404.text}"

        # PATCH с отрицательной ставкой -> 400
        res_400 = client.patch(
            f"/api/v1/clients/{client_id}",
            json={"rate_individual": -500.0},
            headers=teacher_user["headers"],
        )
        assert res_400.status_code == 400, f"Expected 400, got {res_400.status_code}: {res_400.text}"

        # Запрос от неавторизованного пользователя -> 401
        res_401 = client.patch(
            f"/api/v1/clients/{client_id}",
            json={"name": "Аноним"},
        )
        assert res_401.status_code == 401, f"Expected 401, got {res_401.status_code}: {res_401.text}"
