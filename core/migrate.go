package core

import (
	"embed"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/mysql"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
	"github.com/cephei8/greener/assets"
)

func normalizeSQLiteURL(databaseURL string) string {
	if !strings.HasPrefix(databaseURL, "sqlite:///") {
		return databaseURL
	}

	path := strings.TrimPrefix(databaseURL, "sqlite:///")

	if strings.HasPrefix(path, "/") {
		return "sqlite:" + path
	}

	return "sqlite:" + path
}

func Migrate(databaseURL string, verbose bool) error {
	databaseURL = normalizeSQLiteURL(databaseURL)

	u, err := url.Parse(databaseURL)
	if err != nil {
		return fmt.Errorf("invalid database URL: %w", err)
	}

	var migrationFS embed.FS
	var migrationDir string
	switch u.Scheme {
	case "sqlite":
		migrationFS = assets.SqliteMigrationFS
		migrationDir = assets.SqliteMigrationDir
	case "postgres":
		migrationFS = assets.PostgresMigrationFS
		migrationDir = assets.PostgresMigrationDir
	case "mysql":
		migrationFS = assets.MysqlMigrationFS
		migrationDir = assets.MysqlMigrationDir
	default:
		return fmt.Errorf(
			"unsupported database protocol: %s (supported: sqlite, postgres, mysql)",
			u.Scheme,
		)
	}

	db := dbmate.New(u)
	db.FS = migrationFS
	db.MigrationsDir = []string{migrationDir}
	db.AutoDumpSchema = false

	if !verbose {
		db.Verbose = false
		db.Log = io.Discard
	}

	if err := db.CreateAndMigrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
