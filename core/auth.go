package core

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/cephei8/greener/core/model/db"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/pbkdf2"
)

type APIKeyData struct {
	APIKeyID     string `json:"apiKeyId"`
	APIKeySecret string `json:"apiKeySecret"`
}

const (
	contextKeyAPIKey = "apikey"
	contextKeyUserID = "user_id"
)

func HashSecret(secret string, salt []byte) []byte {
	return pbkdf2.Key([]byte(secret), salt, 100000, 32, sha256.New)
}

func APIKeyAuth(db *bun.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			apiKeyHeader := c.Request().Header.Get("x-api-key")
			if apiKeyHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing X-API-Key header")
			}

			decodedData, err := base64.StdEncoding.DecodeString(apiKeyHeader)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API key format")
			}

			var keyData APIKeyData
			if err := json.Unmarshal(decodedData, &keyData); err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API key format")
			}

			apiKeyID, err := uuid.Parse(keyData.APIKeyID)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API key format")
			}

			var apiKey model_db.APIKey
			err = db.NewSelect().
				Model(&apiKey).
				Where("id = ?", model_db.BinaryUUID(apiKeyID)).
				Scan(context.Background())
			if err != nil {
				c.Logger().Errorf("Failed to find API key %s: %v", apiKeyID, err)
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API key")
			}

			c.Logger().Debugf("Found API key: %s", apiKey.ID)

			secretHash := HashSecret(keyData.APIKeySecret, apiKey.SecretSalt)
			if subtle.ConstantTimeCompare(secretHash, apiKey.SecretHash) != 1 {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API key")
			}

			c.Set(contextKeyAPIKey, &apiKey)
			c.Set(contextKeyUserID, apiKey.UserID)

			return next(c)
		}
	}
}

func GetAPIKey(c echo.Context) *model_db.APIKey {
	if apiKey, ok := c.Get(contextKeyAPIKey).(*model_db.APIKey); ok {
		return apiKey
	}
	return nil
}

func GetUserId(c echo.Context) model_db.BinaryUUID {
	if userID, ok := c.Get(contextKeyUserID).(model_db.BinaryUUID); ok {
		return userID
	}
	return model_db.BinaryUUID(uuid.Nil)
}
