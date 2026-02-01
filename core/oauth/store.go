package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	model_db "github.com/cephei8/greener/core/model/db"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type OAuthClient struct {
	bun.BaseModel `bun:"table:oauth_clients"`

	ID           string              `bun:"id,pk"`
	SecretHash   []byte              `bun:"secret_hash"`
	Name         string              `bun:"name,notnull"`
	RedirectURIs string              `bun:"redirect_uris,notnull"`
	UserID       *model_db.BinaryUUID `bun:"user_id"`
	CreatedAt    time.Time           `bun:"created_at,nullzero,notnull"`
}

type OAuthToken struct {
	bun.BaseModel `bun:"table:oauth_tokens"`

	Code                string              `bun:"code,pk"`
	Access              string              `bun:"access"`
	Refresh             string              `bun:"refresh"`
	ClientID            string              `bun:"client_id,notnull"`
	UserID              model_db.BinaryUUID `bun:"user_id,notnull"`
	RedirectURI         string              `bun:"redirect_uri"`
	Scope               string              `bun:"scope"`
	CodeChallenge       string              `bun:"code_challenge"`
	CodeChallengeMethod string              `bun:"code_challenge_method"`
	CodeExpiresAt       *time.Time          `bun:"code_expires_at"`
	AccessExpiresAt     *time.Time          `bun:"access_expires_at"`
	RefreshExpiresAt    *time.Time          `bun:"refresh_expires_at"`
	CreatedAt           time.Time           `bun:"created_at,nullzero,notnull"`
}

type BunClientStore struct {
	db *bun.DB
}

func NewBunClientStore(db *bun.DB) *BunClientStore {
	return &BunClientStore{db: db}
}

func (s *BunClientStore) GetByID(ctx context.Context, id string) (oauth2.ClientInfo, error) {
	var client OAuthClient
	err := s.db.NewSelect().
		Model(&client).
		Where("? = ?", bun.Ident("id"), id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	var redirectURIs []string
	if err := json.Unmarshal([]byte(client.RedirectURIs), &redirectURIs); err != nil {
		return nil, err
	}

	secret := ""

	return &models.Client{
		ID:     client.ID,
		Secret: secret,
		Domain: redirectURIs[0],
		UserID: client.UserID.String(),
	}, nil
}

func (s *BunClientStore) Create(ctx context.Context, client *OAuthClient) error {
	_, err := s.db.NewInsert().Model(client).Exec(ctx)
	return err
}

type BunTokenStore struct {
	db *bun.DB
}

func NewBunTokenStore(db *bun.DB) *BunTokenStore {
	return &BunTokenStore{db: db}
}

func (s *BunTokenStore) Create(ctx context.Context, info oauth2.TokenInfo) error {
	userID, err := uuid.Parse(info.GetUserID())
	if err != nil {
		return err
	}

	token := &OAuthToken{
		Code:        info.GetCode(),
		Access:      info.GetAccess(),
		Refresh:     info.GetRefresh(),
		ClientID:    info.GetClientID(),
		UserID:      model_db.BinaryUUID(userID),
		RedirectURI: info.GetRedirectURI(),
		Scope:       info.GetScope(),
		CreatedAt:   time.Now(),
	}

	if info.GetCode() != "" {
		expiresAt := info.GetCodeCreateAt().Add(info.GetCodeExpiresIn())
		token.CodeExpiresAt = &expiresAt
	}
	if info.GetAccess() != "" {
		expiresAt := info.GetAccessCreateAt().Add(info.GetAccessExpiresIn())
		token.AccessExpiresAt = &expiresAt
	}
	if info.GetRefresh() != "" {
		expiresAt := info.GetRefreshCreateAt().Add(info.GetRefreshExpiresIn())
		token.RefreshExpiresAt = &expiresAt
	}

	_, err = s.db.NewInsert().Model(token).Exec(ctx)
	return err
}

func (s *BunTokenStore) RemoveByCode(ctx context.Context, code string) error {
	_, err := s.db.NewDelete().
		Model((*OAuthToken)(nil)).
		Where("? = ?", bun.Ident("code"), code).
		Exec(ctx)
	return err
}

func (s *BunTokenStore) RemoveByAccess(ctx context.Context, access string) error {
	_, err := s.db.NewDelete().
		Model((*OAuthToken)(nil)).
		Where("? = ?", bun.Ident("access"), access).
		Exec(ctx)
	return err
}

func (s *BunTokenStore) RemoveByRefresh(ctx context.Context, refresh string) error {
	_, err := s.db.NewDelete().
		Model((*OAuthToken)(nil)).
		Where("? = ?", bun.Ident("refresh"), refresh).
		Exec(ctx)
	return err
}

func (s *BunTokenStore) GetByCode(ctx context.Context, code string) (oauth2.TokenInfo, error) {
	var token OAuthToken
	err := s.db.NewSelect().
		Model(&token).
		Where("? = ?", bun.Ident("code"), code).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return s.toTokenInfo(&token), nil
}

func (s *BunTokenStore) GetByAccess(ctx context.Context, access string) (oauth2.TokenInfo, error) {
	var token OAuthToken
	err := s.db.NewSelect().
		Model(&token).
		Where("? = ?", bun.Ident("access"), access).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return s.toTokenInfo(&token), nil
}

func (s *BunTokenStore) GetByRefresh(ctx context.Context, refresh string) (oauth2.TokenInfo, error) {
	var token OAuthToken
	err := s.db.NewSelect().
		Model(&token).
		Where("? = ?", bun.Ident("refresh"), refresh).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return s.toTokenInfo(&token), nil
}

func (s *BunTokenStore) toTokenInfo(token *OAuthToken) oauth2.TokenInfo {
	ti := models.NewToken()
	ti.SetClientID(token.ClientID)
	ti.SetUserID(token.UserID.String())
	ti.SetRedirectURI(token.RedirectURI)
	ti.SetScope(token.Scope)

	if token.Code != "" {
		ti.SetCode(token.Code)
		ti.SetCodeCreateAt(token.CreatedAt)
		if token.CodeExpiresAt != nil {
			ti.SetCodeExpiresIn(token.CodeExpiresAt.Sub(token.CreatedAt))
		}
	}

	if token.Access != "" {
		ti.SetAccess(token.Access)
		ti.SetAccessCreateAt(token.CreatedAt)
		if token.AccessExpiresAt != nil {
			ti.SetAccessExpiresIn(token.AccessExpiresAt.Sub(token.CreatedAt))
		}
	}

	if token.Refresh != "" {
		ti.SetRefresh(token.Refresh)
		ti.SetRefreshCreateAt(token.CreatedAt)
		if token.RefreshExpiresAt != nil {
			ti.SetRefreshExpiresIn(token.RefreshExpiresAt.Sub(token.CreatedAt))
		}
	}

	return ti
}

func (s *BunTokenStore) CreateWithPKCE(ctx context.Context, info oauth2.TokenInfo, codeChallenge, codeChallengeMethod string) error {
	userID, err := uuid.Parse(info.GetUserID())
	if err != nil {
		return err
	}

	token := &OAuthToken{
		Code:                info.GetCode(),
		Access:              info.GetAccess(),
		Refresh:             info.GetRefresh(),
		ClientID:            info.GetClientID(),
		UserID:              model_db.BinaryUUID(userID),
		RedirectURI:         info.GetRedirectURI(),
		Scope:               info.GetScope(),
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		CreatedAt:           time.Now(),
	}

	if info.GetCode() != "" {
		expiresAt := info.GetCodeCreateAt().Add(info.GetCodeExpiresIn())
		token.CodeExpiresAt = &expiresAt
	}
	if info.GetAccess() != "" {
		expiresAt := info.GetAccessCreateAt().Add(info.GetAccessExpiresIn())
		token.AccessExpiresAt = &expiresAt
	}
	if info.GetRefresh() != "" {
		expiresAt := info.GetRefreshCreateAt().Add(info.GetRefreshExpiresIn())
		token.RefreshExpiresAt = &expiresAt
	}

	_, err = s.db.NewInsert().Model(token).Exec(ctx)
	return err
}

func (s *BunTokenStore) GetPKCEByCode(ctx context.Context, code string) (codeChallenge, codeChallengeMethod string, err error) {
	var token OAuthToken
	err = s.db.NewSelect().
		Model(&token).
		Column("code_challenge", "code_challenge_method").
		Where("? = ?", bun.Ident("code"), code).
		Scan(ctx)
	if err != nil {
		return "", "", err
	}
	return token.CodeChallenge, token.CodeChallengeMethod, nil
}

func (s *BunTokenStore) UpdateCodeToToken(ctx context.Context, code string, info oauth2.TokenInfo) error {
	accessExpiresAt := info.GetAccessCreateAt().Add(info.GetAccessExpiresIn())
	refreshExpiresAt := info.GetRefreshCreateAt().Add(info.GetRefreshExpiresIn())

	_, err := s.db.NewUpdate().
		Model((*OAuthToken)(nil)).
		Set("? = ?", bun.Ident("access"), info.GetAccess()).
		Set("? = ?", bun.Ident("refresh"), info.GetRefresh()).
		Set("? = ?", bun.Ident("access_expires_at"), accessExpiresAt).
		Set("? = ?", bun.Ident("refresh_expires_at"), refreshExpiresAt).
		Where("? = ?", bun.Ident("code"), code).
		Exec(ctx)
	return err
}

func (s *BunTokenStore) GetClientByCode(ctx context.Context, code string) (clientID, userID, redirectURI, scope string, err error) {
	var token OAuthToken
	err = s.db.NewSelect().
		Model(&token).
		Column("client_id", "user_id", "redirect_uri", "scope").
		Where("? = ?", bun.Ident("code"), code).
		Scan(ctx)
	if err != nil {
		return "", "", "", "", err
	}
	return token.ClientID, token.UserID.String(), token.RedirectURI, token.Scope, nil
}

func (s *BunClientStore) GetRedirectURIs(ctx context.Context, clientID string) ([]string, error) {
	var client OAuthClient
	err := s.db.NewSelect().
		Model(&client).
		Column("redirect_uris").
		Where("? = ?", bun.Ident("id"), clientID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	var uris []string
	if err := json.Unmarshal([]byte(client.RedirectURIs), &uris); err != nil {
		return nil, err
	}
	return uris, nil
}

func (s *BunClientStore) ValidateRedirectURI(ctx context.Context, clientID, redirectURI string) error {
	uris, err := s.GetRedirectURIs(ctx, clientID)
	if err != nil {
		return err
	}

	for _, uri := range uris {
		if uri == redirectURI {
			return nil
		}
	}
	return errors.New("invalid redirect_uri")
}

func (s *BunClientStore) GetClientName(ctx context.Context, clientID string) (string, error) {
	var client OAuthClient
	err := s.db.NewSelect().
		Model(&client).
		Column("name").
		Where("? = ?", bun.Ident("id"), clientID).
		Scan(ctx)
	if err != nil {
		return "", err
	}
	return client.Name, nil
}
