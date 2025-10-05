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
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/urfave/cli/v3"
	"golang.org/x/crypto/pbkdf2"
)

type User struct {
	Id           uuid.UUID `bun:"id,notnull"`
	Username     string    `bun:"username,notnull"`
	PasswordSalt []byte    `bun:"password_salt,notnull"`
	PasswordHash []byte    `bun:"password_hash,notnull"`
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull"`
	UpdatedAt    time.Time `bun:"updated_at,nullzero,notnull"`
}

type UserSqlite struct {
	bun.BaseModel `bun:"table:users"`

	Id           []byte    `bun:"id,notnull"`
	Username     string    `bun:"username,notnull"`
	PasswordSalt []byte    `bun:"password_salt,notnull"`
	PasswordHash []byte    `bun:"password_hash,notnull"`
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull"`
	UpdatedAt    time.Time `bun:"updated_at,nullzero,notnull"`
}

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
			&cli.StringFlag{
				Name:     "db-type",
				Usage:    "Database type (postgres, mysql, sqlite)",
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
			dbType := cmd.String("db-type")
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
	dbType := cmd.String("db-type")
	username := cmd.String("username")
	email := cmd.String("email")
	name := cmd.String("name")
	password := cmd.String("password")

	db, err := initDB(url, dbType)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	salt, passwordHash, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	switch db.Dialect().Name() {
	case dialect.SQLite:
		id := uuid.New()
		user := &UserSqlite{
			Id:           id[:],
			Username:     username,
			PasswordSalt: salt,
			PasswordHash: passwordHash,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		_, err = db.NewInsert().Model(user).Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	default:
		user := &User{
			Id:           uuid.New(),
			Username:     username,
			PasswordSalt: salt,
			PasswordHash: passwordHash,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		_, err = db.NewInsert().Model(user).Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	}

	fmt.Printf("User created successfully: %s (%s) - %s\n", name, email, username)
	return nil
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

func initDB(url, dbType string) (*bun.DB, error) {
	var sqldb *sql.DB
	var db *bun.DB

	switch strings.ToLower(dbType) {
	case "postgres", "postgresql":
		sqldb = sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(url)))
		db = bun.NewDB(sqldb, pgdialect.New())
	case "mysql":
		var err error
		sqldb, err = sql.Open("mysql", url)
		if err != nil {
			return nil, fmt.Errorf("failed to open mysql connection: %w", err)
		}
		db = bun.NewDB(sqldb, mysqldialect.New())
	case "sqlite", "sqlite3":
		sqldb, err := sql.Open(sqliteshim.ShimName, url)
		if err != nil {
			return nil, fmt.Errorf("failed to open sqlite connection: %w", err)
		}
		db = bun.NewDB(sqldb, sqlitedialect.New())
	default:
		return nil, fmt.Errorf("unsupported database type: %s (supported: postgres, mysql, sqlite)", dbType)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
