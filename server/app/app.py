from __future__ import annotations

import os

from advanced_alchemy.extensions.litestar import (
    AlembicAsyncConfig,
    AsyncSessionConfig,
    SQLAlchemyAsyncConfig,
)
from litestar.exceptions import ImproperlyConfiguredException

from app.app_init import make_app

APP_PATH = os.path.dirname(os.path.abspath(__file__))
ALEMBIC_PATH = os.path.join(APP_PATH, "..", "alembic")

DATABASE_URL = os.getenv("GREENER_DATABASE_URL")
if not DATABASE_URL:
    raise ImproperlyConfiguredException("GREENER_DATABASE_URL is not set")

TOKEN_SECRET = os.getenv("GREENER_JWT_SECRET")
if not TOKEN_SECRET:
    raise ImproperlyConfiguredException("GREENER_JWT_SECRET is not set")


alchemy_config = SQLAlchemyAsyncConfig(
    connection_string=DATABASE_URL,
    before_send_handler="autocommit",
    session_config=AsyncSessionConfig(expire_on_commit=False),
    alembic_config=AlembicAsyncConfig(
        script_location=ALEMBIC_PATH,
    ),
    create_all=True,
)

app = make_app(alchemy_config, TOKEN_SECRET)
