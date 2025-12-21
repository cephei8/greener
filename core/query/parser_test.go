package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupSpecification(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		checkFunc func(*testing.T, Query)
	}{
		{
			name:    "group_by with group specification",
			input:   `status = "pass" group_by(session_id) group = ("value1")`,
			wantErr: false,
			checkFunc: func(t *testing.T, q Query) {
				require.NotNil(t, q.GroupQuery)
				assert.Equal(t, []string{"value1"}, q.GroupSelector)
			},
		},
		{
			name:    "group_by with multiple group values",
			input:   `name = "test" group_by(session_id, #"env") group = ("session1", "prod")`,
			wantErr: false,
			checkFunc: func(t *testing.T, q Query) {
				require.NotNil(t, q.GroupQuery)
				assert.Equal(t, []string{"session1", "prod"}, q.GroupSelector)
			},
		},
		{
			name:    "group without group_by should fail",
			input:   `status = "pass" group = ("value1")`,
			wantErr: true,
		},
		{
			name:    "group_by without group (valid)",
			input:   `status = "pass" group_by(session_id)`,
			wantErr: false,
			checkFunc: func(t *testing.T, q Query) {
				require.NotNil(t, q.GroupQuery)
				assert.Nil(t, q.GroupSelector)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			q, err := parser.Parse()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, q)
				}
			}
		})
	}
}

func TestOffsetLimit(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		checkFunc func(*testing.T, Query)
	}{
		{
			name:    "offset only",
			input:   `offset=10`,
			wantErr: false,
			checkFunc: func(t *testing.T, q Query) {
				assert.Equal(t, 10, q.Offset)
				assert.Equal(t, 0, q.Limit)
			},
		},
		{
			name:    "limit only",
			input:   `limit=50`,
			wantErr: false,
			checkFunc: func(t *testing.T, q Query) {
				assert.Equal(t, 0, q.Offset)
				assert.Equal(t, 50, q.Limit)
			},
		},
		{
			name:    "offset then limit",
			input:   `offset=20 limit=30`,
			wantErr: false,
			checkFunc: func(t *testing.T, q Query) {
				assert.Equal(t, 20, q.Offset)
				assert.Equal(t, 30, q.Limit)
			},
		},
		{
			name:    "limit then offset",
			input:   `limit=40 offset=15`,
			wantErr: false,
			checkFunc: func(t *testing.T, q Query) {
				assert.Equal(t, 15, q.Offset)
				assert.Equal(t, 40, q.Limit)
			},
		},
		{
			name:    "query with offset",
			input:   `status = "pass" offset=10`,
			wantErr: false,
			checkFunc: func(t *testing.T, q Query) {
				assert.Equal(t, 10, q.Offset)
				assert.Equal(t, 0, q.Limit)
			},
		},
		{
			name:    "query with limit",
			input:   `status = "pass" limit=25`,
			wantErr: false,
			checkFunc: func(t *testing.T, q Query) {
				assert.Equal(t, 0, q.Offset)
				assert.Equal(t, 25, q.Limit)
			},
		},
		{
			name:    "query with offset and limit",
			input:   `name = "test" offset=5 limit=100`,
			wantErr: false,
			checkFunc: func(t *testing.T, q Query) {
				assert.Equal(t, 5, q.Offset)
				assert.Equal(t, 100, q.Limit)
			},
		},
		{
			name:    "query with limit and offset",
			input:   `name = "test" limit=75 offset=12`,
			wantErr: false,
			checkFunc: func(t *testing.T, q Query) {
				assert.Equal(t, 12, q.Offset)
				assert.Equal(t, 75, q.Limit)
			},
		},
		{
			name:    "group_by with offset and limit",
			input:   `group_by(session_id) offset=3 limit=50`,
			wantErr: false,
			checkFunc: func(t *testing.T, q Query) {
				require.NotNil(t, q.GroupQuery)
				assert.Equal(t, 3, q.Offset)
				assert.Equal(t, 50, q.Limit)
			},
		},
		{
			name:    "compound query with group_by and offset/limit",
			input:   `status = "pass" group_by(#"env") offset=10 limit=20`,
			wantErr: false,
			checkFunc: func(t *testing.T, q Query) {
				require.NotNil(t, q.GroupQuery)
				assert.Equal(t, 10, q.Offset)
				assert.Equal(t, 20, q.Limit)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			q, err := parser.Parse()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, q)
				}
			}
		})
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		queryType QueryType
		wantErr   bool
	}{
		{
			name:      "Sessions: no group_by - valid",
			input:     `status = "pass"`,
			queryType: QueryTypeSession,
			wantErr:   false,
		},
		{
			name:      "Sessions: group_by without group - invalid",
			input:     `status = "pass" group_by(session_id)`,
			queryType: QueryTypeSession,
			wantErr:   true,
		},
		{
			name:      "Sessions: group_by with group - valid",
			input:     `status = "pass" group_by(session_id) group = ("val1")`,
			queryType: QueryTypeSession,
			wantErr:   false,
		},
		{
			name:      "Testcases: no group_by - valid",
			input:     `name = "test"`,
			queryType: QueryTypeTestcase,
			wantErr:   false,
		},
		{
			name:      "Testcases: group_by without group - invalid",
			input:     `name = "test" group_by(#"env")`,
			queryType: QueryTypeTestcase,
			wantErr:   true,
		},
		{
			name:      "Testcases: group_by with group - valid",
			input:     `name = "test" group_by(#"env") group = ("prod")`,
			queryType: QueryTypeTestcase,
			wantErr:   false,
		},
		{
			name:      "Groups: no group_by - invalid",
			input:     `status = "pass"`,
			queryType: QueryTypeGroup,
			wantErr:   true,
		},
		{
			name:      "Groups: group_by without group - valid",
			input:     `status = "pass" group_by(session_id)`,
			queryType: QueryTypeGroup,
			wantErr:   false,
		},
		{
			name:      "Groups: group_by with group - valid",
			input:     `status = "pass" group_by(session_id) group = ("val1")`,
			queryType: QueryTypeGroup,
			wantErr:   false,
		},
		{
			name:      "Mismatched group values count",
			input:     `status = "pass" group_by(session_id, #"env") group = ("val1")`,
			queryType: QueryTypeGroup,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			q, err := parser.Parse()
			require.NoError(t, err, "Parse() should not error")

			err = Validate(q, tt.queryType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
