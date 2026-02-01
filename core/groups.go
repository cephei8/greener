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

	isHTMX := c.Request().Header.Get("HX-Request") == "true"
	templateName := "groups.html"
	if isHTMX {
		templateName = "groups_table.html"
	}

	if queryStr == "" {
		return c.Render(http.StatusOK, templateName, map[string]any{
			"Groups":       []model_api.Group{},
			"LoadedCount":  0,
			"TotalRecords": 0,
			"Query":        "",
			"ActivePage":   "groups",
		})
	}

	result, err := svc.QueryGroups(ctx, model_db.BinaryUUID(userId), QueryParams{
		Query: queryStr,
	})
	if err != nil {
		c.Response().Header().Set("Content-Type", "text/html")
		return c.HTML(http.StatusBadRequest, fmt.Sprintf("<span>%v</span>", err))
	}

	return c.Render(http.StatusOK, templateName, map[string]any{
		"Groups":       result.Results,
		"LoadedCount":  len(result.Results),
		"TotalRecords": result.TotalCount,
		"Query":        queryStr,
		"ActivePage":   "groups",
	})
}
