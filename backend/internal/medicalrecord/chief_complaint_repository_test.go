package medicalrecord

// chief_complaint_repository_test.go — ChiefComplaintTypeRepository の統合テスト（実 Postgres テスト DB）。
// happy path・not-found・clinic_id 隔離・ソフトデリート除外・Update/Delete/Reorder を対象とする。
// CountUsageByChiefComplaintTypeID は既知の不具合（inquiries に deleted_at 列が存在しない）に
// より常にエラーを返すため、その実挙動を明示的に検証する。
//
// 移動元: internal/repository/chief_complaint_repository_test.go — BE9-2C roll-up。
//
// makeChiefComplaintType は internal/repository/inquiry_repository_test.go からも使う共有
// ヘルパーだったため、旧パッケージ側にそのまま残置（inquiry は BE9-2D 対象、本バッチ対象外）。
// ここでは medicalrecord 自身のテスト専用ローカルコピーとして再定義する。

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

// setupChiefComplaintTypeRepositoryTestDB は chief_complaint_types テーブルと、
// CountUsageByChiefComplaintTypeID が JOIN する medical_records / inquiries を用意する。
func setupChiefComplaintTypeRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ChiefComplaintType{}, &model.MedicalRecord{}, &model.Inquiry{}))
	db.Exec("TRUNCATE TABLE inquiries CASCADE")
	db.Exec("TRUNCATE TABLE chief_complaint_types CASCADE")
	return db
}

// makeChiefComplaintType はテスト用の ChiefComplaintType を作成して返す。
func makeChiefComplaintType(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.ChiefComplaintType {
	t.Helper()
	c := &model.ChiefComplaintType{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(c).Error)
	return c
}

func TestChiefComplaintTypeRepository_Create_And_FindByID(t *testing.T) {
	db := setupChiefComplaintTypeRepositoryTestDB(t)
	repo := NewChiefComplaintTypeRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	t.Run("happy path: Create してから FindByID で取得できる", func(t *testing.T) {
		c := &model.ChiefComplaintType{ClinicID: clinicID, Name: "嘔吐", Description: "嘔吐症状"}
		require.NoError(t, repo.Create(ctx, c))
		require.NotZero(t, c.ID)

		got, err := repo.FindByID(ctx, clinicID, c.ID)
		require.NoError(t, err)
		assert.Equal(t, "嘔吐", got.Name)
		assert.Equal(t, "嘔吐症状", got.Description)
	})

	t.Run("存在しない ID は NotFound エラー", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicID, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("別クリニックからは FindByID できない（clinic_id 隔離）", func(t *testing.T) {
		c := makeChiefComplaintType(t, db, clinicID, "医院1限定区分")
		_, err := repo.FindByID(ctx, uint64(999), c.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "別クリニックからは NotFound であるべき: %v", err)
	})
}

func TestChiefComplaintTypeRepository_FindAll_ClinicIsolation(t *testing.T) {
	db := setupChiefComplaintTypeRepositoryTestDB(t)
	repo := NewChiefComplaintTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	makeChiefComplaintType(t, db, clinicA, "区分A-1")
	makeChiefComplaintType(t, db, clinicA, "区分A-2")
	makeChiefComplaintType(t, db, clinicB, "区分B-1")

	t.Run("clinicA は自院の2件のみ", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("clinicB は自院の1件のみ（混入なし）", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicB)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "区分B-1", got[0].Name)
	})
}

func TestChiefComplaintTypeRepository_FindAll_ExcludesSoftDeleted(t *testing.T) {
	db := setupChiefComplaintTypeRepositoryTestDB(t)
	repo := NewChiefComplaintTypeRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	active := makeChiefComplaintType(t, db, clinicID, "現役区分")
	deleted := makeChiefComplaintType(t, db, clinicID, "削除済み区分")
	require.NoError(t, repo.Delete(ctx, clinicID, deleted.ID))

	got, err := repo.FindAll(ctx, clinicID)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, active.ID, got[0].ID)

	var rawCount int64
	db.Unscoped().Model(&model.ChiefComplaintType{}).Where("id = ?", deleted.ID).Count(&rawCount)
	assert.Equal(t, int64(1), rawCount, "ソフトデリートされた行はDBにまだ存在する")
}

func TestChiefComplaintTypeRepository_Update(t *testing.T) {
	db := setupChiefComplaintTypeRepositoryTestDB(t)
	repo := NewChiefComplaintTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	c := makeChiefComplaintType(t, db, clinicA, "更新前区分")

	t.Run("同一クリニックでは Update が反映される", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicA, c.ID, map[string]any{"name": "更新後区分"})
		require.NoError(t, err)
		assert.Equal(t, "更新後区分", got.Name)
	})

	t.Run("別クリニックからの Update は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicB, c.ID, map[string]any{"name": "改ざん試行"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, c.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後区分", got.Name, "別クリニックからの Update で名称が変わってはならない")
	})

	t.Run("存在しない ID の Update は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicA, 999999, map[string]any{"name": "x"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestChiefComplaintTypeRepository_Delete(t *testing.T) {
	db := setupChiefComplaintTypeRepositoryTestDB(t)
	repo := NewChiefComplaintTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	c := makeChiefComplaintType(t, db, clinicA, "削除対象区分")

	t.Run("別クリニックからの Delete は NotFound で行が残る", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, c.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, c.ID)
		require.NoError(t, err)
		assert.Equal(t, c.ID, got.ID, "別クリニックからの Delete で行が消えてはならない")
	})

	t.Run("同一クリニックでは Delete が成功し以後 FindByID は NotFound", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, c.ID))
		_, err := repo.FindByID(ctx, clinicA, c.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の Delete は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestChiefComplaintTypeRepository_Reorder(t *testing.T) {
	db := setupChiefComplaintTypeRepositoryTestDB(t)
	repo := NewChiefComplaintTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	c1 := makeChiefComplaintType(t, db, clinicA, "区分X")
	c2 := makeChiefComplaintType(t, db, clinicA, "区分Y")
	c3 := makeChiefComplaintType(t, db, clinicA, "区分Z")

	t.Run("並び順が指定順に更新される", func(t *testing.T) {
		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{c3.ID, c1.ID, c2.ID}))

		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, c3.ID, got[0].ID)
		assert.Equal(t, c1.ID, got[1].ID)
		assert.Equal(t, c2.ID, got[2].ID)
	})

	t.Run("別クリニックの ID を含む Reorder はエラーで中断する", func(t *testing.T) {
		other := makeChiefComplaintType(t, db, clinicB, "他院区分")
		err := repo.Reorder(ctx, clinicA, []uint64{c1.ID, other.ID})
		require.Error(t, err, "clinicA スコープに存在しない ID を含む Reorder は失敗すべき")
	})
}

// TestChiefComplaintTypeRepository_CountUsageByChiefComplaintTypeID_KnownInquiriesColumnBug は
// CountUsageByChiefComplaintTypeID の既知の不具合（本テストで顕在化）を検証する。
//
// 既知の不具合: chief_complaint_repository.go のクエリが `inquiries.deleted_at IS NULL` を
// 参照するが、inquiries テーブルには deleted_at 列が存在しない（migrations/001_init.sql の
// CREATE TABLE inquiries 定義・model.Inquiry 構造体のいずれにも soft-delete 列がない）。
// medicine_repository.go / procedure_repository.go の care_plan_items.deleted_at と同型の不具合
// （medicine_repository_test.go の TestMedicineRepository_CountUsageByMedicineID_TreatmentUsage
// を参照）。そのため使用件数の有無に関わらず、本メソッドは呼び出す度にエラーを返す。この既知の
// 不具合は move-not-modify の対象外（挙動保存のため意図的に未修正のまま移動）。
func TestChiefComplaintTypeRepository_CountUsageByChiefComplaintTypeID_KnownInquiriesColumnBug(t *testing.T) {
	db := setupChiefComplaintTypeRepositoryTestDB(t)
	repo := NewChiefComplaintTypeRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	cc := makeChiefComplaintType(t, db, clinicA, "使用中区分")

	mr := &model.MedicalRecord{ClinicID: clinicA, RecordNo: "CC-0001", Date: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)}
	require.NoError(t, db.WithContext(ctx).Create(mr).Error)

	inquiry := &model.Inquiry{MedicalRecordID: mr.ID, ChiefComplaintTypeID: &cc.ID}
	require.NoError(t, db.WithContext(ctx).Create(inquiry).Error)

	_, err := repo.CountUsageByChiefComplaintTypeID(ctx, clinicA, cc.ID)
	require.Error(t, err, "inquiries.deleted_at 列が存在しないため既知の不具合でエラーになる")
}
