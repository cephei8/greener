package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/schema"
	"github.com/urfave/cli/v3"
	"golang.org/x/crypto/pbkdf2"
)

type DBType int

const (
	DBTypePostgres DBType = iota
	DBTypeMySQL
	DBTypeSQLite
)

var dbDialect schema.Dialect

type BinaryUUID uuid.UUID

func (u *BinaryUUID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("cannot scan nil into BinaryUUID")
	}

	switch dbDialect.(type) {
	case *pgdialect.Dialect:
		switch v := value.(type) {
		case string:
			parsed, err := uuid.Parse(v)
			if err != nil {
				return fmt.Errorf("failed to parse UUID string: %w", err)
			}
			*u = BinaryUUID(parsed)
			return nil
		case []byte:
			parsed, err := uuid.ParseBytes(v)
			if err != nil {
				return fmt.Errorf("failed to parse UUID bytes: %w", err)
			}
			*u = BinaryUUID(parsed)
			return nil
		default:
			return fmt.Errorf("unsupported type for PostgreSQL UUID: %T", value)
		}
	case *mysqldialect.Dialect, *sqlitedialect.Dialect:
		bytes, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("expected []byte for MySQL/SQLite UUID, got %T", value)
		}
		if len(bytes) != 16 {
			return fmt.Errorf("expected 16 bytes for UUID, got %d", len(bytes))
		}
		parsed, err := uuid.FromBytes(bytes)
		if err != nil {
			return fmt.Errorf("failed to parse UUID from bytes: %w", err)
		}
		*u = BinaryUUID(parsed)
		return nil
	default:
		return fmt.Errorf("unknown dialect type: %T", dbDialect)
	}
}

func (u BinaryUUID) Value() (driver.Value, error) {
	id := uuid.UUID(u)

	switch dbDialect.(type) {
	case *pgdialect.Dialect:
		return id.String(), nil
	case *mysqldialect.Dialect, *sqlitedialect.Dialect:
		bytes, err := id.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal UUID to binary: %w", err)
		}
		return bytes, nil
	default:
		return nil, fmt.Errorf("unknown dialect type: %T", dbDialect)
	}
}

func (u BinaryUUID) String() string {
	return uuid.UUID(u).String()
}

func (u BinaryUUID) UUID() uuid.UUID {
	return uuid.UUID(u)
}

type User struct {
	bun.BaseModel `bun:"table:users"`

	Id           BinaryUUID `bun:"id,notnull"`
	Username     string     `bun:"username,notnull"`
	PasswordSalt []byte     `bun:"password_salt,notnull"`
	PasswordHash []byte     `bun:"password_hash,notnull"`
	CreatedAt    time.Time  `bun:"created_at,nullzero,notnull"`
	UpdatedAt    time.Time  `bun:"updated_at,nullzero,notnull"`
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
			fmt.Printf("Connecting to %d database at: %s\n", dbType, url)

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

	user := &User{
		Id:           BinaryUUID(uuid.New()),
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

	fmt.Printf("User created successfully: %s\n", username)
	return nil
}

func detectDBType(url string) (DBType, error) {
	lowerURL := strings.ToLower(url)

	if strings.HasPrefix(lowerURL, "postgres://") {
		return DBTypePostgres, nil
	}
	if strings.HasPrefix(lowerURL, "mysql://") {
		return DBTypeMySQL, nil
	}
	if strings.HasPrefix(lowerURL, "sqlite:///") {
		return DBTypeSQLite, nil
	}

	return 0, fmt.Errorf(
		"unable to detect database type from URL: %s (supported: postgres://, mysql:// or sqlite:///)",
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

func initDB(url string, dbType DBType) (*bun.DB, error) {
	var sqldb *sql.DB

	switch dbType {
	case DBTypePostgres:
		sqldb = sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(url)))
		dbDialect = pgdialect.New()
	case DBTypeMySQL:
		var err error
		if strings.HasPrefix(url, "mysql://") {
			url, err = convertMySQLURL(url)
			if err != nil {
				return nil, fmt.Errorf("failed to convert MySQL URL: %w", err)
			}
		}
		sqldb, err = sql.Open("mysql", url)
		if err != nil {
			return nil, fmt.Errorf("failed to open mysql connection: %w", err)
		}
		dbDialect = mysqldialect.New()
	case DBTypeSQLite:
		if cut, ok := strings.CutPrefix(url, "sqlite:///"); ok {
			url = cut
			if url == "/:memory:" {
				url = ":memory:"
			}
		}
		var err error
		sqldb, err = sql.Open(sqliteshim.ShimName, url)
		if err != nil {
			return nil, fmt.Errorf("failed to open sqlite connection: %w", err)
		}
		dbDialect = sqlitedialect.New()
	default:
		return nil, fmt.Errorf("unsupported database type: %d (supported: postgres, mysql, sqlite)", dbType)
	}

	db := bun.NewDB(sqldb, dbDialect)
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
