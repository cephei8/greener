import uuid

import pytest

from app.query import (
    ClassnameQuery,
    ComparisonOperator,
    CompoundQuery,
    EmptyQuery,
    FileQuery,
    GroupByTokenType,
    IdQuery,
    LogicalOperator,
    NameQuery,
    QueryParseError,
    QueryParser,
    QueryWithGroupBy,
    SessionQuery,
    StatusQuery,
    TagQuery,
    TagValueQuery,
    TestsuiteQuery,
)


def test_parse_error_handling():
    parser = QueryParser()
    with pytest.raises(QueryParseError):
        parser.parse("invalid syntax here")


def test_parse_error_handling_complex():
    parser = QueryParser()
    with pytest.raises(QueryParseError):
        parser.parse("status = AND group_by")


def test_parse_error_handling_incomplete():
    parser = QueryParser()
    with pytest.raises(QueryParseError):
        parser.parse("status =")


def test_parse_error_handling_unbalanced():
    parser = QueryParser()
    with pytest.raises(QueryParseError):
        parser.parse('status = "unclosed quote')


def test_empty_query_string():
    parser = QueryParser()
    result = parser.parse("")
    assert isinstance(result, EmptyQuery)


def test_whitespace_only_query_string():
    parser = QueryParser()
    test_cases = [
        "   ",
        "\t",
        "\n",
        "\r\n",
        "  \t  \n  ",
    ]

    for query_string in test_cases:
        result = parser.parse(query_string)
        assert isinstance(result, EmptyQuery)


def test_session_query_with_valid_uuid():
    parser = QueryParser()
    test_uuid = uuid.uuid4()
    query_string = f'session_id = "{test_uuid}"'

    result = parser.parse(query_string)

    assert isinstance(result, SessionQuery)
    assert result.session_id == test_uuid
    assert result.operator == ComparisonOperator.EQUALS


def test_session_query_with_empty_string():
    parser = QueryParser()
    query_string = 'session_id = ""'

    with pytest.raises(QueryParseError) as exc_info:
        parser.parse(query_string)

    assert "session_id cannot be empty" in str(exc_info.value)


def test_session_query_with_invalid_uuid():
    parser = QueryParser()
    query_string = 'session_id = "invalid-uuid"'

    with pytest.raises(QueryParseError):
        parser.parse(query_string)


@pytest.mark.parametrize("token", ["session_id", "SESSION_ID", "session_ID"])
def test_session_query_case_insensitive(token: str):
    parser = QueryParser()
    uid = uuid.uuid4()
    result = parser.parse(f'{token} = "{uid}"')
    assert isinstance(result, SessionQuery)
    assert result.session_id == uid
    assert result.operator == ComparisonOperator.EQUALS


def test_session_query_not_equals():
    parser = QueryParser()
    test_uuid = uuid.uuid4()
    query_string = f'session_id != "{test_uuid}"'

    result = parser.parse(query_string)

    assert isinstance(result, SessionQuery)
    assert result.session_id == test_uuid
    assert result.operator == ComparisonOperator.NOT_EQUALS


def test_session_query_not_equals_empty_string():
    parser = QueryParser()
    query_string = 'session_id != ""'

    with pytest.raises(QueryParseError) as exc_info:
        parser.parse(query_string)

    assert "session_id cannot be empty" in str(exc_info.value)


def test_session_query_not_equals_invalid_uuid():
    parser = QueryParser()
    query_string = 'session_id != "invalid-uuid"'

    with pytest.raises(QueryParseError):
        parser.parse(query_string)


def test_tag_query_with_non_empty_string():
    parser = QueryParser()
    query_string = '#"a"'

    result = parser.parse(query_string)

    assert isinstance(result, TagQuery)
    assert result.tag == "a"
    assert result.operator == ComparisonOperator.EQUALS


def test_tag_query_with_longer_string():
    parser = QueryParser()
    query_string = '#"my-tag"'

    result = parser.parse(query_string)

    assert isinstance(result, TagQuery)
    assert result.tag == "my-tag"
    assert result.operator == ComparisonOperator.EQUALS


def test_tag_query_with_empty_string():
    parser = QueryParser()
    query_string = '#""'

    with pytest.raises(QueryParseError):
        parser.parse(query_string)


def test_tag_assignment_query_with_non_empty_values():
    parser = QueryParser()
    query_string = '#"a" = "bcd"'

    result = parser.parse(query_string)

    assert isinstance(result, TagValueQuery)
    assert result.tag == "a"
    assert result.value == "bcd"
    assert result.operator == ComparisonOperator.EQUALS


def test_tag_assignment_query_with_empty_value():
    parser = QueryParser()
    query_string = '#"tag" = ""'

    result = parser.parse(query_string)

    assert isinstance(result, TagValueQuery)
    assert result.tag == "tag"
    assert result.value == ""


def test_tag_assignment_query_with_empty_tag():
    parser = QueryParser()
    query_string = '#"" = "value"'

    with pytest.raises(QueryParseError):
        parser.parse(query_string)


def test_tag_assignment_query_with_special_characters():
    parser = QueryParser()
    query_string = '#"tag-with_special.chars" = "value with spaces!"'

    result = parser.parse(query_string)

    assert isinstance(result, TagValueQuery)
    assert result.tag == "tag-with_special.chars"
    assert result.value == "value with spaces!"
    assert result.operator == ComparisonOperator.EQUALS


def test_tag_assignment_query_not_equals():
    parser = QueryParser()
    query_string = '#"environment" != "development"'

    result = parser.parse(query_string)

    assert isinstance(result, TagValueQuery)
    assert result.tag == "environment"
    assert result.value == "development"
    assert result.operator == ComparisonOperator.NOT_EQUALS


def test_tag_assignment_query_not_equals_empty_value():
    parser = QueryParser()
    query_string = '#"status" != ""'

    result = parser.parse(query_string)

    assert isinstance(result, TagValueQuery)
    assert result.tag == "status"
    assert result.value == ""
    assert result.operator == ComparisonOperator.NOT_EQUALS


def test_whitespace_handling():
    parser = QueryParser()
    query_string = '  #"tag"  '

    result = parser.parse(query_string)

    assert isinstance(result, TagQuery)
    assert result.tag == "tag"
    assert result.operator == ComparisonOperator.EQUALS


def test_tag_query_with_negation():
    parser = QueryParser()
    query_string = '!#"environment"'

    result = parser.parse(query_string)

    assert isinstance(result, TagQuery)
    assert result.tag == "environment"
    assert result.operator == ComparisonOperator.NOT_EQUALS


def test_tag_query_with_negation_longer_string():
    parser = QueryParser()
    query_string = '!#"test-environment-tag"'

    result = parser.parse(query_string)

    assert isinstance(result, TagQuery)
    assert result.tag == "test-environment-tag"
    assert result.operator == ComparisonOperator.NOT_EQUALS


def test_tag_query_with_negation_empty_string():
    parser = QueryParser()
    query_string = '!#""'

    with pytest.raises(QueryParseError):
        parser.parse(query_string)


def test_invalid_syntax():
    parser = QueryParser()
    invalid_queries = [
        "invalid query",
        'session_id "no equals"',
        "#tag_without_quotes",
        '#"tag" == "double_equals"',
        "session_id = unquoted_value",
    ]

    for invalid_query in invalid_queries:
        with pytest.raises(QueryParseError):
            parser.parse(invalid_query)


def test_real_world_examples():
    parser = QueryParser()
    real_uuid = uuid.UUID("550e8400-e29b-41d4-a716-446655440000")
    session_query = f'session_id = "{real_uuid}"'

    result = parser.parse(session_query)
    assert isinstance(result, SessionQuery)
    assert result.session_id == real_uuid

    tag_queries = [
        '#"user-profile"',
        '#"system:health"',
        '#"log.level"',
    ]

    for query in tag_queries:
        result = parser.parse(query)
        assert isinstance(result, TagQuery)

    assignment_queries = [
        '#"environment" = "production"',
        '#"log.level" = "debug"',
        '#"user.role" = "admin"',
    ]

    for query in assignment_queries:
        result = parser.parse(query)
        assert isinstance(result, TagValueQuery)


def test_compound_query_with_and():
    parser = QueryParser()
    test_uuid = uuid.uuid4()
    query_string = f'session_id = "{test_uuid}" and #"tag" = "value"'

    result = parser.parse(query_string)

    assert isinstance(result, CompoundQuery)
    assert result.operators == [LogicalOperator.AND]
    assert len(result.queries) == 2

    assert isinstance(result.queries[0], SessionQuery)
    assert result.queries[0].session_id == test_uuid
    assert result.queries[0].operator == ComparisonOperator.EQUALS

    assert isinstance(result.queries[1], TagValueQuery)
    assert result.queries[1].tag == "tag"
    assert result.queries[1].value == "value"
    assert result.queries[1].operator == ComparisonOperator.EQUALS


def test_compound_query_with_or():
    parser = QueryParser()
    query_string = '#"tag1" or #"tag2"'

    result = parser.parse(query_string)

    assert isinstance(result, CompoundQuery)
    assert result.operators == [LogicalOperator.OR]
    assert len(result.queries) == 2

    assert isinstance(result.queries[0], TagQuery)
    assert result.queries[0].tag == "tag1"
    assert result.queries[0].operator == ComparisonOperator.EQUALS

    assert isinstance(result.queries[1], TagQuery)
    assert result.queries[1].tag == "tag2"
    assert result.queries[1].operator == ComparisonOperator.EQUALS


@pytest.mark.parametrize("operator", ["and", "AND", "And"])
def test_compound_query_case_insensitive_and(operator: str):
    parser = QueryParser()
    query_string = f'#"tag1" {operator} #"tag2"'

    result = parser.parse(query_string)

    assert isinstance(result, CompoundQuery)
    assert result.operators == [LogicalOperator.AND]


@pytest.mark.parametrize("operator", ["or", "OR", "Or"])
def test_compound_query_case_insensitive_or(operator: str):
    parser = QueryParser()
    query_string = f'#"tag1" {operator} #"tag2"'

    result = parser.parse(query_string)

    assert isinstance(result, CompoundQuery)
    assert result.operators == [LogicalOperator.OR]


def test_multiple_and_operators():
    parser = QueryParser()
    test_uuid = uuid.uuid4()
    query_string = f'session_id = "{test_uuid}" and #"tag1" and #"tag2" = "value"'

    result = parser.parse(query_string)

    assert isinstance(result, CompoundQuery)
    assert result.operators == [LogicalOperator.AND, LogicalOperator.AND]
    assert len(result.queries) == 3

    assert isinstance(result.queries[0], SessionQuery)
    assert isinstance(result.queries[1], TagQuery)
    assert isinstance(result.queries[2], TagValueQuery)


def test_multiple_or_operators():
    parser = QueryParser()
    query_string = '#"tag1" or #"tag2" or #"tag3"'

    result = parser.parse(query_string)

    assert isinstance(result, CompoundQuery)
    assert result.operators == [LogicalOperator.OR, LogicalOperator.OR]
    assert len(result.queries) == 3


def test_compound_query_with_session_and_tags():
    parser = QueryParser()
    test_uuid = uuid.uuid4()
    query_string = f'session_id = "{test_uuid}" and #"a" = "b"'

    result = parser.parse(query_string)

    assert isinstance(result, CompoundQuery)
    assert result.operators == [LogicalOperator.AND]
    assert len(result.queries) == 2

    assert isinstance(result.queries[0], SessionQuery)
    assert result.queries[0].session_id == test_uuid

    assert isinstance(result.queries[1], TagValueQuery)
    assert result.queries[1].tag == "a"
    assert result.queries[1].value == "b"
    assert result.queries[1].operator == ComparisonOperator.EQUALS


def test_compound_query_with_not_equals_operators():
    parser = QueryParser()
    test_uuid = uuid.uuid4()
    query_string = f'session_id = "{test_uuid}" and #"environment" != "development"'

    result = parser.parse(query_string)

    assert isinstance(result, CompoundQuery)
    assert result.operators == [LogicalOperator.AND]
    assert len(result.queries) == 2

    assert isinstance(result.queries[0], SessionQuery)
    assert result.queries[0].session_id == test_uuid
    assert result.queries[0].operator == ComparisonOperator.EQUALS

    assert isinstance(result.queries[1], TagValueQuery)
    assert result.queries[1].tag == "environment"
    assert result.queries[1].value == "development"
    assert result.queries[1].operator == ComparisonOperator.NOT_EQUALS


def test_compound_query_with_tag_negation():
    parser = QueryParser()
    test_uuid = uuid.uuid4()
    query_string = f'session_id = "{test_uuid}" and !#"debug" or #"prod"'

    result = parser.parse(query_string)

    assert isinstance(result, CompoundQuery)
    assert result.operators == [LogicalOperator.AND, LogicalOperator.OR]
    assert len(result.queries) == 3

    assert isinstance(result.queries[0], SessionQuery)
    assert result.queries[0].session_id == test_uuid
    assert result.queries[0].operator == ComparisonOperator.EQUALS

    assert isinstance(result.queries[1], TagQuery)
    assert result.queries[1].tag == "debug"
    assert result.queries[1].operator == ComparisonOperator.NOT_EQUALS

    assert isinstance(result.queries[2], TagQuery)
    assert result.queries[2].tag == "prod"
    assert result.queries[2].operator == ComparisonOperator.EQUALS


def test_group_by_session_id_only():
    parser = QueryParser()
    query_string = "group_by(session_id)"

    result = parser.parse(query_string)

    assert isinstance(result, QueryWithGroupBy)
    assert isinstance(result.main_query, EmptyQuery)
    assert len(result.group_by.tokens) == 1
    assert result.group_by.tokens[0].token_type == GroupByTokenType.SESSION_ID
    assert result.group_by.tokens[0].value == ""


def test_id_query_with_valid_uuid():
    parser = QueryParser()
    test_uuid = uuid.uuid4()
    query_string = f'id = "{test_uuid}"'

    result = parser.parse(query_string)

    assert isinstance(result, IdQuery)
    assert result.id == test_uuid
    assert result.operator == ComparisonOperator.EQUALS


def test_id_query_with_empty_string():
    parser = QueryParser()
    query_string = 'id = ""'

    with pytest.raises(QueryParseError) as exc_info:
        parser.parse(query_string)

    assert "id cannot be empty" in str(exc_info.value)


def test_id_query_with_invalid_uuid():
    parser = QueryParser()
    query_string = 'id = "invalid-uuid"'

    with pytest.raises(QueryParseError):
        parser.parse(query_string)


@pytest.mark.parametrize("token", ["id", "ID", "Id"])
def test_id_query_case_insensitive(token: str):
    parser = QueryParser()
    uid = uuid.uuid4()
    result = parser.parse(f'{token} = "{uid}"')
    assert isinstance(result, IdQuery)
    assert result.id == uid
    assert result.operator == ComparisonOperator.EQUALS


def test_id_query_not_equals():
    parser = QueryParser()
    test_uuid = uuid.uuid4()
    query_string = f'id != "{test_uuid}"'

    result = parser.parse(query_string)

    assert isinstance(result, IdQuery)
    assert result.id == test_uuid
    assert result.operator == ComparisonOperator.NOT_EQUALS


def test_name_query_with_non_empty_string():
    parser = QueryParser()
    query_string = 'name = "test_name"'

    result = parser.parse(query_string)

    assert isinstance(result, NameQuery)
    assert result.name == "test_name"
    assert result.operator == ComparisonOperator.EQUALS


def test_name_query_with_empty_string_fails():
    parser = QueryParser()
    query_string = 'name = ""'

    with pytest.raises(QueryParseError):
        parser.parse(query_string)


@pytest.mark.parametrize("token", ["name", "NAME", "Name"])
def test_name_query_case_insensitive(token: str):
    parser = QueryParser()
    result = parser.parse(f'{token} = "test"')
    assert isinstance(result, NameQuery)
    assert result.name == "test"
    assert result.operator == ComparisonOperator.EQUALS


def test_name_query_not_equals():
    parser = QueryParser()
    query_string = 'name != "unwanted_name"'

    result = parser.parse(query_string)

    assert isinstance(result, NameQuery)
    assert result.name == "unwanted_name"
    assert result.operator == ComparisonOperator.NOT_EQUALS


def test_classname_query_with_string():
    parser = QueryParser()
    query_string = 'classname = "TestClass"'

    result = parser.parse(query_string)

    assert isinstance(result, ClassnameQuery)
    assert result.classname == "TestClass"
    assert result.operator == ComparisonOperator.EQUALS


def test_classname_query_with_empty_string():
    parser = QueryParser()
    query_string = 'classname = ""'

    result = parser.parse(query_string)

    assert isinstance(result, ClassnameQuery)
    assert result.classname == ""
    assert result.operator == ComparisonOperator.EQUALS


@pytest.mark.parametrize("token", ["classname", "CLASSNAME", "ClassName"])
def test_classname_query_case_insensitive(token: str):
    parser = QueryParser()
    result = parser.parse(f'{token} = "MyClass"')
    assert isinstance(result, ClassnameQuery)
    assert result.classname == "MyClass"
    assert result.operator == ComparisonOperator.EQUALS


def test_testsuite_query_with_string():
    parser = QueryParser()
    query_string = 'testsuite = "integration"'

    result = parser.parse(query_string)

    assert isinstance(result, TestsuiteQuery)
    assert result.testsuite == "integration"
    assert result.operator == ComparisonOperator.EQUALS


def test_testsuite_query_with_empty_string():
    parser = QueryParser()
    query_string = 'testsuite = ""'

    result = parser.parse(query_string)

    assert isinstance(result, TestsuiteQuery)
    assert result.testsuite == ""
    assert result.operator == ComparisonOperator.EQUALS


@pytest.mark.parametrize("token", ["testsuite", "TESTSUITE", "TestSuite"])
def test_testsuite_query_case_insensitive(token: str):
    parser = QueryParser()
    result = parser.parse(f'{token} = "unit"')
    assert isinstance(result, TestsuiteQuery)
    assert result.testsuite == "unit"
    assert result.operator == ComparisonOperator.EQUALS


def test_file_query_with_string():
    parser = QueryParser()
    query_string = 'file = "test_file.py"'

    result = parser.parse(query_string)

    assert isinstance(result, FileQuery)
    assert result.file == "test_file.py"
    assert result.operator == ComparisonOperator.EQUALS


def test_file_query_with_empty_string():
    parser = QueryParser()
    query_string = 'file = ""'

    result = parser.parse(query_string)

    assert isinstance(result, FileQuery)
    assert result.file == ""
    assert result.operator == ComparisonOperator.EQUALS


@pytest.mark.parametrize("token", ["file", "FILE", "File"])
def test_file_query_case_insensitive(token: str):
    parser = QueryParser()
    result = parser.parse(f'{token} = "main.py"')
    assert isinstance(result, FileQuery)
    assert result.file == "main.py"
    assert result.operator == ComparisonOperator.EQUALS


def test_status_query_with_valid_status():
    parser = QueryParser()
    query_string = 'status = "pass"'

    result = parser.parse(query_string)

    assert isinstance(result, StatusQuery)
    assert result.status == "pass"
    assert result.operator == ComparisonOperator.EQUALS


def test_status_query_with_empty_string():
    parser = QueryParser()
    query_string = 'status = ""'

    result = parser.parse(query_string)

    assert isinstance(result, StatusQuery)
    assert result.status == ""
    assert result.operator == ComparisonOperator.EQUALS


def test_status_query_with_invalid_status():
    parser = QueryParser()
    query_string = 'status = "invalid"'

    with pytest.raises(QueryParseError):
        parser.parse(query_string)


@pytest.mark.parametrize("status_value", ["pass", "fail", "error", "skip"])
def test_status_query_valid_values(status_value: str):
    parser = QueryParser()
    result = parser.parse(f'status = "{status_value}"')
    assert isinstance(result, StatusQuery)
    assert result.status == status_value
    assert result.operator == ComparisonOperator.EQUALS


@pytest.mark.parametrize("token", ["status", "STATUS", "Status"])
def test_status_query_case_insensitive(token: str):
    parser = QueryParser()
    result = parser.parse(f'{token} = "pass"')
    assert isinstance(result, StatusQuery)
    assert result.status == "pass"
    assert result.operator == ComparisonOperator.EQUALS


def test_compound_query_with_new_tokens():
    parser = QueryParser()
    test_uuid = uuid.uuid4()
    query_string = f'id = "{test_uuid}" and name = "test_method" and status = "pass"'

    result = parser.parse(query_string)

    assert isinstance(result, CompoundQuery)
    assert result.operators == [LogicalOperator.AND, LogicalOperator.AND]
    assert len(result.queries) == 3

    assert isinstance(result.queries[0], IdQuery)
    assert result.queries[0].id == test_uuid

    assert isinstance(result.queries[1], NameQuery)
    assert result.queries[1].name == "test_method"

    assert isinstance(result.queries[2], StatusQuery)
    assert result.queries[2].status == "pass"


def test_mixed_query_types():
    parser = QueryParser()
    session_uuid = uuid.uuid4()
    query_string = f'session_id = "{session_uuid}" and #"env" = "prod" and file = "test.py" and status != "skip"'

    result = parser.parse(query_string)

    assert isinstance(result, CompoundQuery)
    assert result.operators == [
        LogicalOperator.AND,
        LogicalOperator.AND,
        LogicalOperator.AND,
    ]
    assert len(result.queries) == 4

    assert isinstance(result.queries[0], SessionQuery)
    assert isinstance(result.queries[1], TagValueQuery)
    assert isinstance(result.queries[2], FileQuery)
    assert isinstance(result.queries[3], StatusQuery)
    assert result.queries[3].operator == ComparisonOperator.NOT_EQUALS


def test_queries_without_whitespace_around_operators():
    parser = QueryParser()
    test_uuid = uuid.uuid4()
    test_id_uuid = uuid.uuid4()

    test_cases = [
        (f'session_id="{test_uuid}"', SessionQuery),
        (f'id="{test_id_uuid}"', IdQuery),
        ('name="test"', NameQuery),
        ('classname="TestClass"', ClassnameQuery),
        ('testsuite="unit"', TestsuiteQuery),
        ('file="test.py"', FileQuery),
        ('status="pass"', StatusQuery),
        ('#"tag"="value"', TagValueQuery),
    ]

    for query_string, expected_type in test_cases:
        result = parser.parse(query_string)
        assert isinstance(result, expected_type), f"Failed for query: {query_string}"


def test_queries_without_whitespace_not_equals():
    parser = QueryParser()
    test_uuid = uuid.uuid4()

    test_cases = [
        (f'session_id!="{test_uuid}"', SessionQuery),
        (f'id!="{test_uuid}"', IdQuery),
        ('name!="unwanted"', NameQuery),
        ('classname!="BadClass"', ClassnameQuery),
        ('testsuite!="integration"', TestsuiteQuery),
        ('file!="bad.py"', FileQuery),
        ('status!="fail"', StatusQuery),
        ('#"env"!="dev"', TagValueQuery),
    ]

    for query_string, expected_type in test_cases:
        result = parser.parse(query_string)
        assert isinstance(result, expected_type), f"Failed for query: {query_string}"
        assert result.operator == ComparisonOperator.NOT_EQUALS


def test_compound_queries_without_whitespace():
    parser = QueryParser()
    test_uuid = uuid.uuid4()
    query_string = f'session_id="{test_uuid}"and name="test"and status="pass"'

    result = parser.parse(query_string)

    assert isinstance(result, CompoundQuery)
    assert result.operators == [LogicalOperator.AND, LogicalOperator.AND]
    assert len(result.queries) == 3

    assert isinstance(result.queries[0], SessionQuery)
    assert isinstance(result.queries[1], NameQuery)
    assert isinstance(result.queries[2], StatusQuery)


def test_mixed_whitespace_patterns():
    parser = QueryParser()
    test_uuid = uuid.uuid4()
    query_string = f'session_id= "{test_uuid}" and name="test"and status !="skip"'

    result = parser.parse(query_string)

    assert isinstance(result, CompoundQuery)
    assert result.operators == [LogicalOperator.AND, LogicalOperator.AND]
    assert len(result.queries) == 3

    assert isinstance(result.queries[0], SessionQuery)
    assert result.queries[0].operator == ComparisonOperator.EQUALS

    assert isinstance(result.queries[1], NameQuery)
    assert result.queries[1].operator == ComparisonOperator.EQUALS

    assert isinstance(result.queries[2], StatusQuery)
    assert result.queries[2].operator == ComparisonOperator.NOT_EQUALS


def test_group_by_tag_only():
    parser = QueryParser()
    query_string = 'group_by(#"environment")'

    result = parser.parse(query_string)

    assert isinstance(result, QueryWithGroupBy)
    assert isinstance(result.main_query, EmptyQuery)
    assert len(result.group_by.tokens) == 1
    assert result.group_by.tokens[0].token_type == GroupByTokenType.TAG
    assert result.group_by.tokens[0].value == "environment"


def test_group_by_multiple_tokens():
    parser = QueryParser()
    query_string = 'group_by(session_id, #"env", #"user")'

    result = parser.parse(query_string)

    assert isinstance(result, QueryWithGroupBy)
    assert isinstance(result.main_query, EmptyQuery)
    assert len(result.group_by.tokens) == 3

    assert result.group_by.tokens[0].token_type == GroupByTokenType.SESSION_ID
    assert result.group_by.tokens[0].value == ""

    assert result.group_by.tokens[1].token_type == GroupByTokenType.TAG
    assert result.group_by.tokens[1].value == "env"

    assert result.group_by.tokens[2].token_type == GroupByTokenType.TAG
    assert result.group_by.tokens[2].value == "user"


def test_group_by_with_main_query():
    parser = QueryParser()
    test_uuid = uuid.uuid4()
    query_string = f'session_id = "{test_uuid}" and #"status" = "active" group_by(session_id, #"env")'

    result = parser.parse(query_string)

    assert isinstance(result, QueryWithGroupBy)
    assert isinstance(result.main_query, CompoundQuery)
    assert result.main_query.operators == [LogicalOperator.AND]
    assert len(result.main_query.queries) == 2

    assert len(result.group_by.tokens) == 2
    assert result.group_by.tokens[0].token_type == GroupByTokenType.SESSION_ID
    assert result.group_by.tokens[1].token_type == GroupByTokenType.TAG
    assert result.group_by.tokens[1].value == "env"


def test_group_by_duplicate_tokens_validation():
    parser = QueryParser()
    query_string = 'group_by(#"env", #"env")'

    with pytest.raises(QueryParseError):
        parser.parse(query_string)


def test_group_by_duplicate_session_id_validation():
    parser = QueryParser()
    query_string = "group_by(session_id, session_id)"

    with pytest.raises(QueryParseError):
        parser.parse(query_string)


def test_group_by_with_tag_query_and_assignment():
    parser = QueryParser()
    query_string = '#"debug" or #"trace" = "enabled" group_by(#"level")'

    result = parser.parse(query_string)

    assert isinstance(result, QueryWithGroupBy)
    assert isinstance(result.main_query, CompoundQuery)
    assert result.main_query.operators == [LogicalOperator.OR]

    assert len(result.group_by.tokens) == 1
    assert result.group_by.tokens[0].token_type == GroupByTokenType.TAG
    assert result.group_by.tokens[0].value == "level"


def test_group_by_case_sensitivity():
    parser = QueryParser()
    test_cases = [
        "group_by(session_id)",
        "GROUP_BY(session_id)",
        "Group_By(session_id)",
    ]

    for query_string in test_cases:
        result = parser.parse(query_string)
        assert isinstance(result, QueryWithGroupBy)
        assert result.group_by.tokens[0].token_type == GroupByTokenType.SESSION_ID


def test_group_by_with_not_equals_queries():
    parser = QueryParser()
    test_uuid = uuid.uuid4()
    query_string = (
        f'session_id != "{test_uuid}" and #"env" != "development" group_by(session_id)'
    )

    result = parser.parse(query_string)

    assert isinstance(result, QueryWithGroupBy)
    assert isinstance(result.main_query, CompoundQuery)

    assert result.main_query.queries[0].operator == ComparisonOperator.NOT_EQUALS
    assert result.main_query.queries[1].operator == ComparisonOperator.NOT_EQUALS

    assert len(result.group_by.tokens) == 1
    assert result.group_by.tokens[0].token_type == GroupByTokenType.SESSION_ID


def test_mixed_and_or_operators():
    parser = QueryParser()
    test_uuid = uuid.uuid4()
    query_string = f'session_id = "{test_uuid}" and id = "{test_uuid}" or #"tag"'

    result = parser.parse(query_string)

    assert isinstance(result, CompoundQuery)
    assert len(result.queries) == 3
    assert result.operators == [LogicalOperator.AND, LogicalOperator.OR]

    assert isinstance(result.queries[0], SessionQuery)
    assert result.queries[0].session_id == test_uuid
    assert result.queries[0].operator == ComparisonOperator.EQUALS

    assert isinstance(result.queries[1], IdQuery)
    assert result.queries[1].id == test_uuid
    assert result.queries[1].operator == ComparisonOperator.EQUALS

    assert isinstance(result.queries[2], TagQuery)
    assert result.queries[2].tag == "tag"


def test_complex_mixed_operators():
    parser = QueryParser()
    test_uuid = uuid.uuid4()
    query_string = f'#"tag1" or session_id = "{test_uuid}" and id = "{test_uuid}" or #"tag2" or #"tag3"'

    result = parser.parse(query_string)

    assert isinstance(result, CompoundQuery)
    assert len(result.queries) == 5
    assert result.operators == [
        LogicalOperator.OR,
        LogicalOperator.AND,
        LogicalOperator.OR,
        LogicalOperator.OR,
    ]

    assert isinstance(result.queries[0], TagQuery)
    assert isinstance(result.queries[1], SessionQuery)
    assert isinstance(result.queries[2], IdQuery)
    assert isinstance(result.queries[3], TagQuery)
    assert isinstance(result.queries[4], TagQuery)


def test_left_to_right_parsing_no_precedence():
    parser = QueryParser()
    query_string = '#"A" or #"B" and #"C" or #"D"'

    result = parser.parse(query_string)

    assert isinstance(result, CompoundQuery)
    assert len(result.queries) == 4
    assert result.operators == [
        LogicalOperator.OR,
        LogicalOperator.AND,
        LogicalOperator.OR,
    ]

    assert isinstance(result.queries[0], TagQuery)
    assert result.queries[0].tag == "A"
    assert isinstance(result.queries[1], TagQuery)
    assert result.queries[1].tag == "B"
    assert isinstance(result.queries[2], TagQuery)
    assert result.queries[2].tag == "C"
    assert isinstance(result.queries[3], TagQuery)
    assert result.queries[3].tag == "D"
