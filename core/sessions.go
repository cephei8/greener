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

func SessionsHandler(c echo.Context) error {
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
	c.Logger().Infof("Sessions handler called, queryStr='%s', userID=%s", queryStr, userId)

	isHTMX := c.Request().Header.Get("HX-Request") == "true"
	templateName := "sessions.html"
	if isHTMX {
		templateName = "sessions_table.html"
	}

	var queryAST query.Query

	if queryStr != "" {
		parser := query.NewParser(queryStr)
		parsedQuery, err := parser.Parse()
		if err != nil {
			c.Response().Header().Set("Content-Type", "text/html")
			return c.HTML(http.StatusBadRequest, fmt.Sprintf("<span>Invalid query: %v</span>", err))
		}

		if err := query.Validate(parsedQuery, query.QueryTypeSession); err != nil {
			c.Response().Header().Set("Content-Type", "text/html")
			return c.HTML(http.StatusBadRequest, fmt.Sprintf("<span>Invalid query: %v</span>", err))
		}

		queryAST = parsedQuery
	}

	q, err := BuildSessionsQuery(db, model_db.BinaryUUID(userId), queryAST)
	if err != nil {
		c.Response().Header().Set("Content-Type", "text/html")
		return c.HTML(http.StatusBadRequest, fmt.Sprintf("<span>Failed to build query: %v</span>", err))
	}

	type SessionResult struct {
		model_db.Session
		AggregatedStatus int64 `bun:"aggregated_status"`
		TotalCount       int64 `bun:"total_count"`
	}

	var results []SessionResult
	err = q.Scan(ctx, &results)
	if err != nil {
		c.Logger().Errorf("Failed to execute query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to execute query")
	}

	sessions := []model_api.Session{}
	totalCount := 0

	for _, result := range results {
		if totalCount == 0 && result.TotalCount > 0 {
			totalCount = int(result.TotalCount)
		}

		sessionID, err := uuid.FromBytes(result.ID[:])
		if err != nil {
			c.Logger().Errorf("Failed to parse session UUID: %v", err)
			continue
		}

		status := TestcaseStatusToString(model_db.TestcaseStatus(result.AggregatedStatus))

		desc := "No description"
		if result.Description != nil {
			desc = *result.Description
		}

		sessions = append(sessions, model_api.Session{
			ID:          sessionID.String(),
			Description: desc,
			Status:      status,
			CreatedAt:   result.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.Logger().Infof("Returning %d sessions, totalCount=%d, template=%s", len(sessions), totalCount, templateName)
	return c.Render(http.StatusOK, templateName, map[string]any{
		"Sessions":     sessions,
		"LoadedCount":  len(sessions),
		"TotalRecords": totalCount,
		"Query":        queryStr,
		"ActivePage":   "sessions",
	})
}

func SessionDetailHandler(c echo.Context) error {
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

	sessionIdStr := c.Param("id")
	sessionId, err := uuid.Parse(sessionIdStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid session ID")
	}

	db := c.Get("db").(*bun.DB)
	ctx := context.Background()

	type SessionWithStatus struct {
		model_db.Session
		AggregatedStatus *int64 `bun:"aggregated_status"`
	}

	var sessionData SessionWithStatus
	err = db.NewSelect().
		TableExpr("?", bun.Ident("sessions")).
		ColumnExpr("?.*", bun.Ident("sessions")).
		ColumnExpr("MIN(?) AS ?", bun.Ident("testcases.status"), bun.Ident("aggregated_status")).
		Join("LEFT JOIN ? ON ? = ?", bun.Ident("testcases"), bun.Ident("sessions.id"), bun.Ident("testcases.session_id")).
		Where("? = ?", bun.Ident("sessions.id"), model_db.BinaryUUID(sessionId)).
		Where("? = ?", bun.Ident("sessions.user_id"), model_db.BinaryUUID(userId)).
		Group("sessions.id").
		Scan(ctx, &sessionData)

	if err != nil {
		c.Logger().Errorf("Failed to fetch session: %v", err)
		return echo.NewHTTPError(http.StatusNotFound, "Session not found")
	}

	sessionIDStr, err := uuid.FromBytes(sessionData.ID[:])
	if err != nil {
		c.Logger().Errorf("Failed to parse session UUID: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to parse session ID")
	}

	var baggageStr string
	if sessionData.Baggage != nil {
		var baggageData interface{}
		if err := json.Unmarshal(sessionData.Baggage, &baggageData); err == nil {
			if formatted, err := json.MarshalIndent(baggageData, "", "  "); err == nil {
				baggageStr = string(formatted)
			} else {
				baggageStr = string(sessionData.Baggage)
			}
		} else {
			baggageStr = string(sessionData.Baggage)
		}
	}

	var descriptionStr string
	if sessionData.Description != nil {
		descriptionStr = *sessionData.Description
	}

	var statusStr string
	if sessionData.AggregatedStatus != nil {
		statusStr = TestcaseStatusToString(model_db.TestcaseStatus(*sessionData.AggregatedStatus))
	} else {
		statusStr = "pass" // Default status if no testcases
	}

	// Fetch labels for this session
	var labels []model_db.Label
	err = db.NewSelect().
		Model(&labels).
		Where("? = ?", bun.Ident("session_id"), model_db.BinaryUUID(sessionId)).
		Where("? = ?", bun.Ident("user_id"), model_db.BinaryUUID(userId)).
		OrderBy("key", bun.OrderAsc).
		Scan(ctx)

	if err != nil {
		c.Logger().Errorf("Failed to fetch labels: %v", err)
		// Don't fail the request, just log the error
		labels = []model_db.Label{}
	}

	type LabelData struct {
		Key   string
		Value string
	}

	labelList := []LabelData{}
	for _, label := range labels {
		value := ""
		if label.Value != nil {
			value = *label.Value
		}
		labelList = append(labelList, LabelData{
			Key:   label.Key,
			Value: value,
		})
	}

	return c.Render(http.StatusOK, "session_detail.html", map[string]any{
		"Session": map[string]any{
			"ID":          sessionIDStr.String(),
			"Description": descriptionStr,
			"Status":      statusStr,
			"Baggage":     baggageStr,
			"CreatedAt":   sessionData.CreatedAt.Format("2006-01-02 15:04:05"),
		},
		"Labels":     labelList,
		"ActivePage": "sessions",
	})
}
