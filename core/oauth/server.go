package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/uptrace/bun"
)

type Config struct {
	Issuer               string
	AuthCodeExpiry       time.Duration
	AccessTokenExpiry    time.Duration
	RefreshTokenExpiry   time.Duration
	AllowedResponseTypes []oauth2.ResponseType
	AllowedGrantTypes    []oauth2.GrantType
}

func DefaultConfig(issuer string) Config {
	return Config{
		Issuer:             issuer,
		AuthCodeExpiry:     10 * time.Minute,
		AccessTokenExpiry:  1 * time.Hour,
		RefreshTokenExpiry: 30 * 24 * time.Hour,
		AllowedResponseTypes: []oauth2.ResponseType{
			oauth2.Code,
		},
		AllowedGrantTypes: []oauth2.GrantType{
			oauth2.AuthorizationCode,
			oauth2.Refreshing,
		},
	}
}

type Server struct {
	manager     *manage.Manager
	srv         *server.Server
	clientStore *BunClientStore
	tokenStore  *BunTokenStore
	config      Config
	db          *bun.DB
}

func NewServer(db *bun.DB, issuer string) *Server {
	cfg := DefaultConfig(issuer)

	manager := manage.NewDefaultManager()

	manager.SetAuthorizeCodeTokenCfg(&manage.Config{
		AccessTokenExp:    cfg.AccessTokenExpiry,
		RefreshTokenExp:   cfg.RefreshTokenExpiry,
		IsGenerateRefresh: true,
	})

	clientStore := NewBunClientStore(db)
	tokenStore := NewBunTokenStore(db)

	manager.MapClientStorage(clientStore)
	manager.MapTokenStorage(tokenStore)

	manager.MapAccessGenerate(generates.NewAccessGenerate())

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(false)
	srv.SetClientInfoHandler(server.ClientFormHandler)

	srv.Config.AllowedResponseTypes = cfg.AllowedResponseTypes
	srv.Config.AllowedGrantTypes = cfg.AllowedGrantTypes

	return &Server{
		manager:     manager,
		srv:         srv,
		clientStore: clientStore,
		tokenStore:  tokenStore,
		config:      cfg,
		db:          db,
	}
}

func (s *Server) GetManager() *manage.Manager {
	return s.manager
}

func (s *Server) GetServer() *server.Server {
	return s.srv
}

func (s *Server) GetClientStore() *BunClientStore {
	return s.clientStore
}

func (s *Server) GetTokenStore() *BunTokenStore {
	return s.tokenStore
}

func (s *Server) GetConfig() Config {
	return s.config
}

func (s *Server) Metadata() map[string]any {
	return map[string]any{
		"issuer":                                s.config.Issuer,
		"authorization_endpoint":                s.config.Issuer + "/oauth/authorize",
		"token_endpoint":                        s.config.Issuer + "/oauth/token",
		"registration_endpoint":                 s.config.Issuer + "/oauth/register",
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
		"code_challenge_methods_supported":      []string{"S256"},
		"token_endpoint_auth_methods_supported": []string{"none"}, // Public clients with PKCE
		"scopes_supported":                      []string{"read:testcases", "read:sessions"},
	}
}

func GenerateClientID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

type RegisterClientRequest struct {
	ClientName   string   `json:"client_name"`
	RedirectURIs []string `json:"redirect_uris"`
}

type RegisterClientResponse struct {
	ClientID     string   `json:"client_id"`
	ClientName   string   `json:"client_name"`
	RedirectURIs []string `json:"redirect_uris"`
}

func (s *Server) RegisterClient(ctx context.Context, req RegisterClientRequest) (*RegisterClientResponse, error) {
	clientID, err := GenerateClientID()
	if err != nil {
		return nil, err
	}

	redirectURIsJSON, err := json.Marshal(req.RedirectURIs)
	if err != nil {
		return nil, err
	}

	client := &OAuthClient{
		ID:           clientID,
		Name:         req.ClientName,
		RedirectURIs: string(redirectURIsJSON),
		CreatedAt:    time.Now(),
	}

	if err := s.clientStore.Create(ctx, client); err != nil {
		return nil, err
	}

	return &RegisterClientResponse{
		ClientID:     clientID,
		ClientName:   req.ClientName,
		RedirectURIs: req.RedirectURIs,
	}, nil
}
