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
	admittedAt := time.Date(2026, 5, 28, 10, 0, 0, 0, time.UTC)
	dischargedAt := time.Date(2026, 5, 30, 11, 0, 0, 0, time.UTC)
	kind := stayKindHotel
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
	}, HospitalizationStayProjection{
		StayKind: &kind, PlannedEntryAt: &startDate, PlannedExitAt: &endDate,
		Lifecycle: lifecycleDischarged, AdmittedAt: &admittedAt, DischargedAt: &dischargedAt,
	})

	assert.Equal(t, uint64(1), resp.ID)
	assert.Equal(t, uint64(2), resp.ClinicID)
	assert.Equal(t, uint64(3), resp.OwnerID)
	assert.Equal(t, uint64(4), resp.PetID)
	// EMR-499 U1: planned interval, actual admission/discharge and lifecycle are separate facts.
	require.NotNil(t, resp.StayKind)
	assert.Equal(t, stayKindHotel, *resp.StayKind)
	require.NotNil(t, resp.PlannedEntryAt)
	assert.True(t, resp.PlannedEntryAt.Equal(startDate))
	require.NotNil(t, resp.PlannedExitAt)
	assert.True(t, resp.PlannedExitAt.Equal(endDate))
	assert.Equal(t, lifecycleDischarged, resp.LifecycleState)
	require.NotNil(t, resp.AdmittedAt)
	assert.True(t, resp.AdmittedAt.Equal(admittedAt))
	require.NotNil(t, resp.DischargedAt)
	assert.True(t, resp.DischargedAt.Equal(dischargedAt))
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
		CageID:               nil,
		DoctorID:             nil,
		InsuranceCompanyName: nil,
		InsuranceNumber:      nil,
	}, unknownHospitalizationStay())

	assert.Equal(t, uint64(5), resp.ID)
	// A hospitalization with no readable stay: nothing is defaulted, the lifecycle is explicitly unknown.
	assert.Nil(t, resp.StayKind)
	assert.Nil(t, resp.PlannedEntryAt)
	assert.Nil(t, resp.PlannedExitAt)
	assert.Nil(t, resp.AdmittedAt)
	assert.Nil(t, resp.DischargedAt)
	assert.Equal(t, lifecycleUnknown, resp.LifecycleState)
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

// EMR-301/EMR-319 レビュー対応: medicalrecord PetSummaryResponse は
// HospitalizationResponse.pet 経由で birth_date / deceased_at を wire へ含める
// （HospitalizationPatientHeader は /pets/:id を再取得せずこの値を使う）。
func TestToHospitalizationResponse_PetSummaryIncludesBirthDateAndDeceasedAt(t *testing.T) {
	birthDate := time.Date(2019, time.October, 2, 0, 0, 0, 0, time.UTC)
	deceasedAt := time.Date(2026, time.July, 10, 12, 0, 0, 0, time.UTC)

	resp := toHospitalizationResponse(&model.Hospitalization{
		ID: 1,
		Pet: &model.Pet{
			ID:         4,
			Name:       "ポチ",
			Status:     model.PetStatusAlive,
			BirthDate:  &birthDate,
			DeceasedAt: &deceasedAt,
		},
	}, unknownHospitalizationStay())

	require.NotNil(t, resp.Pet)
	require.NotNil(t, resp.Pet.BirthDate)
	assert.Equal(t, "2019-10-02", *resp.Pet.BirthDate)
	assert.NotNil(t, resp.Pet.DeceasedAt, "deceasedAt-only の半死亡データも wire へ伝播する")
}
