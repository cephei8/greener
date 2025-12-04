import uuid

from app.models import TestcaseStatus


async def test_validate_non_grouping(test_client, auth_headers):
    response = await test_client.get(
        "/api/v1/groups/validate-query",
        params={"queryStr": f'session_id = "{uuid.uuid4()}"'},
        headers=auth_headers,
    )

    assert response.status_code == 200, response.json()
    assert response.json()["isGrouping"] is False


async def test_validate_grouping(test_client, auth_headers):
    response = await test_client.get(
        "/api/v1/groups/validate-query",
        params={"queryStr": "group_by(session_id)"},
        headers=auth_headers,
    )

    assert response.status_code == 200, response.json()
    assert response.json()["isGrouping"] is True


async def test_validate_invalid_query(test_client, auth_headers):
    response = await test_client.get(
        "/api/v1/groups/validate-query",
        params={"queryStr": "abc"},
        headers=auth_headers,
    )

    assert response.status_code == 400, response.json()
    assert response.json()["detail"].startswith("Invalid query")


async def test_list_empty_query(test_client, auth_headers):
    response = await test_client.get(
        "/api/v1/groups",
        params={"queryStr": "", "offset": 0, "limit": 10},
        headers=auth_headers,
    )

    assert response.status_code == 200
    assert response.json()["items"] == []


async def test_list_invalid_query(test_client, auth_headers):
    response = await test_client.get(
        "/api/v1/groups",
        params={
            "queryStr": "invalid_query_syntax",
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )

    assert response.status_code == 400
    assert "Invalid query" in response.json()["detail"]


async def test_session_id(test_client, session_factory, testcase_factory, auth_headers):
    session1 = await session_factory()
    session2 = await session_factory()

    await testcase_factory(session1, status=TestcaseStatus.PASS)
    await testcase_factory(session1, status=TestcaseStatus.FAIL)
    await testcase_factory(session2, status=TestcaseStatus.PASS)
    await testcase_factory(session2, status=TestcaseStatus.ERROR)
    await testcase_factory(session2, status=TestcaseStatus.SKIP)

    response = await test_client.get(
        "/api/v1/groups",
        params={
            "queryStr": "group_by(session_id)",
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )
    assert response.status_code == 200, response.text
    assert response.json() == {
        "items": sorted(
            [
                {
                    "columns": [str(session1.id)],
                    "status": 1,
                },
                {
                    "columns": [str(session2.id)],
                    "status": 0,
                },
            ],
            key=lambda x: x["columns"][0],
        ),
        "offset": 0,
        "limit": 10,
        "total": 2,
        "header": ["session_id"],
        "aggregatedStatus": 0,
    }

    response = await test_client.get(
        "/api/v1/groups",
        params={
            "queryStr": f'session_id="{session1.id}" group_by(session_id)',
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )
    assert response.status_code == 200, response.text
    assert response.json() == {
        "items": [
            {
                "columns": [str(session1.id)],
                "status": 1,
            },
        ],
        "offset": 0,
        "limit": 10,
        "total": 1,
        "header": ["session_id"],
        "aggregatedStatus": 1,
    }


async def test_label1(
    test_client, session_factory, testcase_factory, label_factory, auth_headers
):
    session1 = await session_factory()
    session2 = await session_factory()
    session3 = await session_factory()

    await label_factory(session1, "environment", "dev")
    await label_factory(session2, "environment", "test")
    await label_factory(session3, "environment", "production")
    await label_factory(session1, "priority", "medium")
    await label_factory(session2, "priority", "medium")
    await label_factory(session3, "priority", "high")
    await label_factory(session3, "access", "restricted")

    await testcase_factory(session1, status=TestcaseStatus.PASS)
    await testcase_factory(session2, status=TestcaseStatus.PASS)
    await testcase_factory(session3, status=TestcaseStatus.PASS)

    response = await test_client.get(
        "/api/v1/groups",
        params={
            "queryStr": 'group_by(#"priority")',
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )
    assert response.status_code == 200, response.text
    assert response.json() == {
        "items": [
            {
                "columns": ["high"],
                "status": 2,
            },
            {
                "columns": ["medium"],
                "status": 2,
            },
        ],
        "offset": 0,
        "limit": 10,
        "total": 2,
        "header": ['#"priority"'],
        "aggregatedStatus": 2,
    }


async def test_label2(
    test_client, session_factory, testcase_factory, label_factory, auth_headers
):
    session1 = await session_factory()
    session2 = await session_factory()
    session3 = await session_factory()

    await label_factory(session1, "environment", "dev")
    await label_factory(session2, "environment", "test")
    await label_factory(session3, "environment", "production")
    await label_factory(session1, "priority", "medium")
    await label_factory(session2, "priority", "medium")
    await label_factory(session3, "priority", "high")
    await label_factory(session3, "access", "restricted")

    await testcase_factory(session1, status=TestcaseStatus.PASS)
    await testcase_factory(session2, status=TestcaseStatus.PASS)
    await testcase_factory(session3, status=TestcaseStatus.PASS)

    response = await test_client.get(
        "/api/v1/groups",
        params={
            "queryStr": 'group_by(#"priority", #"environment")',
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )
    assert response.status_code == 200, response.text
    assert response.json() == {
        "items": [
            {
                "columns": ["high", "production"],
                "status": 2,
            },
            {
                "columns": ["medium", "dev"],
                "status": 2,
            },
            {
                "columns": ["medium", "test"],
                "status": 2,
            },
        ],
        "offset": 0,
        "limit": 10,
        "total": 3,
        "header": ['#"priority"', '#"environment"'],
        "aggregatedStatus": 2,
    }


async def test_label3(
    test_client, session_factory, testcase_factory, label_factory, auth_headers
):
    session1 = await session_factory()
    session2 = await session_factory()
    session3 = await session_factory()

    await label_factory(session1, "environment", "dev")
    await label_factory(session2, "environment", "test")
    await label_factory(session3, "environment", "production")
    await label_factory(session1, "priority", "medium")
    await label_factory(session2, "priority", "medium")
    await label_factory(session3, "priority", "high")
    await label_factory(session3, "access", "restricted")

    await testcase_factory(session1, status=TestcaseStatus.PASS)
    await testcase_factory(session2, status=TestcaseStatus.PASS)
    await testcase_factory(session3, status=TestcaseStatus.PASS)

    response = await test_client.get(
        "/api/v1/groups",
        params={
            "queryStr": 'group_by(#"environment", #"priority")',
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )
    assert response.status_code == 200, response.text
    assert response.json() == {
        "items": [
            {
                "columns": ["dev", "medium"],
                "status": 2,
            },
            {
                "columns": ["production", "high"],
                "status": 2,
            },
            {
                "columns": ["test", "medium"],
                "status": 2,
            },
        ],
        "offset": 0,
        "limit": 10,
        "total": 3,
        "header": ['#"environment"', '#"priority"'],
        "aggregatedStatus": 2,
    }


async def test_label4(
    test_client, session_factory, testcase_factory, label_factory, auth_headers
):
    session1 = await session_factory()
    session2 = await session_factory()
    session3 = await session_factory()

    await label_factory(session1, "environment", "dev")
    await label_factory(session2, "environment", "test")
    await label_factory(session3, "environment", "production")
    await label_factory(session1, "priority", "medium")
    await label_factory(session2, "priority", "medium")
    await label_factory(session3, "priority", "high")
    await label_factory(session3, "access", "restricted")

    await testcase_factory(session1, status=TestcaseStatus.PASS)
    await testcase_factory(session2, status=TestcaseStatus.PASS)
    await testcase_factory(session3, status=TestcaseStatus.PASS)

    response = await test_client.get(
        "/api/v1/groups",
        params={
            "queryStr": '#"priority" = "medium" group_by(#"environment", #"priority")',
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )
    assert response.status_code == 200, response.text
    assert response.json() == {
        "items": [
            {
                "columns": ["dev", "medium"],
                "status": 2,
            },
            {
                "columns": ["test", "medium"],
                "status": 2,
            },
        ],
        "offset": 0,
        "limit": 10,
        "total": 2,
        "header": ['#"environment"', '#"priority"'],
        "aggregatedStatus": 2,
    }


async def test_tag_where(
    test_client, session_factory, label_factory, testcase_factory, auth_headers
):
    session1 = await session_factory()
    session2 = await session_factory()
    session3 = await session_factory()

    await label_factory(session1, "environment", "dev")
    await label_factory(session2, "environment", "test")
    await label_factory(session3, "environment", "production")
    await label_factory(session2, "access", "qa-team")
    await label_factory(session3, "access", "restricted")

    await testcase_factory(session1, status=TestcaseStatus.PASS)
    await testcase_factory(session2, status=TestcaseStatus.PASS)
    await testcase_factory(session3, status=TestcaseStatus.PASS)

    response = await test_client.get(
        "/api/v1/groups",
        params={
            "queryStr": '#"access" group_by(session_id)',
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )

    assert response.status_code == 200, response.text
    assert response.json() == {
        "items": sorted(
            [
                {
                    "columns": [str(session2.id)],
                    "status": 2,
                },
                {
                    "columns": [str(session3.id)],
                    "status": 2,
                },
            ],
            key=lambda x: x["columns"][0],
        ),
        "offset": 0,
        "limit": 10,
        "total": 2,
        "header": ["session_id"],
        "aggregatedStatus": 2,
    }


async def test_not_tag_where(
    test_client, session_factory, label_factory, testcase_factory, auth_headers
):
    session1 = await session_factory()
    session2 = await session_factory()
    session3 = await session_factory()

    await label_factory(session1, "environment", "dev")
    await label_factory(session2, "environment", "test")
    await label_factory(session3, "environment", "production")
    await label_factory(session3, "access", "restricted")

    await testcase_factory(session1, status=TestcaseStatus.PASS)
    await testcase_factory(session2, status=TestcaseStatus.PASS)
    await testcase_factory(session3, status=TestcaseStatus.PASS)

    response = await test_client.get(
        "/api/v1/groups",
        params={
            "queryStr": '!#"access" group_by(session_id)',
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )

    assert response.status_code == 200, response.text
    assert response.json() == {
        "items": sorted(
            [
                {
                    "columns": [str(session1.id)],
                    "status": 2,
                },
                {
                    "columns": [str(session2.id)],
                    "status": 2,
                },
            ],
            key=lambda x: x["columns"][0],
        ),
        "offset": 0,
        "limit": 10,
        "total": 2,
        "header": ["session_id"],
        "aggregatedStatus": 2,
    }


async def test_not_tag_where2(
    test_client, session_factory, label_factory, testcase_factory, auth_headers
):
    session1 = await session_factory()
    session2 = await session_factory()
    session3 = await session_factory()

    await label_factory(session1, "environment", "dev")
    await label_factory(session2, "environment", "test")
    await label_factory(session3, "environment", "production")
    await label_factory(session3, "access", "restricted")

    await testcase_factory(session1, status=TestcaseStatus.PASS)
    await testcase_factory(session2, status=TestcaseStatus.PASS)
    await testcase_factory(session3, status=TestcaseStatus.PASS)

    response = await test_client.get(
        "/api/v1/groups",
        params={
            "queryStr": '!#"access" or #"access" group_by(session_id)',
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )

    assert response.status_code == 200, response.text
    assert response.json() == {
        "items": sorted(
            [
                {
                    "columns": [str(session1.id)],
                    "status": 2,
                },
                {
                    "columns": [str(session2.id)],
                    "status": 2,
                },
                {
                    "columns": [str(session3.id)],
                    "status": 2,
                },
            ],
            key=lambda x: x["columns"][0],
        ),
        "offset": 0,
        "limit": 10,
        "total": 3,
        "header": ["session_id"],
        "aggregatedStatus": 2,
    }


async def test_not_tag_where3(
    test_client, session_factory, label_factory, testcase_factory, auth_headers
):
    session1 = await session_factory()
    session2 = await session_factory()
    session3 = await session_factory()

    await label_factory(session1, "environment", "dev")
    await label_factory(session2, "environment", "test")
    await label_factory(session3, "environment", "production")
    await label_factory(session3, "access", "restricted")

    await testcase_factory(session1, status=TestcaseStatus.PASS)
    await testcase_factory(session2, status=TestcaseStatus.PASS)
    await testcase_factory(session3, status=TestcaseStatus.PASS)

    response = await test_client.get(
        "/api/v1/groups",
        params={
            "queryStr": '!#"access" and #"environment"="test" group_by(session_id)',
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )

    assert response.status_code == 200, response.text
    assert response.json() == {
        "items": [
            {
                "columns": [str(session2.id)],
                "status": 2,
            },
        ],
        "offset": 0,
        "limit": 10,
        "total": 1,
        "header": ["session_id"],
        "aggregatedStatus": 2,
    }


async def test_list_endpoint_valueless_label(
    test_client, session_factory, testcase_factory, label_factory, auth_headers
):
    session1 = await session_factory()
    session2 = await session_factory()

    await testcase_factory(session1, status=TestcaseStatus.PASS)
    await testcase_factory(session2, status=TestcaseStatus.PASS)

    await label_factory(session1, "triaged")
    await label_factory(session2, "triaged")

    response = await test_client.get(
        "/api/v1/groups",
        params={
            "queryStr": 'group_by(#"triaged")',
            "offset": 0,
            "limit": 10,
        },
        headers=auth_headers,
    )

    assert response.status_code == 200, response.text
    data = response.json()

    assert len(data["items"]) == 1

    item = data["items"][0]
    assert data["header"] == ['#"triaged"']
    assert item["columns"] == [None]
