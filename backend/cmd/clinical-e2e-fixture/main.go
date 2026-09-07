// Command clinical-e2e-fixture creates or deletes a disposable clinical E2E clinic.
// It never logs passwords, hashes, cookies, or emails.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/animal-ekarte/backend/internal/clinicale2e"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/dbconn"
)

const commandTimeout = 2 * time.Minute

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "clinical-e2e-fixture: %s\n", err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: clinical-e2e-fixture setup|teardown [--clinic-id N]")
	}
	command := args[0]
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	clinicID := fs.Uint64("clinic-id", 0, "synthetic clinic id for teardown")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	appEnv := os.Getenv("APP_ENV")
	params, err := dbconn.FromEnv()
	if err != nil {
		return err
	}
	database := os.Getenv("DB_NAME")
	if database == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if err := clinicale2e.Allow(appEnv, params.Host); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	db, err := gorm.Open(postgres.Open(params.DSN(database)), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return fmt.Errorf("open database")
	}

	switch command {
	case "setup":
		password := os.Getenv("E2E_LOGIN_PASSWORD")
		if password == "" {
			return fmt.Errorf("E2E_LOGIN_PASSWORD is required")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), config.BcryptCost)
		if err != nil {
			return fmt.Errorf("hash login password")
		}
		result, err := clinicale2e.Create(ctx, db, clinicale2e.Request{
			AppEnv:       appEnv,
			DBHost:       params.Host,
			PasswordHash: string(hash),
		})
		if err != nil {
			return err
		}
		payload, err := clinicale2e.EncodeResult(result)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(os.Stdout, "%s\n", payload)
		return err
	case "teardown":
		if *clinicID == 0 {
			return fmt.Errorf("--clinic-id is required")
		}
		return clinicale2e.Delete(ctx, db, appEnv, params.Host, *clinicID)
	default:
		return fmt.Errorf("unknown command %q", strings.TrimSpace(command))
	}
}
