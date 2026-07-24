package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// makeMedicineMaster / makeDoctor は clinic 隔離テスト群で共有するマスタ生成ヘルパー。
// （旧 medical_record_owner_medication_test.go に同居していたが、#158 エンドポイント削除に伴い分離）

func makeMedicineMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Medicine {
	t.Helper()
	m := &model.Medicine{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(m).Error)
	return m
}

// makeHistoryMedicalRecord は旧 treatment_repository_test.go の同名ヘルパーの最小限の複製
// （BE9-2D ④b: 移動後も master_preload_clinic_isolation_test.go が本パッケージから参照するため）。
func makeHistoryMedicalRecord(t *testing.T, db *gorm.DB, clinicID, petID uint64, recordNo string, date time.Time) *model.MedicalRecord {
	t.Helper()
	pet := petID
	mr := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: recordNo,
		Date:     date,
		PetID:    &pet,
		Status:   model.MedicalRecordStatusFinalized,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(mr).Error)
	return mr
}

// makeProcedure は旧 treatment_repository_test.go の同名ヘルパーの最小限の複製
// （BE9-2D ④b: treatment repository test の medicalrecord 移動後も care_plan_item /
// master_preload の各テストが本パッケージから引き続き参照するため。makeShiftEntryWithType 先例）。
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

func makeDoctor(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {
	t.Helper()
	s := &model.Staff{ClinicID: clinicID, Name: name, StaffType: model.StaffTypeDoctor}
	require.NoError(t, db.WithContext(context.Background()).Create(s).Error)
	return s
}

// makeClinicScopedClinicalReadParents builds a fully consistent owner → pet →
// medical-record graph plus an active doctor assignment. Read-isolation tests
// can then make one relation intentionally cross-clinic without accidentally
// failing a different patient/staff predicate first.
func makeClinicScopedClinicalReadParents(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	label string,
) (*model.Pet, *model.MedicalRecord, *model.Staff) {
	t.Helper()

	seedClinicsForFK(t, db, clinicID)
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

// makeShiftEntryWithType は shiftentry/repository_test.go の同名ヘルパーの最小限の複製
// （BE8-4 batch13: shift_entry_repository_test.go の移動先パッケージから本パッケージの
// reservation_type_occupation_repository_test.go が引き続き参照するため）。
func makeShiftEntryWithType(t *testing.T, db *gorm.DB, clinicID, staffID uint64, date time.Time, shiftType model.ShiftType) *model.ShiftEntry {
	t.Helper()
	e := &model.ShiftEntry{
		ClinicID:  clinicID,
		StaffID:   staffID,
		Date:      date,
		ShiftType: shiftType,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(e).Error)
	return e
}
