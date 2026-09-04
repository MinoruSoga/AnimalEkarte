package pet

// animal_species_repository_test.go — AnimalSpeciesRepository 統合テスト。
//
// 保護する不変条件:
//   - AnimalSpecies はクリニック横断のグローバルマスタ（clinic_id を持たない）。
//   - FindAll は sort_order ASC, name ASC で返す。
//   - FindByID / Update / Delete は対象なしで NotFound を返す。
//   - AnimalSpecies にソフトデリート列は無いため Delete は物理削除。
//   - Reorder は ids の順序で sort_order を 1 始まりで振り直す。存在しない id は
//     InvalidInput エラーで失敗する（NotFound ではない）。
//   - FindAll/FindByID/Create/Update/Delete/Reorder は persistence.DBOrTx 経由で
//     ambient tx に参加し、後続失敗時に書き込みがロールバックされる
//     （BE-ACT-ANIMAL-SPECIES-DBORTX / POC-07）。

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

var errAnimalSpeciesAmbientTxSentinel = errors.New("simulated post-write failure in ambient tx")

// setupAnimalSpeciesTestDB は animal_species テーブルを用意する。
func setupAnimalSpeciesTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}))
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func TestAnimalSpeciesRepository_FindAll(t *testing.T) {
	db := setupAnimalSpeciesTestDB(t)
	repo := NewAnimalSpeciesRepository(db)
	ctx := context.Background()

	// sort_order が同じ場合は name ASC、sort_order が異なる場合は sort_order ASC が優先される。
	dog := &model.AnimalSpecies{Name: "犬", SortOrder: 2}
	cat := &model.AnimalSpecies{Name: "猫", SortOrder: 1}
	bird := &model.AnimalSpecies{Name: "鳥", SortOrder: 1}
	require.NoError(t, db.WithContext(ctx).Create(dog).Error)
	require.NoError(t, db.WithContext(ctx).Create(cat).Error)
	require.NoError(t, db.WithContext(ctx).Create(bird).Error)

	got, err := repo.FindAll(ctx)
	require.NoError(t, err)
	require.Len(t, got, 3)
	// sort_order=1 の中では name ASC (鳥 < 猫 は Unicode 順ではないため実際の DB 照合順序に依存する可能性があるが、
	// 少なくとも sort_order=1 の2件が sort_order=2 の1件より先に来ることを検証する。
	assert.Equal(t, 1, got[0].SortOrder)
	assert.Equal(t, 1, got[1].SortOrder)
	assert.Equal(t, 2, got[2].SortOrder)
	assert.Equal(t, "犬", got[2].Name)
}

func TestAnimalSpeciesRepository_FindByID(t *testing.T) {
	db := setupAnimalSpeciesTestDB(t)
	repo := NewAnimalSpeciesRepository(db)
	ctx := context.Background()

	species := &model.AnimalSpecies{Name: "うさぎ"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)

	t.Run("found", func(t *testing.T) {
		got, err := repo.FindByID(ctx, species.ID)
		require.NoError(t, err)
		assert.Equal(t, "うさぎ", got.Name)
	})

	t.Run("not found for nonexistent id", func(t *testing.T) {
		got, err := repo.FindByID(ctx, 999999)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestAnimalSpeciesRepository_Create(t *testing.T) {
	db := setupAnimalSpeciesTestDB(t)
	repo := NewAnimalSpeciesRepository(db)
	ctx := context.Background()

	species := &model.AnimalSpecies{Name: "ハムスター"}
	require.NoError(t, repo.Create(ctx, species))
	assert.NotZero(t, species.ID)

	got, err := repo.FindByID(ctx, species.ID)
	require.NoError(t, err)
	assert.Equal(t, "ハムスター", got.Name)
}

// BUG-455-S2: gorm default:true omits zero bools from INSERT; explicit false must survive.
func TestAnimalSpeciesRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupAnimalSpeciesTestDB(t)
	repo := NewAnimalSpeciesRepository(db)
	ctx := context.Background()

	species := &model.AnimalSpecies{Name: "inactive species", IsActive: false}
	require.NoError(t, repo.Create(ctx, species))
	require.NotZero(t, species.ID)
	assert.False(t, species.IsActive, "in-memory struct must keep false after Create")

	got, err := repo.FindByID(ctx, species.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive, "DB read-back must keep explicit false")

	var rawActive bool
	require.NoError(t, db.WithContext(ctx).
		Model(&model.AnimalSpecies{}).
		Select("is_active").
		Where("id = ?", species.ID).
		Scan(&rawActive).Error)
	assert.False(t, rawActive, "raw is_active column must be false")
}

func TestAnimalSpeciesRepository_Create_IsActiveTruePersists(t *testing.T) {
	db := setupAnimalSpeciesTestDB(t)
	repo := NewAnimalSpeciesRepository(db)
	ctx := context.Background()

	species := &model.AnimalSpecies{Name: "active species", IsActive: true}
	require.NoError(t, repo.Create(ctx, species))
	assert.True(t, species.IsActive)

	got, err := repo.FindByID(ctx, species.ID)
	require.NoError(t, err)
	assert.True(t, got.IsActive)

	var rawActive bool
	require.NoError(t, db.WithContext(ctx).
		Model(&model.AnimalSpecies{}).
		Select("is_active").
		Where("id = ?", species.ID).
		Scan(&rawActive).Error)
	assert.True(t, rawActive, "raw is_active column must be true")
}

func TestAnimalSpeciesRepository_Update(t *testing.T) {
	db := setupAnimalSpeciesTestDB(t)
	repo := NewAnimalSpeciesRepository(db)
	ctx := context.Background()

	species := &model.AnimalSpecies{Name: "更新前"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)

	t.Run("updates successfully", func(t *testing.T) {
		got, err := repo.Update(ctx, species.ID, map[string]any{"name": "更新後"})
		require.NoError(t, err)
		assert.Equal(t, "更新後", got.Name)
	})

	t.Run("not found for nonexistent id", func(t *testing.T) {
		got, err := repo.Update(ctx, 999999, map[string]any{"name": "x"})
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestAnimalSpeciesRepository_Delete(t *testing.T) {
	db := setupAnimalSpeciesTestDB(t)
	repo := NewAnimalSpeciesRepository(db)
	ctx := context.Background()

	species := &model.AnimalSpecies{Name: "削除対象"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)

	t.Run("not found for nonexistent id", func(t *testing.T) {
		err := repo.Delete(ctx, 999999)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("deletes the row physically (no soft-delete column)", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, species.ID))

		var count int64
		require.NoError(t, db.Model(&model.AnimalSpecies{}).Where("id = ?", species.ID).Count(&count).Error)
		assert.Equal(t, int64(0), count, "AnimalSpecies has no deleted_at column, so Delete must remove the row entirely")
	})

	t.Run("not found for already-deleted id", func(t *testing.T) {
		err := repo.Delete(ctx, species.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("ペットが参照していれば Conflict で行は残る", func(t *testing.T) {
		require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Owner{}, &model.Pet{}))
		used := &model.AnimalSpecies{Name: "使用中動物種"}
		require.NoError(t, db.WithContext(ctx).Create(used).Error)
		owner := testdb.MakeTestOwner(t, db, 1, "動物種削除飼主")
		pet := &model.Pet{ClinicID: 1, OwnerID: owner.ID, AnimalSpeciesID: used.ID, Name: "参照ペット"}
		require.NoError(t, db.WithContext(ctx).Create(pet).Error)

		err := repo.Delete(ctx, used.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)

		got, findErr := repo.FindByID(ctx, used.ID)
		require.NoError(t, findErr)
		assert.Equal(t, used.ID, got.ID)
	})
}

func TestAnimalSpeciesRepository_Reorder(t *testing.T) {
	db := setupAnimalSpeciesTestDB(t)
	repo := NewAnimalSpeciesRepository(db)
	ctx := context.Background()

	first := &model.AnimalSpecies{Name: "A", SortOrder: 10}
	second := &model.AnimalSpecies{Name: "B", SortOrder: 20}
	third := &model.AnimalSpecies{Name: "C", SortOrder: 30}
	require.NoError(t, db.WithContext(ctx).Create(first).Error)
	require.NoError(t, db.WithContext(ctx).Create(second).Error)
	require.NoError(t, db.WithContext(ctx).Create(third).Error)

	t.Run("reorders ids to 1-based sort_order in the given order", func(t *testing.T) {
		require.NoError(t, repo.Reorder(ctx, []uint64{third.ID, first.ID, second.ID}))

		gotThird, err := repo.FindByID(ctx, third.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, gotThird.SortOrder)

		gotFirst, err := repo.FindByID(ctx, first.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, gotFirst.SortOrder)

		gotSecond, err := repo.FindByID(ctx, second.ID)
		require.NoError(t, err)
		assert.Equal(t, 3, gotSecond.SortOrder)
	})

	t.Run("nonexistent id in the list returns InvalidInput error and rolls back atomically", func(t *testing.T) {
		// Capture orders before the partial-update attempt so a non-atomic
		// reorder cannot leave the first id already reassigned.
		beforeFirst, err := repo.FindByID(ctx, first.ID)
		require.NoError(t, err)
		beforeSecond, err := repo.FindByID(ctx, second.ID)
		require.NoError(t, err)
		beforeThird, err := repo.FindByID(ctx, third.ID)
		require.NoError(t, err)

		err = repo.Reorder(ctx, []uint64{first.ID, 999999})
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))

		gotFirst, err := repo.FindByID(ctx, first.ID)
		require.NoError(t, err)
		assert.Equal(t, beforeFirst.SortOrder, gotFirst.SortOrder, "partial reorder must roll back")
		gotSecond, err := repo.FindByID(ctx, second.ID)
		require.NoError(t, err)
		assert.Equal(t, beforeSecond.SortOrder, gotSecond.SortOrder)
		gotThird, err := repo.FindByID(ctx, third.ID)
		require.NoError(t, err)
		assert.Equal(t, beforeThird.SortOrder, gotThird.SortOrder)
	})
}

// ---- Ambient transaction participation (BE-ACT-ANIMAL-SPECIES-DBORTX) ----

func TestAnimalSpeciesRepository_Create_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupAnimalSpeciesTestDB(t)
	repo := NewAnimalSpeciesRepository(db)
	ctx := context.Background()

	tx := persistence.NewTransactor(db)
	err := tx.WithTx(ctx, func(txCtx context.Context) error {
		species := &model.AnimalSpecies{Name: "ambient-create-rollback", IsActive: true}
		if err := repo.Create(txCtx, species); err != nil {
			return err
		}
		require.NotZero(t, species.ID)
		// Prove the write is visible inside the ambient transaction.
		got, err := repo.FindByID(txCtx, species.ID)
		require.NoError(t, err)
		assert.Equal(t, "ambient-create-rollback", got.Name)
		return errAnimalSpeciesAmbientTxSentinel
	})
	require.Error(t, err)
	require.ErrorIs(t, err, errAnimalSpeciesAmbientTxSentinel)

	var count int64
	require.NoError(t, db.WithContext(ctx).
		Model(&model.AnimalSpecies{}).
		Where("name = ?", "ambient-create-rollback").
		Count(&count).Error)
	assert.Equal(t, int64(0), count, "Create must participate in ambient tx and roll back")
}

func TestAnimalSpeciesRepository_Update_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupAnimalSpeciesTestDB(t)
	repo := NewAnimalSpeciesRepository(db)
	ctx := context.Background()

	species := &model.AnimalSpecies{Name: "ambient-update-before", SortOrder: 5}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)

	tx := persistence.NewTransactor(db)
	err := tx.WithTx(ctx, func(txCtx context.Context) error {
		got, err := repo.Update(txCtx, species.ID, map[string]any{"name": "ambient-update-after"})
		if err != nil {
			return err
		}
		assert.Equal(t, "ambient-update-after", got.Name)
		return errAnimalSpeciesAmbientTxSentinel
	})
	require.Error(t, err)
	require.ErrorIs(t, err, errAnimalSpeciesAmbientTxSentinel)

	reloaded, err := repo.FindByID(ctx, species.ID)
	require.NoError(t, err)
	assert.Equal(t, "ambient-update-before", reloaded.Name, "Update must roll back with ambient tx")
}

func TestAnimalSpeciesRepository_Delete_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupAnimalSpeciesTestDB(t)
	repo := NewAnimalSpeciesRepository(db)
	ctx := context.Background()

	species := &model.AnimalSpecies{Name: "ambient-delete-target"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)

	tx := persistence.NewTransactor(db)
	err := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.Delete(txCtx, species.ID); err != nil {
			return err
		}
		return errAnimalSpeciesAmbientTxSentinel
	})
	require.Error(t, err)
	require.ErrorIs(t, err, errAnimalSpeciesAmbientTxSentinel)

	reloaded, err := repo.FindByID(ctx, species.ID)
	require.NoError(t, err)
	assert.Equal(t, "ambient-delete-target", reloaded.Name, "Delete must roll back with ambient tx")
}

func TestAnimalSpeciesRepository_Reorder_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupAnimalSpeciesTestDB(t)
	repo := NewAnimalSpeciesRepository(db)
	ctx := context.Background()

	first := &model.AnimalSpecies{Name: "reorder-tx-A", SortOrder: 10}
	second := &model.AnimalSpecies{Name: "reorder-tx-B", SortOrder: 20}
	require.NoError(t, db.WithContext(ctx).Create(first).Error)
	require.NoError(t, db.WithContext(ctx).Create(second).Error)

	tx := persistence.NewTransactor(db)
	err := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.Reorder(txCtx, []uint64{second.ID, first.ID}); err != nil {
			return err
		}
		// Visible inside ambient tx.
		gotSecond, err := repo.FindByID(txCtx, second.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, gotSecond.SortOrder)
		return errAnimalSpeciesAmbientTxSentinel
	})
	require.Error(t, err)
	require.ErrorIs(t, err, errAnimalSpeciesAmbientTxSentinel)

	gotFirst, err := repo.FindByID(ctx, first.ID)
	require.NoError(t, err)
	assert.Equal(t, 10, gotFirst.SortOrder, "Reorder must roll back with ambient tx")
	gotSecond, err := repo.FindByID(ctx, second.ID)
	require.NoError(t, err)
	assert.Equal(t, 20, gotSecond.SortOrder, "Reorder must roll back with ambient tx")
}
