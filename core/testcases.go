package core

import (
	"context"
	"fmt"
	"net/http"

	"git.sr.ht/~cephei8/greener/core/model/api"
	"git.sr.ht/~cephei8/greener/core/model/db"
	"git.sr.ht/~cephei8/greener/core/query"
	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

func TestcasesHandler(c echo.Context) error {
	sess, _ := session.Get("session", c)
	if auth, ok := sess.Values["authenticated"].(bool); !ok || !auth {
		return c.Redirect(http.StatusFound, "/login")
	}

	userIdStr, ok := sess.Values["user_id"].(string)
	if !ok {
		sess.Values["authenticated"] = false
		sess.Save(c.Request(), c.Response())
		return c.Redirect(http.StatusFound, "/login")
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		c.Logger().Errorf("Invalid user_id in session: %v", err)
		sess.Values["authenticated"] = false
		sess.Save(c.Request(), c.Response())
		return c.Redirect(http.StatusFound, "/login")
	}

	db := c.Get("db").(*bun.DB)
	ctx := context.Background()

	queryStr := c.FormValue("query")

	isHTMX := c.Request().Header.Get("HX-Request") == "true"
	templateName := "testcases.html"
	if isHTMX {
		templateName = "testcases_table.html"
	}

	var queryAST query.Query
	if queryStr != "" {
		parser := query.NewParser(queryStr)
		parsedQuery, err := parser.Parse()
		if err != nil {
			c.Response().Header().Set("Content-Type", "text/html")
			return c.HTML(http.StatusBadRequest, fmt.Sprintf("<span>Invalid query: %v</span>", err))
		}

		if err := query.Validate(parsedQuery, query.QueryTypeTestcase); err != nil {
			c.Response().Header().Set("Content-Type", "text/html")
			return c.HTML(http.StatusBadRequest, fmt.Sprintf("<span>Invalid query: %v</span>", err))
		}

		queryAST = parsedQuery
	}

	q, err := BuildTestcasesQuery(db, model_db.BinaryUUID(userId), queryAST)
	if err != nil {
		c.Response().Header().Set("Content-Type", "text/html")
		return c.HTML(http.StatusBadRequest, fmt.Sprintf("<span>Invalid query: %v</span>", err))
	}

	type TestcaseResult struct {
		model_db.Testcase
		TotalCount       int64 `bun:"total_count"`
		AggregatedStatus int64 `bun:"aggregated_status"`
	}

	var results []TestcaseResult
	err = q.Scan(ctx, &results)
	if err != nil {
		c.Logger().Errorf("Failed to execute query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to execute query")
	}

	testcases := []model_api.Testcase{}
	totalCount := 0

	for _, result := range results {
		if totalCount == 0 && result.TotalCount > 0 {
			totalCount = int(result.TotalCount)
		}

		testcaseID, err := uuid.FromBytes(result.ID[:])
		if err != nil {
			c.Logger().Errorf("Failed to parse testcase UUID: %v", err)
			continue
		}

		sessionID, err := uuid.FromBytes(result.SessionID[:])
		if err != nil {
			c.Logger().Errorf("Failed to parse session UUID: %v", err)
			continue
		}

		status := TestcaseStatusToString(result.Status)

		testcases = append(testcases, model_api.Testcase{
			ID:        testcaseID.String(),
			SessionID: sessionID.String(),
			Name:      result.Name,
			Status:    status,
		})
	}

	return c.Render(http.StatusOK, templateName, map[string]any{
		"Testcases":    testcases,
		"LoadedCount":  len(testcases),
		"TotalRecords": totalCount,
	})
}
