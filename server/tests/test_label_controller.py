import uuid


async def test_success(test_client, session_factory, label_factory, auth_headers):
    session1 = await session_factory()
    session2 = await session_factory()

    await label_factory(session1, "environment", "dev")
    await label_factory(session1, "priority", "high")
    await label_factory(session2, "environment", "prod")

    response = await test_client.get(
        "/api/v1/labels",
        params={"session_id": str(session1.id), "offset": 0, "limit": 10},
        headers=auth_headers,
    )

    assert response.status_code == 200, f"Response: {response.text}"
    data = response.json()

    labels = {item["key"]: item["value"] for item in data["items"]}
    assert labels == {"environment": "dev", "priority": "high"}


async def test_empty(test_client, session_factory, auth_headers):
    session = await session_factory()

    response = await test_client.get(
        "/api/v1/labels",
        params={"session_id": str(session.id), "offset": 0, "limit": 10},
        headers=auth_headers,
    )

    assert response.status_code == 200
    assert response.json()["items"] == []


async def test_nonexistent_session(test_client, auth_headers):
    nonexistent_session_id = str(uuid.uuid4())

    response = await test_client.get(
        "/api/v1/labels",
        params={"session_id": nonexistent_session_id, "offset": 0, "limit": 10},
        headers=auth_headers,
    )

    assert response.status_code == 404


async def test_invalid_session_id(test_client, auth_headers):
    response = await test_client.get(
        "/api/v1/labels",
        params={"session_id": "invalid-uuid", "offset": 0, "limit": 10},
        headers=auth_headers,
    )

    assert response.status_code == 400
