// Command stg-uat-skeleton applies the STG UAT clinic skeleton (clinics 1/2
// and F6 seed bindings) without writing any F6 cutover table, including staffs.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/animal-ekarte/backend/internal/dbconn"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	allowRemoteEnv      = "STG_UAT_SKELETON_ALLOW_REMOTE"
	allowRemoteSentinel = "YES_I_UNDERSTAND"
	commandTimeout      = 2 * time.Minute

	companyID               int64 = 1
	clinicHachiojiID        int64 = 1
	clinicJoutoID           int64 = 2
	clinicHachiojiName            = "八王子病院"
	clinicJoutoName               = "城東センター病院"
	examTypeName                  = "検査"
	trimmingTypeName              = "トリミング"
	trimmingCategory              = "trimming"
	fallbackAnimalSpeciesID int64 = 1 // 002_master 犬; not created by skeleton
	hachiojiBandStart       int64 = 0
	hachiojiBandEnd         int64 = 10_000_000
)

// clinicBindingIDs are the six non-PHI numeric seed IDs required by
// cmd/csv-import-stg-uat. Names of owners/pets/staff are never logged.
type clinicBindingIDs struct {
	ClinicID                  int64
	AnimalSpeciesID           int64
	ExamTypeID                int64
	TrimmingReservationTypeID int64
	CashPaymentMethodID       int64
	CreditCardPaymentMethodID int64
}

var skeletonAllowlist = []string{
	"clinics",
	"payment_methods",
	"exam_types",
	"reservation_types",
	"permission_groups",
	"permission_group_rules",
	"clinic_settings",
	"accounts",
}

var mutatingTableRe = regexp.MustCompile(`(?is)\b(?:INSERT\s+INTO|COPY|UPDATE|DELETE\s+FROM|TRUNCATE(?:\s+TABLE)?)\s+(?:ONLY\s+)?(?:public\.)?"?([A-Za-z_][A-Za-z0-9_]*)"?`)

var identRe = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

type options struct {
	command               string
	confirmTargetHost     string
	confirmTargetDatabase string
}

type bootstrapOpts struct {
	email        string
	passwordHash string
}

type writeOp struct {
	table string
	sql   string
	args  []any
}

type clinicSeed struct {
	id   int64
	name string
}

type permissionGroupSeed struct {
	id          int64
	clinicID    int64
	name        string
	description string
	sortOrder   int
	executive   bool
}

type rowScanner interface {
	Scan(dest ...any) error
}

type execer interface {
	Exec(ctx context.Context, sql string, args ...any) error
	QueryRow(ctx context.Context, sql string, args ...any) rowScanner
}

type dbSession interface {
	Begin(ctx context.Context) (dbTx, error)
	Close()
}

type dbTx interface {
	execer
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type runDependencies struct {
	fromEnv func() (dbconn.ConnParams, error)
	openDB  func(ctx context.Context, config *pgx.ConnConfig) (dbSession, error)
}

type pgxSession struct {
	conn *pgx.Conn
}

type pgxTx struct {
	tx pgx.Tx
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	if err := run(ctx, os.Args[1:], logger, productionRunDependencies()); err != nil {
		logger.Error("stg-uat-skeleton failed", "error", err)
		os.Exit(1)
	}
}

func productionRunDependencies() runDependencies {
	return runDependencies{
		fromEnv: dbconn.FromEnv,
		openDB: func(ctx context.Context, config *pgx.ConnConfig) (dbSession, error) {
			conn, err := pgx.ConnectConfig(ctx, config)
			if err != nil {
				return nil, err
			}
			return &pgxSession{conn: conn}, nil
		},
	}
}

func run(ctx context.Context, args []string, logger *slog.Logger, deps runDependencies) error {
	opt, err := parseOptions(args)
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
		return fmt.Errorf("stg-uat-skeleton refuses non-local DB_HOST without %s=%s", allowRemoteEnv, allowRemoteSentinel)
	}

	bootstrap, err := bootstrapFromEnv()
	if err != nil {
		return err
	}

	pgxConfig, err := conn.PGXConfig(database)
	if err != nil {
		return err
	}
	session, err := deps.openDB(ctx, pgxConfig)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer session.Close()

	tx, err := session.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Binding ensure first. Cutover emptiness is checked after seed IDs are
	// resolved so a dirty local demo DB still prints the six non-PHI IDs.
	if err := applySkeletonBindings(ctx, tx, bootstrap); err != nil {
		return err
	}
	bindingByClinic := make([]clinicBindingIDs, 0, len(clinicSeeds()))
	for _, clinic := range clinicSeeds() {
		ids, err := resolveClinicBindingIDs(ctx, tx, clinic.id)
		if err != nil {
			return err
		}
		bindingByClinic = append(bindingByClinic, ids)
	}
	if err := commitSkeletonIfPreconditionsPass(ctx, tx); err != nil {
		return err
	}
	if logger != nil {
		logger.Info("stg-uat-skeleton apply complete", "clinic_ids", []int64{clinicHachiojiID, clinicJoutoID})
		for _, ids := range bindingByClinic {
			logSkeletonSeedIDs(logger, ids)
		}
	}
	return nil
}

func commitSkeletonIfPreconditionsPass(ctx context.Context, tx dbTx) error {
	if err := verifyCutoverTablesEmpty(ctx, tx); err != nil {
		if rollbackErr := tx.Rollback(context.WithoutCancel(ctx)); rollbackErr != nil {
			return fmt.Errorf("%w; rollback failed: %w", err, rollbackErr)
		}
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func resolveClinicBindingIDs(ctx context.Context, e execer, clinicID int64) (clinicBindingIDs, error) {
	ids := clinicBindingIDs{
		ClinicID:        clinicID,
		AnimalSpeciesID: fallbackAnimalSpeciesID,
	}
	var err error
	if ids.ExamTypeID, err = lookupID(ctx, e,
		`SELECT id FROM exam_types WHERE clinic_id = $1 AND name = $2 AND deleted_at IS NULL ORDER BY id LIMIT 1`,
		[]any{clinicID, examTypeName},
		fmt.Sprintf("clinic %d exam type %s", clinicID, examTypeName),
	); err != nil {
		return clinicBindingIDs{}, err
	}
	if ids.TrimmingReservationTypeID, err = lookupID(ctx, e,
		`SELECT id FROM reservation_types WHERE clinic_id = $1 AND category = $2 AND deleted_at IS NULL ORDER BY id LIMIT 1`,
		[]any{clinicID, trimmingCategory},
		fmt.Sprintf("clinic %d trimming reservation type", clinicID),
	); err != nil {
		return clinicBindingIDs{}, err
	}
	if ids.CashPaymentMethodID, err = lookupID(ctx, e,
		`SELECT id FROM payment_methods WHERE clinic_id = $1 AND system_key = $2 AND deleted_at IS NULL ORDER BY id LIMIT 1`,
		[]any{clinicID, "cash"},
		fmt.Sprintf("clinic %d cash payment method", clinicID),
	); err != nil {
		return clinicBindingIDs{}, err
	}
	if ids.CreditCardPaymentMethodID, err = lookupID(ctx, e,
		`SELECT id FROM payment_methods WHERE clinic_id = $1 AND system_key = $2 AND deleted_at IS NULL ORDER BY id LIMIT 1`,
		[]any{clinicID, "credit_card"},
		fmt.Sprintf("clinic %d credit_card payment method", clinicID),
	); err != nil {
		return clinicBindingIDs{}, err
	}
	return ids, nil
}

func lookupID(ctx context.Context, e execer, sql string, args []any, label string) (int64, error) {
	var id int64
	if err := e.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		return 0, fmt.Errorf("lookup %s: %w", label, err)
	}
	if id <= 0 {
		return 0, fmt.Errorf("%s id must be positive", label)
	}
	return id, nil
}

func logSkeletonSeedIDs(logger *slog.Logger, ids clinicBindingIDs) {
	if logger == nil {
		return
	}
	logger.Info("stg-uat-skeleton seed ids",
		"clinic_id", ids.ClinicID,
		"animal_species_id", ids.AnimalSpeciesID,
		"exam_type_id", ids.ExamTypeID,
		"trimming_reservation_type_id", ids.TrimmingReservationTypeID,
		"cash_payment_method_id", ids.CashPaymentMethodID,
		"credit_card_payment_method_id", ids.CreditCardPaymentMethodID,
	)
}

func parseOptions(args []string) (options, error) {
	if len(args) == 0 || args[0] != "apply" {
		return options{}, fmt.Errorf("usage: stg-uat-skeleton apply --confirm-target-host=HOST --confirm-target-database=DATABASE")
	}
	var opt options
	opt.command = args[0]
	fs := flag.NewFlagSet("stg-uat-skeleton apply", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&opt.confirmTargetHost, "confirm-target-host", "", "must exactly equal DB_HOST")
	fs.StringVar(&opt.confirmTargetDatabase, "confirm-target-database", "", "must exactly equal DB_NAME")
	if err := fs.Parse(args[1:]); err != nil {
		return options{}, err
	}
	if fs.NArg() != 0 {
		return options{}, fmt.Errorf("unexpected arguments")
	}
	return opt, nil
}

func requireStagingTarget(opt options, host, database string) error {
	if strings.TrimSpace(os.Getenv("APP_ENV")) != "staging" {
		return fmt.Errorf("stg-uat-skeleton requires APP_ENV=staging")
	}
	if opt.confirmTargetHost == "" || opt.confirmTargetHost != host {
		return fmt.Errorf("target host confirmation must exactly match DB_HOST")
	}
	if opt.confirmTargetDatabase == "" || opt.confirmTargetDatabase != database {
		return fmt.Errorf("target database confirmation must exactly match DB_NAME")
	}
	return nil
}

func bootstrapFromEnv() (bootstrapOpts, error) {
	email := strings.TrimSpace(os.Getenv("STG_UAT_SKELETON_BOOTSTRAP_EMAIL"))
	hash := os.Getenv("STG_UAT_SKELETON_BOOTSTRAP_PASSWORD_HASH")
	if email == "" && hash == "" {
		return bootstrapOpts{}, nil
	}
	if email == "" || hash == "" {
		return bootstrapOpts{}, fmt.Errorf("bootstrap account requires both STG_UAT_SKELETON_BOOTSTRAP_EMAIL and STG_UAT_SKELETON_BOOTSTRAP_PASSWORD_HASH")
	}
	return bootstrapOpts{email: email, passwordHash: hash}, nil
}

// skeletonPermissionRule mirrors clinic.defaultPermissionRuleTable (CreateClinic
// attach-receiver bits). Copied because that table is unexported; unknown
// resources stay deny-all.
type skeletonPermissionRule struct {
	resource                                   model.Resource
	execView, execCreate, execEdit, execDelete bool
	genView, genCreate, genEdit, genDelete     bool
}

var skeletonPermissionRules = []skeletonPermissionRule{
	{model.ResourceReception, true, true, true, true, true, false, false, false},
	{model.ResourceOwners, true, true, true, true, true, true, true, false},
	{model.ResourceReservations, true, true, true, true, true, true, true, false},
	{model.ResourceMedicalRecords, true, true, true, true, true, true, true, false},
	{model.ResourceHospitalization, true, true, true, true, true, true, true, false},
	{model.ResourceTrimming, true, true, true, true, true, true, true, false},
	{model.ResourceExaminations, true, true, true, true, true, true, true, false},
	{model.ResourceExaminationUnconfirm, false, false, false, false, false, false, false, false},
	{model.ResourceAccounting, true, true, true, true, true, false, false, false},
	{model.ResourceVaccinations, true, true, true, true, true, true, true, false},
	{model.ResourceCheckups, true, true, true, true, true, false, false, false},
	{model.ResourceInventory, true, true, true, true, true, false, false, false},
	{model.ResourceEstimates, true, true, true, true, true, false, false, false},
	{model.ResourceShifts, true, true, true, true, true, true, true, false},
	{model.ResourceHospitalSettings, true, false, true, false, true, false, false, false},
	{model.ResourceMasterAnimalSpecies, true, false, false, false, true, false, false, false},
	{model.ResourceMasterMedical, true, true, true, true, true, false, false, false},
	{model.ResourceMasterReservationType, true, true, true, true, true, false, false, false},
	{model.ResourceMasterHospitalization, true, true, true, true, true, false, false, false},
	{model.ResourceMasterTrimming, true, true, true, true, true, false, false, false},
	{model.ResourceMasterPermission, true, true, true, true, false, false, false, false},
	{model.ResourceMasterStaff, true, true, true, true, true, false, false, false},
	{model.ResourceMasterInsurance, true, true, true, true, true, false, false, false},
	{model.ResourceMasterMerchandise, true, true, true, true, true, false, false, false},
	{model.ResourceDiscount, true, true, true, true, false, false, false, false},
	{model.ResourceAccountingCancel, true, false, true, false, true, false, false, false},
	{model.ResourceAccountingPostCloseEdit, true, false, true, false, true, false, false, false},
	{model.ResourceCashRegisterClose, true, false, true, false, true, false, false, false},
	{model.ResourceAccountingReports, true, false, true, false, true, false, false, false},
	{model.ResourceClosingSettings, true, true, true, true, true, false, false, false},
	{model.ResourcePaymentMethod, true, false, true, false, true, false, false, false},
	{model.ResourceLstepCsvImport, true, false, true, false, true, false, false, false},
	{model.ResourceLstepAnalytics, true, false, true, false, true, false, false, false},
	{model.ResourceManualEdit, true, false, true, false, true, false, false, false},
	{model.ResourceLabImport, true, false, true, false, true, false, false, false},
	{model.ResourceIdentityLinks, false, false, false, false, false, false, false, false},
	{model.ResourceCheckupPackageImport, false, false, false, false, false, false, false, false},
}
