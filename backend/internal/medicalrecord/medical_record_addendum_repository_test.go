package medicalrecord

// medical_record_addendum_repository_test.go — MedicalRecordAddendumRepository の統合テスト。
//
// 保護する不変条件:
//   - FindByID / FindByMedicalRecordID は clinic_id で正しく分離される。
//   - FindByMedicalRecordID は created_at 昇順で返す。
//   - FindByID は存在しないIDに対して NotFound ラップされたエラーを返す。

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

// setupMedicalRecordAddendumTestDB は medical_record_addenda と、その FK 先である
// pets/animal_species/staffs/medical_records を整備する。
func setupMedicalRecordAddendumTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}, &model.Staff{}, &model.MedicalRecordAddendum{}))
	db.Exec("TRUNCATE TABLE medical_record_addenda CASCADE")
	db.Exec("TRUNCATE TABLE staffs CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func makeAddendum(t *testing.T, db *gorm.DB, clinicID, medicalRecordID, authorUserID uint64, afterText, reason string) *model.MedicalRecordAddendum {
	t.Helper()
	a := &model.MedicalRecordAddendum{
		MedicalRecordID: medicalRecordID,
		ClinicID:        clinicID,
		AuthorUserID:    authorUserID,
		AfterText:       afterText,
		Reason:          reason,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(a).Error)
	return a
}

func TestMedicalRecordAddendumRepository_Create(t *testing.T) {
	db := setupMedicalRecordAddendumTestDB(t)
	repo := NewMedicalRecordAddendumRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "追記飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "追記犬")
	staff := makeDoctor(t, db, clinicA, "追記医師")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "ADD-1", time.Now())

	addendum := &model.MedicalRecordAddendum{
		MedicalRecordID: mr.ID,
		ClinicID:        clinicA,
		AuthorUserID:    staff.ID,
		BeforeText:      "旧テキスト",
		AfterText:       "新テキスト",
		Reason:          "誤記訂正",
	}
	require.NoError(t, repo.Create(ctx, addendum))
	assert.NotZero(t, addendum.ID)

	var stored model.MedicalRecordAddendum
	require.NoError(t, db.First(&stored, addendum.ID).Error)
	assert.Equal(t, "新テキスト", stored.AfterText)
	assert.Equal(t, "誤記訂正", stored.Reason)
}

func TestMedicalRecordAddendumRepository_FindByID(t *testing.T) {
	db := setupMedicalRecordAddendumTestDB(t)
	repo := NewMedicalRecordAddendumRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "追記飼主2")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "追記犬2")
	staff := makeDoctor(t, db, clinicA, "追記医師2")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "ADD-2", time.Now())
	addendum := makeAddendum(t, db, clinicA, mr.ID, staff.ID, "新テキスト2", "理由2")

	t.Run("同一クリニックIDでは取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, addendum.ID)
		require.NoError(t, err)
		assert.Equal(t, addendum.ID, got.ID)
	})

	t.Run("別クリニックIDでは取得できない（clinic_id隔離）", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, addendum.ID)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDはNotFoundを返す", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, uint64(9999999))
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestMedicalRecordAddendumRepository_FindByMedicalRecordID(t *testing.T) {
	db := setupMedicalRecordAddendumTestDB(t)
	repo := NewMedicalRecordAddendumRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "追記飼主3")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "追記犬3")
	staff := makeDoctor(t, db, clinicA, "追記医師3")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "ADD-3", time.Now())

	first := makeAddendum(t, db, clinicA, mr.ID, staff.ID, "1回目修正", "理由1")
	time.Sleep(2 * time.Millisecond)
	second := makeAddendum(t, db, clinicA, mr.ID, staff.ID, "2回目修正", "理由2")
	// 同じ medical_record_id だが別クリニックの追記行（isolation 検証用）
	makeAddendum(t, db, clinicB, mr.ID, staff.ID, "別クリニックの修正", "理由X")

	t.Run("同一クリニック・同一カルテの追記のみ古い順で返す", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mr.ID)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, first.ID, got[0].ID)
		assert.Equal(t, second.ID, got[1].ID)
	})

	t.Run("別クリニックIDでは0件になる（親カルテclinic相関）", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicB, mr.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("追記が無いカルテIDは空スライスを返す", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, uint64(9999999))
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}
