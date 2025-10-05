import base64
import json
import uuid

from app.app_init import provide_apikey_repo


async def test_create(test_client, auth_headers):
    response = await test_client.post(
        "/api/v1/api-keys",
        json={"description": "Test API Key"},
        headers=auth_headers,
    )
    assert response.status_code == 201

    data = response.json()
    key_decoded = json.loads(base64.b64decode(data["key"]).decode())
    assert "apiKeyId" in key_decoded
    assert "apiKeySecret" in key_decoded
    assert key_decoded["apiKeyId"] == data["id"]


async def test_get(test_client, apikey_factory, auth_headers):
    apikey = await apikey_factory(description="Test Key")

    response = await test_client.get(
        f"/api/v1/api-keys/{apikey.id}", headers=auth_headers
    )
    assert response.status_code == 200
    assert response.json()["id"] == str(apikey.id)


async def test_list(test_client, apikey_factory, auth_headers):
    key1 = await apikey_factory(description="Key 1")
    key2 = await apikey_factory(description="Key 2")

    response = await test_client.get("/api/v1/api-keys", headers=auth_headers)
    assert response.status_code == 200

    ids = {uuid.UUID(item["id"]) for item in response.json()["items"]}
    assert {key1.id, key2.id} == ids


async def test_delete(test_client, apikey_factory, db_session_factory, auth_headers):
    apikey = await apikey_factory(description="Key to delete")

    response = await test_client.delete(
        f"/api/v1/api-keys/{apikey.id}", headers=auth_headers
    )
    assert response.status_code == 204

    async with db_session_factory() as db_session:
        repo = await provide_apikey_repo(db_session)
        deleted_key = await repo.get_one_or_none(id=apikey.id)
        assert deleted_key is None
