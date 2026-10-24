package medicalrecord

import (
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestNewListHospitalizationQuery(t *testing.T) {
	values := url.Values{
		"pet_id":             []string{"10"},
		"owner_id":           []string{"20"},
		"lifecycle_state":    []string{"admitted"},
		"planned_entry_from": []string{"2026-05-01"},
		"planned_entry_to":   []string{"2026-05-31"},
	}

	q := newListHospitalizationQuery(values)

	if q.PetID != "10" || q.OwnerID != "20" {
		t.Errorf("PetID/OwnerID = %q/%q, want 10/20", q.PetID, q.OwnerID)
	}
	if q.LifecycleState != "admitted" {
		t.Errorf("LifecycleState = %q, want admitted", q.LifecycleState)
	}
	if q.PlannedEntryFrom != "2026-05-01" || q.PlannedEntryTo != "2026-05-31" {
		t.Errorf("planned entry range = %q..%q", q.PlannedEntryFrom, q.PlannedEntryTo)
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
		PetID:            "10",
		OwnerID:          "20",
		LifecycleState:   "admitted",
		PlannedEntryFrom: "2026-05-01",
		PlannedEntryTo:   "2026-05-31",
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
	if filters.LifecycleState == nil || *filters.LifecycleState != "admitted" {
		t.Fatalf("LifecycleState = %v, want admitted", filters.LifecycleState)
	}
	if filters.PlannedEntryFrom == nil || *filters.PlannedEntryFrom != "2026-05-01" {
		t.Fatalf("PlannedEntryFrom = %v, want 2026-05-01", filters.PlannedEntryFrom)
	}
	if filters.PlannedEntryTo == nil || *filters.PlannedEntryTo != "2026-05-31" {
		t.Fatalf("PlannedEntryTo = %v, want 2026-05-31", filters.PlannedEntryTo)
	}
}

func TestListHospitalizationQuery_ToServiceFilters_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		query listHospitalizationQuery
	}{
		{name: "pet_id", query: listHospitalizationQuery{PetID: "abc"}},
		{name: "owner_id", query: listHospitalizationQuery{OwnerID: "abc"}},
		{name: "planned_entry_from", query: listHospitalizationQuery{PlannedEntryFrom: "2026/05/01"}},
		{name: "planned_entry_to", query: listHospitalizationQuery{PlannedEntryTo: "2026/05/31"}},
		{name: "lifecycle_state", query: listHospitalizationQuery{LifecycleState: "released"}},
		// The pre-stay names are refused, not ignored: ignoring would return an unfiltered list.
		{name: "retired status", query: listHospitalizationQuery{RetiredStatus: true}},
		{name: "retired start_date", query: listHospitalizationQuery{RetiredStartDate: true}},
		{name: "retired end_date", query: listHospitalizationQuery{RetiredEndDate: true}},
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

// Retired names are refused by key presence: an empty or repeated value must not pass as "no filter".
func TestListHospitalizationQuery_RefusesRetiredKeyRegardlessOfValue(t *testing.T) {
	for _, key := range []string{"status", "start_date", "end_date"} {
		for _, raw := range []string{key + "=", key, key + "=&" + key + "=2026-10-24"} {
			t.Run(raw, func(t *testing.T) {
				values, err := url.ParseQuery(raw)
				if err != nil {
					t.Fatalf("ParseQuery(%q) error = %v", raw, err)
				}
				q := newListHospitalizationQuery(values)
				filters, err := q.toServiceFilters()
				if !apperrors.IsInvalidInput(err) {
					t.Fatalf("error = %v, want invalid input", err)
				}
				if filters != (listHospitalizationFilters{}) {
					t.Fatalf("filters = %#v, want zero value", filters)
				}
			})
		}
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
		StayKind:             "hotel",
		PlannedEntryAt:       &startDate,
		PlannedExitAt:        &endDate,
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
	if input.StayKind != stayKindHotel {
		t.Fatalf("StayKind = %s, want %s", input.StayKind, stayKindHotel)
	}
	if !input.PlannedEntryAt.Equal(startDate) || !input.PlannedExitAt.Equal(endDate) {
		t.Fatalf("planned interval = %v..%v, want %v..%v", input.PlannedEntryAt, input.PlannedExitAt, startDate, endDate)
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
	cageID := uint64(10)
	input, err := (&createHospitalizationRequest{
		OwnerID: 1,
		PetID:   2,
		CageID:  &cageID,
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

// The pre-stay lifecycle names are refused with the stable 422 code, never silently dropped.
func TestCreateHospitalizationRequest_ToServiceInput_RefusesRetiredLifecycleFields(t *testing.T) {
	v, at := "admitted", time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	tests := map[string]createHospitalizationRequest{
		"hospitalization_type": {OwnerID: 1, PetID: 2, RetiredType: &v},
		"start_date":           {OwnerID: 1, PetID: 2, RetiredStartDate: &at},
		"end_date":             {OwnerID: 1, PetID: 2, RetiredEndDate: &at},
		"status":               {OwnerID: 1, PetID: 2, RetiredStatus: &v},
	}
	for name, req := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := req.toServiceInput()
			requireStayLifecycleWriteRefused(t, err)
		})
	}
}

func requireStayLifecycleWriteRefused(t *testing.T, err error) {
	t.Helper()
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != codeStayLifecycleWriteUnavailable || !errors.Is(err, apperrors.ErrUnprocessable) {
		t.Fatalf("error = %v, want 422 %s", err, codeStayLifecycleWriteUnavailable)
	}
}

// EMR-316: BUG-037 の必須化を撤回。cage_id 省略は「ケージ未割当」、0 のみ不正値として拒否する。
func TestCreateHospitalizationRequest_ToServiceInput_CageIDOptional(t *testing.T) {
	base := createHospitalizationRequest{
		OwnerID: 1,
		PetID:   2,
	}
	t.Run("nil cage_id is accepted as cageless", func(t *testing.T) {
		input, err := base.toServiceInput()
		if err != nil {
			t.Fatalf("toServiceInput() error = %v, want nil for cageless create", err)
		}
		if input.CageID != nil {
			t.Fatalf("CageID = %v, want nil", *input.CageID)
		}
	})
	t.Run("zero cage_id is rejected", func(t *testing.T) {
		zero := uint64(0)
		req := base
		req.CageID = &zero
		_, err := req.toServiceInput()
		if err == nil {
			t.Fatal("toServiceInput() error = nil, want cage_id invalid")
		}
	})
	t.Run("positive cage_id is preserved", func(t *testing.T) {
		id := uint64(10)
		req := base
		req.CageID = &id
		input, err := req.toServiceInput()
		if err != nil {
			t.Fatalf("toServiceInput() error = %v", err)
		}
		if input.CageID != &id {
			t.Fatal("CageID pointer was not preserved")
		}
	})
}

func TestUpdateHospitalizationRequest_ToServiceInput_ClearCageID(t *testing.T) {
	t.Run("clear_cage_id is forwarded", func(t *testing.T) {
		expected := uint64(10)
		input, err := (&updateHospitalizationRequest{ClearCageID: true, ExpectedCageID: &expected}).toServiceInput()
		if err != nil {
			t.Fatalf("toServiceInput() error = %v", err)
		}
		if !input.ClearCageID || input.CageID != nil {
			t.Fatalf("ClearCageID=%v CageID=%v, want true/nil", input.ClearCageID, input.CageID)
		}
		if input.ExpectedCageID == nil || *input.ExpectedCageID != expected {
			t.Fatalf("ExpectedCageID=%v, want %d", input.ExpectedCageID, expected)
		}
	})
	t.Run("clear_cage_id without expected_cage_id is rejected", func(t *testing.T) {
		_, err := (&updateHospitalizationRequest{ClearCageID: true}).toServiceInput()
		if err == nil {
			t.Fatal("toServiceInput() error = nil, want missing expected_cage_id error")
		}
	})
	t.Run("clear_cage_id with zero expected_cage_id is rejected", func(t *testing.T) {
		zero := uint64(0)
		_, err := (&updateHospitalizationRequest{ClearCageID: true, ExpectedCageID: &zero}).toServiceInput()
		if err == nil {
			t.Fatal("toServiceInput() error = nil, want zero expected_cage_id error")
		}
	})
	t.Run("clear_cage_id with cage_id is rejected", func(t *testing.T) {
		id := uint64(10)
		_, err := (&updateHospitalizationRequest{ClearCageID: true, CageID: &id, ExpectedCageID: &id}).toServiceInput()
		if err == nil {
			t.Fatal("toServiceInput() error = nil, want conflict error")
		}
	})
	t.Run("clear_cage_id with expected_cage_is_null is forwarded", func(t *testing.T) {
		input, err := (&updateHospitalizationRequest{ClearCageID: true, ExpectedCageIsNull: true}).toServiceInput()
		if err != nil {
			t.Fatalf("toServiceInput() error = %v", err)
		}
		if !input.ClearCageID || !input.ExpectedCageIsNull {
			t.Fatalf("ClearCageID=%v ExpectedCageIsNull=%v, want true/true", input.ClearCageID, input.ExpectedCageIsNull)
		}
	})
}

func TestUpdateHospitalizationRequest_ToServiceInput_CageSetGuards(t *testing.T) {
	id := uint64(10)
	zero := uint64(0)
	tests := []struct {
		name    string
		req     updateHospitalizationRequest
		wantErr bool
	}{
		{name: "cage_id with expected_cage_id is accepted", req: updateHospitalizationRequest{CageID: &id, ExpectedCageID: &id}},
		{name: "cage_id with expected_cage_is_null is accepted", req: updateHospitalizationRequest{CageID: &id, ExpectedCageIsNull: true}},
		{name: "cage_id without expected is rejected", req: updateHospitalizationRequest{CageID: &id}, wantErr: true},
		{name: "cage_id with zero expected_cage_id is rejected", req: updateHospitalizationRequest{CageID: &id, ExpectedCageID: &zero}, wantErr: true},
		{name: "cage_id with both expected guards is rejected", req: updateHospitalizationRequest{CageID: &id, ExpectedCageID: &id, ExpectedCageIsNull: true}, wantErr: true},
		{name: "zero cage_id is rejected", req: updateHospitalizationRequest{CageID: &zero, ExpectedCageID: &id}, wantErr: true},
		{name: "expected_cage_id without cage operation is rejected", req: updateHospitalizationRequest{ExpectedCageID: &id}, wantErr: true},
		{name: "expected_cage_is_null without cage operation is rejected", req: updateHospitalizationRequest{ExpectedCageIsNull: true}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := tt.req.toServiceInput()
			if tt.wantErr {
				if err == nil {
					t.Fatal("toServiceInput() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("toServiceInput() error = %v", err)
			}
			if input.CageID != &id {
				t.Fatal("CageID pointer was not preserved")
			}
		})
	}
}

func TestUpdateHospitalizationRequest_ToServiceInput(t *testing.T) {
	ownerID := uint64(1)
	petID := uint64(2)
	cageID := uint64(10)
	expectedCageID := uint64(9)
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
		CageID:               &cageID,
		ExpectedCageID:       &expectedCageID,
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
	if input.ExpectedCageID != &expectedCageID {
		t.Fatalf("ExpectedCageID pointer was not preserved")
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

	if input.OwnerID != nil || input.Memo != nil || input.CageID != nil {
		t.Fatalf("expected nil optional fields, got %#v", input)
	}
}

func TestUpdateHospitalizationRequest_ToServiceInput_RefusesStayOwnedFields(t *testing.T) {
	v, at := "admitted", time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	tests := map[string]updateHospitalizationRequest{
		"stay_kind":            {StayKind: &v},
		"planned_entry_at":     {PlannedEntryAt: &at},
		"planned_exit_at":      {PlannedExitAt: &at},
		"hospitalization_type": {HospitalizationType: &v},
		"start_date":           {StartDate: &at},
		"end_date":             {EndDate: &at},
		"status":               {Status: &v},
	}
	for name, req := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := req.toServiceInput()
			requireStayLifecycleWriteRefused(t, err)
		})
	}
}

func TestValidatePlannedStayInterval(t *testing.T) {
	entry := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
	later := entry.Add(time.Hour)
	if err := validatePlannedStayInterval(&entry, &later); err != nil {
		t.Fatalf("expected ok for exit after entry, got %v", err)
	}
	if err := validatePlannedStayInterval(&entry, nil); err != nil {
		t.Fatalf("a missing side is not invented or rejected, got %v", err)
	}
	if err := validatePlannedStayInterval(nil, nil); err != nil {
		t.Fatalf("no interval is valid, got %v", err)
	}
	if err := validatePlannedStayInterval(&entry, &entry); err == nil {
		t.Fatal("expected error when exit == entry (stay_cases CHECK exit > entry)")
	}
	earlier := entry.Add(-time.Hour)
	if err := validatePlannedStayInterval(&entry, &earlier); err == nil {
		t.Fatal("expected error when exit < entry")
	}
}

func TestCreateHospitalizationRequest_ToServiceInput_RejectsInvertedPlannedInterval(t *testing.T) {
	entry := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	exit := entry.Add(-24 * time.Hour)
	_, err := (&createHospitalizationRequest{OwnerID: 1, PetID: 2, StayKind: "hotel", PlannedEntryAt: &entry, PlannedExitAt: &exit}).toServiceInput()
	if err == nil {
		t.Fatal("expected inverted planned interval to fail")
	}
}
