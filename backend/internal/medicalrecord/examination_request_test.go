package medicalrecord

import (
	"net/url"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestNewListExaminationQuery(t *testing.T) {
	tests := []struct {
		name   string
		values url.Values
		want   listExaminationQuery
	}{
		{
			name: "extracts all query parameters",
			values: url.Values{
				"pet_id":     []string{"10"},
				"owner_id":   []string{"20"},
				"status":     []string{"completed"},
				"start_date": []string{"2026-05-01"},
				"end_date":   []string{"2026-05-31"},
			},
			want: listExaminationQuery{
				PetID:     "10",
				OwnerID:   "20",
				Status:    "completed",
				StartDate: "2026-05-01",
				EndDate:   "2026-05-31",
			},
		},
		{
			name:   "returns zero value for empty query",
			values: url.Values{},
			want:   listExaminationQuery{},
		},
		{
			name:   "returns zero value for nil query",
			values: nil,
			want:   listExaminationQuery{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newListExaminationQuery(tt.values)
			if got != tt.want {
				t.Errorf("newListExaminationQuery() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestListExaminationQuery_ToServiceFilters(t *testing.T) {
	filters, err := (&listExaminationQuery{
		PetID:     "10",
		OwnerID:   "20",
		Status:    "completed",
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
	if filters.Status == nil || *filters.Status != "completed" {
		t.Fatalf("Status = %v, want completed", filters.Status)
	}
	if filters.StartDate == nil || *filters.StartDate != "2026-05-01" {
		t.Fatalf("StartDate = %v, want 2026-05-01", filters.StartDate)
	}
	if filters.EndDate == nil || *filters.EndDate != "2026-05-31" {
		t.Fatalf("EndDate = %v, want 2026-05-31", filters.EndDate)
	}
}

func TestListExaminationQuery_ToServiceFilters_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		query listExaminationQuery
	}{
		{name: "pet_id", query: listExaminationQuery{PetID: "abc"}},
		{name: "owner_id", query: listExaminationQuery{OwnerID: "abc"}},
		{name: "start_date", query: listExaminationQuery{StartDate: "2026/05/01"}},
		{name: "end_date", query: listExaminationQuery{EndDate: "2026/05/31"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters, err := tt.query.toServiceFilters()
			if err == nil {
				t.Fatal("toServiceFilters returned nil error")
			}
			if filters != (listExaminationFilters{}) {
				t.Fatalf("filters = %#v, want zero value", filters)
			}
			if !apperrors.IsInvalidInput(err) {
				t.Fatalf("error = %v, want invalid input", err)
			}
		})
	}
}

func TestCreateExaminationRequest_ToServiceInput(t *testing.T) {
	medicalRecordID := uint64(10)
	petID := uint64(20)
	doctorID := uint64(30)
	date := time.Date(2026, 5, 28, 9, 30, 0, 0, time.UTC)
	req := createExaminationRequest{
		MedicalRecordID: &medicalRecordID,
		PetID:           &petID,
		ExamTypeID:      40,
		DoctorID:        &doctorID,
		Date:            date,
		ResultSummary:   "normal",
		Machine:         "machine-a",
		Status:          "completed",
	}

	input := req.toServiceInput()

	if input.MedicalRecordID == nil || *input.MedicalRecordID != medicalRecordID {
		t.Errorf("MedicalRecordID = %v, want %d", input.MedicalRecordID, medicalRecordID)
	}
	if input.PetID == nil || *input.PetID != petID {
		t.Errorf("PetID = %v, want %d", input.PetID, petID)
	}
	if input.ExamTypeID != req.ExamTypeID {
		t.Errorf("ExamTypeID = %d, want %d", input.ExamTypeID, req.ExamTypeID)
	}
	if input.Date != date {
		t.Errorf("Date = %v, want %v", input.Date, date)
	}
	if string(input.Status) != req.Status {
		t.Errorf("Status = %q, want %q", input.Status, req.Status)
	}
}

func TestUpdateExaminationRequest_ToServiceInput(t *testing.T) {
	petID := uint64(20)
	date := time.Date(2026, 5, 28, 9, 30, 0, 0, time.UTC)
	resultSummary := ""
	machine := "machine-b"
	status := "result_entered"
	req := updateExaminationRequest{
		PetID:         &petID,
		ExamTypeID:    40,
		Date:          &date,
		ResultSummary: &resultSummary,
		Machine:       &machine,
		Status:        &status,
	}

	input := req.toServiceInput()

	if input.PetID == nil || *input.PetID != petID {
		t.Errorf("PetID = %v, want %d", input.PetID, petID)
	}
	if input.ExamTypeID == nil || *input.ExamTypeID != req.ExamTypeID {
		t.Errorf("ExamTypeID = %v, want %d", input.ExamTypeID, req.ExamTypeID)
	}
	if input.ResultSummary == nil || *input.ResultSummary != resultSummary {
		t.Errorf("ResultSummary = %v, want empty string pointer", input.ResultSummary)
	}
	if input.Status == nil || string(*input.Status) != status {
		t.Errorf("Status = %v, want %q", input.Status, status)
	}
}

func TestUpdateExaminationRequest_ToServiceInput_EmptyExamTypeID(t *testing.T) {
	req := updateExaminationRequest{}

	input := req.toServiceInput()

	if input.ExamTypeID != nil {
		t.Errorf("ExamTypeID = %v, want nil", input.ExamTypeID)
	}
	if input.Status != nil {
		t.Errorf("Status = %v, want nil", input.Status)
	}
}

func TestReplaceExamItemsRequest_ToServiceInput(t *testing.T) {
	fieldID := uint64(7)
	req := replaceExamItemsRequest{
		Items: []upsertExamItemRequest{
			{
				ExamTypeFieldID: &fieldID,
				Name:            "WBC",
				InspectionValue: "2.5",
				NormalValue:     "1.2-3.4",
				Unit:            "K/uL",
				ReferenceValue:  "ref",
				SortOrder:       2,
			},
		},
	}

	input := req.toServiceInput()

	if len(input) != 1 {
		t.Fatalf("len(input) = %d, want 1", len(input))
	}
	if input[0].ExamTypeFieldID == nil || *input[0].ExamTypeFieldID != fieldID {
		t.Errorf("ExamTypeFieldID = %v, want %d", input[0].ExamTypeFieldID, fieldID)
	}
	if input[0].Name != "WBC" {
		t.Errorf("Name = %q, want WBC", input[0].Name)
	}
	if input[0].SortOrder != 2 {
		t.Errorf("SortOrder = %d, want 2", input[0].SortOrder)
	}
}

func TestReplaceExamItemsRequest_ToServiceInput_NilItems(t *testing.T) {
	req := replaceExamItemsRequest{}

	input := req.toServiceInput()

	if len(input) != 0 {
		t.Errorf("len(input) = %d, want 0", len(input))
	}
	if input == nil {
		t.Error("input = nil, want empty slice")
	}
}

// BUG-450: include_items は明細同梱の opt-in である。既定（未指定）で明細を返さない
// 後方互換が、この2 case で固定される。
func TestListExaminationQuery_ToServiceFilters_IncludeItems(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "true opts into item delivery", value: "true", want: true},
		{name: "unset keeps the default lean response", value: "", want: false},
		{name: "false keeps the default lean response", value: "false", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters, err := (&listExaminationQuery{IncludeItems: tt.value}).toServiceFilters()
			if err != nil {
				t.Fatalf("toServiceFilters returned error: %v", err)
			}
			if filters.IncludeItems != tt.want {
				t.Errorf("IncludeItems = %t, want %t", filters.IncludeItems, tt.want)
			}
		})
	}
}
