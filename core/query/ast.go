package query

import (
	"github.com/google/uuid"
)

////////////////////////////////////////////////////////////

type EqualityOperator int

const (
	OpEq EqualityOperator = iota
	OpNEq
)

////////////////////////////////////////////////////////////

type LogicalOperator int

const (
	OpAnd LogicalOperator = iota
	OpOr
)

////////////////////////////////////////////////////////////

type SelectQuery interface {
	isSelectQuery()
}

////////////////////////////////////////////////////////////

type SessionSelectQuery struct {
	SessionId uuid.UUID
	Operator  EqualityOperator
}

func (SessionSelectQuery) isSelectQuery() {}

////////////////////////////////////////////////////////////

type TagSelectQuery struct {
	Tag      string
	Operator EqualityOperator
}

func (TagSelectQuery) isSelectQuery() {}

////////////////////////////////////////////////////////////

type TagValueSelectQuery struct {
	Tag      string
	Value    string
	Operator EqualityOperator
}

func (TagValueSelectQuery) isSelectQuery() {}

////////////////////////////////////////////////////////////

type IdSelectQuery struct {
	Id       uuid.UUID
	Operator EqualityOperator
}

func (IdSelectQuery) isSelectQuery() {}

////////////////////////////////////////////////////////////

type NameSelectQuery struct {
	Name     string
	Operator EqualityOperator
}

func (NameSelectQuery) isSelectQuery() {}

////////////////////////////////////////////////////////////

type ClassnameSelectQuery struct {
	Classname string
	Operator  EqualityOperator
}

func (ClassnameSelectQuery) isSelectQuery() {}

////////////////////////////////////////////////////////////

type TestsuiteSelectQuery struct {
	Testsuite string
	Operator  EqualityOperator
}

func (TestsuiteSelectQuery) isSelectQuery() {}

////////////////////////////////////////////////////////////

type FileSelectQuery struct {
	File     string
	Operator EqualityOperator
}

func (FileSelectQuery) isSelectQuery() {}

////////////////////////////////////////////////////////////

type TestcaseStatus string

const (
	StatusPass  TestcaseStatus = "pass"
	StatusFail  TestcaseStatus = "fail"
	StatusError TestcaseStatus = "error"
	StatusSkip  TestcaseStatus = "skip"
)

////////////////////////////////////////////////////////////

type StatusSelectQuery struct {
	Status   TestcaseStatus
	Operator EqualityOperator
}

func (StatusSelectQuery) isSelectQuery() {}

////////////////////////////////////////////////////////////

type EmptySelectQuery struct{}

func (EmptySelectQuery) isSelectQuery() {}

////////////////////////////////////////////////////////////

type CompoundSelectQueryPart struct {
	Operator LogicalOperator
	Query    SelectQuery
}

type CompoundSelectQuery struct {
	Parts []CompoundSelectQueryPart
}

////////////////////////////////////////////////////////////

type GroupToken interface {
	isGroupToken()
}

type SessionGroupToken struct {
}

func (SessionGroupToken) isGroupToken() {}

type TagGroupToken struct {
	Tag string
}

func (TagGroupToken) isGroupToken() {}

type GroupQuery struct {
	Tokens []GroupToken
}

////////////////////////////////////////////////////////////

type Query struct {
	SelectQuery   CompoundSelectQuery
	GroupQuery    *GroupQuery
	GroupSelector []string
}
