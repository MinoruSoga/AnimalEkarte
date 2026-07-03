package repository

// helpers_test.go — reorderByClinicID / reorderGlobal の統合テスト。
//
// 保護する不変条件:
//   - reorderByClinicID / reorderGlobal は渡された ids の順序通りに sort_order (1始まり) を設定する。
//   - reorderByClinicID は clinicScope を通すため、別クリニックの id を混ぜると
//     "not found in this clinic" エラーとなりトランザクション全体がロールバックされる。
//   - 存在しない id を含む場合はエラーとなり、既に反映されかけた更新も含めて全てロールバックされる。
//
// テスト対象モデルは helpers.go の呼び出し元と同型の既存マスタ
// （model.ChiefComplaintType / model.AnimalSpecies）を再利用する。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// setupHelpersTestDB は reorderByClinicID / reorderGlobal 検証用にテーブルを用意する。
func setupHelpersTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.ChiefComplaintType{}, &model.AnimalSpecies{}))
	db.Exec("TRUNCATE TABLE chief_complaint_types CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
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

		require.NoError(t, reorderByClinicID(ctx, db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicID, []uint64{c.ID, a.ID, b.ID}))

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

		err := reorderByClinicID(ctx, db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicA, []uint64{own.ID, foreign.ID})
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

		err := reorderByClinicID(ctx, db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicID, []uint64{existing.ID, 9_999_999})
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

		require.NoError(t, reorderGlobal(ctx, db, &model.AnimalSpecies{}, "animal_species", []uint64{rabbit.ID, dog.ID, cat.ID}))

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

		err := reorderGlobal(ctx, db, &model.AnimalSpecies{}, "animal_species", []uint64{existing.ID, 9_999_998})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))

		var reloaded model.AnimalSpecies
		require.NoError(t, db.First(&reloaded, existing.ID).Error)
		assert.Equal(t, 99, reloaded.SortOrder, "トランザクション全体がロールバックされる")
	})
}
