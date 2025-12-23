package core

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"time"

	"github.com/cephei8/greener/core/model/db"
	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/pbkdf2"
)

func APIKeysHandler(c echo.Context) error {
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

	var apiKeys []model_db.APIKey
	err = db.NewSelect().
		Model(&apiKeys).
		Where("user_id = ?", model_db.BinaryUUID(userId)).
		OrderExpr("created_at DESC").
		Scan(ctx)
	if err != nil {
		c.Logger().Errorf("Failed to load API keys: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load API keys")
	}

	type APIKeyView struct {
		ID          string
		Description string
		CreatedAt   string
	}

	apiKeyViews := make([]APIKeyView, len(apiKeys))
	for i, key := range apiKeys {
		description := ""
		if key.Description != nil {
			description = *key.Description
		}
		apiKeyViews[i] = APIKeyView{
			ID:          key.ID.String(),
			Description: description,
			CreatedAt:   key.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return c.Render(http.StatusOK, "apikeys.html", map[string]any{
		"APIKeys": apiKeyViews,
	})
}

func CreateAPIKeyHandler(c echo.Context) error {
	sess, _ := session.Get("session", c)
	if auth, ok := sess.Values["authenticated"].(bool); !ok || !auth {
		return c.HTML(http.StatusUnauthorized, `<div class="alert alert-error">Unauthorized</div>`)
	}

	userIdStr, ok := sess.Values["user_id"].(string)
	if !ok {
		sess.Values["authenticated"] = false
		sess.Save(c.Request(), c.Response())
		return c.HTML(http.StatusUnauthorized, `<div class="alert alert-error">Unauthorized</div>`)
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		c.Logger().Errorf("Invalid user_id in session: %v", err)
		sess.Values["authenticated"] = false
		sess.Save(c.Request(), c.Response())
		return c.HTML(http.StatusUnauthorized, `<div class="alert alert-error">Unauthorized</div>`)
	}

	description := c.FormValue("description")

	db := c.Get("db").(*bun.DB)
	ctx := context.Background()

	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		c.Logger().Errorf("Failed to generate secret: %v", err)
		return c.HTML(http.StatusInternalServerError, `<div class="alert alert-error">Failed to generate API key</div>`)
	}
	secretStr := base64.URLEncoding.EncodeToString(secret)

	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		c.Logger().Errorf("Failed to generate salt: %v", err)
		return c.HTML(http.StatusInternalServerError, `<div class="alert alert-error">Failed to generate API key</div>`)
	}
	secretHash := pbkdf2.Key([]byte(secretStr), salt, 100000, 32, sha256.New)

	id := uuid.New()
	apiKey := &model_db.APIKey{
		ID:         model_db.BinaryUUID(id),
		SecretSalt: salt,
		SecretHash: secretHash,
		UserID:     model_db.BinaryUUID(userId),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if description != "" {
		apiKey.Description = &description
	}

	_, err = db.NewInsert().Model(apiKey).Exec(ctx)
	if err != nil {
		c.Logger().Errorf("Failed to insert API key: %v", err)
		return c.HTML(http.StatusInternalServerError, `<div class="alert alert-error">Failed to create API key</div>`)
	}

	keyData := map[string]string{
		"apiKeyId":     id.String(),
		"apiKeySecret": secretStr,
	}
	keyJSON, _ := json.Marshal(keyData)
	key := base64.StdEncoding.EncodeToString(keyJSON)

	descDisplay := description
	if descDisplay == "" {
		descDisplay = "No description"
	}

	html := fmt.Sprintf(`
		<h3 class="font-bold text-lg mb-4">API Key Created</h3>
		<div role="alert" class="alert alert-warning mb-4">
			<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 shrink-0 stroke-current" fill="none" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
			</svg>
			<span><strong>Warning:</strong> The key will not be accessible again, copy it now.</span>
		</div>

		<div class="space-y-3">
			<div>
				<label class="label">
					<span class="label-text font-semibold">ID</span>
				</label>
				<div class="font-mono text-sm bg-base-200 p-2 rounded">%s</div>
			</div>

			<div>
				<label class="label">
					<span class="label-text font-semibold">Description</span>
				</label>
				<div class="text-sm bg-base-200 p-2 rounded">%s</div>
			</div>

			<div>
				<label class="label">
					<span class="label-text font-semibold">Created At</span>
				</label>
				<div class="text-sm bg-base-200 p-2 rounded">%s</div>
			</div>

			<div>
				<label class="label">
					<span class="label-text font-semibold">API Key</span>
				</label>
				<div class="flex gap-2">
					<input
						type="text"
						value="%s"
						class="input input-bordered flex-1 font-mono text-xs"
						readonly />
					<button
						type="button"
						class="btn btn-primary"
						onclick="copyToClipboard('%s')">
						Copy
					</button>
				</div>
			</div>
		</div>

		<div class="modal-action">
			<button
				type="button"
				class="btn btn-primary"
				hx-get="/api-keys"
				hx-target="body"
				hx-push-url="true">
				Done
			</button>
		</div>
	`, id.String(), descDisplay, apiKey.CreatedAt.Format("2006-01-02 15:04:05"), key, key)

	c.Response().Header().Set("Content-Type", "text/html")
	return c.HTML(http.StatusOK, html)
}

func DeleteAPIKeyHandler(c echo.Context) error {
	sess, _ := session.Get("session", c)

	if auth, ok := sess.Values["authenticated"].(bool); !ok || !auth {
		return c.HTML(http.StatusUnauthorized, `<div class="alert alert-error">Unauthorized</div>`)
	}

	userIDStr, ok := sess.Values["user_id"].(string)
	if !ok {
		sess.Values["authenticated"] = false
		sess.Save(c.Request(), c.Response())
		return c.HTML(http.StatusUnauthorized, `<div class="alert alert-error">Unauthorized</div>`)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.Logger().Errorf("Invalid user_id in session: %v", err)
		sess.Values["authenticated"] = false
		sess.Save(c.Request(), c.Response())
		return c.HTML(http.StatusUnauthorized, `<div class="alert alert-error">Unauthorized</div>`)
	}

	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.HTML(http.StatusBadRequest, `<div class="alert alert-error">Invalid ID</div>`)
	}

	db := c.Get("db").(*bun.DB)
	ctx := context.Background()

	_, err = db.NewDelete().
		Model((*model_db.APIKey)(nil)).
		Where("id = ? AND user_id = ?", model_db.BinaryUUID(id), model_db.BinaryUUID(userID)).
		Exec(ctx)
	if err != nil {
		c.Logger().Errorf("Failed to delete API key: %v", err)
		return c.HTML(http.StatusInternalServerError, `<div class="alert alert-error">Failed to delete API key</div>`)
	}

	var apiKeys []model_db.APIKey
	err = db.NewSelect().
		Model(&apiKeys).
		Where("user_id = ?", model_db.BinaryUUID(userID)).
		OrderExpr("created_at DESC").
		Scan(ctx)
	if err != nil {
		c.Logger().Errorf("Failed to load API keys: %v", err)
		return c.HTML(http.StatusInternalServerError, `<div class="alert alert-error">Failed to load API keys</div>`)
	}

	tableHTML := ""
	if len(apiKeys) > 0 {
		tableHTML = `<div class="overflow-x-auto">
			<table class="table table-zebra w-full">
				<thead>
					<tr>
						<th class="w-80">ID</th>
						<th>Description</th>
						<th class="w-48">Created At</th>
						<th class="w-24">Actions</th>
					</tr>
				</thead>
				<tbody>`

		for _, key := range apiKeys {
			description := "No description"
			if key.Description != nil {
				description = html.EscapeString(*key.Description)
			}
			htmlRow := fmt.Sprintf(`
					<tr>
						<td class="font-mono text-xs">%s</td>
						<td>%s</td>
						<td class="text-sm">%s</td>
						<td>
							<button
								class="btn btn-sm btn-error"
								hx-delete="/api-keys/%s"
								hx-target="#apikeys-table"
								hx-confirm="Are you sure you want to delete this API key?">
								Delete
							</button>
						</td>
					</tr>`,
				html.EscapeString(key.ID.String()),
				description,
				html.EscapeString(key.CreatedAt.Format("2006-01-02 15:04:05")),
				html.EscapeString(key.ID.String()),
			)
			tableHTML += htmlRow
		}

		tableHTML += `
				</tbody>
			</table>
		</div>`
	} else {
		tableHTML = `<div class="text-center py-12 text-gray-500">
			<p>No API keys found. Create one to get started.</p>
		</div>`
	}

	c.Response().Header().Set("Content-Type", "text/html")
	return c.HTML(http.StatusOK, tableHTML)
}
