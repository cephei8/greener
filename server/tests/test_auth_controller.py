import pytest


class TestLogin:
    async def test_success(self, test_client, user):
        response = await test_client.post(
            "/api/v1/auth/login",
            json={
                "username": user.username,
                "password": "testpass123",
            },
        )

        assert response.status_code == 201

    @pytest.mark.usefixtures("user")
    async def test_invalid_username(self, test_client):
        response = await test_client.post(
            "/api/v1/auth/login",
            json={
                "username": "nonexistent",
                "password": "somepassword",
            },
        )

        assert response.status_code == 401

    async def test_invalid_password(self, test_client, user):
        response = await test_client.post(
            "/api/v1/auth/login",
            json={
                "username": user.username,
                "password": "wrongpassword",
            },
        )

        assert response.status_code == 401


class TestRefreshToken:
    async def test_refresh_token_generates_new_access_token(self, test_client, user):
        login_response = await test_client.post(
            "/api/v1/auth/login",
            json={
                "username": user.username,
                "password": "testpass123",
            },
        )
        assert login_response.status_code == 201

        refresh_token = login_response.json()["refreshToken"]
        refresh_response = await test_client.post(
            "/api/v1/auth/refresh",
            json={"refreshToken": refresh_token},
        )
        assert refresh_response.status_code == 201

    async def test_invalid_token_fails(self, test_client):
        response = await test_client.post(
            "/api/v1/auth/refresh",
            json={"refreshToken": "invalid token"},
        )

        assert response.status_code == 401

    async def test_refresh_with_access_token_fails(self, test_client, user):
        login_response = await test_client.post(
            "/api/v1/auth/login",
            json={
                "username": user.username,
                "password": "testpass123",
            },
        )
        assert login_response.status_code == 201

        access_token = login_response.json()["accessToken"]
        refresh_response = await test_client.post(
            "/api/v1/auth/refresh",
            json={"refreshToken": access_token},
        )
        assert refresh_response.status_code == 401

    async def test_refresh_token_cannot_be_used_as_access_token(
        self, test_client, user
    ):
        login_response = await test_client.post(
            "/api/v1/auth/login",
            json={
                "username": user.username,
                "password": "testpass123",
            },
        )
        assert login_response.status_code == 201

        refresh_token = login_response.json()["refreshToken"]
        sessions_response = await test_client.get(
            "/api/v1/sessions",
            headers={"Authorization": f"Bearer {refresh_token}"},
        )

        assert sessions_response.status_code == 401


class TestChangePassword:
    async def test_success(self, test_client, auth_headers):
        response = await test_client.post(
            "/api/v1/auth/change-password",
            json={
                "passwordOld": "testpass123",
                "passwordNew": "newpass456",
            },
            headers=auth_headers,
        )
        assert response.status_code == 201

    async def test_wrong_password(self, test_client, auth_headers):
        response = await test_client.post(
            "/api/v1/auth/change-password",
            json={
                "passwordOld": "wrongpassword",
                "passwordNew": "newpass456",
            },
            headers=auth_headers,
        )

        assert response.status_code == 401

    async def test_empty_password(self, test_client, auth_headers):
        response = await test_client.post(
            "/api/v1/auth/change-password",
            json={
                "passwordOld": "testpass123",
                "passwordNew": "",
            },
            headers=auth_headers,
        )

        assert response.status_code == 400
