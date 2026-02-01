package core

import (
	"context"
	"fmt"
	"net/http"

	model_api "github.com/cephei8/greener/core/model/api"
	model_db "github.com/cephei8/greener/core/model/db"
	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func GroupsHandler(c echo.Context) error {
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
	templateName := "groups.html"
	if isHTMX {
		templateName = "groups_table.html"
	}

	if queryStr == "" {
		return c.Render(http.StatusOK, templateName, map[string]any{
			"Groups":          []model_api.Group{},
			"LoadedCount":     0,
			"TotalRecords":    0,
			"Query":           "",
			"ActivePage":      "groups",
			"IsAuthenticated": auth,
		})
	}

	result, err := svc.QueryGroups(ctx, model_db.BinaryUUID(uuid.Nil), QueryParams{
		Query: queryStr,
	})
	if err != nil {
		c.Response().Header().Set("Content-Type", "text/html")
		return c.HTML(http.StatusBadRequest, fmt.Sprintf("<span>%v</span>", err))
	}

	return c.Render(http.StatusOK, templateName, map[string]any{
		"Groups":          result.Results,
		"LoadedCount":     len(result.Results),
		"TotalRecords":    result.TotalCount,
		"Query":           queryStr,
		"ActivePage":      "groups",
		"IsAuthenticated": auth,
	})
}
