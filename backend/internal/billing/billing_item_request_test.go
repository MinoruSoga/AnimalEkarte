package billing

import (
	"net/url"
	"strings"
	"testing"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestNewUnbilledItemsQuery(t *testing.T) {
	tests := []struct {
		name      string
		values    url.Values
		wantPetID string
	}{
		{
			name:      "extracts pet_id from query values",
			values:    url.Values{"pet_id": []string{"42"}},
			wantPetID: "42",
		},
		{
			name:      "returns empty PetID when pet_id absent",
			values:    url.Values{},
			wantPetID: "",
		},
		{
			name:      "nil values yields empty PetID",
			values:    nil,
			wantPetID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := newUnbilledItemsQuery(tt.values)
			if q.PetID != tt.wantPetID {
				t.Fatalf("PetID = %q, want %q", q.PetID, tt.wantPetID)
			}
		})
	}
}

func TestUnbilledItemsQuery_ToPetID(t *testing.T) {
	petID, err := (unbilledItemsQuery{PetID: "10"}).toPetID()
	if err != nil {
		t.Fatalf("toPetID returned error: %v", err)
	}
	if petID != 10 {
		t.Fatalf("PetID = %d, want 10", petID)
	}
}

func TestUnbilledItemsQuery_ToPetID_InvalidInput(t *testing.T) {
	tests := []unbilledItemsQuery{
		{},
		{PetID: "0"},
		{PetID: "abc"},
	}

	for _, tt := range tests {
		_, err := tt.toPetID()
		if err == nil {
			t.Fatalf("toPetID returned nil error for %#v", tt)
		}
		if !apperrors.IsInvalidInput(err) {
			t.Fatalf("error = %v, want invalid input", err)
		}
	}
}

func TestCreateBillingItemRequest_ToServiceInput(t *testing.T) {
	treatmentID := uint64(100)
	appointmentID := uint64(200)
	trimmingCourseID := uint64(300)
	trimmingOptionID := uint64(400)
	req := createBillingItemRequest{
		BillingID:             10,
		Category:              string(model.ItemCategoryMedicine),
		Name:                  "medicine item",
		UnitPrice:             1200,
		Quantity:              2.5,
		TaxType:               string(model.TaxTypeExcluded),
		TaxRate:               0.1,
		IsInsuranceApplicable: true,
		Source:                string(model.ItemSourceManual),
		TreatmentID:           &treatmentID,
		AppointmentID:         &appointmentID,
		TrimmingCourseID:      &trimmingCourseID,
		TrimmingOptionID:      &trimmingOptionID,
		SortOrder:             3,
	}

	input := req.toServiceInput(1)

	if input.ClinicID != 1 {
		t.Fatalf("ClinicID = %d, want 1", input.ClinicID)
	}
	if input.BillingID != req.BillingID {
		t.Fatalf("BillingID = %d, want %d", input.BillingID, req.BillingID)
	}
	if input.Category != req.Category {
		t.Fatalf("Category = %q, want %q", input.Category, req.Category)
	}
	if input.Name != req.Name {
		t.Fatalf("Name = %q, want %q", input.Name, req.Name)
	}
	if input.UnitPrice != req.UnitPrice {
		t.Fatalf("UnitPrice = %d, want %d", input.UnitPrice, req.UnitPrice)
	}
	if input.Quantity != req.Quantity {
		t.Fatalf("Quantity = %f, want %f", input.Quantity, req.Quantity)
	}
	if input.TaxType != req.TaxType || input.TaxRate != req.TaxRate {
		t.Fatalf("tax = (%q, %f), want (%q, %f)", input.TaxType, input.TaxRate, req.TaxType, req.TaxRate)
	}
	if !input.IsInsuranceApplicable {
		t.Fatal("IsInsuranceApplicable = false, want true")
	}
	if input.Source != req.Source || input.SortOrder != req.SortOrder {
		t.Fatalf("source/order = (%q, %d), want (%q, %d)", input.Source, input.SortOrder, req.Source, req.SortOrder)
	}
	if input.TreatmentID == nil || *input.TreatmentID != treatmentID {
		t.Fatalf("TreatmentID = %v, want %d", input.TreatmentID, treatmentID)
	}
	if input.AppointmentID == nil || *input.AppointmentID != appointmentID {
		t.Fatalf("AppointmentID = %v, want %d", input.AppointmentID, appointmentID)
	}
	if input.TrimmingCourseID == nil || *input.TrimmingCourseID != trimmingCourseID {
		t.Fatalf("TrimmingCourseID = %v, want %d", input.TrimmingCourseID, trimmingCourseID)
	}
	if input.TrimmingOptionID == nil || *input.TrimmingOptionID != trimmingOptionID {
		t.Fatalf("TrimmingOptionID = %v, want %d", input.TrimmingOptionID, trimmingOptionID)
	}
}

func TestUpdateBillingItemRequest_ToServiceInput(t *testing.T) {
	unitPrice := int64(0)
	quantity := 0.0
	taxType := string(model.TaxTypeExempt)
	taxRate := 0.0
	isInsuranceApplicable := false
	req := updateBillingItemRequest{
		UnitPrice:             &unitPrice,
		Quantity:              &quantity,
		TaxType:               &taxType,
		TaxRate:               &taxRate,
		IsInsuranceApplicable: &isInsuranceApplicable,
	}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput returned error: %v", err)
	}

	if input.UnitPrice == nil || *input.UnitPrice != 0 {
		t.Fatalf("UnitPrice = %v, want explicit zero", input.UnitPrice)
	}
	if input.Quantity == nil || *input.Quantity != 0 {
		t.Fatalf("Quantity = %v, want explicit zero", input.Quantity)
	}
	if input.TaxType == nil || *input.TaxType != model.TaxTypeExempt {
		t.Fatalf("TaxType = %v, want %q", input.TaxType, model.TaxTypeExempt)
	}
	if input.TaxRate == nil || *input.TaxRate != 0 {
		t.Fatalf("TaxRate = %v, want explicit zero", input.TaxRate)
	}
	if input.IsInsuranceApplicable == nil || *input.IsInsuranceApplicable {
		t.Fatalf("IsInsuranceApplicable = %v, want explicit false", input.IsInsuranceApplicable)
	}
}

func TestUpdateBillingItemRequest_ToServiceInput_NilTaxType(t *testing.T) {
	req := updateBillingItemRequest{}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput returned error: %v", err)
	}

	if input.TaxType != nil {
		t.Fatalf("TaxType = %v, want nil", input.TaxType)
	}
}

func TestUpdateBillingItemRequest_ToServiceInput_InvalidTaxType(t *testing.T) {
	taxType := "invalid"
	req := updateBillingItemRequest{TaxType: &taxType}

	input, err := req.toServiceInput()
	if err == nil {
		t.Fatal("toServiceInput returned nil error")
	}
	if input != nil {
		t.Fatalf("input = %#v, want nil", input)
	}
}

func TestBillingItemRequest_PostCloseReasonMax(t *testing.T) {
	tests := []struct {
		name    string
		payload map[string]any
		dest    any
		wantErr bool
	}{
		{
			name: "create post_close_reason at 500 is accepted",
			payload: map[string]any{
				"billing_id":        1,
				"name":              "明細",
				"post_close_reason": strings.Repeat("a", 500),
			},
			dest: &createBillingItemRequest{},
		},
		{
			name: "create post_close_reason over 500 is rejected",
			payload: map[string]any{
				"billing_id":        1,
				"name":              "明細",
				"post_close_reason": strings.Repeat("a", 501),
			},
			dest:    &createBillingItemRequest{},
			wantErr: true,
		},
		{
			name:    "update post_close_reason at 500 is accepted",
			payload: map[string]any{"post_close_reason": strings.Repeat("a", 500)},
			dest:    &updateBillingItemRequest{},
		},
		{
			name:    "update post_close_reason over 500 is rejected",
			payload: map[string]any{"post_close_reason": strings.Repeat("a", 501)},
			dest:    &updateBillingItemRequest{},
			wantErr: true,
		},
		{
			name:    "delete post_close_reason at 500 is accepted",
			payload: map[string]any{"post_close_reason": strings.Repeat("a", 500)},
			dest:    &deleteBillingItemRequest{},
		},
		{
			name:    "delete post_close_reason over 500 is rejected",
			payload: map[string]any{"post_close_reason": strings.Repeat("a", 501)},
			dest:    &deleteBillingItemRequest{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bindJSONBody(t, tt.payload, tt.dest)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ShouldBindJSON = nil, want over-max error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ShouldBindJSON = %v, want nil", err)
			}
		})
	}
}

func TestCreateBillingItemRequest_NameMax(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{name: "255 chars accepted", length: 255, wantErr: false},
		{name: "256 chars rejected", length: 256, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bindJSONBody(t, map[string]any{
				"billing_id": 1,
				"name":       strings.Repeat("a", tt.length),
			}, &createBillingItemRequest{})
			if tt.wantErr {
				if err == nil {
					t.Fatal("ShouldBindJSON = nil, want over-max error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ShouldBindJSON = %v, want nil", err)
			}
		})
	}
}
