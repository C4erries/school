import os
import uuid
from typing import Any, Dict, Generator, Tuple
import pytest
import httpx
from faker import Faker

fake = Faker()

BASE_URL = os.getenv("BASE_URL", "http://nginx:80").rstrip("/")


@pytest.fixture(scope="session")
def base_url() -> str:
    return BASE_URL


@pytest.fixture(scope="session")
def client() -> Generator[httpx.Client, None, None]:
    """Базовый HTTP-клиент для вызовов API через Nginx."""
    with httpx.Client(base_url=BASE_URL, timeout=10.0) as c:
        yield c


@pytest.fixture
def random_user_payload():
    """Генератор случайных валидных данных пользователя."""
    def _generator(role: str = "student") -> Dict[str, Any]:
        unique_id = uuid.uuid4().hex[:8]
        return {
            "full_name": fake.name(),
            "email": f"test_{unique_id}@{fake.free_email_domain()}",
            "password": f"Pass_{fake.password(length=10, special_chars=True)}",
            "phone": fake.phone_number(),
            "role": role,
        }
    return _generator


@pytest.fixture
def registered_user(client: httpx.Client, random_user_payload):
    """Фикстура, создающая зарегистрированного пользователя и возвращающая (payload, response_json)."""
    def _create(role: str = "student") -> Tuple[Dict[str, Any], Dict[str, Any]]:
        payload = random_user_payload(role=role)
        res = client.post("/api/v1/auth/register", json=payload)
        assert res.status_code == 201, f"Не удалось зарегистрировать пользователя: {res.text}"
        return payload, res.json()
    return _create


@pytest.fixture
def auth_headers():
    """Генератор заголовков авторизации Bearer."""
    def _headers(token: str) -> Dict[str, str]:
        return {"Authorization": f"Bearer {token}"}
    return _headers


@pytest.fixture
def admin_user(registered_user):
    """Создает пользователя с ролью администратора (owner)."""
    payload, reg_data = registered_user(role="owner")
    token = reg_data["tokens"]["access_token"]
    return {
        "payload": payload,
        "user": reg_data["user"],
        "id": reg_data["user"]["id"],
        "token": token,
        "headers": {"Authorization": f"Bearer {token}"},
    }


@pytest.fixture
def teacher_user(registered_user):
    """Создает пользователя с ролью преподавателя (teacher)."""
    payload, reg_data = registered_user(role="teacher")
    token = reg_data["tokens"]["access_token"]
    return {
        "payload": payload,
        "user": reg_data["user"],
        "id": reg_data["user"]["id"],
        "token": token,
        "headers": {"Authorization": f"Bearer {token}"},
    }


@pytest.fixture
def student_user(registered_user):
    """Создает пользователя с ролью ученика (student)."""
    payload, reg_data = registered_user(role="student")
    token = reg_data["tokens"]["access_token"]
    return {
        "payload": payload,
        "user": reg_data["user"],
        "id": reg_data["user"]["id"],
        "token": token,
        "headers": {"Authorization": f"Bearer {token}"},
    }
