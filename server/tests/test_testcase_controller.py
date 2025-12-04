import json
import uuid
from urllib.parse import quote

from app.models import TestcaseStatus


async def test_list_no_filters(
    test_client, session_factory, testcase_factory, auth_headers
):
    session1 = await session_factory()
    session2 = await session_factory()

    tc1 = await testcase_factory(session1, status=TestcaseStatus.PASS, name="test_one")
    tc2 = await testcase_factory(session2, status=TestcaseStatus.FAIL, name="test_two")

    response = await test_client.get(
        "/api/v1/testcases", params={"offset": 0, "limit": 10}, headers=auth_headers
    )

    assert response.status_code == 200

    ids = {uuid.UUID(item["id"]) for item in response.json()["items"]}
    assert ids == {tc1.id, tc2.id}


async def test_list_empty(test_client, auth_headers):
    response = await test_client.get(
        "/api/v1/testcases", params={"offset": 0, "limit": 10}, headers=auth_headers
    )

    assert response.status_code == 200
    assert response.json()["items"] == []


async def test_list_with_query(
    test_client, session_factory, testcase_factory, auth_headers
):
    session = await session_factory()

    tc = await testcase_factory(session, status=TestcaseStatus.PASS, name="test_pass")
    await testcase_factory(session, status=TestcaseStatus.FAIL, name="test_fail")
    await testcase_factory(session, status=TestcaseStatus.ERROR, name="test_error")

    response = await test_client.get(
        "/api/v1/testcases",
        params={"queryStr": 'status="pass"', "offset": 0, "limit": 10},
        headers=auth_headers,
    )

    assert response.status_code == 200, f"Response: {response.text}"

    data = response.json()
    assert data["total"] == 1
    assert len(data["items"]) == 1
    assert data["items"][0]["id"] == str(tc.id)


async def test_list_missing_group(test_client, auth_headers):
    response = await test_client.get(
        "/api/v1/testcases",
        params={"queryStr": "group_by(session_id)", "offset": 0, "limit": 10},
        headers=auth_headers,
    )

    assert response.status_code == 400
    data = response.json()
    assert "Group parameter is required when using a grouping query" in data["detail"]


async def test_list_group_no_grouping(test_client, auth_headers):
    group_data = [["session_id"], ["abc-123"]]
    group_param = quote(json.dumps(group_data))

    response = await test_client.get(
        "/api/v1/testcases",
        params={
            "queryStr": 'status="pass"',
            "group": group_param,
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )

    assert response.status_code == 400
    data = response.json()
    assert "Group parameter can only be used with grouping queries" in data["detail"]


async def test_list_mismatched_keys(test_client, auth_headers):
    group_data = [['#"tag1"'], ["value1"]]
    group_param = quote(json.dumps(group_data))

    response = await test_client.get(
        "/api/v1/testcases",
        params={
            "queryStr": "group_by(session_id)",
            "group": group_param,
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )

    assert response.status_code == 400
    data = response.json()
    assert "do not match the grouping query keys" in data["detail"]


async def test_list_session_group(
    test_client, session_factory, testcase_factory, auth_headers
):
    session1 = await session_factory()
    session2 = await session_factory()

    tc1 = await testcase_factory(session1)
    tc2 = await testcase_factory(session1)
    await testcase_factory(session2)

    group_data = [["session_id"], [str(session1.id)]]
    group_param = quote(json.dumps(group_data))

    response = await test_client.get(
        "/api/v1/testcases",
        params={
            "queryStr": "group_by(session_id)",
            "group": group_param,
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )

    assert response.status_code == 200

    data = response.json()
    ids = {uuid.UUID(item["id"]) for item in data["items"]}
    assert ids == {tc1.id, tc2.id}


async def test_list_tag_group(
    test_client,
    session_factory,
    testcase_factory,
    label_factory,
    auth_headers,
):
    session1 = await session_factory()
    session2 = await session_factory()
    session3 = await session_factory()

    await label_factory(session1, "environment", "dev")
    await label_factory(session2, "environment", "test")
    await label_factory(session3, "environment", "prod")

    tc = await testcase_factory(session1)
    await testcase_factory(session2)
    await testcase_factory(session3)

    group_data = [['#"environment"'], ["dev"]]
    group_param = quote(json.dumps(group_data))

    response = await test_client.get(
        "/api/v1/testcases",
        params={
            "queryStr": 'group_by(#"environment")',
            "group": group_param,
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )

    assert response.status_code == 200
    data = response.json()

    assert data["total"] == 1
    assert data["items"][0]["id"] == str(tc.id)


async def test_list_invalid_group_format(test_client, auth_headers):
    response = await test_client.get(
        "/api/v1/testcases",
        params={
            "queryStr": "group_by(session_id)",
            "group": "invalid-json",
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )

    assert response.status_code == 400
    data = response.json()
    assert "Invalid group identifier" in data["detail"]


async def test_list_invalid_group_structure(test_client, auth_headers):
    group_data = ["session_id"]
    group_param = quote(json.dumps(group_data))

    response = await test_client.get(
        "/api/v1/testcases",
        params={
            "queryStr": "group_by(session_id)",
            "group": group_param,
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )

    assert response.status_code == 400
    data = response.json()
    assert (
        "Group identifier must be a tuple/array with exactly 2 elements"
        in data["detail"]
    )


async def test_list_mismatched_group_lengths(test_client, auth_headers):
    group_data = [["session_id", '#"env"'], ["value1"]]  # 2 keys, 1 value
    group_param = quote(json.dumps(group_data))

    response = await test_client.get(
        "/api/v1/testcases",
        params={
            "queryStr": 'group_by(session_id, #"env")',
            "group": group_param,
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )

    assert response.status_code == 400
    data = response.json()
    assert "Group keys and values must have the same length" in data["detail"]


async def test_list_non_string_group_values(test_client, auth_headers):
    group_data = [["session_id"], [123]]
    group_param = quote(json.dumps(group_data))

    response = await test_client.get(
        "/api/v1/testcases",
        params={
            "queryStr": "group_by(session_id)",
            "group": group_param,
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )

    assert response.status_code == 400
    data = response.json()
    assert "All group values must be strings or None" in data["detail"]


async def test_list_invalid_query_syntax(test_client, auth_headers):
    response = await test_client.get(
        "/api/v1/testcases",
        params={"queryStr": "invalid_syntax_here", "offset": 0, "limit": 10},
        headers=auth_headers,
    )

    assert response.status_code == 400
    data = response.json()
    assert "Invalid query" in data["detail"]


async def test_get_testcase(
    test_client, session_factory, testcase_factory, auth_headers
):
    session = await session_factory()
    testcase = await testcase_factory(
        session, name="test_get", status=TestcaseStatus.PASS
    )

    response = await test_client.get(
        f"/api/v1/testcases/{testcase.id}", headers=auth_headers
    )

    assert response.status_code == 200
    assert response.json()["id"] == str(testcase.id)


async def test_get_validation_errors(test_client, auth_headers):
    response = await test_client.get(
        "/api/v1/testcases",
        params={
            "query_str": "group_by session_id",
            "group": "invalid_json_format_here!@#",
        },
        headers=auth_headers,
    )
    assert response.status_code == 400

    response = await test_client.get(
        "/api/v1/testcases",
        params={
            "query_str": "group_by session_id",
            "group": '{"not": "a list"}',
        },
        headers=auth_headers,
    )
    assert response.status_code == 400

    response = await test_client.get(
        "/api/v1/testcases",
        params={
            "query_str": "group_by session_id",
            "group": '[[123, "not_string_key"], ["value1", "value2"]]',
        },
        headers=auth_headers,
    )
    assert response.status_code == 400


async def test_list_valueless_label(
    test_client,
    session_factory,
    testcase_factory,
    label_factory,
    auth_headers,
):
    session1 = await session_factory()
    session2 = await session_factory()

    tc1 = await testcase_factory(session1, name="test_case_1")
    tc2 = await testcase_factory(session2, name="test_case_2")

    await label_factory(session1, "triaged")
    await label_factory(session2, "triaged")

    group_keys = ['#"triaged"']
    group_values = [None]
    group_param = [group_keys, group_values]

    response = await test_client.get(
        "/api/v1/testcases",
        params={
            "queryStr": 'group_by(#"triaged")',
            "group": json.dumps(group_param),
        },
        headers=auth_headers,
    )

    assert response.status_code == 200
    data = response.json()

    ids = {uuid.UUID(item["id"]) for item in data["items"]}
    assert ids == {tc1.id, tc2.id}


async def test_list_session_and_tag_group(
    test_client,
    session_factory,
    testcase_factory,
    label_factory,
    auth_headers,
):
    session1 = await session_factory()
    session2 = await session_factory()
    session3 = await session_factory()

    tc = await testcase_factory(session1, name="test_case_1")
    await testcase_factory(session2, name="test_case_2")
    await testcase_factory(session3, name="test_case_3")

    await label_factory(session1, "environment", "dev")
    await label_factory(session2, "environment", "test")
    await label_factory(session3, "environment", "prod")

    group_keys = ["session_id", '#"environment"']
    group_values = [str(session1.id), "dev"]
    group_param = [group_keys, group_values]

    response = await test_client.get(
        "/api/v1/testcases",
        params={
            "queryStr": 'group_by(session_id, #"environment")',
            "group": json.dumps(group_param),
        },
        headers=auth_headers,
    )

    assert response.status_code == 200
    data = response.json()

    assert len(data["items"]) == 1
    assert data["items"][0]["id"] == str(tc.id)


async def test_list_multi_tag_group(
    test_client,
    session_factory,
    testcase_factory,
    label_factory,
    auth_headers,
):
    session1 = await session_factory()
    session2 = await session_factory()
    session3 = await session_factory()

    tc = await testcase_factory(session1, name="test_case_1")
    await testcase_factory(session2, name="test_case_2")
    await testcase_factory(session3, name="test_case_3")

    await label_factory(session1, "environment", "dev")
    await label_factory(session1, "priority", "high")

    await label_factory(session2, "environment", "test")
    await label_factory(session2, "priority", "medium")

    await label_factory(session3, "environment", "prod")
    await label_factory(session3, "priority", "low")

    group_keys = ['#"environment"', '#"priority"']
    group_values = ["dev", "high"]
    group_param = [group_keys, group_values]

    response = await test_client.get(
        "/api/v1/testcases",
        params={
            "queryStr": 'group_by(#"environment", #"priority")',
            "group": json.dumps(group_param),
        },
        headers=auth_headers,
    )

    assert response.status_code == 200
    data = response.json()

    assert len(data["items"]) == 1
    assert data["items"][0]["id"] == str(tc.id)
