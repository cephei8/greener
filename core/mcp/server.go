package mcp

import (
	"net/http"

	"github.com/cephei8/greener/core"
	"github.com/cephei8/greener/core/oauth"
	"github.com/cephei8/greener/core/sse"
	"github.com/labstack/echo/v4"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/uptrace/bun"
)

type MCPServer struct {
	server       *mcpserver.MCPServer
	db           *bun.DB
	sseHub       *sse.Hub
	queryService core.QueryServiceInterface
}

func NewMCPServer(db *bun.DB, sseHub *sse.Hub) *MCPServer {
	return NewMCPServerWithQueryService(db, sseHub, core.NewQueryService(db))
}

func NewMCPServerWithQueryService(db *bun.DB, sseHub *sse.Hub, queryService core.QueryServiceInterface) *MCPServer {
	server := mcpserver.NewMCPServer(
		"Greener",
		"1.0.0",
		mcpserver.WithToolCapabilities(false),
	)

	s := &MCPServer{
		server:       server,
		db:           db,
		sseHub:       sseHub,
		queryService: queryService,
	}

	s.RegisterTools()

	return s
}

func (s *MCPServer) EchoHandler() echo.HandlerFunc {
	httpServer := mcpserver.NewStreamableHTTPServer(s.server)

	return func(c echo.Context) error {
		userID := oauth.GetOAuthUserID(c)
		clientID := oauth.GetOAuthClientID(c)
		scope := oauth.GetOAuthScope(c)

		ctx := c.Request().Context()
		ctx = ContextWithUserID(ctx, userID)
		ctx = ContextWithClientID(ctx, clientID)
		ctx = ContextWithScope(ctx, scope)

		req := c.Request().WithContext(ctx)

		httpServer.ServeHTTP(c.Response().Writer, req)

		return nil
	}
}

func (s *MCPServer) GetServer() *mcpserver.MCPServer {
	return s.server
}

func (s *MCPServer) GetDB() *bun.DB {
	return s.db
}

func (s *MCPServer) GetSSEHub() *sse.Hub {
	return s.sseHub
}

func (s *MCPServer) ServerInfo() map[string]any {
	return map[string]any{
		"name":    "Greener MCP Server",
		"version": "1.0.0",
		"tools": []string{
			"query_testcases",
			"query_sessions",
			"query_groups",
			"get_testcase",
			"get_session",
		},
	}
}

func (s *MCPServer) HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
