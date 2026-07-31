package inventory

import (
	"strings"
	"testing"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestCreateInventoryRequest_ToServiceInput(t *testing.T) {
	expiryDate := "2026-12-31"
	lastRestocked := "2026-05-28T10:30:00Z"

	req := createInventoryRequest{
		Name:          "Antibiotic",
		Category:      "medicine",
		Quantity:      12,
		Unit:          "box",
		MinStockLevel: 3,
		Location:      "Shelf A",
		ExpiryDate:    &expiryDate,
		Supplier:      "Supplier Inc.",
		LastRestocked: &lastRestocked,
	}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput returned error: %v", err)
	}

	if input.Name != req.Name {
		t.Fatalf("Name = %q, want %q", input.Name, req.Name)
	}
	if input.Category != req.Category {
		t.Fatalf("Category = %q, want %q", input.Category, req.Category)
	}
	if input.Quantity != req.Quantity {
		t.Fatalf("Quantity = %d, want %d", input.Quantity, req.Quantity)
	}
	if input.Unit != req.Unit {
		t.Fatalf("Unit = %q, want %q", input.Unit, req.Unit)
	}
	if input.MinStockLevel != req.MinStockLevel {
		t.Fatalf("MinStockLevel = %d, want %d", input.MinStockLevel, req.MinStockLevel)
	}
	if input.Location != req.Location {
		t.Fatalf("Location = %q, want %q", input.Location, req.Location)
	}
	if input.ExpiryDate == nil || input.ExpiryDate.Format("2006-01-02") != expiryDate {
		t.Fatalf("ExpiryDate = %v, want %s", input.ExpiryDate, expiryDate)
	}
	if input.Supplier != req.Supplier {
		t.Fatalf("Supplier = %q, want %q", input.Supplier, req.Supplier)
	}
	if input.LastRestocked == nil || input.LastRestocked.Format("2006-01-02T15:04:05Z07:00") != lastRestocked {
		t.Fatalf("LastRestocked = %v, want %s", input.LastRestocked, lastRestocked)
	}
	// SD-4: create request に status フィールドは存在しない（client write 経路 0）
}

func TestInventoryRequest_ToServiceInput_InvalidDate(t *testing.T) {
	invalidDate := "not-a-date"

	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "create/expiry_date",
			call: func() error {
				_, err := (&createInventoryRequest{ExpiryDate: &invalidDate}).toServiceInput()
				return err
			},
		},
		{
			name: "create/last_restocked",
			call: func() error {
				_, err := (&createInventoryRequest{LastRestocked: &invalidDate}).toServiceInput()
				return err
			},
		},
		{
			name: "update/expiry_date",
			call: func() error {
				_, err := (&updateInventoryRequest{ExpiryDate: &invalidDate}).toServiceInput()
				return err
			},
		},
		{
			name: "update/last_restocked",
			call: func() error {
				_, err := (&updateInventoryRequest{LastRestocked: &invalidDate}).toServiceInput()
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			if err == nil {
				t.Fatal("toServiceInput returned nil error")
			}
			if !apperrors.IsInvalidInput(err) {
				t.Fatalf("error = %v, want invalid input", err)
			}
			if strings.Contains(err.Error(), invalidDate) {
				t.Fatalf("error = %v, must not leak raw input %q", err, invalidDate)
			}
		})
	}
}

func TestUpdateInventoryRequest_ToServiceInput(t *testing.T) {
	name := ""
	category := "consumable"
	quantity := 0
	unit := ""
	minStockLevel := 0
	location := ""
	expiryDate := "2026-12-31"
	supplier := ""
	lastRestocked := "2026-05-28"

	req := updateInventoryRequest{
		Name:          &name,
		Category:      &category,
		Quantity:      &quantity,
		Unit:          &unit,
		MinStockLevel: &minStockLevel,
		Location:      &location,
		ExpiryDate:    &expiryDate,
		Supplier:      &supplier,
		LastRestocked: &lastRestocked,
	}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput returned error: %v", err)
	}

	if input.Name != &name {
		t.Fatalf("Name pointer was not preserved")
	}
	if input.Category == nil || string(*input.Category) != category {
		t.Fatalf("Category = %v, want %q", input.Category, category)
	}
	if input.Quantity != &quantity {
		t.Fatalf("Quantity pointer was not preserved")
	}
	if input.Unit != &unit {
		t.Fatalf("Unit pointer was not preserved")
	}
	if input.MinStockLevel != &minStockLevel {
		t.Fatalf("MinStockLevel pointer was not preserved")
	}
	if input.Location != &location {
		t.Fatalf("Location pointer was not preserved")
	}
	if input.ExpiryDate == nil || input.ExpiryDate.Format("2006-01-02") != expiryDate {
		t.Fatalf("ExpiryDate = %v, want %s", input.ExpiryDate, expiryDate)
	}
	if input.Supplier != &supplier {
		t.Fatalf("Supplier pointer was not preserved")
	}
	if input.LastRestocked == nil || input.LastRestocked.Format("2006-01-02") != lastRestocked {
		t.Fatalf("LastRestocked = %v, want %s", input.LastRestocked, lastRestocked)
	}
	// SD-4: update request に status フィールドは存在しない（client write 経路 0）
}

func TestUpdateInventoryRequest_ToServiceInput_NilFields(t *testing.T) {
	input, err := (&updateInventoryRequest{}).toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput returned error: %v", err)
	}

	if input.Name != nil {
		t.Fatalf("Name = %v, want nil", input.Name)
	}
	if input.Category != nil {
		t.Fatalf("Category = %v, want nil", input.Category)
	}
	if input.Quantity != nil {
		t.Fatalf("Quantity = %v, want nil", input.Quantity)
	}
	if input.Unit != nil {
		t.Fatalf("Unit = %v, want nil", input.Unit)
	}
	if input.MinStockLevel != nil {
		t.Fatalf("MinStockLevel = %v, want nil", input.MinStockLevel)
	}
	if input.Location != nil {
		t.Fatalf("Location = %v, want nil", input.Location)
	}
	if input.ExpiryDate != nil {
		t.Fatalf("ExpiryDate = %v, want nil", input.ExpiryDate)
	}
	if input.Supplier != nil {
		t.Fatalf("Supplier = %v, want nil", input.Supplier)
	}
	if input.LastRestocked != nil {
		t.Fatalf("LastRestocked = %v, want nil", input.LastRestocked)
	}
}
