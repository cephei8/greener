package main

import (
	"embed"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/mysql"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
)

//go:embed db/sqlite/*.sql
var sqliteMigrationFS embed.FS

//go:embed db/postgres/*.sql
var postgresMigrationFS embed.FS

//go:embed db/mysql/*.sql
var mysqlMigrationFS embed.FS

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

func main() {
	var databaseURL string
	flag.StringVar(&databaseURL, "u", "", "Database URL")
	flag.Parse()

	if databaseURL == "" {
		fmt.Fprintf(os.Stderr, "Error: database URL is required\n")
		fmt.Fprintf(os.Stderr, "Usage: migration -u <database-url>\n")
		os.Exit(1)
	}

	databaseURL = normalizeSQLiteURL(databaseURL)

	u, err := url.Parse(databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid database URL: %v\n", err)
		os.Exit(1)
	}

	var migrationFS embed.FS
	var migrationDir string
	switch u.Scheme {
	case "sqlite":
		migrationFS = sqliteMigrationFS
		migrationDir = "db/sqlite"
	case "postgres":
		migrationFS = postgresMigrationFS
		migrationDir = "db/postgres"
	case "mysql":
		migrationFS = mysqlMigrationFS
		migrationDir = "db/mysql"
	default:
		fmt.Fprintf(
			os.Stderr,
			"Unsupported database protocol: %s (supported: sqlite, postgres, mysql)\n",
			u.Scheme,
		)
		os.Exit(1)
	}

	db := dbmate.New(u)
	db.FS = migrationFS
	db.MigrationsDir = []string{migrationDir}
	db.AutoDumpSchema = false

	if err := db.CreateAndMigrate(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run migrations: %v\n", err)
		os.Exit(1)
	}

}
