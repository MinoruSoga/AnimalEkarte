package repository

// clinic_repository_test.go — ClinicRepository の統合テスト。
//
// clinics / companies テーブルはどの setupTestDB / setupXTestDB からも TRUNCATE されない
// （本番同様の永続シードデータが乗っている前提）。そのため各テストは makeClinicFixture で
// 都度あたらしい company + clinic を作成し、他テストのシードデータや実行順に依存しない
// 完全に隔離された clinic_id を得てから検証する。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// setupClinicTestDB は clinic_repository のテスト用に DB を整備する。
func setupClinicTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&model.Company{}, &model.Clinic{}, &model.Staff{}, &model.StaffClinicAssignment{}, &model.PermissionGroup{},
	))
	return db
}

// makeClinicFixture は新規 company + clinic を作成して返す。
// clinics/companies は他テストから TRUNCATE されないため、既存シードデータを汚さず・
// 汚されないよう常に新規行を作る（自動採番の ID は過去に一度も参照されていないことが保証される）。
func makeClinicFixture(t *testing.T, db *gorm.DB, name string) *model.Clinic {
	t.Helper()
	ctx := context.Background()
	company := &model.Company{Name: "テスト法人_" + name}
	require.NoError(t, db.WithContext(ctx).Create(company).Error)
	clinic := &model.Clinic{CompanyID: company.ID, Name: name}
	require.NoError(t, db.WithContext(ctx).Create(clinic).Error)
	return clinic
}

func TestClinicRepository_FindAll(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	clinicA := makeClinicFixture(t, db, "Aテスト診療所_順序確認")
	clinicZ := makeClinicFixture(t, db, "Zテスト診療所_順序確認")

	got, err := repo.FindAll(ctx)
	require.NoError(t, err)

	idxA, idxZ := -1, -1
	for i, c := range got {
		if c.ID == clinicA.ID {
			idxA = i
		}
		if c.ID == clinicZ.ID {
			idxZ = i
		}
	}
	require.NotEqual(t, -1, idxA, "clinicA が結果に含まれること")
	require.NotEqual(t, -1, idxZ, "clinicZ が結果に含まれること")
	assert.Less(t, idxA, idxZ, "name ASC 順で A が Z より前に来ること")
}

func TestClinicRepository_FindByStaffID(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	clinicA := makeClinicFixture(t, db, "スタッフ検索用A")
	clinicB := makeClinicFixture(t, db, "スタッフ検索用B")
	clinicC := makeClinicFixture(t, db, "スタッフ検索用Cソフト削除")

	staff := &model.Staff{ClinicID: clinicA.ID, Name: "兼務スタッフ", StaffType: model.StaffTypeDoctor}
	require.NoError(t, db.WithContext(ctx).Create(staff).Error)

	// 有効な割当: clinic A, B
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicA.ID, IsMain: true}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicB.ID}).Error)

	// ソフト削除済み割当: clinic C（結果に含まれてはならない）
	assignC := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicC.ID}
	require.NoError(t, db.WithContext(ctx).Create(assignC).Error)
	require.NoError(t, db.WithContext(ctx).Delete(assignC).Error)

	got, err := repo.FindByStaffID(ctx, staff.ID)
	require.NoError(t, err)

	ids := make([]uint64, 0, len(got))
	for _, c := range got {
		ids = append(ids, c.ID)
	}
	assert.Contains(t, ids, clinicA.ID)
	assert.Contains(t, ids, clinicB.ID)
	assert.NotContains(t, ids, clinicC.ID, "ソフト削除された割当のクリニックは含まれない")
}

func TestClinicRepository_FindByID(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	clinic := makeClinicFixture(t, db, "単件取得テスト")

	t.Run("存在するIDは取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinic.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, clinic.Name, got.Name)
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		got, err := repo.FindByID(ctx, 999888001)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})
}

func TestClinicRepository_FindCompany(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	// companies は永続テーブルのため、既存シードが無ければ最小限の1件を用意する。
	var existing int64
	db.Model(&model.Company{}).Count(&existing)
	if existing == 0 {
		require.NoError(t, db.WithContext(ctx).Create(&model.Company{Name: "初期法人"}).Error)
	}

	got, err := repo.FindCompany(ctx)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.NotEmpty(t, got.Name)
}

func TestClinicRepository_Create(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	company := &model.Company{Name: "Create用法人"}
	require.NoError(t, db.WithContext(ctx).Create(company).Error)

	clinic := &model.Clinic{CompanyID: company.ID, Name: "新規作成クリニック"}
	err := repo.Create(ctx, clinic)
	require.NoError(t, err)
	assert.NotZero(t, clinic.ID)

	got, err := repo.FindByID(ctx, clinic.ID)
	require.NoError(t, err)
	assert.Equal(t, "新規作成クリニック", got.Name)
	assert.InDelta(t, 0.10, got.StandardTaxRate, 0.0001, "未指定時は DB デフォルト税率が適用される")
}

func TestClinicRepository_Update(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	clinic := makeClinicFixture(t, db, "更新前クリニック")

	t.Run("成功", func(t *testing.T) {
		err := repo.Update(ctx, clinic.ID, map[string]any{"name": "更新後クリニック"})
		require.NoError(t, err)
		got, err := repo.FindByID(ctx, clinic.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後クリニック", got.Name)
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		err := repo.Update(ctx, 999888002, map[string]any{"name": "x"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})
}

func TestClinicRepository_Delete(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	t.Run("子データのないクリニックは削除できる", func(t *testing.T) {
		clinic := makeClinicFixture(t, db, "削除対象クリニック")
		err := repo.Delete(ctx, clinic.ID)
		require.NoError(t, err)

		_, err = repo.FindByID(ctx, clinic.ID)
		assert.True(t, apperrors.IsNotFound(err), "削除後は NotFound になるべき")
	})

	t.Run("ソフト削除済みPermissionGroupは事前にハード削除されクリニックを削除できる", func(t *testing.T) {
		clinic := makeClinicFixture(t, db, "PG付き削除対象クリニック")
		pg := &model.PermissionGroup{ClinicID: clinic.ID, Name: "削除予定グループ"}
		require.NoError(t, db.WithContext(ctx).Create(pg).Error)
		require.NoError(t, db.WithContext(ctx).Delete(pg).Error) // ソフト削除

		err := repo.Delete(ctx, clinic.ID)
		require.NoError(t, err)

		var count int64
		db.Unscoped().Model(&model.PermissionGroup{}).Where("id = ?", pg.ID).Count(&count)
		assert.Equal(t, int64(0), count, "ソフト削除済み PermissionGroup は物理削除されている")
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, 999888003)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})
}

func TestClinicRepository_CountOwnersByClinicID(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	clinicA := makeClinicFixture(t, db, "飼主数A")
	clinicB := makeClinicFixture(t, db, "飼主数B")

	makeOwner(t, db, clinicA.ID, "飼主1")
	makeOwner(t, db, clinicA.ID, "飼主2")
	deletedOwner := makeOwner(t, db, clinicA.ID, "削除済み飼主")
	require.NoError(t, db.WithContext(ctx).Delete(deletedOwner).Error)
	makeOwner(t, db, clinicB.ID, "別クリニック飼主")

	got, err := repo.CountOwnersByClinicID(ctx, clinicA.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), got, "ソフト削除・別クリニックを除外して2件")
}

func TestClinicRepository_CountStaffByClinicID(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	clinicA := makeClinicFixture(t, db, "スタッフ数A")
	clinicB := makeClinicFixture(t, db, "スタッフ数B")

	staff1 := &model.Staff{ClinicID: clinicB.ID, Name: "スタッフ1", StaffType: model.StaffTypeDoctor}
	require.NoError(t, db.WithContext(ctx).Create(staff1).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{StaffID: staff1.ID, ClinicID: clinicA.ID, IsMain: true}).Error)

	staff2 := &model.Staff{ClinicID: clinicB.ID, Name: "スタッフ2", StaffType: model.StaffTypeNurse}
	require.NoError(t, db.WithContext(ctx).Create(staff2).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{StaffID: staff2.ID, ClinicID: clinicA.ID}).Error)

	// ソフト削除された割当は除外される
	staff3 := &model.Staff{ClinicID: clinicB.ID, Name: "スタッフ3(割当ソフト削除)", StaffType: model.StaffTypeNurse}
	require.NoError(t, db.WithContext(ctx).Create(staff3).Error)
	assign3 := &model.StaffClinicAssignment{StaffID: staff3.ID, ClinicID: clinicA.ID}
	require.NoError(t, db.WithContext(ctx).Create(assign3).Error)
	require.NoError(t, db.WithContext(ctx).Delete(assign3).Error)

	// 別クリニックのみに割当のスタッフは対象外
	staff4 := &model.Staff{ClinicID: clinicB.ID, Name: "スタッフ4(別クリニックのみ)", StaffType: model.StaffTypeNurse}
	require.NoError(t, db.WithContext(ctx).Create(staff4).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{StaffID: staff4.ID, ClinicID: clinicB.ID, IsMain: true}).Error)

	got, err := repo.CountStaffByClinicID(ctx, clinicA.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), got)
}

// Fixed (#236 root cause, 2026-07-13): model.ClinicSettings now carries explicit gorm
// "type:time"/"column:" tags matching backend/migrations/001_init.sql, so AutoMigrate no
// longer fails and this table dependency check runs normally. See
// TestClinicSettingsRepository_* in clinic_settings_repository_test.go for the same fix.
func TestClinicRepository_CountBlockingReferencesByClinicID(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	t.Run("依存データが無ければ空スライス", func(t *testing.T) {
		clinic := makeClinicFixture(t, db, "依存なしクリニック")
		got, err := repo.CountBlockingReferencesByClinicID(ctx, clinic.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("会計データがあれば件数付きでラベルが返る", func(t *testing.T) {
		clinic := makeClinicFixture(t, db, "会計依存クリニック")
		makeBilling(t, db, clinic.ID, nil, nil, 1000, model.BillingStatusWaiting, time.Now())
		makeBilling(t, db, clinic.ID, nil, nil, 2000, model.BillingStatusWaiting, time.Now())

		got, err := repo.CountBlockingReferencesByClinicID(ctx, clinic.ID)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "会計", got[0].Label)
		assert.Equal(t, int64(2), got[0].Count)
	})

	t.Run("ソフト削除された会計は除外される(P2)", func(t *testing.T) {
		clinic := makeClinicFixture(t, db, "会計ソフト削除クリニック")
		makeBilling(t, db, clinic.ID, nil, nil, 1000, model.BillingStatusWaiting, time.Now())

		var b model.Billing
		require.NoError(t, db.WithContext(ctx).Where("clinic_id = ?", clinic.ID).First(&b).Error)
		require.NoError(t, db.WithContext(ctx).Delete(&b).Error)

		got, err := repo.CountBlockingReferencesByClinicID(ctx, clinic.ID)
		require.NoError(t, err)
		assert.Empty(t, got, "唯一の会計がソフト削除されれば依存なしになる")
	})

	// Fixed (#236 root cause, 2026-07-13): see comment above TestClinicRepository_CountBlockingReferencesByClinicID.
	t.Run("clinic_settingsはソフトデリート対象外テーブルとして検出される", func(t *testing.T) {
		clinic := makeClinicFixture(t, db, "医院設定依存クリニック")
		require.NoError(t, db.WithContext(ctx).Create(&model.ClinicSettings{ClinicID: clinic.ID}).Error)

		got, err := repo.CountBlockingReferencesByClinicID(ctx, clinic.ID)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "医院設定", got[0].Label)
		assert.Equal(t, int64(1), got[0].Count)
	})
}
