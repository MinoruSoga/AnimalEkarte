package medicalrecord

import (
	"net/url"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestNewListHospitalizationQuery(t *testing.T) {
	values := url.Values{
		"pet_id":     []string{"10"},
		"owner_id":   []string{"20"},
		"status":     []string{"admitted"},
		"start_date": []string{"2026-05-01"},
		"end_date":   []string{"2026-05-31"},
	}

	q := newListHospitalizationQuery(values)

	if q.PetID != "10" {
		t.Errorf("PetID = %q, want %q", q.PetID, "10")
	}
	if q.OwnerID != "20" {
		t.Errorf("OwnerID = %q, want %q", q.OwnerID, "20")
	}
	if q.Status != "admitted" {
		t.Errorf("Status = %q, want %q", q.Status, "admitted")
	}
	if q.StartDate != "2026-05-01" {
		t.Errorf("StartDate = %q, want %q", q.StartDate, "2026-05-01")
	}
	if q.EndDate != "2026-05-31" {
		t.Errorf("EndDate = %q, want %q", q.EndDate, "2026-05-31")
	}
}

func TestNewListHospitalizationQuery_Empty(t *testing.T) {
	q := newListHospitalizationQuery(url.Values{})

	if q != (listHospitalizationQuery{}) {
		t.Fatalf("q = %#v, want zero value", q)
	}
}

func TestListHospitalizationQuery_ToServiceFilters(t *testing.T) {
	filters, err := (&listHospitalizationQuery{
		PetID:     "10",
		OwnerID:   "20",
		Status:    "admitted",
		StartDate: "2026-05-01",
		EndDate:   "2026-05-31",
	}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters returned error: %v", err)
	}

	if filters.PetID == nil || *filters.PetID != 10 {
		t.Fatalf("PetID = %v, want 10", filters.PetID)
	}
	if filters.OwnerID == nil || *filters.OwnerID != 20 {
		t.Fatalf("OwnerID = %v, want 20", filters.OwnerID)
	}
	if filters.Status == nil || *filters.Status != "admitted" {
		t.Fatalf("Status = %v, want admitted", filters.Status)
	}
	if filters.StartDate == nil || *filters.StartDate != "2026-05-01" {
		t.Fatalf("StartDate = %v, want 2026-05-01", filters.StartDate)
	}
	if filters.EndDate == nil || *filters.EndDate != "2026-05-31" {
		t.Fatalf("EndDate = %v, want 2026-05-31", filters.EndDate)
	}
}

func TestListHospitalizationQuery_ToServiceFilters_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		query listHospitalizationQuery
	}{
		{name: "pet_id", query: listHospitalizationQuery{PetID: "abc"}},
		{name: "owner_id", query: listHospitalizationQuery{OwnerID: "abc"}},
		{name: "start_date", query: listHospitalizationQuery{StartDate: "2026/05/01"}},
		{name: "end_date", query: listHospitalizationQuery{EndDate: "2026/05/31"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters, err := tt.query.toServiceFilters()
			if err == nil {
				t.Fatal("toServiceFilters returned nil error")
			}
			if filters != (listHospitalizationFilters{}) {
				t.Fatalf("filters = %#v, want zero value", filters)
			}
			if !apperrors.IsInvalidInput(err) {
				t.Fatalf("error = %v, want invalid input", err)
			}
		})
	}
}

func TestDischargeWithBillingRequest_ToServiceInput(t *testing.T) {
	dischargeDate := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)

	input := (&dischargeWithBillingRequest{
		DischargeDate:    dischargeDate,
		CreateAccounting: true,
	}).toServiceInput(42)

	if input.DischargeDate != dischargeDate {
		t.Fatalf("DischargeDate = %v, want %v", input.DischargeDate, dischargeDate)
	}
	if !input.CreateAccounting {
		t.Fatalf("CreateAccounting = false, want true")
	}
	if input.ActorID == nil || *input.ActorID != 42 {
		t.Fatalf("ActorID = %v, want 42", input.ActorID)
	}
}

func TestCreateHospitalizationRequest_ToServiceInput(t *testing.T) {
	startDate := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC)
	cageID := uint64(10)
	doctorID := uint64(20)
	insuranceCompanyName := "Pet Insurance"
	insuranceNumber := "INS-001"

	input, err := (&createHospitalizationRequest{
		OwnerID:              1,
		PetID:                2,
		HospitalizationType:  string(model.HospitalizationTypeHotel),
		StartDate:            startDate,
		EndDate:              endDate,
		Status:               string(model.HospitalizationStatusReserved),
		CageID:               &cageID,
		DoctorID:             &doctorID,
		Memo:                 "memo",
		OwnerRequest:         "quiet room",
		StaffNotes:           "notes",
		IsInsurance:          true,
		InsuranceCompanyName: &insuranceCompanyName,
		InsuranceNumber:      &insuranceNumber,
	}).toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if input.OwnerID != 1 || input.PetID != 2 {
		t.Fatalf("owner/pet IDs = %d/%d, want 1/2", input.OwnerID, input.PetID)
	}
	if input.HospitalizationType != model.HospitalizationTypeHotel {
		t.Fatalf("HospitalizationType = %s, want %s", input.HospitalizationType, model.HospitalizationTypeHotel)
	}
	if input.Status != model.HospitalizationStatusReserved {
		t.Fatalf("Status = %s, want %s", input.Status, model.HospitalizationStatusReserved)
	}
	if input.CageID != &cageID {
		t.Fatalf("CageID pointer was not preserved")
	}
	if input.DoctorID != &doctorID {
		t.Fatalf("DoctorID pointer was not preserved")
	}
	if input.InsuranceCompanyName != &insuranceCompanyName {
		t.Fatalf("InsuranceCompanyName pointer was not preserved")
	}
	if input.InsuranceNumber != &insuranceNumber {
		t.Fatalf("InsuranceNumber pointer was not preserved")
	}
	if len(input.TreatmentPlans) != 0 {
		t.Fatalf("TreatmentPlans = %d, want 0 when omitted", len(input.TreatmentPlans))
	}
}

func TestCreateHospitalizationRequest_ToServiceInput_NestedTreatmentPlans(t *testing.T) {
	startDate := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC)
	cageID := uint64(10)
	input, err := (&createHospitalizationRequest{
		OwnerID:             1,
		PetID:               2,
		HospitalizationType: string(model.HospitalizationTypeInpatient),
		StartDate:           startDate,
		EndDate:             endDate,
		CageID:              &cageID,
		TreatmentPlans: []createTreatmentPlanRequest{
			{TreatmentContent: "adm", UnitPrice: 990, Quantity: 1, SortOrder: 0},
			{TreatmentContent: "monitor", UnitPrice: 500, Quantity: 2, DiscountRate: 10, SortOrder: 1},
		},
	}).toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}
	if len(input.TreatmentPlans) != 2 {
		t.Fatalf("TreatmentPlans = %d, want 2", len(input.TreatmentPlans))
	}
	if input.TreatmentPlans[0].TreatmentContent != "adm" || input.TreatmentPlans[0].UnitPrice != 990 {
		t.Fatalf("plan[0] = %+v", input.TreatmentPlans[0])
	}
	if input.TreatmentPlans[1].DiscountRate != 10 || input.TreatmentPlans[1].Quantity != 2 {
		t.Fatalf("plan[1] = %+v", input.TreatmentPlans[1])
	}
}

func TestCreateHospitalizationRequest_ToServiceInput_EmptyStatus(t *testing.T) {
	cageID := uint64(10)
	input, err := (&createHospitalizationRequest{
		OwnerID:             1,
		PetID:               2,
		HospitalizationType: string(model.HospitalizationTypeInpatient),
		Status:              "",
		CageID:              &cageID,
	}).toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}
	if input.Status != "" {
		t.Fatalf("Status = %q, want empty (zero value)", input.Status)
	}
}

func TestCreateHospitalizationRequest_ToServiceInput_RequiresCageID(t *testing.T) {
	startDate := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC)
	base := createHospitalizationRequest{
		OwnerID:             1,
		PetID:               2,
		HospitalizationType: string(model.HospitalizationTypeInpatient),
		StartDate:           startDate,
		EndDate:             endDate,
	}
	t.Run("nil cage_id", func(t *testing.T) {
		_, err := base.toServiceInput()
		if err == nil {
			t.Fatal("toServiceInput() error = nil, want cage_id required")
		}
	})
	t.Run("zero cage_id", func(t *testing.T) {
		zero := uint64(0)
		req := base
		req.CageID = &zero
		_, err := req.toServiceInput()
		if err == nil {
			t.Fatal("toServiceInput() error = nil, want cage_id required")
		}
	})
}

func TestCreateHospitalizationRequest_ToServiceInput_InvalidHospitalizationType(t *testing.T) {
	_, err := (&createHospitalizationRequest{
		OwnerID:             1,
		PetID:               2,
		HospitalizationType: "invalid",
	}).toServiceInput()
	if err == nil {
		t.Fatal("toServiceInput() error = nil, want error")
	}
}

func TestCreateHospitalizationRequest_ToServiceInput_InvalidStatus(t *testing.T) {
	_, err := (&createHospitalizationRequest{
		OwnerID:             1,
		PetID:               2,
		HospitalizationType: string(model.HospitalizationTypeInpatient),
		Status:              "invalid",
	}).toServiceInput()
	if err == nil {
		t.Fatal("toServiceInput() error = nil, want error")
	}
}

func TestUpdateHospitalizationRequest_ToServiceInput(t *testing.T) {
	ownerID := uint64(1)
	petID := uint64(2)
	hospitalizationType := string(model.HospitalizationTypeInpatient)
	startDate := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC)
	status := string(model.HospitalizationStatusAdmitted)
	cageID := uint64(10)
	doctorID := uint64(20)
	memo := "memo"
	ownerRequest := "request"
	staffNotes := "notes"
	isInsurance := false
	insuranceCompanyName := ""
	insuranceNumber := ""

	input, err := (&updateHospitalizationRequest{
		OwnerID:              &ownerID,
		PetID:                &petID,
		HospitalizationType:  &hospitalizationType,
		StartDate:            &startDate,
		EndDate:              &endDate,
		Status:               &status,
		CageID:               &cageID,
		DoctorID:             &doctorID,
		Memo:                 &memo,
		OwnerRequest:         &ownerRequest,
		StaffNotes:           &staffNotes,
		IsInsurance:          &isInsurance,
		InsuranceCompanyName: &insuranceCompanyName,
		InsuranceNumber:      &insuranceNumber,
	}).toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if input.OwnerID != &ownerID {
		t.Fatalf("OwnerID pointer was not preserved")
	}
	if input.HospitalizationType == nil || *input.HospitalizationType != model.HospitalizationTypeInpatient {
		t.Fatalf("HospitalizationType = %v, want %s", input.HospitalizationType, model.HospitalizationTypeInpatient)
	}
	if input.Status == nil || *input.Status != model.HospitalizationStatusAdmitted {
		t.Fatalf("Status = %v, want %s", input.Status, model.HospitalizationStatusAdmitted)
	}
	if input.IsInsurance != &isInsurance {
		t.Fatalf("IsInsurance pointer was not preserved")
	}
	if input.InsuranceCompanyName != &insuranceCompanyName {
		t.Fatalf("InsuranceCompanyName pointer was not preserved")
	}
	if input.InsuranceNumber != &insuranceNumber {
		t.Fatalf("InsuranceNumber pointer was not preserved")
	}
}

func TestUpdateHospitalizationRequest_ToServiceInput_NilFields(t *testing.T) {
	input, err := (&updateHospitalizationRequest{}).toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if input.OwnerID != nil || input.HospitalizationType != nil || input.Status != nil {
		t.Fatalf("expected nil optional fields, got %#v", input)
	}
}

func TestUpdateHospitalizationRequest_ToServiceInput_InvalidHospitalizationType(t *testing.T) {
	invalid := "invalid"
	_, err := (&updateHospitalizationRequest{HospitalizationType: &invalid}).toServiceInput()
	if err == nil {
		t.Fatal("toServiceInput() error = nil, want error")
	}
}

func TestUpdateHospitalizationRequest_ToServiceInput_InvalidStatus(t *testing.T) {
	invalid := "invalid"
	_, err := (&updateHospitalizationRequest{Status: &invalid}).toServiceInput()
	if err == nil {
		t.Fatal("toServiceInput() error = nil, want error")
	}
}

func TestValidateHospitalizationDateRange_MRB06(t *testing.T) {
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	endOK := time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC)
	endBad := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	if err := validateHospitalizationDateRange(start, endOK); err != nil {
		t.Fatalf("expected ok for end>=start, got %v", err)
	}
	if err := validateHospitalizationDateRange(start, start); err != nil {
		t.Fatalf("expected ok for equal dates, got %v", err)
	}
	if err := validateHospitalizationDateRange(start, endBad); err == nil {
		t.Fatal("expected error when end < start")
	}
}

func TestCreateHospitalizationRequest_ToServiceInput_RejectsInvertedDates(t *testing.T) {
	req := createHospitalizationRequest{
		OwnerID:             1,
		PetID:               2,
		HospitalizationType: string(model.HospitalizationTypeInpatient),
		StartDate:           time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		EndDate:             time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	_, err := req.toServiceInput()
	if err == nil {
		t.Fatal("expected inverted date range to fail")
	}
}
