package lstep

import (
	"net/url"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
)

func TestNewLstepDeliveryMonitorSummaryQuery(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)

	query, err := newLstepDeliveryMonitorSummaryQuery(1, url.Values{
		"from":         {"2026-05-01"},
		"to":           {"2026-05-31"},
		"trigger_type": {"after_visit"},
	}, now)
	if err != nil {
		t.Fatalf("newLstepDeliveryMonitorSummaryQuery returned error: %v", err)
	}

	if query.ClinicID != 1 {
		t.Errorf("ClinicID = %d, want 1", query.ClinicID)
	}
	if got := query.From.Format("2006-01-02"); got != "2026-05-01" {
		t.Errorf("From = %q, want 2026-05-01", got)
	}
	if got := query.To.Format("2006-01-02"); got != "2026-06-01" {
		t.Errorf("To = %q, want 2026-06-01", got)
	}
	if query.TriggerType != "after_visit" {
		t.Errorf("TriggerType = %q, want after_visit", query.TriggerType)
	}
}

func TestNewLstepDeliveryMonitorLogsQuery(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)

	query, err := newLstepDeliveryMonitorLogsQuery(1, url.Values{
		"from":         {"2026-05-01"},
		"to":           {"2026-05-31"},
		"trigger_type": {"after_visit"},
		"status":       {"failed"},
		"page":         {"2"},
		"per_page":     {"50"},
	}, now)
	if err != nil {
		t.Fatalf("newLstepDeliveryMonitorLogsQuery returned error: %v", err)
	}

	if query.Status != "failed" {
		t.Errorf("Status = %q, want failed", query.Status)
	}
	if query.Page != 2 {
		t.Errorf("Page = %d, want 2", query.Page)
	}
	if query.PerPage != 50 {
		t.Errorf("PerPage = %d, want 50", query.PerPage)
	}
}

func TestNewLstepDeliveryMonitorLogsQuery_Defaults(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, config.JST)

	query, err := newLstepDeliveryMonitorLogsQuery(1, url.Values{}, now)
	if err != nil {
		t.Fatalf("newLstepDeliveryMonitorLogsQuery returned error: %v", err)
	}

	if got := query.From.Format("2006-01-02"); got != "2026-05-28" {
		t.Errorf("From = %q, want 2026-05-28", got)
	}
	if got := query.To.Format("2006-01-02"); got != "2026-05-29" {
		t.Errorf("To = %q, want 2026-05-29", got)
	}
	if query.Page != 1 {
		t.Errorf("Page = %d, want 1", query.Page)
	}
	if query.PerPage != 20 {
		t.Errorf("PerPage = %d, want 20", query.PerPage)
	}
}

func TestNewLstepDeliveryMonitorLogsQuery_Invalid(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		values url.Values
	}{
		{"from invalid", url.Values{"from": {"2026/05/01"}}},
		{"to invalid", url.Values{"to": {"2026/05/31"}}},
		{"page zero", url.Values{"page": {"0"}}},
		{"page non numeric", url.Values{"page": {"abc"}}},
		{"per_page zero", url.Values{"per_page": {"0"}}},
		{"per_page too large", url.Values{"per_page": {"101"}}},
		{"per_page non numeric", url.Values{"per_page": {"abc"}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := newLstepDeliveryMonitorLogsQuery(1, tc.values, now); err == nil {
				t.Fatal("newLstepDeliveryMonitorLogsQuery returned nil error")
			} else if !apperrors.IsInvalidInput(err) {
				t.Fatalf("error = %v, want invalid input", err)
			}
		})
	}
}
