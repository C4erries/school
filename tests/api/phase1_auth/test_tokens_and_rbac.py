import pytest
import httpx


@pytest.mark.auth
class TestTokensAndRBAC:
    """Тестирование JWT токенов, ротации refresh токенов в Valkey и защиты роутов."""

    def test_get_current_user_me_authorized(self, client: httpx.Client, registered_user):
        payload, reg_data = registered_user()
        access_token = reg_data["tokens"]["access_token"]

        headers = {"Authorization": f"Bearer {access_token}"}
        res = client.get("/api/v1/auth/me", headers=headers)

        assert res.status_code == 200
        user = res.json()
        assert user["email"] == payload["email"]
        assert user["full_name"] == payload["full_name"]
        assert user["role"] == payload["role"]

    def test_get_current_user_me_unauthorized_without_header(self, client: httpx.Client):
        res = client.get("/api/v1/auth/me")
        assert res.status_code == 401

    def test_get_current_user_me_invalid_token(self, client: httpx.Client):
        headers = {"Authorization": "Bearer invalid.jwt.token"}
        res = client.get("/api/v1/auth/me", headers=headers)
        assert res.status_code == 401

    def test_refresh_token_rotation_success(self, client: httpx.Client, registered_user):
        _, reg_data = registered_user()
        initial_refresh_token = reg_data["tokens"]["refresh_token"]

        # 1. Запрос на обновление токенов
        refresh_res = client.post("/api/v1/auth/refresh", json={
            "refresh_token": initial_refresh_token,
        })
        assert refresh_res.status_code == 200
        new_tokens = refresh_res.json()["tokens"]

        assert new_tokens["access_token"] != reg_data["tokens"]["access_token"]
        assert new_tokens["refresh_token"] != initial_refresh_token

        # 2. Проверяем, что новый access_token валиден для /auth/me
        me_res = client.get("/api/v1/auth/me", headers={
            "Authorization": f"Bearer {new_tokens['access_token']}",
        })
        assert me_res.status_code == 200

        # 3. Проверяем ротацию: СТАРЫЙ refresh_token больше не должен работать
        old_refresh_res = client.post("/api/v1/auth/refresh", json={
            "refresh_token": initial_refresh_token,
        })
        assert old_refresh_res.status_code == 401

    @pytest.mark.smoke
    def test_healthcheck_endpoint(self, client: httpx.Client):
        res = client.get("/api/v1/health")
        assert res.status_code == 200
        data = res.json()
        assert data["status"] == "ok"
        assert data["service"] == "school-api"

    def test_user_default_rates(self, client: httpx.Client, teacher_user):
        """Проверка получения и обновления дефолтных ставок преподавателя."""
        headers = teacher_user["headers"]

        # 1. Получаем текущие дефолтные ставки
        get_res = client.get("/api/v1/users/me/rates", headers=headers)
        assert get_res.status_code == 200
        initial_rates = get_res.json()
        assert "rate_individual" in initial_rates
        assert "rate_pair" in initial_rates
        assert "rate_group" in initial_rates

        # 2. Обновляем дефолтные ставки
        put_payload = {
            "rate_individual": 2700.0,
            "rate_pair": 1900.0,
            "rate_group": 1300.0,
        }
        put_res = client.put("/api/v1/users/me/rates", json=put_payload, headers=headers)
        assert put_res.status_code == 200
        updated = put_res.json()
        assert updated["rate_individual"] == 2700.0
        assert updated["rate_pair"] == 1900.0
        assert updated["rate_group"] == 1300.0

        # 3. Повторный GET возвращает обновленные значения
        get_res2 = client.get("/api/v1/users/me/rates", headers=headers)
        assert get_res2.status_code == 200
        assert get_res2.json()["rate_individual"] == 2700.0
        assert get_res2.json()["rate_pair"] == 1900.0
        assert get_res2.json()["rate_group"] == 1300.0

