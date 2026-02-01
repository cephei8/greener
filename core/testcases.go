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

	svc := c.Get("queryService").(QueryServiceInterface)
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

	result, err := svc.QueryTestcases(ctx, model_db.BinaryUUID(userId), QueryParams{
		Query: queryStr,
	})
	if err != nil {
		c.Response().Header().Set("Content-Type", "text/html")
		return c.HTML(http.StatusBadRequest, fmt.Sprintf("<span>%v</span>", err))
	}

	return c.Render(http.StatusOK, templateName, map[string]any{
		"Testcases":    result.Results,
		"LoadedCount":  len(result.Results),
		"TotalRecords": result.TotalCount,
		"Query":        queryStr,
		"ActivePage":   "testcases",
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

	svc := c.Get("queryService").(QueryServiceInterface)
	ctx := context.Background()

	result, err := svc.GetTestcase(ctx, model_db.BinaryUUID(userId), testcaseId)
	if err != nil {
		c.Logger().Errorf("Failed to fetch testcase: %v", err)
		return echo.NewHTTPError(http.StatusNotFound, "Testcase not found")
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

	return c.Render(http.StatusOK, "testcase_detail.html", map[string]any{
		"Testcase": map[string]any{
			"ID":        result.ID,
			"SessionID": result.SessionID,
			"Name":      result.Name,
			"Classname": result.Classname,
			"File":      result.File,
			"Testsuite": result.Testsuite,
			"Status":    result.Status,
			"Baggage":   baggageStr,
			"Output":    result.Output,
			"CreatedAt": result.CreatedAt,
		},
		"Labels":     labelList,
		"ActivePage": "testcases",
	})
}
