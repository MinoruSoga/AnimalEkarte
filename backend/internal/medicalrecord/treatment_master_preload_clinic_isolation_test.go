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
	"github.com/animal-ekarte/backend/internal/repository/repotest"
)

// makeMedicineMaster は internal/repository/isolation_test_helpers_test.go の同名ヘルパーの
// 最小限の複製（BE9-2D ④b: 原本は旧 package の他テストが引き続き使うため移動不可）。
func makeMedicineMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Medicine {
	t.Helper()
	m := &model.Medicine{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(m).Error)
	return m
}

// setupTreatmentMasterPreloadTestDB は treatments + マスタ(procedures/medicines)隔離テスト用に DB を整備する。
func setupTreatmentMasterPreloadTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := repotest.SetupTestDB(t)
	require.NoError(t, repotest.EnsureAutoMigrated(db,
		&model.AnimalSpecies{}, &model.Pet{}, &model.Procedure{}, &model.Medicine{}, &model.Treatment{},
	))
	db.Exec("TRUNCATE TABLE treatments CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE procedures CASCADE")
	db.Exec("TRUNCATE TABLE medicines CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

// TestTreatmentRepository_FindByMedicalRecordID_MasterPreloadClinicIsolation は
// clinic A のカルテ配下の treatment が、別クリニック(B)のマスタを指す FK を持っていても
// 別クリニックのマスタを Preload で漏らさないことを検証する。
func TestTreatmentRepository_FindByMedicalRecordID_MasterPreloadClinicIsolation(t *testing.T) {
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
	procedureB := makeProcedure(t, db, clinicB, "医院Bの手技マスタ", model.AnesthesiaTypeNone, false)
	medicineB := makeMedicineMaster(t, db, clinicB, "医院Bの薬剤マスタ")

	// clinic A の treatment に別クリニックのマスタ FK を植え付ける
	// （write 側の clinic 未検証 / 過去のデータ汚染 #124/#125 を再現）。
	cross := &model.Treatment{
		MedicalRecordID: mrA.ID,
		ProcedureID:     &procedureB.ID,
		MedicineID:      &medicineB.ID,
		ItemType:        model.TreatmentItemTypeProcedure,
		Content:         "cross-clinic FK planted",
		SortOrder:       0,
	}
	require.NoError(t, db.WithContext(ctx).Create(cross).Error)

	got, err := repo.FindByMedicalRecordID(ctx, clinicA, mrA.ID)
	require.NoError(t, err)
	require.Len(t, got, 1, "treatment 自体は clinic A 所属なので返る")

	// treatment 行自体は clinic A のものなので返るが、Preload されるマスタは clinic B のもので、
	// 応答に混入してはならない。clinic_id スコープ Preload なら nil になる。
	assert.Nil(t, got[0].Procedure, "別クリニックの手技マスタが Preload で混入してはならない")
	assert.Nil(t, got[0].Medicine, "別クリニックの薬剤マスタが Preload で混入してはならない")
}

// TestTreatmentRepository_FindByMedicalRecordID_SameClinicMasterPreloaded は
// clinic 隔離 Preload を追加しても同一クリニックの正当なマスタは従来どおり Preload されることを検証する
// （修正が正規挙動を壊さないことの担保）。
func TestTreatmentRepository_FindByMedicalRecordID_SameClinicMasterPreloaded(t *testing.T) {
	db := setupTreatmentMasterPreloadTestDB(t)
	repo := NewTreatmentRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)

	ownerA := makeTestOwner(t, db, clinicA, "正常飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "タマA")
	mrA := makeHistoryMedicalRecord(t, db, clinicA, petA.ID, "MR-PRELOAD-OK", time.Now())

	procedureA := makeProcedure(t, db, clinicA, "医院Aの手技", model.AnesthesiaTypeNone, false)
	medicineA := makeMedicineMaster(t, db, clinicA, "医院Aの薬剤")

	legit := &model.Treatment{
		MedicalRecordID: mrA.ID,
		ProcedureID:     &procedureA.ID,
		MedicineID:      &medicineA.ID,
		ItemType:        model.TreatmentItemTypeProcedure,
		Content:         "same-clinic FK",
		SortOrder:       0,
	}
	require.NoError(t, db.WithContext(ctx).Create(legit).Error)

	got, err := repo.FindByMedicalRecordID(ctx, clinicA, mrA.ID)
	require.NoError(t, err)
	require.Len(t, got, 1)

	require.NotNil(t, got[0].Procedure, "同一クリニックの手技マスタは Preload されるべき")
	assert.Equal(t, procedureA.ID, got[0].Procedure.ID)
	require.NotNil(t, got[0].Medicine, "同一クリニックの薬剤マスタは Preload されるべき")
	assert.Equal(t, medicineA.ID, got[0].Medicine.ID)
}
