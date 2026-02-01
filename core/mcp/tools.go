package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
)

const queryLanguageDoc = `
GREENER QUERY LANGUAGE SYNTAX:

IMPORTANT: All string values MUST be enclosed in double quotes.

FIELD FILTERS:
- status = "pass"              Filter by status (values: "pass", "fail", "error", "skip")
- status != "fail"             Negated status filter
- name = "test_login"          Filter by test name (exact match)
- classname = "TestAuth"       Filter by class name
- testsuite = "api_tests"      Filter by test suite name
- file = "tests/test_api.py"   Filter by file path
- session_id = "uuid-here"     Filter by session UUID
- id = "uuid-here"             Filter by testcase UUID

TAG FILTERS (use # prefix, tag names must be quoted):
- #"os"                        Tests that have the "os" tag (any value)
- #"os" = "linux"              Tests where tag "os" equals "linux"
- #"os" != "windows"           Tests where tag "os" does not equal "windows"
- !#"os"                       Tests that do NOT have the "os" tag

LOGICAL OPERATORS:
- and                          Combine conditions with AND
- or                           Combine conditions with OR

DATE FILTERS (format: "YYYY/MM/DD HH:MM:SS"):
- start_date = "2025/01/01 00:00:00"   Filter from this date
- end_date = "2025/12/31 23:59:59"     Filter until this date

PAGINATION:
- offset=10                    Skip first N results
- limit=50                     Return at most N results

EXAMPLES:
- status = "fail"
- status = "fail" and #"os" = "linux"
- name = "test_login" or name = "test_logout"
- status = "error" and classname = "TestAPI" start_date = "2025/01/01 00:00:00"
- #"browser" = "chrome" and status = "fail" limit=10
- status = "pass" and !#"flaky"
`

const groupQueryDoc = `
GROUPING (for query_groups tool):

GROUP BY CLAUSE (required for query_groups):
- group_by(session_id)              Group by session
- group_by(#"tag_name")             Group by tag value (tag name must be quoted)
- group_by(session_id, #"env")      Group by multiple fields

GROUP SELECTOR (optional, selects specific group):
- group = ("value1")                Select specific group
- group = ("session-uuid", "prod")  Select group when grouping by multiple fields

EXAMPLES:
- group_by(session_id)
- status = "fail" group_by(session_id)
- group_by(#"os") group = ("linux")
- status = "pass" group_by(session_id, #"env") group = ("uuid", "prod")
`

func (s *MCPServer) RegisterTools() {
	s.server.AddTool(
		mcp.NewTool("query_testcases",
			mcp.WithDescription("Query test cases using the Greener query language."+queryLanguageDoc),
			mcp.WithString("query",
				mcp.Description(`Query string in Greener query language. All string values must be quoted. Examples:
- Empty query returns all testcases
- status = "fail"
- status = "fail" and #"os" = "linux"
- name = "test_login" and status = "pass"
- #"browser" = "chrome" start_date = "2025/01/01 00:00:00"`),
			),
			mcp.WithNumber("offset",
				mcp.Description("Number of results to skip (default: 0)"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of results to return (default: 100, max: 100)"),
			),
			mcp.WithBoolean("trigger_sse",
				mcp.Description("Whether to trigger browser SSE update (default: true)"),
			),
		),
		s.handleQueryTestcases,
	)

	s.server.AddTool(
		mcp.NewTool("query_sessions",
			mcp.WithDescription("Query test sessions using the Greener query language."+queryLanguageDoc),
			mcp.WithString("query",
				mcp.Description(`Query string in Greener query language. All string values must be quoted. Examples:
- Empty query returns all sessions
- status = "fail"
- name = "nightly_run" and status = "pass"
- start_date = "2025/01/01 00:00:00" end_date = "2025/01/31 23:59:59"`),
			),
			mcp.WithNumber("offset",
				mcp.Description("Number of results to skip (default: 0)"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of results to return (default: 100, max: 100)"),
			),
			mcp.WithBoolean("trigger_sse",
				mcp.Description("Whether to trigger browser SSE update (default: true)"),
			),
		),
		s.handleQuerySessions,
	)

	s.server.AddTool(
		mcp.NewTool("query_groups",
			mcp.WithDescription("Query grouped test results using the Greener query language. Requires a group_by clause."+queryLanguageDoc+groupQueryDoc),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description(`Query string with group_by clause. Tag names in group_by must be quoted. Examples:
- group_by(session_id)
- status = "fail" group_by(session_id)
- group_by(#"os")
- status = "pass" group_by(#"env") group = ("production")`),
			),
			mcp.WithNumber("offset",
				mcp.Description("Number of results to skip (default: 0)"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of results to return (default: 100, max: 100)"),
			),
			mcp.WithBoolean("trigger_sse",
				mcp.Description("Whether to trigger browser SSE update (default: true)"),
			),
		),
		s.handleQueryGroups,
	)

	s.server.AddTool(
		mcp.NewTool("get_testcase",
			mcp.WithDescription("Get detailed information about a specific test case including full output, error messages, and metadata."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("UUID of the test case (e.g., \"550e8400-e29b-41d4-a716-446655440000\")"),
			),
			mcp.WithBoolean("trigger_sse",
				mcp.Description("Whether to trigger browser SSE update (default: true)"),
			),
		),
		s.handleGetTestcase,
	)

	s.server.AddTool(
		mcp.NewTool("get_session",
			mcp.WithDescription("Get detailed information about a specific test session including summary statistics and metadata."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("UUID of the session (e.g., \"550e8400-e29b-41d4-a716-446655440000\")"),
			),
			mcp.WithBoolean("trigger_sse",
				mcp.Description("Whether to trigger browser SSE update (default: true)"),
			),
		),
		s.handleGetSession,
	)
}
