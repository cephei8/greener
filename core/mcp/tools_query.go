package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cephei8/greener/core"
	model_db "github.com/cephei8/greener/core/model/db"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
)

type QueryResponse struct {
	Results    any    `json:"results"`
	TotalCount int    `json:"total_count"`
	Query      string `json:"query"`
}

func getTriggerSSE(request mcp.CallToolRequest) bool {
	val, err := request.RequireBool("trigger_sse")
	if err != nil {
		return true // default to true
	}
	return val
}

func (s *MCPServer) handleQueryTestcases(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID := UserIDFromContext(ctx)
	if userID == model_db.BinaryUUID(uuid.Nil) {
		return mcp.NewToolResultError("unauthorized: no user context"), nil
	}

	queryStr := request.GetString("query", "")
	offset := int(request.GetFloat("offset", 0))
	limit := int(request.GetFloat("limit", 0))

	result, err := s.queryService.QueryTestcases(ctx, userID, core.QueryParams{
		Query:  queryStr,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if s.sseHub != nil && getTriggerSSE(request) {
		s.sseHub.BroadcastMCPQuery(userID.String(), "/testcases", queryStr)
	}

	response := QueryResponse{
		Results:    result.Results,
		TotalCount: result.TotalCount,
		Query:      queryStr,
	}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func (s *MCPServer) handleQuerySessions(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID := UserIDFromContext(ctx)
	if userID == model_db.BinaryUUID(uuid.Nil) {
		return mcp.NewToolResultError("unauthorized: no user context"), nil
	}

	queryStr := request.GetString("query", "")
	offset := int(request.GetFloat("offset", 0))
	limit := int(request.GetFloat("limit", 0))

	result, err := s.queryService.QuerySessions(ctx, userID, core.QueryParams{
		Query:  queryStr,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if s.sseHub != nil && getTriggerSSE(request) {
		s.sseHub.BroadcastMCPQuery(userID.String(), "/sessions", queryStr)
	}

	response := QueryResponse{
		Results:    result.Results,
		TotalCount: result.TotalCount,
		Query:      queryStr,
	}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func (s *MCPServer) handleQueryGroups(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID := UserIDFromContext(ctx)
	if userID == model_db.BinaryUUID(uuid.Nil) {
		return mcp.NewToolResultError("unauthorized: no user context"), nil
	}

	queryStr := request.GetString("query", "")
	if queryStr == "" {
		return mcp.NewToolResultError("query is required for group queries"), nil
	}

	offset := int(request.GetFloat("offset", 0))
	limit := int(request.GetFloat("limit", 0))

	result, err := s.queryService.QueryGroups(ctx, userID, core.QueryParams{
		Query:  queryStr,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if s.sseHub != nil && getTriggerSSE(request) {
		s.sseHub.BroadcastMCPQuery(userID.String(), "/groups", queryStr)
	}

	response := QueryResponse{
		Results:    result.Results,
		TotalCount: result.TotalCount,
		Query:      queryStr,
	}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func (s *MCPServer) handleGetTestcase(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID := UserIDFromContext(ctx)
	if userID == model_db.BinaryUUID(uuid.Nil) {
		return mcp.NewToolResultError("unauthorized: no user context"), nil
	}

	idStr, err := request.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError("id is required"), nil
	}

	testcaseID, err := uuid.Parse(idStr)
	if err != nil {
		return mcp.NewToolResultError("invalid testcase ID format"), nil
	}

	result, err := s.queryService.GetTestcase(ctx, userID, testcaseID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if s.sseHub != nil && getTriggerSSE(request) {
		s.sseHub.BroadcastMCPQuery(userID.String(), "/testcases/"+idStr+"/details", "")
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func (s *MCPServer) handleGetSession(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID := UserIDFromContext(ctx)
	if userID == model_db.BinaryUUID(uuid.Nil) {
		return mcp.NewToolResultError("unauthorized: no user context"), nil
	}

	idStr, err := request.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError("id is required"), nil
	}

	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		return mcp.NewToolResultError("invalid session ID format"), nil
	}

	result, err := s.queryService.GetSession(ctx, userID, sessionID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if s.sseHub != nil && getTriggerSSE(request) {
		s.sseHub.BroadcastMCPQuery(userID.String(), "/sessions/"+idStr+"/details", "")
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}
