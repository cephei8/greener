from __future__ import annotations

import json
from datetime import datetime
from typing import Any, assert_never
from urllib.parse import unquote
from uuid import UUID

from advanced_alchemy.filters import LimitOffset, OrderBy
from litestar import Controller, Request, get
from litestar.exceptions import NotFoundException, ValidationException
from litestar.pagination import OffsetPagination
from litestar.params import Parameter
from litestar.security.jwt import Token
from sqlalchemy import and_, func, select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import aliased

from app.controllers.util import build_query_conditions
from app.models import (
    Label,
    Session_,
    Testcase,
    TestcaseOffsetPagination,
    TestcaseReadDTO,
    TestcaseRepository,
    User,
)
from app.query import QueryParseError, QueryParser, QueryWithGroupBy
from app.query.ast import GroupByTokenType


class TestcaseController(Controller):
    path = "/testcases"
    return_dto = TestcaseReadDTO
    tags = ["Testcases"]

    @get("/", summary="List testcases")
    async def list(
        self,
        request: Request[User, Token, Any],
        db_session: AsyncSession,
        offset: int = 0,
        limit: int = 100,
        query_str: str | None = Parameter(query="queryStr", default=None),
        start_date: datetime | None = Parameter(query="startDate", default=None),
        end_date: datetime | None = Parameter(query="endDate", default=None),
        group: str | None = None,
    ) -> OffsetPagination[Testcase]:
        cte_query = select(Testcase).where(Testcase.user_id == request.user.id)

        query_ast = None
        if query_str:
            try:
                parser = QueryParser()
                query_ast = parser.parse(query_str)
            except QueryParseError as e:
                raise ValidationException(f"Invalid query: {e}") from e

        self._validate_query_group_relationship(query_ast, group)

        if query_ast:
            where_cond = build_query_conditions(
                query_ast.main_query if hasattr(query_ast, "main_query") else query_ast
            )
            if where_cond is not None:
                cte_query = cte_query.where(where_cond)

        if start_date is not None:
            cte_query = cte_query.where(Testcase.created_at >= start_date)
        if end_date is not None:
            cte_query = cte_query.where(Testcase.created_at < end_date)

        if group:
            try:
                group_decoded = unquote(group)
                parsed_group = json.loads(group_decoded)
                group_keys, group_values = self._validate_group_format(parsed_group)
                cte_query = self._apply_group_filter(
                    cte_query, group_keys, group_values
                )
            except (json.JSONDecodeError, ValueError) as e:
                raise ValidationException(f"Invalid group identifier: {e}") from e

        cte_query = cte_query.order_by(Testcase.created_at.desc())
        cte = cte_query.cte("cte")
        query = (
            select(
                cte,
                func.count(1).over().label("total_count"),
                func.min(cte.c.status).over().label("aggregated_status"),
            )
            .select_from(cte)
            .offset(offset)
            .limit(limit)
        )

        aggregated_status = None
        total_count = None
        items = []

        testcase_cols = Testcase.__table__.columns.keys()

        result = await db_session.execute(query)
        for row in result:
            if aggregated_status is None:
                aggregated_status = row.aggregated_status
            if total_count is None:
                total_count = row.total_count
            mapping = row._mapping
            testcase_data = {k: mapping[k] for k in testcase_cols if k in mapping}
            item = Testcase(**testcase_data)
            items.append(item)

        return TestcaseOffsetPagination(
            items=items,
            total=total_count or 0,
            limit=limit,
            offset=offset,
            aggregatedStatus=aggregated_status,
        )

    @get("/{id:uuid}", summary="Get testcase")
    async def get(
        self,
        id: UUID,
        request: Request[User, Token, Any],
        testcase_repo: TestcaseRepository,
    ) -> Testcase:
        if testcase := await testcase_repo.get_one_or_none(
            id=id, user_id=request.user.id
        ):
            return testcase
        raise NotFoundException()

    def _apply_group_filter(
        self, query, group_keys: list[str], group_values: list[str | None]
    ):
        for i, (key, value) in enumerate(zip(group_keys, group_values)):
            if key == "session_id":
                query = query.join(Session_, Testcase.session_id == Session_.id)
                query = query.where(Session_.id == value)
            elif key.startswith('#"') and key.endswith('"'):
                tag_name = key[2:-1]
                label_alias = aliased(Label, name=f"label_{i}")
                if value is None:
                    query = query.join(
                        label_alias,
                        and_(
                            Testcase.session_id == label_alias.session_id,
                            label_alias.key == tag_name,
                            label_alias.value.is_(None),
                        ),
                    )
                else:
                    query = query.join(
                        label_alias,
                        and_(
                            Testcase.session_id == label_alias.session_id,
                            label_alias.key == tag_name,
                            label_alias.value == value,
                        ),
                    )
        return query

    @staticmethod
    def _validate_group_format(parsed_group) -> tuple[list[str], list[str]]:
        if not isinstance(parsed_group, (list, tuple)) or len(parsed_group) != 2:
            raise ValueError(
                "Group identifier must be a tuple/array with exactly 2 elements"
            )

        group_keys, group_values = parsed_group

        if not isinstance(group_keys, list) or not isinstance(group_values, list):
            raise ValueError("Both elements must be arrays/lists")

        if len(group_keys) != len(group_values):
            raise ValueError("Group keys and values must have the same length")

        if not all(isinstance(key, str) for key in group_keys):
            raise ValueError("All group keys must be strings")

        if not all(isinstance(value, (str, type(None))) for value in group_values):
            raise ValueError("All group values must be strings or None")

        return group_keys, group_values

    def _validate_query_group_relationship(self, query_ast, group: str | None) -> None:
        is_grouping_query = query_ast is not None and isinstance(
            query_ast, QueryWithGroupBy
        )
        has_group = group is not None and group.strip() != ""

        if is_grouping_query and not has_group:
            raise ValidationException(
                "Group parameter is required when using a grouping query"
            )

        if not is_grouping_query and has_group:
            raise ValidationException(
                "Group parameter can only be used with grouping queries"
            )

        if is_grouping_query and has_group:
            try:
                group_decoded = unquote(group)
                parsed_group = json.loads(group_decoded)
                provided_keys, _ = self._validate_group_format(parsed_group)
                expected_keys = self._extract_group_keys_from_query(query_ast)

                if provided_keys != expected_keys:
                    raise ValidationException(
                        f"Group keys {provided_keys} do not match the grouping query keys {expected_keys}"
                    )
            except (json.JSONDecodeError, ValueError) as e:
                raise ValidationException(f"Invalid group identifier: {e}") from e

    @staticmethod
    def _extract_group_keys_from_query(query_ast: QueryWithGroupBy) -> list[str]:
        group_keys = []
        for token in query_ast.group_by.tokens:
            match token.token_type:
                case GroupByTokenType.SESSION_ID:
                    group_keys.append("session_id")
                case GroupByTokenType.TAG:
                    label = token.value
                    group_keys.append(f'#"{label}"')
                case _:  # pragma: no cover
                    assert_never(token.token_type)  # pragma: no cover
        return group_keys
