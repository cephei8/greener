package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func (s *Server) MetadataHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.Metadata())
}

func (s *Server) AuthorizePageHandler(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login?return_to="+c.Request().URL.String())
	}

	auth, ok := sess.Values["authenticated"].(bool)
	if !ok || !auth {
		return c.Redirect(http.StatusFound, "/login?return_to="+c.Request().URL.String())
	}

	clientID := c.QueryParam("client_id")
	redirectURI := c.QueryParam("redirect_uri")
	responseType := c.QueryParam("response_type")
	scope := c.QueryParam("scope")
	state := c.QueryParam("state")
	codeChallenge := c.QueryParam("code_challenge")
	codeChallengeMethod := c.QueryParam("code_challenge_method")

	if clientID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "client_id is required")
	}
	if redirectURI == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "redirect_uri is required")
	}
	if responseType != "code" {
		return echo.NewHTTPError(http.StatusBadRequest, "only response_type=code is supported")
	}
	if codeChallenge == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "code_challenge is required (PKCE)")
	}
	if codeChallengeMethod != "S256" {
		return echo.NewHTTPError(http.StatusBadRequest, "only code_challenge_method=S256 is supported")
	}

	ctx := context.Background()
	if err := s.clientStore.ValidateRedirectURI(ctx, clientID, redirectURI); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid client_id or redirect_uri")
	}

	clientName, err := s.clientStore.GetClientName(ctx, clientID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid client_id")
	}

	scopes := []string{}
	if scope != "" {
		scopes = strings.Split(scope, " ")
	}

	return c.Render(http.StatusOK, "oauth_authorize.html", map[string]any{
		"ClientName":          clientName,
		"ClientID":            clientID,
		"RedirectURI":         redirectURI,
		"Scope":               scope,
		"Scopes":              scopes,
		"State":               state,
		"CodeChallenge":       codeChallenge,
		"CodeChallengeMethod": codeChallengeMethod,
	})
}

func (s *Server) AuthorizeHandler(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, oauthError{
			Error:            "access_denied",
			ErrorDescription: "not authenticated",
		})
	}

	auth, ok := sess.Values["authenticated"].(bool)
	if !ok || !auth {
		return c.JSON(http.StatusUnauthorized, oauthError{
			Error:            "access_denied",
			ErrorDescription: "not authenticated",
		})
	}

	userIDStr, ok := sess.Values["user_id"].(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, oauthError{
			Error:            "access_denied",
			ErrorDescription: "invalid session",
		})
	}

	decision := c.FormValue("decision")
	if decision == "deny" {
		redirectURI := c.FormValue("redirect_uri")
		state := c.FormValue("state")
		errorRedirect := redirectURI + "?error=access_denied"
		if state != "" {
			errorRedirect += "&state=" + state
		}
		return c.Redirect(http.StatusFound, errorRedirect)
	}

	clientID := c.FormValue("client_id")
	redirectURI := c.FormValue("redirect_uri")
	scope := c.FormValue("scope")
	state := c.FormValue("state")
	codeChallenge := c.FormValue("code_challenge")
	codeChallengeMethod := c.FormValue("code_challenge_method")

	ctx := context.Background()
	if err := s.clientStore.ValidateRedirectURI(ctx, clientID, redirectURI); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid redirect_uri")
	}

	code, err := generateCode()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate code")
	}

	ti := models.NewToken()
	ti.SetClientID(clientID)
	ti.SetUserID(userIDStr)
	ti.SetRedirectURI(redirectURI)
	ti.SetScope(scope)
	ti.SetCode(code)
	ti.SetCodeCreateAt(time.Now())
	ti.SetCodeExpiresIn(s.config.AuthCodeExpiry)

	if err := s.tokenStore.CreateWithPKCE(ctx, ti, codeChallenge, codeChallengeMethod); err != nil {
		c.Logger().Errorf("Failed to store authorization code: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create authorization")
	}

	callbackURL := redirectURI + "?code=" + code
	if state != "" {
		callbackURL += "&state=" + state
	}

	return c.Redirect(http.StatusFound, callbackURL)
}

func (s *Server) TokenHandler(c echo.Context) error {
	grantType := c.FormValue("grant_type")
	ctx := context.Background()

	switch grantType {
	case "authorization_code":
		return s.handleAuthorizationCodeGrant(c, ctx)
	case "refresh_token":
		return s.handleRefreshTokenGrant(c, ctx)
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "unsupported_grant_type",
			"error_description": "only authorization_code and refresh_token are supported",
		})
	}
}

func (s *Server) handleAuthorizationCodeGrant(c echo.Context, ctx context.Context) error {
	code := c.FormValue("code")
	redirectURI := c.FormValue("redirect_uri")
	clientID := c.FormValue("client_id")
	codeVerifier := c.FormValue("code_verifier")

	if code == "" || clientID == "" || codeVerifier == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_request",
			"error_description": "code, client_id, and code_verifier are required",
		})
	}

	codeChallenge, codeChallengeMethod, err := s.tokenStore.GetPKCEByCode(ctx, code)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "invalid or expired authorization code",
		})
	}

	if !ValidatePKCE(codeVerifier, codeChallenge, codeChallengeMethod) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "PKCE validation failed",
		})
	}

	ti, err := s.tokenStore.GetByCode(ctx, code)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "invalid or expired authorization code",
		})
	}

	if ti.GetClientID() != clientID {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "client_id mismatch",
		})
	}
	if redirectURI != "" && ti.GetRedirectURI() != redirectURI {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "redirect_uri mismatch",
		})
	}

	if ti.GetCodeCreateAt().Add(ti.GetCodeExpiresIn()).Before(time.Now()) {
		s.tokenStore.RemoveByCode(ctx, code)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "authorization code has expired",
		})
	}

	accessToken, err := generateToken()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate access token")
	}
	refreshToken, err := generateToken()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate refresh token")
	}

	newTI := models.NewToken()
	newTI.SetClientID(ti.GetClientID())
	newTI.SetUserID(ti.GetUserID())
	newTI.SetRedirectURI(ti.GetRedirectURI())
	newTI.SetScope(ti.GetScope())
	newTI.SetAccess(accessToken)
	newTI.SetAccessCreateAt(time.Now())
	newTI.SetAccessExpiresIn(s.config.AccessTokenExpiry)
	newTI.SetRefresh(refreshToken)
	newTI.SetRefreshCreateAt(time.Now())
	newTI.SetRefreshExpiresIn(s.config.RefreshTokenExpiry)

	if err := s.tokenStore.UpdateCodeToToken(ctx, code, newTI); err != nil {
		c.Logger().Errorf("Failed to update token: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create tokens")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    int(s.config.AccessTokenExpiry.Seconds()),
		"refresh_token": refreshToken,
		"scope":         ti.GetScope(),
	})
}

func (s *Server) handleRefreshTokenGrant(c echo.Context, ctx context.Context) error {
	refreshToken := c.FormValue("refresh_token")
	clientID := c.FormValue("client_id")

	if refreshToken == "" || clientID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_request",
			"error_description": "refresh_token and client_id are required",
		})
	}

	ti, err := s.tokenStore.GetByRefresh(ctx, refreshToken)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "invalid refresh token",
		})
	}

	if ti.GetClientID() != clientID {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "client_id mismatch",
		})
	}

	if ti.GetRefreshCreateAt().Add(ti.GetRefreshExpiresIn()).Before(time.Now()) {
		s.tokenStore.RemoveByRefresh(ctx, refreshToken)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "refresh token has expired",
		})
	}

	newAccessToken, err := generateToken()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate access token")
	}
	newRefreshToken, err := generateToken()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate refresh token")
	}

	if err := s.tokenStore.RemoveByRefresh(ctx, refreshToken); err != nil {
		c.Logger().Errorf("Failed to remove old refresh token: %v", err)
	}

	newTI := models.NewToken()
	newTI.SetCode(uuid.New().String())
	newTI.SetClientID(ti.GetClientID())
	newTI.SetUserID(ti.GetUserID())
	newTI.SetRedirectURI(ti.GetRedirectURI())
	newTI.SetScope(ti.GetScope())
	newTI.SetAccess(newAccessToken)
	newTI.SetAccessCreateAt(time.Now())
	newTI.SetAccessExpiresIn(s.config.AccessTokenExpiry)
	newTI.SetRefresh(newRefreshToken)
	newTI.SetRefreshCreateAt(time.Now())
	newTI.SetRefreshExpiresIn(s.config.RefreshTokenExpiry)

	if err := s.tokenStore.Create(ctx, newTI); err != nil {
		c.Logger().Errorf("Failed to create new token: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create tokens")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"access_token":  newAccessToken,
		"token_type":    "Bearer",
		"expires_in":    int(s.config.AccessTokenExpiry.Seconds()),
		"refresh_token": newRefreshToken,
		"scope":         ti.GetScope(),
	})
}

func (s *Server) RegisterHandler(c echo.Context) error {
	var req RegisterClientRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.ClientName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "client_name is required")
	}
	if len(req.RedirectURIs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "redirect_uris is required")
	}

	resp, err := s.RegisterClient(context.Background(), req)
	if err != nil {
		c.Logger().Errorf("Failed to register client: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to register client")
	}

	return c.JSON(http.StatusCreated, resp)
}

func generateCode() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func GetUserIDFromSession(sess *sessions.Session) (string, bool) {
	userIDStr, ok := sess.Values["user_id"].(string)
	return userIDStr, ok
}
