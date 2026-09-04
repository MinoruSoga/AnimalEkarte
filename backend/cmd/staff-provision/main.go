// Command staff-provision runs secret-managed staff batch preflight/apply.
//
// Real production apply is operator-only. This binary never logs names, emails,
// passwords, or file bodies — only digests and counts.
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

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/dbconn"
	"github.com/animal-ekarte/backend/internal/staff"
)

const commandTimeout = 10 * time.Minute

type options struct {
	command      string
	manifestPath string
	secretsPath  string
	repoRoot     string
}

type runDependencies struct {
	configureTimeZone func() error
	fromEnv           func() (dbconn.ConnParams, error)
	openDB            func(dsn string) (*gorm.DB, error)
	repoRoots         func(explicit string) ([]string, error)
	newProvisioner    func(db *gorm.DB, repoRoots []string) *staff.StaffProvisioner
}

func productionRunDependencies() runDependencies {
	return runDependencies{
		configureTimeZone: config.ConfigureTimeZone,
		fromEnv:           dbconn.FromEnv,
		openDB: func(dsn string) (*gorm.DB, error) {
			return gorm.Open(postgres.Open(dsn), &gorm.Config{
				// Never emit SQL bind values that may include emails/password hashes.
				Logger: gormlogger.Default.LogMode(gormlogger.Silent),
			})
		},
		repoRoots: defaultRepoRoots,
		newProvisioner: func(db *gorm.DB, repoRoots []string) *staff.StaffProvisioner {
			return staff.NewStaffProvisioner(staff.NewStaffProvisioningRepository(db), repoRoots)
		},
	}
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		// Keep operator logs free of request bodies / PII fields.
		Level: slog.LevelInfo,
	}))
	if err := run(context.Background(), os.Args[1:], logger, productionRunDependencies()); err != nil {
		// Log only the safe error surface; never dump input payloads.
		logger.Error("staff-provision failed", "error", sanitizeError(err))
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
	if opt.command == "apply" && !dbconn.IsLocalHost(conn.Host) {
		// Hard stop for non-local hosts unless the operator is intentionally on a
		// local compose network. Staging/production apply remains a separate,
		// human-reviewed procedure described in the ops doc — this binary still
		// refuses accidental remote apply from a developer laptop without an
		// explicit local host.
		//
		// Note: authorized operators targeting remote environments must use the
		// documented remote execution path (container in that environment) where
		// DB_HOST is the environment-local name, not a public endpoint from a
		// developer workstation.
		if os.Getenv("STAFF_PROVISION_ALLOW_REMOTE") != "YES_I_UNDERSTAND" {
			return fmt.Errorf("apply refuses non-local DB_HOST without STAFF_PROVISION_ALLOW_REMOTE=YES_I_UNDERSTAND")
		}
	}

	db, err := deps.openDB(conn.DSN(database))
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer closeGormDBQuietly(db)

	runCtx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()

	provisioner := deps.newProvisioner(db, repoRoots)
	switch opt.command {
	case "preflight":
		result, preflightErr := provisioner.Preflight(runCtx, opt.manifestPath, opt.secretsPath)
		if preflightErr != nil {
			return preflightErr
		}
		logger.Info("staff-provision preflight PASS",
			"batch_id", result.BatchID,
			"digest", result.Digest,
			"staff_count", result.StaffCount,
			"clinic_scope_count", len(result.ClinicScope),
		)
		return writeJSON(os.Stdout, result)
	case "apply":
		result, applyErr := provisioner.Apply(runCtx, opt.manifestPath, opt.secretsPath)
		if applyErr != nil {
			return applyErr
		}
		logger.Info("staff-provision apply complete",
			"status", result.Status,
			"batch_id", result.BatchID,
			"digest", result.Digest,
			"staff_count", result.StaffCount,
			"clinic_scope_count", len(result.ClinicScope),
		)
		return writeJSON(os.Stdout, result)
	default:
		return fmt.Errorf("unknown command %q", opt.command)
	}
}

func parseOptions(args []string) (options, error) {
	if len(args) == 0 {
		return options{}, fmt.Errorf("usage: staff-provision <preflight|apply> --manifest=/abs/path --secrets=/abs/path")
	}
	command := args[0]
	if command != "preflight" && command != "apply" {
		return options{}, fmt.Errorf("command must be preflight or apply")
	}
	fs := flag.NewFlagSet("staff-provision", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	manifestPath := fs.String("manifest", "", "absolute path to manifest JSON (mode 0600, outside repo)")
	secretsPath := fs.String("secrets", "", "absolute path to secrets JSON (mode 0600, outside repo)")
	repoRoot := fs.String("repo-root", "", "optional absolute repository root to exclude as input location")
	if err := fs.Parse(args[1:]); err != nil {
		return options{}, fmt.Errorf("parse flags: %w", err)
	}
	if strings.TrimSpace(*manifestPath) == "" || strings.TrimSpace(*secretsPath) == "" {
		return options{}, fmt.Errorf("--manifest and --secrets are required")
	}
	return options{
		command:      command,
		manifestPath: *manifestPath,
		secretsPath:  *secretsPath,
		repoRoot:     strings.TrimSpace(*repoRoot),
	}, nil
}

func defaultRepoRoots(explicit string) ([]string, error) {
	roots := make([]string, 0, 3)
	if explicit != "" {
		if !filepath.IsAbs(explicit) {
			return nil, fmt.Errorf("--repo-root must be absolute")
		}
		roots = append(roots, filepath.Clean(explicit))
	}
	// Prefer module root detection from working directory when present.
	if wd, err := os.Getwd(); err == nil {
		if moduleRoot := findGoModRoot(wd); moduleRoot != "" {
			roots = appendModuleRootAndParent(roots, moduleRoot)
		}
	}
	if envRoot := os.Getenv("STAFF_PROVISION_REPO_ROOT"); envRoot != "" {
		if !filepath.IsAbs(envRoot) {
			return nil, fmt.Errorf("STAFF_PROVISION_REPO_ROOT must be absolute")
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

// closeGormDBQuietly closes the underlying sql.DB when present. Incomplete test
// doubles (nil config) are skipped without panicking.
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
	// Prefer stable AppError messages which are already free of secret bodies.
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) && appErr.Message != "" {
		return appErr.Code + ": " + appErr.Message
	}
	msg := err.Error()
	// Defense in depth: strip common secret-bearing patterns if a lower layer
	// ever embeds them (should not happen for this command's own errors).
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "password") || strings.Contains(msg, "@") {
		return "staff provision failed (details redacted)"
	}
	return msg
}
