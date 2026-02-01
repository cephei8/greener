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
	"github.com/google/uuid"
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
		Role:         model_db.RoleEditor,
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
