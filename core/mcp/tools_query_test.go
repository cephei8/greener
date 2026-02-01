package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/cephei8/greener/core"
	model_api "github.com/cephei8/greener/core/model/api"
	model_db "github.com/cephei8/greener/core/model/db"
	"github.com/google/uuid"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func createTestMCPServer(t *testing.T) (*MCPServer, *core.MockQueryServiceInterface) {
	mockService := core.NewMockQueryServiceInterface(t)
	server := NewMCPServerWithQueryService(nil, nil, mockService)
	return server, mockService
}

func createToolRequest(args map[string]interface{}) mcpgo.CallToolRequest {
	return mcpgo.CallToolRequest{
		Params: mcpgo.CallToolParams{
			Arguments: args,
		},
	}
}

func TestHandleQueryTestcases_Success(t *testing.T) {
	server, mockService := createTestMCPServer(t)

	userID := uuid.New()
	ctx := ContextWithUserID(context.Background(), model_db.BinaryUUID(userID))

	expectedResult := &core.QueryResult[model_api.Testcase]{
		Results: []model_api.Testcase{
			{ID: uuid.New().String(), Name: "test1", Status: "pass", CreatedAt: "2024-01-01 12:00:00"},
			{ID: uuid.New().String(), Name: "test2", Status: "fail", CreatedAt: "2024-01-01 12:00:00"},
		},
		TotalCount: 2,
	}

	mockService.EXPECT().
		QueryTestcases(mock.Anything, model_db.BinaryUUID(userID), core.QueryParams{
			Query:  `status="pass"`,
			Offset: 0,
			Limit:  10,
		}).
		Return(expectedResult, nil)

	request := createToolRequest(map[string]interface{}{
		"query":  `status="pass"`,
		"limit":  float64(10),
		"offset": float64(0),
	})

	result, err := server.handleQueryTestcases(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	textContent, ok := result.Content[0].(mcpgo.TextContent)
	require.True(t, ok)

	var response QueryResponse
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err)
	assert.Equal(t, 2, response.TotalCount)
	assert.Equal(t, `status="pass"`, response.Query)
}

func TestHandleQueryTestcases_NoUserContext(t *testing.T) {
	server, _ := createTestMCPServer(t)

	ctx := context.Background()

	request := createToolRequest(map[string]interface{}{
		"query": "",
	})

	result, err := server.handleQueryTestcases(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestHandleQueryTestcases_ServiceError(t *testing.T) {
	server, mockService := createTestMCPServer(t)

	userID := uuid.New()
	ctx := ContextWithUserID(context.Background(), model_db.BinaryUUID(userID))

	mockService.EXPECT().
		QueryTestcases(mock.Anything, model_db.BinaryUUID(userID), mock.Anything).
		Return(nil, errors.New("query failed"))

	request := createToolRequest(map[string]interface{}{
		"query": "invalid",
	})

	result, err := server.handleQueryTestcases(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestHandleQuerySessions_Success(t *testing.T) {
	server, mockService := createTestMCPServer(t)

	userID := uuid.New()
	ctx := ContextWithUserID(context.Background(), model_db.BinaryUUID(userID))

	expectedResult := &core.QueryResult[model_api.Session]{
		Results: []model_api.Session{
			{ID: uuid.New().String(), Description: "session1", Status: "pass"},
		},
		TotalCount: 1,
	}

	mockService.EXPECT().
		QuerySessions(mock.Anything, model_db.BinaryUUID(userID), core.QueryParams{
			Query:  "",
			Offset: 0,
			Limit:  0,
		}).
		Return(expectedResult, nil)

	request := createToolRequest(map[string]interface{}{
		"query": "",
	})

	result, err := server.handleQuerySessions(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
}

func TestHandleQuerySessions_ServiceError(t *testing.T) {
	server, mockService := createTestMCPServer(t)

	userID := uuid.New()
	ctx := ContextWithUserID(context.Background(), model_db.BinaryUUID(userID))

	mockService.EXPECT().
		QuerySessions(mock.Anything, model_db.BinaryUUID(userID), mock.Anything).
		Return(nil, errors.New("query failed"))

	request := createToolRequest(map[string]interface{}{
		"query": "",
	})

	result, err := server.handleQuerySessions(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestHandleQueryGroups_Success(t *testing.T) {
	server, mockService := createTestMCPServer(t)

	userID := uuid.New()
	ctx := ContextWithUserID(context.Background(), model_db.BinaryUUID(userID))

	expectedResult := &core.QueryResult[model_api.Group]{
		Results: []model_api.Group{
			{Group: "production", Status: "pass"},
			{Group: "staging", Status: "fail"},
		},
		TotalCount: 2,
	}

	query := `group_by(#"env")`
	mockService.EXPECT().
		QueryGroups(mock.Anything, model_db.BinaryUUID(userID), core.QueryParams{
			Query:  query,
			Offset: 0,
			Limit:  0,
		}).
		Return(expectedResult, nil)

	request := createToolRequest(map[string]interface{}{
		"query": query,
	})

	result, err := server.handleQueryGroups(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
}

func TestHandleQueryGroups_EmptyQuery(t *testing.T) {
	server, _ := createTestMCPServer(t)

	userID := uuid.New()
	ctx := ContextWithUserID(context.Background(), model_db.BinaryUUID(userID))

	request := createToolRequest(map[string]interface{}{
		"query": "",
	})

	result, err := server.handleQueryGroups(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestHandleGetTestcase_Success(t *testing.T) {
	server, mockService := createTestMCPServer(t)

	userID := uuid.New()
	testcaseID := uuid.New()
	ctx := ContextWithUserID(context.Background(), model_db.BinaryUUID(userID))

	expectedResult := &core.TestcaseDetail{
		ID:        testcaseID.String(),
		SessionID: uuid.New().String(),
		Name:      "test_login",
		Status:    "pass",
		Classname: "TestAuth",
		CreatedAt: "2024-01-01 12:00:00",
	}

	mockService.EXPECT().
		GetTestcase(mock.Anything, model_db.BinaryUUID(userID), testcaseID).
		Return(expectedResult, nil)

	request := createToolRequest(map[string]interface{}{
		"id": testcaseID.String(),
	})

	result, err := server.handleGetTestcase(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	textContent, ok := result.Content[0].(mcpgo.TextContent)
	require.True(t, ok)

	var response core.TestcaseDetail
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err)
	assert.Equal(t, testcaseID.String(), response.ID)
	assert.Equal(t, "test_login", response.Name)
}

func TestHandleGetTestcase_InvalidID(t *testing.T) {
	server, _ := createTestMCPServer(t)

	userID := uuid.New()
	ctx := ContextWithUserID(context.Background(), model_db.BinaryUUID(userID))

	request := createToolRequest(map[string]interface{}{
		"id": "invalid-uuid",
	})

	result, err := server.handleGetTestcase(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestHandleGetTestcase_NotFound(t *testing.T) {
	server, mockService := createTestMCPServer(t)

	userID := uuid.New()
	testcaseID := uuid.New()
	ctx := ContextWithUserID(context.Background(), model_db.BinaryUUID(userID))

	mockService.EXPECT().
		GetTestcase(mock.Anything, model_db.BinaryUUID(userID), testcaseID).
		Return(nil, errors.New("not found"))

	request := createToolRequest(map[string]interface{}{
		"id": testcaseID.String(),
	})

	result, err := server.handleGetTestcase(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestHandleGetSession_Success(t *testing.T) {
	server, mockService := createTestMCPServer(t)

	userID := uuid.New()
	sessionID := uuid.New()
	ctx := ContextWithUserID(context.Background(), model_db.BinaryUUID(userID))

	expectedResult := &core.SessionDetail{
		ID:          sessionID.String(),
		Description: "Test session",
		Status:      "pass",
		Labels:      map[string]string{"env": "production"},
		CreatedAt:   "2024-01-01 12:00:00",
	}

	mockService.EXPECT().
		GetSession(mock.Anything, model_db.BinaryUUID(userID), sessionID).
		Return(expectedResult, nil)

	request := createToolRequest(map[string]interface{}{
		"id": sessionID.String(),
	})

	result, err := server.handleGetSession(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	textContent, ok := result.Content[0].(mcpgo.TextContent)
	require.True(t, ok)

	var response core.SessionDetail
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err)
	assert.Equal(t, sessionID.String(), response.ID)
	assert.Equal(t, "Test session", response.Description)
}

func TestHandleGetSession_NotFound(t *testing.T) {
	server, mockService := createTestMCPServer(t)

	userID := uuid.New()
	sessionID := uuid.New()
	ctx := ContextWithUserID(context.Background(), model_db.BinaryUUID(userID))

	mockService.EXPECT().
		GetSession(mock.Anything, model_db.BinaryUUID(userID), sessionID).
		Return(nil, errors.New("not found"))

	request := createToolRequest(map[string]interface{}{
		"id": sessionID.String(),
	})

	result, err := server.handleGetSession(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestHandleGetTestcase_MissingID(t *testing.T) {
	server, _ := createTestMCPServer(t)

	userID := uuid.New()
	ctx := ContextWithUserID(context.Background(), model_db.BinaryUUID(userID))

	request := createToolRequest(map[string]interface{}{})

	result, err := server.handleGetTestcase(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestHandleGetSession_MissingID(t *testing.T) {
	server, _ := createTestMCPServer(t)

	userID := uuid.New()
	ctx := ContextWithUserID(context.Background(), model_db.BinaryUUID(userID))

	request := createToolRequest(map[string]interface{}{})

	result, err := server.handleGetSession(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}
