from __future__ import annotations

from typing import Any
from uuid import UUID

from advanced_alchemy.filters import LimitOffset, OrderBy
from litestar import Controller, Request, get
from litestar.exceptions import NotFoundException
from litestar.pagination import OffsetPagination
from litestar.security.jwt import Token

from app.models import (
    Session_,
    SessionReadDTO,
    SessionRepository,
    User,
)


class SessionController(Controller):
    path = "/sessions"
    return_dto = SessionReadDTO
    tags = ["Sessions"]

    @get("/", summary="List sessions")
    async def list(
        self,
        request: Request[User, Token, Any],
        session_repo: SessionRepository,
        offset: int = 0,
        limit: int = 100,
    ) -> OffsetPagination[Session_]:
        items, total = await session_repo.list_and_count(
            LimitOffset(limit=limit, offset=offset),
            OrderBy(field_name="created_at", sort_order="desc"),
            user_id=request.user.id,
        )

        return OffsetPagination[Session_](
            items=items,
            total=total,
            limit=limit,
            offset=offset,
        )

    @get("/{id:uuid}", summary="Get session")
    async def get(
        self,
        id: UUID,
        request: Request[User, Token, Any],
        session_repo: SessionRepository,
    ) -> Session_:
        if session := await session_repo.get_one_or_none(
            id=id, user_id=request.user.id
        ):
            return session
        raise NotFoundException()
