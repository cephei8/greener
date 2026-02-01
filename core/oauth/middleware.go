package oauth

import (
	"context"
	"net/http"
	"strings"
	"time"

	model_db "github.com/cephei8/greener/core/model/db"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	ContextKeyUserID   = "oauth_user_id"
	ContextKeyClientID = "oauth_client_id"
	ContextKeyScope    = "oauth_scope"
)

type oauthError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

func (s *Server) BearerAuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, oauthError{
					Error:            "invalid_request",
					ErrorDescription: "missing Authorization header",
				})
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return c.JSON(http.StatusUnauthorized, oauthError{
					Error:            "invalid_request",
					ErrorDescription: "invalid Authorization header format",
				})
			}
			accessToken := parts[1]

			ctx := context.Background()
			ti, err := s.tokenStore.GetByAccess(ctx, accessToken)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, oauthError{
					Error:            "invalid_token",
					ErrorDescription: "invalid access token",
				})
			}

			if ti.GetAccessCreateAt().Add(ti.GetAccessExpiresIn()).Before(time.Now()) {
				return c.JSON(http.StatusUnauthorized, oauthError{
					Error:            "invalid_token",
					ErrorDescription: "access token has expired",
				})
			}

			userID, err := uuid.Parse(ti.GetUserID())
			if err != nil {
				c.Logger().Errorf("Invalid user ID in token: %v", err)
				return c.JSON(http.StatusInternalServerError, oauthError{
					Error:            "server_error",
					ErrorDescription: "invalid token data",
				})
			}

			c.Set(ContextKeyUserID, model_db.BinaryUUID(userID))
			c.Set(ContextKeyClientID, ti.GetClientID())
			c.Set(ContextKeyScope, ti.GetScope())

			return next(c)
		}
	}
}

func GetOAuthUserID(c echo.Context) model_db.BinaryUUID {
	if userID, ok := c.Get(ContextKeyUserID).(model_db.BinaryUUID); ok {
		return userID
	}
	return model_db.BinaryUUID(uuid.Nil)
}

func GetOAuthClientID(c echo.Context) string {
	if clientID, ok := c.Get(ContextKeyClientID).(string); ok {
		return clientID
	}
	return ""
}

func GetOAuthScope(c echo.Context) string {
	if scope, ok := c.Get(ContextKeyScope).(string); ok {
		return scope
	}
	return ""
}

func HasScope(c echo.Context, requiredScope string) bool {
	scope := GetOAuthScope(c)
	if scope == "" {
		return false
	}

	scopes := strings.Split(scope, " ")
	for _, s := range scopes {
		if s == requiredScope {
			return true
		}
	}
	return false
}
