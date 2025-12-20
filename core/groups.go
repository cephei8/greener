package core

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"git.sr.ht/~cephei8/greener/core/model/api"
	"git.sr.ht/~cephei8/greener/core/model/db"
	"git.sr.ht/~cephei8/greener/core/query"
	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
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

	db := c.Get("db").(*bun.DB)
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
		})
	}

	parser := query.NewParser(queryStr)
	queryAST, err := parser.Parse()
	if err != nil {
		c.Response().Header().Set("Content-Type", "text/html")
		return c.HTML(http.StatusBadRequest, fmt.Sprintf("<span>Invalid query: %v</span>", err))
	}

	if err := query.Validate(queryAST, query.QueryTypeGroup); err != nil {
		c.Response().Header().Set("Content-Type", "text/html")
		return c.HTML(http.StatusBadRequest, fmt.Sprintf("<span>Invalid query: %v</span>", err))
	}

	if queryAST.GroupQuery == nil {
		return c.Render(http.StatusOK, templateName, map[string]any{
			"Groups":       []model_api.Group{},
			"LoadedCount":  0,
			"TotalRecords": 0,
			"Query":        queryStr,
		})
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

	q, err := BuildGroupsQuery(db, model_db.BinaryUUID(userId), queryAST, queryAST.GroupQuery)
	if err != nil {
		c.Response().Header().Set("Content-Type", "text/html")
		return c.HTML(http.StatusBadRequest, fmt.Sprintf("<span>Failed to build query: %v</span>", err))
	}

	rows, err := q.Rows(ctx)
	if err != nil {
		c.Logger().Errorf("Failed to execute query: %v", err)
		c.Response().Header().Set("Content-Type", "text/html")
		return c.HTML(http.StatusInternalServerError, "<span>Failed to execute query</span>")
	}
	defer rows.Close()

	groups := []model_api.Group{}
	totalCount := 0
	aggregatedStatus := -1

	for rows.Next() {
		values := make([]any, len(groupColumns)+3) // + group_status, total_count, aggregated_status
		valuePtrs := make([]any, len(values))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			c.Logger().Errorf("Failed to scan row: %v", err)
			continue
		}

		groupValues := []string{}
		for i := 0; i < len(groupColumns); i++ {
			if values[i] != nil {
				if groupColumns[i].isUUID {
					if bytesVal, ok := values[i].([]byte); ok {
						if len(bytesVal) == 16 {
							u, err := uuid.FromBytes(bytesVal)
							if err == nil {
								groupValues = append(groupValues, u.String())
								continue
							}
						}
					}
				}
				groupValues = append(groupValues, fmt.Sprintf("%v", values[i]))
			} else {
				groupValues = append(groupValues, "null")
			}
		}

		statusInt := values[len(groupColumns)].(int64)
		status := TestcaseStatusToString(model_db.TestcaseStatus(statusInt))

		if totalCount == 0 {
			totalCount = int(values[len(groupColumns)+1].(int64))
		}
		if aggregatedStatus == -1 {
			aggregatedStatus = int(values[len(groupColumns)+2].(int64))
		}

		groups = append(groups, model_api.Group{
			Status: status,
			Group:  strings.Join(groupValues, ", "),
		})
	}

	return c.Render(http.StatusOK, templateName, map[string]any{
		"Groups":       groups,
		"LoadedCount":  len(groups),
		"TotalRecords": totalCount,
		"Query":        queryStr,
	})
}
