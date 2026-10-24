package medicalrecord

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestListMedicalRecordQuery_ToServiceFilters(t *testing.T) {
	filters, err := (&listMedicalRecordQuery{
		PetID:     "10",
		OwnerID:   "20",
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
	if filters.StartDate == nil || *filters.StartDate != "2026-05-01" {
		t.Fatalf("StartDate = %v, want 2026-05-01", filters.StartDate)
	}
	if filters.EndDate == nil || *filters.EndDate != "2026-05-31" {
		t.Fatalf("EndDate = %v, want 2026-05-31", filters.EndDate)
	}
}

// TestListMedicalRecordQuery_ToServiceFilters_ColumnTextFilters は EMR-245 の
// 表示列フィルタ（owner_name / pet_name / chief_complaint）が独立した
// フィルタ値としてそのまま service 層へ渡ることを検証する。
func TestListMedicalRecordQuery_ToServiceFilters_ColumnTextFilters(t *testing.T) {
	filters, err := (&listMedicalRecordQuery{
		OwnerName:      "山田",
		PetName:        "ポチ",
		ChiefComplaint: "嘔吐",
		Search:         "横断",
	}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters returned error: %v", err)
	}

	if filters.OwnerName != "山田" {
		t.Fatalf("OwnerName = %q, want 山田", filters.OwnerName)
	}
	if filters.PetName != "ポチ" {
		t.Fatalf("PetName = %q, want ポチ", filters.PetName)
	}
	if filters.ChiefComplaint != "嘔吐" {
		t.Fatalf("ChiefComplaint = %q, want 嘔吐", filters.ChiefComplaint)
	}
	if filters.Search != "横断" {
		t.Fatalf("Search = %q, want 横断", filters.Search)
	}
}

func TestListMedicalRecordQuery_ToServiceFilters_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		query listMedicalRecordQuery
	}{
		{name: "pet_id", query: listMedicalRecordQuery{PetID: "abc"}},
		{name: "owner_id", query: listMedicalRecordQuery{OwnerID: "abc"}},
		{name: "start_date", query: listMedicalRecordQuery{StartDate: "2026/05/01"}},
		{name: "end_date", query: listMedicalRecordQuery{EndDate: "2026/05/31"}},
		{name: "medicine_id", query: listMedicalRecordQuery{MedicineID: "abc"}},
		{name: "procedure_id", query: listMedicalRecordQuery{ProcedureID: "abc"}},
		{name: "consultation_id", query: listMedicalRecordQuery{ConsultationID: "abc"}},
		{name: "inventory_id", query: listMedicalRecordQuery{InventoryID: "abc"}},
		{name: "owner_name too long", query: listMedicalRecordQuery{OwnerName: strings.Repeat("あ", 256)}},
		{name: "pet_name too long", query: listMedicalRecordQuery{PetName: strings.Repeat("あ", 256)}},
		{name: "chief_complaint too long", query: listMedicalRecordQuery{ChiefComplaint: strings.Repeat("あ", 256)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters, err := tt.query.toServiceFilters()
			if err == nil {
				t.Fatal("toServiceFilters returned nil error")
			}
			if filters != (listMedicalRecordFilters{}) {
				t.Fatalf("filters = %#v, want zero value", filters)
			}
			if !apperrors.IsInvalidInput(err) {
				t.Fatalf("error = %v, want invalid input", err)
			}
		})
	}
}

// EMR-495: list の status フィルタ（旧 InvalidStatusDoesNotLeakEnumError）は status 列の廃止とともに廃止した。
// 未知の query は無視されるため、status を渡してもフィルタは空のままであることだけを確認する。
func TestListMedicalRecordQuery_ToServiceFilters_HasNoStatusFilter(t *testing.T) {
	filters, err := (&listMedicalRecordQuery{PetID: "10"}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters returned error: %v", err)
	}
	if filters.PetID == nil || *filters.PetID != 10 {
		t.Fatalf("PetID = %v, want 10", filters.PetID)
	}
}

func TestCreateMedicalRecordRequest_ToServiceInput(t *testing.T) {
	ownerID := "10"
	petID := "20"
	doctorID := "30"
	appointmentID := "40"
	visitDate := "2026-05-28"
	reason := "checkup"
	req := createMedicalRecordRequest{
		RecordNo:             "MR-1",
		VisitDate:            &visitDate,
		VisitType:            string(model.VisitTypeFirst),
		OwnerID:              &ownerID,
		PetID:                &petID,
		DoctorID:             &doctorID,
		AppointmentID:        &appointmentID,
		RecommendationReason: &reason,
	}

	input, err := req.toServiceInput(9)
	if err != nil {
		t.Fatalf("toServiceInput returned error: %v", err)
	}
	if got := input.Date.Format("2006-01-02"); got != visitDate {
		t.Fatalf("Date = %q, want %q", got, visitDate)
	}
	if input.OwnerID == nil || *input.OwnerID != 10 {
		t.Fatalf("OwnerID = %v, want 10", input.OwnerID)
	}
	if input.PetID == nil || *input.PetID != 20 {
		t.Fatalf("PetID = %v, want 20", input.PetID)
	}
	if input.DoctorID == nil || *input.DoctorID != 30 {
		t.Fatalf("DoctorID = %v, want 30", input.DoctorID)
	}
	if input.AppointmentID == nil || *input.AppointmentID != 40 {
		t.Fatalf("AppointmentID = %v, want 40", input.AppointmentID)
	}
	if input.VisitType != model.VisitTypeFirst {
		t.Fatalf("VisitType = %v, want %q", input.VisitType, model.VisitTypeFirst)
	}
	if input.RecommendationReason == nil || *input.RecommendationReason != reason {
		t.Fatalf("RecommendationReason = %v, want %q", input.RecommendationReason, reason)
	}
	if input.EnteredBy == nil || *input.EnteredBy != 9 {
		t.Fatalf("EnteredBy = %v, want 9", input.EnteredBy)
	}
}

func TestCreateMedicalRecordRequest_ToServiceInput_InvalidOwnerID(t *testing.T) {
	ownerID := "abc"
	petID := "20"
	req := createMedicalRecordRequest{
		OwnerID: &ownerID,
		PetID:   &petID,
	}

	input, err := req.toServiceInput(9)
	if err == nil {
		t.Fatal("toServiceInput returned nil error")
	}
	if input.EnteredBy != nil {
		t.Fatalf("input = %#v, want zero value", input)
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
}

func TestCreateMedicalRecordRequest_ToSubRecordsInput(t *testing.T) {
	complaint := "cough"
	plan := "observe"
	slot := int16(2)
	typeID := uint64(11)
	snapshot := "snapshot"
	req := createMedicalRecordRequest{
		ChiefComplaint: &complaint,
		Plan:           &plan,
		Diagnoses: []clinicalPlanDiagnosisRequest{
			{DiagnosisSlot: &slot, DiagnosisTypeID: &typeID, DiagnosisNameSnapshot: &snapshot},
		},
	}

	input, err := req.toSubRecordsInput()
	if err != nil {
		t.Fatalf("toSubRecordsInput returned error: %v", err)
	}

	if input.ChiefComplaint == nil || *input.ChiefComplaint != complaint {
		t.Fatalf("ChiefComplaint = %v, want %q", input.ChiefComplaint, complaint)
	}
	if input.Plan == nil || *input.Plan != plan {
		t.Fatalf("Plan = %v, want %q", input.Plan, plan)
	}
	if len(input.Diagnoses) != 1 || input.Diagnoses[0].Slot != 2 ||
		input.Diagnoses[0].DiagnosisTypeID == nil || *input.Diagnoses[0].DiagnosisTypeID != typeID ||
		input.Diagnoses[0].DiagnosisNameSnapshot == nil || *input.Diagnoses[0].DiagnosisNameSnapshot != snapshot {
		t.Fatalf("Diagnoses = %#v, want the single slot-2 row as sent", input.Diagnoses)
	}
}

func TestCreateMedicalRecordRequest_ToSubRecordsInput_RowWithoutSlotIsInvalid(t *testing.T) {
	snapshot := "snapshot"
	req := createMedicalRecordRequest{Diagnoses: []clinicalPlanDiagnosisRequest{{DiagnosisNameSnapshot: &snapshot}}}

	if _, err := req.toSubRecordsInput(); err == nil || !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
}

func TestUpdateMedicalRecordRequest_ToServiceInput(t *testing.T) {
	date := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	version := 3
	req := updateMedicalRecordRequest{
		Date:    &date,
		Version: &version,
	}

	input, err := req.toServiceInput(9)
	if err != nil {
		t.Fatalf("toServiceInput returned error: %v", err)
	}

	if input.Date == nil || !input.Date.Equal(date) {
		t.Fatalf("Date = %v, want %v", input.Date, date)
	}
	if input.Version == nil || *input.Version != version {
		t.Fatalf("Version = %v, want %d", input.Version, version)
	}
	if input.ActorID == nil || *input.ActorID != 9 {
		t.Fatalf("ActorID = %v, want 9", input.ActorID)
	}
}

// SPEC-02: 更新では status キーが存在した時点で（draft / finalized / 空 / その他）固定エラー。未指定のみ受理。
func TestUpdateMedicalRecordRequest_ToServiceInput_StatusIsAlwaysRetired(t *testing.T) {
	for _, status := range []string{"finalized", "draft", "", "archived"} {
		s := status
		_, err := (updateMedicalRecordRequest{Status: &s}).toServiceInput(9)
		if !errors.Is(err, apperrors.ErrUnprocessable) {
			t.Fatalf("status %q: error = %v, want 422 class", status, err)
		}
		var appErr *apperrors.AppError
		if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeMedicalRecordStatusRetired {
			t.Fatalf("status %q: error = %v, want code %q", status, err, apperrors.CodeMedicalRecordStatusRetired)
		}
	}
	if _, err := (updateMedicalRecordRequest{}).toServiceInput(9); err != nil {
		t.Fatalf("absent status must be accepted, got %v", err)
	}
}

func TestUpdateMedicalRecordRequest_ToServiceInput_InvalidEnumDoesNotLeakEnumError(t *testing.T) {
	invalidVisitType := "walkin"
	emptyVisitType := ""
	mixedVisitType := "First"

	tests := []struct {
		name    string
		req     updateMedicalRecordRequest
		wantMsg string
	}{
		{
			name:    "unknown visit_type",
			req:     updateMedicalRecordRequest{VisitType: &invalidVisitType},
			wantMsg: "来院区分の値が不正です",
		},
		{
			name:    "empty visit_type",
			req:     updateMedicalRecordRequest{VisitType: &emptyVisitType},
			wantMsg: "来院区分の値が不正です",
		},
		{
			name:    "mixed case visit_type",
			req:     updateMedicalRecordRequest{VisitType: &mixedVisitType},
			wantMsg: "来院区分の値が不正です",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := tt.req.toServiceInput(9)
			if input.ActorID != nil {
				t.Fatalf("input = %#v, want zero value", input)
			}
			assertFixedInvalidInputMessage(t, err, tt.wantMsg, "invalid value")
		})
	}
}

// TestResolveMedicalRecordSort は B-1 follow-up の列ソート server 化における
// sort/order query の許可リスト検証を網羅する。
func TestResolveMedicalRecordSort(t *testing.T) {
	tests := []struct {
		name      string
		sort      string
		order     string
		wantSort  string
		wantOrder string
	}{
		{name: "allowed key date + asc", sort: "date", order: "asc", wantSort: "date", wantOrder: "asc"},
		{name: "allowed key owner_name + desc", sort: "owner_name", order: "desc", wantSort: "owner_name", wantOrder: "desc"},
		{name: "allowed key pet_name, no order defaults to desc", sort: "pet_name", order: "", wantSort: "pet_name", wantOrder: "desc"},
		{name: "retired key status falls back to empty (EMR-495: status column removed)", sort: "status", order: "invalid", wantSort: "", wantOrder: ""},
		{name: "unknown sort key falls back to empty (default order)", sort: "unknown_column", order: "asc", wantSort: "", wantOrder: ""},
		{name: "empty sort falls back to empty (default order)", sort: "", order: "asc", wantSort: "", wantOrder: ""},
		{name: "SQL-injection-shaped sort key is rejected", sort: "date; DROP TABLE medical_records;--", order: "asc", wantSort: "", wantOrder: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSort, gotOrder := resolveMedicalRecordSort(tt.sort, tt.order)
			if gotSort != tt.wantSort {
				t.Fatalf("sort = %q, want %q", gotSort, tt.wantSort)
			}
			if gotOrder != tt.wantOrder {
				t.Fatalf("order = %q, want %q", gotOrder, tt.wantOrder)
			}
		})
	}
}

// TestListMedicalRecordQuery_ToServiceFilters_Sort は toServiceFilters が sort/order を
// listMedicalRecordFilters へ正しく伝播することを検証する。
func TestListMedicalRecordQuery_ToServiceFilters_Sort(t *testing.T) {
	filters, err := (&listMedicalRecordQuery{Sort: "owner_name", Order: "asc"}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters returned error: %v", err)
	}
	if filters.Sort != "owner_name" {
		t.Fatalf("Sort = %q, want owner_name", filters.Sort)
	}
	if filters.Order != "asc" {
		t.Fatalf("Order = %q, want asc", filters.Order)
	}
}

func TestPatchMedicalRecordRecommendationReasonRequest_ToServiceInput(t *testing.T) {
	req := patchMedicalRecordRecommendationReasonRequest{Reason: "exam"}

	input := req.toServiceInput()

	if input.Reason != req.Reason {
		t.Fatalf("Reason = %q, want %q", input.Reason, req.Reason)
	}
}

func assertFixedInvalidInputMessage(t *testing.T, err error, wantMessage string, forbiddenSubstrings ...string) {
	t.Helper()
	if err == nil {
		t.Fatal("error = nil, want invalid input")
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
	got := err.Error()
	if !strings.Contains(got, wantMessage) {
		t.Fatalf("error = %q, want to contain %q", got, wantMessage)
	}
	for _, leak := range forbiddenSubstrings {
		if strings.Contains(got, leak) {
			t.Fatalf("error leaked %q: %q", leak, got)
		}
	}
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error = %T (%v), want AppError", err, err)
	}
	if appErr.Message != wantMessage {
		t.Fatalf("AppError.Message = %q, want %q", appErr.Message, wantMessage)
	}
}
