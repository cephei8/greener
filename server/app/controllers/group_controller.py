from __future__ import annotations

from datetime import datetime
from typing import Any, assert_never

from litestar import Controller, Request, get
from litestar.exceptions import ValidationException
from litestar.params import Parameter
from litestar.security.jwt import Token
from sqlalchemy import and_, func, select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import aliased

from app.controllers.util import build_query_conditions
from app.models import (
    GroupItem,
    GroupOffsetPagination,
    GroupOffsetPaginationDTO,
    Label,
    QueryValidationResult,
    QueryValidationResultDTO,
    Session_,
    Testcase,
    User,
)
from app.query import (
    QueryParseError,
    QueryParser,
    QueryWithGroupBy,
)
from app.query.ast import GroupByTokenType


class GroupController(Controller):
    path = "/groups"
    tags = ["Groups"]

    @get(
        "/validate-query", summary="Validate query", return_dto=QueryValidationResultDTO
    )
    async def validate_query(
        self,
        query_str: str = Parameter(query="queryStr", default=""),
    ) -> QueryValidationResult:

        try:
            parser = QueryParser()
            parsed_query = parser.parse(query_str)
            return QueryValidationResult(
                is_grouping=isinstance(parsed_query, QueryWithGroupBy)
            )
        except QueryParseError as e:
            raise ValidationException(f"Invalid query: {e}") from e

    @get("/", summary="Execute query", return_dto=GroupOffsetPaginationDTO)
    async def list(
        self,
        request: Request[User, Token, Any],
        db_session: AsyncSession,
        query_str: str = Parameter(query="queryStr", default=""),
        offset: int = 0,
        limit: int = 10,
        start_date: datetime | None = Parameter(query="startDate", default=None),
        end_date: datetime | None = Parameter(query="endDate", default=None),
    ) -> GroupOffsetPagination:
        try:
            parser = QueryParser()
            query_ast = parser.parse(query_str)
        except QueryParseError as e:
            raise ValidationException(f"Invalid query: {e}") from e

        if not isinstance(query_ast, QueryWithGroupBy):
            return GroupOffsetPagination(
                items=[],
                total=0,
                limit=limit,
                offset=offset,
                header=None,
                aggregated_status=None,
            )

        group_tables = []
        group_columns = []
        group_column_labels = []
        group_column_headers = []
        for i, token in enumerate(query_ast.group_by.tokens):
            match token.token_type:
                case GroupByTokenType.SESSION_ID:
                    group_tables.append(Session_)
                    group_columns.append(Session_.id)
                    group_column_labels.append(
                        f"{Session_.__table__}__{Session_.id}__{i}"
                    )
                    group_column_headers.append("session_id")
                case GroupByTokenType.TAG:
                    label = token.value
                    label_tbl = aliased(Label, name=f"{Label.__table__}__{i}")
                    group_tables.append(label_tbl)
                    group_columns.append(label_tbl.value)
                    group_column_labels.append(f"{Label.__table__}__{i}__{Label.value}")
                    group_column_headers.append(f'#"{label}"')
                case _:  # pragma: no cover
                    assert_never(token.token_type)  # pragma: no cover

        select_columns = [
            c.label(la) for c, la in zip(group_columns, group_column_labels)
        ]
        cte_query = select(
            *select_columns, func.min(Testcase.status).label("group_status")
        ).select_from(Testcase)

        for token, tbl in zip(query_ast.group_by.tokens, group_tables):
            match token.token_type:
                case GroupByTokenType.SESSION_ID:
                    cte_query = cte_query.join(tbl, Testcase.session_id == tbl.id)
                case GroupByTokenType.TAG:
                    cte_query = cte_query.join(
                        tbl,
                        and_(
                            Testcase.session_id == tbl.session_id,
                            tbl.key == token.value,
                        ),
                    )
                case _:  # pragma: no cover
                    assert_never(token.token_type)  # pragma: no cover

        cte_query = (
            cte_query.where(and_(Testcase.user_id == request.user.id))
            .group_by(*group_columns)
            .order_by(*group_columns)
        )

        where_cond = build_query_conditions(query_ast.main_query)
        if where_cond is not None:
            cte_query = cte_query.where(where_cond)
        if start_date is not None:
            cte_query = cte_query.where(Testcase.created_at >= start_date)
        if end_date is not None:
            cte_query = cte_query.where(Testcase.created_at < end_date)

        cte = cte_query.cte("cte")
        query = (
            select(
                cte,
                func.count(1).over().label("total_count"),
                func.min(cte.c.group_status).over().label("aggregated_status"),
            )
            .select_from(cte)
            .offset(offset)
            .limit(limit)
        )

        query_str = str(query)

        result = await db_session.execute(query)

        aggregated_status = None
        total_count = None
        header = None
        items = []
        for row in result:
            if aggregated_status is None:
                aggregated_status = row.aggregated_status
            if total_count is None:
                total_count = row.total_count
            if header is None:
                header = [str(c) for c in group_column_headers]
            status = row.group_status
            columns = [
                str(value) if (value := getattr(row, la)) is not None else None
                for la in group_column_labels
            ]

            items.append(
                GroupItem(
                    columns=columns,
                    status=status,
                )
            )

        return GroupOffsetPagination(
            items=items,
            total=total_count or 0,
            limit=limit,
            offset=offset,
            header=header,
            aggregated_status=aggregated_status,
        )
