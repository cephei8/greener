# MCP Server

Greener includes an MCP (Model Context Protocol) server that allows AI agents to query test results.

The MCP server is available at `/api/v1/mcp` and uses OAuth 2.0 for authentication.

## How It Works

When an AI agent queries test results through the MCP server, you can view the results in two ways:

1. **In the agent interface** - The agent receives the query results directly and can analyze, summarize, or discuss them with you
2. **In the Greener UI** - Queries can trigger a server-sent event (SSE) that updates the Greener browser tab to display the same results

This means you can ask your AI agent questions like "show me all failing tests from today" and see the results both in your conversation and in the Greener web interface simultaneously.

To disable the browser update, agents can set `trigger_sse: false` when calling MCP tools. Just tell the agent how you prefer to see the results.

## Available Tools

| Tool              | Description                                           |
|:------------------|:------------------------------------------------------|
| query_testcases   | Query test cases using the Greener query language     |
| query_sessions    | Query test sessions using the Greener query language  |
| query_groups      | Query grouped test results with group_by clause       |
| get_testcase      | Get detailed information about a specific test case   |
| get_session       | Get detailed information about a specific test session|

## OAuth Endpoints

| Endpoint                                  | Description                       |
|:------------------------------------------|:----------------------------------|
| `/.well-known/oauth-authorization-server` | OAuth server metadata             |
| `/oauth/authorize`                        | Authorization endpoint            |
| `/oauth/token`                            | Token endpoint                    |
| `/oauth/register`                         | Dynamic client registration       |

## Configuration

Set `GREENER_AUTH_ISSUER` environment variable to your external base URL if it differs from localhost.
