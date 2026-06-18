package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// makeMedicalRecordForPet はテスト用カルテ 1 件を作成する。deleted=true なら論理削除する。
// pet_id は medical_records → pets の FK 制約があるため、呼び出し側で実在ペットの ID を渡すこと。
// 既存ヘルパー（makeOwner / makeSpeciesAndPet）に合わせ context は内部生成する。
func makeMedicalRecordForPet(t *testing.T, db *gorm.DB, clinicID, petID uint64, date time.Time, deleted bool) {
	t.Helper()
	ctx := context.Background()
	pid := petID
	mr := &model.MedicalRecord{ClinicID: clinicID, PetID: &pid, Date: date}
	require.NoError(t, db.WithContext(ctx).Create(mr).Error)
	if deleted {
		// gorm.DeletedAt により deleted_at がセットされる（論理削除）。
		require.NoError(t, db.WithContext(ctx).Delete(mr).Error)
	}
}

// TestMedicalRecordRepository_FindFirstVisitDateByPetID は #158 飼主レポートの初診日導出を検証する。
// 期待: 同一ペット・同一医院の有効カルテのうち最古の date を返し、
// 論理削除カルテ・別医院カルテ・別ペットのカルテは集計から除外する。捏造はしない（無ければ nil）。
func TestMedicalRecordRepository_FindFirstVisitDateByPetID(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.AnimalSpecies{}, &model.Pet{}))
	// 共有テスト DB の他テスト残骸と混ざらないよう、FK 連鎖ごと初期化する。
	db.Exec("TRUNCATE TABLE medical_records, pets, animal_species CASCADE")

	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	d := func(s string) time.Time {
		tm, err := time.Parse("2006-01-02", s)
		require.NoError(t, err)
		return tm
	}

	owner := makeOwner(t, db, clinicA, "山田太郎")
	petA := makeSpeciesAndPet(t, db, clinicA, owner.ID, "ポチ")
	// otherPet の所属医院（clinicB）はテスト対象外。pet スコープ除外の検証では
	// otherPet.ID を使った clinicA のカルテを作り、petA への問い合わせから除外されることを見る。
	otherPet := makeSpeciesAndPet(t, db, clinicB, owner.ID, "タマ")
	deletedOnlyPet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "ミケ")

	// clinicA / petA: 有効カルテの最古は 2023-03-01。
	makeMedicalRecordForPet(t, db, clinicA, petA.ID, d("2024-01-15"), false)
	makeMedicalRecordForPet(t, db, clinicA, petA.ID, d("2023-03-01"), false) // 最古の有効カルテ
	// より古いが論理削除済み → 除外される。
	makeMedicalRecordForPet(t, db, clinicA, petA.ID, d("2022-01-01"), true)
	// 同じ pet だが別医院 (clinicB) のより古いカルテ → clinic スコープで除外される。
	makeMedicalRecordForPet(t, db, clinicB, petA.ID, d("2020-01-01"), false)
	// clinicA の別ペットのより古いカルテ → pet スコープで除外される。
	makeMedicalRecordForPet(t, db, clinicA, otherPet.ID, d("2019-01-01"), false)
	// 全カルテ論理削除のペット。
	makeMedicalRecordForPet(t, db, clinicA, deletedOnlyPet.ID, d("2021-05-05"), true)

	t.Run("最古の有効カルテ date を返し、論理削除・別医院・別ペットを除外する", func(t *testing.T) {
		got, err := repo.FindFirstVisitDateByPetID(ctx, clinicA, petA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "2023-03-01", got.Format("2006-01-02"))
	})

	t.Run("カルテが無いペットは nil を返す（捏造しない）", func(t *testing.T) {
		got, err := repo.FindFirstVisitDateByPetID(ctx, clinicA, uint64(99999))
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("全カルテが論理削除されたペットは nil を返す", func(t *testing.T) {
		got, err := repo.FindFirstVisitDateByPetID(ctx, clinicA, deletedOnlyPet.ID)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
