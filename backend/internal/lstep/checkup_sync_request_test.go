package lstep

import (
	"net/url"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin/binding"
)

func TestCheckupSyncPreviewQuery_ToServiceInput(t *testing.T) {
	input, err := (&checkupSyncPreviewQuery{
		CheckupType:         "annual",
		Species:             "dog",
		CPMStage:            string(CPMStageCore),
		LastVisitBefore:     "2026-05-01",
		LastVisitAfter:      "2026-01-01",
		MinAgeYears:         "2",
		MaxAgeYears:         "10",
		HasChronicCondition: "false",
		MinTotalAmount:      "10000",
		MinAnnualVisitCount: "3",
		LastCheckupBefore:   "2026-04-01",
		LastCheckupAfter:    "2026-02-01",
	}).toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if input.CheckupType != "annual" || input.Species != "dog" {
		t.Fatalf("basic fields = %q/%q", input.CheckupType, input.Species)
	}
	if input.MinAgeYears == nil || *input.MinAgeYears != 2 {
		t.Fatalf("MinAgeYears = %v, want 2", input.MinAgeYears)
	}
	if input.HasChronicCondition == nil || *input.HasChronicCondition {
		t.Fatalf("HasChronicCondition = %v, want false", input.HasChronicCondition)
	}
	if input.MinTotalAmount == nil || *input.MinTotalAmount != 10000 {
		t.Fatalf("MinTotalAmount = %v, want 10000", input.MinTotalAmount)
	}
	if input.LastVisitBefore == nil || input.LastCheckupAfter == nil {
		t.Fatalf("date filters were not parsed")
	}
}

func TestCheckupSyncPreviewQuery_ToServiceInput_InvalidBool(t *testing.T) {
	_, err := (&checkupSyncPreviewQuery{HasChronicCondition: "yes"}).toServiceInput()
	if err == nil {
		t.Fatalf("toServiceInput() error = nil, want error")
	}
}

func TestCheckupSyncPreviewQuery_ToServiceInput_InvalidCPMStage(t *testing.T) {
	_, err := (&checkupSyncPreviewQuery{CPMStage: "invalid"}).toServiceInput()
	if err == nil {
		t.Fatalf("toServiceInput() error = nil, want error")
	}
}

func TestCheckupSyncRequest_ToServiceInput(t *testing.T) {
	input, err := (&checkupSyncRequest{
		CheckupType: "annual",
		OwnerIDs:    []string{"1", "2"},
		TagName:     "annual_checkup",
	}).toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if input.CheckupType != "annual" {
		t.Fatalf("CheckupType = %q, want annual", input.CheckupType)
	}
	if len(input.OwnerIDs) != 2 || input.OwnerIDs[1] != 2 {
		t.Fatalf("OwnerIDs = %#v, want [1 2]", input.OwnerIDs)
	}
	if input.TagName != "annual_checkup" {
		t.Fatalf("TagName = %q, want annual_checkup", input.TagName)
	}
}

func TestCheckupSyncRequest_ToServiceInput_InvalidOwnerID(t *testing.T) {
	_, err := (&checkupSyncRequest{
		CheckupType: "annual",
		OwnerIDs:    []string{"x"},
		TagName:     "annual_checkup",
	}).toServiceInput()
	if err == nil {
		t.Fatalf("toServiceInput() error = nil, want error")
	}
}

func TestCheckupSyncRequest_ToServiceInput_InvalidTagName(t *testing.T) {
	_, err := (&checkupSyncRequest{
		CheckupType: "annual",
		OwnerIDs:    []string{"1"},
		TagName:     "invalid tag!",
	}).toServiceInput()
	if err == nil {
		t.Fatalf("toServiceInput() error = nil, want error")
	}
}

func TestCheckupSyncRequest_OwnerIDsBoundary(t *testing.T) {
	ownerIDs := func(count int) []string {
		ids := make([]string, count)
		for i := range ids {
			ids[i] = strconv.Itoa(i + 1)
		}
		return ids
	}

	tests := []struct {
		name      string
		ownerIDs  []string
		wantError bool
	}{
		{
			name:     "accepts 100 owners",
			ownerIDs: ownerIDs(100),
		},
		{
			name:      "rejects 101 owners",
			ownerIDs:  ownerIDs(101),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := checkupSyncRequest{
				CheckupType: "annual",
				OwnerIDs:    tt.ownerIDs,
				TagName:     "annual_checkup",
			}

			err := binding.Validator.ValidateStruct(req)
			if tt.wantError && err == nil {
				t.Fatal("ValidateStruct() error = nil, want owner_ids max validation error")
			}
			if !tt.wantError && err != nil {
				t.Fatalf("ValidateStruct() error = %v, want nil", err)
			}
		})
	}
}

func TestCheckupSyncPreviewQuery_ToServiceInput_InvalidFields(t *testing.T) {
	tests := []struct {
		name  string
		query checkupSyncPreviewQuery
	}{
		{name: "invalid last_visit_before", query: checkupSyncPreviewQuery{LastVisitBefore: "not-a-date"}},
		{name: "invalid last_visit_after", query: checkupSyncPreviewQuery{LastVisitAfter: "not-a-date"}},
		{name: "invalid min_age_years", query: checkupSyncPreviewQuery{MinAgeYears: "abc"}},
		{name: "negative min_age_years", query: checkupSyncPreviewQuery{MinAgeYears: "-1"}},
		{name: "invalid max_age_years", query: checkupSyncPreviewQuery{MaxAgeYears: "abc"}},
		{name: "negative max_age_years", query: checkupSyncPreviewQuery{MaxAgeYears: "-1"}},
		{name: "invalid min_total_amount", query: checkupSyncPreviewQuery{MinTotalAmount: "abc"}},
		{name: "negative min_total_amount", query: checkupSyncPreviewQuery{MinTotalAmount: "-1"}},
		{name: "invalid min_annual_visit_count", query: checkupSyncPreviewQuery{MinAnnualVisitCount: "abc"}},
		{name: "negative min_annual_visit_count", query: checkupSyncPreviewQuery{MinAnnualVisitCount: "-1"}},
		{name: "invalid last_checkup_before", query: checkupSyncPreviewQuery{LastCheckupBefore: "not-a-date"}},
		{name: "invalid last_checkup_after", query: checkupSyncPreviewQuery{LastCheckupAfter: "not-a-date"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.query
			if _, err := q.toServiceInput(); err == nil {
				t.Fatalf("toServiceInput() error = nil, want error")
			}
		})
	}
}

func TestCheckupSyncPreviewQuery_ToServiceInput_HasChronicConditionTrue(t *testing.T) {
	input, err := (&checkupSyncPreviewQuery{HasChronicCondition: "true"}).toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}
	if input.HasChronicCondition == nil || !*input.HasChronicCondition {
		t.Fatalf("HasChronicCondition = %v, want true", input.HasChronicCondition)
	}
}

func TestCheckupSyncPreviewQuery_ToServiceInput_EmptyValues(t *testing.T) {
	input, err := (&checkupSyncPreviewQuery{}).toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}
	if input.CPMStage != "" {
		t.Fatalf("CPMStage = %q, want empty", input.CPMStage)
	}
	if input.HasChronicCondition != nil {
		t.Fatalf("HasChronicCondition = %v, want nil", input.HasChronicCondition)
	}
	if input.MinAgeYears != nil || input.MaxAgeYears != nil {
		t.Fatalf("age filters = %v/%v, want nil", input.MinAgeYears, input.MaxAgeYears)
	}
	if input.LastVisitBefore != nil || input.LastVisitAfter != nil {
		t.Fatalf("last visit filters = %v/%v, want nil", input.LastVisitBefore, input.LastVisitAfter)
	}
	if input.LastCheckupBefore != nil || input.LastCheckupAfter != nil {
		t.Fatalf("last checkup filters = %v/%v, want nil", input.LastCheckupBefore, input.LastCheckupAfter)
	}
}

func TestNewCheckupSyncPreviewQuery(t *testing.T) {
	values := url.Values{
		"checkup_type":           {"annual"},
		"species":                {"dog"},
		"cpm_stage":              {"cpm_core"},
		"last_visit_before":      {"2026-05-01"},
		"last_visit_after":       {"2026-01-01"},
		"min_age_years":          {"2"},
		"max_age_years":          {"10"},
		"has_chronic_condition":  {"true"},
		"min_total_amount":       {"10000"},
		"min_annual_visit_count": {"3"},
		"last_checkup_before":    {"2026-04-01"},
		"last_checkup_after":     {"2026-02-01"},
	}

	q := newCheckupSyncPreviewQuery(values)

	want := checkupSyncPreviewQuery{
		CheckupType:         "annual",
		Species:             "dog",
		CPMStage:            "cpm_core",
		LastVisitBefore:     "2026-05-01",
		LastVisitAfter:      "2026-01-01",
		MinAgeYears:         "2",
		MaxAgeYears:         "10",
		HasChronicCondition: "true",
		MinTotalAmount:      "10000",
		MinAnnualVisitCount: "3",
		LastCheckupBefore:   "2026-04-01",
		LastCheckupAfter:    "2026-02-01",
	}
	if q != want {
		t.Fatalf("newCheckupSyncPreviewQuery() = %+v, want %+v", q, want)
	}
}

func TestNewCheckupSyncPreviewQuery_Empty(t *testing.T) {
	q := newCheckupSyncPreviewQuery(url.Values{})
	if q != (checkupSyncPreviewQuery{}) {
		t.Fatalf("newCheckupSyncPreviewQuery() = %+v, want zero value", q)
	}
}
