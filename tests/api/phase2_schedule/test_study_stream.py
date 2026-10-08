import uuid
from datetime import datetime, timedelta, timezone
import pytest
import httpx


def to_rfc3339(dt: datetime) -> str:
    """Форматирует datetime в ISO/RFC3339 строку с суффиксом Z."""
    return dt.strftime("%Y-%m-%dT%H:%M:%SZ")


@pytest.mark.schedule
class TestStudyStreamAndNotesWorkflow:
    """
    E2E автотесты Спринта 2.4.2 (Полноэкранный Дневник-Мессенджер и Быстрые Заметки):
    - Создание свободной заметки ученика: POST /api/v1/crm/clients/{id}/notes
    - Валидация пустой заметки: 400 Bad Request
    - Удаление свободной заметки: DELETE /api/v1/crm/clients/notes/{id}
    - Получение объединенной ленты Study Stream: GET /api/v1/crm/clients/{id}/stream
    - Проверка сортировки ленты Timestamp ASC (старые сверху, новые снизу)
    - Проверка наличия и структуры upcoming_lesson в ленте
    - Проверка сортировки клиентов по last_lesson_at в CRM
    - Мультиарендная изоляция (чужой преподаватель получает 403 Forbidden / 404 Not Found)
    """

    @pytest.fixture
    def setup_teacher_and_client(self, client: httpx.Client, registered_user):
        """Создает изолированного преподавателя и его клиента."""
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
                "name": f"Ученик Стрима {unique_id}",
                "phone": "+79997771122",
                "rate_individual": 2000.0,
            },
            headers=teacher["headers"],
        )
        assert c_res.status_code == 201, f"Failed to create client: {c_res.text}"
        client_data = c_res.json()

        # Пополняем абонемент
        sub_res = client.post(
            f"/api/v1/clients/{client_data['id']}/subscriptions",
            json={"format": "individual", "balance": 10.0},
            headers=teacher["headers"],
        )
        assert sub_res.status_code == 201, f"Failed to create subscription: {sub_res.text}"

        return teacher, client_data

    def test_client_notes_lifecycle(self, client: httpx.Client, setup_teacher_and_client):
        """Проверяет создание, валидацию и удаление свободной заметки."""
        teacher, client_data = setup_teacher_and_client
        client_id = client_data["id"]

        # 1. Валидация пустой заметки
        empty_res = client.post(
            f"/api/v1/crm/clients/{client_id}/notes",
            json={"content": "   "},
            headers=teacher["headers"],
        )
        assert empty_res.status_code == 400

        # 2. Успешное создание заметки
        create_res = client.post(
            f"/api/v1/crm/clients/{client_id}/notes",
            json={"content": "Мама просила сделать упор на планиметрию"},
            headers=teacher["headers"],
        )
        assert create_res.status_code == 201
        note_data = create_res.json()
        assert note_data["content"] == "Мама просила сделать упор на планиметрию"
        assert note_data["client_id"] == client_id
        assert note_data["teacher_id"] == teacher["id"]
        note_id = note_data["id"]

        # 3. Удаление заметки
        del_res = client.delete(
            f"/api/v1/crm/clients/notes/{note_id}",
            headers=teacher["headers"],
        )
        assert del_res.status_code == 204

        # 4. Повторное удаление возвращает 404
        del_again = client.delete(
            f"/api/v1/crm/clients/notes/{note_id}",
            headers=teacher["headers"],
        )
        assert del_again.status_code == 404

    def test_study_stream_aggregation_and_ordering(self, client: httpx.Client, setup_teacher_and_client):
        """Проверяет объединение отчетов и заметок в единый поток, сортировку ASC и upcoming_lesson."""
        teacher, client_data = setup_teacher_and_client
        client_id = client_data["id"]
        now = datetime.now(timezone.utc)

        # 1. Создаем прошедший физический урок и заполняем отчет
        past_lesson_res = client.post(
            "/api/v1/lessons",
            json={
                "teacher_id": teacher["id"],
                "client_id": client_id,
                "title": "Урок по геометрии",
                "start_time": to_rfc3339(now - timedelta(days=2, hours=2)),
                "end_time": to_rfc3339(now - timedelta(days=2, hours=1)),
                "format": "individual",
            },
            headers=teacher["headers"],
        )
        assert past_lesson_res.status_code == 201, f"Failed past lesson: {past_lesson_res.text}"
        past_lesson_id = past_lesson_res.json()["id"]

        journal_res = client.put(
            f"/api/v1/schedule/lessons/{past_lesson_id}/journal",
            json={
                "topic": "Теорема синусов",
                "notes": "Разобрали примеры из ЕГЭ",
                "performance_score": 5,
            },
            headers=teacher["headers"],
        )
        assert journal_res.status_code == 200

        # 2. Создаем свободную заметку (позже по времени, чем урок)
        note_res = client.post(
            f"/api/v1/crm/clients/{client_id}/notes",
            json={"content": "Ученик быстро освоил материал, дать олимпиадную задачу"},
            headers=teacher["headers"],
        )
        assert note_res.status_code == 201

        # 3. Создаем будущий запланированный урок (upcoming)
        future_start = now + timedelta(days=1, hours=3)
        future_end = future_start + timedelta(hours=1)
        future_lesson_res = client.post(
            "/api/v1/lessons",
            json={
                "teacher_id": teacher["id"],
                "client_id": client_id,
                "title": "Подготовка к контрольной",
                "start_time": to_rfc3339(future_start),
                "end_time": to_rfc3339(future_end),
                "format": "individual",
                "notes": "Повторение тригонометрии",
            },
            headers=teacher["headers"],
        )
        assert future_lesson_res.status_code == 201, f"Failed future lesson: {future_lesson_res.text}"

        # 4. Запрашиваем study stream
        stream_res = client.get(
            f"/api/v1/crm/clients/{client_id}/stream",
            headers=teacher["headers"],
        )
        assert stream_res.status_code == 200
        stream_data = stream_res.json()

        assert stream_data["client_id"] == client_id
        items = stream_data["items"]
        assert len(items) == 2

        # Проверка сортировки: старое сверху (lesson_report), новое снизу (note)
        assert items[0]["type"] == "lesson_report"
        assert items[0]["lesson_report"]["journal"]["topic"] == "Теорема синусов"
        assert items[1]["type"] == "note"
        assert items[1]["note"]["content"] == "Ученик быстро освоил материал, дать олимпиадную задачу"

        # Проверка upcoming_lesson
        upcoming = stream_data.get("upcoming_lesson")
        assert upcoming is not None
        assert upcoming["title"] == "Подготовка к контрольной"
        assert upcoming["topic"] == "Повторение тригонометрии"

    def test_client_last_lesson_at_sorting(self, client: httpx.Client, registered_user):
        """Проверяет корректность расчета поля last_lesson_at у клиентов и сортировки по нему."""
        _, reg_data = registered_user(role="teacher")
        token = reg_data["tokens"]["access_token"]
        headers = {"Authorization": f"Bearer {token}"}
        teacher_id = reg_data["user"]["id"]
        now = datetime.now(timezone.utc)

        # Создаем двух учеников
        c1_res = client.post(
            "/api/v1/clients",
            json={"name": "Ученик 1 (старый урок)", "phone": "+79991112233", "rate_individual": 1500.0},
            headers=headers,
        )
        assert c1_res.status_code == 201, f"Failed c1: {c1_res.text}"
        c2_res = client.post(
            "/api/v1/clients",
            json={"name": "Ученик 2 (свежий урок)", "phone": "+79992223344", "rate_individual": 1500.0},
            headers=headers,
        )
        assert c2_res.status_code == 201, f"Failed c2: {c2_res.text}"
        c1 = c1_res.json()
        c2 = c2_res.json()

        # Ученик 1: урок был 5 дней назад
        t1_start = now - timedelta(days=5)
        l1_res = client.post(
            "/api/v1/lessons",
            json={
                "teacher_id": teacher_id,
                "client_id": c1["id"],
                "title": "Урок 5 дней назад",
                "start_time": to_rfc3339(t1_start),
                "end_time": to_rfc3339(t1_start + timedelta(hours=1)),
                "format": "individual",
            },
            headers=headers,
        )
        assert l1_res.status_code == 201, f"Failed l1: {l1_res.text}"

        # Ученик 2: урок был 1 день назад
        t2_start = now - timedelta(days=1)
        l2_res = client.post(
            "/api/v1/lessons",
            json={
                "teacher_id": teacher_id,
                "client_id": c2["id"],
                "title": "Урок вчера",
                "start_time": to_rfc3339(t2_start),
                "end_time": to_rfc3339(t2_start + timedelta(hours=1)),
                "format": "individual",
            },
            headers=headers,
        )
        assert l2_res.status_code == 201, f"Failed l2: {l2_res.text}"

        # Запрашиваем клиентов через GET /api/v1/clients
        list_res = client.get("/api/v1/clients", headers=headers)
        assert list_res.status_code == 200
        clients_list = list_res.json()

        c1_found = next(c for c in clients_list if c["id"] == c1["id"])
        c2_found = next(c for c in clients_list if c["id"] == c2["id"])

        assert c1_found.get("last_lesson_at") is not None
        assert c2_found.get("last_lesson_at") is not None
        assert c2_found["last_lesson_at"] > c1_found["last_lesson_at"]

    def test_multi_tenant_isolation(self, client: httpx.Client, setup_teacher_and_client, registered_user):
        """Чужой преподаватель не может читать stream и создавать/удалять заметки клиента."""
        teacher1, client_data = setup_teacher_and_client
        client_id = client_data["id"]

        # Создаем второго преподавателя
        _, reg_data2 = registered_user(role="teacher")
        teacher2_headers = {"Authorization": f"Bearer {reg_data2['tokens']['access_token']}"}

        # 1. Чужой стрим -> 403 Forbidden
        foreign_stream = client.get(
            f"/api/v1/crm/clients/{client_id}/stream",
            headers=teacher2_headers,
        )
        assert foreign_stream.status_code == 403

        # 2. Создание заметки чужому клиенту -> 403 Forbidden
        foreign_create = client.post(
            f"/api/v1/crm/clients/{client_id}/notes",
            json={"content": "Взлом заметки"},
            headers=teacher2_headers,
        )
        assert foreign_create.status_code == 403

        # 3. Учитель 1 создает заметку
        n_res = client.post(
            f"/api/v1/crm/clients/{client_id}/notes",
            json={"content": "Своя заметка"},
            headers=teacher1["headers"],
        )
        assert n_res.status_code == 201
        note_id = n_res.json()["id"]

        # 4. Учитель 2 пытается удалить заметку Учителя 1 -> 404 (строгая изоляция по id + teacher_id)
        foreign_del = client.delete(
            f"/api/v1/crm/clients/notes/{note_id}",
            headers=teacher2_headers,
        )
        assert foreign_del.status_code == 404

