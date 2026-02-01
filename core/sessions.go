package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	model_db "github.com/cephei8/greener/core/model/db"
	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
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

	svc := c.Get("queryService").(QueryServiceInterface)
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

	result, err := svc.QuerySessions(ctx, model_db.BinaryUUID(userId), QueryParams{
		Query: queryStr,
	})
	if err != nil {
		c.Response().Header().Set("Content-Type", "text/html")
		return c.HTML(http.StatusBadRequest, fmt.Sprintf("<span>%v</span>", err))
	}

	// Add "No description" for empty descriptions for template display
	sessions := result.Results
	for i := range sessions {
		if sessions[i].Description == "" {
			sessions[i].Description = "No description"
		}
	}

	c.Logger().Infof("Returning %d sessions, totalCount=%d, template=%s", len(sessions), result.TotalCount, templateName)
	return c.Render(http.StatusOK, templateName, map[string]any{
		"Sessions":     sessions,
		"LoadedCount":  len(sessions),
		"TotalRecords": result.TotalCount,
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

	svc := c.Get("queryService").(QueryServiceInterface)
	ctx := context.Background()

	result, err := svc.GetSession(ctx, model_db.BinaryUUID(userId), sessionId)
	if err != nil {
		c.Logger().Errorf("Failed to fetch session: %v", err)
		return echo.NewHTTPError(http.StatusNotFound, "Session not found")
	}

	var baggageStr string
	if result.Baggage != nil {
		if formatted, err := json.MarshalIndent(result.Baggage, "", "  "); err == nil {
			baggageStr = string(formatted)
		}
	}

	type LabelData struct {
		Key   string
		Value string
	}

	labelList := []LabelData{}
	for k, v := range result.Labels {
		labelList = append(labelList, LabelData{
			Key:   k,
			Value: v,
		})
	}

	return c.Render(http.StatusOK, "session_detail.html", map[string]any{
		"Session": map[string]any{
			"ID":          result.ID,
			"Description": result.Description,
			"Status":      result.Status,
			"Baggage":     baggageStr,
			"CreatedAt":   result.CreatedAt,
		},
		"Labels":     labelList,
		"ActivePage": "sessions",
	})
}
