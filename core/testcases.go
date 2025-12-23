package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cephei8/greener/core/model/api"
	"github.com/cephei8/greener/core/model/db"
	"github.com/cephei8/greener/core/query"
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
	if queryStr == "" {
		queryStr = c.QueryParam("query")
	}

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
			CreatedAt: result.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return c.Render(http.StatusOK, templateName, map[string]any{
		"Testcases":    testcases,
		"LoadedCount":  len(testcases),
		"TotalRecords": totalCount,
		"Query":        queryStr,
	})
}

func TestcaseDetailHandler(c echo.Context) error {
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

	testcaseIdStr := c.Param("id")
	testcaseId, err := uuid.Parse(testcaseIdStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid testcase ID")
	}

	db := c.Get("db").(*bun.DB)
	ctx := context.Background()

	var testcase model_db.Testcase
	err = db.NewSelect().
		Model(&testcase).
		Where("? = ?", bun.Ident("id"), model_db.BinaryUUID(testcaseId)).
		Where("? = ?", bun.Ident("user_id"), model_db.BinaryUUID(userId)).
		Scan(ctx)

	if err != nil {
		c.Logger().Errorf("Failed to fetch testcase: %v", err)
		return echo.NewHTTPError(http.StatusNotFound, "Testcase not found")
	}

	testcaseIDStr, err := uuid.FromBytes(testcase.ID[:])
	if err != nil {
		c.Logger().Errorf("Failed to parse testcase UUID: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to parse testcase ID")
	}

	sessionIDStr, err := uuid.FromBytes(testcase.SessionID[:])
	if err != nil {
		c.Logger().Errorf("Failed to parse session UUID: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to parse session ID")
	}

	status := TestcaseStatusToString(testcase.Status)

	var baggageStr string
	if testcase.Baggage != nil {
		var baggageData interface{}
		if err := json.Unmarshal(testcase.Baggage, &baggageData); err == nil {
			if formatted, err := json.MarshalIndent(baggageData, "", "  "); err == nil {
				baggageStr = string(formatted)
			} else {
				baggageStr = string(testcase.Baggage)
			}
		} else {
			baggageStr = string(testcase.Baggage)
		}
	}

	var outputStr string
	if testcase.Output != nil {
		outputStr = *testcase.Output
	}

	var classnameStr string
	if testcase.Classname != nil {
		classnameStr = *testcase.Classname
	}

	var fileStr string
	if testcase.File != nil {
		fileStr = *testcase.File
	}

	var testsuiteStr string
	if testcase.Testsuite != nil {
		testsuiteStr = *testcase.Testsuite
	}

	return c.Render(http.StatusOK, "testcase_detail.html", map[string]any{
		"Testcase": map[string]any{
			"ID":        testcaseIDStr.String(),
			"SessionID": sessionIDStr.String(),
			"Name":      testcase.Name,
			"Classname": classnameStr,
			"File":      fileStr,
			"Testsuite": testsuiteStr,
			"Status":    status,
			"Baggage":   baggageStr,
			"Output":    outputStr,
			"CreatedAt": testcase.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}
