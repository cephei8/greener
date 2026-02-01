package core

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"net/http"

	"github.com/cephei8/greener/core/model/db"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/pbkdf2"
)

func IndexHandler(c echo.Context) error {
	sess, _ := session.Get("session", c)
	if auth, ok := sess.Values["authenticated"].(bool); ok && auth {
		return c.Redirect(http.StatusFound, "/sessions")
	}

	if AllowUnauthenticatedViewers(c) {
		return c.Redirect(http.StatusFound, "/sessions")
	}

	return c.Redirect(http.StatusFound, "/login")
}

func LoginPageHandler(c echo.Context) error {
	sess, _ := session.Get("session", c)
	if auth, ok := sess.Values["authenticated"].(bool); ok && auth {
		return c.Redirect(http.StatusFound, "/sessions")
	}

	return c.Render(http.StatusOK, "login.html", map[string]any{
		"ShowSidebar": false,
	})
}

func LoginHandler(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	db := c.Get("db").(*bun.DB)
	ctx := context.Background()

	var user model_db.User
	err := db.NewSelect().
		Model(&user).
		Where("? = ?", bun.Ident("username"), username).
		Scan(ctx)
	if err != nil {
		c.Logger().Errorf("Failed to find user %s: %v", username, err)
		c.Response().Header().Set("Content-Type", "text/html")
		return c.HTML(http.StatusUnauthorized, `<span>Invalid username or password</span>`)
	}

	passwordHash := pbkdf2.Key([]byte(password), user.PasswordSalt, 100000, 32, sha256.New)

	if subtle.ConstantTimeCompare(passwordHash, user.PasswordHash) != 1 {
		c.Response().Header().Set("Content-Type", "text/html")
		return c.HTML(http.StatusUnauthorized, `<span>Invalid username or password</span>`)
	}

	sess, _ := session.Get("session", c)
	sess.Values["authenticated"] = true
	sess.Values["username"] = username
	sess.Values["user_id"] = user.ID.String()
	sess.Values["role"] = string(user.Role)
	sess.Save(c.Request(), c.Response())

	c.Response().Header().Set("HX-Redirect", "/sessions")
	return c.NoContent(http.StatusOK)
}

func LogoutHandler(c echo.Context) error {
	sess, _ := session.Get("session", c)
	sess.Values["authenticated"] = false
	delete(sess.Values, "username")
	delete(sess.Values, "user_id")
	delete(sess.Values, "role")
	sess.Save(c.Request(), c.Response())

	c.Response().Header().Set("HX-Redirect", "/login")
	return c.NoContent(http.StatusOK)
}
