package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCheckupFieldType_IsValid guards the #211 checkup package field-type request
// boundary check: only the 6 defined field types are accepted.
func TestCheckupFieldType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		in   CheckupFieldType
		want bool
	}{
		{"number is valid", CheckupFieldTypeNumber, true},
		{"single_select is valid", CheckupFieldTypeSingleSelect, true},
		{"multi_select is valid", CheckupFieldTypeMultiSelect, true},
		{"boolean is valid", CheckupFieldTypeBoolean, true},
		{"checklist is valid", CheckupFieldTypeChecklist, true},
		{"text is valid", CheckupFieldTypeText, true},
		{"unknown value is invalid", CheckupFieldType("unknown"), false},
		{"empty value is invalid", CheckupFieldType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.in.IsValid())
		})
	}
}
