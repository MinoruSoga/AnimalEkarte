// Package testdb はテスト用DB接続・ENUM型・共有ベースモデルの AutoMigrate を提供する。
// BE8-3: backend/internal/repository/db_setup_test.go の setupTestDB 等は `_test.go` 内定義で
// パッケージ外から import 不能だったため、importable な通常 .go ファイルへ抽出した。
// これによりサブパッケージ（例: repository/paymentmethod）からもパッケージローカルな
// DB 統合テストを書けるようになる。既存フラット側の呼び出し元は db_setup_test.go の薄い
// alias 経由でこのパッケージへ委譲し、無変更のまま動作する。
package testdb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// テストDB接続・ENUM型・共有ベースモデルの AutoMigrate はプロセス全体（テスト実行単位）で一度だけ
// 実行する（sharedTestDBOnce/sharedTestSchemaOnce）。130+ ファイル・158+ 箇所の SetupTestDB(t)
// 呼び出し毎にこれらを繰り返すと、接続確立×2・ENUM存在チェック46回・AutoMigrate スキーマ内省
// クエリが呼び出し回数分積み上がり、repository テストスイート全体の支配的コストになる。
// TRUNCATE のみ SetupTestDB 内で呼び出し毎に実行し、テスト間データ分離を維持する。
//
// ドメインテーブルの AutoMigrate も EnsureAutoMigrated 経由でモデル型ごとに一度だけに抑える。
// テスト間データ分離は各ヘルパーの TRUNCATE に依存する。呼び出し側パッケージに t.Parallel() は無い。
var (
	sharedTestDB         *gorm.DB
	sharedTestDBOnce     sync.Once
	sharedTestDBErr      error
	sharedTestSchemaOnce sync.Once
	sharedTestSchemaErr  error
	// autoMigratedTypes: EnsureAutoMigrated 済みの *model 要素型（reflect.Type of non-pointer struct）。
	autoMigratedTypes sync.Map
	autoMigrateMu     sync.Mutex
)

// modelElemType は AutoMigrate 引数（通常 &model.X{}）からキャッシュキーとなる構造体型を返す。
func modelElemType(m any) reflect.Type {
	t := reflect.TypeOf(m)
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

// EnsureAutoMigrated は各モデル型についてプロセス全体で一度だけ db.AutoMigrate を実行する。
// スキップするのは既に migrate 済みの型（プロセス内キャッシュ）であり、既存テーブルをスキップするわけではない。
// DROP TABLE 後の再作成には使わないこと（キャッシュが再 CREATE を抑止する）。
// many2many join table が既にある場合、GORM AutoMigrate は列追加で widen しないことがある
// （staff_permission_groups.created_at は setupSharedTestSchema の ALTER で補う）。
func EnsureAutoMigrated(db *gorm.DB, models ...any) error {
	if len(models) == 0 {
		return nil
	}

	// 高速パス: ロック前に全型が済みなら no-op
	pending := make([]any, 0, len(models))
	for _, m := range models {
		t := modelElemType(m)
		if t == nil {
			continue
		}
		if _, ok := autoMigratedTypes.Load(t); ok {
			continue
		}
		pending = append(pending, m)
	}
	if len(pending) == 0 {
		return nil
	}

	autoMigrateMu.Lock()
	defer autoMigrateMu.Unlock()

	still := make([]any, 0, len(pending))
	types := make([]reflect.Type, 0, len(pending))
	for _, m := range pending {
		t := modelElemType(m)
		if t == nil {
			continue
		}
		if _, ok := autoMigratedTypes.Load(t); ok {
			continue
		}
		still = append(still, m)
		types = append(types, t)
	}
	if len(still) == 0 {
		return nil
	}
	if err := db.AutoMigrate(still...); err != nil {
		return fmt.Errorf("auto-migrate test database models: %w", err)
	}
	for _, t := range types {
		autoMigratedTypes.Store(t, struct{}{})
	}
	return nil
}

// MarkAutoMigrated は SetupSharedTestSchema 等で生 AutoMigrate した型をキャッシュに登録する。
func MarkAutoMigrated(models ...any) {
	for _, m := range models {
		if t := modelElemType(m); t != nil {
			autoMigratedTypes.Store(t, struct{}{})
		}
	}
}

// CloseSharedTestDB は共有 DB 接続プールを閉じる。呼び出し側パッケージの TestMain から
// プロセス終了直前に一度だけ呼ぶ（db_setup_test.go の TestMain 参照）。
func CloseSharedTestDB() {
	if sharedTestDB != nil {
		if sqlDB, err := sharedTestDB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
}

// SetupTestDB はテスト用の DB を返し、共有ベーステーブルを TRUNCATE してクリーンな状態にします。
// DB接続確立・ENUM型作成・ベースモデルの AutoMigrate はプロセス全体で一度だけ実行されます。
func SetupTestDB(t *testing.T) *gorm.DB {
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

// EnsureClinicSettingsTable は clinic_settings テーブルを実マイグレーション
// （backend/migrations/001_init.sql の CREATE TABLE clinic_settings + 011番 ALTER TABLE ADD COLUMN
// closing_am_start）と一致する生SQLで用意する。GORM の AutoMigrate は gorm:"type:time" タグを無視し
// timestamptz として CREATE TABLE を生成する既知の相性問題があり model.ClinicSettings を直接
// AutoMigrate できないため（#236 BUG#3 調査時に確認済み）、clinic_settings に依存する複数のテストファイル
// がこの生SQLヘルパーを共有する。IF NOT EXISTS のため複数回呼んでも安全。
func EnsureClinicSettingsTable(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`
		CREATE TABLE IF NOT EXISTS clinic_settings (
			clinic_id              bigint       PRIMARY KEY REFERENCES clinics(id) ON DELETE CASCADE,
			closing_am_pm_boundary time         NOT NULL DEFAULT '14:00',
			closing_weekday_end    time         NOT NULL DEFAULT '18:30',
			closing_sunday_end     time         NOT NULL DEFAULT '17:30',
			closing_am_start       time         NOT NULL DEFAULT '09:00',
			closed_weekdays        smallint[]   NOT NULL DEFAULT '{}',
			cpm_version            varchar(8)   NOT NULL DEFAULT 'v1'
			                       CHECK (cpm_version IN ('v1', 'v2')),
			dormant_prevention_180_days integer  NOT NULL DEFAULT 180,
			dormant_prevention_210_days integer  NOT NULL DEFAULT 210,
			dormant_prevention_240_days integer  NOT NULL DEFAULT 240,
			dormant_prevention_365_days integer  NOT NULL DEFAULT 365,
			cpm_v2_coming_threshold  INT NOT NULL DEFAULT 2  CHECK (cpm_v2_coming_threshold  >= 1),
			cpm_v2_good_threshold    INT NOT NULL DEFAULT 4  CHECK (cpm_v2_good_threshold    >= 1),
			cpm_v2_family_threshold  INT NOT NULL DEFAULT 8  CHECK (cpm_v2_family_threshold  >= 1),
			cpm_v2_noah_threshold    INT NOT NULL DEFAULT 13 CHECK (cpm_v2_noah_threshold    >= 1),
			cpm_v1_dormant_days       INT     NOT NULL DEFAULT 240    CHECK (cpm_v1_dormant_days       >= 1),
			cpm_v1_noah_days          INT     NOT NULL DEFAULT 365    CHECK (cpm_v1_noah_days          >= 1),
			cpm_v1_noah_annual_visits INT     NOT NULL DEFAULT 3      CHECK (cpm_v1_noah_annual_visits >= 1),
			cpm_v1_noah_ltv           BIGINT  NOT NULL DEFAULT 80000  CHECK (cpm_v1_noah_ltv           >= 0),
			cpm_v1_core_days          INT     NOT NULL DEFAULT 180    CHECK (cpm_v1_core_days          >= 1),
			cpm_v1_core_annual_visits INT     NOT NULL DEFAULT 2      CHECK (cpm_v1_core_annual_visits >= 1),
			cpm_v1_core_ltv           BIGINT  NOT NULL DEFAULT 50000  CHECK (cpm_v1_core_ltv           >= 0),
			cpm_v1_spot_min_amount    BIGINT  NOT NULL DEFAULT 30000  CHECK (cpm_v1_spot_min_amount    >= 0),
			cpm_v1_spot_inactive_days INT     NOT NULL DEFAULT 90     CHECK (cpm_v1_spot_inactive_days >= 1),
			cpm_v1_growing_max_days   INT     NOT NULL DEFAULT 90     CHECK (cpm_v1_growing_max_days   >= 1),
			cpm_v1_growing_min_visits INT     NOT NULL DEFAULT 2      CHECK (cpm_v1_growing_min_visits >= 1),
			cpm_v1_growing_max_visits INT     NOT NULL DEFAULT 3      CHECK (cpm_v1_growing_max_visits >= 1),
			cpm_v1_ltv_break_low      BIGINT  NOT NULL DEFAULT 20000  CHECK (cpm_v1_ltv_break_low      >= 0),
			health_prevention_lookback_days INT NOT NULL DEFAULT 365,
			vaccine_deadline_days           INT NOT NULL DEFAULT 60,
			created_at             timestamptz  NOT NULL DEFAULT now(),
			updated_at             timestamptz  NOT NULL DEFAULT now()
		)
	`).Error)
}

// EnsureStaffPermissionGroupsCreatedAt self-heals a stale ekarte_db_test where
// staff_permission_groups was created as a bare many2many join (staff_id, group_id only).
// model.StaffPermissionGroup.CreatedAt and 001_init.sql both require created_at.
// GORM AutoMigrate does not always widen an existing join table that
// PermissionGroup.Staffs many2many already provisioned without the extra column.
func EnsureStaffPermissionGroupsCreatedAt(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, ensureStaffPermissionGroupsCreatedAt(db))
}

func ensureStaffPermissionGroupsCreatedAt(db *gorm.DB) error {
	if err := db.Exec(`
		ALTER TABLE staff_permission_groups
		ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now()
	`).Error; err != nil {
		return fmt.Errorf("failed to ensure staff_permission_groups.created_at: %w", err)
	}
	return nil
}

// MakeTestOwner はテスト用の Owner を作成して返す。
func MakeTestOwner(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Owner {
	t.Helper()
	o := &model.Owner{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(o).Error)
	return o
}

// SetupIsolatedTestDB は SetupTestDB と異なり、プロセス全体で共有しない「呼び出し毎に完全に新しい」
// DB 接続を返す（最適化前の setupTestDB と同じ挙動）。
//
// checkup_field_results/checkup_type_fields は checkup_field_repository_test.go（AutoMigrate 由来
// スキーマ）・checkup_field_cascade_test.go（migration 010 の実 DDL）・
// checkup_field_composite_fk_test.go（010 実 DDL + migration 012 複合 FK）という 3 種の異なる
// ヘルパーが同じテーブルを意図的に毎回 DROP+CREATE し合う（migration drift 検出が目的で、
// 挙動として必須）。この cluster に SetupTestDB の共有コネクションプールを使うと、いずれかの
// ヘルパーが DROP TABLE/DROP TYPE した瞬間に、別テストが既に保持していた同一物理コネクション上の
// キャッシュ済み prepared statement（古いテーブル/型 OID 参照）が
// "cache lookup failed"（SQLSTATE XX000）で壊れる。3 ヘルパーの意図的な毎回 DROP+CREATE は
// 統合できないため、この cluster だけは共有プールから外し、テスト毎に使い捨ての新規コネクションを
// 割り当てることでキャッシュ汚染を根本的に回避する（対象は少数のためスループット影響は軽微）。
func SetupIsolatedTestDB(t *testing.T) *gorm.DB {
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

// EnumType is one hand-maintained ENUM type double used by setupSharedTestSchema.
// Kept as a package-level type/var (rather than a func-local literal) so
// test_schema_enum_parity_test.go (G12-2) can compare it against 001_init.sql directly instead
// of re-parsing Go source via go/ast. Exported (BE8-3): the flat repository package's
// test_schema_enum_parity_test.go reads this type/var across the package boundary.
type EnumType struct {
	Name   string
	Create string
}

// SharedTestSchemaEnumTypes hand-duplicates every PostgreSQL ENUM type from 001_init.sql
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
var SharedTestSchemaEnumTypes = []EnumType{
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
	{"lab_import_job_status", "CREATE TYPE lab_import_job_status AS ENUM ('received', 'validated', 'mapped', 'persisted', 'duplicate', 'needs_review', 'failed', 'reverted')"},
	{"lab_import_source_type", "CREATE TYPE lab_import_source_type AS ENUM ('fixture', 'drwan', 'manual', 'fuji_nx600', 'fuji_au10v', 'arkray_pu4010', 'idexx_vetlab')"},
	// #211 健診パッケージ（migration 010 → 001_init.sql 統合済み）
	{"checkup_field_type", "CREATE TYPE checkup_field_type AS ENUM ('number', 'single_select', 'multi_select', 'boolean', 'checklist', 'text')"},
}

// EnumValueRe extracts the ordered list of quoted ENUM value literals out of a
// "CREATE TYPE ... AS ENUM ('a', 'b', ...)" definition string. Exported (BE8-3):
// test_schema_enum_parity_test.go (flat repository package) reuses this regex.
var EnumValueRe = regexp.MustCompile(`'[^']*'`)

// reconcileEnumTypeDefinition self-heals a stale ekarte_db_test so SharedTestSchemaEnumTypes
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

	expected := EnumValueRe.FindAllString(create, -1)

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

// setupSharedTestSchema は PostgreSQL カスタム ENUM 型の作成とベースモデルの AutoMigrate を行います。
// SetupTestDB から sharedTestSchemaOnce 経由でプロセス全体につき一度だけ呼ばれます。
func setupSharedTestSchema(db *gorm.DB) error {
	// AutoMigrate の前に、PostgreSQL カスタム ENUM 型を作成する（SharedTestSchemaEnumTypes 参照）。
	for _, et := range SharedTestSchemaEnumTypes {
		if err := reconcileEnumTypeDefinition(db, et.Name, et.Create); err != nil {
			return err
		}
	}

	// Shared bootstrap for packages that query beyond the original 4 billing tables
	// without calling EnsureAutoMigrated first (e.g. unbilled vaccinations on CreateItem).
	coreModels := []any{
		&model.Company{},
		&model.Clinic{},
		&model.Account{},
		&model.Staff{},
		&model.Owner{},
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.MedicalRecord{},
		&model.Billing{},
		&model.Payment{},
		&model.PaymentMethodMaster{},
		&model.BillingRefund{},
		&model.BillingItem{},
		&model.Treatment{},
		&model.PaymentSplit{},
		&model.Estimate{},
		&model.EstimateItem{},
		&model.MerchandiseItem{},
		&model.CashRegisterClose{},
		&model.Examination{},
		&model.ExamResult{},
		&model.Inquiry{},
		&model.VitalRecord{},
		&model.TreatmentPlan{},
		&model.Vaccine{},
		&model.Vaccination{},
		&model.PasswordResetToken{},
		&model.PermissionGroup{},
		&model.PermissionGroupRule{},
	}
	if err := db.AutoMigrate(coreModels...); err != nil {
		return fmt.Errorf("failed to migrate test db: %w", err)
	}
	MarkAutoMigrated(coreModels...)
	// PermissionGroup.Staffs many2many creates a bare join table during the
	// AutoMigrate above. Registering StaffPermissionGroup later does not
	// reliably ADD created_at; ALTER is the source of truth for test DB.
	if err := ensureStaffPermissionGroupsCreatedAt(db); err != nil {
		return err
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
	// SetupTestDB 内で全 ENUM を一度きり idempotent 作成するのに加え、DROP+CREATE を行う
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
	// 接続枯渇防止: このプールは全テストで共有する唯一の接続で、プロセス終了時に呼び出し側の
	// TestMain が CloseSharedTestDB() 経由で一度だけ閉じる（テスト毎に開閉すると、full suite で
	// 接続確立オーバーヘッドが呼び出し回数分積み上がる）。
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
