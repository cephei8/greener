import contextlib

import pytest
import testcontainers.mysql
import testcontainers.postgres
from advanced_alchemy.extensions.litestar import (
    AsyncSessionConfig,
    SQLAlchemyAsyncConfig,
)
from litestar.testing import AsyncTestClient
from sqlalchemy.ext.asyncio import async_sessionmaker, create_async_engine

from app.app_init import make_app
from app.models import METADATA, APIKey, Label, Session_, Testcase, TestcaseStatus, User
from app.util import hash_secret


def pytest_addoption(parser):
    parser.addoption(
        "--sqlite",
        action="store_true",
        default=True,
        dest="sqlite",
        help="run sqlite tests",
    )
    parser.addoption(
        "--no-sqlite",
        action="store_false",
        default=None,
        dest="sqlite",
        help="skip sqlite tests",
    )
    parser.addoption(
        "--postgres", action="store_true", default=False, help="run postgres tests"
    )
    parser.addoption(
        "--mysql", action="store_true", default=False, help="run mysql tests"
    )


def pytest_generate_tests(metafunc):
    if "db_conn" in metafunc.fixturenames:
        sqlite = metafunc.config.getoption("sqlite")
        postgres = metafunc.config.getoption("--postgres")
        mysql = metafunc.config.getoption("--mysql")

        params = []
        if sqlite:
            params.append("db_sqlite")
        if postgres:
            params.append("db_postgres")
        if mysql:
            params.append("db_mysql")

        metafunc.parametrize("db_conn", params, indirect=True)


@pytest.fixture
def db_sqlite(tmp_path):
    return f"sqlite+aiosqlite:///{tmp_path / 'testdb.sqlite'}"


@pytest.fixture
def db_postgres():
    with testcontainers.postgres.PostgresContainer("postgres:latest") as postgres:
        yield (
            "postgresql+asyncpg://"
            f"{postgres.username}:{postgres.password}@"
            f"{postgres.get_container_host_ip()}:{postgres.get_exposed_port(5432)}/"
            f"{postgres.dbname}"
        )


@pytest.fixture
def db_mysql():
    with testcontainers.mysql.MySqlContainer("mysql:latest") as mysql:
        yield (
            "mysql+asyncmy://"
            f"{mysql.username}:{mysql.password}@"
            f"{mysql.get_container_host_ip()}:{mysql.get_exposed_port(3306)}/"
            f"{mysql.dbname}"
        )


@pytest.fixture
def db_conn(request):
    return request.getfixturevalue(request.param)


@pytest.fixture
def test_app(db_conn):
    app = make_app(
        alchemy_config=SQLAlchemyAsyncConfig(
            connection_string=db_conn,
            session_config=AsyncSessionConfig(expire_on_commit=False),
            create_all=False,
        ),
        token_secret="abcd",
    )
    app.debug = True
    return app


@pytest.fixture
async def test_client(test_app):
    async with AsyncTestClient(app=test_app) as client:

        yield client


@pytest.fixture
async def db_engine(test_client, db_conn):
    _ = test_client

    engine = create_async_engine(db_conn, echo=False)
    async with engine.begin() as conn:
        await conn.run_sync(METADATA.create_all)

    yield engine

    await engine.dispose()


@pytest.fixture
def db_session_factory(db_engine):
    @contextlib.asynccontextmanager
    async def _make():
        async_session_maker = async_sessionmaker(bind=db_engine, expire_on_commit=False)
        async with async_session_maker() as db_session:
            yield db_session

    return _make


@pytest.fixture
async def user(db_session_factory):
    password_salt = b"test_salt_12345678901234567890123"
    password_hash = hash_secret("testpass123", password_salt)

    user = User(
        username="test_user",
        password_salt=password_salt,
        password_hash=password_hash,
    )

    async with db_session_factory() as db_session:
        db_session.add(user)
        await db_session.commit()
        await db_session.refresh(user)

    return user


@pytest.fixture
async def user_token(user, test_client):
    response = await test_client.post(
        "/api/v1/auth/login",
        json={
            "username": user.username,
            "password": "testpass123",
        },
    )
    assert response.status_code == 201
    return response.json()["accessToken"]


@pytest.fixture
def auth_headers(user_token):
    return {"Authorization": f"Bearer {user_token}"}


@pytest.fixture
def session_factory(user, db_session_factory):
    async def _create():
        session = Session_(
            user_id=user.id,
        )

        async with db_session_factory() as db_session:
            db_session.add(session)
            await db_session.commit()
            await db_session.refresh(session)

        return session

    return _create


@pytest.fixture
def testcase_factory(db_session_factory):
    async def _create(
        session,
        *,
        status=TestcaseStatus.PASS,
        name="test_api",
        classname=None,
        file=None,
    ):
        testcase = Testcase(
            status=status,
            name=name,
            classname=classname,
            file=file,
            user_id=session.user_id,
            session_id=session.id,
        )

        async with db_session_factory() as db_session:
            db_session.add(testcase)
            await db_session.commit()
            await db_session.refresh(testcase)

        return testcase

    return _create


@pytest.fixture
def label_factory(db_session_factory):
    async def _create(session, key, value=None):
        label = Label(
            key=key,
            value=value,
            user_id=session.user_id,
            session_id=session.id,
        )

        async with db_session_factory() as db_session:
            db_session.add(label)
            await db_session.commit()
            await db_session.refresh(label)

        return label

    return _create


@pytest.fixture
def apikey_factory(user, db_session_factory):
    async def _create(description="Test API Key", secret="test_secret_12345"):

        secret_salt = b"test_salt_12345678901234567890123"
        secret_hash = hash_secret(secret, secret_salt)

        apikey = APIKey(
            description=description,
            secret_salt=secret_salt,
            secret_hash=secret_hash,
            user_id=user.id,
        )

        async with db_session_factory() as db_session:
            db_session.add(apikey)
            await db_session.commit()
            await db_session.refresh(apikey)

        return apikey

    return _create
