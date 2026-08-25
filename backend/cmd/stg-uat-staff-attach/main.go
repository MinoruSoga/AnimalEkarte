// Command stg-uat-staff-attach links synthetic UAT accounts onto existing staffs
// rows without inserting staffs. Operator output is digest/count/ids only.
package main

import (
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormlogger "gorm.io/gorm/logger"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/dbconn"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/staff"
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
	command     string
	rosterPath  string
	secretsPath string
	repoRoot    string
}

type runDependencies struct {
	configureTimeZone func() error
	fromEnv           func() (dbconn.ConnParams, error)
	openDB            func(dsn string) (*gorm.DB, error)
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
		openDB: func(dsn string) (*gorm.DB, error) {
			return gorm.Open(postgres.Open(dsn), &gorm.Config{
				Logger: gormlogger.Default.LogMode(gormlogger.Silent),
			})
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
	if opt.command == "apply" && !dbconn.IsLocalHost(conn.Host) {
		if os.Getenv(allowRemoteEnv) != allowRemoteSentinel {
			return fmt.Errorf("apply refuses non-local DB_HOST without %s=%s", allowRemoteEnv, allowRemoteSentinel)
		}
	}

	db, err := deps.openDB(conn.DSN(database))
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
	if err := fs.Parse(args[1:]); err != nil {
		return options{}, fmt.Errorf("parse flags: %w", err)
	}
	if strings.TrimSpace(*rosterPath) == "" || strings.TrimSpace(*secretsPath) == "" {
		return options{}, fmt.Errorf("--roster and --secrets are required")
	}
	return options{
		command:     command,
		rosterPath:  *rosterPath,
		secretsPath: *secretsPath,
		repoRoot:    strings.TrimSpace(*repoRoot),
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
	if wd, err := os.Getwd(); err == nil {
		if moduleRoot := findGoModRoot(wd); moduleRoot != "" {
			roots = append(roots, moduleRoot)
			parent := filepath.Dir(moduleRoot)
			if parent != moduleRoot && parent != string(filepath.Separator) {
				roots = append(roots, parent)
			}
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

func newStaffAttacher(repo attachRepository, repoRoots []string) *attacher {
	roots := make([]string, 0, len(repoRoots))
	for _, root := range repoRoots {
		if root == "" {
			continue
		}
		roots = append(roots, root)
	}
	return &attacher{repo: repo, repoRoots: roots}
}

func (a *attacher) Preflight(ctx context.Context, rosterPath, secretsPath string) (*attachResult, error) {
	if a == nil || a.repo == nil {
		return nil, apperrors.WrapInternalServerError("staff attacher is not configured")
	}
	inputs, err := a.loadInputs(rosterPath, secretsPath)
	if err != nil {
		return nil, err
	}
	if err := a.validateRefs(ctx, inputs.roster); err != nil {
		return nil, err
	}
	return newAttachResult("preflight", inputs), nil
}

func (a *attacher) Apply(ctx context.Context, rosterPath, secretsPath string) (*attachResult, error) {
	if a == nil || a.repo == nil {
		return nil, apperrors.WrapInternalServerError("staff attacher is not configured")
	}
	inputs, err := a.loadInputs(rosterPath, secretsPath)
	if err != nil {
		return nil, err
	}

	applied := 0
	if err := a.repo.WithTx(ctx, func(txCtx context.Context) error {
		applied = 0
		for _, entry := range inputs.roster.Staff {
			staffRow, findErr := a.repo.FindStaffByID(txCtx, entry.StaffID)
			if findErr != nil {
				return findErr
			}
			clinicID, clinicErr := resolveClinic(entry, staffRow)
			if clinicErr != nil {
				return clinicErr
			}
			if emailErr := validateStaffEmail(entry.StaffID, entry.Email); emailErr != nil {
				return emailErr
			}
			if groupErr := validatePermissionGroupIDs(entry.PermissionGroupIDs); groupErr != nil {
				return groupErr
			}
			if belongErr := a.repo.PermissionGroupsBelongToClinic(txCtx, clinicID, entry.PermissionGroupIDs); belongErr != nil {
				return belongErr
			}

			staffDigest := inputs.staffDigests[entry.StaffID]
			if staffRow.AccountID != nil {
				linked, linkErr := a.linkedAccountDecision(txCtx, staffRow, entry.StaffID, staffDigest)
				if linkErr != nil {
					return linkErr
				}
				if linked {
					continue
				}
			}

			password := inputs.secrets[entry.SecretRef]
			hashed, hashErr := bcrypt.GenerateFromPassword([]byte(password), config.BcryptCost)
			password = ""
			if hashErr != nil {
				return apperrors.Wrap(hashErr, "failed to hash credential")
			}

			account := &model.Account{
				Email:         expectedStaffEmail(entry.StaffID),
				PasswordHash:  string(hashed),
				IsActive:      true,
				IsSystemAdmin: false,
			}
			if createErr := a.repo.CreateAccount(txCtx, account); createErr != nil {
				return createErr
			}
			if updateErr := a.repo.UpdateStaffAccount(txCtx, staffRow.ID, clinicID, account.ID, entry.SetActive); updateErr != nil {
				return updateErr
			}
			if assignErr := a.repo.EnsureClinicAssignment(txCtx, staffRow.ID, clinicID); assignErr != nil {
				return assignErr
			}
			if groupErr := a.repo.AssignPermissionGroups(txCtx, clinicID, staffRow.ID, entry.PermissionGroupIDs); groupErr != nil {
				return groupErr
			}
			if saveErr := a.repo.SaveAttachDigest(txCtx, staffRow.ID, staffDigest); saveErr != nil {
				return saveErr
			}
			applied++
		}
		return nil
	}); err != nil {
		return nil, err
	}

	status := "applied"
	if applied == 0 {
		status = "noop"
	}
	return newAttachResult(status, inputs), nil
}

func (a *attacher) linkedAccountDecision(
	ctx context.Context,
	staffRow *model.Staff,
	staffID uint64,
	staffDigest string,
) (bool, error) {
	last, err := a.repo.LastAttachDigest(ctx, staffID)
	if err != nil {
		return false, err
	}
	if last == staffDigest {
		return true, nil
	}
	if last != "" {
		return false, apperrors.WrapConflict("staff is already linked with a different attach digest")
	}
	account, err := a.repo.FindAccountByID(ctx, *staffRow.AccountID)
	if err != nil {
		return false, err
	}
	if normalizeEmail(account.Email) == expectedStaffEmail(staffID) {
		if saveErr := a.repo.SaveAttachDigest(ctx, staffID, staffDigest); saveErr != nil {
			return false, saveErr
		}
		return true, nil
	}
	return false, apperrors.WrapConflict("staff is already linked to a different account")
}

func (a *attacher) validateRefs(ctx context.Context, roster *rosterFile) error {
	for _, entry := range roster.Staff {
		staffRow, err := a.repo.FindStaffByID(ctx, entry.StaffID)
		if err != nil {
			return err
		}
		clinicID, err := resolveClinic(entry, staffRow)
		if err != nil {
			return err
		}
		if err := validateStaffEmail(entry.StaffID, entry.Email); err != nil {
			return err
		}
		if err := validatePermissionGroupIDs(entry.PermissionGroupIDs); err != nil {
			return err
		}
		if err := a.repo.PermissionGroupsBelongToClinic(ctx, clinicID, entry.PermissionGroupIDs); err != nil {
			return err
		}
	}
	return nil
}

func (a *attacher) loadInputs(rosterPath, secretsPath string) (*loadedInputs, error) {
	rosterHandle, _, err := staff.OpenSecureInputFile(rosterPath, a.repoRoots)
	if err != nil {
		return nil, apperrors.Wrap(err, "roster input rejected")
	}
	defer func() { _ = rosterHandle.Close() }()

	secretsHandle, _, err := staff.OpenSecureInputFile(secretsPath, a.repoRoots)
	if err != nil {
		return nil, apperrors.Wrap(err, "secrets input rejected")
	}
	defer func() { _ = secretsHandle.Close() }()

	var roster rosterFile
	if err := decodeStrictJSON(rosterHandle, &roster, "roster"); err != nil {
		return nil, err
	}
	var secrets secretsFile
	if err := decodeStrictJSON(secretsHandle, &secrets, "secrets"); err != nil {
		return nil, err
	}
	if err := validateRoster(&roster); err != nil {
		return nil, err
	}
	secretMap, err := validateSecrets(&roster, &secrets)
	if err != nil {
		return nil, err
	}
	digest, staffDigests, err := computeAttachDigests(&roster)
	if err != nil {
		return nil, err
	}
	return &loadedInputs{
		roster:       &roster,
		secrets:      secretMap,
		digest:       digest,
		staffDigests: staffDigests,
	}, nil
}

func decodeStrictJSON(r io.Reader, dest any, label string) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return apperrors.WrapInvalidInput(fmt.Sprintf("%s decode failed: %v", label, err))
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return apperrors.WrapInvalidInput(fmt.Sprintf("%s must contain exactly one JSON value", label))
	}
	return nil
}

func validateRoster(roster *rosterFile) error {
	if roster == nil {
		return apperrors.WrapInvalidInput("roster is required")
	}
	if roster.SchemaVersion != rosterSchemaVersion {
		return apperrors.WrapInvalidInput("unsupported schema_version")
	}
	if len(roster.Staff) == 0 {
		return apperrors.WrapInvalidInput("staff must not be empty")
	}

	seenID := make(map[uint64]struct{}, len(roster.Staff))
	for i := range roster.Staff {
		entry := &roster.Staff[i]
		prefix := fmt.Sprintf("staff[%d]", i)
		if entry.StaffID == 0 {
			return apperrors.WrapInvalidInput(prefix + ": staff_id is required")
		}
		if _, dup := seenID[entry.StaffID]; dup {
			return apperrors.WrapInvalidInput("duplicate staff_id in roster")
		}
		seenID[entry.StaffID] = struct{}{}

		entry.Email = normalizeEmail(entry.Email)
		if err := validateStaffEmail(entry.StaffID, entry.Email); err != nil {
			return apperrors.Wrap(err, prefix)
		}
		entry.SecretRef = strings.TrimSpace(entry.SecretRef)
		if entry.SecretRef == "" {
			return apperrors.WrapInvalidInput(prefix + ": secret_ref is required")
		}
		if err := validatePermissionGroupIDs(entry.PermissionGroupIDs); err != nil {
			return apperrors.Wrap(err, prefix)
		}
		if entry.ClinicID == 0 && len(entry.ClinicIDs) == 0 {
			return apperrors.WrapInvalidInput(prefix + ": clinic_id or clinic_ids is required")
		}
		for _, clinicID := range entry.ClinicIDs {
			if clinicID == 0 {
				return apperrors.WrapInvalidInput(prefix + ": clinic_ids must not contain zero")
			}
		}
	}

	slices.SortFunc(roster.Staff, func(a, b rosterStaff) int {
		return cmp.Compare(a.StaffID, b.StaffID)
	})
	return nil
}

func validateSecrets(roster *rosterFile, secrets *secretsFile) (map[string]string, error) {
	if secrets == nil {
		return nil, apperrors.WrapInvalidInput("secrets are required")
	}
	if len(secrets.Secrets) == 0 {
		return nil, apperrors.WrapInvalidInput("secrets must not be empty")
	}
	required := make(map[string]struct{}, len(roster.Staff))
	for _, entry := range roster.Staff {
		required[entry.SecretRef] = struct{}{}
	}
	out := make(map[string]string, len(secrets.Secrets))
	for i, secret := range secrets.Secrets {
		ref := strings.TrimSpace(secret.SecretRef)
		if ref == "" {
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("secrets[%d]: secret_ref is required", i))
		}
		if _, exists := out[ref]; exists {
			return nil, apperrors.WrapInvalidInput("duplicate secret_ref in secrets file")
		}
		if _, ok := required[ref]; !ok {
			return nil, apperrors.WrapInvalidInput("secrets file contains unreferenced secret_ref")
		}
		if strings.TrimSpace(secret.Password) == "" {
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("secrets[%d]: password is required", i))
		}
		out[ref] = secret.Password
	}
	if len(out) != len(required) {
		return nil, apperrors.WrapInvalidInput("secrets file must cover every staff secret_ref")
	}
	return out, nil
}

func validatePermissionGroupIDs(ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("permission_group_ids must not be empty")
	}
	seen := make(map[uint64]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			return apperrors.WrapInvalidInput("permission_group_ids must not contain zero")
		}
		if _, dup := seen[id]; dup {
			return apperrors.WrapInvalidInput("permission_group_ids must not contain duplicates")
		}
		seen[id] = struct{}{}
	}
	return nil
}

func expectedStaffEmail(staffID uint64) string {
	return fmt.Sprintf("stg-staff-%d@example.test", staffID)
}

func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func validateStaffEmail(staffID uint64, email string) error {
	if normalizeEmail(email) != expectedStaffEmail(staffID) {
		return apperrors.WrapInvalidInput("email must match stg-staff-{id}@example.test")
	}
	return nil
}

func resolveClinic(entry rosterStaff, staffRow *model.Staff) (uint64, error) {
	if staffRow == nil {
		return 0, apperrors.WrapInvalidInput("staff is required")
	}
	if entry.ClinicID != 0 {
		if staffRow.ClinicID != entry.ClinicID {
			return 0, apperrors.WrapInvalidInput("staff clinic does not match roster clinic")
		}
		return entry.ClinicID, nil
	}
	if len(entry.ClinicIDs) == 0 {
		return 0, apperrors.WrapInvalidInput("clinic_id or clinic_ids is required")
	}
	if !slices.Contains(entry.ClinicIDs, staffRow.ClinicID) {
		return 0, apperrors.WrapInvalidInput("staff clinic is not in roster clinic_ids")
	}
	return staffRow.ClinicID, nil
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

func computeAttachDigests(roster *rosterFile) (string, map[uint64]string, error) {
	entries := make([]digestStaff, 0, len(roster.Staff))
	staffDigests := make(map[uint64]string, len(roster.Staff))
	for _, entry := range roster.Staff {
		clinicIDs := append([]uint64(nil), entry.ClinicIDs...)
		slices.Sort(clinicIDs)
		groupIDs := append([]uint64(nil), entry.PermissionGroupIDs...)
		slices.Sort(groupIDs)
		identitySum := sha256.Sum256([]byte(entry.Email))
		item := digestStaff{
			StaffID:             entry.StaffID,
			ClinicID:            entry.ClinicID,
			ClinicIDs:           clinicIDs,
			IdentityFingerprint: hex.EncodeToString(identitySum[:]),
			PermissionGroupIDs:  groupIDs,
			SetActive:           entry.SetActive,
			SecretRef:           entry.SecretRef,
		}
		staffPayload, err := json.Marshal(item)
		if err != nil {
			return "", nil, apperrors.Wrap(err, "failed to marshal staff attach digest")
		}
		staffSum := sha256.Sum256(staffPayload)
		staffDigests[entry.StaffID] = hex.EncodeToString(staffSum[:])
		entries = append(entries, item)
	}
	payload, err := json.Marshal(digestRoot{
		SchemaVersion: roster.SchemaVersion,
		Staff:         entries,
	})
	if err != nil {
		return "", nil, apperrors.Wrap(err, "failed to marshal attach digest")
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), staffDigests, nil
}

func newAttachResult(status string, inputs *loadedInputs) *attachResult {
	ids := make([]uint64, 0, len(inputs.roster.Staff))
	for _, entry := range inputs.roster.Staff {
		ids = append(ids, entry.StaffID)
	}
	return &attachResult{
		Status:     status,
		Digest:     inputs.digest,
		StaffCount: len(ids),
		StaffIDs:   ids,
	}
}

func (r *gormAttachRepository) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(persistence.WithTxValue(ctx, tx))
	}); err != nil {
		return apperrors.Wrap(err, "staff attach transaction failed")
	}
	return nil
}

func (r *gormAttachRepository) FindStaffByID(ctx context.Context, id uint64) (*model.Staff, error) {
	var row model.Staff
	if err := persistence.DBOrTx(ctx, r.db).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&row).Error; err != nil {
		return nil, apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", id))
	}
	return &row, nil
}

func (r *gormAttachRepository) FindAccountByID(ctx context.Context, accountID uint64) (*model.Account, error) {
	var account model.Account
	db := persistence.DBOrTx(ctx, r.db)
	if err := db.Session(&gorm.Session{Logger: db.Logger.LogMode(gormlogger.Silent)}).
		First(&account, "id = ? AND deleted_at IS NULL", accountID).Error; err != nil {
		return nil, apperrors.FromGORM(err, "account", fmt.Sprintf("%d", accountID))
	}
	return &account, nil
}

func (r *gormAttachRepository) PermissionGroupsBelongToClinic(
	ctx context.Context,
	clinicID uint64,
	groupIDs []uint64,
) error {
	if len(groupIDs) == 0 {
		return apperrors.WrapInvalidInput("permission_group_ids must not be empty")
	}
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.PermissionGroup{}).
		Where("clinic_id = ? AND deleted_at IS NULL AND id IN ?", clinicID, groupIDs).
		Count(&count).Error; err != nil {
		return apperrors.FromGORM(err, "permission_group", "")
	}
	if int(count) != len(groupIDs) {
		return apperrors.WrapInvalidInput("permission_group_ids contains invalid permission group")
	}
	return nil
}

func (r *gormAttachRepository) CreateAccount(ctx context.Context, account *model.Account) error {
	db := persistence.DBOrTx(ctx, r.db)
	if err := db.Session(&gorm.Session{Logger: db.Logger.LogMode(gormlogger.Silent)}).
		Create(account).Error; err != nil {
		return apperrors.FromGORM(err, "account", "")
	}
	return nil
}

func (r *gormAttachRepository) UpdateStaffAccount(
	ctx context.Context,
	staffID, clinicID, accountID uint64,
	setActive bool,
) error {
	db := persistence.DBOrTx(ctx, r.db)
	var row model.Staff
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", staffID, clinicID).
		First(&row).Error; err != nil {
		return apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", staffID))
	}
	if row.AccountID != nil {
		return apperrors.WrapConflict("staff is already linked")
	}
	updates := map[string]any{"account_id": accountID}
	if setActive {
		updates["is_active"] = true
	}
	result := db.Model(&model.Staff{}).
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL AND account_id IS NULL", staffID, clinicID).
		Updates(updates)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "staff", fmt.Sprintf("%d", staffID))
	}
	if result.RowsAffected != 1 {
		return apperrors.WrapConflict("staff account attach did not update exactly one row")
	}
	return nil
}

func (r *gormAttachRepository) EnsureClinicAssignment(ctx context.Context, staffID, clinicID uint64) error {
	db := persistence.DBOrTx(ctx, r.db)
	var count int64
	if err := db.Model(&model.StaffClinicAssignment{}).
		Where("staff_id = ? AND clinic_id = ? AND deleted_at IS NULL", staffID, clinicID).
		Count(&count).Error; err != nil {
		return apperrors.FromGORM(err, "staff_clinic_assignment", fmt.Sprintf("staff=%d,clinic=%d", staffID, clinicID))
	}
	if count > 0 {
		return nil
	}
	assignment := &model.StaffClinicAssignment{
		StaffID:  staffID,
		ClinicID: clinicID,
		IsMain:   true,
	}
	if err := db.Create(assignment).Error; err != nil {
		return apperrors.FromGORM(err, "staff_clinic_assignment", fmt.Sprintf("staff=%d,clinic=%d", staffID, clinicID))
	}
	return nil
}

func (r *gormAttachRepository) AssignPermissionGroups(
	ctx context.Context,
	clinicID, staffID uint64,
	groupIDs []uint64,
) error {
	if err := r.PermissionGroupsBelongToClinic(ctx, clinicID, groupIDs); err != nil {
		return err
	}
	db := persistence.DBOrTx(ctx, r.db)
	var assignment model.StaffClinicAssignment
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("staff_id = ? AND clinic_id = ? AND deleted_at IS NULL", staffID, clinicID).
		First(&assignment).Error; err != nil {
		return apperrors.FromGORM(
			err,
			"staff_clinic_assignment",
			fmt.Sprintf("staff=%d,clinic=%d", staffID, clinicID),
		)
	}
	rows := make([]model.StaffPermissionGroup, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		rows = append(rows, model.StaffPermissionGroup{StaffID: staffID, GroupID: groupID})
	}
	if err := db.Create(&rows).Error; err != nil {
		return apperrors.FromGORM(err, "staff_permission_group", fmt.Sprintf("staff:%d", staffID))
	}
	return nil
}

func (r *gormAttachRepository) LastAttachDigest(ctx context.Context, staffID uint64) (string, error) {
	var entry model.AuditLog
	err := persistence.DBOrTx(ctx, r.db).
		Where("resource = ? AND resource_id = ? AND action = ?", model.AuditResourceStaff, staffID, attachAuditAction).
		Order("id DESC").
		First(&entry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", apperrors.FromGORM(err, "audit_log", fmt.Sprintf("staff:%d", staffID))
	}
	return parseAttachDigestReceipt(entry.NewValue)
}

func (r *gormAttachRepository) SaveAttachDigest(ctx context.Context, staffID uint64, digest string) error {
	staffRow, err := r.FindStaffByID(ctx, staffID)
	if err != nil {
		return err
	}
	clinicID := staffRow.ClinicID
	payload, err := json.Marshal(attachDigestReceipt{Digest: digest})
	if err != nil {
		return apperrors.Wrap(err, "failed to marshal attach digest receipt")
	}
	entry := &model.AuditLog{
		ClinicID:   &clinicID,
		ActorType:  model.AuditActorTypeSystem,
		Action:     attachAuditAction,
		Resource:   model.AuditResourceStaff,
		ResourceID: &staffID,
		NewValue:   payload,
		UserAgent:  attachAuditUserAgent,
	}
	if err := persistence.DBOrTx(ctx, r.db).Create(entry).Error; err != nil {
		return apperrors.FromGORM(err, "audit_log", fmt.Sprintf("staff:%d", staffID))
	}
	return nil
}

func parseAttachDigestReceipt(raw json.RawMessage) (string, error) {
	if len(raw) == 0 {
		return "", nil
	}
	var receipt attachDigestReceipt
	if err := json.Unmarshal(raw, &receipt); err != nil {
		return "", apperrors.WrapInvalidInput("attach digest receipt is unreadable")
	}
	return receipt.Digest, nil
}
