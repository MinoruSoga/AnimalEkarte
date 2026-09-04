package staff_test

// occupation_repository_test.go
// occupation_repository.go の実 DB 結合テスト。
// 既存の makeOccupation / makeStaffWithOccupation（staff_occupation_preload_clinic_isolation_test.go）
// および seedClinicsForFK / makeStaffClinicAssignment（staff_preload_clinic_isolation_test.go）を再利用する。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	. "github.com/animal-ekarte/backend/internal/staff"
)

// setupOccupationRepoTestDB は occupation_repository のテストに必要なテーブルを整備する。
// CountUsageByOccupationID は staff_clinic_assignments を JOIN するため staffs / assignments も migrate する。
func setupOccupationRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Occupation{}, &model.Staff{}, &model.StaffClinicAssignment{},
	))
	require.NoError(t, db.Exec("TRUNCATE TABLE staff_clinic_assignments, staffs, occupations CASCADE").Error)
	return db
}

func TestOccupationRepository_Create_FindByID(t *testing.T) {
	db := setupOccupationRepoTestDB(t)
	repo := NewOccupationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("作成した職種を同一クリニックで取得できる", func(t *testing.T) {
		occ := &model.Occupation{ClinicID: clinicA, Name: "獣医師"}
		require.NoError(t, repo.Create(ctx, occ))
		require.NotZero(t, occ.ID)

		got, err := repo.FindByID(ctx, clinicA, occ.ID)
		require.NoError(t, err)
		assert.Equal(t, "獣医師", got.Name)
	})

	t.Run("explicit is_active false persists (BUG-455-S7)", func(t *testing.T) {
		occ := &model.Occupation{ClinicID: clinicA, Name: "inactive occupation", IsActive: false}
		require.NoError(t, repo.Create(ctx, occ))
		assert.False(t, occ.IsActive)

		got, err := repo.FindByID(ctx, clinicA, occ.ID)
		require.NoError(t, err)
		assert.False(t, got.IsActive)

		var raw bool
		require.NoError(t, db.WithContext(ctx).
			Model(&model.Occupation{}).
			Select("is_active").
			Where("id = ?", occ.ID).
			Scan(&raw).Error)
		assert.False(t, raw, "raw is_active must be false")
	})

	t.Run("別クリニックからは取得できずNotFoundを返す", func(t *testing.T) {
		occ := makeOccupation(t, db, clinicA, "看護師")
		_, err := repo.FindByID(ctx, clinicB, occ.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "別クリニックの職種取得はNotFoundでラップされるべき")
	})

	t.Run("存在しないIDはNotFoundを返す", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestOccupationRepository_FindAll(t *testing.T) {
	db := setupOccupationRepoTestDB(t)
	repo := NewOccupationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("クリニックで隔離され sort_order/name の昇順で返る", func(t *testing.T) {
		occB := makeOccupation(t, db, clinicB, "医院Bの職種")
		occA2 := &model.Occupation{ClinicID: clinicA, Name: "B職種", SortOrder: 2}
		require.NoError(t, db.WithContext(ctx).Create(occA2).Error)
		occA1 := &model.Occupation{ClinicID: clinicA, Name: "A職種", SortOrder: 1}
		require.NoError(t, db.WithContext(ctx).Create(occA1).Error)

		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 2, "clinic Aの職種のみ返るべき")
		assert.Equal(t, occA1.ID, got[0].ID, "sort_order昇順で先頭")
		assert.Equal(t, occA2.ID, got[1].ID)
		for _, o := range got {
			assert.NotEqual(t, occB.ID, o.ID, "別クリニックの職種が混入してはならない")
		}
	})

	t.Run("ソフトデリート済みは一覧から除外されるがレコードは残る", func(t *testing.T) {
		db2 := setupOccupationRepoTestDB(t)
		repo2 := NewOccupationRepository(db2)
		occ := makeOccupation(t, db2, clinicA, "削除予定職種")

		require.NoError(t, repo2.Delete(ctx, clinicA, occ.ID))

		got, err := repo2.FindAll(ctx, clinicA)
		require.NoError(t, err)
		assert.Len(t, got, 0, "ソフトデリート済みは一覧から除外される")

		var raw model.Occupation
		require.NoError(t, db2.WithContext(ctx).Unscoped().First(&raw, occ.ID).Error, "行自体はDBに残っている")
		assert.NotNil(t, raw.DeletedAt)
	})
}

func TestOccupationRepository_Update(t *testing.T) {
	db := setupOccupationRepoTestDB(t)
	repo := NewOccupationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("同一クリニックの更新は反映される", func(t *testing.T) {
		occ := makeOccupation(t, db, clinicA, "更新前職種")
		got, err := repo.Update(ctx, clinicA, occ.ID, map[string]any{"name": "更新後職種"})
		require.NoError(t, err)
		assert.Equal(t, "更新後職種", got.Name)
	})

	t.Run("別クリニックの更新はNotFound", func(t *testing.T) {
		occ := makeOccupation(t, db, clinicA, "他院からの更新対象")
		_, err := repo.Update(ctx, clinicB, occ.ID, map[string]any{"name": "越境更新"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDの更新はNotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicA, 999999, map[string]any{"name": "存在しない"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestOccupationRepository_Delete(t *testing.T) {
	db := setupOccupationRepoTestDB(t)
	repo := NewOccupationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("同一クリニックの削除は成功しその後取得できない", func(t *testing.T) {
		occ := makeOccupation(t, db, clinicA, "削除対象職種")
		require.NoError(t, repo.Delete(ctx, clinicA, occ.ID))

		_, err := repo.FindByID(ctx, clinicA, occ.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックの削除はNotFoundで対象データは残る", func(t *testing.T) {
		occ := makeOccupation(t, db, clinicA, "越境削除対象")
		err := repo.Delete(ctx, clinicB, occ.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, occ.ID)
		require.NoError(t, err, "越境削除の試行では実際には削除されていないはず")
		assert.Equal(t, occ.ID, got.ID)
	})

	t.Run("存在しないIDの削除はNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("使用中なら Conflict で行は残る", func(t *testing.T) {
		seedClinicsForFK(t, db, clinicA)
		occ := makeOccupation(t, db, clinicA, "使用中職種")
		staff := makeStaffWithOccupation(t, db, clinicA, occ.ID, "使用中職種スタッフ")
		makeStaffClinicAssignment(t, db, staff.ID, clinicA)

		err := repo.Delete(ctx, clinicA, occ.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)

		got, findErr := repo.FindByID(ctx, clinicA, occ.ID)
		require.NoError(t, findErr)
		assert.Equal(t, occ.ID, got.ID)
	})
}

func TestOccupationRepository_Reorder(t *testing.T) {
	db := setupOccupationRepoTestDB(t)
	repo := NewOccupationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("指定した順序でsort_orderが1始まりに更新される", func(t *testing.T) {
		occ1 := makeOccupation(t, db, clinicA, "職種1")
		occ2 := makeOccupation(t, db, clinicA, "職種2")
		occ3 := makeOccupation(t, db, clinicA, "職種3")

		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{occ3.ID, occ1.ID, occ2.ID}))

		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 3)
		byID := make(map[uint64]model.Occupation, 3)
		for _, o := range got {
			byID[o.ID] = o
		}
		assert.Equal(t, 1, byID[occ3.ID].SortOrder)
		assert.Equal(t, 2, byID[occ1.ID].SortOrder)
		assert.Equal(t, 3, byID[occ2.ID].SortOrder)
	})

	t.Run("別クリニックのIDが混ざるとエラーになる", func(t *testing.T) {
		occA := makeOccupation(t, db, clinicA, "医院A職種")
		occB := makeOccupation(t, db, clinicB, "医院B職種")

		err := repo.Reorder(ctx, clinicA, []uint64{occA.ID, occB.ID})
		require.Error(t, err, "別クリニックのIDを含むReorderは失敗するべき")
	})
}

func TestOccupationRepository_CountUsageByOccupationID(t *testing.T) {
	db := setupOccupationRepoTestDB(t)
	repo := NewOccupationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	t.Run("clinic A に配属されたスタッフのみカウントする", func(t *testing.T) {
		occ := makeOccupation(t, db, clinicA, "カウント対象職種")

		staffInA := makeStaffWithOccupation(t, db, clinicA, occ.ID, "医院Aスタッフ")
		makeStaffClinicAssignment(t, db, staffInA.ID, clinicA)

		// 主クリニックはAだが実際の配属はBのみ（JOINのclinic_idはclinic_assignmentsで判定）
		staffAssignedElsewhere := makeStaffWithOccupation(t, db, clinicA, occ.ID, "配属先違いスタッフ")
		makeStaffClinicAssignment(t, db, staffAssignedElsewhere.ID, clinicB)

		count, err := repo.CountUsageByOccupationID(ctx, clinicA, occ.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "clinic Aに配属済みのスタッフのみカウントされるべき")
	})

	t.Run("未使用の職種は0を返す", func(t *testing.T) {
		unused := makeOccupation(t, db, clinicA, "未使用職種")
		count, err := repo.CountUsageByOccupationID(ctx, clinicA, unused.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("ソフトデリート済みスタッフはカウントに含まれない", func(t *testing.T) {
		occ := makeOccupation(t, db, clinicA, "削除スタッフ用職種")
		staff := makeStaffWithOccupation(t, db, clinicA, occ.ID, "退職予定スタッフ")
		makeStaffClinicAssignment(t, db, staff.ID, clinicA)

		require.NoError(t, db.WithContext(ctx).Delete(&model.Staff{}, staff.ID).Error)

		count, err := repo.CountUsageByOccupationID(ctx, clinicA, occ.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count, "ソフトデリート済みスタッフは参照カウントに含まれない")
	})
}
