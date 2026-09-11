// Command synthetic-closing-fixture creates or deletes a disposable S09 closing clinic.
// It never logs passwords, hashes, cookies, or emails.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/animal-ekarte/backend/internal/billing"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/dbconn"
)

const commandTimeout = 2 * time.Minute

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "synthetic-closing-fixture: %s\n", err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: synthetic-closing-fixture setup|teardown --target-date YYYY-MM-DD|--clinic-id N --cleanup-token TOKEN")
	}
	command := args[0]
	if command != "setup" && command != "teardown" {
		return fmt.Errorf("unknown command %q", strings.TrimSpace(command))
	}
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	clinicID := fs.Uint64("clinic-id", 0, "synthetic clinic id for teardown")
	targetDate := fs.String("target-date", "", "JST calendar date for setup")
	cleanupToken := fs.String("cleanup-token", "", "cleanup token from setup")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	appEnv := os.Getenv("APP_ENV")
	params, err := dbconn.FromEnv()
	if err != nil {
		return err
	}
	if err := billing.AllowUATSyntheticClosing(appEnv, params.Host); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	database := os.Getenv("DB_NAME")
	if database == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	db, err := gorm.Open(postgres.Open(params.DSN(database)), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return fmt.Errorf("open database")
	}

	switch command {
	case "setup":
		jst, err := time.LoadLocation("Asia/Tokyo")
		if err != nil {
			return err
		}
		day, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(*targetDate), jst)
		if err != nil {
			return fmt.Errorf("target-date must be YYYY-MM-DD")
		}
		password := os.Getenv("UAT_SYNTHETIC_CLOSING_PASSWORD")
		if password == "" {
			return fmt.Errorf("UAT_SYNTHETIC_CLOSING_PASSWORD is required")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), config.BcryptCost)
		if err != nil {
			return fmt.Errorf("hash login password")
		}
		result, err := billing.CreateSyntheticClosingFixture(ctx, db, billing.SyntheticClosingRequest{
			AppEnv:       appEnv,
			DBHost:       params.Host,
			TargetDate:   day,
			PasswordHash: string(hash),
		})
		if err != nil {
			return err
		}
		payload, err := json.Marshal(map[string]any{
			"clinicId":     result.ClinicID,
			"loginEmail":   result.LoginEmail,
			"billingIds":   result.BillingIDs,
			"completedAt":  result.CompletedAt,
			"cleanupToken": result.CleanupToken,
		})
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(os.Stdout, "%s\n", payload)
		return err
	case "teardown":
		if *clinicID == 0 {
			return fmt.Errorf("--clinic-id is required")
		}
		return billing.DeleteSyntheticClosingFixture(ctx, db, appEnv, params.Host, *clinicID, *cleanupToken)
	}
	return fmt.Errorf("unknown command %q", strings.TrimSpace(command))
}
