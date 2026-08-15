package persistence

// scope_helpers_integration_test.go — ReorderByClinicID / ReorderGlobal の統合テスト。
//
// 保護する不変条件:
//   - ReorderByClinicID / ReorderGlobal は渡された ids の順序通りに sort_order (1始まり) を設定する。
//   - ReorderByClinicID は ClinicScope を通すため、別クリニックの id を混ぜると
//     "not found in this clinic" エラーとなりトランザクション全体がロールバックされる。
//   - 存在しない id を含む場合はエラーとなり、既に反映されかけた更新も含めて全てロールバックされる。
//
// テスト対象モデルは scope.go の呼び出し元と同型の既存マスタ
// （model.ChiefComplaintType / model.AnimalSpecies）を再利用する。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupHelpersTestDB は ReorderByClinicID / ReorderGlobal 検証用にテーブルを用意する。
func setupHelpersTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ChiefComplaintType{}, &model.AnimalSpecies{}))
	require.NoError(t, db.Exec("TRUNCATE TABLE chief_complaint_types CASCADE").Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE animal_species CASCADE").Error)
	return db
}

func TestReorderByClinicID(t *testing.T) {
	db := setupHelpersTestDB(t)
	ctx := context.Background()

	t.Run("渡されたid順にsort_orderを1始まりで設定する", func(t *testing.T) {
		const clinicID = uint64(1)
		a := &model.ChiefComplaintType{ClinicID: clinicID, Name: "A"}
		b := &model.ChiefComplaintType{ClinicID: clinicID, Name: "B"}
		c := &model.ChiefComplaintType{ClinicID: clinicID, Name: "C"}
		require.NoError(t, db.Create(a).Error)
		require.NoError(t, db.Create(b).Error)
		require.NoError(t, db.Create(c).Error)

		require.NoError(t, ReorderByClinicID(ctx, db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicID, []uint64{c.ID, a.ID, b.ID}, "sort_order"))

		var reloadedA, reloadedB, reloadedC model.ChiefComplaintType
		require.NoError(t, db.First(&reloadedA, a.ID).Error)
		require.NoError(t, db.First(&reloadedB, b.ID).Error)
		require.NoError(t, db.First(&reloadedC, c.ID).Error)
		assert.Equal(t, 1, reloadedC.SortOrder, "先頭に指定したcがsort_order=1")
		assert.Equal(t, 2, reloadedA.SortOrder)
		assert.Equal(t, 3, reloadedB.SortOrder)
	})

	t.Run("別クリニックのidを含むとエラーになり全体がロールバックされる", func(t *testing.T) {
		const clinicA, clinicB = uint64(10), uint64(20)
		own := &model.ChiefComplaintType{ClinicID: clinicA, Name: "own", SortOrder: 99}
		foreign := &model.ChiefComplaintType{ClinicID: clinicB, Name: "foreign", SortOrder: 99}
		require.NoError(t, db.Create(own).Error)
		require.NoError(t, db.Create(foreign).Error)

		err := ReorderByClinicID(ctx, db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicA, []uint64{own.ID, foreign.ID}, "sort_order")
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))

		var reloadedOwn model.ChiefComplaintType
		require.NoError(t, db.First(&reloadedOwn, own.ID).Error)
		assert.Equal(t, 99, reloadedOwn.SortOrder, "先に処理された own.ID の更新もロールバックされ元の値のまま")
	})

	t.Run("存在しないidを含むとエラーになりロールバックされる", func(t *testing.T) {
		const clinicID = uint64(30)
		existing := &model.ChiefComplaintType{ClinicID: clinicID, Name: "existing", SortOrder: 99}
		require.NoError(t, db.Create(existing).Error)

		err := ReorderByClinicID(ctx, db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicID, []uint64{existing.ID, 9_999_999}, "sort_order")
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))

		var reloaded model.ChiefComplaintType
		require.NoError(t, db.First(&reloaded, existing.ID).Error)
		assert.Equal(t, 99, reloaded.SortOrder, "トランザクション全体がロールバックされる")
	})
}

func TestReorderGlobal(t *testing.T) {
	db := setupHelpersTestDB(t)
	ctx := context.Background()

	t.Run("渡されたid順にsort_orderを1始まりで設定する（クリニック横断マスタ）", func(t *testing.T) {
		dog := &model.AnimalSpecies{Name: "犬"}
		cat := &model.AnimalSpecies{Name: "猫"}
		rabbit := &model.AnimalSpecies{Name: "うさぎ"}
		require.NoError(t, db.Create(dog).Error)
		require.NoError(t, db.Create(cat).Error)
		require.NoError(t, db.Create(rabbit).Error)

		require.NoError(t, ReorderGlobal(ctx, db, &model.AnimalSpecies{}, "animal_species", []uint64{rabbit.ID, dog.ID, cat.ID}, "sort_order"))

		var reloadedDog, reloadedCat, reloadedRabbit model.AnimalSpecies
		require.NoError(t, db.First(&reloadedDog, dog.ID).Error)
		require.NoError(t, db.First(&reloadedCat, cat.ID).Error)
		require.NoError(t, db.First(&reloadedRabbit, rabbit.ID).Error)
		assert.Equal(t, 1, reloadedRabbit.SortOrder)
		assert.Equal(t, 2, reloadedDog.SortOrder)
		assert.Equal(t, 3, reloadedCat.SortOrder)
	})

	t.Run("存在しないidを含むとエラーになりロールバックされる", func(t *testing.T) {
		existing := &model.AnimalSpecies{Name: "鳥", SortOrder: 99}
		require.NoError(t, db.Create(existing).Error)

		err := ReorderGlobal(ctx, db, &model.AnimalSpecies{}, "animal_species", []uint64{existing.ID, 9_999_998}, "sort_order")
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))

		var reloaded model.AnimalSpecies
		require.NoError(t, db.First(&reloaded, existing.ID).Error)
		assert.Equal(t, 99, reloaded.SortOrder, "トランザクション全体がロールバックされる")
	})
}

func TestUpdateScopedByID(t *testing.T) {
	db := setupHelpersTestDB(t)
	ctx := context.Background()

	t.Run("clinic内の対象idを更新できる", func(t *testing.T) {
		const clinicID = uint64(1)
		rec := &model.ChiefComplaintType{ClinicID: clinicID, Name: "before"}
		require.NoError(t, db.Create(rec).Error)

		err := UpdateScopedByID(ctx, db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicID, rec.ID, map[string]any{"name": "after"})
		require.NoError(t, err)

		var reloaded model.ChiefComplaintType
		require.NoError(t, db.First(&reloaded, rec.ID).Error)
		assert.Equal(t, "after", reloaded.Name)
	})

	t.Run("存在しないidはWrapNotFoundを返す", func(t *testing.T) {
		const clinicID = uint64(2)
		err := UpdateScopedByID(ctx, db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicID, 9_999_997, map[string]any{"name": "x"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックのidはWrapNotFoundを返し更新されない", func(t *testing.T) {
		const clinicA, clinicB = uint64(11), uint64(12)
		foreign := &model.ChiefComplaintType{ClinicID: clinicB, Name: "foreign"}
		require.NoError(t, db.Create(foreign).Error)

		err := UpdateScopedByID(ctx, db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicA, foreign.ID, map[string]any{"name": "hacked"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		var reloaded model.ChiefComplaintType
		require.NoError(t, db.First(&reloaded, foreign.ID).Error)
		assert.Equal(t, "foreign", reloaded.Name, "別クリニックのレコードは変更されない")
	})
}

func TestFindByIDScoped(t *testing.T) {
	db := setupHelpersTestDB(t)
	ctx := context.Background()

	t.Run("clinic内の対象idを取得できる", func(t *testing.T) {
		const clinicID = uint64(5)
		rec := &model.ChiefComplaintType{ClinicID: clinicID, Name: "target"}
		require.NoError(t, db.Create(rec).Error)

		got, err := FindByIDScoped[model.ChiefComplaintType](ctx, db, "chief_complaint_type", clinicID, rec.ID)
		require.NoError(t, err)
		assert.Equal(t, rec.ID, got.ID)
		assert.Equal(t, "target", got.Name)
	})

	t.Run("存在しないidはNotFoundを返す", func(t *testing.T) {
		const clinicID = uint64(6)
		_, err := FindByIDScoped[model.ChiefComplaintType](ctx, db, "chief_complaint_type", clinicID, 9_999_995)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックのidはNotFoundを返す", func(t *testing.T) {
		const clinicA, clinicB = uint64(31), uint64(32)
		foreign := &model.ChiefComplaintType{ClinicID: clinicB, Name: "foreign3"}
		require.NoError(t, db.Create(foreign).Error)

		_, err := FindByIDScoped[model.ChiefComplaintType](ctx, db, "chief_complaint_type", clinicA, foreign.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestDeleteScopedByID(t *testing.T) {
	db := setupHelpersTestDB(t)
	ctx := context.Background()

	t.Run("clinic内の対象idを削除できる", func(t *testing.T) {
		const clinicID = uint64(3)
		rec := &model.ChiefComplaintType{ClinicID: clinicID, Name: "to-delete"}
		require.NoError(t, db.Create(rec).Error)

		err := DeleteScopedByID(ctx, db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicID, rec.ID)
		require.NoError(t, err)

		var count int64
		require.NoError(t, db.Model(&model.ChiefComplaintType{}).Where("id = ?", rec.ID).Count(&count).Error)
		assert.Equal(t, int64(0), count, "ソフトデリートされ通常クエリでは見えなくなる")
	})

	t.Run("存在しないidはWrapNotFoundを返す", func(t *testing.T) {
		const clinicID = uint64(4)
		err := DeleteScopedByID(ctx, db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicID, 9_999_996)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックのidはWrapNotFoundを返し削除されない", func(t *testing.T) {
		const clinicA, clinicB = uint64(21), uint64(22)
		foreign := &model.ChiefComplaintType{ClinicID: clinicB, Name: "foreign2"}
		require.NoError(t, db.Create(foreign).Error)

		err := DeleteScopedByID(ctx, db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicA, foreign.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		var count int64
		require.NoError(t, db.Model(&model.ChiefComplaintType{}).Where("id = ?", foreign.ID).Count(&count).Error)
		assert.Equal(t, int64(1), count, "別クリニックのレコードは削除されない")
	})
}
