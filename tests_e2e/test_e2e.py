import httpx


def test_e2e(server_url, api_key, auth_tokens):
    with httpx.Client() as client:
        session_response = client.post(
            f"{server_url}/api/v1/ingress/sessions",
            json={
                "description": "E2E test session",
                "baggage": {"test_run": "e2e_001", "environment": "test"},
                "labels": [
                    {"key": "branch", "value": "main"},
                    {"key": "commit", "value": "abc123"},
                    {"key": "ci", "value": None},
                ],
            },
            headers={"X-API-Key": api_key},
        )
        assert session_response.status_code == 201
        session_data = session_response.json()
        session_id = session_data["id"]

        testcases_response = client.post(
            f"{server_url}/api/v1/ingress/testcases",
            json={
                "testcases": [
                    {
                        "sessionId": session_id,
                        "testcaseName": "test_login_success",
                        "testcaseClassname": "TestAuth",
                        "testcaseFile": "tests/test_auth.py",
                        "testsuite": "auth",
                        "status": "pass",
                        "output": "Test passed successfully",
                        "baggage": {"duration": 1.5, "retries": 0},
                    },
                    {
                        "sessionId": session_id,
                        "testcaseName": "test_login_failure",
                        "testcaseClassname": "TestAuth",
                        "testcaseFile": "tests/test_auth.py",
                        "testsuite": "auth",
                        "status": "pass",
                        "output": "Test passed successfully",
                        "baggage": {"duration": 0.8, "retries": 0},
                    },
                    {
                        "sessionId": session_id,
                        "testcaseName": "test_database_connection",
                        "testcaseClassname": "TestDatabase",
                        "testcaseFile": "tests/test_db.py",
                        "testsuite": "database",
                        "status": "fail",
                        "output": "Connection timeout after 5 seconds",
                        "baggage": {"duration": 5.2, "retries": 3},
                    },
                ]
            },
            headers={"X-API-Key": api_key},
        )
        assert testcases_response.status_code == 201, testcases_response.content

        testcases_list_response = client.get(
            f"{server_url}/api/v1/testcases",
            headers={"Authorization": f"Bearer {auth_tokens['access_token']}"},
        )
        assert testcases_list_response.status_code == 200
        testcases_list_data = testcases_list_response.json()

        testcase_names = {tc["name"] for tc in testcases_list_data["items"]}
        assert {
            "test_login_success",
            "test_login_failure",
            "test_database_connection",
        } == testcase_names
        assert testcases_list_data["aggregatedStatus"] == 1
