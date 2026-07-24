package medicalrecord

// care_plan_item_repository_test.go — CarePlanItemRepository 統合テスト。
//
// 保護する不変条件:
//   - FindByHospitalizationID / FindByID / Update / Delete は hospitalizations を JOIN して
//     clinic_id でテナント隔離される（care_plan_items 自体は clinic_id 列を持たない）。
//   - FindByHospitalizationID は sort_order ASC で返す。
//   - Medicine / Procedure の Preload は clinic_id 述語付きで、別クリニックのマスタは
//     読み込まれない（P3.1）。
//   - Update / Delete は対象なしで NotFound を返す。
//   - Delete は Unscoped() 指定だが CarePlanItem に deleted_at 列は無いため実質的に物理削除。

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

// setupCarePlanItemTestDB は hospitalizations 経由の clinic_id 隔離と Medicine/Procedure Preload
// を検証するために必要なテーブル一式を用意する。
func setupCarePlanItemTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.Hospitalization{},
		&model.Medicine{},
		&model.Procedure{},
		&model.CarePlanItem{},
	))
	db.Exec("TRUNCATE TABLE care_plan_items CASCADE")
	db.Exec("TRUNCATE TABLE hospitalizations CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	db.Exec("TRUNCATE TABLE medicines CASCADE")
	db.Exec("TRUNCATE TABLE procedures CASCADE")
	return db
}

// makeCarePlanItem はテスト用 CarePlanItem を作成して返す。
func makeCarePlanItem(t *testing.T, db *gorm.DB, hospitalizationID uint64, name string, sortOrder int, medicineID, procedureID *uint64) *model.CarePlanItem {
	t.Helper()
	item := &model.CarePlanItem{
		HospitalizationID: hospitalizationID,
		Type:              model.CarePlanTypeItem,
		Name:              name,
		SortOrder:         sortOrder,
		MedicineID:        medicineID,
		ProcedureID:       procedureID,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(item).Error)
	return item
}

func TestCarePlanItemRepository_FindByHospitalizationID(t *testing.T) {
	db := setupCarePlanItemTestDB(t)
	repo := NewCarePlanItemRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	ownerA := makeTestOwner(t, db, clinicA, "入院飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "入院ポチA")
	hospA := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)
	hospOther := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)

	ownerB := makeTestOwner(t, db, clinicB, "入院飼主B")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "入院ポチB")
	hospB := makeHospitalizationRec(t, db, clinicB, ownerB.ID, petB.ID, nil)

	medA := makeMedicineMaster(t, db, clinicA, "医院Aの薬")
	medB := makeMedicineMaster(t, db, clinicB, "医院Bの薬")
	procA := makeProcedure(t, db, clinicA, "医院Aの処置", model.AnesthesiaTypeNone, false)

	itemSecond := makeCarePlanItem(t, db, hospA.ID, "後・アイテム", 2, nil, nil)
	itemFirst := makeCarePlanItem(t, db, hospA.ID, "先・アイテム", 1, &medA.ID, &procA.ID)
	// 別クリニックの Medicine を紐付けたアイテム（クロステナント Preload 漏洩防止の検証用）
	itemCrossTenantMedicine := makeCarePlanItem(t, db, hospA.ID, "越境薬アイテム", 3, &medB.ID, nil)
	makeCarePlanItem(t, db, hospOther.ID, "別入院のアイテム", 1, nil, nil)
	makeCarePlanItem(t, db, hospB.ID, "別クリニック入院のアイテム", 1, nil, nil)

	t.Run("returns items for the hospitalization ordered by sort_order ASC with Medicine/Procedure preloaded", func(t *testing.T) {
		got, err := repo.FindByHospitalizationID(ctx, clinicA, hospA.ID)
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, itemFirst.ID, got[0].ID)
		require.NotNil(t, got[0].Medicine)
		assert.Equal(t, "医院Aの薬", got[0].Medicine.Name)
		require.NotNil(t, got[0].Procedure)
		assert.Equal(t, "医院Aの処置", got[0].Procedure.Name)

		assert.Equal(t, itemSecond.ID, got[1].ID)
		assert.Equal(t, itemCrossTenantMedicine.ID, got[2].ID)
	})

	t.Run("cross-tenant Medicine is not preloaded (P3.1)", func(t *testing.T) {
		got, err := repo.FindByHospitalizationID(ctx, clinicA, hospA.ID)
		require.NoError(t, err)
		for _, item := range got {
			if item.ID == itemCrossTenantMedicine.ID {
				assert.Nil(t, item.Medicine, "別クリニックの Medicine は Preload されるべきではない")
			}
		}
	})

	t.Run("clinic isolation: different clinic scope returns empty", func(t *testing.T) {
		got, err := repo.FindByHospitalizationID(ctx, clinicB, hospA.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("empty for hospitalization with no items", func(t *testing.T) {
		got, err := repo.FindByHospitalizationID(ctx, clinicA, uint64(999999))
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestCarePlanItemRepository_FindByID(t *testing.T) {
	db := setupCarePlanItemTestDB(t)
	repo := NewCarePlanItemRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	ownerA := makeTestOwner(t, db, clinicA, "飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ポチA")
	hospA := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)
	item := makeCarePlanItem(t, db, hospA.ID, "対象アイテム", 1, nil, nil)

	t.Run("found", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		assert.Equal(t, item.ID, got.ID)
	})

	t.Run("not found for nonexistent id", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, uint64(999999))
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("clinic isolation", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, item.ID)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestCarePlanItemRepository_Create(t *testing.T) {
	db := setupCarePlanItemTestDB(t)
	repo := NewCarePlanItemRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	ownerA := makeTestOwner(t, db, clinicA, "飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ポチA")
	hospA := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)

	item := &model.CarePlanItem{
		HospitalizationID: hospA.ID,
		Type:              model.CarePlanTypeFood,
		Name:              "新規アイテム",
	}
	require.NoError(t, repo.Create(ctx, item))
	assert.NotZero(t, item.ID)
}

func TestCarePlanItemRepository_Update(t *testing.T) {
	db := setupCarePlanItemTestDB(t)
	repo := NewCarePlanItemRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	ownerA := makeTestOwner(t, db, clinicA, "飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ポチA")
	hospA := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)
	item := makeCarePlanItem(t, db, hospA.ID, "更新前", 1, nil, nil)

	t.Run("updates successfully", func(t *testing.T) {
		require.NoError(t, repo.Update(ctx, clinicA, item.ID, map[string]any{"name": "更新後"}))
		got, err := repo.FindByID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後", got.Name)
	})

	t.Run("not found for nonexistent id", func(t *testing.T) {
		err := repo.Update(ctx, clinicA, uint64(999999), map[string]any{"name": "x"})
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("clinic isolation: wrong clinic returns NotFound", func(t *testing.T) {
		err := repo.Update(ctx, clinicB, item.ID, map[string]any{"name": "乗っ取り"})
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestCarePlanItemRepository_Delete(t *testing.T) {
	db := setupCarePlanItemTestDB(t)
	repo := NewCarePlanItemRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	ownerA := makeTestOwner(t, db, clinicA, "飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ポチA")
	hospA := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)
	item := makeCarePlanItem(t, db, hospA.ID, "削除対象", 1, nil, nil)

	t.Run("clinic isolation: wrong clinic cannot delete", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, item.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("deletes the row physically", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, item.ID))

		var count int64
		require.NoError(t, db.Model(&model.CarePlanItem{}).Where("id = ?", item.ID).Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})

	t.Run("not found for already-deleted id", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, item.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
