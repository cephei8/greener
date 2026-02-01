package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	model_api "github.com/cephei8/greener/core/model/api"
	model_db "github.com/cephei8/greener/core/model/db"
	"github.com/cephei8/greener/core/query"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type QueryServiceInterface interface {
	QueryTestcases(ctx context.Context, userID model_db.BinaryUUID, params QueryParams) (*QueryResult[model_api.Testcase], error)
	QuerySessions(ctx context.Context, userID model_db.BinaryUUID, params QueryParams) (*QueryResult[model_api.Session], error)
	QueryGroups(ctx context.Context, userID model_db.BinaryUUID, params QueryParams) (*QueryResult[model_api.Group], error)
	GetTestcase(ctx context.Context, userID model_db.BinaryUUID, testcaseID uuid.UUID) (*TestcaseDetail, error)
	GetSession(ctx context.Context, userID model_db.BinaryUUID, sessionID uuid.UUID) (*SessionDetail, error)
}

type QueryService struct {
	db *bun.DB
}

func NewQueryService(db *bun.DB) *QueryService {
	return &QueryService{db: db}
}

type QueryParams struct {
	Query  string
	Offset int
	Limit  int
}

type QueryResult[T any] struct {
	Results    []T
	TotalCount int
}

type TestcaseDetail struct {
	ID        string
	SessionID string
	Name      string
	Status    string
	Classname string
	File      string
	Testsuite string
	Output    string
	Baggage   any
	Labels    map[string]string
	CreatedAt string
}

type SessionDetail struct {
	ID          string
	Description string
	Status      string
	Baggage     any
	Labels      map[string]string
	CreatedAt   string
}

func (s *QueryService) QueryTestcases(ctx context.Context, userID model_db.BinaryUUID, params QueryParams) (*QueryResult[model_api.Testcase], error) {
	var queryAST query.Query

	if params.Query != "" {
		parser := query.NewParser(params.Query)
		parsedQuery, err := parser.Parse()
		if err != nil {
			return nil, fmt.Errorf("invalid query: %w", err)
		}

		if err := query.Validate(parsedQuery, query.QueryTypeTestcase); err != nil {
			return nil, fmt.Errorf("invalid query: %w", err)
		}

		queryAST = parsedQuery
	}

	if params.Offset > 0 {
		queryAST.Offset = params.Offset
	}
	if params.Limit > 0 && params.Limit <= 100 {
		queryAST.Limit = params.Limit
	}

	q, err := BuildTestcasesQuery(s.db, userID, queryAST)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	type dbResult struct {
		model_db.Testcase
		TotalCount       int64 `bun:"total_count"`
		AggregatedStatus int64 `bun:"aggregated_status"`
	}

	var results []dbResult
	if err := q.Scan(ctx, &results); err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	testcases := []model_api.Testcase{}
	totalCount := 0

	for _, result := range results {
		if totalCount == 0 && result.TotalCount > 0 {
			totalCount = int(result.TotalCount)
		}

		testcaseID, _ := uuid.FromBytes(result.ID[:])
		sessionID, _ := uuid.FromBytes(result.SessionID[:])

		testcases = append(testcases, model_api.Testcase{
			ID:        testcaseID.String(),
			SessionID: sessionID.String(),
			Name:      result.Name,
			Status:    TestcaseStatusToString(result.Status),
			CreatedAt: result.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &QueryResult[model_api.Testcase]{
		Results:    testcases,
		TotalCount: totalCount,
	}, nil
}

func (s *QueryService) QuerySessions(ctx context.Context, userID model_db.BinaryUUID, params QueryParams) (*QueryResult[model_api.Session], error) {
	var queryAST query.Query

	if params.Query != "" {
		parser := query.NewParser(params.Query)
		parsedQuery, err := parser.Parse()
		if err != nil {
			return nil, fmt.Errorf("invalid query: %w", err)
		}

		if err := query.Validate(parsedQuery, query.QueryTypeSession); err != nil {
			return nil, fmt.Errorf("invalid query: %w", err)
		}

		queryAST = parsedQuery
	}

	if params.Offset > 0 {
		queryAST.Offset = params.Offset
	}
	if params.Limit > 0 && params.Limit <= 100 {
		queryAST.Limit = params.Limit
	}

	q, err := BuildSessionsQuery(s.db, userID, queryAST)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	type dbResult struct {
		model_db.Session
		AggregatedStatus int64 `bun:"aggregated_status"`
		TotalCount       int64 `bun:"total_count"`
	}

	var results []dbResult
	if err := q.Scan(ctx, &results); err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	sessions := []model_api.Session{}
	totalCount := 0

	for _, result := range results {
		if totalCount == 0 && result.TotalCount > 0 {
			totalCount = int(result.TotalCount)
		}

		sessionID, _ := uuid.FromBytes(result.ID[:])

		desc := ""
		if result.Description != nil {
			desc = *result.Description
		}

		sessions = append(sessions, model_api.Session{
			ID:          sessionID.String(),
			Description: desc,
			Status:      TestcaseStatusToString(model_db.TestcaseStatus(result.AggregatedStatus)),
			CreatedAt:   result.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &QueryResult[model_api.Session]{
		Results:    sessions,
		TotalCount: totalCount,
	}, nil
}

func (s *QueryService) QueryGroups(ctx context.Context, userID model_db.BinaryUUID, params QueryParams) (*QueryResult[model_api.Group], error) {
	if params.Query == "" {
		return nil, fmt.Errorf("query is required for group queries")
	}

	parser := query.NewParser(params.Query)
	queryAST, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	if err := query.Validate(queryAST, query.QueryTypeGroup); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	if queryAST.GroupQuery == nil {
		return &QueryResult[model_api.Group]{
			Results:    []model_api.Group{},
			TotalCount: 0,
		}, nil
	}

	if params.Offset > 0 {
		queryAST.Offset = params.Offset
	}
	if params.Limit > 0 && params.Limit <= 100 {
		queryAST.Limit = params.Limit
	}

	type groupColumn struct {
		header string
		isUUID bool
	}
	var groupColumns []groupColumn

	for _, token := range queryAST.GroupQuery.Tokens {
		switch t := token.(type) {
		case query.SessionGroupToken:
			groupColumns = append(groupColumns, groupColumn{
				header: "session_id",
				isUUID: true,
			})
		case query.TagGroupToken:
			groupColumns = append(groupColumns, groupColumn{
				header: fmt.Sprintf("#\"%s\"", t.Tag),
				isUUID: false,
			})
		}
	}

	q, err := BuildGroupsQuery(s.db, userID, queryAST, queryAST.GroupQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := q.Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	groups := []model_api.Group{}
	totalCount := 0

	for rows.Next() {
		groupDests := make([]any, len(groupColumns))
		for i, col := range groupColumns {
			if col.isUUID {
				groupDests[i] = new(model_db.BinaryUUID)
			} else {
				groupDests[i] = new(string)
			}
		}

		var aggregatedStatus, testcaseCount, rowTotalCount int64
		scanDests := append(groupDests, &aggregatedStatus, &testcaseCount, &rowTotalCount)
		if err := rows.Scan(scanDests...); err != nil {
			continue
		}

		if totalCount == 0 && rowTotalCount > 0 {
			totalCount = int(rowTotalCount)
		}

		groupValues := []string{}
		for i, col := range groupColumns {
			if col.isUUID {
				if binUUID, ok := groupDests[i].(*model_db.BinaryUUID); ok && binUUID != nil {
					groupValues = append(groupValues, uuid.UUID(*binUUID).String())
				}
			} else {
				if strVal, ok := groupDests[i].(*string); ok && strVal != nil {
					groupValues = append(groupValues, *strVal)
				}
			}
		}

		status := TestcaseStatusToString(model_db.TestcaseStatus(aggregatedStatus))

		groups = append(groups, model_api.Group{
			Group:  strings.Join(groupValues, ", "),
			Status: status,
		})
	}

	if totalCount == 0 {
		totalCount = len(groups)
	}

	return &QueryResult[model_api.Group]{
		Results:    groups,
		TotalCount: totalCount,
	}, nil
}

func (s *QueryService) GetTestcase(ctx context.Context, userID model_db.BinaryUUID, testcaseID uuid.UUID) (*TestcaseDetail, error) {
	var testcase model_db.Testcase
	err := s.db.NewSelect().
		Model(&testcase).
		Where("? = ?", bun.Ident("id"), model_db.BinaryUUID(testcaseID)).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("testcase not found")
	}

	testcaseIDStr, _ := uuid.FromBytes(testcase.ID[:])
	sessionIDStr, _ := uuid.FromBytes(testcase.SessionID[:])

	result := &TestcaseDetail{
		ID:        testcaseIDStr.String(),
		SessionID: sessionIDStr.String(),
		Name:      testcase.Name,
		Status:    TestcaseStatusToString(testcase.Status),
		CreatedAt: testcase.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if testcase.Classname != nil {
		result.Classname = *testcase.Classname
	}
	if testcase.File != nil {
		result.File = *testcase.File
	}
	if testcase.Testsuite != nil {
		result.Testsuite = *testcase.Testsuite
	}
	if testcase.Output != nil {
		result.Output = *testcase.Output
	}
	if testcase.Baggage != nil {
		var baggage any
		if err := json.Unmarshal(testcase.Baggage, &baggage); err == nil {
			result.Baggage = baggage
		}
	}

	var labels []model_db.Label
	err = s.db.NewSelect().
		Model(&labels).
		Where("? = ?", bun.Ident("session_id"), testcase.SessionID).
		OrderBy("key", bun.OrderAsc).
		Scan(ctx)

	if err == nil && len(labels) > 0 {
		result.Labels = make(map[string]string)
		for _, label := range labels {
			if label.Value != nil {
				result.Labels[label.Key] = *label.Value
			} else {
				result.Labels[label.Key] = ""
			}
		}
	}

	return result, nil
}

func (s *QueryService) GetSession(ctx context.Context, userID model_db.BinaryUUID, sessionID uuid.UUID) (*SessionDetail, error) {
	type SessionWithStatus struct {
		model_db.Session
		AggregatedStatus *int64 `bun:"aggregated_status"`
	}

	var sessionData SessionWithStatus
	err := s.db.NewSelect().
		TableExpr("?", bun.Ident("sessions")).
		ColumnExpr("?.*", bun.Ident("sessions")).
		ColumnExpr("MIN(?) AS ?", bun.Ident("testcases.status"), bun.Ident("aggregated_status")).
		Join("LEFT JOIN ? ON ? = ?", bun.Ident("testcases"), bun.Ident("sessions.id"), bun.Ident("testcases.session_id")).
		Where("? = ?", bun.Ident("sessions.id"), model_db.BinaryUUID(sessionID)).
		Group("sessions.id").
		Scan(ctx, &sessionData)

	if err != nil {
		return nil, fmt.Errorf("session not found")
	}

	sessionIDStr, _ := uuid.FromBytes(sessionData.ID[:])

	result := &SessionDetail{
		ID:        sessionIDStr.String(),
		CreatedAt: sessionData.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if sessionData.Description != nil {
		result.Description = *sessionData.Description
	}

	if sessionData.AggregatedStatus != nil {
		result.Status = TestcaseStatusToString(model_db.TestcaseStatus(*sessionData.AggregatedStatus))
	} else {
		result.Status = "pass"
	}

	if sessionData.Baggage != nil {
		var baggage any
		if err := json.Unmarshal(sessionData.Baggage, &baggage); err == nil {
			result.Baggage = baggage
		}
	}

	var labels []model_db.Label
	err = s.db.NewSelect().
		Model(&labels).
		Where("? = ?", bun.Ident("session_id"), model_db.BinaryUUID(sessionID)).
		OrderBy("key", bun.OrderAsc).
		Scan(ctx)

	if err == nil && len(labels) > 0 {
		result.Labels = make(map[string]string)
		for _, label := range labels {
			value := ""
			if label.Value != nil {
				value = *label.Value
			}
			result.Labels[label.Key] = value
		}
	}

	return result, nil
}
