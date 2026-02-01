package core_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cephei8/greener/core"
	model_db "github.com/cephei8/greener/core/model/db"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

type apiKeyTestRenderer struct{}

func (t *apiKeyTestRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return nil
}

func setupAPIKeyContext(t *testing.T, method, path string, body string, authenticated bool, userID string, role string, db *bun.DB) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	e.Renderer = &apiKeyTestRenderer{}

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
	sess.Values["role"] = role

	c.Set("_session_store", store)
	c.Set("session", sess)
	c.Set("db", db)

	return c, rec
}

func TestCreateAPIKeyHandler_EditorCanCreate(t *testing.T) {
	userID := uuid.New()

	c, rec := setupAPIKeyContext(t, http.MethodPost, "/api-keys/create", "description=test+key", true, userID.String(), string(model_db.RoleEditor), nil)

	defer func() {
		if r := recover(); r != nil {
			assert.NotEqual(t, http.StatusForbidden, rec.Code, "Editor should pass role check")
		}
	}()

	err := core.CreateAPIKeyHandler(c)
	require.NoError(t, err)
	assert.NotEqual(t, http.StatusForbidden, rec.Code, "Editor should be able to create API keys")
}

func TestCreateAPIKeyHandler_ViewerCannotCreate(t *testing.T) {
	userID := uuid.New()

	c, rec := setupAPIKeyContext(t, http.MethodPost, "/api-keys/create", "description=test+key", true, userID.String(), string(model_db.RoleViewer), nil)

	err := core.CreateAPIKeyHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code, "Viewer should not be able to create API keys")
	assert.Contains(t, rec.Body.String(), "Viewers cannot create API keys")
}

func TestCreateAPIKeyHandler_Unauthenticated(t *testing.T) {
	c, rec := setupAPIKeyContext(t, http.MethodPost, "/api-keys/create", "", false, "", "", nil)

	err := core.CreateAPIKeyHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCreateAPIKeyHandler_InvalidUserID(t *testing.T) {
	c, rec := setupAPIKeyContext(t, http.MethodPost, "/api-keys/create", "", true, "invalid-uuid", string(model_db.RoleEditor), nil)

	err := core.CreateAPIKeyHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCreateAPIKeyHandler_EmptyRole(t *testing.T) {
	userID := uuid.New()

	c, rec := setupAPIKeyContext(t, http.MethodPost, "/api-keys/create", "description=test+key", true, userID.String(), "", nil)

	defer func() {
		if r := recover(); r != nil {
			assert.NotEqual(t, http.StatusForbidden, rec.Code, "Empty role should pass role check")
		}
	}()

	err := core.CreateAPIKeyHandler(c)
	require.NoError(t, err)
	assert.NotEqual(t, http.StatusForbidden, rec.Code, "Empty role should not be treated as viewer")
}
