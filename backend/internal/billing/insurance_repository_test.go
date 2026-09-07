package billing

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

func setupInsuranceRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Insurance{}, &model.AnimalSpecies{}, &model.Owner{}, &model.Pet{}))
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE owners CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	db.Exec("TRUNCATE TABLE insurances CASCADE")
	return db
}

func TestInsuranceRepository_Create_And_FindByID(t *testing.T) {
	db := setupInsuranceRepositoryTestDB(t)
	repo := NewInsuranceRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	t.Run("happy path: Create してから FindByID で取得できる", func(t *testing.T) {
		ins := &model.Insurance{ClinicID: clinicID, Name: "アニペット", CoverageRate: 70, ContactPhone: "03-1234-5678"}
		require.NoError(t, repo.Create(ctx, ins))
		require.NotZero(t, ins.ID)

		got, err := repo.FindByID(ctx, clinicID, ins.ID)
		require.NoError(t, err)
		assert.Equal(t, "アニペット", got.Name)
		assert.Equal(t, 70, got.CoverageRate)
	})

	t.Run("存在しない ID は NotFound エラー", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicID, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("別クリニックからは FindByID できない（clinic_id 隔離）", func(t *testing.T) {
		ins := testdb.MakeInsurance(t, db, clinicID, "医院1限定保険")
		_, err := repo.FindByID(ctx, uint64(999), ins.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "別クリニックからは NotFound であるべき: %v", err)
	})
}

// TestInsuranceRepository_Create_IsActiveFalsePersists is BUG-455-S3:
// gorm:"default:true" omits zero bools from INSERT; explicit false must survive Create.
func TestInsuranceRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupInsuranceRepositoryTestDB(t)
	repo := NewInsuranceRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	ins := &model.Insurance{
		ClinicID: clinicID,
		Name:     "create inactive insurance",
		IsActive: false,
	}
	require.NoError(t, repo.Create(ctx, ins))
	require.NotZero(t, ins.ID)
	assert.False(t, ins.IsActive, "in-memory struct must keep false after Create")

	got, err := repo.FindByID(ctx, clinicID, ins.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive, "DB read-back must keep explicit false")

	var rawActive bool
	require.NoError(t, db.WithContext(ctx).
		Model(&model.Insurance{}).
		Select("is_active").
		Where("id = ?", ins.ID).
		Scan(&rawActive).Error)
	assert.False(t, rawActive, "raw is_active column must be false")
}

// TestInsuranceRepository_Create_IsActiveTruePersists is true-path regression for BUG-455-S3.
func TestInsuranceRepository_Create_IsActiveTruePersists(t *testing.T) {
	db := setupInsuranceRepositoryTestDB(t)
	repo := NewInsuranceRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	active := &model.Insurance{
		ClinicID: clinicID,
		Name:     "create active true insurance",
		IsActive: true,
	}
	require.NoError(t, repo.Create(ctx, active))
	assert.True(t, active.IsActive)

	got, err := repo.FindByID(ctx, clinicID, active.ID)
	require.NoError(t, err)
	assert.True(t, got.IsActive)

	var rawActive bool
	require.NoError(t, db.WithContext(ctx).
		Model(&model.Insurance{}).
		Select("is_active").
		Where("id = ?", active.ID).
		Scan(&rawActive).Error)
	assert.True(t, rawActive)
}

func TestInsuranceRepository_FindAll_ClinicIsolation(t *testing.T) {
	db := setupInsuranceRepositoryTestDB(t)
	repo := NewInsuranceRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	testdb.MakeInsurance(t, db, clinicA, "保険A-1")
	testdb.MakeInsurance(t, db, clinicA, "保険A-2")
	testdb.MakeInsurance(t, db, clinicB, "保険B-1")

	t.Run("clinicA は自院の2件のみ", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("clinicB は自院の1件のみ（混入なし）", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicB)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "保険B-1", got[0].Name)
	})
}

func TestInsuranceRepository_FindAll_ExcludesSoftDeleted(t *testing.T) {
	db := setupInsuranceRepositoryTestDB(t)
	repo := NewInsuranceRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	active := testdb.MakeInsurance(t, db, clinicID, "現役保険")
	deleted := testdb.MakeInsurance(t, db, clinicID, "削除済み保険")
	require.NoError(t, repo.Delete(ctx, clinicID, deleted.ID))

	got, err := repo.FindAll(ctx, clinicID)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, active.ID, got[0].ID)

	var rawCount int64
	db.Unscoped().Model(&model.Insurance{}).Where("id = ?", deleted.ID).Count(&rawCount)
	assert.Equal(t, int64(1), rawCount, "ソフトデリートされた行はDBにまだ存在する")
}

func TestInsuranceRepository_Update(t *testing.T) {
	db := setupInsuranceRepositoryTestDB(t)
	repo := NewInsuranceRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ins := testdb.MakeInsurance(t, db, clinicA, "更新前保険")

	t.Run("同一クリニックでは Update が反映される", func(t *testing.T) {
		name := "更新後保険"
		rate := 50
		got, err := repo.Update(ctx, clinicA, ins.ID, UpdateInsuranceInput{Name: &name, CoverageRate: &rate})
		require.NoError(t, err)
		assert.Equal(t, "更新後保険", got.Name)
		assert.Equal(t, 50, got.CoverageRate)
	})

	t.Run("別クリニックからの Update は NotFound", func(t *testing.T) {
		name := "改ざん試行"
		_, err := repo.Update(ctx, clinicB, ins.ID, UpdateInsuranceInput{Name: &name})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, ins.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後保険", got.Name, "別クリニックからの Update で名称が変わってはならない")
	})

	t.Run("存在しない ID の Update は NotFound", func(t *testing.T) {
		name := "x"
		_, err := repo.Update(ctx, clinicA, 999999, UpdateInsuranceInput{Name: &name})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestInsuranceRepository_Delete(t *testing.T) {
	db := setupInsuranceRepositoryTestDB(t)
	repo := NewInsuranceRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ins := testdb.MakeInsurance(t, db, clinicA, "削除対象保険")

	t.Run("別クリニックからの Delete は NotFound で行が残る", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, ins.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, ins.ID)
		require.NoError(t, err)
		assert.Equal(t, ins.ID, got.ID, "別クリニックからの Delete で行が消えてはならない")
	})

	t.Run("同一クリニックでは Delete が成功し以後 FindByID は NotFound", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, ins.ID))
		_, err := repo.FindByID(ctx, clinicA, ins.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の Delete は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("ペットで使用中なら Conflict で行は残る", func(t *testing.T) {
		used := testdb.MakeInsurance(t, db, clinicA, "使用中削除対象保険")
		owner := testdb.MakeTestOwner(t, db, clinicA, "使用中保険飼主")
		testdb.MakePetWithInsurance(t, db, clinicA, owner.ID, &used.ID, "使用中保険ペット")

		err := repo.Delete(ctx, clinicA, used.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
		assert.Contains(t, err.Error(), "この保険はペット情報で使用中のため削除できません")

		got, findErr := repo.FindByID(ctx, clinicA, used.ID)
		require.NoError(t, findErr)
		assert.Equal(t, used.ID, got.ID)
	})

	t.Run("CountUsage==0 直後にペットが紐づいても削除は失敗する", func(t *testing.T) {
		target := testdb.MakeInsurance(t, db, clinicA, "TOCTOU保険")
		count, err := repo.CountUsageByInsuranceID(ctx, clinicA, target.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), count)

		owner := testdb.MakeTestOwner(t, db, clinicA, "TOCTOU保険飼主")
		testdb.MakePetWithInsurance(t, db, clinicA, owner.ID, &target.ID, "TOCTOU保険ペット")

		err = repo.Delete(ctx, clinicA, target.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
		assert.Contains(t, err.Error(), "この保険はペット情報で使用中のため削除できません")

		got, findErr := repo.FindByID(ctx, clinicA, target.ID)
		require.NoError(t, findErr)
		assert.Equal(t, target.ID, got.ID)
	})
}

func TestInsuranceRepository_Reorder(t *testing.T) {
	db := setupInsuranceRepositoryTestDB(t)
	repo := NewInsuranceRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	i1 := testdb.MakeInsurance(t, db, clinicA, "保険X")
	i2 := testdb.MakeInsurance(t, db, clinicA, "保険Y")
	i3 := testdb.MakeInsurance(t, db, clinicA, "保険Z")

	t.Run("並び順が指定順に更新される", func(t *testing.T) {
		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{i3.ID, i1.ID, i2.ID}))

		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, i3.ID, got[0].ID)
		assert.Equal(t, i1.ID, got[1].ID)
		assert.Equal(t, i2.ID, got[2].ID)
	})

	t.Run("別クリニックの ID を含む Reorder はエラーで中断する", func(t *testing.T) {
		other := testdb.MakeInsurance(t, db, clinicB, "他院保険")
		err := repo.Reorder(ctx, clinicA, []uint64{i1.ID, other.ID})
		require.Error(t, err, "clinicA スコープに存在しない ID を含む Reorder は失敗すべき")
	})
}

func TestInsuranceRepository_CountUsageByInsuranceID(t *testing.T) {
	db := setupInsuranceRepositoryTestDB(t)
	repo := NewInsuranceRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ins := testdb.MakeInsurance(t, db, clinicA, "使用中保険")
	owner := testdb.MakeTestOwner(t, db, clinicA, "飼主A")
	testdb.MakePetWithInsurance(t, db, clinicA, owner.ID, &ins.ID, "ポチ")
	testdb.MakePetWithInsurance(t, db, clinicA, owner.ID, &ins.ID, "タマ")

	t.Run("同一クリニックでは2件カウントされる", func(t *testing.T) {
		count, err := repo.CountUsageByInsuranceID(ctx, clinicA, ins.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("別クリニックからは0件（clinic_id 隔離）", func(t *testing.T) {
		count, err := repo.CountUsageByInsuranceID(ctx, clinicB, ins.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("未使用保険は0件", func(t *testing.T) {
		unused := testdb.MakeInsurance(t, db, clinicA, "未使用保険")
		count, err := repo.CountUsageByInsuranceID(ctx, clinicA, unused.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("ソフトデリート済みペットは数えない（P2）", func(t *testing.T) {
		pet := testdb.MakePetWithInsurance(t, db, clinicA, owner.ID, &ins.ID, "削除対象ペット")
		require.NoError(t, db.WithContext(ctx).Delete(pet).Error)

		count, err := repo.CountUsageByInsuranceID(ctx, clinicA, ins.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count, "ソフトデリート済みペットは使用件数に含まれない")
	})
}
