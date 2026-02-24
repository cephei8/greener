package query

import "fmt"

const selectorCountMismatchError = "group selector must have same number of values as group_by columns count"

type QueryType int

const (
	QueryTypeSession QueryType = iota
	QueryTypeTestcase
	QueryTypeGroup
)

func Validate(q Query, queryType QueryType) error {
	switch queryType {
	case QueryTypeSession, QueryTypeTestcase:
		return validateGroupSelectorPresent(q)

	case QueryTypeGroup:
		return validateGroupSelectorAbsent(q)

	default:
		panic(fmt.Sprintf("unknown query type: %d", queryType))
	}
}

func validateGroupSelectorPresent(q Query) error {
	if q.GroupQuery != nil && q.GroupSelector == nil {
		return &QueryError{
			Message: "group selector is required when grouping results. Use: group_by(...) group = (...)",
		}
	}

	if q.GroupQuery != nil && q.GroupSelector != nil {
		if len(q.GroupQuery.Tokens) != len(q.GroupSelector) {
			return &QueryError{
				Message: selectorCountMismatchError,
			}
		}
	}

	return nil
}

func validateGroupSelectorAbsent(q Query) error {
	if q.GroupQuery == nil {
		return &QueryError{
			Message: "group_by is required",
		}
	}

	if q.GroupSelector != nil {
		if len(q.GroupQuery.Tokens) != len(q.GroupSelector) {
			return &QueryError{
				Message: selectorCountMismatchError,
			}
		}
	}

	return nil
}
