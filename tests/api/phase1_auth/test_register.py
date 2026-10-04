import pytest
import httpx


@pytest.mark.auth
class TestUserRegistration:
    """Тестирование сценариев регистрации новых пользователей."""

    def test_register_student_success(self, client: httpx.Client, random_user_payload):
        payload = random_user_payload(role="student")
        response = client.post("/api/v1/auth/register", json=payload)

        assert response.status_code == 201
        data = response.json()

        assert "user" in data
        assert "tokens" in data

        user = data["user"]
        assert user["email"] == payload["email"]
        assert user["full_name"] == payload["full_name"]
        assert user["role"] == "student"
        assert "id" in user
        assert "password" not in user
        assert "password_hash" not in user

        tokens = data["tokens"]
        assert "access_token" in tokens
        assert "refresh_token" in tokens
        assert tokens["token_type"] == "Bearer"
        assert tokens["expires_in"] == 900

    def test_register_teacher_success(self, client: httpx.Client, random_user_payload):
        payload = random_user_payload(role="teacher")
        response = client.post("/api/v1/auth/register", json=payload)

        assert response.status_code == 201
        data = response.json()
        assert data["user"]["role"] == "teacher"

    def test_register_duplicate_email_conflict(self, client: httpx.Client, random_user_payload):
        payload = random_user_payload()
        # Первая регистрация
        res1 = client.post("/api/v1/auth/register", json=payload)
        assert res1.status_code == 201

        # Повторная регистрация с тем же email
        res2 = client.post("/api/v1/auth/register", json=payload)
        assert res2.status_code == 409
        error = res2.json()
        assert "error" in error

    @pytest.mark.parametrize("invalid_email", [
        "not-an-email",
        "@example.com",
        "user@",
        "",
    ])
    def test_register_invalid_email_bad_request(self, client: httpx.Client, random_user_payload, invalid_email):
        payload = random_user_payload()
        payload["email"] = invalid_email

        response = client.post("/api/v1/auth/register", json=payload)
        assert response.status_code == 400

    def test_register_short_password_bad_request(self, client: httpx.Client, random_user_payload):
        payload = random_user_payload()
        payload["password"] = "123"  # Минимальная длина 8

        response = client.post("/api/v1/auth/register", json=payload)
        assert response.status_code == 400

    def test_register_empty_name_bad_request(self, client: httpx.Client, random_user_payload):
        payload = random_user_payload()
        payload["full_name"] = " "

        response = client.post("/api/v1/auth/register", json=payload)
        assert response.status_code == 400
