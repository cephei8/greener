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
	auth, _ := sess.Values["authenticated"].(bool)

	if !auth && !AllowUnauthenticatedViewers(c) {
		return c.Redirect(http.StatusFound, "/login")
	}

	svc := c.Get("queryService").(QueryServiceInterface)
	ctx := context.Background()

	queryStr := c.FormValue("query")
	if queryStr == "" {
		queryStr = c.QueryParam("query")
	}

	isHTMX := c.Request().Header.Get("HX-Request") == "true"
	templateName := "sessions.html"
	if isHTMX {
		templateName = "sessions_table.html"
	}

	result, err := svc.QuerySessions(ctx, model_db.BinaryUUID(uuid.Nil), QueryParams{
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

	return c.Render(http.StatusOK, templateName, map[string]any{
		"Sessions":        sessions,
		"LoadedCount":     len(sessions),
		"TotalRecords":    result.TotalCount,
		"Query":           queryStr,
		"ActivePage":      "sessions",
		"IsAuthenticated": auth,
	})
}

func SessionDetailHandler(c echo.Context) error {
	sess, _ := session.Get("session", c)
	auth, _ := sess.Values["authenticated"].(bool)

	if !auth && !AllowUnauthenticatedViewers(c) {
		return c.Redirect(http.StatusFound, "/login")
	}

	sessionIdStr := c.Param("id")
	sessionId, err := uuid.Parse(sessionIdStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid session ID")
	}

	svc := c.Get("queryService").(QueryServiceInterface)
	ctx := context.Background()

	result, err := svc.GetSession(ctx, model_db.BinaryUUID(uuid.Nil), sessionId)
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
		"Labels":          labelList,
		"ActivePage":      "sessions",
		"IsAuthenticated": auth,
	})
}
