from __future__ import annotations

import secrets
from datetime import datetime, timedelta, timezone
from typing import Any

import advanced_alchemy.exceptions
from litestar import Controller, Request, Response, post
from litestar.connection import ASGIConnection
from litestar.datastructures import State
from litestar.exceptions import NotAuthorizedException
from litestar.security.jwt import JWTAuth, Token
from sqlalchemy.ext.asyncio import AsyncSession

from app.models import (
    ChangePasswordData,
    ChangePasswordDataDTO,
    LoginData,
    LoginDataDTO,
    RefreshData,
    RefreshDataDTO,
    TokenData,
    TokenDataDTO,
    User,
    UserRepository,
)
from app.util import hash_secret


async def retrieve_user_handler(
    token: Token, connection: ASGIConnection
) -> User | None:
    token_type = token.extras.get("type") if token.extras else None
    if token_type == "refresh":
        return None

    return User(id=token.sub)


def create_tokens(user_id: str, jwt_auth: JWTAuth[User]) -> TokenData:
    access_token_expiration = timedelta(hours=1)
    access_token = jwt_auth.create_token(
        identifier=user_id,
        token_expiration=access_token_expiration,
    )

    refresh_token_expiration = timedelta(days=7)
    refresh_token = jwt_auth.create_token(
        identifier=user_id,
        token_expiration=refresh_token_expiration,
        token_extras={"type": "refresh"},
    )

    return TokenData(
        access_token=access_token,
        access_token_expires_at=datetime.now(timezone.utc) + access_token_expiration,
        refresh_token=refresh_token,
        refresh_token_expires_at=datetime.now(timezone.utc) + refresh_token_expiration,
    )


class AuthController(Controller):
    path = "/auth"
    tags = ["Auth"]

    @post("/login", dto=LoginDataDTO, return_dto=TokenDataDTO)
    async def login(
        self,
        data: LoginData,
        user_repo: UserRepository,
        request: ASGIConnection,
    ) -> TokenData:
        try:
            user = await user_repo.get_one(username=data.username)
        except advanced_alchemy.exceptions.NotFoundError:
            raise NotAuthorizedException()

        password_hash = hash_secret(data.password, user.password_salt)
        if not secrets.compare_digest(password_hash, user.password_hash):
            raise NotAuthorizedException()

        jwt_auth = await request.app.dependencies["jwt_auth"].dependency()
        return create_tokens(str(user.id), jwt_auth)

    @post("/refresh", dto=RefreshDataDTO, return_dto=TokenDataDTO)
    async def refresh(
        self,
        data: RefreshData,
        request: ASGIConnection,
    ) -> TokenData:
        jwt_auth = await request.app.dependencies["jwt_auth"].dependency()

        try:
            decoded = Token.decode(
                encoded_token=data.refresh_token,
                secret=jwt_auth.token_secret,
                algorithm=jwt_auth.algorithm,
            )
        except Exception as e:
            raise NotAuthorizedException(f"Invalid refresh token: {e}")

        token_type = decoded.extras.get("type") if decoded.extras else None
        if token_type != "refresh":
            raise NotAuthorizedException("Invalid token type")

        return create_tokens(decoded.sub, jwt_auth)

    @post("/change-password", dto=ChangePasswordDataDTO)
    async def change_password(
        self,
        data: ChangePasswordData,
        request: Request[User, Token, Any],
        user_repo: UserRepository,
    ) -> None:
        user = await user_repo.get_one(id=request.user.id)

        old_password_hash = hash_secret(data.password_old, user.password_salt)
        if not secrets.compare_digest(old_password_hash, user.password_hash):
            raise NotAuthorizedException("Current password is incorrect")

        new_salt = secrets.token_bytes(32)
        new_password_hash = hash_secret(data.password_new, new_salt)

        user.password_salt = new_salt
        user.password_hash = new_password_hash
        await user_repo.update(user)

        return None

    @post("/logout")
    async def logout(self) -> None:
        return None
