import uuid
from pathlib import Path

from lark import Lark, Transformer, exceptions

from .ast import (
    ClassnameQuery,
    ComparisonOperator,
    CompoundQuery,
    EmptyQuery,
    FileQuery,
    GroupByClause,
    GroupByToken,
    GroupByTokenType,
    IdQuery,
    LogicalOperator,
    NameQuery,
    Query,
    QueryWithGroupBy,
    SessionQuery,
    StatusQuery,
    TagQuery,
    TagValueQuery,
    TestcaseStatus,
    TestsuiteQuery,
)


class QueryParseError(Exception):
    pass


class QueryTransformer(Transformer):
    def quoted_string(self, s):
        escaped_string = str(s[0])
        return escaped_string[1:-1]

    def session_query(self, items):
        operator_token = str(items[1])
        operator = (
            ComparisonOperator.EQUALS
            if operator_token == "="
            else ComparisonOperator.NOT_EQUALS
        )
        session_id_str = items[2]
        if not session_id_str:
            raise ValueError("session_id cannot be empty")
        try:
            session_id = uuid.UUID(session_id_str)
        except ValueError:
            raise ValueError(f"Invalid UUID format for session_id: {session_id_str}")
        return SessionQuery(session_id=session_id, operator=operator)

    def tag_query(self, items):
        return items[0]

    def negated_tag_query(self, items):
        tag = items[0]
        if not tag:
            raise ValueError("Tag must be non-empty")
        return TagQuery(tag=tag, operator=ComparisonOperator.NOT_EQUALS)

    def regular_tag_query(self, items):
        tag = items[0]
        if not tag:
            raise ValueError("Tag must be non-empty")
        return TagQuery(tag=tag, operator=ComparisonOperator.EQUALS)

    def tag_assignment_query(self, items):
        tag = items[0]
        if not tag:
            raise ValueError("Tag must be non-empty")
        operator_token = str(items[1])
        operator = (
            ComparisonOperator.EQUALS
            if operator_token == "="
            else ComparisonOperator.NOT_EQUALS
        )
        value = items[2]
        return TagValueQuery(tag=tag, value=value, operator=operator)

    def id_query(self, items):
        operator_token = str(items[1])
        operator = (
            ComparisonOperator.EQUALS
            if operator_token == "="
            else ComparisonOperator.NOT_EQUALS
        )
        id_str = items[2]
        if not id_str:
            raise ValueError("id cannot be empty")
        try:
            id_value = uuid.UUID(id_str)
        except ValueError:
            raise ValueError(f"Invalid UUID format for id: {id_str}")
        return IdQuery(id=id_value, operator=operator)

    def name_query(self, items):
        operator_token = str(items[1])
        operator = (
            ComparisonOperator.EQUALS
            if operator_token == "="
            else ComparisonOperator.NOT_EQUALS
        )
        name_value = items[2]
        if not name_value:
            raise ValueError("Name must be non-empty")
        return NameQuery(name=name_value, operator=operator)

    def classname_query(self, items):
        operator_token = str(items[1])
        operator = (
            ComparisonOperator.EQUALS
            if operator_token == "="
            else ComparisonOperator.NOT_EQUALS
        )
        classname_value = items[2]
        return ClassnameQuery(classname=classname_value, operator=operator)

    def testsuite_query(self, items):
        operator_token = str(items[1])
        operator = (
            ComparisonOperator.EQUALS
            if operator_token == "="
            else ComparisonOperator.NOT_EQUALS
        )
        testsuite_value = items[2]
        return TestsuiteQuery(testsuite=testsuite_value, operator=operator)

    def file_query(self, items):
        operator_token = str(items[1])
        operator = (
            ComparisonOperator.EQUALS
            if operator_token == "="
            else ComparisonOperator.NOT_EQUALS
        )
        file_value = items[2]
        return FileQuery(file=file_value, operator=operator)

    def status_query(self, items):
        operator_token = str(items[1])
        operator = (
            ComparisonOperator.EQUALS
            if operator_token == "="
            else ComparisonOperator.NOT_EQUALS
        )
        status_value = items[2]
        if status_value:
            try:
                TestcaseStatus(status_value)
            except ValueError:
                valid_statuses = [status.value for status in TestcaseStatus]
                raise ValueError(
                    f"Invalid status '{status_value}'. Must be one of: {valid_statuses}"
                )
        return StatusQuery(status=status_value, operator=operator)

    def atomic_query(self, items):
        return items[0]

    def logical_op(self, items):
        op = items[0]
        if hasattr(op, "data"):
            op_name = op.data
        else:
            op_name = str(op)

        if "and" in op_name.lower():
            return LogicalOperator.AND
        elif "or" in op_name.lower():
            return LogicalOperator.OR
        else:
            raise ValueError(f"Unknown logical operator: {op_name}")

    def compound_query(self, items):
        if len(items) == 1:
            return items[0]

        queries = []
        operators = []

        for i in range(len(items)):
            if i % 2 == 0:
                queries.append(items[i])
            else:
                operators.append(items[i])

        if len(queries) < 2:
            raise ValueError("Compound query must have at least 2 queries")

        if len(operators) != len(queries) - 1:
            raise ValueError(
                f"Expected {len(queries) - 1} operators for {len(queries)} queries, got {len(operators)}"
            )

        return CompoundQuery(queries=queries, operators=operators)

    def empty_query(self, items):
        return EmptyQuery()

    def tag_token(self, items):
        tag_name = items[0]
        if not tag_name:
            raise ValueError("TAG tokens must have a non-empty value")
        return GroupByToken(token_type=GroupByTokenType.TAG, value=tag_name)

    def group_by_token(self, items):
        token = items[0]
        if hasattr(token, "token_type"):
            return token
        else:
            return GroupByToken(token_type=GroupByTokenType.SESSION_ID, value="")

    def group_by_tokens(self, items):
        return items

    def group_by_clause(self, items):
        tokens = items[0]
        if not tokens:
            raise ValueError("Group by clause must have at least one token")

        seen = set()
        for token in tokens:
            key = (token.token_type, token.value)
            if key in seen:
                raise ValueError(f"Duplicate group_by token: {token}")
            seen.add(key)

        return GroupByClause(tokens=tokens)

    def main_query(self, items):
        return items[0]

    def query(self, items):
        if len(items) == 1:
            return items[0]
        else:
            main_query = items[0]
            group_by = items[1]
            return QueryWithGroupBy(main_query=main_query, group_by=group_by)


class QueryParser:
    def __init__(self):
        grammar_path = Path(__file__).parent / "grammar.lark"
        with open(grammar_path, "r") as f:
            grammar = f.read()

        self.parser = Lark(grammar, start="start", parser="lalr")
        self.transformer = QueryTransformer()

    def parse(self, query_string: str) -> Query:
        try:
            tree = self.parser.parse(query_string.strip())
            query = self.transformer.transform(tree)
            return query
        except (exceptions.LarkError, ValueError) as e:
            raise QueryParseError(f"Failed to parse query '{query_string}': {e}")
