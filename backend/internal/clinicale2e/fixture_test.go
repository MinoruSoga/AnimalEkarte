package clinicale2e

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func testModels() []any {
	return []any{
		&model.Company{},
		&model.Clinic{},
		&model.Account{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.Owner{},
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.MedicalRecord{},
		&model.ExaminationType{},
		&model.Examination{},
		&model.Vaccine{},
		&model.Vaccination{},
		&model.CheckupType{},
		&model.Checkup{},
		&model.Hospitalization{},
		&model.Estimate{},
		&model.EstimateItem{},
	}
}

func TestCreate_RejectsUnsafeRequest(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()

	t.Run("staging env", func(t *testing.T) {
		_, err := Create(ctx, db, Request{AppEnv: "staging", DBHost: "db", PasswordHash: "x"})
		require.Error(t, err)
		assert.ErrorContains(t, err, "APP_ENV")
	})

	t.Run("empty password hash", func(t *testing.T) {
		_, err := Create(ctx, db, Request{AppEnv: "test", DBHost: "db"})
		require.Error(t, err)
		assert.ErrorContains(t, err, "password hash")
	})

	t.Run("reserved clinic remains rejected", func(t *testing.T) {
		require.Error(t, RejectReservedClinicID(1))
	})
}

func TestCreateAndDelete_DisposableClinicGraph(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, testModels()...))
	ctx := context.Background()

	got, err := Create(ctx, db, Request{AppEnv: "test", DBHost: "db", PasswordHash: "test-hash-not-for-login"})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NoError(t, RejectReservedClinicID(got.ClinicID))
	assert.GreaterOrEqual(t, got.ClinicID, clinicIDBase)
	assert.Equal(t, ownerSearchToken, got.OwnerSearch)
	assert.Contains(t, got.OwnerName, ownerSearchToken)
	assert.NotZero(t, got.PetID)
	assert.NotZero(t, got.OutsideFirstPagePetID)
	assert.NotEqual(t, got.PetID, got.OutsideFirstPagePetID)
	assert.Contains(t, got.OutsideFirstPagePetName, outsideFirstPagePrefix)
	assert.Equal(t, medicalRecordCount, got.MedicalRecordCount)
	assert.Contains(t, got.EstimateTitle, "e2e-est-")

	var clinic model.Clinic
	require.NoError(t, db.First(&clinic, got.ClinicID).Error)
	assert.True(t, clinic.IsActive)

	var account model.Account
	require.NoError(t, db.Where("email = ?", LoginEmail(got.ClinicID)).First(&account).Error)
	assert.True(t, account.IsSystemAdmin)
	assert.Equal(t, "test-hash-not-for-login", account.PasswordHash)

	var pets []model.Pet
	require.NoError(t, db.Where("clinic_id = ?", got.ClinicID).Order("id ASC").Find(&pets).Error)
	require.Len(t, pets, padPetCount+2)
	assert.Equal(t, got.PetID, pets[0].ID)
	assert.Equal(t, got.OutsideFirstPagePetID, pets[len(pets)-1].ID)

	var records []model.MedicalRecord
	require.NoError(t, db.Where("clinic_id = ?", got.ClinicID).Find(&records).Error)
	require.Len(t, records, medicalRecordCount)
	for _, record := range records {
		assert.Equal(t, model.MedicalRecordStatusFinalized, record.Status)
	}

	var hospitalizations []model.Hospitalization
	require.NoError(t, db.Where("clinic_id = ?", got.ClinicID).Find(&hospitalizations).Error)
	require.Len(t, hospitalizations, 1)
	assert.Equal(t, model.HospitalizationStatusAdmitted, hospitalizations[0].Status)

	encoded, err := EncodeResult(got)
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "password")
	assert.NotContains(t, string(encoded), account.PasswordHash)
	var decoded Result
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	assert.Equal(t, got.ClinicID, decoded.ClinicID)

	require.NoError(t, Delete(ctx, db, "test", "db", got.ClinicID))
	var leftover model.Clinic
	err = db.First(&leftover, got.ClinicID).Error
	require.Error(t, err)
	var leftoverPets int64
	require.NoError(t, db.Model(&model.Pet{}).Where("clinic_id = ?", got.ClinicID).Count(&leftoverPets).Error)
	assert.Zero(t, leftoverPets)
}

func TestDelete_RejectsReservedAndMismatchedClinic(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Company{}, &model.Clinic{}))
	ctx := context.Background()

	err := Delete(ctx, db, "test", "db", 1)
	require.Error(t, err)
	assert.ErrorContains(t, err, "reserved")

	company := &model.Company{Name: "not-e2e-company"}
	require.NoError(t, db.Create(company).Error)
	foreign := &model.Clinic{
		ID:        clinicIDBase + uint64(time.Now().UnixNano()%7000) + 50,
		CompanyID: company.ID,
		Name:      "foreign-clinic",
		IsActive:  true,
	}
	require.NoError(t, db.Create(foreign).Error)
	err = Delete(ctx, db, "test", "db", foreign.ID)
	require.Error(t, err)
	assert.ErrorContains(t, err, "not a clinical e2e fixture")
	var still model.Clinic
	require.NoError(t, db.First(&still, foreign.ID).Error)
}
