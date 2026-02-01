package core_test

import (
	"context"
	"testing"

	"github.com/cephei8/greener/core"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *BaseSuite) TestQueryServiceQueryTestcases() {
	tests := []struct {
		name           string
		params         core.QueryParams
		expectedIds    []uuid.UUID
		expectedCount  int
		expectedStatus []string
		expectErr      bool
		errContains    string
	}{
		{
			name:           "empty query - returns all testcases",
			params:         core.QueryParams{},
			expectedIds:    []uuid.UUID{s.testcase6Id, s.testcase5Id, s.testcase4Id, s.testcase3Id, s.testcase2Id, s.testcase1Id},
			expectedCount:  6,
			expectedStatus: []string{"pass", "pass", "error", "pass", "fail", "pass"},
		},
		{
			name:           "filter by status pass",
			params:         core.QueryParams{Query: `status = "pass"`},
			expectedIds:    []uuid.UUID{s.testcase6Id, s.testcase5Id, s.testcase3Id, s.testcase1Id},
			expectedCount:  4,
			expectedStatus: []string{"pass", "pass", "pass", "pass"},
		},
		{
			name:           "filter by status fail",
			params:         core.QueryParams{Query: `status = "fail"`},
			expectedIds:    []uuid.UUID{s.testcase2Id},
			expectedCount:  1,
			expectedStatus: []string{"fail"},
		},
		{
			name:           "filter by status error",
			params:         core.QueryParams{Query: `status = "error"`},
			expectedIds:    []uuid.UUID{s.testcase4Id},
			expectedCount:  1,
			expectedStatus: []string{"error"},
		},
		{
			name:           "filter by name",
			params:         core.QueryParams{Query: `name = "test_login_success"`},
			expectedIds:    []uuid.UUID{s.testcase1Id},
			expectedCount:  1,
			expectedStatus: []string{"pass"},
		},
		{
			name:           "filter by classname",
			params:         core.QueryParams{Query: `classname = "TestAuth"`},
			expectedIds:    []uuid.UUID{s.testcase2Id, s.testcase1Id},
			expectedCount:  2,
			expectedStatus: []string{"fail", "pass"},
		},
		{
			name:           "filter by file",
			params:         core.QueryParams{Query: `file = "test_auth.py"`},
			expectedIds:    []uuid.UUID{s.testcase2Id, s.testcase1Id},
			expectedCount:  2,
			expectedStatus: []string{"fail", "pass"},
		},
		{
			name:           "filter by testsuite",
			params:         core.QueryParams{Query: `testsuite = "api_tests"`},
			expectedIds:    []uuid.UUID{s.testcase6Id, s.testcase5Id, s.testcase4Id, s.testcase3Id},
			expectedCount:  4,
			expectedStatus: []string{"pass", "pass", "error", "pass"},
		},
		{
			name:           "filter by tag value",
			params:         core.QueryParams{Query: `#"env" = "production"`},
			expectedIds:    []uuid.UUID{s.testcase6Id, s.testcase5Id, s.testcase2Id, s.testcase1Id},
			expectedCount:  4,
			expectedStatus: []string{"pass", "pass", "fail", "pass"},
		},
		{
			name:           "filter by tag value - staging",
			params:         core.QueryParams{Query: `#"env" = "staging"`},
			expectedIds:    []uuid.UUID{s.testcase4Id, s.testcase3Id},
			expectedCount:  2,
			expectedStatus: []string{"error", "pass"},
		},
		{
			name:           "compound query - AND",
			params:         core.QueryParams{Query: `status = "pass" AND classname = "TestAuth"`},
			expectedIds:    []uuid.UUID{s.testcase1Id},
			expectedCount:  1,
			expectedStatus: []string{"pass"},
		},
		{
			name:           "compound query - OR",
			params:         core.QueryParams{Query: `status = "fail" OR status = "error"`},
			expectedIds:    []uuid.UUID{s.testcase4Id, s.testcase2Id},
			expectedCount:  2,
			expectedStatus: []string{"error", "fail"},
		},
		{
			name:          "with limit",
			params:        core.QueryParams{Limit: 2},
			expectedIds:   []uuid.UUID{s.testcase6Id, s.testcase5Id},
			expectedCount: 6,
		},
		{
			name:          "with offset",
			params:        core.QueryParams{Offset: 2},
			expectedIds:   []uuid.UUID{s.testcase4Id, s.testcase3Id, s.testcase2Id, s.testcase1Id},
			expectedCount: 6,
		},
		{
			name:          "with offset and limit",
			params:        core.QueryParams{Offset: 2, Limit: 2},
			expectedIds:   []uuid.UUID{s.testcase4Id, s.testcase3Id},
			expectedCount: 6,
		},
		{
			name:        "invalid query syntax",
			params:      core.QueryParams{Query: "status ="},
			expectErr:   true,
			errContains: "invalid query",
		},
		{
			name:        "invalid field for testcase query",
			params:      core.QueryParams{Query: "description = foo"},
			expectErr:   true,
			errContains: "invalid query",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			svc := core.NewQueryService(s.db)

			result, err := svc.QueryTestcases(ctx, s.userID, tt.params)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tt.expectedCount, result.TotalCount)
			assert.Equal(t, len(tt.expectedIds), len(result.Results))

			actualIds := make([]uuid.UUID, len(result.Results))
			for i, tc := range result.Results {
				id, err := uuid.Parse(tc.ID)
				require.NoError(t, err)
				actualIds[i] = id
			}

			assert.Equal(t, tt.expectedIds, actualIds)

			if tt.expectedStatus != nil {
				for i, tc := range result.Results {
					assert.Equal(t, tt.expectedStatus[i], tc.Status)
				}
			}

			for _, tc := range result.Results {
				assert.NotEmpty(t, tc.ID)
				assert.NotEmpty(t, tc.SessionID)
				assert.NotEmpty(t, tc.Name)
				assert.NotEmpty(t, tc.Status)
				assert.NotEmpty(t, tc.CreatedAt)
			}
		})
	}
}

func (s *BaseSuite) TestQueryServiceQuerySessions() {
	tests := []struct {
		name           string
		params         core.QueryParams
		expectedIds    []uuid.UUID
		expectedCount  int
		expectedStatus []string
		expectErr      bool
		errContains    string
	}{
		{
			name:           "empty query - returns all sessions",
			params:         core.QueryParams{},
			expectedIds:    []uuid.UUID{s.session3Id, s.session2Id, s.session1Id},
			expectedCount:  3,
			expectedStatus: []string{"pass", "error", "fail"},
		},
		{
			name:           "filter by tag value - production",
			params:         core.QueryParams{Query: `#"env" = "production"`},
			expectedIds:    []uuid.UUID{s.session3Id, s.session1Id},
			expectedCount:  2,
			expectedStatus: []string{"pass", "fail"},
		},
		{
			name:           "filter by tag value - staging",
			params:         core.QueryParams{Query: `#"env" = "staging"`},
			expectedIds:    []uuid.UUID{s.session2Id},
			expectedCount:  1,
			expectedStatus: []string{"error"},
		},
		{
			name:           "filter by tag - branch",
			params:         core.QueryParams{Query: `#"branch"`},
			expectedIds:    []uuid.UUID{s.session3Id, s.session2Id, s.session1Id},
			expectedCount:  3,
			expectedStatus: []string{"pass", "error", "fail"},
		},
		{
			name:           "filter by tag value - branch main",
			params:         core.QueryParams{Query: `#"branch" = "main"`},
			expectedIds:    []uuid.UUID{s.session3Id, s.session1Id},
			expectedCount:  2,
			expectedStatus: []string{"pass", "fail"},
		},
		{
			name:          "with limit",
			params:        core.QueryParams{Limit: 2},
			expectedIds:   []uuid.UUID{s.session3Id, s.session2Id},
			expectedCount: 3,
		},
		{
			name:          "with offset",
			params:        core.QueryParams{Offset: 1},
			expectedIds:   []uuid.UUID{s.session2Id, s.session1Id},
			expectedCount: 3,
		},
		{
			name:          "with offset and limit",
			params:        core.QueryParams{Offset: 1, Limit: 1},
			expectedIds:   []uuid.UUID{s.session2Id},
			expectedCount: 3,
		},
		{
			name:        "invalid query syntax",
			params:      core.QueryParams{Query: "#env ="},
			expectErr:   true,
			errContains: "invalid query",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			svc := core.NewQueryService(s.db)

			result, err := svc.QuerySessions(ctx, s.userID, tt.params)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tt.expectedCount, result.TotalCount)
			assert.Equal(t, len(tt.expectedIds), len(result.Results))

			actualIds := make([]uuid.UUID, len(result.Results))
			for i, sess := range result.Results {
				id, err := uuid.Parse(sess.ID)
				require.NoError(t, err)
				actualIds[i] = id
			}

			assert.Equal(t, tt.expectedIds, actualIds)

			if tt.expectedStatus != nil {
				for i, sess := range result.Results {
					assert.Equal(t, tt.expectedStatus[i], sess.Status)
				}
			}

			for _, sess := range result.Results {
				assert.NotEmpty(t, sess.ID)
				assert.NotEmpty(t, sess.Status)
				assert.NotEmpty(t, sess.CreatedAt)
			}
		})
	}
}

func (s *BaseSuite) TestQueryServiceQueryGroups() {
	tests := []struct {
		name           string
		params         core.QueryParams
		expectedGroups []string
		expectedCount  int
		expectErr      bool
		errContains    string
	}{
		{
			name:           "group by session_id",
			params:         core.QueryParams{Query: "group_by(session_id)"},
			expectedGroups: []string{s.session1Id.String(), s.session2Id.String(), s.session3Id.String()},
			expectedCount:  3,
		},
		{
			name:           "group by env tag",
			params:         core.QueryParams{Query: `group_by(#"env")`},
			expectedGroups: []string{"production", "staging"},
			expectedCount:  2,
		},
		{
			name:           "group by branch tag",
			params:         core.QueryParams{Query: `group_by(#"branch")`},
			expectedGroups: []string{"main", "develop"},
			expectedCount:  2,
		},
		{
			name:           "group by env with status filter",
			params:         core.QueryParams{Query: `status = "pass" group_by(#"env")`},
			expectedGroups: []string{"production", "staging"},
			expectedCount:  2,
		},
		{
			name:        "empty query - error",
			params:      core.QueryParams{},
			expectErr:   true,
			errContains: "query is required",
		},
		{
			name:        "query without group_by returns error",
			params:      core.QueryParams{Query: `status = "pass"`},
			expectErr:   true,
			errContains: "group_by is required",
		},
		{
			name:        "invalid query syntax",
			params:      core.QueryParams{Query: "group_by("},
			expectErr:   true,
			errContains: "invalid query",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			svc := core.NewQueryService(s.db)

			result, err := svc.QueryGroups(ctx, s.userID, tt.params)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tt.expectedCount, result.TotalCount)
			assert.Equal(t, len(tt.expectedGroups), len(result.Results))

			if len(tt.expectedGroups) > 0 {
				actualGroups := make([]string, len(result.Results))
				for i, g := range result.Results {
					actualGroups[i] = g.Group
				}

				for _, expected := range tt.expectedGroups {
					assert.Contains(t, actualGroups, expected, "expected group not found: %s", expected)
				}
			}

			for _, g := range result.Results {
				assert.NotEmpty(t, g.Group)
				assert.NotEmpty(t, g.Status)
			}
		})
	}
}

func (s *BaseSuite) TestQueryServiceGetTestcase() {
	tests := []struct {
		name        string
		testcaseID  uuid.UUID
		expectErr   bool
		errContains string
		validate    func(t *testing.T, result *core.TestcaseDetail)
	}{
		{
			name:       "get existing testcase - testcase1",
			testcaseID: s.testcase1Id,
			validate: func(t *testing.T, result *core.TestcaseDetail) {
				assert.Equal(t, s.testcase1Id.String(), result.ID)
				assert.Equal(t, s.session1Id.String(), result.SessionID)
				assert.Equal(t, "test_login_success", result.Name)
				assert.Equal(t, "pass", result.Status)
				assert.Equal(t, "TestAuth", result.Classname)
				assert.Equal(t, "test_auth.py", result.File)
				assert.Equal(t, "auth_tests", result.Testsuite)
				assert.NotEmpty(t, result.CreatedAt)
			},
		},
		{
			name:       "get existing testcase - testcase2 (fail status)",
			testcaseID: s.testcase2Id,
			validate: func(t *testing.T, result *core.TestcaseDetail) {
				assert.Equal(t, s.testcase2Id.String(), result.ID)
				assert.Equal(t, "test_login_failure", result.Name)
				assert.Equal(t, "fail", result.Status)
			},
		},
		{
			name:       "get existing testcase - testcase4 (error status)",
			testcaseID: s.testcase4Id,
			validate: func(t *testing.T, result *core.TestcaseDetail) {
				assert.Equal(t, s.testcase4Id.String(), result.ID)
				assert.Equal(t, "test_api_error", result.Name)
				assert.Equal(t, "error", result.Status)
			},
		},
		{
			name:        "get non-existent testcase",
			testcaseID:  uuid.New(),
			expectErr:   true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			svc := core.NewQueryService(s.db)

			result, err := svc.GetTestcase(ctx, s.userID, tt.testcaseID)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func (s *BaseSuite) TestQueryServiceGetSession() {
	tests := []struct {
		name        string
		sessionID   uuid.UUID
		expectErr   bool
		errContains string
		validate    func(t *testing.T, result *core.SessionDetail)
	}{
		{
			name:      "get existing session - session1",
			sessionID: s.session1Id,
			validate: func(t *testing.T, result *core.SessionDetail) {
				assert.Equal(t, s.session1Id.String(), result.ID)
				assert.Equal(t, "First test session", result.Description)
				assert.Equal(t, "fail", result.Status)
				assert.NotEmpty(t, result.CreatedAt)
				assert.NotNil(t, result.Labels)
				assert.Equal(t, "production", result.Labels["env"])
				assert.Equal(t, "main", result.Labels["branch"])
			},
		},
		{
			name:      "get existing session - session2",
			sessionID: s.session2Id,
			validate: func(t *testing.T, result *core.SessionDetail) {
				assert.Equal(t, s.session2Id.String(), result.ID)
				assert.Equal(t, "Second test session", result.Description)
				assert.Equal(t, "error", result.Status)
				assert.NotNil(t, result.Labels)
				assert.Equal(t, "staging", result.Labels["env"])
				assert.Equal(t, "develop", result.Labels["branch"])
				assert.Equal(t, "linux", result.Labels["platform"])
			},
		},
		{
			name:      "get existing session - session3 (all pass)",
			sessionID: s.session3Id,
			validate: func(t *testing.T, result *core.SessionDetail) {
				assert.Equal(t, s.session3Id.String(), result.ID)
				assert.Equal(t, "Third test session", result.Description)
				assert.Equal(t, "pass", result.Status)
				assert.NotNil(t, result.Labels)
				assert.Equal(t, "production", result.Labels["env"])
				assert.Equal(t, "main", result.Labels["branch"])
			},
		},
		{
			name:        "get non-existent session",
			sessionID:   uuid.New(),
			expectErr:   true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			svc := core.NewQueryService(s.db)

			result, err := svc.GetSession(ctx, s.userID, tt.sessionID)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func (s *BaseSuite) TestQueryServiceQueryTestcasesFilterBySession() {
	ctx := context.Background()
	svc := core.NewQueryService(s.db)

	result, err := svc.QueryTestcases(ctx, s.userID, core.QueryParams{
		Query: `session_id = "` + s.session1Id.String() + `"`,
	})

	require.NoError(s.T(), err)
	require.NotNil(s.T(), result)
	assert.Equal(s.T(), 2, result.TotalCount)
	assert.Equal(s.T(), 2, len(result.Results))

	for _, tc := range result.Results {
		assert.Equal(s.T(), s.session1Id.String(), tc.SessionID)
	}
}

func (s *BaseSuite) TestQueryServiceQueryTestcasesFilterByID() {
	ctx := context.Background()
	svc := core.NewQueryService(s.db)

	result, err := svc.QueryTestcases(ctx, s.userID, core.QueryParams{
		Query: `id = "` + s.testcase1Id.String() + `"`,
	})

	require.NoError(s.T(), err)
	require.NotNil(s.T(), result)
	assert.Equal(s.T(), 1, result.TotalCount)
	assert.Equal(s.T(), 1, len(result.Results))
	assert.Equal(s.T(), s.testcase1Id.String(), result.Results[0].ID)
}

func (s *BaseSuite) TestQueryServiceQuerySessionsFilterBySession() {
	ctx := context.Background()
	svc := core.NewQueryService(s.db)

	result, err := svc.QuerySessions(ctx, s.userID, core.QueryParams{
		Query: `session_id = "` + s.session1Id.String() + `"`,
	})

	require.NoError(s.T(), err)
	require.NotNil(s.T(), result)
	assert.Equal(s.T(), 1, result.TotalCount)
	assert.Equal(s.T(), 1, len(result.Results))
	assert.Equal(s.T(), s.session1Id.String(), result.Results[0].ID)
}

func (s *BaseSuite) TestQueryServiceLimitEnforcement() {
	ctx := context.Background()
	svc := core.NewQueryService(s.db)

	result, err := svc.QueryTestcases(ctx, s.userID, core.QueryParams{
		Limit: 150,
	})

	require.NoError(s.T(), err)
	require.NotNil(s.T(), result)
	assert.Equal(s.T(), 6, len(result.Results))
}

func (s *BaseSuite) TestQueryServiceDateTimeFormat() {
	ctx := context.Background()
	svc := core.NewQueryService(s.db)

	result, err := svc.QueryTestcases(ctx, s.userID, core.QueryParams{Limit: 1})

	require.NoError(s.T(), err)
	require.NotNil(s.T(), result)
	require.NotEmpty(s.T(), result.Results)

	assert.Regexp(s.T(), `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`, result.Results[0].CreatedAt)
}
