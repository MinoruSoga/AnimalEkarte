package lstep

import (
	"testing"
	"time"
)

func TestUpdateTriggerPrioritiesRequest_ToServiceInput(t *testing.T) {
	req := updateTriggerPrioritiesRequest{
		Items: []updateTriggerPriorityItemRequest{
			{TriggerType: "after_visit", Priority: 1},
			{TriggerType: "vaccine_reminder", Priority: 2},
		},
	}

	input := req.toServiceInput()

	if len(input.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(input.Items))
	}
	if input.Items[0].TriggerType != "after_visit" || input.Items[0].Priority != 1 {
		t.Errorf("Items[0] = %+v, want after_visit priority 1", input.Items[0])
	}
}

func TestLstepDeliveryMonitorQuery_ToSummaryServiceInput(t *testing.T) {
	from := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 1)
	query := lstepDeliveryMonitorQuery{
		ClinicID:    1,
		From:        from,
		To:          to,
		TriggerType: "after_visit",
	}

	input := query.toSummaryServiceInput()

	if input.ClinicID != query.ClinicID {
		t.Errorf("ClinicID = %d, want %d", input.ClinicID, query.ClinicID)
	}
	if input.TriggerType != query.TriggerType {
		t.Errorf("TriggerType = %q, want %q", input.TriggerType, query.TriggerType)
	}
}

func TestLstepDeliveryMonitorQuery_ToLogsServiceInput(t *testing.T) {
	from := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 1)
	query := lstepDeliveryMonitorQuery{
		ClinicID:    1,
		From:        from,
		To:          to,
		TriggerType: "after_visit",
		Status:      "failed",
		Page:        2,
		PerPage:     50,
	}

	input := query.toLogsServiceInput()

	if input.Status != query.Status {
		t.Errorf("Status = %q, want %q", input.Status, query.Status)
	}
	if input.Page != query.Page {
		t.Errorf("Page = %d, want %d", input.Page, query.Page)
	}
	if input.PerPage != query.PerPage {
		t.Errorf("PerPage = %d, want %d", input.PerPage, query.PerPage)
	}
}
