package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// テストDB接続・ENUM型・共有ベースモデルの AutoMigrate はプロセス全体（package repository の
// go test 実行単位）で一度だけ実行する（sharedTestSchemaOnce/TestMain）。130+ ファイル・158+ 箇所の
// setupTestDB(t) 呼び出し毎にこれらを繰り返すと、接続確立×2・ENUM存在チェック46回・AutoMigrate
// スキーマ内省クエリが呼び出し回数分積み上がり、repository テストスイート全体の支配的コストになる
// （2026-07 計測: ローカル 191s → 本最適化後は setupTestDB 呼び出し側を一切変更せず短縮）。
// TRUNCATE のみ setupTestDB 内で呼び出し毎に実行し、テスト間データ分離を維持する。
var (
	sharedTestDB         *gorm.DB
	sharedTestDBOnce     sync.Once
	sharedTestDBErr      error
	sharedTestSchemaOnce sync.Once
	sharedTestSchemaErr  error
)

// TestMain は internal/repository package の全テストで共有する DB 接続プールを管理する。
// 個々のテストは接続を閉じず（sharedTestDBOnce で一度だけ確立・全テストで再利用）、プロセス終了時に
// ここで一度だけ閉じる。
func TestMain(m *testing.M) {
	code := m.Run()
	if sharedTestDB != nil {
		if sqlDB, err := sharedTestDB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
	os.Exit(code)
}

// setupTestDB はテスト用の DB を返し、共有ベーステーブルを TRUNCATE してクリーンな状態にします。
// DB接続確立・ENUM型作成・ベースモデルの AutoMigrate はプロセス全体で一度だけ実行されます。
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := getTestDatabaseConnection(t)

	sharedTestSchemaOnce.Do(func() {
		sharedTestSchemaErr = setupSharedTestSchema(db)
	})
	if sharedTestSchemaErr != nil {
		t.Fatalf("failed to set up shared test schema: %v", sharedTestSchemaErr)
	}

	// Truncate tables to ensure clean state (data isolation between tests)
	db.Exec("TRUNCATE TABLE billing_refunds CASCADE")
	db.Exec("TRUNCATE TABLE payments CASCADE")
	db.Exec("TRUNCATE TABLE billings CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE owners CASCADE")

	return db
}

// makeTestOwner はテスト用の Owner を作成して返す。
func makeTestOwner(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Owner {
	t.Helper()
	o := &model.Owner{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(o).Error)
	return o
}

// setupIsolatedTestDB は setupTestDB と異なり、プロセス全体で共有しない「呼び出し毎に完全に新しい」
// DB 接続を返す（最適化前の setupTestDB と同じ挙動）。
//
// checkup_field_results/checkup_type_fields は checkup_field_repository_test.go（AutoMigrate 由来
// スキーマ）・checkup_field_cascade_test.go（migration 010 の実 DDL）・
// checkup_field_composite_fk_test.go（010 実 DDL + migration 012 複合 FK）という 3 種の異なる
// ヘルパーが同じテーブルを意図的に毎回 DROP+CREATE し合う（migration drift 検出が目的で、
// 挙動として必須）。この cluster に setupTestDB の共有コネクションプールを使うと、いずれかの
// ヘルパーが DROP TABLE/DROP TYPE した瞬間に、別テストが既に保持していた同一物理コネクション上の
// キャッシュ済み prepared statement（古いテーブル/型 OID 参照）が
// "cache lookup failed"（SQLSTATE XX000）で壊れる。3 ヘルパーの意図的な毎回 DROP+CREATE は
// 統合できないため、この cluster だけは共有プールから外し、テスト毎に使い捨ての新規コネクションを
// 割り当てることでキャッシュ汚染を根本的に回避する（対象は少数のためスループット影響は軽微）。
func setupIsolatedTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := connectTestDatabase()
	if err != nil {
		t.Fatalf("failed to connect to isolated test db: %v", err)
	}
	if sqlDB, derr := db.DB(); derr == nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}

	if err := setupSharedTestSchema(db); err != nil {
		t.Fatalf("failed to set up shared test schema (isolated): %v", err)
	}

	db.Exec("TRUNCATE TABLE billing_refunds CASCADE")
	db.Exec("TRUNCATE TABLE payments CASCADE")
	db.Exec("TRUNCATE TABLE billings CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE owners CASCADE")

	return db
}

// testSchemaEnumType is one hand-maintained ENUM type double used by setupSharedTestSchema.
// Kept as a package-level type/var (rather than a func-local literal) so
// test_schema_enum_parity_test.go (G12-2) can compare it against 001_init.sql directly instead
// of re-parsing Go source via go/ast.
type testSchemaEnumType struct {
	name   string
	create string
}

// sharedTestSchemaEnumTypes hand-duplicates every PostgreSQL ENUM type from 001_init.sql
// (54 types total, 2026-07-04 consolidated migration + 009 #201 薬量計算 4 型を含む）。
// model.Medicine が calculation_type を持つため、本 setup を使う全テストの medicines
// AutoMigrate に medicine_calculation_type が必須（欠落で CREATE TABLE 失敗）。
//
// G12-2 (BE-refactor.md): this list previously drifted from 001_init.sql — item_source was
// missing 'trimming' (blocking any integration test that persists a trimming billing_items
// row via billing_item_repository.go's Source: model.ItemSourceTrimming) and 4 whole types
// were absent. test_schema_enum_parity_test.go now gates this list against 001_init.sql on
// every `go test ./internal/repository/...` run so it cannot silently drift again.
//
// checkup_field_type is included here even though checkup_field_repository_test.go (and its
// sibling _cascade_test.go/_composite_fk_test.go) also DROP+CREATE it: those three helpers run
// on setupIsolatedTestDB (a throw-away connection per call, see repository/CLAUDE.md), and no
// setupTestDB-based (shared-connection) test AutoMigrates CheckupTypeField/CheckupFieldResult,
// so there is no column depending on this type via the shared connection. Package tests never
// run in parallel (no t.Parallel() in this package), so the isolated helpers' own DROP+CREATE
// cannot race this one either — adding it here is a no-op once created, not a collision.
var sharedTestSchemaEnumTypes = []testSchemaEnumType{
	// ペット関連
	{"pet_status", "CREATE TYPE pet_status AS ENUM ('alive', 'deceased')"},
	{"pet_gender", "CREATE TYPE pet_gender AS ENUM ('male', 'female', 'unknown')"},
	{"acquisition_type", "CREATE TYPE acquisition_type AS ENUM ('purchased', 'transferred', 'rescued', 'other')"},
	{"danger_level", "CREATE TYPE danger_level AS ENUM ('low', 'medium', 'high')"},
	{"membership_type", "CREATE TYPE membership_type AS ENUM ('non_member', 'member', 'deceased', 'transferred')"},
	// マスタ共通
	{"inventory_category", "CREATE TYPE inventory_category AS ENUM ('medicine', 'consumable', 'food', 'other')"},
	{"inventory_status", "CREATE TYPE inventory_status AS ENUM ('sufficient', 'low', 'out_of_stock')"},
	{"dosage_form", "CREATE TYPE dosage_form AS ENUM ('tablet', 'liquid', 'injection', 'topical', 'powder')"},
	{"medicine_unit", "CREATE TYPE medicine_unit AS ENUM ('per_tablet', 'per_ml', 'per_dose', 'per_gram')"},
	// #201 薬量自動計算（migration 009）: medicines.calculation_type + medicine_dose_params 用
	{"medicine_calculation_type", "CREATE TYPE medicine_calculation_type AS ENUM ('none', 'per_weight')"},
	{"medicine_dose_basis", "CREATE TYPE medicine_dose_basis AS ENUM ('per_administration', 'per_day')"},
	{"medicine_rounding_mode", "CREATE TYPE medicine_rounding_mode AS ENUM ('up', 'down', 'nearest')"},
	{"medicine_dose_species", "CREATE TYPE medicine_dose_species AS ENUM ('dog', 'cat')"},
	{"cage_type", "CREATE TYPE cage_type AS ENUM ('icu', 'dog', 'cat', 'general')"},
	{"cage_size", "CREATE TYPE cage_size AS ENUM ('small', 'medium', 'large')"},
	{"body_size", "CREATE TYPE body_size AS ENUM ('small', 'medium', 'large')"},
	{"billing_unit", "CREATE TYPE billing_unit AS ENUM ('per_day', 'per_night')"},
	{"target_size", "CREATE TYPE target_size AS ENUM ('small', 'medium', 'large', 'cat')"},
	{"anesthesia_type", "CREATE TYPE anesthesia_type AS ENUM ('none', 'local', 'sedation', 'general')"},
	{"vaccine_species", "CREATE TYPE vaccine_species AS ENUM ('dog', 'cat', 'both')"},
	// 電子カルテ関連
	{"medical_record_status", "CREATE TYPE medical_record_status AS ENUM ('draft', 'finalized')"},
	{"treatment_item_type", "CREATE TYPE treatment_item_type AS ENUM ('consultation', 'procedure', 'medicine', 'other')"},
	{"treatment_status", "CREATE TYPE treatment_status AS ENUM ('pending', 'completed', 'not_applicable')"},
	{"exam_status", "CREATE TYPE exam_status AS ENUM ('pending', 'in_progress', 'result_entered', 'completed', 'confirmed')"},
	{"exam_result_status", "CREATE TYPE exam_result_status AS ENUM ('normal', 'high', 'low')"},
	{"next_schedule_type", "CREATE TYPE next_schedule_type AS ENUM ('3weeks', '4weeks', '1year', 'other')"},
	{"appetite_level", "CREATE TYPE appetite_level AS ENUM ('normal', 'increased', 'decreased', 'none')"},
	{"water_intake_level", "CREATE TYPE water_intake_level AS ENUM ('normal', 'increased', 'decreased', 'none')"},
	{"medical_image_type", "CREATE TYPE medical_image_type AS ENUM ('xray', 'echo', 'photo', 'endoscope', 'ct', 'mri', 'microscope', 'other')"},
	{"estimate_status", "CREATE TYPE estimate_status AS ENUM ('draft', 'sent', 'approved', 'rejected')"},
	{"confirmation_status", "CREATE TYPE confirmation_status AS ENUM ('pending', 'confirmed', 'returned')"},
	{"item_category", "CREATE TYPE item_category AS ENUM ('examination', 'test', 'procedure', 'surgery', 'medicine', 'food', 'goods', 'other', 'vaccine', 'trimming', 'hotel', 'training')"},
	// G12-2: 'trimming' was missing — billing_item_repository.go:271 persists
	// Source: model.ItemSourceTrimming, so its integration path was untestable under this schema.
	{"item_source", "CREATE TYPE item_source AS ENUM ('medical_record', 'manual', 'hospitalization', 'trimming')"},
	{"campaign_discount_type", "CREATE TYPE campaign_discount_type AS ENUM ('rate', 'amount')"},
	// 予約・会計・入院関連
	{"visit_type", "CREATE TYPE visit_type AS ENUM ('first', 'revisit')"},
	{"reservation_status", "CREATE TYPE reservation_status AS ENUM ('confirmed', 'pending', 'cancelled', 'checked_in', 'in_consultation', 'accounting', 'completed', 'no_show')"},
	{"staff_type", "CREATE TYPE staff_type AS ENUM ('doctor', 'nurse', 'trimmer', 'resource')"},
	{"reservation_source", "CREATE TYPE reservation_source AS ENUM ('manual', 'line')"},
	{"billing_status", "CREATE TYPE billing_status AS ENUM ('waiting', 'completed', 'cancelled', 'pending')"},
	{"hospitalization_type", "CREATE TYPE hospitalization_type AS ENUM ('hospitalization', 'hotel')"},
	{"hospitalization_status", "CREATE TYPE hospitalization_status AS ENUM ('admitted', 'discharged', 'reserved')"},
	{"care_plan_type", "CREATE TYPE care_plan_type AS ENUM ('food', 'medicine', 'treatment', 'instruction', 'item')"},
	{"care_plan_status", "CREATE TYPE care_plan_status AS ENUM ('active', 'completed', 'discontinued')"},
	{"care_log_type", "CREATE TYPE care_log_type AS ENUM ('food', 'excretion', 'medicine', 'treatment', 'other')"},
	{"care_log_status", "CREATE TYPE care_log_status AS ENUM ('completed', 'partial', 'skipped')"},
	{"plan_timing", "CREATE TYPE plan_timing AS ENUM ('morning', 'noon', 'night')"},
	{"body_weight_unit", "CREATE TYPE body_weight_unit AS ENUM ('Kg', 'g')"},
	// トリミング・シフト関連
	{"reservation_type_category", "CREATE TYPE reservation_type_category AS ENUM ('general', 'trimming')"},
	{"payment_method", "CREATE TYPE payment_method AS ENUM ('cash', 'credit_card', 'electronic_money', 'bank_transfer')"},
	{"shift_type", "CREATE TYPE shift_type AS ENUM ('full', 'morning', 'afternoon', 'off', 'paid_leave')"},
	{"tax_type", "CREATE TYPE tax_type AS ENUM ('included', 'excluded', 'exempt')"},
	// lab_import（検査結果取込ジョブ）関連
	{"lab_import_job_status", "CREATE TYPE lab_import_job_status AS ENUM ('received', 'validated', 'mapped', 'persisted', 'duplicate', 'needs_review', 'failed')"},
	{"lab_import_source_type", "CREATE TYPE lab_import_source_type AS ENUM ('fixture', 'drwan', 'manual')"},
	// #211 健診パッケージ（migration 010 → 001_init.sql 統合済み）
	{"checkup_field_type", "CREATE TYPE checkup_field_type AS ENUM ('number', 'single_select', 'multi_select', 'boolean', 'checklist', 'text')"},
}

// enumValueRe extracts the ordered list of quoted ENUM value literals out of a
// "CREATE TYPE ... AS ENUM ('a', 'b', ...)" definition string.
var enumValueRe = regexp.MustCompile(`'[^']*'`)

// reconcileEnumTypeDefinition self-heals a stale ekarte_db_test so sharedTestSchemaEnumTypes
// edits (G12-2: e.g. item_source lacking 'trimming') take effect without a manual DB reset. A
// previous IF NOT EXISTS guard silently kept stale definitions forever once a type had been
// created once in the test DB.
//
// It prefers the non-destructive path: if the existing value set is an unchanged, order-preserving
// prefix of the new definition (a pure append — the only kind of drift this task actually hit),
// it widens the type in place with ALTER TYPE ... ADD VALUE. An earlier version of this function
// unconditionally did DROP TYPE ... CASCADE + recreate for any mismatch; verified empirically
// (scoped test run) that this transiently breaks any already-provisioned column of the type —
// billing_item_lstep_queries_test.go / billing_item_repository_tx_atomicity_test.go both assume
// billing_items.source already exists and do not themselves AutoMigrate(&model.BillingItem{}),
// so a blanket CASCADE drop leaves them broken until some other test file happens to run its own
// AutoMigrate first. ALTER TYPE ADD VALUE avoids that class of collateral breakage entirely.
//
// DROP+recreate remains the fallback for genuinely incompatible drift (reordered/removed/renamed
// values) — none of the 54 types hit that case for G12-2, but a future migration edit could.
func reconcileEnumTypeDefinition(db *gorm.DB, name, create string) error {
	var existing []string
	if err := db.Raw(`
		SELECT e.enumlabel FROM pg_type t
		JOIN pg_enum e ON t.oid = e.enumtypid
		WHERE t.typname = ?
		ORDER BY e.enumsortorder`, name).Scan(&existing).Error; err != nil {
		return fmt.Errorf("failed to inspect existing ENUM %s: %w", name, err)
	}

	expected := enumValueRe.FindAllString(create, -1)

	if len(existing) == 0 {
		if err := db.Exec(create).Error; err != nil {
			return fmt.Errorf("failed to create ENUM type %s: %w", name, err)
		}
		return nil
	}

	if enumValuesEqual(existing, expected) {
		return nil
	}

	if appended, ok := enumAppendedValues(existing, expected); ok {
		for _, v := range appended {
			if err := db.Exec(fmt.Sprintf("ALTER TYPE %s ADD VALUE IF NOT EXISTS %s", name, v)).Error; err != nil {
				return fmt.Errorf("failed to widen ENUM type %s with value %s: %w", name, v, err)
			}
		}
		return nil
	}

	if err := db.Exec(fmt.Sprintf("DROP TYPE IF EXISTS %s CASCADE", name)).Error; err != nil {
		return fmt.Errorf("failed to drop stale ENUM type %s: %w", name, err)
	}
	if err := db.Exec(create).Error; err != nil {
		return fmt.Errorf("failed to create ENUM type %s: %w", name, err)
	}
	return nil
}

// enumValuesEqual reports whether existing (unquoted labels from pg_enum) exactly matches
// expectedQuoted (quoted literals extracted from a CREATE TYPE string), in the same order.
func enumValuesEqual(existing, expectedQuoted []string) bool {
	if len(existing) != len(expectedQuoted) {
		return false
	}
	for i, v := range expectedQuoted {
		if "'"+existing[i]+"'" != v {
			return false
		}
	}
	return true
}

// enumAppendedValues reports whether expectedQuoted equals existing with one or more values
// appended at the end (an order-preserving prefix match), returning just the appended
// (still-quoted) values in order. Returns ok=false for any reorder/removal/rename.
func enumAppendedValues(existing, expectedQuoted []string) (appended []string, ok bool) {
	if len(expectedQuoted) <= len(existing) {
		return nil, false
	}
	for i, v := range existing {
		if "'"+v+"'" != expectedQuoted[i] {
			return nil, false
		}
	}
	return expectedQuoted[len(existing):], true
}

// TestEnumValuesEqual_TestEnumAppendedValues pins reconcileEnumTypeDefinition's pure helpers —
// the logic that decides between a non-destructive ALTER TYPE ADD VALUE and a destructive
// DROP+recreate (G12-2).
func TestEnumValuesEqual_TestEnumAppendedValues(t *testing.T) {
	t.Run("enumValuesEqual: identical values match", func(t *testing.T) {
		assert.True(t, enumValuesEqual([]string{"manual", "trimming"}, []string{"'manual'", "'trimming'"}))
	})
	t.Run("enumValuesEqual: different length does not match", func(t *testing.T) {
		assert.False(t, enumValuesEqual([]string{"manual"}, []string{"'manual'", "'trimming'"}))
	})
	t.Run("enumValuesEqual: same length different value does not match", func(t *testing.T) {
		assert.False(t, enumValuesEqual([]string{"manual", "hospitalization"}, []string{"'manual'", "'trimming'"}))
	})

	t.Run("enumAppendedValues: pure trailing append is detected (item_source G12-2 case)", func(t *testing.T) {
		appended, ok := enumAppendedValues(
			[]string{"medical_record", "manual", "hospitalization"},
			[]string{"'medical_record'", "'manual'", "'hospitalization'", "'trimming'"},
		)
		assert.True(t, ok)
		assert.Equal(t, []string{"'trimming'"}, appended)
	})
	t.Run("enumAppendedValues: no new values is not an append", func(t *testing.T) {
		_, ok := enumAppendedValues([]string{"a", "b"}, []string{"'a'", "'b'"})
		assert.False(t, ok)
	})
	t.Run("enumAppendedValues: reordered values is not a pure append", func(t *testing.T) {
		_, ok := enumAppendedValues([]string{"a", "b"}, []string{"'b'", "'a'", "'c'"})
		assert.False(t, ok)
	})
	t.Run("enumAppendedValues: removed value is not a pure append", func(t *testing.T) {
		_, ok := enumAppendedValues([]string{"a", "b"}, []string{"'a'"})
		assert.False(t, ok)
	})
}

// setupSharedTestSchema は PostgreSQL カスタム ENUM 型の作成とベースモデルの AutoMigrate を行います。
// setupTestDB から sharedTestSchemaOnce 経由でプロセス全体につき一度だけ呼ばれます。
func setupSharedTestSchema(db *gorm.DB) error {
	// AutoMigrate の前に、PostgreSQL カスタム ENUM 型を作成する（sharedTestSchemaEnumTypes 参照）。
	for _, et := range sharedTestSchemaEnumTypes {
		if err := reconcileEnumTypeDefinition(db, et.name, et.create); err != nil {
			return err
		}
	}

	if err := db.AutoMigrate(
		&model.Owner{},
		&model.MedicalRecord{},
		&model.Billing{},
		&model.Payment{},
		&model.BillingRefund{},
		&model.Treatment{},
	); err != nil {
		return fmt.Errorf("failed to migrate test db: %w", err)
	}
	return nil
}

// getTestDatabaseConnection はテスト用の DB コネクションを返す。接続確立（テストDB存在確認込み）は
// sharedTestDBOnce によりプロセス全体で一度だけ行われ、以降の呼び出しは共有プールを再利用する。
func getTestDatabaseConnection(t *testing.T) *gorm.DB {
	t.Helper()
	sharedTestDBOnce.Do(func() {
		sharedTestDB, sharedTestDBErr = connectTestDatabase()
	})
	if sharedTestDBErr != nil {
		t.Fatalf("failed to connect to test db: %v", sharedTestDBErr)
	}
	return sharedTestDB
}

// connectTestDatabase はテスト用 DB への接続を確立する（sharedTestDBOnce によりプロセス全体で一度だけ呼ばれる）。
func connectTestDatabase() (*gorm.DB, error) {
	// 環境変数から DB パラメータを取得（デフォルト: ekarte_db）
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "db"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "ekarte_user"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "ekarte_password"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "ekarte_db"
	}

	testDBName := dbName + "_test"
	mainDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

	// まず本番DBに接続してテストDBを作成
	mainDB, err := gorm.Open(postgres.Open(mainDSN), &gorm.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to connect to main db: %v\n", err)
	} else {
		if err := ensureTestDatabaseExists(mainDB, testDBName); err != nil {
			return nil, err
		}
		// 接続リーク防止: mainDB は CREATE DATABASE 専用。閉じないと接続が漏れ、
		// PostgreSQL の max_connections を使い切る（FATAL: sorry, too many clients already / SQLSTATE 53300）。
		if sqlMainDB, derr := mainDB.DB(); derr == nil {
			_ = sqlMainDB.Close()
		}
	}

	// テストDB接続
	// 共有プール上で ENUM/テーブルの DROP+CREATE を毎テスト実行すると、サーバサイド prepared
	// statement キャッシュ（pgx デフォルト cache_statement モード）が古い型/リレーション OID を
	// 保持し続け "cache lookup failed" (SQLSTATE XX000) で失敗する。この対策としては
	// setupTestDB 内で全 ENUM を一度きり idempotent 作成するのに加え、DROP+CREATE を行う
	// 個別ヘルパー（checkup_field_repository_test.go / medicine_dose_param_clinic_isolation_test.go）
	// 側もプロセス全体で一度だけ実行するよう sync.Once 化した。これによりプロセス起動後は
	// スキーマが不変となるため、接続プロトコルは pgx デフォルト（cache_statement、最速）のままでよい。
	testDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, testDBName)
	if envDSN := os.Getenv("TEST_DATABASE_URL"); envDSN != "" {
		testDSN = envDSN
	}
	db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test db: %w", err)
	}
	// 接続枯渇防止: このプールは全テストで共有する唯一の接続で、プロセス終了時に TestMain が一度だけ閉じる
	// （テスト毎に開閉すると、full suite で接続確立オーバーヘッドが呼び出し回数分積み上がる）。
	if sqlDB, derr := db.DB(); derr == nil {
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetMaxIdleConns(2)
	}
	return db, nil
}

func ensureTestDatabaseExists(mainDB *gorm.DB, testDBName string) error {
	lockKey := "setup_test_database:" + testDBName
	if err := mainDB.Exec("SELECT pg_advisory_lock(hashtext(?))", lockKey).Error; err != nil {
		return fmt.Errorf("failed to acquire test database creation lock: %w", err)
	}
	defer func() {
		_ = mainDB.Exec("SELECT pg_advisory_unlock(hashtext(?))", lockKey).Error
	}()

	var exists bool
	if err := mainDB.Raw("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = ?)", testDBName).Scan(&exists).Error; err != nil {
		return fmt.Errorf("failed to check test database existence: %w", err)
	}
	if exists {
		return nil
	}

	if err := mainDB.Exec("CREATE DATABASE " + quotePostgresIdentifier(testDBName)).Error; err != nil {
		if isDuplicateDatabaseError(err) {
			return nil
		}
		return fmt.Errorf("failed to create test database %s: %w", testDBName, err)
	}
	return nil
}

func quotePostgresIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func isDuplicateDatabaseError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42P04"
}

func TestQuotePostgresIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		want       string
	}{
		{
			name:       "simple identifier",
			identifier: "ekarte_db_test",
			want:       `"ekarte_db_test"`,
		},
		{
			name:       "embedded quote is escaped",
			identifier: `ekarte"db_test`,
			want:       `"ekarte""db_test"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, quotePostgresIdentifier(tt.identifier))
		})
	}
}

func TestIsDuplicateDatabaseError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "duplicate_database is true",
			err:  &pgconn.PgError{Code: "42P04"},
			want: true,
		},
		{
			name: "wrapped duplicate_database is true",
			err:  fmt.Errorf("create failed: %w", &pgconn.PgError{Code: "42P04"}),
			want: true,
		},
		{
			name: "different pg error is false",
			err:  &pgconn.PgError{Code: "23505"},
			want: false,
		},
		{
			name: "plain error is false",
			err:  errors.New("plain error"),
			want: false,
		},
		{
			name: "nil is false",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isDuplicateDatabaseError(tt.err))
		})
	}
}
