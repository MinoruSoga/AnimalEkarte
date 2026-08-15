package medicalrecord

// master_preload_clinic_isolation_test.go
// クロステナント READ IDOR remediation follow-up — (b) single-clinic master Preload 隔離回帰テスト。
//
// 保護する不変条件: clinic-scoped マスタを FK 値で Preload する際、別クリニックの
// マスタ(名前/価格等)を応答へ混入させない。必須マスタを指す Checkup は、現在の
// relation scope により汚染行そのものを fail-closed で除外し、同一クリニックの
// マスタと整合した患者・医師関係は従来どおり Preload する。

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

// --- inline master/parent creators (既存 helper に無いもののみ) ---

// makeClinicScopedClinicalReadParents builds a fully consistent owner → pet →
// medical-record graph plus an active doctor assignment.
func makeClinicScopedClinicalReadParents(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	label string,
) (*model.Pet, *model.MedicalRecord, *model.Staff) {
	t.Helper()

	testdb.SeedClinicsForFK(t, db, clinicID)
	owner := makeTestOwner(t, db, clinicID, label+"飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, label+"ペット")
	record := makeHistoryMedicalRecord(t, db, clinicID, pet.ID, label+"-MR", time.Now())
	require.NoError(t, db.WithContext(context.Background()).Model(record).Update("owner_id", owner.ID).Error)
	record.OwnerID = &owner.ID

	doctor := makeDoctor(t, db, clinicID, label+"医師")
	require.NoError(t, db.WithContext(context.Background()).Create(&model.StaffClinicAssignment{
		StaffID:  doctor.ID,
		ClinicID: clinicID,
		IsMain:   true,
	}).Error)

	return pet, record, doctor
}

// --- (b1) hospitalization: Cage ---

// FindAll を使う（Cage を Preload するが CarePlanItems/DailyRecords/TreatmentPlans は Preload しないため
// テスト DB の migrate 範囲が最小で済む。検証対象は L56 の Cage Preload clinic スコープ）。
func TestHospitalizationRepository_FindAll_CagePreloadClinicIsolation(t *testing.T) {
	db := testdb.SetupTestDB(t)
	// TRUNCATE first: 既存の null 制約違反行を除去してから AutoMigrate（cages は既存スキーマに存在しうる）。
	db.Exec("TRUNCATE TABLE hospitalizations CASCADE")
	db.Exec("TRUNCATE TABLE cages CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}, &model.Cage{}, &model.Hospitalization{}))
	repo := NewHospitalizationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "入院飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "入院ポチA")
	cageB := makeCageMaster(t, db, clinicB, "医院Bのケージ")
	cageA := makeCageMaster(t, db, clinicA, "医院Aのケージ")

	cross := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, &cageB.ID)
	legit := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, &cageA.ID)

	got, _, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, nil, 1, 100)
	require.NoError(t, err)

	byID := map[uint64]model.Hospitalization{}
	for _, h := range got {
		byID[h.ID] = h
	}
	require.Contains(t, byID, cross.ID)
	require.Contains(t, byID, legit.ID)
	assert.Nil(t, byID[cross.ID].Cage, "別クリニックのケージマスタが Preload で混入してはならない")
	require.NotNil(t, byID[legit.ID].Cage, "同一クリニックのケージは Preload されるべき")
	assert.Equal(t, cageA.ID, byID[legit.ID].Cage.ID)
}

// --- (b2) checkup: CheckupType ---

func TestCheckupRepository_FindByID_CheckupTypePreloadClinicIsolation(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Company{},
		&model.Clinic{},
		&model.Owner{},
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.MedicalRecord{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.CheckupType{},
		&model.Checkup{},
	))
	db.Exec("TRUNCATE TABLE checkups CASCADE")
	db.Exec("TRUNCATE TABLE checkup_types CASCADE")
	db.Exec("TRUNCATE TABLE staff_clinic_assignments CASCADE")
	db.Exec("TRUNCATE TABLE staffs CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	repo := NewCheckupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	petA, mrA, doctorA := makeClinicScopedClinicalReadParents(t, db, clinicA, "健診")
	typeB := makeCheckupTypeMaster(t, db, clinicB, "医院Bの健診種別")
	typeA := makeCheckupTypeMaster(t, db, clinicA, "医院Aの健診種別")

	crossID := makePreloadCheckupRec(t, db, clinicA, mrA.ID, petA.ID, doctorA.ID, typeB.ID)
	legitID := makePreloadCheckupRec(t, db, clinicA, mrA.ID, petA.ID, doctorA.ID, typeA.ID)

	tests := []struct {
		name         string
		id           uint64
		wantNotFound bool
		wantTypeID   uint64
	}{
		{
			name:         "別クリニックの必須健診種別を指す行は取得対象外",
			id:           crossID,
			wantNotFound: true,
		},
		{
			name:       "同一クリニックの健診種別と整合した患者医師関係を取得",
			id:         legitID,
			wantTypeID: typeA.ID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.FindByID(ctx, clinicA, tt.id)
			if tt.wantNotFound {
				require.Error(t, err)
				assert.True(t, apperrors.IsNotFound(err))
				assert.Nil(t, got, "別クリニックの健診種別を参照する行を返してはならない")
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got.CheckupType, "同一クリニックの健診種別は Preload されるべき")
			assert.Equal(t, tt.wantTypeID, got.CheckupType.ID)
			require.NotNil(t, got.MedicalRecord)
			require.NotNil(t, got.MedicalRecord.Pet)
			require.NotNil(t, got.MedicalRecord.Pet.Owner)
			require.NotNil(t, got.Doctor)
			assert.Equal(t, doctorA.ID, got.Doctor.ID)
		})
	}
}

func makePreloadCheckupRec(t *testing.T, db *gorm.DB, clinicID, mrID, petID, doctorID, checkupTypeID uint64) uint64 {
	t.Helper()
	pid := petID
	did := doctorID
	c := &model.Checkup{
		ClinicID:        clinicID,
		MedicalRecordID: mrID,
		PetID:           &pid,
		DoctorID:        &did,
		CheckupTypeID:   checkupTypeID,
		Date:            time.Now(),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(c).Error)
	return c.ID
}

// --- (b3) care_plan_item: Medicine / Procedure ---

func TestCarePlanItemRepository_FindByID_MasterPreloadClinicIsolation(t *testing.T) {
	db := testdb.SetupTestDB(t)
	// TRUNCATE first: 既存の null 制約違反行を除去してから AutoMigrate（cages は既存スキーマに存在しうる）。
	db.Exec("TRUNCATE TABLE care_plan_items CASCADE")
	db.Exec("TRUNCATE TABLE hospitalizations CASCADE")
	db.Exec("TRUNCATE TABLE cages CASCADE")
	db.Exec("TRUNCATE TABLE medicines CASCADE")
	db.Exec("TRUNCATE TABLE procedures CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.AnimalSpecies{}, &model.Pet{}, &model.Cage{}, &model.Hospitalization{},
		&model.Medicine{}, &model.Procedure{}, &model.CarePlanItem{},
	))
	repo := NewCarePlanItemRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "ケアプラン飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ケアプランポチA")
	hospA := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)
	medB := makeMedicineMaster(t, db, clinicB, "医院Bの薬剤")
	procB := makeProcedure(t, db, clinicB, "医院Bの手技", model.AnesthesiaTypeNone, false)

	item := &model.CarePlanItem{
		HospitalizationID: hospA.ID,
		Type:              model.CarePlanTypeMedicine,
		Name:              "cross-clinic care item",
		MedicineID:        &medB.ID,
		ProcedureID:       &procB.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(item).Error)

	got, err := repo.FindByID(ctx, clinicA, item.ID)
	require.NoError(t, err)
	assert.Nil(t, got.Medicine, "別クリニックの薬剤マスタが Preload で混入してはならない")
	assert.Nil(t, got.Procedure, "別クリニックの手技マスタが Preload で混入してはならない")
}

// --- (b4) clinical_plan: DiagnosisType / DiagnosisName ---

func TestClinicalPlanRepository_FindByMedicalRecordID_DiagnosisPreloadClinicIsolation(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.AnimalSpecies{}, &model.Pet{}, &model.DiagnosisType{}, &model.DiagnosisName{}, &model.ClinicalPlan{},
	))
	db.Exec("TRUNCATE TABLE clinical_plans CASCADE")
	db.Exec("TRUNCATE TABLE diagnosis_names CASCADE")
	db.Exec("TRUNCATE TABLE diagnosis_types CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	repo := NewClinicalPlanRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "診断計画飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "診断計画ポチA")
	mrA := makeHistoryMedicalRecord(t, db, clinicA, petA.ID, "MR-CP-A", time.Now())
	typeB := makeDiagnosisTypeMaster(t, db, clinicB, "医院Bの診断分類")
	nameB := makeDiagnosisNameRec(t, db, clinicB, typeB.ID, "医院Bの診断名")

	plan := &model.ClinicalPlan{
		MedicalRecordID: mrA.ID,
		DiagnosisTypeID: &typeB.ID,
		DiagnosisNameID: &nameB.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(plan).Error)

	got, err := repo.FindByMedicalRecordID(ctx, clinicA, mrA.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Nil(t, got.DiagnosisType, "別クリニックの診断分類マスタが Preload で混入してはならない")
	assert.Nil(t, got.DiagnosisName, "別クリニックの診断名マスタが Preload で混入してはならない")
}

// --- (b6) diagnosis: Names (子マスタ) ---

func TestDiagnosisTypeRepository_FindAll_NamesPreloadClinicIsolation(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.DiagnosisType{}, &model.DiagnosisName{}))
	db.Exec("TRUNCATE TABLE diagnosis_names CASCADE")
	db.Exec("TRUNCATE TABLE diagnosis_types CASCADE")
	repo := NewDiagnosisTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	typeA := makeDiagnosisTypeMaster(t, db, clinicA, "医院Aの診断分類")
	// clinic A の type を指す clinic B の診断名（クロステナント子の植え付け）
	nameCross := makeDiagnosisNameRec(t, db, clinicB, typeA.ID, "医院Bの診断名(越境)")
	nameLegit := makeDiagnosisNameRec(t, db, clinicA, typeA.ID, "医院Aの診断名")

	got, _, err := repo.FindAll(ctx, clinicA, 1, 100)
	require.NoError(t, err)

	var found *model.DiagnosisType
	for i := range got {
		if got[i].ID == typeA.ID {
			found = &got[i]
			break
		}
	}
	require.NotNil(t, found, "clinic A の診断分類は取得できる")
	for _, n := range found.Names {
		assert.NotEqual(t, nameCross.ID, n.ID, "別クリニックの診断名が Names Preload で混入してはならない")
		assert.Equal(t, clinicA, n.ClinicID, "Preload された診断名は全て clinic A 所属であるべき")
	}
	// 同一クリニックの診断名は混入する（非破壊）
	var sawLegit bool
	for _, n := range found.Names {
		if n.ID == nameLegit.ID {
			sawLegit = true
		}
	}
	assert.True(t, sawLegit, "同一クリニックの診断名は Names に含まれるべき")
}
