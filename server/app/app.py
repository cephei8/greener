from __future__ import annotations

import os

from advanced_alchemy.extensions.litestar import (
    AsyncSessionConfig,
    SQLAlchemyAsyncConfig,
)
from litestar.exceptions import ImproperlyConfiguredException
from sqlalchemy.engine import make_url

from app.app_init import make_app

DATABASE_URL = os.getenv("GREENER_DATABASE_URL")
if not DATABASE_URL:
    raise ImproperlyConfiguredException("GREENER_DATABASE_URL is not set")

url = make_url(DATABASE_URL)
if not url.database:
    raise ImproperlyConfiguredException(
        "GREENER_DATABASE_URL must have database part specified"
    )

driver_mapping = {"sqlite": "aiosqlite", "postgres": "asyncpg"}
driver = driver_mapping.get(url.database)
if not driver:
    raise ImproperlyConfiguredException(
        f"Database is not supported, required {list(driver_mapping.keys())}, got '{url.database}'"
    )

DATABASE_URL = url.set(drivername=f"{url.drivername}+{driver}").render_as_string(
    hide_password=False
)

TOKEN_SECRET = os.getenv("GREENER_JWT_SECRET")
if not TOKEN_SECRET:
    raise ImproperlyConfiguredException("GREENER_JWT_SECRET is not set")


alchemy_config = SQLAlchemyAsyncConfig(
    connection_string=DATABASE_URL,
    before_send_handler="autocommit",
    session_config=AsyncSessionConfig(expire_on_commit=False),
)

app = make_app(alchemy_config, TOKEN_SECRET)
