package medicalrecord

// 移動元 internal/repository（BE9-2D sub-batch④b）。setupTestDB/ensureAutoMigrated は
// repotest 直呼びへ（vital_repository_test.go 先例）。makeHistoryMedicalRecord は本 package の
// diagnosis_name_repository_test.go に byte-identical 定義が既存のため移動時に統合（複製削除）。

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

// setupTreatmentHistoryTestDB は setupTestDB を拡張し pets / treatments / procedures 周りを整備する。
// FindHistoryByPetID は medical_records JOIN で clinic 隔離するため medical_records も使う。
// medical_records.pet_id は pets への FK があるため pets/animal_species/owners も用意する。
func setupTreatmentHistoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	if err := testdb.EnsureAutoMigrated(
		db,
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.Consultation{},
		&model.Procedure{},
		&model.Medicine{},
		&model.InventoryItem{},
		&model.Treatment{},
	); err != nil {
		t.Fatalf("failed to migrate treatment read graph: %v", err)
	}
	db.Exec("TRUNCATE TABLE treatments CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE consultations CASCADE")
	db.Exec("TRUNCATE TABLE procedures CASCADE")
	db.Exec("TRUNCATE TABLE medicines CASCADE")
	db.Exec("TRUNCATE TABLE inventory_items CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func makeHistoryTreatment(t *testing.T, db *gorm.DB, medicalRecordID uint64, itemType model.TreatmentItemType, content string, sortOrder int) {
	t.Helper()
	tr := &model.Treatment{
		MedicalRecordID: medicalRecordID,
		ItemType:        itemType,
		Content:         content,
		SortOrder:       sortOrder,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(tr).Error)
}

func makeHistoryTreatmentWithProcedure(t *testing.T, db *gorm.DB, medicalRecordID, procedureID uint64, content string) {
	t.Helper()
	tr := &model.Treatment{
		MedicalRecordID: medicalRecordID,
		ProcedureID:     &procedureID,
		ItemType:        model.TreatmentItemTypeProcedure,
		Content:         content,
		SortOrder:       0,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(tr).Error)
}

func makeProcedure(t *testing.T, db *gorm.DB, clinicID uint64, name string, anesthesia model.AnesthesiaType, isSurgery bool) *model.Procedure {
	t.Helper()
	p := &model.Procedure{
		ClinicID:   clinicID,
		Name:       name,
		Anesthesia: anesthesia,
		IsSurgery:  isSurgery,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(p).Error)
	return p
}

func TestTreatmentRepository_FindHistoryByPetID(t *testing.T) {
	db := setupTreatmentHistoryTestDB(t)
	repo := NewTreatmentRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	owner := makeTestOwner(t, db, clinicA, "飼主A")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "ポチ")

	older := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC)

	// clinic A, pet: 2 カルテ・3 treatment（投薬2 + 処置1）
	mrOld := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "A-OLD", older)
	mrNew := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "A-NEW", newer)
	makeHistoryTreatment(t, db, mrOld.ID, model.TreatmentItemTypeMedicine, "旧・投薬", 0)
	makeHistoryTreatment(t, db, mrNew.ID, model.TreatmentItemTypeMedicine, "新・投薬", 0)
	makeHistoryTreatment(t, db, mrNew.ID, model.TreatmentItemTypeProcedure, "新・処置", 1)

	// 同一 pet だが clinic B のカルテ（clinic_id フィルタが効くかを検証）→ 混入してはならない
	mrOther := makeHistoryMedicalRecord(t, db, clinicB, pet.ID, "B-1", newer)
	makeHistoryTreatment(t, db, mrOther.ID, model.TreatmentItemTypeMedicine, "別clinic・投薬", 0)

	t.Run("returns only same-clinic treatments (clinic isolation)", func(t *testing.T) {
		got, total, err := repo.FindHistoryByPetID(ctx, clinicA, pet.ID, model.PetTreatmentHistoryFilter{}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total, "別 clinic の treatment は混入しない")
		require.Len(t, got, 3)
		for _, tr := range got {
			assert.NotEqual(t, "別clinic・投薬", tr.Content)
		}
	})

	t.Run("orders by medical_records.date DESC", func(t *testing.T) {
		got, _, err := repo.FindHistoryByPetID(ctx, clinicA, pet.ID, model.PetTreatmentHistoryFilter{}, 1, 100)
		require.NoError(t, err)
		require.Len(t, got, 3)
		// 先頭 2 件は新しいカルテ(6/3)、最後が古いカルテ(6/1)
		assert.Equal(t, mrNew.ID, got[0].MedicalRecordID)
		assert.Equal(t, mrNew.ID, got[1].MedicalRecordID)
		assert.Equal(t, mrOld.ID, got[2].MedicalRecordID)
		// preload された MedicalRecord.Date が降順
		require.NotNil(t, got[0].MedicalRecord)
		require.NotNil(t, got[2].MedicalRecord)
		assert.True(t, got[0].MedicalRecord.Date.After(got[2].MedicalRecord.Date))
	})

	t.Run("filters by item_type", func(t *testing.T) {
		medicine := model.TreatmentItemTypeMedicine
		got, total, err := repo.FindHistoryByPetID(ctx, clinicA, pet.ID, model.PetTreatmentHistoryFilter{ItemType: &medicine}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		require.Len(t, got, 2)
		for _, tr := range got {
			assert.Equal(t, model.TreatmentItemTypeMedicine, tr.ItemType)
		}
	})

	t.Run("returns empty for pet with no history", func(t *testing.T) {
		got, total, err := repo.FindHistoryByPetID(ctx, clinicA, uint64(999999), model.PetTreatmentHistoryFilter{}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, got)
	})
}

func TestTreatmentRepository_FindHistoryByPetID_RejectsCorruptCrossClinicPetRelation(t *testing.T) {
	db := setupTreatmentHistoryTestDB(t)
	repo := NewTreatmentRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "治療履歴取得対象飼主")
	crossClinicPet := makeSpeciesAndPet(t, db, clinicB, ownerA.ID, "破損した別医院ペット")
	mr := makeHistoryMedicalRecord(t, db, clinicA, crossClinicPet.ID, "MR-CORRUPT-HISTORY", time.Now())
	makeHistoryTreatment(t, db, mr.ID, model.TreatmentItemTypeMedicine, "別医院ペット由来治療", 0)

	got, total, err := repo.FindHistoryByPetID(
		ctx,
		clinicA,
		crossClinicPet.ID,
		model.PetTreatmentHistoryFilter{},
		1,
		100,
	)

	require.NoError(t, err)
	assert.Empty(t, got)
	assert.Zero(t, total)
}

func TestTreatmentRepository_FindHistoryByPetID_ProcedureFilters(t *testing.T) {
	db := setupTreatmentHistoryTestDB(t)
	repo := NewTreatmentRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "飼主B")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "タマ")

	date := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "PROC-1", date)

	// 麻酔あり処置（anesthesia != none）
	procAnesthesia := makeProcedure(t, db, clinicA, "麻酔処置", model.AnesthesiaTypeGeneral, false)
	// 手術処置（is_surgery = true）
	procSurgery := makeProcedure(t, db, clinicA, "手術", model.AnesthesiaTypeNone, true)
	// 麻酔なし・手術なし（通常処置）
	procNormal := makeProcedure(t, db, clinicA, "通常処置", model.AnesthesiaTypeNone, false)
	// 汚染 fixture: clinic A の treatment が clinic B の procedure を参照しても、
	// procedure filter の list/count には混入させない。
	procForeign := makeProcedure(t, db, clinicA+1, "別院の麻酔手術", model.AnesthesiaTypeGeneral, true)
	// 投薬（procedure_id なし）
	makeHistoryTreatment(t, db, mr.ID, model.TreatmentItemTypeMedicine, "投薬", 3)

	makeHistoryTreatmentWithProcedure(t, db, mr.ID, procAnesthesia.ID, "麻酔処置実施")
	makeHistoryTreatmentWithProcedure(t, db, mr.ID, procSurgery.ID, "手術実施")
	makeHistoryTreatmentWithProcedure(t, db, mr.ID, procNormal.ID, "通常処置実施")
	makeHistoryTreatmentWithProcedure(t, db, mr.ID, procForeign.ID, "別院処置の汚染")

	t.Run("AnesthesiaOnly returns only treatments linked to anesthetic procedures", func(t *testing.T) {
		got, total, err := repo.FindHistoryByPetID(ctx, clinicA, pet.ID, model.PetTreatmentHistoryFilter{AnesthesiaOnly: true}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total, "麻酔あり処置のみ")
		require.Len(t, got, 1)
		require.NotNil(t, got[0].Procedure)
		assert.Equal(t, model.AnesthesiaTypeGeneral, got[0].Procedure.Anesthesia)
		assert.Equal(t, "麻酔処置実施", got[0].Content)
	})

	t.Run("IsSurgery returns only treatments linked to surgery procedures", func(t *testing.T) {
		got, total, err := repo.FindHistoryByPetID(ctx, clinicA, pet.ID, model.PetTreatmentHistoryFilter{IsSurgery: true}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total, "手術処置のみ")
		require.Len(t, got, 1)
		require.NotNil(t, got[0].Procedure)
		assert.True(t, got[0].Procedure.IsSurgery)
		assert.Equal(t, "手術実施", got[0].Content)
	})

	t.Run("no filter returns all treatments including medicine", func(t *testing.T) {
		got, total, err := repo.FindHistoryByPetID(ctx, clinicA, pet.ID, model.PetTreatmentHistoryFilter{}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(5), total, "filter 未指定では既存の全 treatment を返す")
		assert.Len(t, got, 5)
	})
}

// setupTreatmentBillingTestDB は setupTreatmentHistoryTestDB を拡張し、FindUnbilledByPetID /
// CountFinalizedUnconfirmedByPetAndDate のテストに必要な billing_confirmations / billings / billing_items
// も整備する（billings は setupTestDB で AutoMigrate 済み・medical_records/pets 等の重複 TRUNCATE は無害）。
func setupTreatmentBillingTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTreatmentHistoryTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.BillingConfirmation{}, &model.BillingItem{}))
	db.Exec("TRUNCATE TABLE billing_items CASCADE")
	db.Exec("TRUNCATE TABLE billing_confirmations CASCADE")
	return db
}

func TestTreatmentRepository_FindByID(t *testing.T) {
	db := setupTreatmentHistoryTestDB(t)
	repo := NewTreatmentRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	owner := makeTestOwner(t, db, clinicA, "検索対象飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "検索対象ペット")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-FINDBYID", time.Now())
	makeHistoryTreatment(t, db, mr.ID, model.TreatmentItemTypeMedicine, "対象治療", 0)

	var tr model.Treatment
	require.NoError(t, db.Where("medical_record_id = ?", mr.ID).First(&tr).Error)

	t.Run("found", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, tr.ID)
		require.NoError(t, err)
		assert.Equal(t, tr.ID, got.ID)
	})

	t.Run("not found for nonexistent id", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, uint64(999999))
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("clinic isolation", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, tr.ID)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("soft-deleted treatment is not found", func(t *testing.T) {
		require.NoError(t, db.Delete(&tr).Error)
		got, err := repo.FindByID(ctx, clinicA, tr.ID)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestTreatmentRepository_Create(t *testing.T) {
	db := setupTreatmentHistoryTestDB(t)
	repo := NewTreatmentRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	owner := makeTestOwner(t, db, clinicA, "作成飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "作成ペット")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-CREATE", time.Now())

	tr := &model.Treatment{
		MedicalRecordID: mr.ID,
		ItemType:        model.TreatmentItemTypeOther,
		Content:         "新規治療",
	}
	require.NoError(t, repo.Create(ctx, tr))
	assert.NotZero(t, tr.ID)
}

func TestTreatmentRepository_Update(t *testing.T) {
	db := setupTreatmentHistoryTestDB(t)
	repo := NewTreatmentRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	owner := makeTestOwner(t, db, clinicA, "更新飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "更新ペット")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-UPDATE", time.Now())
	makeHistoryTreatment(t, db, mr.ID, model.TreatmentItemTypeMedicine, "更新前", 0)

	var tr model.Treatment
	require.NoError(t, db.Where("medical_record_id = ?", mr.ID).First(&tr).Error)

	t.Run("updates successfully", func(t *testing.T) {
		require.NoError(t, repo.Update(ctx, clinicA, tr.ID, map[string]any{"content": "更新後"}))
		got, err := repo.FindByID(ctx, clinicA, tr.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後", got.Content)
	})

	t.Run("not found for nonexistent id", func(t *testing.T) {
		err := repo.Update(ctx, clinicA, uint64(999999), map[string]any{"content": "x"})
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("clinic isolation: wrong clinic returns NotFound", func(t *testing.T) {
		err := repo.Update(ctx, clinicB, tr.ID, map[string]any{"content": "乗っ取り"})
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestTreatmentRepository_Delete(t *testing.T) {
	db := setupTreatmentHistoryTestDB(t)
	repo := NewTreatmentRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	owner := makeTestOwner(t, db, clinicA, "削除飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "削除ペット")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-DELETE", time.Now())
	makeHistoryTreatment(t, db, mr.ID, model.TreatmentItemTypeMedicine, "削除対象", 0)

	var tr model.Treatment
	require.NoError(t, db.Where("medical_record_id = ?", mr.ID).First(&tr).Error)

	t.Run("clinic isolation: wrong clinic cannot delete", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, tr.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("soft-deletes successfully", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, tr.ID))

		var unscoped int64
		require.NoError(t, db.Unscoped().Model(&model.Treatment{}).Where("id = ?", tr.ID).Count(&unscoped).Error)
		assert.Equal(t, int64(1), unscoped, "物理的には行が残る（ソフトデリート）")

		var scoped int64
		require.NoError(t, db.Model(&model.Treatment{}).Where("id = ?", tr.ID).Count(&scoped).Error)
		assert.Equal(t, int64(0), scoped)
	})

	t.Run("not found for already-deleted id", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, tr.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestTreatmentRepository_BulkUpdateSortOrder(t *testing.T) {
	db := setupTreatmentHistoryTestDB(t)
	repo := NewTreatmentRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	owner := makeTestOwner(t, db, clinicA, "並び替え飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "並び替えペット")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-SORT", time.Now())
	makeHistoryTreatment(t, db, mr.ID, model.TreatmentItemTypeMedicine, "1番目", 0)
	makeHistoryTreatment(t, db, mr.ID, model.TreatmentItemTypeMedicine, "2番目", 1)

	var treatments []model.Treatment
	require.NoError(t, db.Where("medical_record_id = ?", mr.ID).Order("sort_order ASC").Find(&treatments).Error)
	require.Len(t, treatments, 2)

	t.Run("updates sort_order for matching clinic", func(t *testing.T) {
		updates := []TreatmentSortUpdate{
			{ID: treatments[0].ID, ClinicID: clinicA, MedicalRecordID: mr.ID, SortOrder: 9},
			{ID: treatments[1].ID, ClinicID: clinicA, MedicalRecordID: mr.ID, SortOrder: 8},
		}
		require.NoError(t, repo.BulkUpdateSortOrder(ctx, updates))

		got0, err := repo.FindByID(ctx, clinicA, treatments[0].ID)
		require.NoError(t, err)
		assert.Equal(t, 9, got0.SortOrder)

		got1, err := repo.FindByID(ctx, clinicA, treatments[1].ID)
		require.NoError(t, err)
		assert.Equal(t, 8, got1.SortOrder)
	})

	// MRD-01: 別 clinic / 存在しない ID は NotFound（silent skip 禁止）
	t.Run("wrong clinic_id returns NotFound", func(t *testing.T) {
		updates := []TreatmentSortUpdate{
			{ID: treatments[0].ID, ClinicID: clinicB, MedicalRecordID: mr.ID, SortOrder: 100},
		}
		err := repo.BulkUpdateSortOrder(ctx, updates)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "got %v", err)

		got, err := repo.FindByID(ctx, clinicA, treatments[0].ID)
		require.NoError(t, err)
		assert.Equal(t, 9, got.SortOrder, "別クリニックからの更新は反映されない")
	})

	t.Run("wrong medical_record_id returns NotFound", func(t *testing.T) {
		updates := []TreatmentSortUpdate{
			{ID: treatments[0].ID, ClinicID: clinicA, MedicalRecordID: mr.ID + 999999, SortOrder: 50},
		}
		err := repo.BulkUpdateSortOrder(ctx, updates)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "got %v", err)
	})
}

// TestTreatmentRepository_FindUnbilledByPetID は #77 の「未会計対象化済みだが取り残された」
// treatment 抽出ロジックを検証する: 医師確認(confirmed)済みかつ、その診察カルテ・治療明細のいずれにも
// 未取消の billing/billing_item が紐付いていない treatment のみを返す。
func TestTreatmentRepository_FindUnbilledByPetID(t *testing.T) {
	db := setupTreatmentBillingTestDB(t)
	repo := NewTreatmentRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	owner := makeTestOwner(t, db, clinicA, "未会計飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "未会計ペット")

	t.Run("returns treatments for confirmed, unbilled medical record", func(t *testing.T) {
		mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-UNBILLED-1", time.Now())
		makeHistoryTreatment(t, db, mr.ID, model.TreatmentItemTypeMedicine, "未会計治療", 0)
		require.NoError(t, db.Create(&model.BillingConfirmation{
			MedicalRecordID: mr.ID,
			Status:          model.ConfirmationStatusConfirmed,
		}).Error)

		got, err := repo.FindUnbilledByPetID(ctx, clinicA, pet.ID)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "未会計治療", got[0].Content)
	})

	t.Run("excludes treatment when a non-cancelled billing already exists for the medical record", func(t *testing.T) {
		mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-BILLED", time.Now())
		makeHistoryTreatment(t, db, mr.ID, model.TreatmentItemTypeMedicine, "会計済み治療", 0)
		require.NoError(t, db.Create(&model.BillingConfirmation{
			MedicalRecordID: mr.ID,
			Status:          model.ConfirmationStatusConfirmed,
		}).Error)
		require.NoError(t, db.Create(&model.Billing{
			ClinicID:        clinicA,
			MedicalRecordID: &mr.ID,
			Status:          model.BillingStatusWaiting,
			ScheduledDate:   time.Now(),
		}).Error)

		got, err := repo.FindUnbilledByPetID(ctx, clinicA, pet.ID)
		require.NoError(t, err)
		for _, item := range got {
			assert.NotEqual(t, "会計済み治療", item.Content)
		}
	})

	t.Run("includes treatment when the only billing is cancelled", func(t *testing.T) {
		mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-CANCELLED-BILLING", time.Now())
		makeHistoryTreatment(t, db, mr.ID, model.TreatmentItemTypeMedicine, "取消済会計の治療", 0)
		require.NoError(t, db.Create(&model.BillingConfirmation{
			MedicalRecordID: mr.ID,
			Status:          model.ConfirmationStatusConfirmed,
		}).Error)
		require.NoError(t, db.Create(&model.Billing{
			ClinicID:        clinicA,
			MedicalRecordID: &mr.ID,
			Status:          model.BillingStatusCancelled,
			ScheduledDate:   time.Now(),
		}).Error)

		got, err := repo.FindUnbilledByPetID(ctx, clinicA, pet.ID)
		require.NoError(t, err)
		var found bool
		for _, item := range got {
			if item.Content == "取消済会計の治療" {
				found = true
			}
		}
		assert.True(t, found, "会計が取消済みなら未会計対象化候補として再度返るべき")
	})

	t.Run("excludes medical record without confirmed billing_confirmation", func(t *testing.T) {
		mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-PENDING-CONFIRM", time.Now())
		makeHistoryTreatment(t, db, mr.ID, model.TreatmentItemTypeMedicine, "未確認治療", 0)
		require.NoError(t, db.Create(&model.BillingConfirmation{
			MedicalRecordID: mr.ID,
			Status:          model.ConfirmationStatusPending,
		}).Error)

		got, err := repo.FindUnbilledByPetID(ctx, clinicA, pet.ID)
		require.NoError(t, err)
		for _, item := range got {
			assert.NotEqual(t, "未確認治療", item.Content)
		}
	})

	t.Run("excludes treatment directly referenced by a billing_item via a non-cancelled billing", func(t *testing.T) {
		mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-ITEM-LEVEL", time.Now())
		makeHistoryTreatment(t, db, mr.ID, model.TreatmentItemTypeMedicine, "アイテム紐付け治療", 0)
		require.NoError(t, db.Create(&model.BillingConfirmation{
			MedicalRecordID: mr.ID,
			Status:          model.ConfirmationStatusConfirmed,
		}).Error)

		var tr model.Treatment
		require.NoError(t, db.Where("medical_record_id = ?", mr.ID).First(&tr).Error)

		// 別カルテの billing に、当該 treatment を明細として直接紐付ける
		// （billing_items.treatment_id の NOT EXISTS 分岐を検証する構成）。
		mrOther := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-ITEM-LEVEL-BILLING", time.Now())
		billing := &model.Billing{
			ClinicID:        clinicA,
			MedicalRecordID: &mrOther.ID,
			Status:          model.BillingStatusWaiting,
			ScheduledDate:   time.Now(),
		}
		require.NoError(t, db.Create(billing).Error)
		require.NoError(t, db.Create(&model.BillingItem{
			BillingID:   billing.ID,
			Category:    model.ItemCategoryOther,
			Name:        "直接紐付け明細",
			TreatmentID: &tr.ID,
		}).Error)

		got, err := repo.FindUnbilledByPetID(ctx, clinicA, pet.ID)
		require.NoError(t, err)
		for _, item := range got {
			assert.NotEqual(t, "アイテム紐付け治療", item.Content)
		}
	})

	t.Run("foreign-clinic billing for the medical record does not suppress clinic treatment", func(t *testing.T) {
		mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-FOREIGN-BILLING", time.Now())
		makeHistoryTreatment(t, db, mr.ID, model.TreatmentItemTypeMedicine, "別院会計汚染の治療", 0)
		require.NoError(t, db.Create(&model.BillingConfirmation{
			MedicalRecordID: mr.ID,
			Status:          model.ConfirmationStatusConfirmed,
		}).Error)
		require.NoError(t, db.Create(&model.Billing{
			ClinicID:        clinicB,
			MedicalRecordID: &mr.ID,
			Status:          model.BillingStatusWaiting,
			ScheduledDate:   time.Now(),
		}).Error)

		got, err := repo.FindUnbilledByPetID(ctx, clinicA, pet.ID)
		require.NoError(t, err)
		assert.True(t, treatmentContentExists(got, "別院会計汚染の治療"))
	})

	t.Run("foreign-clinic billing item does not suppress clinic treatment", func(t *testing.T) {
		mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-FOREIGN-BILLING-ITEM", time.Now())
		makeHistoryTreatment(t, db, mr.ID, model.TreatmentItemTypeMedicine, "別院会計明細汚染の治療", 0)
		require.NoError(t, db.Create(&model.BillingConfirmation{
			MedicalRecordID: mr.ID,
			Status:          model.ConfirmationStatusConfirmed,
		}).Error)

		var tr model.Treatment
		require.NoError(t, db.Where("medical_record_id = ?", mr.ID).First(&tr).Error)
		billing := &model.Billing{
			ClinicID:      clinicB,
			Status:        model.BillingStatusWaiting,
			ScheduledDate: time.Now(),
		}
		require.NoError(t, db.Create(billing).Error)
		require.NoError(t, db.Create(&model.BillingItem{
			BillingID:   billing.ID,
			Category:    model.ItemCategoryOther,
			Name:        "別院からの汚染明細",
			TreatmentID: &tr.ID,
		}).Error)

		got, err := repo.FindUnbilledByPetID(ctx, clinicA, pet.ID)
		require.NoError(t, err)
		assert.True(t, treatmentContentExists(got, "別院会計明細汚染の治療"))
	})

	t.Run("clinic isolation: other clinic scope returns empty", func(t *testing.T) {
		got, err := repo.FindUnbilledByPetID(ctx, clinicB, pet.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestTreatmentRepository_FindUnbilledByPetID_RejectsCorruptCrossClinicPetRelation(t *testing.T) {
	db := setupTreatmentBillingTestDB(t)
	repo := NewTreatmentRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "未会計取得対象飼主")
	crossClinicPet := makeSpeciesAndPet(t, db, clinicB, ownerA.ID, "破損した別医院ペット")
	mr := makeHistoryMedicalRecord(t, db, clinicA, crossClinicPet.ID, "MR-CORRUPT-UNBILLED", time.Now())
	makeHistoryTreatment(t, db, mr.ID, model.TreatmentItemTypeMedicine, "別医院ペット由来治療", 0)
	require.NoError(t, db.Create(&model.BillingConfirmation{
		MedicalRecordID: mr.ID,
		Status:          model.ConfirmationStatusConfirmed,
	}).Error)

	got, err := repo.FindUnbilledByPetID(ctx, clinicA, crossClinicPet.ID)

	require.NoError(t, err)
	assert.Empty(t, got)
}

func treatmentContentExists(treatments []model.Treatment, content string) bool {
	for _, treatment := range treatments {
		if treatment.Content == content {
			return true
		}
	}
	return false
}

// TestTreatmentRepository_CountFinalizedUnconfirmedByPetAndDate は #77 の
// 「同日同ペットの取り残し診察カルテ件数」ロジックを検証する。
func TestTreatmentRepository_CountFinalizedUnconfirmedByPetAndDate(t *testing.T) {
	db := setupTreatmentBillingTestDB(t)
	repo := NewTreatmentRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	owner := makeTestOwner(t, db, clinicA, "取り残し飼主")
	date := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	t.Run("counts finalized medical record with no billing_confirmation and no billing", func(t *testing.T) {
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "取り残しペット1")
		makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-LEFTOVER-1", date)

		count, err := repo.CountFinalizedUnconfirmedByPetAndDate(ctx, clinicA, pet.ID, date)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("excludes when billing_confirmation is confirmed", func(t *testing.T) {
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "取り残しペット2")
		mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-CONFIRMED", date)
		require.NoError(t, db.Create(&model.BillingConfirmation{
			MedicalRecordID: mr.ID,
			Status:          model.ConfirmationStatusConfirmed,
		}).Error)

		count, err := repo.CountFinalizedUnconfirmedByPetAndDate(ctx, clinicA, pet.ID, date)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count, "確認済みは取り残し候補ではない")
	})

	t.Run("counts when billing_confirmation exists but is not confirmed", func(t *testing.T) {
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "取り残しペット3")
		mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-STILL-PENDING", date)
		require.NoError(t, db.Create(&model.BillingConfirmation{
			MedicalRecordID: mr.ID,
			Status:          model.ConfirmationStatusPending,
		}).Error)

		count, err := repo.CountFinalizedUnconfirmedByPetAndDate(ctx, clinicA, pet.ID, date)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("excludes when a non-cancelled billing already exists", func(t *testing.T) {
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "取り残しペット4")
		mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-ALREADY-BILLED", date)
		require.NoError(t, db.Create(&model.Billing{
			ClinicID:        clinicA,
			MedicalRecordID: &mr.ID,
			Status:          model.BillingStatusWaiting,
			ScheduledDate:   date,
		}).Error)

		count, err := repo.CountFinalizedUnconfirmedByPetAndDate(ctx, clinicA, pet.ID, date)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("foreign-clinic billing does not suppress clinic leftover", func(t *testing.T) {
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "取り残し別院会計汚染ペット")
		mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-FOREIGN-BILLING-LEFTOVER", date)
		require.NoError(t, db.Create(&model.Billing{
			ClinicID:        clinicB,
			MedicalRecordID: &mr.ID,
			Status:          model.BillingStatusWaiting,
			ScheduledDate:   date,
		}).Error)

		count, err := repo.CountFinalizedUnconfirmedByPetAndDate(ctx, clinicA, pet.ID, date)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("excludes draft (non-finalized) medical records", func(t *testing.T) {
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "取り残しペット5")
		petID := pet.ID
		mr := &model.MedicalRecord{
			ClinicID: clinicA,
			RecordNo: "MR-DRAFT",
			Date:     date,
			PetID:    &petID,
			Status:   model.MedicalRecordStatusDraft,
		}
		require.NoError(t, db.Create(mr).Error)

		count, err := repo.CountFinalizedUnconfirmedByPetAndDate(ctx, clinicA, pet.ID, date)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count, "draft は対象外")
	})

	t.Run("excludes different date", func(t *testing.T) {
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "取り残しペット6")
		makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-OTHER-DATE", date)

		otherDate := date.AddDate(0, 0, 1)
		count, err := repo.CountFinalizedUnconfirmedByPetAndDate(ctx, clinicA, pet.ID, otherDate)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("clinic isolation: other clinic scope returns zero", func(t *testing.T) {
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "取り残しペット7")
		makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-CLINIC-ISO", date)

		count, err := repo.CountFinalizedUnconfirmedByPetAndDate(ctx, clinicB, pet.ID, date)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestTreatmentRepository_CountFinalizedUnconfirmedByPetAndDate_RejectsCorruptCrossClinicPetRelation(t *testing.T) {
	db := setupTreatmentBillingTestDB(t)
	repo := NewTreatmentRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "取り残し取得対象飼主")
	crossClinicPet := makeSpeciesAndPet(t, db, clinicB, ownerA.ID, "破損した別医院ペット")
	date := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	makeHistoryMedicalRecord(t, db, clinicA, crossClinicPet.ID, "MR-CORRUPT-COUNT", date)

	count, err := repo.CountFinalizedUnconfirmedByPetAndDate(ctx, clinicA, crossClinicPet.ID, date)

	require.NoError(t, err)
	assert.Zero(t, count)
}
