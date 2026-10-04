import pytest
import httpx


@pytest.mark.auth
class TestUserLogin:
    """Тестирование аутентификации пользователей и валидации учетных данных."""

    def test_login_success(self, client: httpx.Client, registered_user):
        payload, _ = registered_user()

        login_res = client.post("/api/v1/auth/login", json={
            "email": payload["email"],
            "password": payload["password"],
        })

        assert login_res.status_code == 200
        data = login_res.json()

        assert "user" in data
        assert "tokens" in data
        assert data["user"]["email"] == payload["email"]
        assert len(data["tokens"]["access_token"]) > 20
        assert len(data["tokens"]["refresh_token"]) > 20

    def test_login_wrong_password_unauthorized(self, client: httpx.Client, registered_user):
        payload, _ = registered_user()

        login_res = client.post("/api/v1/auth/login", json={
            "email": payload["email"],
            "password": "WrongPassword123!",
        })

        assert login_res.status_code == 401
        data = login_res.json()
        assert "error" in data

    def test_login_nonexistent_user_unauthorized(self, client: httpx.Client):
        login_res = client.post("/api/v1/auth/login", json={
            "email": "nonexistent_ghost_user@school.test",
            "password": "AnyPassword123!",
        })

        assert login_res.status_code == 401
