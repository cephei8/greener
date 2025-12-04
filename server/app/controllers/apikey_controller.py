from __future__ import annotations

import base64
import json
import secrets
from typing import Any
from uuid import UUID

from advanced_alchemy.filters import LimitOffset, OrderBy
from litestar import Controller, Request, delete, get, post
from litestar.exceptions import NotFoundException
from litestar.pagination import OffsetPagination
from litestar.security.jwt import Token
from sqlalchemy.ext.asyncio import AsyncSession

from app.models import (
    APIKey,
    APIKeyReadDTO,
    APIKeyRepository,
    APIKeyUnencrypted,
    APIKeyUnencryptedReadDTO,
    APIKeyWriteDTO,
    User,
)
from app.util import authenticate_api_key, hash_secret


class APIKeyController(Controller):
    path = "/api-keys"
    dto = APIKeyWriteDTO
    return_dto = APIKeyReadDTO
    tags = ["API Keys"]

    @post("/", summary="Create API key", return_dto=APIKeyUnencryptedReadDTO)
    async def create(
        self,
        data: APIKey,
        request: Request[User, Token, Any],
        apikey_repo: APIKeyRepository,
    ) -> APIKeyUnencrypted:
        secret = secrets.token_urlsafe(32)

        data.secret_salt = secrets.token_bytes(32)
        data.secret_hash = hash_secret(secret, data.secret_salt)
        data.user_id = request.user.id
        apikey = await apikey_repo.add(data)

        key_dict = {"apiKeyId": str(apikey.id), "apiKeySecret": secret}
        key = base64.b64encode(json.dumps(key_dict).encode()).decode()
        return APIKeyUnencrypted(
            id=apikey.id,
            key=key,
            description=apikey.description,
            created_at=apikey.created_at,
        )

    @get("/", summary="List API keys")
    async def list(
        self,
        request: Request[User, Token, Any],
        apikey_repo: APIKeyRepository,
        offset: int = 0,
        limit: int = 100,
    ) -> OffsetPagination[APIKey]:
        items, total = await apikey_repo.list_and_count(
            LimitOffset(limit=limit, offset=offset),
            OrderBy(field_name="created_at", sort_order="desc"),
            user_id=request.user.id,
        )

        return OffsetPagination[APIKey](
            items=items,
            total=total,
            limit=limit,
            offset=offset,
        )

    @get("/{id:uuid}", summary="Get API key")
    async def get(
        self,
        id: UUID,
        request: Request[User, Token, Any],
        apikey_repo: APIKeyRepository,
    ) -> APIKey:
        if apikey := await apikey_repo.get_one_or_none(id=id, user_id=request.user.id):
            return apikey
        raise NotFoundException()

    @delete("/{id:uuid}", summary="Delete API key")
    async def delete(
        self,
        id: UUID,
        request: Request[User, Token, Any],
        db_session: AsyncSession,
        apikey_repo: APIKeyRepository,
    ) -> None:
        if apikey := await apikey_repo.get_one_or_none(id=id, user_id=request.user.id):
            await apikey_repo.delete(apikey.id)
            await db_session.commit()
        else:
            raise NotFoundException()
