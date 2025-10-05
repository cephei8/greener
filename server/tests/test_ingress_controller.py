import base64
import json
import uuid

import pytest

from app.controllers.ingress_controller import StatusRequest
from app.models import APIKey
from app.util import hash_secret


@pytest.fixture
async def api_key_header(user, db_session_factory):
    apikey_secret = "test_secret_12345"
    secret_salt = b"test_salt_12345678901234567890123"
    secret_hash = hash_secret(apikey_secret, secret_salt)

    apikey = APIKey(
        description="Test API Key",
        secret_salt=secret_salt,
        secret_hash=secret_hash,
        user_id=user.id,
    )

    async with db_session_factory() as db_session:
        db_session.add(apikey)
        await db_session.commit()
        await db_session.refresh(apikey)

    api_key_data = {"apiKeyId": str(apikey.id), "apiKeySecret": apikey_secret}
    encoded_key = base64.b64encode(json.dumps(api_key_data).encode()).decode()
    return {"x-api-key": encoded_key}


async def test_session_without_auth(test_client):
    response = await test_client.post(
        "/api/v1/ingress/sessions", json={"description": "test session"}
    )
    assert response.status_code == 401


async def test_session(test_client, api_key_header):
    response = await test_client.post(
        "/api/v1/ingress/sessions",
        json={"description": "test session"},
        headers=api_key_header,
    )
    assert response.status_code == 201
    uuid.UUID(response.json()["id"])


async def test_session_custom_id(test_client, api_key_header):
    session_id = str(uuid.uuid4())
    response = await test_client.post(
        "/api/v1/ingress/sessions",
        json={"id": session_id, "description": "test session"},
        headers=api_key_header,
    )
    assert response.status_code == 201
    assert response.json()["id"] == session_id


async def test_session_invalid_id(test_client, api_key_header):
    response = await test_client.post(
        "/api/v1/ingress/sessions",
        json={"id": "invalid-uuid", "description": "test session"},
        headers=api_key_header,
    )
    assert response.status_code == 400


async def test_session_labels(test_client, api_key_header):
    response = await test_client.post(
        "/api/v1/ingress/sessions",
        json={
            "description": "test session",
            "labels": [
                {"key": "environment", "value": "test"},
                {"key": "version", "value": "1.0"},
                {"key": "flag_only"},
            ],
        },
        headers=api_key_header,
    )
    assert response.status_code == 201


async def test_session_baggage(test_client, api_key_header):
    baggage = {
        "build_number": 123,
        "commit_sha": "abc123",
        "nested": {"key": "value"},
    }
    response = await test_client.post(
        "/api/v1/ingress/sessions",
        json={"description": "test session", "baggage": baggage},
        headers=api_key_header,
    )
    assert response.status_code == 201


async def test_testcases_without_auth(test_client):
    response = await test_client.post(
        "/api/v1/ingress/testcases", json={"testcases": []}
    )
    assert response.status_code == 401


async def test_testcases_empty(test_client, api_key_header):
    response = await test_client.post(
        "/api/v1/ingress/testcases", json={"testcases": []}, headers=api_key_header
    )
    assert response.status_code == 201


async def test_testcases_single(test_client, api_key_header, session_factory):
    session = await session_factory()

    response = await test_client.post(
        "/api/v1/ingress/testcases",
        json={
            "testcases": [
                {
                    "sessionId": str(session.id),
                    "testcaseName": "test_example",
                    "status": StatusRequest.PASS.value,
                    "testcaseClassname": "TestExample",
                    "testcaseFile": "test_example.py",
                    "testsuite": "unit_tests",
                    "output": "Test passed successfully",
                    "baggage": {"duration": 0.123},
                }
            ]
        },
        headers=api_key_header,
    )
    assert response.status_code == 201


async def test_testcases_multiple(test_client, api_key_header, session_factory):
    session = await session_factory()

    response = await test_client.post(
        "/api/v1/ingress/testcases",
        json={
            "testcases": [
                {
                    "sessionId": str(session.id),
                    "testcaseName": "test_pass",
                    "status": StatusRequest.PASS.value,
                },
                {
                    "sessionId": str(session.id),
                    "testcaseName": "test_fail",
                    "status": StatusRequest.FAIL.value,
                    "output": "AssertionError: Expected 1, got 2",
                },
                {
                    "sessionId": str(session.id),
                    "testcaseName": "test_error",
                    "status": StatusRequest.ERROR.value,
                    "output": "ImportError: Module not found",
                },
                {
                    "sessionId": str(session.id),
                    "testcaseName": "test_skip",
                    "status": StatusRequest.SKIP.value,
                    "output": "Skipped due to missing dependency",
                },
            ]
        },
        headers=api_key_header,
    )
    assert response.status_code == 201


async def test_testcases_invalid_session_id(test_client, api_key_header):
    response = await test_client.post(
        "/api/v1/ingress/testcases",
        json={
            "testcases": [
                {
                    "sessionId": "invalid-uuid",
                    "testcaseName": "test_example",
                    "status": StatusRequest.PASS.value,
                }
            ]
        },
        headers=api_key_header,
    )
    assert response.status_code == 400


async def test_testcases_unknown_session(test_client, api_key_header):
    unknown_session_id = str(uuid.uuid4())
    response = await test_client.post(
        "/api/v1/ingress/testcases",
        json={
            "testcases": [
                {
                    "sessionId": unknown_session_id,
                    "testcaseName": "test_example",
                    "status": StatusRequest.PASS.value,
                }
            ]
        },
        headers=api_key_header,
    )
    assert response.status_code == 400


@pytest.mark.parametrize("status", ["pass", "fail", "error", "skip"])
async def test_status_enum_mapping(
    status, test_client, api_key_header, session_factory
):
    session = await session_factory()

    response = await test_client.post(
        "/api/v1/ingress/testcases",
        json={
            "testcases": [
                {
                    "sessionId": str(session.id),
                    "testcaseName": "test_a",
                    "status": status,
                }
            ]
        },
        headers=api_key_header,
    )
    assert response.status_code == 201
