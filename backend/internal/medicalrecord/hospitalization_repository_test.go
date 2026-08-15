package medicalrecord

// hospitalization_repository_test.go — HospitalizationRepository の統合テスト。
// 実 Postgres テスト DB (setupTestDB) に対して実行する。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupHospitalizationRepoTestDB は hospitalizations とその子テーブル群を整備する。
// Cage/Staff/Medicine/Procedure は Hospitalization / CarePlanItem の belongsTo 関連先として
// AutoMigrate 対象に含める（master_preload_clinic_isolation_test.go と同じ組み合わせ）。
// makeCageMaster は internal/repository/master_preload_clinic_isolation_test.go の同名ヘルパーの
// 最小複製（BE9-2D ⑤ Batch A: 原本は旧 package の cage/master_preload テストが引き続き使うため移動不可）。
func makeCageMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Cage {
	t.Helper()
	c := &model.Cage{ClinicID: clinicID, Name: name, CageType: model.CageTypeGeneral, CageSize: model.CageSizeMedium}
	require.NoError(t, db.WithContext(context.Background()).Create(c).Error)
	return c
}

func setupHospitalizationRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.AnimalSpecies{}, &model.Pet{}, &model.Cage{}, &model.Staff{},
		&model.Medicine{}, &model.Procedure{}, &model.HospitalizationPlan{},
		&model.Hospitalization{}, &model.CarePlanItem{}, &model.DailyRecord{}, &model.TreatmentPlan{},
	))
	db.Exec("TRUNCATE TABLE treatment_plans CASCADE")
	db.Exec("TRUNCATE TABLE daily_records CASCADE")
	db.Exec("TRUNCATE TABLE care_plan_items CASCADE")
	db.Exec("TRUNCATE TABLE hospitalizations CASCADE")
	db.Exec("TRUNCATE TABLE hospitalization_plans CASCADE")
	db.Exec("TRUNCATE TABLE medicines CASCADE")
	db.Exec("TRUNCATE TABLE procedures CASCADE")
	db.Exec("TRUNCATE TABLE cages CASCADE")
	db.Exec("TRUNCATE TABLE staffs CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

// makeHospitalizationFixture は status/日付/cage/doctor を柔軟に設定できる Hospitalization を作成する。
func makeHospitalizationFixture(t *testing.T, db *gorm.DB, clinicID, ownerID, petID uint64, opts func(*model.Hospitalization)) *model.Hospitalization {
	t.Helper()
	now := time.Now().UTC().Truncate(24 * time.Hour)
	h := &model.Hospitalization{
		ClinicID:            clinicID,
		OwnerID:             ownerID,
		PetID:               petID,
		HospitalizationType: model.HospitalizationTypeInpatient,
		StartDate:           now,
		EndDate:             now.AddDate(0, 0, 3),
		Status:              model.HospitalizationStatusAdmitted,
	}
	if opts != nil {
		opts(h)
	}
	require.NoError(t, db.WithContext(context.Background()).Create(h).Error)
	return h
}

func TestHospitalizationRepository_CRUDParticipatesInAmbientTransaction(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	repo := NewHospitalizationRepository(db)
	const clinicID = uint64(90301)

	owner := makeTestOwner(t, db, clinicID, "ambient tx 飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "ambient tx ペット")
	errRollback := errors.New("force hospitalization rollback")
	var hospitalizationID uint64

	err := withTx(context.Background(), db, func(txCtx context.Context) error {
		now := time.Now().UTC().Truncate(24 * time.Hour)
		hospitalization := &model.Hospitalization{
			ClinicID:            clinicID,
			OwnerID:             owner.ID,
			PetID:               pet.ID,
			HospitalizationType: model.HospitalizationTypeInpatient,
			StartDate:           now,
			EndDate:             now.AddDate(0, 0, 1),
			Status:              model.HospitalizationStatusAdmitted,
		}
		if err := repo.Create(txCtx, hospitalization); err != nil {
			return err
		}
		hospitalizationID = hospitalization.ID

		created, err := repo.FindByID(txCtx, clinicID, hospitalizationID)
		if err != nil {
			return err
		}
		if created.ID != hospitalizationID {
			t.Errorf("FindByID() ID = %d, want %d", created.ID, hospitalizationID)
		}

		updated, err := repo.Update(txCtx, clinicID, hospitalizationID, map[string]any{"memo": "tx update"})
		if err != nil {
			return err
		}
		if updated.Memo != "tx update" {
			t.Errorf("Update() memo = %q, want %q", updated.Memo, "tx update")
		}
		return errRollback
	})
	require.ErrorIs(t, err, errRollback)
	require.NotZero(t, hospitalizationID)

	_, err = repo.FindByID(context.Background(), clinicID, hospitalizationID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "ambient transaction rollback must remove Create/Update: %v", err)
}

func TestHospitalizationRepository_FindAll(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	repo := NewHospitalizationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "入院飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "入院ポチA")
	cageA := makeCageMaster(t, db, clinicA, "医院Aケージ1")
	doctorA := makeDoctor(t, db, clinicA, "医院A担当医")

	hospA := makeHospitalizationFixture(t, db, clinicA, ownerA.ID, petA.ID, func(h *model.Hospitalization) {
		h.CageID = &cageA.ID
		h.DoctorID = &doctorA.ID
	})

	ownerB := makeTestOwner(t, db, clinicB, "入院飼主B")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "入院ポチB")
	_ = makeHospitalizationFixture(t, db, clinicB, ownerB.ID, petB.ID, nil)

	t.Run("同一クリニックの入院のみ取得しリレーションがPreloadされる", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, nil, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		h := got[0]
		assert.Equal(t, hospA.ID, h.ID)
		require.NotNil(t, h.Pet)
		assert.Equal(t, petA.ID, h.Pet.ID)
		require.NotNil(t, h.Pet.AnimalSpecies)
		require.NotNil(t, h.Owner)
		assert.Equal(t, ownerA.ID, h.Owner.ID)
		require.NotNil(t, h.Cage)
		assert.Equal(t, cageA.ID, h.Cage.ID)
		require.NotNil(t, h.Doctor)
		assert.Equal(t, doctorA.ID, h.Doctor.ID)
	})

	t.Run("pet_idで絞り込む", func(t *testing.T) {
		petA2 := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "入院ポチA2")
		hospA2 := makeHospitalizationFixture(t, db, clinicA, ownerA.ID, petA2.ID, nil)

		got, total, err := repo.FindAll(ctx, clinicA, &petA2.ID, nil, nil, nil, nil, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, hospA2.ID, got[0].ID)
	})

	t.Run("owner_idで絞り込む", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicA, nil, &ownerA.ID, nil, nil, nil, 1, 100)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(1))
		for _, h := range got {
			assert.Equal(t, ownerA.ID, h.OwnerID)
		}
	})

	t.Run("statusで絞り込む", func(t *testing.T) {
		petC := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "退院済みポチ")
		discharged := makeHospitalizationFixture(t, db, clinicA, ownerA.ID, petC.ID, func(h *model.Hospitalization) {
			h.Status = model.HospitalizationStatusDischarged
		})

		status := string(model.HospitalizationStatusDischarged)
		got, total, err := repo.FindAll(ctx, clinicA, nil, nil, &status, nil, nil, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, discharged.ID, got[0].ID)
	})

	t.Run("start_date範囲で絞り込む", func(t *testing.T) {
		petD := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "期間絞り込みポチ")
		early := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
		fixture := makeHospitalizationFixture(t, db, clinicA, ownerA.ID, petD.ID, func(h *model.Hospitalization) {
			h.StartDate = early
			h.EndDate = early.AddDate(0, 0, 2)
		})

		startDate, endDate := "2026-01-01", "2026-01-31"
		got, _, err := repo.FindAll(ctx, clinicA, nil, nil, nil, &startDate, &endDate, 1, 100)
		require.NoError(t, err)
		var found bool
		for _, h := range got {
			if h.ID == fixture.ID {
				found = true
			}
		}
		assert.True(t, found, "範囲内のstart_dateは含まれるべき")

		outOfRangeStart, outOfRangeEnd := "2026-02-01", "2026-02-28"
		got2, _, err := repo.FindAll(ctx, clinicA, nil, nil, nil, &outOfRangeStart, &outOfRangeEnd, 1, 100)
		require.NoError(t, err)
		for _, h := range got2 {
			assert.NotEqual(t, fixture.ID, h.ID, "範囲外のstart_dateは除外されるべき")
		}
	})

	t.Run("別クリニックの入院は含まれない", func(t *testing.T) {
		got, _, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, nil, 1, 100)
		require.NoError(t, err)
		for _, h := range got {
			assert.Equal(t, clinicA, h.ClinicID)
		}
	})
}

func TestHospitalizationRepository_FindAll_CurrentOwnerAfterTransfer(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	repo := NewHospitalizationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(70103)

	fixture := makeCurrentOwnerTransferFixture(
		t,
		db,
		clinicID,
		"MR-HOSPITALIZATION-CURRENT-OWNER",
		time.Now(),
	)
	hospitalization := makeHospitalizationFixture(
		t,
		db,
		clinicID,
		fixture.PreviousOwner.ID,
		fixture.Pet.ID,
		nil,
	)

	got, total, err := repo.FindAll(ctx, clinicID, nil, &fixture.CurrentOwner.ID, nil, nil, nil, 1, 20)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, got, 1)
	assert.Equal(t, hospitalization.ID, got[0].ID)
	assert.Equal(t, fixture.PreviousOwner.ID, got[0].OwnerID, "returned owner_id remains the historical snapshot")

	previous, previousTotal, err := repo.FindAll(ctx, clinicID, nil, &fixture.PreviousOwner.ID, nil, nil, nil, 1, 20)
	require.NoError(t, err)
	assert.Zero(t, previousTotal)
	assert.Empty(t, previous)
}

func TestHospitalizationRepository_FindAll_RejectsCorruptCrossClinicPetRelation(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	repo := NewHospitalizationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "入院取得対象飼主")
	crossClinicPet := makeSpeciesAndPet(t, db, clinicB, ownerA.ID, "破損した別医院ペット")
	makeHospitalizationFixture(t, db, clinicA, ownerA.ID, crossClinicPet.ID, nil)

	got, total, err := repo.FindAll(ctx, clinicA, &crossClinicPet.ID, nil, nil, nil, nil, 1, 100)

	require.NoError(t, err)
	assert.Empty(t, got)
	assert.Zero(t, total)
}

func TestHospitalizationRepository_FindByID_NotFoundAndIsolation(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	repo := NewHospitalizationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "FindByID飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "FindByIDポチA")
	hospA := makeHospitalizationFixture(t, db, clinicA, ownerA.ID, petA.ID, nil)

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, 99999999)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックからは取得できない", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, hospA.ID)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// TestHospitalizationRepository_FindByID_Success は成功パスを検証する。
func TestHospitalizationRepository_FindByID_Success(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	repo := NewHospitalizationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	ownerA := makeTestOwner(t, db, clinicA, "FindByID成功飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "FindByID成功ポチ")
	hospA := makeHospitalizationFixture(t, db, clinicA, ownerA.ID, petA.ID, nil)

	got, err := repo.FindByID(ctx, clinicA, hospA.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, hospA.ID, got.ID)
}

func TestHospitalizationRepository_Create(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	repo := NewHospitalizationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	ownerA := makeTestOwner(t, db, clinicA, "Create飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "Createポチ")

	now := time.Now().UTC().Truncate(24 * time.Hour)
	h := &model.Hospitalization{
		ClinicID:            clinicA,
		OwnerID:             ownerA.ID,
		PetID:               petA.ID,
		HospitalizationType: model.HospitalizationTypeHotel,
		StartDate:           now,
		EndDate:             now.AddDate(0, 0, 1),
	}
	require.NoError(t, repo.Create(ctx, h))
	assert.NotZero(t, h.ID)

	var count int64
	require.NoError(t, db.Model(&model.Hospitalization{}).Where("id = ?", h.ID).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestHospitalizationRepository_Update_NotFoundAndIsolation(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	repo := NewHospitalizationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "Update飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "Updateポチ")
	hospA := makeHospitalizationFixture(t, db, clinicA, ownerA.ID, petA.ID, nil)

	t.Run("別クリニックからの更新はNotFound", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicB, hospA.ID, map[string]any{"memo": "不正更新"})
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDの更新はNotFound", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicA, 99999999, map[string]any{"memo": "x"})
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// TestHospitalizationRepository_Update_Success は成功パスを検証する。
func TestHospitalizationRepository_Update_Success(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	repo := NewHospitalizationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	ownerA := makeTestOwner(t, db, clinicA, "Update成功飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "Update成功ポチ")
	hospA := makeHospitalizationFixture(t, db, clinicA, ownerA.ID, petA.ID, nil)

	got, err := repo.Update(ctx, clinicA, hospA.ID, map[string]any{"memo": "更新後メモ"})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "更新後メモ", got.Memo)
}

func TestHospitalizationRepository_Delete(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	repo := NewHospitalizationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "Delete飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "Deleteポチ")
	hospA := makeHospitalizationFixture(t, db, clinicA, ownerA.ID, petA.ID, nil)

	t.Run("別クリニックからの削除はNotFoundで実際には削除されない", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, hospA.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		var count int64
		require.NoError(t, db.Model(&model.Hospitalization{}).Where("id = ?", hospA.ID).Count(&count).Error)
		assert.Equal(t, int64(1), count)
	})

	t.Run("同一クリニックからの削除は成功しソフトデリートされる", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, hospA.ID))

		var count int64
		require.NoError(t, db.Model(&model.Hospitalization{}).Where("id = ?", hospA.ID).Count(&count).Error)
		assert.Equal(t, int64(0), count, "通常クエリでは見えなくなる")

		var raw model.Hospitalization
		require.NoError(t, db.Unscoped().Where("id = ?", hospA.ID).First(&raw).Error)
		assert.True(t, raw.DeletedAt.Valid)
	})

	t.Run("存在しないIDの削除はNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 99999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// TestHospitalizationRepository_CountCarePlanItemsByHospitalizationID は
// 入院に紐づくケアプラン項目数の集計を検証する。
func TestHospitalizationRepository_CountCarePlanItemsByHospitalizationID(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	repo := NewHospitalizationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	ownerA := makeTestOwner(t, db, clinicA, "ケアプラン集計飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ケアプラン集計ポチ")
	hospA := makeHospitalizationFixture(t, db, clinicA, ownerA.ID, petA.ID, nil)

	item := &model.CarePlanItem{
		HospitalizationID: hospA.ID,
		Type:              model.CarePlanTypeFood,
		Name:              "朝食",
	}
	require.NoError(t, db.WithContext(ctx).Create(item).Error)

	count, err := repo.CountCarePlanItemsByHospitalizationID(ctx, clinicA, hospA.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

// TestHospitalizationRepository_CountDailyRecordsByHospitalizationID は
// 入院に紐づく日次記録数の集計を検証する。
func TestHospitalizationRepository_CountDailyRecordsByHospitalizationID(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	repo := NewHospitalizationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "日次記録集計飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "日次記録集計ポチ")
	hospA := makeHospitalizationFixture(t, db, clinicA, ownerA.ID, petA.ID, nil)

	dr := &model.DailyRecord{ClinicID: clinicA, HospitalizationID: hospA.ID, Date: time.Now()}
	require.NoError(t, db.WithContext(ctx).Create(dr).Error)
	foreign := &model.DailyRecord{
		ClinicID:          clinicB,
		HospitalizationID: hospA.ID,
		Date:              dr.Date.AddDate(0, 0, 1),
	}
	require.NoError(t, db.WithContext(ctx).Create(foreign).Error)

	count, err := repo.CountDailyRecordsByHospitalizationID(ctx, clinicA, hospA.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count, "foreign-clinic child rows must not affect the clinic-scoped count")
}

func TestHospitalizationRepository_CountTreatmentPlansByHospitalizationID(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	repo := NewHospitalizationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "治療計画集計飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "治療計画集計ポチ")
	hospA := makeHospitalizationFixture(t, db, clinicA, ownerA.ID, petA.ID, nil)

	t.Run("0件", func(t *testing.T) {
		count, err := repo.CountTreatmentPlansByHospitalizationID(ctx, clinicA, hospA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	tp1 := &model.TreatmentPlan{ClinicID: clinicA, HospitalizationID: &hospA.ID, TreatmentContent: "内服薬投与"}
	require.NoError(t, db.WithContext(ctx).Create(tp1).Error)
	tp2 := &model.TreatmentPlan{ClinicID: clinicA, HospitalizationID: &hospA.ID, TreatmentContent: "点滴"}
	require.NoError(t, db.WithContext(ctx).Create(tp2).Error)

	t.Run("2件", func(t *testing.T) {
		count, err := repo.CountTreatmentPlansByHospitalizationID(ctx, clinicA, hospA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("ソフトデリートされた治療計画は除外される", func(t *testing.T) {
		require.NoError(t, db.WithContext(ctx).Delete(&model.TreatmentPlan{}, tp1.ID).Error)
		count, err := repo.CountTreatmentPlansByHospitalizationID(ctx, clinicA, hospA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("別クリニックからは0件", func(t *testing.T) {
		count, err := repo.CountTreatmentPlansByHospitalizationID(ctx, clinicB, hospA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestHospitalizationRepository_UpdateIfNotDischarged(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	repo := NewHospitalizationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	ownerA := makeTestOwner(t, db, clinicA, "UpdateIfNotDischarged飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "UpdateIfNotDischargedポチ")

	t.Run("admitted status は退院更新に成功する", func(t *testing.T) {
		hosp := makeHospitalizationFixture(t, db, clinicA, ownerA.ID, petA.ID, func(h *model.Hospitalization) {
			h.Status = model.HospitalizationStatusAdmitted
		})
		dischargeDate := time.Now().UTC().Truncate(time.Second)

		got, err := repo.UpdateIfNotDischarged(ctx, clinicA, hosp.ID, map[string]any{
			"status":   model.HospitalizationStatusDischarged,
			"end_date": dischargeDate,
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, model.HospitalizationStatusDischarged, got.Status)
	})

	t.Run("reserved status は退院更新に成功する", func(t *testing.T) {
		petB := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "UpdateIfNotDischarged予約ポチ")
		hosp := makeHospitalizationFixture(t, db, clinicA, ownerA.ID, petB.ID, func(h *model.Hospitalization) {
			h.Status = model.HospitalizationStatusReserved
		})
		dischargeDate := time.Now().UTC().Truncate(time.Second)

		got, err := repo.UpdateIfNotDischarged(ctx, clinicA, hosp.ID, map[string]any{
			"status":   model.HospitalizationStatusDischarged,
			"end_date": dischargeDate,
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, model.HospitalizationStatusDischarged, got.Status)
	})

	t.Run("already discharged は NotFound", func(t *testing.T) {
		petC := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "UpdateIfNotDischarged退院済みポチ")
		hosp := makeHospitalizationFixture(t, db, clinicA, ownerA.ID, petC.ID, func(h *model.Hospitalization) {
			h.Status = model.HospitalizationStatusDischarged
		})

		got, err := repo.UpdateIfNotDischarged(ctx, clinicA, hosp.ID, map[string]any{
			"status": model.HospitalizationStatusDischarged,
		})
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
