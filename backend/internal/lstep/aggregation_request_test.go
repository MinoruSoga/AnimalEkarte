package lstep

import (
	"net/url"
	"testing"
)

func TestOwnerAggregationQuery_ToServiceInput(t *testing.T) {
	minTotalAmount := int64(100)
	year := 2026
	from := "2026-01-01"
	maxVisitCount := int64(3)
	query := ownerAggregationQuery{
		Sort:            "total_amount",
		MinTotalAmount:  &minTotalAmount,
		LineLinked:      true,
		Page:            2,
		PerPage:         50,
		Year:            &year,
		From:            &from,
		AmountBasis:     "paid_amount",
		IncludeZero:     true,
		Search:          "owner",
		PeriodPreset:    "last_12_months",
		MaxVisitCount:   &maxVisitCount,
		LastVisitBucket: "over_3m",
		IncludeNoVisit:  true,
		Order:           "asc",
		Metric:          "visit_count",
		CPMStage:        "core",
	}

	input := query.toServiceInput()

	if input.MinTotalAmount == nil || *input.MinTotalAmount != minTotalAmount {
		t.Errorf("MinTotalAmount = %v, want %d", input.MinTotalAmount, minTotalAmount)
	}
	if input.Year == nil || *input.Year != year {
		t.Errorf("Year = %v, want %d", input.Year, year)
	}
	if input.From == nil || *input.From != from {
		t.Errorf("From = %v, want %q", input.From, from)
	}
	if input.MaxVisitCount == nil || *input.MaxVisitCount != maxVisitCount {
		t.Errorf("MaxVisitCount = %v, want %d", input.MaxVisitCount, maxVisitCount)
	}
	if input.CPMStage != query.CPMStage {
		t.Errorf("CPMStage = %q, want %q", input.CPMStage, query.CPMStage)
	}
}

func TestNewOwnerAggregationQuery(t *testing.T) {
	values := url.Values{
		"page":              {"2"},
		"per_page":          {"100"},
		"min_total_amount":  {"1000"},
		"max_total_amount":  {"99999"},
		"min_visit_count":   {"1"},
		"max_visit_count":   {"20"},
		"line_linked":       {"true"},
		"sort":              {"annual_amount"},
		"order":             {"asc"},
		"year":              {"2026"},
		"from":              {"2026-01-01"},
		"to":                {"2026-12-31"},
		"amount_basis":      {"net_paid_amount"},
		"include_zero":      {"true"},
		"include_no_visit":  {"true"},
		"search":            {"山田"},
		"period_preset":     {"last_6_months"},
		"last_visit_bucket": {"over_6m"},
		"metric":            {"last_visit"},
		"cpm_stage":         {"spot"},
	}

	query, err := newOwnerAggregationQuery(values)
	if err != nil {
		t.Fatalf("newOwnerAggregationQuery returned error: %v", err)
	}

	if query.Page != 2 {
		t.Errorf("Page = %d, want 2", query.Page)
	}
	if query.PerPage != 100 {
		t.Errorf("PerPage = %d, want 100", query.PerPage)
	}
	if query.MinTotalAmount == nil || *query.MinTotalAmount != 1000 {
		t.Errorf("MinTotalAmount = %v, want 1000", query.MinTotalAmount)
	}
	if query.MaxTotalAmount == nil || *query.MaxTotalAmount != 99999 {
		t.Errorf("MaxTotalAmount = %v, want 99999", query.MaxTotalAmount)
	}
	if query.MinVisitCount == nil || *query.MinVisitCount != 1 {
		t.Errorf("MinVisitCount = %v, want 1", query.MinVisitCount)
	}
	if query.MaxVisitCount == nil || *query.MaxVisitCount != 20 {
		t.Errorf("MaxVisitCount = %v, want 20", query.MaxVisitCount)
	}
	if !query.LineLinked {
		t.Error("LineLinked = false, want true")
	}
	if query.Sort != "annual_amount" {
		t.Errorf("Sort = %q, want annual_amount", query.Sort)
	}
	if query.Order != "asc" {
		t.Errorf("Order = %q, want asc", query.Order)
	}
	if query.Year == nil || *query.Year != 2026 {
		t.Errorf("Year = %v, want 2026", query.Year)
	}
	if query.From == nil || *query.From != "2026-01-01" {
		t.Errorf("From = %v, want 2026-01-01", query.From)
	}
	if query.To == nil || *query.To != "2026-12-31" {
		t.Errorf("To = %v, want 2026-12-31", query.To)
	}
	if query.AmountBasis != "net_paid_amount" {
		t.Errorf("AmountBasis = %q, want net_paid_amount", query.AmountBasis)
	}
	if !query.IncludeZero {
		t.Error("IncludeZero = false, want true")
	}
	if !query.IncludeNoVisit {
		t.Error("IncludeNoVisit = false, want true")
	}
	if query.Search != "山田" {
		t.Errorf("Search = %q, want 山田", query.Search)
	}
	if query.PeriodPreset != "last_6_months" {
		t.Errorf("PeriodPreset = %q, want last_6_months", query.PeriodPreset)
	}
	if query.LastVisitBucket != "over_6m" {
		t.Errorf("LastVisitBucket = %q, want over_6m", query.LastVisitBucket)
	}
	if query.Metric != "last_visit" {
		t.Errorf("Metric = %q, want last_visit", query.Metric)
	}
	if query.CPMStage != "spot" {
		t.Errorf("CPMStage = %q, want spot", query.CPMStage)
	}
}

func TestNewOwnerAggregationQuery_Defaults(t *testing.T) {
	query, err := newOwnerAggregationQuery(url.Values{})
	if err != nil {
		t.Fatalf("newOwnerAggregationQuery returned error: %v", err)
	}

	if query.Page != 1 {
		t.Errorf("Page = %d, want 1", query.Page)
	}
	if query.PerPage != 50 {
		t.Errorf("PerPage = %d, want 50", query.PerPage)
	}
	if query.Sort != "total_amount" {
		t.Errorf("Sort = %q, want total_amount", query.Sort)
	}
	if query.Order != "desc" {
		t.Errorf("Order = %q, want desc", query.Order)
	}
	if query.AmountBasis != "gross_total_amount" {
		t.Errorf("AmountBasis = %q, want gross_total_amount", query.AmountBasis)
	}
	if query.Metric != "annual_sales" {
		t.Errorf("Metric = %q, want annual_sales", query.Metric)
	}
	if query.MinTotalAmount != nil || query.MaxTotalAmount != nil || query.Year != nil {
		t.Errorf("optional filters should be nil: min=%v max=%v year=%v", query.MinTotalAmount, query.MaxTotalAmount, query.Year)
	}
}

func TestNewOwnerAggregationQuery_Aliases(t *testing.T) {
	t.Run("canonical amount names take precedence", func(t *testing.T) {
		query, err := newOwnerAggregationQuery(url.Values{
			"min_amount":       {"100"},
			"min_total_amount": {"200"},
			"min_total_fee":    {"300"},
			"max_amount":       {"900"},
			"max_total_amount": {"800"},
			"max_total_fee":    {"700"},
		})
		if err != nil {
			t.Fatalf("newOwnerAggregationQuery returned error: %v", err)
		}
		if query.MinTotalAmount == nil || *query.MinTotalAmount != 100 {
			t.Errorf("MinTotalAmount = %v, want 100", query.MinTotalAmount)
		}
		if query.MaxTotalAmount == nil || *query.MaxTotalAmount != 900 {
			t.Errorf("MaxTotalAmount = %v, want 900", query.MaxTotalAmount)
		}
	})

	t.Run("FE aliases map when canonical names are absent", func(t *testing.T) {
		query, err := newOwnerAggregationQuery(url.Values{
			"min_total_fee": {"500"},
			"max_total_fee": {"10000"},
			"has_line":      {"true"},
			"sort":          {"total_fee"},
		})
		if err != nil {
			t.Fatalf("newOwnerAggregationQuery returned error: %v", err)
		}
		if query.MinTotalAmount == nil || *query.MinTotalAmount != 500 {
			t.Errorf("MinTotalAmount = %v, want 500", query.MinTotalAmount)
		}
		if query.MaxTotalAmount == nil || *query.MaxTotalAmount != 10000 {
			t.Errorf("MaxTotalAmount = %v, want 10000", query.MaxTotalAmount)
		}
		if !query.LineLinked {
			t.Error("LineLinked = false, want true")
		}
		if query.Sort != "total_amount" {
			t.Errorf("Sort = %q, want total_amount", query.Sort)
		}
	})
}

func TestNewOwnerAggregationQuery_Invalid(t *testing.T) {
	testAggregationRequestInvalidCases(t)
}

// TestAggregationRequest_InvalidBoundaryValues is the unit-verify entrypoint for G2A-05
// (`go test -run AggregationRequest`).
func TestAggregationRequest_InvalidBoundaryValues(t *testing.T) {
	testAggregationRequestInvalidCases(t)
}

func testAggregationRequestInvalidCases(t *testing.T) {
	t.Helper()
	cases := []struct {
		name   string
		values url.Values
	}{
		{"page=0", url.Values{"page": {"0"}}},
		{"page=abc", url.Values{"page": {"abc"}}},
		{"per_page=0", url.Values{"per_page": {"0"}}},
		{"per_page=201", url.Values{"per_page": {"201"}}},
		{"min_amount=-1", url.Values{"min_amount": {"-1"}}},
		{"max_amount=abc", url.Values{"max_amount": {"abc"}}},
		{"min_visit_count=-1", url.Values{"min_visit_count": {"-1"}}},
		{"max_visit_count=abc", url.Values{"max_visit_count": {"abc"}}},
		{"year=1899", url.Values{"year": {"1899"}}},
		{"year=2101", url.Values{"year": {"2101"}}},
		{"year=abc", url.Values{"year": {"abc"}}},
		{"from=bad", url.Values{"from": {"2026/01/01"}}},
		{"to=bad", url.Values{"to": {"not-a-date"}}},
		{"amount_basis=bad", url.Values{"amount_basis": {"unknown"}}},
		{"period_preset=bad", url.Values{"period_preset": {"last_year"}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := newOwnerAggregationQuery(tc.values); err == nil {
				t.Fatal("newOwnerAggregationQuery returned nil error")
			}
		})
	}
}
