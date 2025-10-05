from __future__ import annotations

from typing import Any
from uuid import UUID

from litestar import Controller, Request, get
from litestar.exceptions import NotFoundException
from litestar.pagination import OffsetPagination
from litestar.security.jwt import Token
from sqlalchemy import func, select
from sqlalchemy.ext.asyncio import AsyncSession

from app.models import (
    Label,
    LabelReadDTO,
    Session_,
    User,
)


class LabelController(Controller):
    path = "/labels"
    return_dto = LabelReadDTO
    tags = ["Labels"]

    @get("/", summary="List labels")
    async def list(
        self,
        request: Request[User, Token, Any],
        db_session: AsyncSession,
        session_id: UUID,
        offset: int = 0,
        limit: int = 100,
    ) -> OffsetPagination[Label]:
        session_query = select(Session_).where(
            Session_.id == session_id, Session_.user_id == request.user.id
        )
        session_result = await db_session.execute(session_query)
        session = session_result.scalar_one_or_none()

        if not session:
            raise NotFoundException("Session not found")

        labels_query = (
            select(Label)
            .where(Label.session_id == session.id)
            .offset(offset)
            .limit(limit)
            .order_by(Label.created_at.asc())
        )

        labels_result = await db_session.execute(labels_query)
        items = list(labels_result.scalars().all())

        count_query = select(func.count(Label.id)).where(Label.session_id == session.id)
        count_result = await db_session.execute(count_query)
        total = count_result.scalar() or 0

        return OffsetPagination[Label](
            items=items,
            total=total,
            limit=limit,
            offset=offset,
        )
