package core

import (
	"fmt"

	"git.sr.ht/~cephei8/greener/core/model/db"
	"git.sr.ht/~cephei8/greener/core/query"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type QueryTable string

const (
	testcasesTable QueryTable = "testcases"
	sessionsTable  QueryTable = "sessions"
	labelsTable    QueryTable = "labels"
)

func convertQueryStatusToDBStatus(status query.TestcaseStatus) model_db.TestcaseStatus {
	switch status {
	case query.StatusPass:
		return model_db.StatusPass
	case query.StatusFail:
		return model_db.StatusFail
	case query.StatusError:
		return model_db.StatusError
	case query.StatusSkip:
		return model_db.StatusSkip
	default:
		panic(fmt.Sprintf("unknown status: %s", status))
	}
}

func applyOffsetLimit(bunQuery *bun.SelectQuery, queryAST query.Query) (*bun.SelectQuery, error) {
	offset := queryAST.Offset
	if offset < 0 {
		return nil, fmt.Errorf("offset must be non-negative")
	}
	if offset > 0 {
		bunQuery = bunQuery.Offset(offset)
	}

	limit := queryAST.Limit
	if limit == 0 {
		limit = 100
	}
	if limit < 0 {
		return nil, fmt.Errorf("limit must be positive")
	}
	if limit > 100 {
		return nil, fmt.Errorf("limit cannot exceed 100")
	}
	bunQuery = bunQuery.Limit(limit)

	return bunQuery, nil
}

func BuildTestcasesQuery(
	db *bun.DB,
	userID model_db.BinaryUUID,
	queryAST query.Query,
) (*bun.SelectQuery, error) {
	cteQuery := db.NewSelect().
		Column(fmt.Sprintf("%s.*", testcasesTable)).
		Table(fmt.Sprintf("%s", testcasesTable)).
		Where("? = ?", bun.Ident(fmt.Sprintf("%s.user_id", testcasesTable)), userID).
		OrderBy(fmt.Sprintf("%s.created_at", testcasesTable), bun.OrderDesc)

	if queryAST.StartDate != nil {
		cteQuery = cteQuery.Where("? >= ?", bun.Ident(fmt.Sprintf("%s.created_at", testcasesTable)), queryAST.StartDate)
	}
	if queryAST.EndDate != nil {
		cteQuery = cteQuery.Where("? <= ?", bun.Ident(fmt.Sprintf("%s.created_at", testcasesTable)), queryAST.EndDate)
	}

	cteQuery = applySelectQuery(cteQuery, queryAST.SelectQuery)

	if queryAST.GroupQuery != nil {
		if len(queryAST.GroupQuery.Tokens) != len(queryAST.GroupSelector) {
			panic("logic error: grouping/selector mismatch")
		}

		for i, token := range queryAST.GroupQuery.Tokens {
			groupValue := queryAST.GroupSelector[i]
			switch t := token.(type) {
			case query.SessionGroupToken:
				sessionUUID, err := uuid.Parse(groupValue)
				if err != nil {
					return nil, fmt.Errorf("invalid selector value: %s", groupValue)
				}
				cteQuery = cteQuery.Where(
					"? = ?",
					bun.Ident(fmt.Sprintf("%s.session_id", testcasesTable)),
					model_db.BinaryUUID(sessionUUID),
				)
			case query.TagGroupToken:
				cteQuery = cteQuery.Where(
					"? IN (SELECT ? FROM ? WHERE ? = ? AND ? = ?)",
					bun.Ident(fmt.Sprintf("%s.session_id", testcasesTable)),
					bun.Ident(fmt.Sprintf("%s.session_id", labelsTable)),
					bun.Ident("labels"),
					bun.Ident("key"),
					t.Tag,
					bun.Ident("value"),
					groupValue,
				)
			default:
				panic(fmt.Sprintf("unknown token type: %s", t))
			}
		}
	}

	mainQuery := db.NewSelect().
		Table("cte").
		Column("*").
		ColumnExpr("COUNT(?) OVER() AS ?", 1, bun.Ident("total_count")).
		ColumnExpr("MIN(?) OVER() AS ?", bun.Ident("status"), bun.Ident("aggregated_status")).
		With("cte", cteQuery)

	mainQuery, err := applyOffsetLimit(mainQuery, queryAST)
	if err != nil {
		return nil, err
	}

	return mainQuery, nil
}

func BuildSessionsQuery(
	db *bun.DB,
	userID model_db.BinaryUUID,
	queryAST query.Query,
) (*bun.SelectQuery, error) {
	cteQuery := db.NewSelect().
		Column(fmt.Sprintf("%s.*", sessionsTable)).
		ColumnExpr(
			"MIN(?) AS ?",
			bun.Ident(fmt.Sprintf("%s.status", testcasesTable)),
			bun.Ident("aggregated_status"),
		).
		Table(fmt.Sprintf("%s", sessionsTable)).
		Join(
			"LEFT JOIN ? ON ? = ?",
			bun.Ident(fmt.Sprintf("%s", testcasesTable)),
			bun.Ident(fmt.Sprintf("%s.id", sessionsTable)),
			bun.Ident(fmt.Sprintf("%s.session_id", testcasesTable)),
		).
		Where("? = ?", bun.Ident(fmt.Sprintf("%s.user_id", sessionsTable)), userID).
		Group(fmt.Sprintf("%s.id", sessionsTable)).
		OrderBy(fmt.Sprintf("%s.created_at", sessionsTable), bun.OrderDesc)

	if queryAST.StartDate != nil {
		cteQuery = cteQuery.Where("? >= ?", bun.Ident(fmt.Sprintf("%s.created_at", sessionsTable)), queryAST.StartDate)
	}
	if queryAST.EndDate != nil {
		cteQuery = cteQuery.Where("? <= ?", bun.Ident(fmt.Sprintf("%s.created_at", sessionsTable)), queryAST.EndDate)
	}

	cteQuery = applySelectQuery(cteQuery, queryAST.SelectQuery)

	if queryAST.GroupQuery != nil {
		if len(queryAST.GroupQuery.Tokens) != len(queryAST.GroupSelector) {
			panic("logic error: grouping/selector mismatch")
		}

		for i, token := range queryAST.GroupQuery.Tokens {
			groupValue := queryAST.GroupSelector[i]
			switch t := token.(type) {
			case query.SessionGroupToken:
				sessionUUID, err := uuid.Parse(groupValue)
				if err != nil {
					return nil, fmt.Errorf("invalid selector value: %s", groupValue)
				}
				cteQuery = cteQuery.Where(
					"? = ?",
					bun.Ident(fmt.Sprintf("%s.id", sessionsTable)),
					model_db.BinaryUUID(sessionUUID),
				)
			case query.TagGroupToken:
				cteQuery = cteQuery.Where(
					"? IN (SELECT ? FROM ? WHERE ? = ? AND ? = ?)",
					bun.Ident(fmt.Sprintf("%s.id", sessionsTable)),
					bun.Ident(fmt.Sprintf("%s.session_id", labelsTable)),
					bun.Ident("labels"),
					bun.Ident("key"),
					t.Tag,
					bun.Ident("value"),
					groupValue)
			}
		}
	}

	mainQuery := db.NewSelect().
		Table("cte").
		Column("*").
		ColumnExpr("COUNT(?) OVER() AS ?", 1, bun.Ident("total_count")).
		With("cte", cteQuery)

	mainQuery, err := applyOffsetLimit(mainQuery, queryAST)
	if err != nil {
		return nil, err
	}

	return mainQuery, nil
}

func BuildGroupsQuery(db *bun.DB, userID model_db.BinaryUUID, queryAST query.Query, groupBy *query.GroupQuery) (*bun.SelectQuery, error) {
	if groupBy == nil || len(groupBy.Tokens) == 0 {
		return nil, fmt.Errorf("group_by clause required for groups page")
	}

	if len(queryAST.GroupSelector) > 0 {
		if len(groupBy.Tokens) != len(queryAST.GroupSelector) {
			panic("logic error: grouping/selector mismatch")
		}
	}

	groupCols := []string{}
	orderCols := []string{}

	cteQuery := db.NewSelect().
		Table(fmt.Sprintf("%s", testcasesTable))

	labelJoinIdx := 0
	for _, token := range groupBy.Tokens {
		switch t := token.(type) {
		case query.SessionGroupToken:
			idCol := fmt.Sprintf("%s.id", sessionsTable)
			cteQuery = cteQuery.ColumnExpr("? AS ?", bun.Ident(idCol), bun.Ident("session_id"))
			groupCols = append(groupCols, idCol)
			orderCols = append(orderCols, idCol)

		case query.TagGroupToken:
			alias := fmt.Sprintf("l%d", labelJoinIdx)
			valCol := fmt.Sprintf("%s.value", alias)
			cteQuery = cteQuery.ColumnExpr(
				"? AS ?",
				bun.Ident(valCol),
				bun.Ident(fmt.Sprintf("\"%s\"", t.Tag)),
			)
			groupCols = append(groupCols, valCol)
			orderCols = append(orderCols, valCol)
			labelJoinIdx++
		}
	}

	cteQuery = cteQuery.ColumnExpr(
		"MIN(?) AS ?",
		bun.Ident(fmt.Sprintf("%s.status", testcasesTable)),
		bun.Ident("aggregated_status"),
	)
	cteQuery = cteQuery.ColumnExpr(
		"COUNT(DISTINCT ?) AS ?",
		bun.Ident(fmt.Sprintf("%s.id", testcasesTable)),
		bun.Ident("testcase_count"),
	)

	labelJoinIdx = 0
	for _, token := range groupBy.Tokens {
		switch t := token.(type) {
		case query.SessionGroupToken:
			cteQuery = cteQuery.Join(
				"JOIN ? ON ? = ?",
				bun.Ident(fmt.Sprintf("%s", sessionsTable)),
				bun.Ident(fmt.Sprintf("%s.session_id", testcasesTable)),
				bun.Ident(fmt.Sprintf("%s.id", sessionsTable)),
			)

		case query.TagGroupToken:
			alias := fmt.Sprintf("l%d", labelJoinIdx)
			cteQuery = cteQuery.Join(
				"JOIN ? AS ? ON ? = ? AND ? = ?",
				bun.Ident(fmt.Sprintf("%s", labelsTable)),
				bun.Ident(alias),
				bun.Ident(fmt.Sprintf("%s.session_id", testcasesTable)),
				bun.Ident(fmt.Sprintf("%s.session_id", alias)),
				bun.Ident(fmt.Sprintf("%s.key", alias)),
				t.Tag,
			)
			labelJoinIdx++
		}
	}

	cteQuery = cteQuery.Where("? = ?", bun.Ident(fmt.Sprintf("%s.user_id", testcasesTable)), userID)

	if queryAST.StartDate != nil {
		cteQuery = cteQuery.Where("? >= ?", bun.Ident(fmt.Sprintf("%s.created_at", testcasesTable)), queryAST.StartDate)
	}
	if queryAST.EndDate != nil {
		cteQuery = cteQuery.Where("? <= ?", bun.Ident(fmt.Sprintf("%s.created_at", testcasesTable)), queryAST.EndDate)
	}

	if len(queryAST.GroupSelector) > 0 {
		labelJoinIdx = 0
		for i, token := range groupBy.Tokens {
			groupValue := queryAST.GroupSelector[i]
			switch token.(type) {
			case query.SessionGroupToken:
				sessionUUID, err := uuid.Parse(groupValue)
				if err != nil {
					return nil, fmt.Errorf("invalid selector value: %s", groupValue)
				}
				cteQuery = cteQuery.Where(
					"? = ?",
					bun.Ident(fmt.Sprintf("%s.id", sessionsTable)),
					model_db.BinaryUUID(sessionUUID),
				)
			case query.TagGroupToken:
				alias := fmt.Sprintf("l%d", labelJoinIdx)
				cteQuery = cteQuery.Where("? = ?", bun.Ident(fmt.Sprintf("%s.value", alias)), groupValue)
				labelJoinIdx++
			}
		}
	}

	cteQuery = applySelectQuery(cteQuery, queryAST.SelectQuery)

	for _, col := range groupCols {
		cteQuery = cteQuery.Group(col)
	}

	for _, col := range orderCols {
		cteQuery = cteQuery.Order(col)
	}

	mainQuery := db.NewSelect().
		Table("cte").
		Column("*").
		ColumnExpr("COUNT(?) OVER() AS ?", 1, bun.Ident("total_count")).
		With("cte", cteQuery)

	mainQuery, err := applyOffsetLimit(mainQuery, queryAST)
	if err != nil {
		return nil, err
	}

	return mainQuery, nil
}

func applySelectQuery(bunQuery *bun.SelectQuery, csq query.CompoundSelectQuery) *bun.SelectQuery {
	applyAtomicQuery := func(sq *bun.SelectQuery, atomicQuery query.SelectQuery, useOr bool) *bun.SelectQuery {
		eqCondition := func(op query.EqualityOperator, ident bun.Ident, arg any) *bun.SelectQuery {
			whereFunc := sq.Where
			if useOr {
				whereFunc = sq.WhereOr
			}

			if op == query.OpEq {
				return whereFunc("? = ?", ident, arg)
			} else {
				return whereFunc("? != ?", ident, arg)
			}
		}

		switch qt := atomicQuery.(type) {
		case query.EmptySelectQuery:
			return sq

		case query.SessionSelectQuery:
			return eqCondition(
				qt.Operator,
				bun.Ident(fmt.Sprintf("%s.session_id", testcasesTable)),
				model_db.BinaryUUID(qt.SessionId),
			)

		case query.IdSelectQuery:
			return eqCondition(
				qt.Operator,
				bun.Ident(fmt.Sprintf("%s.id", testcasesTable)),
				model_db.BinaryUUID(qt.Id),
			)

		case query.NameSelectQuery:
			return eqCondition(
				qt.Operator,
				bun.Ident(fmt.Sprintf("%s.name", testcasesTable)),
				qt.Name,
			)

		case query.ClassnameSelectQuery:
			return eqCondition(
				qt.Operator,
				bun.Ident(fmt.Sprintf("%s.classname", testcasesTable)),
				qt.Classname,
			)

		case query.TestsuiteSelectQuery:
			return eqCondition(
				qt.Operator,
				bun.Ident(fmt.Sprintf("%s.testsuite", testcasesTable)),
				qt.Testsuite,
			)

		case query.FileSelectQuery:
			return eqCondition(
				qt.Operator,
				bun.Ident(fmt.Sprintf("%s.file", testcasesTable)),
				qt.File,
			)

		case query.StatusSelectQuery:
			return eqCondition(
				qt.Operator,
				bun.Ident(fmt.Sprintf("%s.status", testcasesTable)),
				convertQueryStatusToDBStatus(qt.Status),
			)

		case query.TagSelectQuery:
			whereFunc := sq.Where
			if useOr {
				whereFunc = sq.WhereOr
			}
			sessionCol := fmt.Sprintf("%s.session_id", testcasesTable)
			tcColSession := bun.Ident(sessionCol)
			laTbl := bun.Ident(labelsTable)
			laColSession := bun.Ident(fmt.Sprintf("%s.session_id", labelsTable))
			laColKey := bun.Ident("key")
			if qt.Operator == query.OpEq {
				return whereFunc(
					"? IN (SELECT ? FROM ? WHERE ? = ?)",
					tcColSession, laColSession, laTbl, laColKey, qt.Tag,
				)
			} else {
				return whereFunc(
					"? NOT IN (SELECT ? FROM ? WHERE ? = ?)",
					tcColSession, laColSession, laTbl, laColKey, qt.Tag,
				)
			}

		case query.TagValueSelectQuery:
			whereFunc := sq.Where
			if useOr {
				whereFunc = sq.WhereOr
			}

			sessionCol := fmt.Sprintf("%s.session_id", testcasesTable)
			tcColSession := bun.Ident(sessionCol)
			laTbl := bun.Ident(labelsTable)
			laColSession := bun.Ident(fmt.Sprintf("%s.session_id", labelsTable))
			laColKey := bun.Ident("key")
			laColValue := bun.Ident("value")
			if qt.Operator == query.OpEq {
				return whereFunc(
					"? IN (SELECT ? FROM ? WHERE ? = ? AND ? = ?)",
					tcColSession, laColSession, laTbl, laColKey, qt.Tag, laColValue, qt.Value,
				)
			} else {
				return whereFunc(
					"? NOT IN (SELECT ? FROM ? WHERE ? = ? AND ? = ?)",
					tcColSession, laColSession, laTbl, laColKey, qt.Tag, laColValue, qt.Value,
				)
			}

		default:
			panic(fmt.Sprintf("unknown atomic query type: %T", atomicQuery))
		}
	}

	if len(csq.Parts) == 0 {
		return bunQuery
	}

	if csq.Parts[0].Operator != query.OpAnd {
		panic("logic error: invalid operator in the first part of compound select query")
	}

	if len(csq.Parts) == 1 {
		if _, ok := csq.Parts[0].Query.(query.EmptySelectQuery); ok {
			return bunQuery
		}
		return applyAtomicQuery(bunQuery, csq.Parts[0].Query, false)
	}

	return bunQuery.WhereGroup(" AND ", func(sq *bun.SelectQuery) *bun.SelectQuery {
		for i := 0; i < len(csq.Parts); i++ {
			useOr := csq.Parts[i].Operator == query.OpOr
			sq = applyAtomicQuery(sq, csq.Parts[i].Query, useOr)
		}
		return sq
	})
}
