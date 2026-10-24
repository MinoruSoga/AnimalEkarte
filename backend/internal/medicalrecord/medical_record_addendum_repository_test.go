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
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}, &model.Staff{}, &model.StaffClinicAssignment{}, &model.MedicalRecordAddendum{}))
	db.Exec("TRUNCATE TABLE medical_record_addenda CASCADE")
	db.Exec("TRUNCATE TABLE staff_clinic_assignments CASCADE")
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

// assignStaffToClinic は staff_clinic_assignments 行を作成する。実運用ではスタッフ
// 作成・所属更新で必ず主医院の is_main 行（＋兼務先の行）が書き込まれるため、
// 著者解決テストでも同じ正本データ形を再現する。
func assignStaffToClinic(t *testing.T, db *gorm.DB, staffID, clinicID uint64, isMain bool) *model.StaffClinicAssignment {
	t.Helper()
	a := &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID, IsMain: isMain}
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
// 追記者（AuthorStaff）が「親カルテの clinic_id に有効な staff_clinic_assignments 行を
// 持つ」スコープかつ deleted_at IS NULL でのみ解決されることを検証する。医院所属の正本は
// staff_clinic_assignments（認証側と同じ所属相関 predicate）。未解決は nil → 応答側で
// author_user_id フォールバック。
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
		assignStaffToClinic(t, db, staff.ID, clinicA, true)
		a := makeAddendum(t, db, clinicA, mr.ID, staff.ID, "同医院追記", "理由")
		got := findAuthorStaff(t, a.ID)
		require.NotNil(t, got)
		assert.Equal(t, "追記医師A", got.Name)
	})

	t.Run("改名後は現行名が返る（契約確認済み）", func(t *testing.T) {
		// 契約: 解決可能な追記者は現行名を表示する。旧名スナップショット列は
		// スキーマに存在せず、監査上も現在名で帰属を把握する方が有用と判断済み。
		// （削除・異動＝所属解除・他医院・削除済み所属のみ ID フォールバック）
		staff := makeDoctor(t, db, clinicA, "改名前医師")
		assignStaffToClinic(t, db, staff.ID, clinicA, true)
		a := makeAddendum(t, db, clinicA, mr.ID, staff.ID, "改名追記", "理由")
		require.NoError(t, db.Model(staff).Update("name", "改名後医師").Error)
		got := findAuthorStaff(t, a.ID)
		require.NotNil(t, got)
		assert.Equal(t, "改名後医師", got.Name)
	})

	t.Run("削除済みスタッフは解決されずnilになる（IDフォールバック）", func(t *testing.T) {
		staff := makeDoctor(t, db, clinicA, "削除予定医師")
		assignStaffToClinic(t, db, staff.ID, clinicA, true)
		a := makeAddendum(t, db, clinicA, mr.ID, staff.ID, "削除追記", "理由")
		require.NoError(t, db.Delete(staff).Error)
		assert.Nil(t, findAuthorStaff(t, a.ID))
	})

	t.Run("他医院スタッフは解決されない（氏名漏洩なし）", func(t *testing.T) {
		other := makeDoctor(t, db, clinicB, "他院医師")
		assignStaffToClinic(t, db, other.ID, clinicB, true)
		a := makeAddendum(t, db, clinicA, mr.ID, other.ID, "他院追記", "理由")
		assert.Nil(t, findAuthorStaff(t, a.ID))
	})

	t.Run("兼務スタッフは主医院が他院でも所属医院の追記で氏名が解決される", func(t *testing.T) {
		// 主医院 B ・兼務所属 A のスタッフが A で追記 → 認証側が所属で許可する
		// 正規の著者なので氏名が解決される（staffs.clinic_id 比較では漏れる）。
		staff := makeDoctor(t, db, clinicB, "兼務医師")
		assignStaffToClinic(t, db, staff.ID, clinicB, true)
		assignStaffToClinic(t, db, staff.ID, clinicA, false)
		a := makeAddendum(t, db, clinicA, mr.ID, staff.ID, "兼務追記", "理由")
		got := findAuthorStaff(t, a.ID)
		require.NotNil(t, got)
		assert.Equal(t, "兼務医師", got.Name)
	})

	t.Run("所属なしスタッフは主医院が別院なら解決されない", func(t *testing.T) {
		// 主医院 B で所属行が無いスタッフ → A の追記では解決されない。
		staff := makeDoctor(t, db, clinicB, "所属なし医師")
		a := makeAddendum(t, db, clinicA, mr.ID, staff.ID, "無所属追記", "理由")
		assert.Nil(t, findAuthorStaff(t, a.ID))
	})

	t.Run("削除済みの兼務所属では解決されない", func(t *testing.T) {
		// 主医院 B ・兼務所属 A が論理削除済み → A の追記では解決されない。
		staff := makeDoctor(t, db, clinicB, "元兼務医師")
		assignStaffToClinic(t, db, staff.ID, clinicB, true)
		assignment := assignStaffToClinic(t, db, staff.ID, clinicA, false)
		require.NoError(t, db.Delete(assignment).Error)
		a := makeAddendum(t, db, clinicA, mr.ID, staff.ID, "元兼務追記", "理由")
		assert.Nil(t, findAuthorStaff(t, a.ID))
	})

	t.Run("is_active=falseでも監査帰属のため氏名は解決される", func(t *testing.T) {
		staff := makeDoctor(t, db, clinicA, "退職医師")
		assignStaffToClinic(t, db, staff.ID, clinicA, true)
		require.NoError(t, db.Model(staff).Update("is_active", false).Error)
		a := makeAddendum(t, db, clinicA, mr.ID, staff.ID, "退職追記", "理由")
		got := findAuthorStaff(t, a.ID)
		require.NotNil(t, got)
		assert.Equal(t, "退職医師", got.Name)
	})
}

// EMR-495 Q2: 最新の診療終了イベントの種別を、同一医院・同一カルテの最大 revision で返す。
// AutoMigrate の test double には 001 の trigger / operation receipt 制約がないため、行を直接挿入する。
func TestMedicalRecordAddendumRepository_LatestConsultationEventKind(t *testing.T) {
	db := setupMedicalRecordAddendumTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.MedicalRecordConsultationEvent{}))
	db.Exec("TRUNCATE TABLE medical_record_consultation_events CASCADE")
	repo := NewMedicalRecordAddendumRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "終了履歴飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "終了履歴犬")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "ADD-EVT-1", time.Now())
	insertEvent := func(revision int64, kind string) {
		require.NoError(t, db.Create(&model.MedicalRecordConsultationEvent{
			ClinicID: clinicA, MedicalRecordID: mr.ID, Revision: revision, EventKind: kind,
			SourceKind: "consultation", OccurredAt: time.Now(), OperationID: 1, RecordedBy: 1,
		}).Error)
	}

	t.Run("no event: found=false", func(t *testing.T) {
		kind, found, err := repo.LatestConsultationEventKind(ctx, clinicA, mr.ID)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Empty(t, kind)
	})

	insertEvent(1, "ended")
	t.Run("ended", func(t *testing.T) {
		kind, found, err := repo.LatestConsultationEventKind(ctx, clinicA, mr.ID)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "ended", kind)
	})

	t.Run("other clinic sees no event", func(t *testing.T) {
		_, found, err := repo.LatestConsultationEventKind(ctx, clinicB, mr.ID)
		require.NoError(t, err)
		assert.False(t, found)
	})

	insertEvent(2, "end_undone")
	t.Run("latest revision wins (end_undone)", func(t *testing.T) {
		kind, found, err := repo.LatestConsultationEventKind(ctx, clinicA, mr.ID)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "end_undone", kind)
	})
}
