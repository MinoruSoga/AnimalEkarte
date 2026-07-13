package repository

// clinic_settings_repository_test.go — ClinicSettingsRepository の統合テスト。
//
// clinic_settings は clinics(id) への1:1（PK=clinic_id）。makeClinicFixture（clinic_repository_test.go）
// で都度新規 clinic を作り、他テストの残存行と衝突しない clinic_id を得てから検証する。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// setupClinicSettingsTestDB は clinic_settings_repository のテスト用に DB を整備する。
//
// FIXED (2026-07-13, Issue #212 → #236 BUG#3): 元々 db.AutoMigrate(&model.ClinicSettings{}) が
// "invalid input syntax for type timestamp with time zone: 14:00" (SQLSTATE 22007) で失敗し
// 全7テストが t.Skip されていた（GORM の AutoMigrate が gorm:"type:time" タグを無視し timestamptz
// として CREATE TABLE を生成する既知の相性問題）。プロダクションコード（model/migration）を変更せず、
// テスト専用に実マイグレーション（backend/migrations/001_init.sql の CREATE TABLE clinic_settings +
// 011番 ALTER TABLE ADD COLUMN closing_am_start、2036-2071行目・3635行目）と一致する生SQLで
// テーブルを用意することでこの AutoMigrate 問題は解消した。
//
// 別途発覚していた BUG#3（model.ClinicSettings の DormantPrevention180Days / CPMV2ComingThreshold /
// CPMV1DormantDays 等、GORM デフォルトの snake_case 変換が実スキーマと異なる列名を生成し
// `SQLSTATE 42703` で Create()/UPSERT が失敗する問題）は、対応する gorm:"column:..." タグの
// 明示付与により解消済み（model/clinic_settings.go 参照）。本ファイルの t.Skip は全て解除し、
// 2026-07-13 時点で TestClinicSettingsRepository_* 全サブテストが green であることを確認済み。
func setupClinicSettingsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Company{}, &model.Clinic{}))
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
	db.Exec("TRUNCATE TABLE clinic_settings CASCADE")
	return db
}

func TestClinicSettingsRepository_FindByClinicID(t *testing.T) {
	db := setupClinicSettingsTestDB(t)
	repo := NewClinicSettingsRepository(db)
	ctx := context.Background()

	t.Run("行が存在しなければデフォルト値を返す(エラーなし)", func(t *testing.T) {
		clinic := makeClinicFixture(t, db, "設定未作成クリニック")
		got, err := repo.FindByClinicID(ctx, clinic.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, clinic.ID, got.ClinicID)
		assert.Equal(t, "14:00", got.ClosingAmPmBoundary)
		assert.Equal(t, "18:30", got.ClosingWeekdayEnd)
		assert.Equal(t, "17:30", got.ClosingSundayEnd)
	})

	t.Run("clinic_id隔離: 他院の設定は返らず自院はデフォルト値のまま", func(t *testing.T) {
		clinicA := makeClinicFixture(t, db, "隔離検証A")
		clinicB := makeClinicFixture(t, db, "隔離検証B")
		sB := &model.ClinicSettings{ClinicID: clinicB.ID, ClosingAmPmBoundary: "15:45", ClosingWeekdayEnd: "20:00", ClosingSundayEnd: "18:00"}
		_, err := repo.Save(ctx, clinicB.ID, sB)
		require.NoError(t, err)

		gotA, err := repo.FindByClinicID(ctx, clinicA.ID)
		require.NoError(t, err)
		assert.Equal(t, "14:00", gotA.ClosingAmPmBoundary, "他院(B)の値が漏れずデフォルト値のまま")

		gotB, err := repo.FindByClinicID(ctx, clinicB.ID)
		require.NoError(t, err)
		assert.Contains(t, gotB.ClosingAmPmBoundary, "15:45")
	})

	t.Run("行が存在すれば実際の値を返す", func(t *testing.T) {
		clinic := makeClinicFixture(t, db, "設定既存クリニック")
		s := &model.ClinicSettings{ClinicID: clinic.ID, ClosingAmPmBoundary: "13:00", ClosingWeekdayEnd: "19:00", ClosingSundayEnd: "16:00"}
		_, err := repo.Save(ctx, clinic.ID, s)
		require.NoError(t, err)

		got, err := repo.FindByClinicID(ctx, clinic.ID)
		require.NoError(t, err)
		assert.Contains(t, got.ClosingAmPmBoundary, "13:00")
	})
}

func TestClinicSettingsRepository_Save(t *testing.T) {
	db := setupClinicSettingsTestDB(t)
	repo := NewClinicSettingsRepository(db)
	ctx := context.Background()
	clinic := makeClinicFixture(t, db, "Save用クリニック")

	first := &model.ClinicSettings{ClinicID: clinic.ID, ClosingAmPmBoundary: "11:00", ClosingWeekdayEnd: "18:00", ClosingSundayEnd: "17:00"}
	saved, err := repo.Save(ctx, clinic.ID, first)
	require.NoError(t, err)
	require.NotNil(t, saved)

	got, err := repo.FindByClinicID(ctx, clinic.ID)
	require.NoError(t, err)
	assert.Contains(t, got.ClosingAmPmBoundary, "11:00")

	// 同一 clinic_id での再 Save は UPSERT で上書きされる（新規行にならない）
	second := &model.ClinicSettings{ClinicID: clinic.ID, ClosingAmPmBoundary: "12:30", ClosingWeekdayEnd: "18:00", ClosingSundayEnd: "17:00"}
	_, err = repo.Save(ctx, clinic.ID, second)
	require.NoError(t, err)

	got2, err := repo.FindByClinicID(ctx, clinic.ID)
	require.NoError(t, err)
	assert.Contains(t, got2.ClosingAmPmBoundary, "12:30", "再 Save で値が上書きされる")

	var count int64
	db.Model(&model.ClinicSettings{}).Where("clinic_id = ?", clinic.ID).Count(&count)
	assert.Equal(t, int64(1), count, "UPSERT のため行は増えない")
}

func TestClinicSettingsRepository_UpdateCPMVersion(t *testing.T) {
	db := setupClinicSettingsTestDB(t)
	repo := NewClinicSettingsRepository(db)
	ctx := context.Background()
	clinic := makeClinicFixture(t, db, "CPMバージョン更新用")

	t.Run("未作成なら新規作成される", func(t *testing.T) {
		err := repo.UpdateCPMVersion(ctx, clinic.ID, "v2")
		require.NoError(t, err)
		got, err := repo.FindByClinicID(ctx, clinic.ID)
		require.NoError(t, err)
		assert.Equal(t, "v2", got.CPMVersion)
	})

	t.Run("既存行の他カラムは保持される(OnConflict列限定)", func(t *testing.T) {
		custom := &model.ClinicSettings{ClinicID: clinic.ID, ClosingAmPmBoundary: "10:00", ClosingWeekdayEnd: "20:00", ClosingSundayEnd: "19:00"}
		_, err := repo.Save(ctx, clinic.ID, custom)
		require.NoError(t, err)

		err = repo.UpdateCPMVersion(ctx, clinic.ID, "v1")
		require.NoError(t, err)

		got, err := repo.FindByClinicID(ctx, clinic.ID)
		require.NoError(t, err)
		assert.Equal(t, "v1", got.CPMVersion)
		assert.Contains(t, got.ClosingAmPmBoundary, "10:00", "UpdateCPMVersion は他カラムを変更しない")
	})
}

func TestClinicSettingsRepository_UpdateDormantThresholds(t *testing.T) {
	db := setupClinicSettingsTestDB(t)
	repo := NewClinicSettingsRepository(db)
	ctx := context.Background()
	clinic := makeClinicFixture(t, db, "休眠閾値更新用")

	err := repo.UpdateDormantThresholds(ctx, clinic.ID, model.DormantThresholds{
		Stage180: 100, Stage210: 200, Stage240: 300, Stage365: 400,
	})
	require.NoError(t, err)

	got, err := repo.FindByClinicID(ctx, clinic.ID)
	require.NoError(t, err)
	assert.Equal(t, 100, got.DormantPrevention180Days)
	assert.Equal(t, 200, got.DormantPrevention210Days)
	assert.Equal(t, 300, got.DormantPrevention240Days)
	assert.Equal(t, 400, got.DormantPrevention365Days)

	t.Run("他カラムは保持される", func(t *testing.T) {
		custom := &model.ClinicSettings{ClinicID: clinic.ID, ClosingAmPmBoundary: "09:30", ClosingWeekdayEnd: "18:00", ClosingSundayEnd: "17:00"}
		_, err := repo.Save(ctx, clinic.ID, custom)
		require.NoError(t, err)

		require.NoError(t, repo.UpdateDormantThresholds(ctx, clinic.ID, model.DormantThresholds{
			Stage180: 111, Stage210: 211, Stage240: 311, Stage365: 411,
		}))

		got, err := repo.FindByClinicID(ctx, clinic.ID)
		require.NoError(t, err)
		assert.Equal(t, 111, got.DormantPrevention180Days)
		assert.Contains(t, got.ClosingAmPmBoundary, "09:30")
	})
}

func TestClinicSettingsRepository_UpdateCPMV2Thresholds(t *testing.T) {
	db := setupClinicSettingsTestDB(t)
	repo := NewClinicSettingsRepository(db)
	ctx := context.Background()
	clinic := makeClinicFixture(t, db, "CPMV2閾値更新用")

	err := repo.UpdateCPMV2Thresholds(ctx, clinic.ID, model.CPMV2Thresholds{
		Coming: 3, Good: 5, Family: 9, Noah: 14,
	})
	require.NoError(t, err)

	got, err := repo.FindByClinicID(ctx, clinic.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, got.CPMV2ComingThreshold)
	assert.Equal(t, 5, got.CPMV2GoodThreshold)
	assert.Equal(t, 9, got.CPMV2FamilyThreshold)
	assert.Equal(t, 14, got.CPMV2NoahThreshold)
}

func TestClinicSettingsRepository_UpdateCPMV1Thresholds(t *testing.T) {
	db := setupClinicSettingsTestDB(t)
	repo := NewClinicSettingsRepository(db)
	ctx := context.Background()
	clinic := makeClinicFixture(t, db, "CPMV1閾値更新用")

	thresholds := model.CPMV1Thresholds{
		DormantDays: 241, NoahDays: 366, NoahAnnualVisits: 4, NoahLTV: 81000,
		CoreDays: 181, CoreAnnualVisits: 3, CoreLTV: 51000,
		SpotMinAmount: 31000, SpotInactiveDays: 91,
		GrowingMaxDays: 91, GrowingMinVisits: 3, GrowingMaxVisits: 4,
		LTVBreakLow: 21000,
	}
	err := repo.UpdateCPMV1Thresholds(ctx, clinic.ID, thresholds)
	require.NoError(t, err)

	got, err := repo.FindByClinicID(ctx, clinic.ID)
	require.NoError(t, err)
	assert.Equal(t, 241, got.CPMV1DormantDays)
	assert.Equal(t, int64(81000), got.CPMV1NoahLTV)
	assert.Equal(t, 91, got.CPMV1SpotInactiveDays)
	assert.Equal(t, int64(21000), got.CPMV1LTVBreakLow)
}

func TestClinicSettingsRepository_UpdateHealthPreventionThresholds(t *testing.T) {
	db := setupClinicSettingsTestDB(t)
	repo := NewClinicSettingsRepository(db)
	ctx := context.Background()
	clinic := makeClinicFixture(t, db, "健診予防閾値更新用")

	err := repo.UpdateHealthPreventionThresholds(ctx, clinic.ID, model.HealthPreventionThresholds{
		LookbackDays: 400, VaccineDeadline: 45,
	})
	require.NoError(t, err)

	got, err := repo.FindByClinicID(ctx, clinic.ID)
	require.NoError(t, err)
	assert.Equal(t, 400, got.HealthPreventionLookbackDays)
	assert.Equal(t, 45, got.VaccineDeadlineDays)
}
