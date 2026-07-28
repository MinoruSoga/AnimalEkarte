package medicalrecord

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToHospitalizationResponse(t *testing.T) {
	startDate := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 5, 27, 9, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 29, 9, 0, 0, 0, time.UTC)
	cageID := uint64(10)
	doctorID := uint64(20)
	insuranceCompanyName := "Pet Insurance"
	insuranceNumber := "INS-001"

	resp := toHospitalizationResponse(&model.Hospitalization{
		ID:                   1,
		ClinicID:             2,
		OwnerID:              3,
		PetID:                4,
		HospitalizationType:  model.HospitalizationTypeHotel,
		StartDate:            startDate,
		EndDate:              endDate,
		Status:               model.HospitalizationStatusAdmitted,
		CageID:               &cageID,
		DoctorID:             &doctorID,
		InsuranceCompanyName: &insuranceCompanyName,
		InsuranceNumber:      &insuranceNumber,
		Memo:                 "経過観察",
		OwnerRequest:         "静かな部屋",
		StaffNotes:           "食欲あり",
		CreatedAt:            createdAt,
		UpdatedAt:            updatedAt,
		Owner:                &model.Owner{ID: 3, Name: "飼主太郎"},
		Pet:                  &model.Pet{ID: 4, Name: "ポチ"},
		Doctor:               &model.Staff{ID: 20, Name: "先生"},
	})

	assert.Equal(t, uint64(1), resp.ID)
	assert.Equal(t, uint64(2), resp.ClinicID)
	assert.Equal(t, uint64(3), resp.OwnerID)
	assert.Equal(t, uint64(4), resp.PetID)
	assert.Equal(t, string(model.HospitalizationTypeHotel), resp.HospitalizationType)
	assert.Equal(t, string(model.HospitalizationStatusAdmitted), resp.Status)
	assert.Equal(t, &cageID, resp.CageID)
	assert.Equal(t, &doctorID, resp.DoctorID)
	assert.Equal(t, &insuranceCompanyName, resp.InsuranceCompanyName)
	assert.Equal(t, &insuranceNumber, resp.InsuranceNumber)
	assert.Equal(t, "経過観察", resp.Memo)
	assert.Equal(t, "静かな部屋", resp.OwnerRequest)
	assert.Equal(t, "食欲あり", resp.StaffNotes)
	require.NotNil(t, resp.Owner)
	assert.Equal(t, "飼主太郎", resp.Owner.OwnerName)
	require.NotNil(t, resp.Pet)
	assert.Equal(t, "ポチ", resp.Pet.Name)
	require.NotNil(t, resp.Doctor)
	assert.Equal(t, "先生", resp.Doctor.Name)
}

func TestToHospitalizationResponse_NilRelations(t *testing.T) {
	resp := toHospitalizationResponse(&model.Hospitalization{
		ID:                   5,
		HospitalizationType:  model.HospitalizationTypeInpatient,
		Status:               model.HospitalizationStatusReserved,
		CageID:               nil,
		DoctorID:             nil,
		InsuranceCompanyName: nil,
		InsuranceNumber:      nil,
	})

	assert.Equal(t, uint64(5), resp.ID)
	assert.Nil(t, resp.CageID)
	assert.Nil(t, resp.DoctorID)
	assert.Nil(t, resp.InsuranceCompanyName)
	assert.Nil(t, resp.InsuranceNumber)
	assert.Nil(t, resp.Owner)
	assert.Nil(t, resp.Pet)
	assert.Nil(t, resp.Doctor)
}

func TestToDischargeWithBillingResponse(t *testing.T) {
	accountingID := uint64(42)

	resp := toDischargeWithBillingResponse(&DischargeWithBillingResult{
		HospitalizationID: 1,
		AccountingID:      &accountingID,
		Status:            "discharged",
	})

	assert.Equal(t, uint64(1), resp.HospitalizationID)
	assert.Equal(t, &accountingID, resp.AccountingID)
	assert.Equal(t, "discharged", resp.Status)
}

func TestToDischargeWithBillingResponse_NilAccountingID(t *testing.T) {
	resp := toDischargeWithBillingResponse(&DischargeWithBillingResult{
		HospitalizationID: 2,
		AccountingID:      nil,
		Status:            "admitted",
	})

	assert.Equal(t, uint64(2), resp.HospitalizationID)
	assert.Nil(t, resp.AccountingID)
	assert.Equal(t, "admitted", resp.Status)
}
