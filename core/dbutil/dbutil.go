package dbutil

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/cephei8/greener/core/model/db"
	_ "github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/driver/sqliteshim"
)

func Init(dbURL string) (*bun.DB, error) {
	var db *sql.DB

	url := strings.ToLower(dbURL)

	if strings.HasPrefix(url, "postgres://") {
		db = sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dbURL)))
		model_db.Dialect = pgdialect.New()
	} else if strings.HasPrefix(url, "mysql://") {
		var err error
		dbURL, err = convertMySQLURL(dbURL)
		if err != nil {
			return nil, fmt.Errorf("failed to convert MySQL URL: %w", err)
		}
		db, err = sql.Open("mysql", dbURL)
		if err != nil {
			return nil, fmt.Errorf("failed to open mysql connection: %w", err)
		}
		model_db.Dialect = mysqldialect.New()
	} else if strings.HasPrefix(url, "sqlite:///") || strings.HasPrefix(url, "sqlite:") {
		url := dbURL
		if cut, ok := strings.CutPrefix(url, "sqlite:///"); ok {
			url = cut
			if url == "/:memory:" {
				url = ":memory:"
			}
		} else if cut, ok := strings.CutPrefix(url, "sqlite:"); ok {
			url = cut
		}
		var err error
		db, err = sql.Open(sqliteshim.ShimName, url)
		if err != nil {
			return nil, fmt.Errorf("failed to open sqlite connection: %w", err)
		}

		_, err = db.Exec(`
			PRAGMA journal_mode=WAL;
			PRAGMA busy_timeout=5000;
			PRAGMA foreign_keys=ON;
		`)
		if err != nil {
			return nil, fmt.Errorf("failed to configure sqlite: %w", err)
		}

		model_db.Dialect = sqlitedialect.New()
	} else {
		return nil, fmt.Errorf("unsupported database URL scheme: %s (supported: postgres://, mysql://, sqlite:///)", dbURL)
	}

	bunDB := bun.NewDB(db, model_db.Dialect)
	if err := bunDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return bunDB, nil
}

func convertMySQLURL(sqlalchemyURL string) (string, error) {
	rest := strings.TrimPrefix(sqlalchemyURL, "mysql://")

	var userPass, hostAndDB string
	if atIdx := strings.Index(rest, "@"); atIdx != -1 {
		userPass = rest[:atIdx]
		hostAndDB = rest[atIdx+1:]
	} else {
		hostAndDB = rest
	}

	slashIdx := strings.Index(hostAndDB, "/")
	if slashIdx == -1 {
		return "", fmt.Errorf("invalid MySQL URL: missing database name")
	}

	host := hostAndDB[:slashIdx]
	dbAndParams := hostAndDB[slashIdx+1:]

	if !strings.Contains(host, ":") {
		host = host + ":3306"
	}

	var dsn string
	if userPass != "" {
		dsn = fmt.Sprintf("%s@tcp(%s)/%s", userPass, host, dbAndParams)
	} else {
		dsn = fmt.Sprintf("tcp(%s)/%s", host, dbAndParams)
	}

	return dsn, nil
}
