package medicalrecord

// 移動元 internal/repository（BE9-2D sub-batch④b）。setupTestDB/ensureAutoMigrated は
// repotest 直呼びへ、makeTestOwner/makeSpeciesAndPet は本 package の共有ヘルパを再利用。

// treatment_master_preload_clinic_isolation_test.go
// クロステナント READ IDOR 監査（72e8887c write 監査の read 版）回帰テスト。
//
// 保護する不変条件: treatments は自前 clinic_id を持たず medical_records.clinic_id で
// テナント隔離されるが、Preload する「マスタ」(Procedure / Medicine / Consultation) は
// FK 値 (treatment.procedure_id 等) で引かれる。FK 値が別クリニックのマスタを指す場合
// (write 側が FK の clinic 所有権を未検証 / 過去のシードデータ汚染 #124/#125)、
// Preload に clinic_id 述語が無いと別クリニックのマスタ名・価格が応答に混入する。
//
// このテストは Preload の "clinic_id = ?" スコープを外すと必ず失敗するよう設計されている。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// makeMedicineMaster は internal/repository/isolation_test_helpers_test.go の同名ヘルパーの
// 最小限の複製（BE9-2D ④b: 原本は旧 package の他テストが引き続き使うため移動不可）。
func makeMedicineMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Medicine {
	t.Helper()
	m := &model.Medicine{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(m).Error)
	return m
}

// setupTreatmentMasterPreloadTestDB は treatments + nullable マスタ隔離テスト用に DB を整備する。
func setupTreatmentMasterPreloadTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.Consultation{},
		&model.Procedure{},
		&model.Medicine{},
		&model.InventoryItem{},
		&model.Treatment{},
	))
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

func makeTreatmentConsultationMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Consultation {
	t.Helper()
	consultation := &model.Consultation{ClinicID: clinicID, Name: name, IsActive: true}
	require.NoError(t, db.Create(consultation).Error)
	return consultation
}

func makeTreatmentInventoryMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.InventoryItem {
	t.Helper()
	item := &model.InventoryItem{
		ClinicID: clinicID, Name: name, Category: model.InventoryCategoryOther,
		Quantity: 1, Unit: "個", Status: model.InventoryStatusSufficient,
	}
	require.NoError(t, db.Create(item).Error)
	return item
}

// TestDB_TreatmentRepositoryClearsForeignNullableMasterFKs は
// clinic A のカルテ配下の treatment が、別クリニック(B)のマスタを指す FK を持っていても
// association だけでなく raw FK も response へ漏らさないことを検証する。
func TestDB_TreatmentRepositoryClearsForeignNullableMasterFKs(t *testing.T) {
	db := setupTreatmentMasterPreloadTestDB(t)
	repo := NewTreatmentRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	// clinic A のカルテ（ペットは clinic A）
	ownerA := makeTestOwner(t, db, clinicA, "隔離飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ポチA")
	mrA := makeHistoryMedicalRecord(t, db, clinicA, petA.ID, "MR-PRELOAD-A", time.Now())

	// clinic B のマスタ（漏れてはならないクロステナントデータ）
	consultationB := makeTreatmentConsultationMaster(t, db, clinicB, "医院Bの診察マスタ")
	procedureB := makeProcedure(t, db, clinicB, "医院Bの手技マスタ", model.AnesthesiaTypeNone, false)
	medicineB := makeMedicineMaster(t, db, clinicB, "医院Bの薬剤マスタ")
	inventoryB := makeTreatmentInventoryMaster(t, db, clinicB, "医院Bの在庫マスタ")

	// clinic A の treatment に別クリニックのマスタ FK を植え付ける
	// （write 側の clinic 未検証 / 過去のデータ汚染 #124/#125 を再現）。
	cross := &model.Treatment{
		MedicalRecordID: mrA.ID,
		ConsultationID:  &consultationB.ID,
		ProcedureID:     &procedureB.ID,
		MedicineID:      &medicineB.ID,
		InventoryID:     &inventoryB.ID,
		ItemType:        model.TreatmentItemTypeOther,
		Content:         "cross-clinic FK planted",
		SortOrder:       0,
	}
	require.NoError(t, db.WithContext(ctx).Create(cross).Error)

	got, err := repo.FindByMedicalRecordID(ctx, clinicA, mrA.ID)
	require.NoError(t, err)
	require.Len(t, got, 1, "treatment 自体は clinic A 所属なので返る")

	assert.Nil(t, got[0].ConsultationID)
	assert.Nil(t, got[0].ProcedureID)
	assert.Nil(t, got[0].MedicineID)
	assert.Nil(t, got[0].InventoryID)
	assert.Nil(t, got[0].Consultation)
	assert.Nil(t, got[0].Procedure, "別クリニックの手技マスタが Preload で混入してはならない")
	assert.Nil(t, got[0].Medicine, "別クリニックの薬剤マスタが Preload で混入してはならない")
	assert.Nil(t, got[0].Inventory)

	response := toTreatmentResponse(&got[0])
	assert.Nil(t, response.ConsultationID)
	assert.Nil(t, response.ProcedureID)
	assert.Nil(t, response.MedicineID)
	assert.Nil(t, response.InventoryID)

	byID, err := repo.FindByID(ctx, clinicA, cross.ID)
	require.NoError(t, err)
	assert.Nil(t, byID.ConsultationID)
	assert.Nil(t, byID.ProcedureID)
	assert.Nil(t, byID.MedicineID)
	assert.Nil(t, byID.InventoryID)

	history, total, err := repo.FindHistoryByPetID(
		ctx,
		clinicA,
		petA.ID,
		model.PetTreatmentHistoryFilter{},
		1,
		100,
	)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, history, 1)
	historyResponse := toPetTreatmentHistoryResponse(&history[0])
	assert.Nil(t, historyResponse.ProcedureID)
	assert.Nil(t, historyResponse.ProcedureName)
	assert.Nil(t, historyResponse.MedicineID)
	assert.Nil(t, historyResponse.MedicineName)
}

// TestDB_TreatmentRepositoryRetainsSameClinicNullableMasterFKs は
// clinic 隔離 Preload を追加しても同一クリニックの正当なマスタは従来どおり Preload されることを検証する
// （修正が正規挙動を壊さないことの担保）。
func TestDB_TreatmentRepositoryRetainsSameClinicNullableMasterFKs(t *testing.T) {
	db := setupTreatmentMasterPreloadTestDB(t)
	repo := NewTreatmentRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)

	ownerA := makeTestOwner(t, db, clinicA, "正常飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "タマA")
	mrA := makeHistoryMedicalRecord(t, db, clinicA, petA.ID, "MR-PRELOAD-OK", time.Now())

	consultationA := makeTreatmentConsultationMaster(t, db, clinicA, "医院Aの診察")
	procedureA := makeProcedure(t, db, clinicA, "医院Aの手技", model.AnesthesiaTypeNone, false)
	medicineA := makeMedicineMaster(t, db, clinicA, "医院Aの薬剤")
	inventoryA := makeTreatmentInventoryMaster(t, db, clinicA, "医院Aの在庫")

	legit := &model.Treatment{
		MedicalRecordID: mrA.ID,
		ConsultationID:  &consultationA.ID,
		ProcedureID:     &procedureA.ID,
		MedicineID:      &medicineA.ID,
		InventoryID:     &inventoryA.ID,
		ItemType:        model.TreatmentItemTypeOther,
		Content:         "same-clinic FK",
		SortOrder:       0,
	}
	require.NoError(t, db.WithContext(ctx).Create(legit).Error)

	got, err := repo.FindByMedicalRecordID(ctx, clinicA, mrA.ID)
	require.NoError(t, err)
	require.Len(t, got, 1)

	require.NotNil(t, got[0].Consultation)
	assert.Equal(t, consultationA.ID, got[0].Consultation.ID)
	require.NotNil(t, got[0].Procedure, "同一クリニックの手技マスタは Preload されるべき")
	assert.Equal(t, procedureA.ID, got[0].Procedure.ID)
	require.NotNil(t, got[0].Medicine, "同一クリニックの薬剤マスタは Preload されるべき")
	assert.Equal(t, medicineA.ID, got[0].Medicine.ID)
	require.NotNil(t, got[0].Inventory)
	assert.Equal(t, inventoryA.ID, got[0].Inventory.ID)
	require.NotNil(t, got[0].ConsultationID)
	require.NotNil(t, got[0].ProcedureID)
	require.NotNil(t, got[0].MedicineID)
	require.NotNil(t, got[0].InventoryID)

	response := toTreatmentResponse(&got[0])
	require.NotNil(t, response.ConsultationID)
	require.NotNil(t, response.ProcedureID)
	require.NotNil(t, response.MedicineID)
	require.NotNil(t, response.InventoryID)
}
