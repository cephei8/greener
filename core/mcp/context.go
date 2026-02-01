package mcp

import (
	"context"

	model_db "github.com/cephei8/greener/core/model/db"
	"github.com/google/uuid"
)

type contextKey string

const (
	userIDKey   contextKey = "mcp_user_id"
	clientIDKey contextKey = "mcp_client_id"
	scopeKey    contextKey = "mcp_scope"
)

func ContextWithUserID(ctx context.Context, userID model_db.BinaryUUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromContext(ctx context.Context) model_db.BinaryUUID {
	if userID, ok := ctx.Value(userIDKey).(model_db.BinaryUUID); ok {
		return userID
	}
	return model_db.BinaryUUID(uuid.Nil)
}

func ContextWithClientID(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, clientIDKey, clientID)
}

func ClientIDFromContext(ctx context.Context) string {
	if clientID, ok := ctx.Value(clientIDKey).(string); ok {
		return clientID
	}
	return ""
}

func ContextWithScope(ctx context.Context, scope string) context.Context {
	return context.WithValue(ctx, scopeKey, scope)
}

func ScopeFromContext(ctx context.Context) string {
	if scope, ok := ctx.Value(scopeKey).(string); ok {
		return scope
	}
	return ""
}
