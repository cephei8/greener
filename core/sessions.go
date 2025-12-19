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
	})
}
