// Command stg-uat-skeleton applies the STG UAT clinic skeleton (clinics 1/2
// and F6 seed bindings) without writing any F6 cutover table, including staffs.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/animal-ekarte/backend/internal/csvimport"
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
	command string
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
	openDB  func(ctx context.Context, dsn string) (dbSession, error)
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
		openDB: func(ctx context.Context, dsn string) (dbSession, error) {
			conn, err := pgx.Connect(ctx, dsn)
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
	if opt.command == "apply" && !dbconn.IsLocalHost(conn.Host) {
		if os.Getenv(allowRemoteEnv) != allowRemoteSentinel {
			return fmt.Errorf("stg-uat-skeleton refuses non-local DB_HOST without %s=%s", allowRemoteEnv, allowRemoteSentinel)
		}
	}

	bootstrap, err := bootstrapFromEnv()
	if err != nil {
		return err
	}

	session, err := deps.openDB(ctx, conn.DSN(database))
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
	cutoverErr := verifyCutoverTablesEmpty(ctx, tx)
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	if logger != nil {
		if cutoverErr == nil {
			logger.Info("stg-uat-skeleton apply complete", "clinic_ids", []int64{clinicHachiojiID, clinicJoutoID})
		} else {
			logger.Info("stg-uat-skeleton bindings committed; cutover emptiness check failed",
				"clinic_ids", []int64{clinicHachiojiID, clinicJoutoID},
				"cutover_error", cutoverErr.Error(),
			)
		}
		for _, ids := range bindingByClinic {
			logSkeletonSeedIDs(logger, ids)
		}
	}
	if cutoverErr != nil {
		return cutoverErr
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
	if len(args) == 0 {
		return options{}, fmt.Errorf("usage: stg-uat-skeleton apply")
	}
	if args[0] != "apply" {
		return options{}, fmt.Errorf("command must be apply")
	}
	if len(args) > 1 {
		return options{}, fmt.Errorf("unexpected arguments")
	}
	return options{command: "apply"}, nil
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

func applySkeleton(ctx context.Context, e execer, bootstrap bootstrapOpts) error {
	if err := applySkeletonBindings(ctx, e, bootstrap); err != nil {
		return err
	}
	return verifyCutoverTablesEmpty(ctx, e)
}

func applySkeletonBindings(ctx context.Context, e execer, bootstrap bootstrapOpts) error {
	if err := assertAllowlistDisjointFromCutover(); err != nil {
		return err
	}
	if err := requireCompany(ctx, e, companyID); err != nil {
		return err
	}
	if err := ensureClinics(ctx, e); err != nil {
		return err
	}
	if err := ensureExamTypes(ctx, e); err != nil {
		return err
	}
	if err := ensureTrimmingReservationTypes(ctx, e); err != nil {
		return err
	}
	if err := ensureClinicSettings(ctx, e); err != nil {
		return err
	}
	if err := ensurePermissionGroups(ctx, e); err != nil {
		return err
	}
	if err := ensurePermissionGroupRules(ctx, e); err != nil {
		return err
	}
	if err := ensureBootstrapAccount(ctx, e, bootstrap); err != nil {
		return err
	}
	if err := ensureDefaultPaymentMethods(ctx, e); err != nil {
		return err
	}
	return verifySkeletonBindings(ctx, e)
}

func ensureClinics(ctx context.Context, e execer) error {
	for _, clinic := range clinicSeeds() {
		var exists bool
		if err := e.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM clinics WHERE id = $1)`, clinic.id).Scan(&exists); err != nil {
			return fmt.Errorf("lookup clinic %d: %w", clinic.id, err)
		}
		if exists {
			continue
		}
		if err := guardedExec(ctx, e, writeOp{
			table: "clinics",
			sql:   `INSERT INTO clinics (id, company_id, name, is_active) VALUES ($1, $2, $3, true)`,
			args:  []any{clinic.id, companyID, clinic.name},
		}); err != nil {
			return err
		}
	}
	return guardedExec(ctx, e, writeOp{
		table: "clinics",
		sql:   `SELECT setval(pg_get_serial_sequence('clinics', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM clinics), 1))`,
	})
}

func ensureExamTypes(ctx context.Context, e execer) error {
	for _, clinic := range clinicSeeds() {
		var n int64
		if err := e.QueryRow(ctx,
			`SELECT count(*) FROM exam_types WHERE clinic_id = $1 AND name = $2 AND deleted_at IS NULL`,
			clinic.id, examTypeName,
		).Scan(&n); err != nil {
			return fmt.Errorf("count clinic %d exam type: %w", clinic.id, err)
		}
		if n >= 1 {
			continue
		}
		if err := guardedExec(ctx, e, writeOp{
			table: "exam_types",
			sql:   `INSERT INTO exam_types (clinic_id, name, is_active) VALUES ($1, $2, true)`,
			args:  []any{clinic.id, examTypeName},
		}); err != nil {
			return err
		}
	}
	return nil
}

func ensureTrimmingReservationTypes(ctx context.Context, e execer) error {
	for _, clinic := range clinicSeeds() {
		var n int64
		if err := e.QueryRow(ctx,
			`SELECT count(*) FROM reservation_types WHERE clinic_id = $1 AND category = $2 AND deleted_at IS NULL`,
			clinic.id, trimmingCategory,
		).Scan(&n); err != nil {
			return fmt.Errorf("count clinic %d trimming reservation type: %w", clinic.id, err)
		}
		if n >= 1 {
			continue
		}
		if err := guardedExec(ctx, e, writeOp{
			table: "reservation_types",
			sql:   `INSERT INTO reservation_types (clinic_id, name, category, is_active, duration_minutes) VALUES ($1, $2, $3, true, 15)`,
			args:  []any{clinic.id, trimmingTypeName, trimmingCategory},
		}); err != nil {
			return err
		}
	}
	return nil
}

func ensureClinicSettings(ctx context.Context, e execer) error {
	for _, clinic := range clinicSeeds() {
		var n int64
		if err := e.QueryRow(ctx,
			`SELECT count(*) FROM clinic_settings WHERE clinic_id = $1`,
			clinic.id,
		).Scan(&n); err != nil {
			return fmt.Errorf("count clinic %d settings: %w", clinic.id, err)
		}
		if n >= 1 {
			continue
		}
		if err := guardedExec(ctx, e, writeOp{
			table: "clinic_settings",
			sql:   `INSERT INTO clinic_settings (clinic_id) VALUES ($1)`,
			args:  []any{clinic.id},
		}); err != nil {
			return err
		}
	}
	return nil
}

func ensurePermissionGroups(ctx context.Context, e execer) error {
	for _, group := range permissionGroupSeeds() {
		var exists bool
		if err := e.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM permission_groups WHERE id = $1)`, group.id).Scan(&exists); err != nil {
			return fmt.Errorf("lookup permission group %d: %w", group.id, err)
		}
		if exists {
			continue
		}
		if err := guardedExec(ctx, e, writeOp{
			table: "permission_groups",
			sql:   `INSERT INTO permission_groups (id, clinic_id, name, description, is_active, sort_order) VALUES ($1, $2, $3, $4, true, $5)`,
			args:  []any{group.id, group.clinicID, group.name, group.description, group.sortOrder},
		}); err != nil {
			return err
		}
	}
	return guardedExec(ctx, e, writeOp{
		table: "permission_groups",
		sql:   `SELECT setval(pg_get_serial_sequence('permission_groups', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM permission_groups), 1))`,
	})
}

func ensurePermissionGroupRules(ctx context.Context, e execer) error {
	for _, group := range permissionGroupSeeds() {
		for _, resource := range model.AllResources {
			var n int64
			if err := e.QueryRow(ctx,
				`SELECT count(*) FROM permission_group_rules WHERE group_id = $1 AND resource = $2`,
				group.id, string(resource),
			).Scan(&n); err != nil {
				return fmt.Errorf("count permission group %d rule %s: %w", group.id, resource, err)
			}
			if n >= 1 {
				continue
			}
			canView, canCreate, canEdit, canDelete := permissionBits(resource, group.executive)
			if err := guardedExec(ctx, e, writeOp{
				table: "permission_group_rules",
				sql:   `INSERT INTO permission_group_rules (group_id, resource, can_view, can_create, can_edit, can_delete) VALUES ($1, $2, $3, $4, $5, $6)`,
				args:  []any{group.id, string(resource), canView, canCreate, canEdit, canDelete},
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureBootstrapAccount(ctx context.Context, e execer, bootstrap bootstrapOpts) error {
	if (bootstrap.email == "") != (bootstrap.passwordHash == "") {
		return fmt.Errorf("bootstrap account requires both email and password hash")
	}
	if bootstrap.email == "" {
		return nil
	}
	var n int64
	if err := e.QueryRow(ctx,
		`SELECT count(*) FROM accounts WHERE email = $1`,
		bootstrap.email,
	).Scan(&n); err != nil {
		return fmt.Errorf("count bootstrap account: %w", err)
	}
	if n >= 1 {
		return nil
	}
	return guardedExec(ctx, e, writeOp{
		table: "accounts",
		sql:   `INSERT INTO accounts (email, password_hash, is_active, is_system_admin) VALUES ($1, $2, true, false)`,
		args:  []any{bootstrap.email, bootstrap.passwordHash},
	})
}

func clinicSeeds() []clinicSeed {
	return []clinicSeed{
		{id: clinicHachiojiID, name: clinicHachiojiName},
		{id: clinicJoutoID, name: clinicJoutoName},
	}
}

func permissionGroupSeeds() []permissionGroupSeed {
	return []permissionGroupSeed{
		{id: 1, clinicID: clinicHachiojiID, name: "執行", description: "執行権限", sortOrder: 1, executive: true},
		{id: 2, clinicID: clinicHachiojiID, name: "一般", description: "一般スタッフ権限", sortOrder: 2, executive: false},
		{id: 3, clinicID: clinicJoutoID, name: "執行", description: "執行権限", sortOrder: 1, executive: true},
		{id: 4, clinicID: clinicJoutoID, name: "一般", description: "一般スタッフ権限", sortOrder: 2, executive: false},
	}
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

func permissionBits(resource model.Resource, executive bool) (canView, canCreate, canEdit, canDelete bool) {
	for _, r := range skeletonPermissionRules {
		if r.resource != resource {
			continue
		}
		if executive {
			return r.execView, r.execCreate, r.execEdit, r.execDelete
		}
		return r.genView, r.genCreate, r.genEdit, r.genDelete
	}
	return false, false, false, false
}

func ensureDefaultPaymentMethods(ctx context.Context, e execer) error {
	// trg_create_default_payment_methods already inserts cash/credit_card on
	// clinic INSERT. Only fill F6-required keys when the trigger did not run.
	required := []struct {
		name       string
		systemKey  string
		displayOrd int
	}{
		{"現金", "cash", 1},
		{"クレジットカード", "credit_card", 2},
	}
	for _, clinic := range clinicSeeds() {
		for _, pm := range required {
			var n int64
			if err := e.QueryRow(ctx,
				`SELECT count(*) FROM payment_methods WHERE clinic_id = $1 AND system_key = $2 AND deleted_at IS NULL`,
				clinic.id, pm.systemKey,
			).Scan(&n); err != nil {
				return fmt.Errorf("count clinic %d payment method %s: %w", clinic.id, pm.systemKey, err)
			}
			if n >= 1 {
				continue
			}
			if err := guardedExec(ctx, e, writeOp{
				table: "payment_methods",
				sql:   `INSERT INTO payment_methods (clinic_id, name, system_key, display_order, is_active) VALUES ($1, $2, $3, $4, true)`,
				args:  []any{clinic.id, pm.name, pm.systemKey, pm.displayOrd},
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func requireCompany(ctx context.Context, e execer, id int64) error {
	var exists bool
	if err := e.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM companies WHERE id = $1)`, id).Scan(&exists); err != nil {
		return fmt.Errorf("lookup company %d: %w", id, err)
	}
	if !exists {
		return fmt.Errorf("company id=%d is required", id)
	}
	return nil
}

func verifySkeleton(ctx context.Context, e execer) error {
	if err := verifySkeletonBindings(ctx, e); err != nil {
		return err
	}
	return verifyCutoverTablesEmpty(ctx, e)
}

func verifySkeletonBindings(ctx context.Context, e execer) error {
	for _, clinic := range clinicSeeds() {
		var id, gotCompanyID int64
		var name string
		if err := e.QueryRow(ctx, `SELECT id, company_id, name FROM clinics WHERE id = $1`, clinic.id).
			Scan(&id, &gotCompanyID, &name); err != nil {
			return fmt.Errorf("clinic %d: %w", clinic.id, err)
		}
		if gotCompanyID != companyID || name != clinic.name {
			return fmt.Errorf("clinic %d identity mismatch", clinic.id)
		}
		if err := requireCountAtLeast(ctx, e,
			`SELECT count(*) FROM payment_methods WHERE clinic_id = $1 AND system_key = $2 AND deleted_at IS NULL`,
			[]any{clinic.id, "cash"},
			fmt.Sprintf("clinic %d cash payment method", clinic.id),
		); err != nil {
			return err
		}
		if err := requireCountAtLeast(ctx, e,
			`SELECT count(*) FROM payment_methods WHERE clinic_id = $1 AND system_key = $2 AND deleted_at IS NULL`,
			[]any{clinic.id, "credit_card"},
			fmt.Sprintf("clinic %d credit_card payment method", clinic.id),
		); err != nil {
			return err
		}
		if err := requireCountAtLeast(ctx, e,
			`SELECT count(*) FROM exam_types WHERE clinic_id = $1 AND name = $2 AND deleted_at IS NULL`,
			[]any{clinic.id, examTypeName},
			fmt.Sprintf("clinic %d exam type %s", clinic.id, examTypeName),
		); err != nil {
			return err
		}
		if err := requireCountAtLeast(ctx, e,
			`SELECT count(*) FROM reservation_types WHERE clinic_id = $1 AND category = $2 AND deleted_at IS NULL`,
			[]any{clinic.id, trimmingCategory},
			fmt.Sprintf("clinic %d trimming reservation type", clinic.id),
		); err != nil {
			return err
		}
		if err := requireCountAtLeast(ctx, e,
			`SELECT count(*) FROM clinic_settings WHERE clinic_id = $1`,
			[]any{clinic.id},
			fmt.Sprintf("clinic %d clinic_settings", clinic.id),
		); err != nil {
			return err
		}
	}

	for _, group := range permissionGroupSeeds() {
		var gotClinic int64
		var name string
		if err := e.QueryRow(ctx,
			`SELECT clinic_id, name FROM permission_groups WHERE id = $1`,
			group.id,
		).Scan(&gotClinic, &name); err != nil {
			return fmt.Errorf("permission group %d: %w", group.id, err)
		}
		if gotClinic != group.clinicID || name != group.name {
			return fmt.Errorf("permission group %d clinic/name mismatch", group.id)
		}
		if err := requireCountAtLeast(ctx, e,
			`SELECT count(*) FROM permission_group_rules WHERE group_id = $1`,
			[]any{group.id},
			fmt.Sprintf("permission group %d rules", group.id),
		); err != nil {
			return err
		}
	}
	return nil
}

func verifyCutoverTablesEmpty(ctx context.Context, e execer) error {
	for _, name := range cutoverTableNames() {
		ident, err := quoteIdent(name)
		if err != nil {
			return err
		}
		var n int64
		if err := e.QueryRow(ctx, `SELECT count(*) FROM `+ident).Scan(&n); err != nil {
			return fmt.Errorf("count %s: %w", name, err)
		}
		if n != 0 {
			return fmt.Errorf("cutover table %s must be empty after skeleton apply, got %d rows", name, n)
		}
	}

	var staffsInBand int64
	if err := e.QueryRow(ctx,
		`SELECT count(*) FROM staffs WHERE id >= $1 AND id < $2`,
		hachiojiBandStart, hachiojiBandEnd,
	).Scan(&staffsInBand); err != nil {
		return fmt.Errorf("count staffs in hachioji band: %w", err)
	}
	if staffsInBand != 0 {
		return fmt.Errorf("staffs count in band [%d,%d) must be 0, got %d", hachiojiBandStart, hachiojiBandEnd, staffsInBand)
	}
	return nil
}

func requireCountAtLeast(ctx context.Context, e execer, sql string, args []any, label string) error {
	var n int64
	if err := e.QueryRow(ctx, sql, args...).Scan(&n); err != nil {
		return fmt.Errorf("count %s: %w", label, err)
	}
	if n < 1 {
		return fmt.Errorf("%s is required", label)
	}
	return nil
}

func guardedExec(ctx context.Context, e execer, op writeOp) error {
	if err := rejectForbiddenWrite(op.table, op.sql); err != nil {
		return err
	}
	if !isAllowlisted(op.table) {
		return fmt.Errorf("stg-uat-skeleton refuses write to non-allowlisted table %s", op.table)
	}
	if err := e.Exec(ctx, op.sql, op.args...); err != nil {
		return fmt.Errorf("exec %s: %w", op.table, err)
	}
	return nil
}

func rejectForbiddenWrite(table, sql string) error {
	forbidden := cutoverTableSet()
	check := func(name string) error {
		name = strings.ToLower(strings.Trim(strings.TrimSpace(name), `"`))
		if name == "" {
			return nil
		}
		if _, ok := forbidden[name]; ok {
			return fmt.Errorf("stg-uat-skeleton refuses write to cutover table %s", name)
		}
		return nil
	}
	if err := check(table); err != nil {
		return err
	}
	if m := mutatingTableRe.FindStringSubmatch(sql); len(m) == 2 {
		if err := check(m[1]); err != nil {
			return err
		}
	}
	return nil
}

func assertAllowlistDisjointFromCutover() error {
	forbidden := cutoverTableSet()
	for _, name := range skeletonAllowlist {
		if _, ok := forbidden[name]; ok {
			return fmt.Errorf("skeleton allowlist includes cutover table %s", name)
		}
	}
	return nil
}

func isAllowlisted(table string) bool {
	table = strings.ToLower(strings.TrimSpace(table))
	for _, name := range skeletonAllowlist {
		if name == table {
			return true
		}
	}
	return false
}

func skeletonAllowlistTables() []string {
	return append([]string(nil), skeletonAllowlist...)
}

func cutoverTableNames() []string {
	specs := csvimport.CutoverTableSpecs()
	names := make([]string, 0, len(specs))
	for _, spec := range specs {
		names = append(names, spec.Name)
	}
	return names
}

func cutoverTableSet() map[string]struct{} {
	names := cutoverTableNames()
	set := make(map[string]struct{}, len(names))
	for _, name := range names {
		set[name] = struct{}{}
	}
	return set
}

func quoteIdent(name string) (string, error) {
	if !identRe.MatchString(name) {
		return "", fmt.Errorf("invalid identifier %q", name)
	}
	return `"` + name + `"`, nil
}

func (s *pgxSession) Begin(ctx context.Context) (dbTx, error) {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &pgxTx{tx: tx}, nil
}

func (s *pgxSession) Close() {
	if s == nil || s.conn == nil {
		return
	}
	_ = s.conn.Close(context.Background())
}

func (t *pgxTx) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := t.tx.Exec(ctx, sql, args...)
	return err
}

func (t *pgxTx) QueryRow(ctx context.Context, sql string, args ...any) rowScanner {
	return t.tx.QueryRow(ctx, sql, args...)
}

func (t *pgxTx) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *pgxTx) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}
