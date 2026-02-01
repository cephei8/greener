package core_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

"github.com/cephei8/greener/core"
	model_api "github.com/cephei8/greener/core/model/api"
	model_db "github.com/cephei8/greener/core/model/db"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type testRenderer struct{}

func (t *testRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return nil
}

func setupEchoContext(t *testing.T, method, path string, body string, authenticated bool, userID string) (echo.Context, *httptest.ResponseRecorder, *core.MockQueryServiceInterface) {
	e := echo.New()
	e.Renderer = &testRenderer{}

	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	store := sessions.NewCookieStore([]byte("test-secret"))
	sess, _ := store.Get(req, "session")
	sess.Values["authenticated"] = authenticated
	sess.Values["user_id"] = userID

	c.Set("_session_store", store)
	c.Set("session", sess)

	mockService := core.NewMockQueryServiceInterface(t)
	c.Set("queryService", mockService)

	return c, rec, mockService
}

func TestTestcasesHandler_Success(t *testing.T) {
	userID := uuid.New()
	c, rec, mockService := setupEchoContext(t, http.MethodGet, "/testcases", "", true, userID.String())

	expectedResult := &core.QueryResult[model_api.Testcase]{
		Results: []model_api.Testcase{
			{ID: uuid.New().String(), Name: "test1", Status: "pass"},
			{ID: uuid.New().String(), Name: "test2", Status: "fail"},
		},
		TotalCount: 2,
	}

	mockService.EXPECT().
		QueryTestcases(mock.Anything, model_db.BinaryUUID(userID), core.QueryParams{Query: ""}).
		Return(expectedResult, nil)

	err := core.TestcasesHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestTestcasesHandler_WithQuery(t *testing.T) {
	userID := uuid.New()
	c, rec, mockService := setupEchoContext(t, http.MethodGet, "/testcases?query=status%3D%22pass%22", "", true, userID.String())

	expectedResult := &core.QueryResult[model_api.Testcase]{
		Results:    []model_api.Testcase{},
		TotalCount: 0,
	}

	mockService.EXPECT().
		QueryTestcases(mock.Anything, model_db.BinaryUUID(userID), core.QueryParams{Query: `status="pass"`}).
		Return(expectedResult, nil)

	err := core.TestcasesHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestTestcasesHandler_QueryError(t *testing.T) {
	userID := uuid.New()
	c, rec, mockService := setupEchoContext(t, http.MethodGet, "/testcases", "", true, userID.String())

	mockService.EXPECT().
		QueryTestcases(mock.Anything, model_db.BinaryUUID(userID), core.QueryParams{Query: ""}).
		Return(nil, errors.New("invalid query"))

	err := core.TestcasesHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTestcasesHandler_Unauthenticated(t *testing.T) {
	c, rec, _ := setupEchoContext(t, http.MethodGet, "/testcases", "", false, "")

	err := core.TestcasesHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/login", rec.Header().Get("Location"))
}

func TestSessionsHandler_Success(t *testing.T) {
	userID := uuid.New()
	c, rec, mockService := setupEchoContext(t, http.MethodGet, "/sessions", "", true, userID.String())

	expectedResult := &core.QueryResult[model_api.Session]{
		Results: []model_api.Session{
			{ID: uuid.New().String(), Description: "session1", Status: "pass"},
		},
		TotalCount: 1,
	}

	mockService.EXPECT().
		QuerySessions(mock.Anything, model_db.BinaryUUID(userID), core.QueryParams{Query: ""}).
		Return(expectedResult, nil)

	err := core.SessionsHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestSessionsHandler_QueryError(t *testing.T) {
	userID := uuid.New()
	c, rec, mockService := setupEchoContext(t, http.MethodGet, "/sessions", "", true, userID.String())

	mockService.EXPECT().
		QuerySessions(mock.Anything, model_db.BinaryUUID(userID), core.QueryParams{Query: ""}).
		Return(nil, errors.New("query error"))

	err := core.SessionsHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGroupsHandler_Success(t *testing.T) {
	userID := uuid.New()
	query := `group_by(#"env")`
	c, rec, mockService := setupEchoContext(t, http.MethodGet, "/groups?query="+url.QueryEscape(query), "", true, userID.String())

	expectedResult := &core.QueryResult[model_api.Group]{
		Results: []model_api.Group{
			{Group: "production", Status: "pass"},
			{Group: "staging", Status: "fail"},
		},
		TotalCount: 2,
	}

	mockService.EXPECT().
		QueryGroups(mock.Anything, model_db.BinaryUUID(userID), core.QueryParams{Query: query}).
		Return(expectedResult, nil)

	err := core.GroupsHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGroupsHandler_EmptyQuery(t *testing.T) {
	userID := uuid.New()
	c, rec, _ := setupEchoContext(t, http.MethodGet, "/groups", "", true, userID.String())

	err := core.GroupsHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGroupsHandler_QueryError(t *testing.T) {
	userID := uuid.New()
	query := `invalid`
	c, rec, mockService := setupEchoContext(t, http.MethodGet, "/groups?query="+url.QueryEscape(query), "", true, userID.String())

	mockService.EXPECT().
		QueryGroups(mock.Anything, model_db.BinaryUUID(userID), core.QueryParams{Query: query}).
		Return(nil, errors.New("invalid query"))

	err := core.GroupsHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTestcaseDetailHandler_Success(t *testing.T) {
	userID := uuid.New()
	testcaseID := uuid.New()
	c, rec, mockService := setupEchoContext(t, http.MethodGet, "/testcases/"+testcaseID.String()+"/details", "", true, userID.String())
	c.SetParamNames("id")
	c.SetParamValues(testcaseID.String())

	expectedResult := &core.TestcaseDetail{
		ID:        testcaseID.String(),
		SessionID: uuid.New().String(),
		Name:      "test_login",
		Status:    "pass",
		Classname: "TestAuth",
		File:      "test_auth.py",
		CreatedAt: "2024-01-01 12:00:00",
	}

	mockService.EXPECT().
		GetTestcase(mock.Anything, model_db.BinaryUUID(userID), testcaseID).
		Return(expectedResult, nil)

	err := core.TestcaseDetailHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestTestcaseDetailHandler_NotFound(t *testing.T) {
	userID := uuid.New()
	testcaseID := uuid.New()
	c, _, mockService := setupEchoContext(t, http.MethodGet, "/testcases/"+testcaseID.String()+"/details", "", true, userID.String())
	c.SetParamNames("id")
	c.SetParamValues(testcaseID.String())

	mockService.EXPECT().
		GetTestcase(mock.Anything, model_db.BinaryUUID(userID), testcaseID).
		Return(nil, errors.New("not found"))

	err := core.TestcaseDetailHandler(c)

	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, httpErr.Code)
}

func TestTestcaseDetailHandler_InvalidID(t *testing.T) {
	userID := uuid.New()
	c, _, _ := setupEchoContext(t, http.MethodGet, "/testcases/invalid-id/details", "", true, userID.String())
	c.SetParamNames("id")
	c.SetParamValues("invalid-id")

	err := core.TestcaseDetailHandler(c)

	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}

func TestSessionDetailHandler_Success(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	c, rec, mockService := setupEchoContext(t, http.MethodGet, "/sessions/"+sessionID.String()+"/details", "", true, userID.String())
	c.SetParamNames("id")
	c.SetParamValues(sessionID.String())

	expectedResult := &core.SessionDetail{
		ID:          sessionID.String(),
		Description: "Test session",
		Status:      "pass",
		Labels:      map[string]string{"env": "prod"},
		CreatedAt:   "2024-01-01 12:00:00",
	}

	mockService.EXPECT().
		GetSession(mock.Anything, model_db.BinaryUUID(userID), sessionID).
		Return(expectedResult, nil)

	err := core.SessionDetailHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestSessionDetailHandler_NotFound(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	c, _, mockService := setupEchoContext(t, http.MethodGet, "/sessions/"+sessionID.String()+"/details", "", true, userID.String())
	c.SetParamNames("id")
	c.SetParamValues(sessionID.String())

	mockService.EXPECT().
		GetSession(mock.Anything, model_db.BinaryUUID(userID), sessionID).
		Return(nil, errors.New("not found"))

	err := core.SessionDetailHandler(c)

	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, httpErr.Code)
}
