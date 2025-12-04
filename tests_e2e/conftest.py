import sqlite3
import subprocess
import time
from pathlib import Path

import httpx
import pytest
import testcontainers.mysql
import testcontainers.postgres
from testcontainers.core.container import DockerContainer
from testcontainers.core.network import Network

E2E_TESTS_PATH = Path(__file__).parent
REPO_ROOT_PATH = E2E_TESTS_PATH.parent


def pytest_addoption(parser):
    parser.addoption(
        "--skip-sqlite",
        action="store_true",
        default=False,
        help="Skip SQLite tests",
    )


def pytest_generate_tests(metafunc):
    if "db_conn" in metafunc.fixturenames:
        db_params = ["db_sqlite", "db_postgres", "db_mysql"]

        if metafunc.config.getoption("--skip-sqlite"):
            db_params = [p for p in db_params if p != "db_sqlite"]

        metafunc.parametrize("db_conn", db_params, indirect=True)


@pytest.fixture
def image_tag():
    return "greener-test:latest"


@pytest.fixture
def network():
    with Network() as n:
        yield n


@pytest.fixture
def db_sqlite(tmp_path):
    db_name = "testdb.sqlite"

    host_url = f"sqlite:///{tmp_path / db_name}"
    network_url = f"sqlite:////data/{db_name}"

    conn = sqlite3.connect(tmp_path / db_name)
    conn.close()

    yield {
        "host_url": host_url,
        "network_url": network_url,
    }


@pytest.fixture
def db_postgres(network):
    with testcontainers.postgres.PostgresContainer("postgres:latest").with_network(
        network
    ).with_network_aliases("db").with_exposed_ports(5432) as postgres:
        network_url = (
            "postgres://"
            f"{postgres.username}:{postgres.password}@"
            "db:5432/"
            f"{postgres.dbname}?sslmode=disable"
        )

        host_url = (
            "postgres://"
            f"{postgres.username}:{postgres.password}@"
            f"{postgres.get_container_host_ip()}:{postgres.get_exposed_port(5432)}/"
            f"{postgres.dbname}?sslmode=disable"
        )

        yield {
            "network_url": network_url,
            "host_url": host_url,
        }


@pytest.fixture
def db_mysql(network):
    with testcontainers.mysql.MySqlContainer("mysql:latest").with_network(
        network
    ).with_network_aliases("db").with_exposed_ports(3306) as mysql:
        network_url = (
            "mysql://"
            f"{mysql.username}:{mysql.password}@"
            "db:3306/"
            f"{mysql.dbname}"
        )

        host_url = (
            "mysql://"
            f"{mysql.username}:{mysql.password}@"
            f"{mysql.get_container_host_ip()}:{mysql.get_exposed_port(3306)}/"
            f"{mysql.dbname}"
        )

        yield {
            "network_url": network_url,
            "host_url": host_url,
        }


@pytest.fixture
def db_conn(request):
    return request.getfixturevalue(request.param)


@pytest.fixture
def jwt_secret():
    return "test-jwt-secret-key"


def wait_for_ready(url: str, timeout: int = 120, interval: float = 1.0) -> None:
    start_time = time.time()
    last_error = None

    while time.time() - start_time < timeout:
        try:
            with httpx.Client() as client:
                response = client.get(f"{url}/api/v1/ready", timeout=5)
                if response.status_code == 200:
                    return
                last_error = f"HTTP {response.status_code}: {response.text}"
        except httpx.RequestError as e:
            last_error = str(e)

        time.sleep(interval)

    raise TimeoutError(
        f"Server did not become ready within {timeout} seconds. "
        f"Last error: {last_error}"
    )


@pytest.fixture
def server(tmp_path, image_tag, db_conn, jwt_secret, network):
    with DockerContainer(image_tag).with_network(network).with_volume_mapping(
        str(tmp_path), "/data", mode="rw"
    ).with_exposed_ports(5096).with_env(
        "GREENER_DATABASE_URL", db_conn["network_url"]
    ).with_env(
        "GREENER_AUTH_SECRET", jwt_secret
    ) as server:
        server_url = (
            f"http://{server.get_container_host_ip()}:{server.get_exposed_port(5096)}"
        )
        wait_for_ready(server_url)
        yield server


@pytest.fixture
def server_url(server):
    return f"http://{server.get_container_host_ip()}:{server.get_exposed_port(5096)}"


@pytest.fixture
def admin_cli_path():
    return REPO_ROOT_PATH / "util" / "greener-admin-cli"


@pytest.fixture
def user(server, db_conn, admin_cli_path):
    _ = server

    username = "testuser"
    password = "testpass123"

    result = subprocess.run(
        [
            "go",
            "run",
            ".",
            "--db-url",
            db_conn["host_url"],
            "create-user",
            "--username",
            username,
            "--password",
            password,
        ],
        cwd=admin_cli_path,
        capture_output=True,
        text=True,
    )

    if result.returncode != 0:
        raise RuntimeError(f"Failed to create user: {result.stderr}")

    return {"username": username, "password": password}


@pytest.fixture
def auth_tokens(user, server, server_url):
    time.sleep(2)

    with httpx.Client() as client:
        response = client.post(
            f"{server_url}/api/v1/auth/login",
            json={
                "username": user["username"],
                "password": user["password"],
            },
        )

        if response.status_code != 201:
            try:
                logs = server.get_wrapped_container().logs().decode("utf-8")
                print(logs)
            except Exception as log_err:
                print(f"Could not get logs: {log_err}")
            response.raise_for_status()

        data = response.json()

        return {
            "access_token": data["accessToken"],
            "refresh_token": data["refreshToken"],
        }


@pytest.fixture
def api_key(server_url, auth_tokens):
    with httpx.Client() as client:
        response = client.post(
            f"{server_url}/api/v1/api-keys",
            json={"description": "Test API Key"},
            headers={"Authorization": f"Bearer {auth_tokens['access_token']}"},
        )
        response.raise_for_status()
        data = response.json()

        return data["key"]
