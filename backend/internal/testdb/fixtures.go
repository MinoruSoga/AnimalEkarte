package testdb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/model"
)

// MakeSpeciesAndPet はテスト用の AnimalSpecies と Pet を作成して返す。
func MakeSpeciesAndPet(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, petName string) *model.Pet {
	t.Helper()
	species := &model.AnimalSpecies{Name: "犬"}
	require.NoError(t, db.WithContext(context.Background()).Create(species).Error)
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         ownerID,
		AnimalSpeciesID: species.ID,
		Name:            petName,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(pet).Error)
	return pet
}

// MakeDoctor はテスト用の医師を作成して返す。
func MakeDoctor(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {
	t.Helper()
	s := &model.Staff{ClinicID: clinicID, Name: name, StaffType: model.StaffTypeDoctor}
	require.NoError(t, db.WithContext(context.Background()).Create(s).Error)
	return s
}

// MakeHistoryMedicalRecord はテスト用の確定済みカルテを作成して返す。
func MakeHistoryMedicalRecord(t *testing.T, db *gorm.DB, clinicID, petID uint64, recordNo string, date time.Time) *model.MedicalRecord {
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

// SeedClinicsForFK は staff_clinic_assignments の外部キーを満たす医院を作成する。
func SeedClinicsForFK(t *testing.T, db *gorm.DB, clinicIDs ...uint64) {
	t.Helper()
	company := &model.Company{Name: "テスト法人(staff preload)"}
	require.NoError(t, db.WithContext(context.Background()).Create(company).Error)
	for _, cid := range clinicIDs {
		c := &model.Clinic{ID: cid, CompanyID: company.ID, Name: fmt.Sprintf("テスト医院%d", cid)}
		require.NoError(t, db.WithContext(context.Background()).Clauses(clause.OnConflict{DoNothing: true}).Create(c).Error)
	}
}

// MakeInsurance はテスト用の保険マスタを作成して返す。
func MakeInsurance(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Insurance {
	t.Helper()
	ins := &model.Insurance{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(ins).Error)
	return ins
}

// MakePetWithInsurance はテスト用の保険付き Pet を作成して返す。
func MakePetWithInsurance(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, insuranceID *uint64, name string) *model.Pet {
	t.Helper()
	species := &model.AnimalSpecies{Name: "犬"}
	require.NoError(t, db.WithContext(context.Background()).Create(species).Error)
	pet := &model.Pet{ClinicID: clinicID, OwnerID: ownerID, AnimalSpeciesID: species.ID, Name: name, InsuranceID: insuranceID}
	require.NoError(t, db.WithContext(context.Background()).Create(pet).Error)
	return pet
}
