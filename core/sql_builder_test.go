package core_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cephei8/greener/core"
	"github.com/cephei8/greener/core/model/db"
	"github.com/cephei8/greener/core/query"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

type BaseSuite struct {
	suite.Suite
	sqlDb       *sql.DB
	db          *bun.DB
	userID      model_db.BinaryUUID
	session1Id  uuid.UUID
	session2Id  uuid.UUID
	session3Id  uuid.UUID
	testcase1Id uuid.UUID
	testcase2Id uuid.UUID
	testcase3Id uuid.UUID
	testcase4Id uuid.UUID
	testcase5Id uuid.UUID
	testcase6Id uuid.UUID
}

func (s *BaseSuite) SetupSuite()   {}
func (s *BaseSuite) TearDownTest() {}

// Test data:
//
// session1
//
//	labels:
//	  - label1: env=production
//	  - label3: branch=main
//	testcases:
//	  - testcase1
//	  - testcase2
//
// session2
//
//		labels:
//		  - label2: env=staging
//		  - label5: branch=develop
//	   - label7: platform=linux
//		testcases:
//		  - testcase3
//		  - testcase4
//
// session3
//
//	labels:
//	  - label4: env=production
//	  - label6: branch=main
//	testcases:
//	  - testcase5
//	  - testcase6
func (s *BaseSuite) setupTestData() {
	ctx := context.Background()

	userId := uuid.New()
	session1Id := uuid.New()
	session2Id := uuid.New()
	session3Id := uuid.New()

	now := time.Now()

	user := &model_db.User{
		ID:           model_db.BinaryUUID(userId),
		Username:     "testuser",
		PasswordSalt: []byte("salt"),
		PasswordHash: []byte("hash"),
	}
	_, err := s.db.NewInsert().Model(user).Exec(ctx)
	s.Require().NoError(err)

	session1 := &model_db.Session{
		ID:          model_db.BinaryUUID(session1Id),
		Description: stringPtr("First test session"),
		CreatedAt:   now,
		UpdatedAt:   now,
		UserID:      model_db.BinaryUUID(userId),
	}
	_, err = s.db.NewInsert().Model(session1).Exec(ctx)
	s.Require().NoError(err)

	session2 := &model_db.Session{
		ID:          model_db.BinaryUUID(session2Id),
		Description: stringPtr("Second test session"),
		CreatedAt:   now.Add(time.Second),
		UpdatedAt:   now.Add(time.Second),
		UserID:      model_db.BinaryUUID(userId),
	}
	_, err = s.db.NewInsert().Model(session2).Exec(ctx)
	s.Require().NoError(err)

	session3 := &model_db.Session{
		ID:          model_db.BinaryUUID(session3Id),
		Description: stringPtr("Third test session"),
		CreatedAt:   now.Add(2 * time.Second),
		UpdatedAt:   now.Add(2 * time.Second),
		UserID:      model_db.BinaryUUID(userId),
	}
	_, err = s.db.NewInsert().Model(session3).Exec(ctx)
	s.Require().NoError(err)

	testcase1Id := uuid.New()
	testcase1 := &model_db.Testcase{
		ID:        model_db.BinaryUUID(testcase1Id),
		SessionID: model_db.BinaryUUID(session1Id),
		Name:      "test_login_success",
		Classname: stringPtr("TestAuth"),
		File:      stringPtr("test_auth.py"),
		Testsuite: stringPtr("auth_tests"),
		Status:    model_db.StatusPass,
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    model_db.BinaryUUID(userId),
	}
	_, err = s.db.NewInsert().Model(testcase1).Exec(ctx)
	s.Require().NoError(err)

	testcase2Id := uuid.New()
	testcase2 := &model_db.Testcase{
		ID:        model_db.BinaryUUID(testcase2Id),
		SessionID: model_db.BinaryUUID(session1Id),
		Name:      "test_login_failure",
		Classname: stringPtr("TestAuth"),
		File:      stringPtr("test_auth.py"),
		Testsuite: stringPtr("auth_tests"),
		Status:    model_db.StatusFail,
		CreatedAt: now.Add(time.Second),
		UpdatedAt: now.Add(time.Second),
		UserID:    model_db.BinaryUUID(userId),
	}
	_, err = s.db.NewInsert().Model(testcase2).Exec(ctx)
	s.Require().NoError(err)

	testcase3Id := uuid.New()
	testcase3 := &model_db.Testcase{
		ID:        model_db.BinaryUUID(testcase3Id),
		SessionID: model_db.BinaryUUID(session2Id),
		Name:      "test_api_endpoint",
		Classname: stringPtr("TestAPI"),
		File:      stringPtr("test_api.py"),
		Testsuite: stringPtr("api_tests"),
		Status:    model_db.StatusPass,
		CreatedAt: now.Add(2 * time.Second),
		UpdatedAt: now.Add(2 * time.Second),
		UserID:    model_db.BinaryUUID(userId),
	}
	_, err = s.db.NewInsert().Model(testcase3).Exec(ctx)
	s.Require().NoError(err)

	testcase4Id := uuid.New()
	testcase4 := &model_db.Testcase{
		ID:        model_db.BinaryUUID(testcase4Id),
		SessionID: model_db.BinaryUUID(session2Id),
		Name:      "test_api_error",
		Classname: stringPtr("TestAPI"),
		File:      stringPtr("test_api.py"),
		Testsuite: stringPtr("api_tests"),
		Status:    model_db.StatusError,
		CreatedAt: now.Add(3 * time.Second),
		UpdatedAt: now.Add(3 * time.Second),
		UserID:    model_db.BinaryUUID(userId),
	}
	_, err = s.db.NewInsert().Model(testcase4).Exec(ctx)
	s.Require().NoError(err)

	testcase5Id := uuid.New()
	testcase5 := &model_db.Testcase{
		ID:        model_db.BinaryUUID(testcase5Id),
		SessionID: model_db.BinaryUUID(session3Id),
		Name:      "test_all_pass_1",
		Classname: stringPtr("TestAPI"),
		File:      stringPtr("test_api.py"),
		Testsuite: stringPtr("api_tests"),
		Status:    model_db.StatusPass,
		CreatedAt: now.Add(4 * time.Second),
		UpdatedAt: now.Add(4 * time.Second),
		UserID:    model_db.BinaryUUID(userId),
	}
	_, err = s.db.NewInsert().Model(testcase5).Exec(ctx)
	s.Require().NoError(err)

	testcase6Id := uuid.New()
	testcase6 := &model_db.Testcase{
		ID:        model_db.BinaryUUID(testcase6Id),
		SessionID: model_db.BinaryUUID(session3Id),
		Name:      "test_all_pass_2",
		Classname: stringPtr("TestAPI"),
		File:      stringPtr("test_api.py"),
		Testsuite: stringPtr("api_tests"),
		Status:    model_db.StatusPass,
		CreatedAt: now.Add(5 * time.Second),
		UpdatedAt: now.Add(5 * time.Second),
		UserID:    model_db.BinaryUUID(userId),
	}
	_, err = s.db.NewInsert().Model(testcase6).Exec(ctx)
	s.Require().NoError(err)

	label1 := &model_db.Label{
		SessionID: model_db.BinaryUUID(session1Id),
		Key:       "env",
		Value:     stringPtr("production"),
		UserID:    model_db.BinaryUUID(userId),
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err = s.db.NewInsert().Model(label1).Exec(ctx)
	s.Require().NoError(err)

	label2 := &model_db.Label{
		SessionID: model_db.BinaryUUID(session2Id),
		Key:       "env",
		Value:     stringPtr("staging"),
		UserID:    model_db.BinaryUUID(userId),
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err = s.db.NewInsert().Model(label2).Exec(ctx)
	s.Require().NoError(err)

	label3 := &model_db.Label{
		SessionID: model_db.BinaryUUID(session1Id),
		Key:       "branch",
		Value:     stringPtr("main"),
		UserID:    model_db.BinaryUUID(userId),
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err = s.db.NewInsert().Model(label3).Exec(ctx)
	s.Require().NoError(err)

	label4 := &model_db.Label{
		SessionID: model_db.BinaryUUID(session3Id),
		Key:       "env",
		Value:     stringPtr("production"),
		UserID:    model_db.BinaryUUID(userId),
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err = s.db.NewInsert().Model(label4).Exec(ctx)
	s.Require().NoError(err)

	label5 := &model_db.Label{
		SessionID: model_db.BinaryUUID(session2Id),
		Key:       "branch",
		Value:     stringPtr("develop"),
		UserID:    model_db.BinaryUUID(userId),
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err = s.db.NewInsert().Model(label5).Exec(ctx)
	s.Require().NoError(err)

	label6 := &model_db.Label{
		SessionID: model_db.BinaryUUID(session3Id),
		Key:       "branch",
		Value:     stringPtr("main"),
		UserID:    model_db.BinaryUUID(userId),
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err = s.db.NewInsert().Model(label6).Exec(ctx)
	s.Require().NoError(err)

	label7 := &model_db.Label{
		SessionID: model_db.BinaryUUID(session2Id),
		Key:       "platform",
		Value:     stringPtr("linux"),
		UserID:    model_db.BinaryUUID(userId),
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err = s.db.NewInsert().Model(label7).Exec(ctx)
	s.Require().NoError(err)

	s.userID = model_db.BinaryUUID(userId)
	s.session1Id = session1Id
	s.session2Id = session2Id
	s.session3Id = session3Id
	s.testcase1Id = testcase1Id
	s.testcase2Id = testcase2Id
	s.testcase3Id = testcase3Id
	s.testcase4Id = testcase4Id
	s.testcase5Id = testcase5Id
	s.testcase6Id = testcase6Id
}

type SQLiteSuite struct {
	BaseSuite
	dbPath string
}

func (s *SQLiteSuite) SetupSuite() {
	tempFile, err := os.CreateTemp("", "testdb_groups_*.sqlite")
	s.Require().NoError(err)
	tempFile.Close()

	s.dbPath = tempFile.Name()

	dbURL := fmt.Sprintf("sqlite:%s", s.dbPath)
	err = core.Migrate(dbURL, testing.Verbose())
	s.Require().NoError(err)

	sqlDb, err := sql.Open(sqliteshim.ShimName, s.dbPath)
	s.Require().NoError(err)

	s.sqlDb = sqlDb
	s.db = bun.NewDB(sqlDb, sqlitedialect.New())
	model_db.Dialect = sqlitedialect.New()

	s.setupTestData()
}

func (s *SQLiteSuite) TearDownSuite() {
	s.sqlDb.Close()
	os.Remove(s.dbPath)
}

type PostgresSuite struct {
	BaseSuite
	container *postgres.PostgresContainer
}

func (s *PostgresSuite) SetupSuite() {
	ctx := context.Background()

	var err error
	s.container, err = postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	s.Require().NoError(err)

	connStr, err := s.container.ConnectionString(ctx, "sslmode=disable")
	s.Require().NoError(err)

	err = core.Migrate(connStr, testing.Verbose())
	s.Require().NoError(err)

	sqlDb, err := sql.Open("postgres", connStr)
	s.Require().NoError(err)

	s.sqlDb = sqlDb
	s.db = bun.NewDB(sqlDb, pgdialect.New())
	model_db.Dialect = pgdialect.New()

	s.setupTestData()
}

func (s *PostgresSuite) TearDownSuite() {
	s.sqlDb.Close()
	s.container.Terminate(context.Background())
}

type MySQLSuite struct {
	BaseSuite
	container *mysql.MySQLContainer
}

func (s *MySQLSuite) SetupSuite() {
	ctx := context.Background()

	var err error
	s.container, err = mysql.Run(ctx,
		"mysql:8.0",
		mysql.WithDatabase("testdb"),
		mysql.WithUsername("testuser"),
		mysql.WithPassword("testpass"),
	)
	s.Require().NoError(err)

	dsnStr, err := s.container.ConnectionString(ctx)
	s.Require().NoError(err)

	mysqlURL := dsnStr
	if strings.Contains(mysqlURL, "@tcp(") {
		mysqlURL = strings.Replace(mysqlURL, "@tcp(", "@", 1)
		mysqlURL = strings.Replace(mysqlURL, ")/", "/", 1)
	}
	if !strings.HasPrefix(mysqlURL, "mysql://") {
		mysqlURL = "mysql://" + mysqlURL
	}

	err = core.Migrate(mysqlURL, testing.Verbose())
	s.Require().NoError(err)

	sqlDb, err := sql.Open("mysql", dsnStr)
	s.Require().NoError(err)

	s.sqlDb = sqlDb
	s.db = bun.NewDB(sqlDb, mysqldialect.New())
	model_db.Dialect = mysqldialect.New()

	s.setupTestData()
}

func (s *MySQLSuite) TearDownSuite() {
	s.sqlDb.Close()
	s.container.Terminate(context.Background())
}

func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func (s *BaseSuite) TestTestcases() {
	type testcaseRow struct {
		ID        model_db.BinaryUUID     `bun:"id"`
		SessionID model_db.BinaryUUID     `bun:"session_id"`
		Name      string                  `bun:"name"`
		Classname *string                 `bun:"classname"`
		File      *string                 `bun:"file"`
		Testsuite *string                 `bun:"testsuite"`
		Output    *string                 `bun:"output"`
		Status    model_db.TestcaseStatus `bun:"status"`
		Baggage   []byte                  `bun:"baggage"`
		CreatedAt time.Time               `bun:"created_at"`
		UpdatedAt time.Time               `bun:"updated_at"`
		UserID    model_db.BinaryUUID     `bun:"user_id"`
	}

	tests := []struct {
		name        string
		queryAST    query.Query
		expectedIds []uuid.UUID
	}{
		{
			name: "empty query - returns all testcases",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase6Id, s.testcase5Id, s.testcase4Id, s.testcase3Id, s.testcase2Id, s.testcase1Id},
		},
		{
			name: "filter by session_id (eq)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.SessionSelectQuery{
							SessionId: s.session1Id,
							Operator:  query.OpEq,
						}},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase2Id, s.testcase1Id},
		},
		{
			name: "filter by session_id (neq)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.SessionSelectQuery{
							SessionId: s.session1Id,
							Operator:  query.OpNEq,
						}},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase6Id, s.testcase5Id, s.testcase4Id, s.testcase3Id},
		},
		{
			name: "filter by testcase id (eq)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.IdSelectQuery{
							Id:       s.testcase1Id,
							Operator: query.OpEq,
						}},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase1Id},
		},
		{
			name: "filter by name (eq)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.NameSelectQuery{
							Name:     "test_login_success",
							Operator: query.OpEq,
						}},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase1Id},
		},
		{
			name: "filter by name (neq)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.NameSelectQuery{
							Name:     "test_login_success",
							Operator: query.OpNEq,
						}},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase6Id, s.testcase5Id, s.testcase4Id, s.testcase3Id, s.testcase2Id},
		},
		{
			name: "filter by classname (eq)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.ClassnameSelectQuery{
							Classname: "TestAuth",
							Operator:  query.OpEq,
						}},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase2Id, s.testcase1Id},
		},
		{
			name: "filter by file (eq)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.FileSelectQuery{
							File:     "test_auth.py",
							Operator: query.OpEq,
						}},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase2Id, s.testcase1Id},
		},
		{
			name: "filter by testsuite (eq)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.TestsuiteSelectQuery{
							Testsuite: "api_tests",
							Operator:  query.OpEq,
						}},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase6Id, s.testcase5Id, s.testcase4Id, s.testcase3Id},
		},
		{
			name: "filter by status pass",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.StatusSelectQuery{
							Status:   query.StatusPass,
							Operator: query.OpEq,
						}},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase6Id, s.testcase5Id, s.testcase3Id, s.testcase1Id},
		},
		{
			name: "filter by status fail",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.StatusSelectQuery{
							Status:   query.StatusFail,
							Operator: query.OpEq,
						}},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase2Id},
		},
		{
			name: "filter by status (neq)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.StatusSelectQuery{
							Status:   query.StatusPass,
							Operator: query.OpNEq,
						}},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase4Id, s.testcase2Id},
		},
		{
			name: "filter by tag (eq)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.TagSelectQuery{
							Tag:      "platform",
							Operator: query.OpEq,
						}},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase4Id, s.testcase3Id},
		},
		{
			name: "filter by tag (neq)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.TagSelectQuery{
							Tag:      "platform",
							Operator: query.OpNEq,
						}},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase6Id, s.testcase5Id, s.testcase2Id, s.testcase1Id},
		},
		{
			name: "filter by tag value (eq)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.TagValueSelectQuery{
							Tag:      "env",
							Value:    "production",
							Operator: query.OpEq,
						}},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase6Id, s.testcase5Id, s.testcase2Id, s.testcase1Id},
		},
		{
			name: "filter by tag value (neq)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.TagValueSelectQuery{
							Tag:      "env",
							Value:    "production",
							Operator: query.OpNEq,
						}},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase4Id, s.testcase3Id},
		},
		{
			name: "compound query - AND",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{
							Query: query.StatusSelectQuery{
								Status:   query.StatusPass,
								Operator: query.OpEq,
							},
							Operator: query.OpAnd,
						},
						{
							Query: query.ClassnameSelectQuery{
								Classname: "TestAuth",
								Operator:  query.OpEq,
							},
							Operator: query.OpAnd,
						},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase1Id},
		},
		{
			name: "compound query - OR",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{
							Query: query.StatusSelectQuery{
								Status:   query.StatusFail,
								Operator: query.OpEq,
							},
							Operator: query.OpAnd,
						},
						{
							Query: query.StatusSelectQuery{
								Status:   query.StatusError,
								Operator: query.OpEq,
							},
							Operator: query.OpOr,
						},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase4Id, s.testcase2Id},
		},
		{
			name: "compound query - AND with tag value",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{
							Query: query.TagValueSelectQuery{
								Tag:      "env",
								Value:    "staging",
								Operator: query.OpEq,
							},
							Operator: query.OpAnd,
						},
						{
							Query: query.ClassnameSelectQuery{
								Classname: "TestAPI",
								Operator:  query.OpEq,
							},
							Operator: query.OpAnd,
						},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase4Id, s.testcase3Id},
		},
		{
			name: "compound query - complex (A AND B) OR C",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{
							Query: query.StatusSelectQuery{
								Status:   query.StatusPass,
								Operator: query.OpEq,
							},
							Operator: query.OpAnd,
						},
						{
							Query: query.ClassnameSelectQuery{
								Classname: "TestAuth",
								Operator:  query.OpEq,
							},
							Operator: query.OpAnd,
						},
						{
							Query: query.StatusSelectQuery{
								Status:   query.StatusError,
								Operator: query.OpEq,
							},
							Operator: query.OpOr,
						},
					},
				},
			},
			expectedIds: []uuid.UUID{s.testcase4Id, s.testcase1Id},
		},
		{
			name: "group_by session with selector",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.SessionGroupToken{},
					},
				},
				GroupSelector: []string{s.session1Id.String()},
			},
			expectedIds: []uuid.UUID{s.testcase2Id, s.testcase1Id},
		},
		{
			name: "group_by tag with selector",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.TagGroupToken{Tag: "env"},
					},
				},
				GroupSelector: []string{"production"},
			},
			expectedIds: []uuid.UUID{s.testcase6Id, s.testcase5Id, s.testcase2Id, s.testcase1Id},
		},
		{
			name: "group_by session and tag with selectors",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.SessionGroupToken{},
						query.TagGroupToken{Tag: "branch"},
					},
				},
				GroupSelector: []string{s.session2Id.String(), "develop"},
			},
			expectedIds: []uuid.UUID{s.testcase4Id, s.testcase3Id},
		},
		{
			name: "group_by tag with selector and status filter",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.StatusSelectQuery{
							Status:   query.StatusPass,
							Operator: query.OpEq,
						}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.TagGroupToken{Tag: "env"},
					},
				},
				GroupSelector: []string{"production"},
			},
			expectedIds: []uuid.UUID{s.testcase6Id, s.testcase5Id, s.testcase1Id},
		},
	}

	farPast := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	farFuture := time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)

	dateTests := []struct {
		name        string
		queryAST    query.Query
		expectedIds []uuid.UUID
	}{
		{
			name: "filter by start_date - all testcases after far past",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				StartDate: &farPast,
			},
			expectedIds: []uuid.UUID{s.testcase6Id, s.testcase5Id, s.testcase4Id, s.testcase3Id, s.testcase2Id, s.testcase1Id},
		},
		{
			name: "filter by end_date - no testcases before far past",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				EndDate: &farPast,
			},
			expectedIds: []uuid.UUID{},
		},
		{
			name: "filter by start_date - no testcases after far future",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				StartDate: &farFuture,
			},
			expectedIds: []uuid.UUID{},
		},
		{
			name: "filter by end_date - all testcases before far future",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				EndDate: &farFuture,
			},
			expectedIds: []uuid.UUID{s.testcase6Id, s.testcase5Id, s.testcase4Id, s.testcase3Id, s.testcase2Id, s.testcase1Id},
		},
		{
			name: "filter by date range - far past to far future (all testcases)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				StartDate: &farPast,
				EndDate:   &farFuture,
			},
			expectedIds: []uuid.UUID{s.testcase6Id, s.testcase5Id, s.testcase4Id, s.testcase3Id, s.testcase2Id, s.testcase1Id},
		},
		{
			name: "date filter with status query",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.StatusSelectQuery{
							Status:   query.StatusPass,
							Operator: query.OpEq,
						}},
					},
				},
				StartDate: &farPast,
				EndDate:   &farFuture,
			},
			expectedIds: []uuid.UUID{s.testcase6Id, s.testcase5Id, s.testcase3Id, s.testcase1Id},
		},
	}

	tests = append(tests, dateTests...)

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			q, err := core.BuildTestcasesQuery(s.db, s.userID, tt.queryAST)
			require.NoError(t, err)

			type result struct {
				testcaseRow
				TotalCount       int64 `bun:"total_count"`
				AggregatedStatus int64 `bun:"aggregated_status"`
			}

			var results []result
			err = q.Scan(ctx, &results)
			require.NoError(t, err)

			actualIds := make([]uuid.UUID, len(results))
			for i, r := range results {
				actualIds[i] = uuid.UUID(r.ID)
			}

			assert.Equal(t, tt.expectedIds, actualIds)

			if len(tt.expectedIds) > 0 {
				require.NotEmpty(t, results)
				assert.Equal(t, int64(len(tt.expectedIds)), results[0].TotalCount)
			} else {
				assert.Empty(t, results, "expected no results")
			}

			if len(results) > 0 {
				minStatus := results[0].Status
				for _, r := range results {
					if r.Status < minStatus {
						minStatus = r.Status
					}
				}
				assert.Equal(t, int64(minStatus), results[0].AggregatedStatus)
			}
		})
	}
}

func (s *BaseSuite) TestTestcasesErrors() {
	tests := []struct {
		name      string
		queryAST  query.Query
		expectErr bool
		panicMsg  string
	}{
		{
			name: "group_by present but selector missing",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.SessionGroupToken{},
					},
				},
				GroupSelector: []string{},
			},
			panicMsg: "logic error: grouping/selector mismatch",
		},
		{
			name: "group_by present but selector nil",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.SessionGroupToken{},
					},
				},
				GroupSelector: nil,
			},
			panicMsg: "logic error: grouping/selector mismatch",
		},
		{
			name: "selector size mismatch - too few selectors",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.SessionGroupToken{},
						query.TagGroupToken{Tag: "env"},
					},
				},
				GroupSelector: []string{s.session1Id.String()},
			},
			panicMsg: "logic error: grouping/selector mismatch",
		},
		{
			name: "selector size mismatch - too many selectors",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.SessionGroupToken{},
					},
				},
				GroupSelector: []string{s.session1Id.String(), "production"},
			},
			panicMsg: "logic error: grouping/selector mismatch",
		},
		{
			name: "invalid session UUID in selector",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.SessionGroupToken{},
					},
				},
				GroupSelector: []string{"invalid-uuid"},
			},
			expectErr: true,
		},
		{
			name: "multiple group tokens with size mismatch - empty selector",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.SessionGroupToken{},
						query.TagGroupToken{Tag: "env"},
						query.TagGroupToken{Tag: "branch"},
					},
				},
				GroupSelector: []string{},
			},
			panicMsg: "logic error: grouping/selector mismatch",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.panicMsg != "" {
				assert.PanicsWithValue(t, tt.panicMsg, func() {
					core.BuildTestcasesQuery(s.db, s.userID, tt.queryAST)
				})
			} else if tt.expectErr {
				_, err := core.BuildTestcasesQuery(s.db, s.userID, tt.queryAST)
				assert.Error(t, err)
			}
		})
	}
}

func (s *BaseSuite) TestSessions() {
	type sessionRow struct {
		ID               model_db.BinaryUUID     `bun:"id"`
		Description      *string                 `bun:"description"`
		Baggage          []byte                  `bun:"baggage"`
		CreatedAt        time.Time               `bun:"created_at"`
		UpdatedAt        time.Time               `bun:"updated_at"`
		UserID           model_db.BinaryUUID     `bun:"user_id"`
		AggregatedStatus model_db.TestcaseStatus `bun:"aggregated_status"`
	}

	tests := []struct {
		name                    string
		queryAST                query.Query
		expectedSessionIds      []uuid.UUID
		expectedAggregatedStats map[uuid.UUID]model_db.TestcaseStatus
	}{
		{
			name: "empty query - returns all sessions",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
			},
			expectedSessionIds: []uuid.UUID{s.session3Id, s.session2Id, s.session1Id},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{
				s.session1Id: model_db.StatusFail,
				s.session2Id: model_db.StatusError,
				s.session3Id: model_db.StatusPass,
			},
		},
		{
			name: "filter by session_id (eq)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.SessionSelectQuery{
							SessionId: s.session1Id,
							Operator:  query.OpEq,
						}},
					},
				},
			},
			expectedSessionIds: []uuid.UUID{s.session1Id},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{
				s.session1Id: model_db.StatusFail,
			},
		},
		{
			name: "filter by session_id (neq)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.SessionSelectQuery{
							SessionId: s.session1Id,
							Operator:  query.OpNEq,
						}},
					},
				},
			},
			expectedSessionIds: []uuid.UUID{s.session3Id, s.session2Id},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{
				s.session2Id: model_db.StatusError,
				s.session3Id: model_db.StatusPass,
			},
		},
		{
			name: "filter by id (eq)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.IdSelectQuery{
							Id:       s.testcase3Id,
							Operator: query.OpEq,
						}},
					},
				},
			},
			expectedSessionIds: []uuid.UUID{s.session2Id},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{
				s.session2Id: model_db.StatusPass,
			},
		},
		{
			name: "filter by tag (eq)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.TagSelectQuery{
							Tag:      "env",
							Operator: query.OpEq,
						}},
					},
				},
			},
			expectedSessionIds: []uuid.UUID{s.session3Id, s.session2Id, s.session1Id},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{
				s.session1Id: model_db.StatusFail,
				s.session2Id: model_db.StatusError,
				s.session3Id: model_db.StatusPass,
			},
		},
		{
			name: "filter by tag (neq) - sessions without env tag",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.TagSelectQuery{
							Tag:      "env",
							Operator: query.OpNEq,
						}},
					},
				},
			},
			expectedSessionIds:      []uuid.UUID{},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{},
		},
		{
			name: "filter by tag value (eq) - production",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.TagValueSelectQuery{
							Tag:      "env",
							Value:    "production",
							Operator: query.OpEq,
						}},
					},
				},
			},
			expectedSessionIds: []uuid.UUID{s.session3Id, s.session1Id},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{
				s.session1Id: model_db.StatusFail,
				s.session3Id: model_db.StatusPass,
			},
		},
		{
			name: "filter by tag value (eq) - staging",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.TagValueSelectQuery{
							Tag:      "env",
							Value:    "staging",
							Operator: query.OpEq,
						}},
					},
				},
			},
			expectedSessionIds: []uuid.UUID{s.session2Id},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{
				s.session2Id: model_db.StatusError,
			},
		},
		{
			name: "filter by tag value (neq) - not production",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.TagValueSelectQuery{
							Tag:      "env",
							Value:    "production",
							Operator: query.OpNEq,
						}},
					},
				},
			},
			expectedSessionIds: []uuid.UUID{s.session2Id},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{
				s.session2Id: model_db.StatusError,
			},
		},
		{
			name: "filter by branch tag (eq)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{

					Parts: []query.CompoundSelectQueryPart{

						{Operator: query.OpAnd, Query: query.TagSelectQuery{
							Tag:      "branch",
							Operator: query.OpEq,
						}},
					},
				},
			},
			expectedSessionIds: []uuid.UUID{s.session3Id, s.session2Id, s.session1Id},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{
				s.session1Id: model_db.StatusFail,
				s.session2Id: model_db.StatusError,
				s.session3Id: model_db.StatusPass,
			},
		},
		{
			name: "compound query - AND with tag values",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{
							Query: query.TagValueSelectQuery{
								Tag:      "env",
								Value:    "production",
								Operator: query.OpEq,
							},
							Operator: query.OpAnd,
						},
						{
							Query: query.TagSelectQuery{
								Tag:      "branch",
								Operator: query.OpEq,
							},
							Operator: query.OpAnd,
						},
					},
				},
			},
			expectedSessionIds: []uuid.UUID{s.session3Id, s.session1Id},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{
				s.session1Id: model_db.StatusFail,
				s.session3Id: model_db.StatusPass,
			},
		},
		{
			name: "compound query - OR with tag values",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{
							Query: query.TagValueSelectQuery{
								Tag:      "env",
								Value:    "production",
								Operator: query.OpEq,
							},
							Operator: query.OpAnd,
						},
						{
							Query: query.TagValueSelectQuery{
								Tag:      "env",
								Value:    "staging",
								Operator: query.OpEq,
							},
							Operator: query.OpOr,
						},
					},
				},
			},
			expectedSessionIds: []uuid.UUID{s.session3Id, s.session2Id, s.session1Id},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{
				s.session1Id: model_db.StatusFail,
				s.session2Id: model_db.StatusError,
				s.session3Id: model_db.StatusPass,
			},
		},
		{
			name: "compound query - complex (A AND B) OR C",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{
							Query: query.TagValueSelectQuery{
								Tag:      "env",
								Value:    "production",
								Operator: query.OpEq,
							},
							Operator: query.OpAnd,
						},
						{
							Query: query.TagSelectQuery{
								Tag:      "branch",
								Operator: query.OpEq,
							},
							Operator: query.OpAnd,
						},
						{
							Query: query.TagValueSelectQuery{
								Tag:      "env",
								Value:    "development",
								Operator: query.OpEq,
							},
							Operator: query.OpOr,
						},
					},
				},
			},
			expectedSessionIds: []uuid.UUID{s.session3Id, s.session1Id},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{
				s.session1Id: model_db.StatusFail,
				s.session3Id: model_db.StatusPass,
			},
		},
		{
			name: "group_by session with selector",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.SessionGroupToken{},
					},
				},
				GroupSelector: []string{s.session1Id.String()},
			},
			expectedSessionIds: []uuid.UUID{s.session1Id},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{
				s.session1Id: model_db.StatusFail,
			},
		},
		{
			name: "group_by tag with selector",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.TagGroupToken{Tag: "env"},
					},
				},
				GroupSelector: []string{"staging"},
			},
			expectedSessionIds: []uuid.UUID{s.session2Id},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{
				s.session2Id: model_db.StatusError,
			},
		},
		{
			name: "group_by session and tag with selectors",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.SessionGroupToken{},
						query.TagGroupToken{Tag: "env"},
					},
				},
				GroupSelector: []string{s.session3Id.String(), "production"},
			},
			expectedSessionIds: []uuid.UUID{s.session3Id},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{
				s.session3Id: model_db.StatusPass,
			},
		},
		{
			name: "group_by tag with selector - multiple sessions match",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.TagGroupToken{Tag: "env"},
					},
				},
				GroupSelector: []string{"production"},
			},
			expectedSessionIds: []uuid.UUID{s.session3Id, s.session1Id},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{
				s.session1Id: model_db.StatusFail,
				s.session3Id: model_db.StatusPass,
			},
		},
	}

	farPast := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	farFuture := time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)

	dateTests := []struct {
		name                    string
		queryAST                query.Query
		expectedSessionIds      []uuid.UUID
		expectedAggregatedStats map[uuid.UUID]model_db.TestcaseStatus
	}{
		{
			name: "filter sessions by start_date - all sessions after far past",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				StartDate: &farPast,
			},
			expectedSessionIds: []uuid.UUID{s.session3Id, s.session2Id, s.session1Id},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{
				s.session1Id: model_db.StatusFail,
				s.session2Id: model_db.StatusError,
				s.session3Id: model_db.StatusPass,
			},
		},
		{
			name: "filter sessions by end_date - no sessions before far past",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				EndDate: &farPast,
			},
			expectedSessionIds:      []uuid.UUID{},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{},
		},
		{
			name: "filter sessions by start_date - no sessions after far future",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				StartDate: &farFuture,
			},
			expectedSessionIds:      []uuid.UUID{},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{},
		},
		{
			name: "filter sessions by end_date - all sessions before far future",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				EndDate: &farFuture,
			},
			expectedSessionIds: []uuid.UUID{s.session3Id, s.session2Id, s.session1Id},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{
				s.session1Id: model_db.StatusFail,
				s.session2Id: model_db.StatusError,
				s.session3Id: model_db.StatusPass,
			},
		},
		{
			name: "date filter with tag query",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.TagValueSelectQuery{
							Tag:      "env",
							Value:    "production",
							Operator: query.OpEq,
						}},
					},
				},
				StartDate: &farPast,
				EndDate:   &farFuture,
			},
			expectedSessionIds: []uuid.UUID{s.session3Id, s.session1Id},
			expectedAggregatedStats: map[uuid.UUID]model_db.TestcaseStatus{
				s.session1Id: model_db.StatusFail,
				s.session3Id: model_db.StatusPass,
			},
		},
	}

	tests = append(tests, dateTests...)

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			q, err := core.BuildSessionsQuery(s.db, s.userID, tt.queryAST)
			require.NoError(t, err)

			type result struct {
				sessionRow
				TotalCount int64 `bun:"total_count"`
			}

			var results []result
			err = q.Scan(ctx, &results)
			require.NoError(t, err)

			actualSessionIds := make([]uuid.UUID, len(results))
			for i, r := range results {
				actualSessionIds[i] = uuid.UUID(r.ID)
			}

			assert.Equal(t, tt.expectedSessionIds, actualSessionIds)

			if len(tt.expectedSessionIds) > 0 {
				require.NotEmpty(t, results)
				assert.Equal(t, int64(len(tt.expectedSessionIds)), results[0].TotalCount)
			} else {
				assert.Empty(t, results)
			}

			for _, r := range results {
				sessionId := uuid.UUID(r.ID)
				expectedStatus, ok := tt.expectedAggregatedStats[sessionId]
				require.True(t, ok)
				assert.Equal(t, expectedStatus, r.AggregatedStatus)
			}
		})
	}
}

func (s *BaseSuite) TestSessionsErrors() {
	tests := []struct {
		name      string
		queryAST  query.Query
		expectErr bool
		panicMsg  string
	}{
		{
			name: "group_by present but selector missing",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.SessionGroupToken{},
					},
				},
				GroupSelector: []string{},
			},
			panicMsg: "logic error: grouping/selector mismatch",
		},
		{
			name: "group_by present but selector nil",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.SessionGroupToken{},
					},
				},
				GroupSelector: nil,
			},
			panicMsg: "logic error: grouping/selector mismatch",
		},
		{
			name: "selector size mismatch - too few selectors",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.SessionGroupToken{},
						query.TagGroupToken{Tag: "env"},
					},
				},
				GroupSelector: []string{s.session1Id.String()},
			},
			panicMsg: "logic error: grouping/selector mismatch",
		},
		{
			name: "selector size mismatch - too many selectors",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.SessionGroupToken{},
					},
				},
				GroupSelector: []string{s.session1Id.String(), "production"},
			},
			panicMsg: "logic error: grouping/selector mismatch",
		},
		{
			name: "invalid session UUID in selector",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.SessionGroupToken{},
					},
				},
				GroupSelector: []string{"invalid-uuid"},
			},
			expectErr: true,
		},
		{
			name: "multiple group tokens with size mismatch - empty selector",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupQuery: &query.GroupQuery{
					Tokens: []query.GroupToken{
						query.SessionGroupToken{},
						query.TagGroupToken{Tag: "env"},
						query.TagGroupToken{Tag: "branch"},
					},
				},
				GroupSelector: []string{},
			},
			panicMsg: "logic error: grouping/selector mismatch",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.panicMsg != "" {
				assert.PanicsWithValue(t, tt.panicMsg, func() {
					core.BuildSessionsQuery(s.db, s.userID, tt.queryAST)
				})
			} else if tt.expectErr {
				_, err := core.BuildSessionsQuery(s.db, s.userID, tt.queryAST)
				assert.Error(t, err)
			}
		})
	}
}

func (s *BaseSuite) TestGroups() {
	type groupRow struct {
		SessionID        *model_db.BinaryUUID    `bun:"session_id"`
		Env              *string                 `bun:"env"`
		Branch           *string                 `bun:"branch"`
		AggregatedStatus model_db.TestcaseStatus `bun:"aggregated_status"`
		TestcaseCount    int64                   `bun:"testcase_count"`
	}

	groupKey := func(g groupRow) string {
		parts := []string{}
		if g.SessionID != nil {
			parts = append(parts, fmt.Sprintf("session:%s", uuid.UUID(*g.SessionID).String()))
		}
		if g.Env != nil {
			parts = append(parts, fmt.Sprintf("env:%s", *g.Env))
		}
		if g.Branch != nil {
			parts = append(parts, fmt.Sprintf("branch:%s", *g.Branch))
		}
		return strings.Join(parts, ",")
	}

	tests := []struct {
		name           string
		queryAST       query.Query
		groupBy        *query.GroupQuery
		expectedGroups []string
	}{
		{
			name: "group by session",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{
					query.SessionGroupToken{},
				},
			},
			expectedGroups: []string{
				fmt.Sprintf("session:%s", s.session2Id),
				fmt.Sprintf("session:%s", s.session1Id),
				fmt.Sprintf("session:%s", s.session3Id),
			},
		},
		{
			name: "group by env tag",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{
					query.TagGroupToken{Tag: "env"},
				},
			},
			expectedGroups: []string{
				"env:staging",
				"env:production",
			},
		},
		{
			name: "group by branch tag",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{
					query.TagGroupToken{Tag: "branch"},
				},
			},
			expectedGroups: []string{
				"branch:develop",
				"branch:main",
			},
		},
		{
			name: "group by session and env",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{
					query.SessionGroupToken{},
					query.TagGroupToken{Tag: "env"},
				},
			},
			expectedGroups: []string{
				fmt.Sprintf("session:%s,env:staging", s.session2Id),
				fmt.Sprintf("session:%s,env:production", s.session1Id),
				fmt.Sprintf("session:%s,env:production", s.session3Id),
			},
		},
		{
			name: "group by env and branch",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{
					query.TagGroupToken{Tag: "env"},
					query.TagGroupToken{Tag: "branch"},
				},
			},
			expectedGroups: []string{
				"env:staging,branch:develop",
				"env:production,branch:main",
			},
		},
		{
			name: "group by session with status filter",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.StatusSelectQuery{
							Status:   query.StatusPass,
							Operator: query.OpEq,
						}},
					},
				},
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{
					query.SessionGroupToken{},
				},
			},
			expectedGroups: []string{
				fmt.Sprintf("session:%s", s.session1Id),
				fmt.Sprintf("session:%s", s.session2Id),
				fmt.Sprintf("session:%s", s.session3Id),
			},
		},
		{
			name: "group by env with tag value filter",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.TagValueSelectQuery{
							Tag:      "env",
							Value:    "production",
							Operator: query.OpEq,
						}},
					},
				},
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{
					query.TagGroupToken{Tag: "env"},
				},
			},
			expectedGroups: []string{
				"env:production",
			},
		},
		{
			name: "group by env with session filter",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.SessionSelectQuery{
							SessionId: s.session1Id,
							Operator:  query.OpEq,
						}},
					},
				},
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{
					query.TagGroupToken{Tag: "env"},
				},
			},
			expectedGroups: []string{
				"env:production",
			},
		},
		{
			name: "group by session and env with compound query",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{
							Query: query.StatusSelectQuery{
								Status:   query.StatusPass,
								Operator: query.OpEq,
							},
							Operator: query.OpAnd,
						},
						{
							Query: query.TagValueSelectQuery{
								Tag:      "env",
								Value:    "production",
								Operator: query.OpEq,
							},
							Operator: query.OpAnd,
						},
					},
				},
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{
					query.SessionGroupToken{},
					query.TagGroupToken{Tag: "env"},
				},
			},
			expectedGroups: []string{
				fmt.Sprintf("session:%s,env:production", s.session1Id),
				fmt.Sprintf("session:%s,env:production", s.session3Id),
			},
		},
		{
			name: "group by session with selector",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupSelector: []string{s.session1Id.String()},
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{
					query.SessionGroupToken{},
				},
			},
			expectedGroups: []string{
				fmt.Sprintf("session:%s", s.session1Id),
			},
		},
		{
			name: "group by env with selector",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupSelector: []string{"staging"},
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{
					query.TagGroupToken{Tag: "env"},
				},
			},
			expectedGroups: []string{
				"env:staging",
			},
		},
		{
			name: "group by session and env with selectors",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupSelector: []string{s.session3Id.String(), "production"},
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{
					query.SessionGroupToken{},
					query.TagGroupToken{Tag: "env"},
				},
			},
			expectedGroups: []string{
				fmt.Sprintf("session:%s,env:production", s.session3Id),
			},
		},
	}

	farPast := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	farFuture := time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)

	dateTests := []struct {
		name           string
		queryAST       query.Query
		groupBy        *query.GroupQuery
		expectedGroups []string
	}{
		{
			name: "group by session with start_date filter - all sessions",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				StartDate: &farPast,
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{
					query.SessionGroupToken{},
				},
			},
			expectedGroups: []string{
				fmt.Sprintf("session:%s", s.session1Id),
				fmt.Sprintf("session:%s", s.session2Id),
				fmt.Sprintf("session:%s", s.session3Id),
			},
		},
		{
			name: "group by session with end_date filter - no sessions",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				EndDate: &farPast,
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{
					query.SessionGroupToken{},
				},
			},
			expectedGroups: []string{},
		},
		{
			name: "group by env with date range filter - all",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				StartDate: &farPast,
				EndDate:   &farFuture,
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{
					query.TagGroupToken{Tag: "env"},
				},
			},
			expectedGroups: []string{
				"env:staging",
				"env:production",
			},
		},
		{
			name: "group by env with date filter and status query",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.StatusSelectQuery{
							Status:   query.StatusPass,
							Operator: query.OpEq,
						}},
					},
				},
				StartDate: &farPast,
				EndDate:   &farFuture,
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{
					query.TagGroupToken{Tag: "env"},
				},
			},
			expectedGroups: []string{
				"env:staging",
				"env:production",
			},
		},
	}

	tests = append(tests, dateTests...)

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			q, err := core.BuildGroupsQuery(s.db, s.userID, tt.queryAST, tt.groupBy)
			require.NoError(t, err)

			type result struct {
				groupRow
				TotalCount int64 `bun:"total_count"`
			}

			var results []result
			err = q.Scan(ctx, &results)
			require.NoError(t, err)

			assert.Equal(t, len(tt.expectedGroups), len(results), "result count should match")

			if len(tt.expectedGroups) > 0 {
				require.NotEmpty(t, results, "expected results but got none")
				assert.Equal(t, int64(len(tt.expectedGroups)), results[0].TotalCount, "total count should match")
			} else {
				assert.Empty(t, results, "expected no results")
			}

			actualKeys := make([]string, len(results))
			for i, r := range results {
				actualKeys[i] = groupKey(r.groupRow)
			}

			for _, expectedKey := range tt.expectedGroups {
				assert.Contains(t, actualKeys, expectedKey, "expected group key not found")
			}

			for _, actualKey := range actualKeys {
				assert.Contains(t, tt.expectedGroups, actualKey, "unexpected group key found")
			}
		})
	}
}

func (s *BaseSuite) TestGroupsErrors() {
	tests := []struct {
		name      string
		queryAST  query.Query
		groupBy   *query.GroupQuery
		expectErr bool
		errMsg    string
		panicMsg  string
	}{
		{
			name: "group_by is nil",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
			},
			groupBy:   nil,
			expectErr: true,
			errMsg:    "group_by clause required for groups page",
		},
		{
			name: "group_by has no tokens",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{},
			},
			expectErr: true,
			errMsg:    "group_by clause required for groups page",
		},
		{
			name: "group_by has empty token list",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
			},
			groupBy: &query.GroupQuery{
				Tokens: nil,
			},
			expectErr: true,
			errMsg:    "group_by clause required for groups page",
		},
		{
			name: "group selector with too few selectors - should panic",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupSelector: []string{s.session1Id.String()},
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{
					query.SessionGroupToken{},
					query.TagGroupToken{Tag: "env"},
				},
			},
			panicMsg: "logic error: grouping/selector mismatch",
		},
		{
			name: "group selector with too many selectors - should panic",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupSelector: []string{s.session1Id.String(), "extra-value"},
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{
					query.SessionGroupToken{},
				},
			},
			panicMsg: "logic error: grouping/selector mismatch",
		},
		{
			name: "group selector with invalid session UUID - should error",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				GroupSelector: []string{"invalid-uuid"},
			},
			groupBy: &query.GroupQuery{
				Tokens: []query.GroupToken{
					query.SessionGroupToken{},
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.panicMsg != "" {
				assert.PanicsWithValue(t, tt.panicMsg, func() {
					core.BuildGroupsQuery(s.db, s.userID, tt.queryAST, tt.groupBy)
				})
			} else {
				_, err := core.BuildGroupsQuery(s.db, s.userID, tt.queryAST, tt.groupBy)
				if tt.expectErr {
					require.Error(t, err)
					if tt.errMsg != "" {
						assert.Contains(t, err.Error(), tt.errMsg)
					}
				} else {
					require.NoError(t, err)
				}
			}
		})
	}
}

func (s *BaseSuite) TestOffsetLimit() {
	tests := []struct {
		name          string
		queryAST      query.Query
		expectedCount int
		expectedFirst uuid.UUID
		expectErr     bool
		errMsg        string
	}{
		{
			name: "default limit (100)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
			},
			expectedCount: 6,
			expectedFirst: s.testcase6Id,
		},
		{
			name: "limit 2",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				Limit: 2,
			},
			expectedCount: 2,
			expectedFirst: s.testcase6Id,
		},
		{
			name: "offset 2",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				Offset: 2,
			},
			expectedCount: 4,
			expectedFirst: s.testcase4Id,
		},
		{
			name: "offset 2 limit 2",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				Offset: 2,
				Limit:  2,
			},
			expectedCount: 2,
			expectedFirst: s.testcase4Id,
		},
		{
			name: "limit 100 (max allowed)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				Limit: 100,
			},
			expectedCount: 6,
			expectedFirst: s.testcase6Id,
		},
		{
			name: "limit 101 (exceeds max)",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				Limit: 101,
			},
			expectErr: true,
			errMsg:    "limit cannot exceed 100",
		},
		{
			name: "negative limit",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				Limit: -1,
			},
			expectErr: true,
			errMsg:    "limit must be positive",
		},
		{
			name: "negative offset",
			queryAST: query.Query{
				SelectQuery: query.CompoundSelectQuery{
					Parts: []query.CompoundSelectQueryPart{
						{Operator: query.OpAnd, Query: query.EmptySelectQuery{}},
					},
				},
				Offset: -1,
			},
			expectErr: true,
			errMsg:    "offset must be non-negative",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			q, err := core.BuildTestcasesQuery(s.db, s.userID, tt.queryAST)
			if tt.expectErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)

			type testcaseRow struct {
				ID        model_db.BinaryUUID     `bun:"id"`
				SessionID model_db.BinaryUUID     `bun:"session_id"`
				Name      string                  `bun:"name"`
				Classname *string                 `bun:"classname"`
				File      *string                 `bun:"file"`
				Testsuite *string                 `bun:"testsuite"`
				Output    *string                 `bun:"output"`
				Status    model_db.TestcaseStatus `bun:"status"`
				Baggage   []byte                  `bun:"baggage"`
				CreatedAt time.Time               `bun:"created_at"`
				UpdatedAt time.Time               `bun:"updated_at"`
				UserID    model_db.BinaryUUID     `bun:"user_id"`
			}

			type result struct {
				testcaseRow
				TotalCount       int64 `bun:"total_count"`
				AggregatedStatus int64 `bun:"aggregated_status"`
			}

			var results []result
			err = q.Scan(ctx, &results)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCount, len(results))

			if len(results) > 0 {
				actualFirst := uuid.UUID(results[0].ID)
				assert.Equal(t, tt.expectedFirst, actualFirst)
			}
		})
	}
}

func TestSQLite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(SQLiteSuite))
}

func TestPostgres(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("Skipping Postgres tests in short mode")
	}
	suite.Run(t, new(PostgresSuite))
}

func TestMySQL(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("Skipping MySQL tests in short mode")
	}
	suite.Run(t, new(MySQLSuite))
}
