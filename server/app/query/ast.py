import uuid
from abc import ABC
from dataclasses import dataclass
from enum import Enum, auto
from typing import List


class ComparisonOperator(Enum):
    EQUALS = "="
    NOT_EQUALS = "!="


class Query(ABC):
    pass


class SimpleQuery(Query):
    pass


@dataclass
class SessionQuery(SimpleQuery):
    session_id: uuid.UUID
    operator: ComparisonOperator


@dataclass
class TagQuery(SimpleQuery):
    tag: str
    operator: ComparisonOperator


@dataclass
class TagValueQuery(SimpleQuery):
    tag: str
    value: str
    operator: ComparisonOperator


class TestcaseStatus(Enum):
    __test__ = False

    PASS = "pass"
    FAIL = "fail"
    ERROR = "error"
    SKIP = "skip"


@dataclass
class IdQuery(SimpleQuery):
    id: uuid.UUID
    operator: ComparisonOperator


@dataclass
class NameQuery(SimpleQuery):
    name: str
    operator: ComparisonOperator


@dataclass
class ClassnameQuery(SimpleQuery):
    classname: str
    operator: ComparisonOperator


@dataclass
class TestsuiteQuery(SimpleQuery):
    __test__ = False

    testsuite: str
    operator: ComparisonOperator


@dataclass
class FileQuery(SimpleQuery):
    file: str
    operator: ComparisonOperator


@dataclass
class StatusQuery(SimpleQuery):
    status: str
    operator: ComparisonOperator


@dataclass
class EmptyQuery(SimpleQuery):
    pass


class GroupByTokenType(Enum):
    SESSION_ID = "session_id"
    TAG = "tag"


@dataclass
class GroupByToken:
    token_type: GroupByTokenType
    value: str


@dataclass
class GroupByClause:
    tokens: List[GroupByToken]


@dataclass
class QueryWithGroupBy(SimpleQuery):
    main_query: Query
    group_by: GroupByClause


class LogicalOperator(Enum):
    AND = auto()
    OR = auto()


@dataclass
class CompoundQuery(Query):
    queries: List[Query]
    operators: List[LogicalOperator]
