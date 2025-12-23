package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cephei8/greener/core/dbutil"
	"github.com/cephei8/greener/core/model/db"
	"github.com/google/uuid"
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
				Usage: "Create a new user",
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
			fmt.Printf("Connecting to database at: %s\n", url)

			db, err := dbutil.Init(url)
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

	db, err := dbutil.Init(url)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	salt, passwordHash, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model_db.User{
		ID:           model_db.BinaryUUID(uuid.New()),
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

func hashPassword(password string) (salt, hash []byte, err error) {
	saltBytes := make([]byte, 32)
	_, err = rand.Read(saltBytes)
	if err != nil {
		return nil, nil, err
	}

	hashBytes := pbkdf2.Key([]byte(password), saltBytes, 100000, 32, sha256.New)

	return saltBytes, hashBytes, nil
}
