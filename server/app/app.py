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

DATABASE_DRIVER = os.getenv("GREENER_DATABASE_DRIVER")
if DATABASE_DRIVER:
    url = make_url(DATABASE_URL)
    url = url.set(drivername=f"{url.drivername}+{DATABASE_DRIVER}")
    DATABASE_URL = url.render_as_string(hide_password=False)

TOKEN_SECRET = os.getenv("GREENER_JWT_SECRET")
if not TOKEN_SECRET:
    raise ImproperlyConfiguredException("GREENER_JWT_SECRET is not set")


alchemy_config = SQLAlchemyAsyncConfig(
    connection_string=DATABASE_URL,
    before_send_handler="autocommit",
    session_config=AsyncSessionConfig(expire_on_commit=False),
)

app = make_app(alchemy_config, TOKEN_SECRET)
