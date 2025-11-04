package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli/v3"
	"golang.org/x/crypto/pbkdf2"
)

func main() {
	app := &cli.Command{
		Name:  "greener-admin-cli",
		Usage: "Admin utility for Greener",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "db-url",
				Usage:    "Database connection URL",
				Required: true,
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "create-user",
				Usage: "Create a new user in the users table",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "username",
						Usage:    "User name",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "password",
						Usage:    "User password",
						Required: true,
					},
				},
				Action: createUserAction,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			url := cmd.String("db-url")
			dbType, err := detectDBType(url)
			if err != nil {
				return err
			}
			fmt.Printf("Connecting to %s database at: %s\n", dbType, url)

			db, err := initDB(url, dbType)
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			defer db.Close()

			fmt.Println("Database initialized successfully")
			return nil
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func createUserAction(ctx context.Context, cmd *cli.Command) error {
	url := cmd.String("db-url")
	username := cmd.String("username")
	password := cmd.String("password")

	dbType, err := detectDBType(url)
	if err != nil {
		return err
	}

	db, err := initDB(url, dbType)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	salt, passwordHash, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	id := uuid.New()
	now := time.Now()

	var query string
	var args []any

	switch strings.ToLower(dbType) {
	case "postgresql":
		query = `INSERT INTO users (id, username, password_salt, password_hash, created_at, updated_at)
		         VALUES ($1, $2, $3, $4, $5, $6)`
		args = []any{id, username, salt, passwordHash, now, now}

	case "mysql":
		query = `INSERT INTO users (id, username, password_salt, password_hash, created_at, updated_at)
		         VALUES (?, ?, ?, ?, ?, ?)`
		args = []any{id[:], username, salt, passwordHash, now, now}

	case "sqlite":
		query = `INSERT INTO users (id, username, password_salt, password_hash, created_at, updated_at)
		         VALUES (?, ?, ?, ?, ?, ?)`
		args = []any{id[:], username, salt, passwordHash, now.Format(time.RFC3339), now.Format(time.RFC3339)}

	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	_, err = db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	fmt.Printf("User created successfully: %s (ID: %s)\n", username, id)
	return nil
}

func detectDBType(url string) (string, error) {
	lowerURL := strings.ToLower(url)

	if strings.HasPrefix(lowerURL, "postgresql://") {
		return "postgresql", nil
	}
	if strings.HasPrefix(lowerURL, "mysql://") {
		return "mysql", nil
	}
	if strings.HasPrefix(lowerURL, "sqlite://") {
		return "sqlite", nil
	}

	return "", fmt.Errorf(
		"unable to detect database type from URL: %s (supported: postgresql://, mysql:// or sqlite://)",
		url,
	)
}

func hashPassword(password string) (salt, hash []byte, err error) {
	saltBytes := make([]byte, 32)
	_, err = rand.Read(saltBytes)
	if err != nil {
		return nil, nil, err
	}

	hashBytes := pbkdf2.Key([]byte(password), saltBytes, 100000, 32, sha256.New)

	return saltBytes, hashBytes, nil
}

func initDB(url, dbType string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	switch strings.ToLower(dbType) {
	case "postgresql":
		db, err = sql.Open("postgres", url)
		if err != nil {
			return nil, fmt.Errorf("failed to open postgres connection: %w", err)
		}
	case "mysql":
		mysqlDSN := url
		if strings.HasPrefix(url, "mysql://") {
			mysqlDSN, err = convertMySQLURL(url)
			if err != nil {
				return nil, fmt.Errorf("failed to convert MySQL URL: %w", err)
			}
		}
		db, err = sql.Open("mysql", mysqlDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to open mysql connection: %w", err)
		}
	case "sqlite":
		sqliteURL := url
		if after, ok := strings.CutPrefix(url, "sqlite://"); ok {
			sqliteURL = after
			if sqliteURL == "/:memory:" {
				sqliteURL = ":memory:"
			}
		}
		db, err = sql.Open("sqlite3", sqliteURL)
		if err != nil {
			return nil, fmt.Errorf("failed to open sqlite connection: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %s (supported: postgresql, mysql, sqlite)", dbType)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
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
