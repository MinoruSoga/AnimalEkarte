// Command stg-uat-staff-attach links synthetic UAT accounts onto existing staffs
// rows without inserting staffs. Operator output is digest/count/ids only.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/dbconn"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	allowRemoteEnv       = "STG_UAT_STAFF_ATTACH_ALLOW_REMOTE"
	allowRemoteSentinel  = "YES_I_UNDERSTAND"
	commandTimeout       = 10 * time.Minute
	rosterSchemaVersion  = "stg-uat-staff-attach-v1"
	attachAuditAction    = "staff.uat_attach"
	attachAuditUserAgent = "stg-uat-staff-attach"
)

type options struct {
	command               string
	rosterPath            string
	secretsPath           string
	repoRoot              string
	confirmTargetHost     string
	confirmTargetDatabase string
}

type runDependencies struct {
	configureTimeZone func() error
	fromEnv           func() (dbconn.ConnParams, error)
	openDB            func(config *pgx.ConnConfig) (*gorm.DB, error)
	repoRoots         func(explicit string) ([]string, error)
	newAttacher       func(db *gorm.DB, repoRoots []string) *attacher
}

type attachResult struct {
	Status     string   `json:"status"`
	Digest     string   `json:"digest"`
	StaffCount int      `json:"staff_count"`
	StaffIDs   []uint64 `json:"staff_ids"`
}

type rosterFile struct {
	SchemaVersion string        `json:"schema_version"`
	Staff         []rosterStaff `json:"staff"`
}

type rosterStaff struct {
	StaffID            uint64   `json:"staff_id"`
	ClinicID           uint64   `json:"clinic_id"`
	ClinicIDs          []uint64 `json:"clinic_ids"`
	Email              string   `json:"email"`
	SecretRef          string   `json:"secret_ref"`
	PermissionGroupIDs []uint64 `json:"permission_group_ids"`
	SetActive          bool     `json:"set_active"`
}

type secretsFile struct {
	Secrets []secretEntry `json:"secrets"`
}

type secretEntry struct {
	SecretRef string `json:"secret_ref"`
	Password  string `json:"password"`
}

type loadedInputs struct {
	roster       *rosterFile
	secrets      map[string]string
	digest       string
	staffDigests map[uint64]string
}

type attachRepository interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
	FindStaffByID(ctx context.Context, id uint64) (*model.Staff, error)
	FindAccountByID(ctx context.Context, accountID uint64) (*model.Account, error)
	PermissionGroupsBelongToClinic(ctx context.Context, clinicID uint64, groupIDs []uint64) error
	CreateAccount(ctx context.Context, account *model.Account) error
	UpdateStaffAccount(ctx context.Context, staffID, clinicID, accountID uint64, setActive bool) error
	EnsureClinicAssignment(ctx context.Context, staffID, clinicID uint64) error
	AssignPermissionGroups(ctx context.Context, clinicID, staffID uint64, groupIDs []uint64) error
	LastAttachDigest(ctx context.Context, staffID uint64) (string, error)
	SaveAttachDigest(ctx context.Context, staffID uint64, digest string) error
}

type attacher struct {
	repo      attachRepository
	repoRoots []string
}

type gormAttachRepository struct {
	db *gorm.DB
}

func productionRunDependencies() runDependencies {
	return runDependencies{
		configureTimeZone: config.ConfigureTimeZone,
		fromEnv:           dbconn.FromEnv,
		openDB: func(config *pgx.ConnConfig) (*gorm.DB, error) {
			sqlDB := stdlib.OpenDB(*config)
			db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
				Logger: gormlogger.Default.LogMode(gormlogger.Silent),
			})
			if err != nil {
				_ = sqlDB.Close()
			}
			return db, err
		},
		repoRoots: defaultRepoRoots,
		newAttacher: func(db *gorm.DB, repoRoots []string) *attacher {
			return newStaffAttacher(&gormAttachRepository{db: db}, repoRoots)
		},
	}
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	if err := run(context.Background(), os.Args[1:], logger, productionRunDependencies()); err != nil {
		logger.Error("stg-uat-staff-attach failed", "error", sanitizeError(err))
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	args []string,
	logger *slog.Logger,
	deps runDependencies,
) error {
	if err := deps.configureTimeZone(); err != nil {
		return fmt.Errorf("timezone configuration failed: %w", err)
	}
	opt, err := parseOptions(args)
	if err != nil {
		return err
	}
	repoRoots, err := deps.repoRoots(opt.repoRoot)
	if err != nil {
		return err
	}

	conn, err := deps.fromEnv()
	if err != nil {
		return err
	}
	database := os.Getenv("DB_NAME")
	if database == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if err := requireStagingTarget(opt, conn.Host, database); err != nil {
		return err
	}
	if !dbconn.IsLocalHost(conn.Host) && os.Getenv(allowRemoteEnv) != allowRemoteSentinel {
		return fmt.Errorf("%s refuses non-local DB_HOST without %s=%s", opt.command, allowRemoteEnv, allowRemoteSentinel)
	}
	pgxConfig, err := conn.PGXConfig(database)
	if err != nil {
		return err
	}
	db, err := deps.openDB(pgxConfig)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer closeGormDBQuietly(db)

	runCtx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()

	att := deps.newAttacher(db, repoRoots)
	switch opt.command {
	case "preflight":
		result, preflightErr := att.Preflight(runCtx, opt.rosterPath, opt.secretsPath)
		if preflightErr != nil {
			return preflightErr
		}
		logger.Info("stg-uat-staff-attach preflight PASS",
			"digest", result.Digest,
			"staff_count", result.StaffCount,
		)
		return writeJSON(os.Stdout, result)
	case "apply":
		result, applyErr := att.Apply(runCtx, opt.rosterPath, opt.secretsPath)
		if applyErr != nil {
			return applyErr
		}
		logger.Info("stg-uat-staff-attach apply complete",
			"status", result.Status,
			"digest", result.Digest,
			"staff_count", result.StaffCount,
		)
		return writeJSON(os.Stdout, result)
	default:
		return fmt.Errorf("unknown command %q", opt.command)
	}
}

func parseOptions(args []string) (options, error) {
	if len(args) == 0 {
		return options{}, fmt.Errorf("usage: stg-uat-staff-attach <preflight|apply> --roster=/abs/path --secrets=/abs/path")
	}
	command := args[0]
	if command != "preflight" && command != "apply" {
		return options{}, fmt.Errorf("command must be preflight or apply")
	}
	fs := flag.NewFlagSet("stg-uat-staff-attach", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	rosterPath := fs.String("roster", "", "absolute path to roster JSON (mode 0600, outside repo)")
	secretsPath := fs.String("secrets", "", "absolute path to secrets JSON (mode 0600, outside repo)")
	repoRoot := fs.String("repo-root", "", "optional absolute repository root to exclude as input location")
	confirmTargetHost := fs.String("confirm-target-host", "", "must exactly equal DB_HOST")
	confirmTargetDatabase := fs.String("confirm-target-database", "", "must exactly equal DB_NAME")
	if err := fs.Parse(args[1:]); err != nil {
		return options{}, fmt.Errorf("parse flags: %w", err)
	}
	if strings.TrimSpace(*rosterPath) == "" || strings.TrimSpace(*secretsPath) == "" {
		return options{}, fmt.Errorf("--roster and --secrets are required")
	}
	return options{
		command:               command,
		rosterPath:            *rosterPath,
		secretsPath:           *secretsPath,
		repoRoot:              strings.TrimSpace(*repoRoot),
		confirmTargetHost:     *confirmTargetHost,
		confirmTargetDatabase: *confirmTargetDatabase,
	}, nil
}

func requireStagingTarget(opt options, host, database string) error {
	if strings.TrimSpace(os.Getenv("APP_ENV")) != "staging" {
		return fmt.Errorf("stg-uat-staff-attach requires APP_ENV=staging")
	}
	if opt.confirmTargetHost == "" || opt.confirmTargetHost != host {
		return fmt.Errorf("target host confirmation must exactly match DB_HOST")
	}
	if opt.confirmTargetDatabase == "" || opt.confirmTargetDatabase != database {
		return fmt.Errorf("target database confirmation must exactly match DB_NAME")
	}
	return nil
}

func defaultRepoRoots(explicit string) ([]string, error) {
	roots := make([]string, 0, 3)
	if explicit != "" {
		if !filepath.IsAbs(explicit) {
			return nil, fmt.Errorf("--repo-root must be absolute")
		}
		roots = append(roots, filepath.Clean(explicit))
	}
	if wd, err := os.Getwd(); err == nil {
		if moduleRoot := findGoModRoot(wd); moduleRoot != "" {
			roots = appendModuleRootAndParent(roots, moduleRoot)
		}
	}
	if envRoot := os.Getenv("STG_UAT_STAFF_ATTACH_REPO_ROOT"); envRoot != "" {
		if !filepath.IsAbs(envRoot) {
			return nil, fmt.Errorf("STG_UAT_STAFF_ATTACH_REPO_ROOT must be absolute")
		}
		roots = append(roots, filepath.Clean(envRoot))
	}
	return uniqueStrings(roots), nil
}

func appendModuleRootAndParent(roots []string, moduleRoot string) []string {
	roots = append(roots, moduleRoot)
	parent := filepath.Dir(moduleRoot)
	if parent != moduleRoot && parent != string(filepath.Separator) {
		roots = append(roots, parent)
	}
	return roots
}

func findGoModRoot(start string) string {
	dir := filepath.Clean(start)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, v := range in {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encode result: %w", err)
	}
	return nil
}

func closeGormDBQuietly(db *gorm.DB) {
	if db == nil || db.Config == nil {
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		return
	}
	_ = sqlDB.Close()
}

func sanitizeError(err error) string {
	if err == nil {
		return ""
	}
	var msg string
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) && appErr.Message != "" {
		msg = appErr.Code + ": " + appErr.Message
	} else {
		msg = err.Error()
	}
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "password") || strings.Contains(msg, "@") {
		return "staff attach failed (details redacted)"
	}
	return msg
}

type digestStaff struct {
	StaffID             uint64   `json:"staff_id"`
	ClinicID            uint64   `json:"clinic_id"`
	ClinicIDs           []uint64 `json:"clinic_ids,omitempty"`
	IdentityFingerprint string   `json:"identity_fingerprint"`
	PermissionGroupIDs  []uint64 `json:"permission_group_ids"`
	SetActive           bool     `json:"set_active"`
	SecretRef           string   `json:"secret_ref"`
}

type attachDigestReceipt struct {
	Digest string `json:"digest"`
}

type digestRoot struct {
	SchemaVersion string        `json:"schema_version"`
	Staff         []digestStaff `json:"staff"`
}
