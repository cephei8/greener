from __future__ import annotations

from datetime import timedelta
from os import environ

from advanced_alchemy.extensions.litestar import (
    SQLAlchemyAsyncConfig,
    SQLAlchemyPlugin,
)
from litestar import Litestar, Request, Response
from litestar.config.cors import CORSConfig
from litestar.di import Provide
from litestar.logging import LoggingConfig
from litestar.openapi.config import OpenAPIConfig
from litestar.openapi.plugins import ScalarRenderPlugin
from litestar.security.jwt import JWTAuth
from litestar.status_codes import HTTP_400_BAD_REQUEST
from pydantic_core import ValidationError
from sqlalchemy.ext.asyncio import AsyncSession

from app.controllers import (
    APIKeyController,
    AuthController,
    GroupController,
    IngressController,
    LabelController,
    ReadyController,
    SessionController,
    TestcaseController,
    retrieve_user_handler,
)
from app.models import (
    APIKeyRepository,
    LabelRepository,
    SessionRepository,
    TestcaseRepository,
    User,
    UserRepository,
)


async def provide_user_repo(db_session: AsyncSession) -> UserRepository:
    return UserRepository(session=db_session)


async def provide_apikey_repo(db_session: AsyncSession) -> APIKeyRepository:
    return APIKeyRepository(session=db_session)


async def provide_label_repo(db_session: AsyncSession) -> LabelRepository:
    return LabelRepository(session=db_session)


async def provide_testcase_repo(db_session: AsyncSession) -> TestcaseRepository:
    return TestcaseRepository(session=db_session)


async def provide_session_repo(db_session: AsyncSession) -> SessionRepository:
    return SessionRepository(session=db_session)


def pydantic_validation_exception_handler(
    request: Request, exc: ValidationError
) -> Response:
    return Response(
        content={"detail": str(exc), "errors": exc.errors()},
        status_code=HTTP_400_BAD_REQUEST,
    )


def make_app(alchemy_config: SQLAlchemyAsyncConfig, token_secret: str) -> Litestar:
    jwt_auth = JWTAuth[User](
        retrieve_user_handler=retrieve_user_handler,
        token_secret=token_secret,
        default_token_expiration=timedelta(hours=1),
        exclude=[
            "/api/v1/auth/login",
            "/api/v1/auth/refresh",
            "/api/v1/ingress/sessions",
            "/api/v1/ingress/testcases",
            "/api/v1/ready",
        ],
    )

    async def provide_jwt_auth() -> JWTAuth[User]:
        return jwt_auth

    app = Litestar(
        path="/api/v1",
        route_handlers=[
            AuthController,
            APIKeyController,
            ReadyController,
            SessionController,
            TestcaseController,
            GroupController,
            LabelController,
            IngressController,
        ],
        dependencies={
            "user_repo": Provide(provide_user_repo),
            "apikey_repo": Provide(provide_apikey_repo),
            "label_repo": Provide(provide_label_repo),
            "testcase_repo": Provide(provide_testcase_repo),
            "session_repo": Provide(provide_session_repo),
            "jwt_auth": Provide(provide_jwt_auth),
        },
        plugins=[SQLAlchemyPlugin(config=alchemy_config)],
        logging_config=LoggingConfig(
            root={"level": "INFO", "handlers": ["queue_listener"]},
            log_exceptions="always",
            disable_stack_trace={400, 401, 404},
        ),
        cors_config=CORSConfig(
            allow_origins=["*"],
            allow_methods=["GET", "POST", "PUT", "DELETE", "OPTIONS"],
            allow_headers=["*"],
            allow_credentials=True,
        ),
        openapi_config=OpenAPIConfig(
            title="Greener server",
            version="0.0.1",
            render_plugins=[ScalarRenderPlugin()],
        ),
        exception_handlers={
            ValidationError: pydantic_validation_exception_handler,
        },
        on_app_init=[jwt_auth.on_app_init],
    )
    return app
