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

	t.Run("同一医院IDでは取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, addendum.ID)
		require.NoError(t, err)
		assert.Equal(t, addendum.ID, got.ID)
	})

	t.Run("別医院IDでは取得できない（clinic_id隔離）", func(t *testing.T) {
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
	// 同じ medical_record_id だが別医院の追記行（isolation 検証用）
	makeAddendum(t, db, clinicB, mr.ID, staff.ID, "別医院の修正", "理由X")

	t.Run("同一医院・同一カルテの追記のみ古い順で返す", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mr.ID)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, first.ID, got[0].ID)
		assert.Equal(t, second.ID, got[1].ID)
	})

	t.Run("別医院IDでは0件になる（親カルテclinic相関）", func(t *testing.T) {
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

// TestMedicalRecordAddendumRepository_FindByMedicalRecordID_AuthorStaffResolution は
// 追記者（AuthorStaff）が「親カルテの clinic_id」スコープかつ deleted_at IS NULL でのみ
// 解決されることを検証する。未解決は nil → 応答側で author_user_id フォールバック。
// is_active=false（退職）でも解決するのは監査帰属の仕様（scopedBillingStaffPreload と同判断）。
func TestMedicalRecordAddendumRepository_FindByMedicalRecordID_AuthorStaffResolution(t *testing.T) {
	db := setupMedicalRecordAddendumTestDB(t)
	repo := NewMedicalRecordAddendumRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	ensureVaccinationTestClinics(t, db, clinicA, clinicB)

	owner := makeTestOwner(t, db, clinicA, "追記飼主4")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "追記犬4")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "ADD-4", time.Now())

	findAuthorStaff := func(t *testing.T, addendumID uint64) *model.Staff {
		t.Helper()
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mr.ID)
		require.NoError(t, err)
		for _, a := range got {
			if a.ID == addendumID {
				return a.AuthorStaff
			}
		}
		t.Fatalf("addendum %d not found in list response", addendumID)
		return nil
	}

	t.Run("同医院の追記者は氏名が解決される", func(t *testing.T) {
		staff := makeDoctor(t, db, clinicA, "追記医師A")
		a := makeAddendum(t, db, clinicA, mr.ID, staff.ID, "同医院追記", "理由")
		got := findAuthorStaff(t, a.ID)
		require.NotNil(t, got)
		assert.Equal(t, "追記医師A", got.Name)
	})

	t.Run("改名後は新しい氏名が返る", func(t *testing.T) {
		staff := makeDoctor(t, db, clinicA, "改名前医師")
		a := makeAddendum(t, db, clinicA, mr.ID, staff.ID, "改名追記", "理由")
		require.NoError(t, db.Model(staff).Update("name", "改名後医師").Error)
		got := findAuthorStaff(t, a.ID)
		require.NotNil(t, got)
		assert.Equal(t, "改名後医師", got.Name)
	})

	t.Run("削除済みスタッフは解決されずnilになる（IDフォールバック）", func(t *testing.T) {
		staff := makeDoctor(t, db, clinicA, "削除予定医師")
		a := makeAddendum(t, db, clinicA, mr.ID, staff.ID, "削除追記", "理由")
		require.NoError(t, db.Delete(staff).Error)
		assert.Nil(t, findAuthorStaff(t, a.ID))
	})

	t.Run("他医院スタッフは解決されない（氏名漏洩なし）", func(t *testing.T) {
		other := makeDoctor(t, db, clinicB, "他院医師")
		a := makeAddendum(t, db, clinicA, mr.ID, other.ID, "他院追記", "理由")
		assert.Nil(t, findAuthorStaff(t, a.ID))
	})

	t.Run("is_active=falseでも監査帰属のため氏名は解決される", func(t *testing.T) {
		staff := makeDoctor(t, db, clinicA, "退職医師")
		require.NoError(t, db.Model(staff).Update("is_active", false).Error)
		a := makeAddendum(t, db, clinicA, mr.ID, staff.ID, "退職追記", "理由")
		got := findAuthorStaff(t, a.ID)
		require.NotNil(t, got)
		assert.Equal(t, "退職医師", got.Name)
	})
}
