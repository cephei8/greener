package model_db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/schema"
)

var Dialect schema.Dialect

type BinaryUUID uuid.UUID

func (u *BinaryUUID) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("cannot scan nil into BinaryUUID")
	}

	switch Dialect.(type) {
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
		return fmt.Errorf("unknown dialect type: %T", Dialect)
	}
}

func (u BinaryUUID) Value() (driver.Value, error) {
	id := uuid.UUID(u)

	switch Dialect.(type) {
	case *pgdialect.Dialect:
		return id.String(), nil
	case *mysqldialect.Dialect, *sqlitedialect.Dialect:
		bytes, err := id.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal UUID to binary: %w", err)
		}
		return bytes, nil
	default:
		return nil, fmt.Errorf("unknown dialect type: %T", Dialect)
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

	ID           BinaryUUID `bun:"id,notnull"`
	Username     string     `bun:"username,notnull"`
	PasswordSalt []byte     `bun:"password_salt,notnull"`
	PasswordHash []byte     `bun:"password_hash,notnull"`
	CreatedAt    time.Time  `bun:"created_at,nullzero,notnull"`
	UpdatedAt    time.Time  `bun:"updated_at,nullzero,notnull"`
}

type APIKey struct {
	bun.BaseModel `bun:"table:apikeys"`

	ID          BinaryUUID `bun:"id,notnull"`
	Description *string    `bun:"description"`
	SecretSalt  []byte     `bun:"secret_salt,notnull"`
	SecretHash  []byte     `bun:"secret_hash,notnull"`
	CreatedAt   time.Time  `bun:"created_at,nullzero,notnull"`
	UpdatedAt   time.Time  `bun:"updated_at,nullzero,notnull"`
	UserID      BinaryUUID `bun:"user_id,notnull"`
}

type Session struct {
	bun.BaseModel `bun:"table:sessions"`

	ID          BinaryUUID      `bun:"id,notnull"`
	Description *string         `bun:"description"`
	Baggage     json.RawMessage `bun:"baggage"`
	CreatedAt   time.Time       `bun:"created_at,nullzero,notnull"`
	UpdatedAt   time.Time       `bun:"updated_at,nullzero,notnull"`
	UserID      BinaryUUID      `bun:"user_id,notnull"`
}

type Label struct {
	bun.BaseModel `bun:"table:labels"`

	ID        int64      `bun:"id,notnull,autoincrement"`
	SessionID BinaryUUID `bun:"session_id,notnull"`
	Key       string     `bun:"key,notnull"`
	Value     *string    `bun:"value"`
	UserID    BinaryUUID `bun:"user_id,notnull"`
	CreatedAt time.Time  `bun:"created_at,nullzero,notnull"`
	UpdatedAt time.Time  `bun:"updated_at,nullzero,notnull"`
}

type TestcaseStatus int

const (
	StatusError TestcaseStatus = iota
	StatusFail
	StatusPass
	StatusSkip
)

type Testcase struct {
	bun.BaseModel `bun:"table:testcases"`

	ID        BinaryUUID      `bun:"id,notnull"`
	SessionID BinaryUUID      `bun:"session_id,notnull"`
	Name      string          `bun:"name,notnull"`
	Classname *string         `bun:"classname"`
	File      *string         `bun:"file"`
	Testsuite *string         `bun:"testsuite"`
	Output    *string         `bun:"output"`
	Status    TestcaseStatus  `bun:"status,notnull"`
	Baggage   json.RawMessage `bun:"baggage"`
	CreatedAt time.Time       `bun:"created_at,nullzero,notnull"`
	UpdatedAt time.Time       `bun:"updated_at,nullzero,notnull"`
	UserID    BinaryUUID      `bun:"user_id,notnull"`
}
