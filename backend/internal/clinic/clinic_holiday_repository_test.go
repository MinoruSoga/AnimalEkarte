package clinic

// clinic_holiday_repository_test.go — 個別休診日 data access の統合テスト。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupClinicHolidayTestDB は個別休診日 data access のテスト用に DB を整備する。
func setupClinicHolidayTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Company{}, &model.Clinic{}, &model.ClinicHoliday{}))
	require.NoError(t, db.Exec("TRUNCATE TABLE clinic_holidays").Error)
	// clinics/companies はテスト全体で TRUNCATE されない共有テーブル。他ファイル
	// （staff_preload_clinic_isolation_test.go の seedClinicsForFK 等）が clinics.id を
	// 明示指定して手動 INSERT すると、bigserial シーケンスの内部カウンタは追従せず、
	// 後続の makeClinicFixture（シーケンス採番）が同じ id を再度払い出して
	// "duplicate key value violates unique constraint clinics_pkey" を起こしうる。
	// シーケンスを現在の最大 id 以上に明示同期し、この実行順依存の衝突を防ぐ。
	db.Exec("SELECT setval('clinics_id_seq', GREATEST((SELECT COALESCE(MAX(id), 0) FROM clinics), 1))")
	// model.ClinicHoliday に uniqueIndex タグが無いため AutoMigrate では複合 UNIQUE(clinic_id, date)
	// が再現されない。Save() の ON CONFLICT ("clinic_id","date") が要求するため明示的に追加する
	// （shift_entry_repository_test.go 等、本パッケージの既存パターンを踏襲）。
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_test_clinic_holiday_clinic_date
		ON clinic_holidays (clinic_id, date)`)
	return db
}

// makeClinicFixture は新規 company + clinic を作成して返す。
func makeClinicFixture(t *testing.T, db *gorm.DB, name string) *model.Clinic {
	t.Helper()
	ctx := context.Background()
	company := &model.Company{Name: "テスト法人_" + name}
	require.NoError(t, db.WithContext(ctx).Create(company).Error)
	clinic := &model.Clinic{CompanyID: company.ID, Name: name}
	require.NoError(t, db.WithContext(ctx).Create(clinic).Error)
	return clinic
}

func TestClinicHolidayRepository_FindAllByYearMonth(t *testing.T) {
	db := setupClinicHolidayTestDB(t)
	repo := NewClinicHolidayRepository(db)
	ctx := context.Background()

	clinicA := makeClinicFixture(t, db, "休診日一覧A")
	clinicB := makeClinicFixture(t, db, "休診日一覧B")

	julyEarly := time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)
	julyLate := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	august := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicHoliday{ClinicID: clinicA.ID, Date: julyLate, Reason: "夏季休診(後半)"}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicHoliday{ClinicID: clinicA.ID, Date: julyEarly, Reason: "夏季休診(前半)"}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicHoliday{ClinicID: clinicA.ID, Date: august, Reason: "8月の休診"}).Error)
	// 別クリニックの休診日（結果に混入してはならない）
	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicHoliday{ClinicID: clinicB.ID, Date: julyEarly, Reason: "別クリニックの休診"}).Error)

	t.Run("yearMonth 指定で該当月のみ date ASC で返る", func(t *testing.T) {
		got, err := repo.FindAllByYearMonth(ctx, clinicA.ID, "2026-07")
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.True(t, got[0].Date.Before(got[1].Date) || got[0].Date.Equal(got[1].Date))
		assert.Equal(t, "夏季休診(前半)", got[0].Reason)
		assert.Equal(t, "夏季休診(後半)", got[1].Reason)
	})

	t.Run("yearMonth 未指定は全件返る(clinic スコープ内)", func(t *testing.T) {
		got, err := repo.FindAllByYearMonth(ctx, clinicA.ID, "")
		require.NoError(t, err)
		assert.Len(t, got, 3)
	})

	t.Run("別クリニックの休診日は含まれない", func(t *testing.T) {
		got, err := repo.FindAllByYearMonth(ctx, clinicB.ID, "2026-07")
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "別クリニックの休診", got[0].Reason)
	})
}

func TestClinicHolidayRepository_Save(t *testing.T) {
	db := setupClinicHolidayTestDB(t)
	repo := NewClinicHolidayRepository(db)
	ctx := context.Background()
	clinic := makeClinicFixture(t, db, "休診日Save用")
	date := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)

	first := &model.ClinicHoliday{ClinicID: clinic.ID, Date: date, Reason: "臨時休診"}
	saved, err := repo.Save(ctx, clinic.ID, first)
	require.NoError(t, err)
	require.NotNil(t, saved)

	got, err := repo.FindByDate(ctx, clinic.ID, date)
	require.NoError(t, err)
	assert.Equal(t, "臨時休診", got.Reason)

	// 同一 (clinic_id, date) の再 Save は上書きせず AlreadyExists（BUG-015）
	second := &model.ClinicHoliday{ClinicID: clinic.ID, Date: date, Reason: "理由変更後"}
	_, err = repo.Save(ctx, clinic.ID, second)
	require.Error(t, err)
	assert.True(t, apperrors.IsAlreadyExists(err))

	got2, err := repo.FindByDate(ctx, clinic.ID, date)
	require.NoError(t, err)
	assert.Equal(t, "臨時休診", got2.Reason, "既存 reason は上書きされない")

	var count int64
	db.Model(&model.ClinicHoliday{}).Where("clinic_id = ? AND date = ?", clinic.ID, date.Format("2006-01-02")).Count(&count)
	assert.Equal(t, int64(1), count, "重複追加では行が増えない")
}

// POC-09: Save は struct の ClinicID より arg clinicID を正として書き込む。
func TestClinicHolidayRepository_Save_OverwritesStructClinicID(t *testing.T) {
	db := setupClinicHolidayTestDB(t)
	repo := NewClinicHolidayRepository(db)
	ctx := context.Background()
	clinic := makeClinicFixture(t, db, "休診日ClinicID上書き")
	date := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)

	holiday := &model.ClinicHoliday{ClinicID: clinic.ID + 9999, Date: date, Reason: "arg優先"}
	saved, err := repo.Save(ctx, clinic.ID, holiday)
	require.NoError(t, err)
	require.Equal(t, clinic.ID, saved.ClinicID)

	got, err := repo.FindByDate(ctx, clinic.ID, date)
	require.NoError(t, err)
	assert.Equal(t, "arg優先", got.Reason)
	assert.Equal(t, clinic.ID, got.ClinicID)
}

func TestClinicHolidayRepository_Delete(t *testing.T) {
	db := setupClinicHolidayTestDB(t)
	repo := NewClinicHolidayRepository(db)
	ctx := context.Background()
	clinic := makeClinicFixture(t, db, "休診日Delete用")
	date := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicHoliday{ClinicID: clinic.ID, Date: date, Reason: "削除対象"}).Error)

	t.Run("成功", func(t *testing.T) {
		err := repo.Delete(ctx, clinic.ID, date)
		require.NoError(t, err)

		_, err = repo.FindByDate(ctx, clinic.ID, date)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない日付はNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinic.ID, time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC))
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestClinicHolidayRepository_FindByDate(t *testing.T) {
	db := setupClinicHolidayTestDB(t)
	repo := NewClinicHolidayRepository(db)
	ctx := context.Background()

	clinicA := makeClinicFixture(t, db, "休診日単件A")
	clinicB := makeClinicFixture(t, db, "休診日単件B")
	date := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)

	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicHoliday{ClinicID: clinicA.ID, Date: date, Reason: "文化の日"}).Error)

	t.Run("同一クリニックでは取得できる", func(t *testing.T) {
		got, err := repo.FindByDate(ctx, clinicA.ID, date)
		require.NoError(t, err)
		assert.Equal(t, "文化の日", got.Reason)
	})

	t.Run("別クリニックからは取得できない(clinic_id 隔離)", func(t *testing.T) {
		got, err := repo.FindByDate(ctx, clinicB.ID, date)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない日付はNotFound", func(t *testing.T) {
		got, err := repo.FindByDate(ctx, clinicA.ID, time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC))
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
