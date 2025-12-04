from __future__ import annotations

from typing import assert_never

from sqlalchemy import and_, or_, select

from app.models import Label, Session_, Testcase, TestcaseStatus
from app.query import (
    ClassnameQuery,
    ComparisonOperator,
    CompoundQuery,
    FileQuery,
    IdQuery,
    LogicalOperator,
    NameQuery,
    SessionQuery,
    SimpleQuery,
    StatusQuery,
    TagQuery,
    TagValueQuery,
    TestsuiteQuery,
)
from app.query.ast import EmptyQuery


def _convert_status_string_to_enum(status_str: str) -> TestcaseStatus:
    status_mapping = {
        "pass": TestcaseStatus.PASS,
        "fail": TestcaseStatus.FAIL,
        "error": TestcaseStatus.ERROR,
        "skip": TestcaseStatus.SKIP,
    }
    if status_str in status_mapping:
        return status_mapping[status_str]
    raise ValueError(f"Unknown status: {status_str}")


def build_query_conditions(query):
    if isinstance(query, EmptyQuery):
        return None

    def process_query(q):
        match q:
            case SessionQuery(
                operator=ComparisonOperator.EQUALS, session_id=session_id
            ):
                return Session_.id == session_id
            case SessionQuery(
                operator=ComparisonOperator.NOT_EQUALS, session_id=session_id
            ):
                return Session_.id != session_id
            case IdQuery(operator=ComparisonOperator.EQUALS, id=id_value):
                return Testcase.id == id_value
            case IdQuery(operator=ComparisonOperator.NOT_EQUALS, id=id_value):
                return Testcase.id != id_value
            case NameQuery(operator=ComparisonOperator.EQUALS, name=name_value):
                return Testcase.name == name_value
            case NameQuery(operator=ComparisonOperator.NOT_EQUALS, name=name_value):
                return Testcase.name != name_value
            case ClassnameQuery(
                operator=ComparisonOperator.EQUALS, classname=classname_value
            ):
                return Testcase.classname == classname_value
            case ClassnameQuery(
                operator=ComparisonOperator.NOT_EQUALS, classname=classname_value
            ):
                return Testcase.classname != classname_value
            case TestsuiteQuery(
                operator=ComparisonOperator.EQUALS, testsuite=testsuite_value
            ):
                return Testcase.testsuite == testsuite_value
            case TestsuiteQuery(
                operator=ComparisonOperator.NOT_EQUALS, testsuite=testsuite_value
            ):
                return Testcase.testsuite != testsuite_value
            case FileQuery(operator=ComparisonOperator.EQUALS, file=file_value):
                return Testcase.file == file_value
            case FileQuery(operator=ComparisonOperator.NOT_EQUALS, file=file_value):
                return Testcase.file != file_value
            case StatusQuery(operator=ComparisonOperator.EQUALS, status=status_value):
                enum_status = _convert_status_string_to_enum(status_value)
                return Testcase.status == enum_status
            case StatusQuery(
                operator=ComparisonOperator.NOT_EQUALS, status=status_value
            ):
                enum_status = _convert_status_string_to_enum(status_value)
                return Testcase.status != enum_status
            case TagValueQuery(
                operator=ComparisonOperator.EQUALS, tag=key, value=value
            ):
                return Testcase.session_id.in_(
                    select(Label.session_id).where(
                        and_(Label.key == key, Label.value == value)
                    )
                )
            case TagValueQuery(
                operator=ComparisonOperator.NOT_EQUALS, tag=key, value=value
            ):
                return Testcase.session_id.not_in(
                    select(Label.session_id).where(
                        and_(Label.key == key, Label.value == value)
                    )
                )
            case TagQuery(operator=ComparisonOperator.EQUALS, tag=key):
                return Testcase.session_id.in_(
                    select(Label.session_id).where(Label.key == key)
                )
            case TagQuery(operator=ComparisonOperator.NOT_EQUALS, tag=key):
                return Testcase.session_id.not_in(
                    select(Label.session_id).where(Label.key == key)
                )
            case _:  # pragma: no cover
                assert_never(q)  # pragma: no cover

    if isinstance(query, CompoundQuery):
        assert len(query.queries) > 1
        cond = process_query(query.queries[0])

        for i in range(1, len(query.queries)):
            next_cond = process_query(query.queries[i])
            op = query.operators[i - 1]

            match op:
                case LogicalOperator.AND:
                    cond = and_(cond, next_cond)
                case LogicalOperator.OR:
                    cond = or_(cond, next_cond)
                case _:  # pragma: no cover
                    assert_never(op)  # pragma: no cover

        return cond

    assert isinstance(query, SimpleQuery)
    return process_query(query)
