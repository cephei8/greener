import uuid


async def test_get_success(test_client, session_factory, auth_headers):
    session = await session_factory()

    response = await test_client.get(
        f"/api/v1/sessions/{session.id}", headers=auth_headers
    )

    assert response.status_code == 200

    data = response.json()
    assert data["id"] == str(session.id)


async def test_get_not_found(test_client, auth_headers):
    nonexistent_id = uuid.uuid4()

    response = await test_client.get(
        f"/api/v1/sessions/{nonexistent_id}", headers=auth_headers
    )

    assert response.status_code == 404


async def test_get_invalid_uuid(test_client, auth_headers):
    response = await test_client.get(
        "/api/v1/sessions/invalid-uuid", headers=auth_headers
    )

    assert response.status_code == 404
