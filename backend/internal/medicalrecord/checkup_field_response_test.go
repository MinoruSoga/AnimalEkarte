package medicalrecord

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToCheckupTypeFieldResponse(t *testing.T) {
	minValue := 1.0
	maxValue := 10.0

	tests := []struct {
		name        string
		field       *model.CheckupTypeField
		wantOptions string
	}{
		{
			name: "nil options fall back to empty array",
			field: &model.CheckupTypeField{
				ID:            1,
				CheckupTypeID: 2,
				Name:          "体重",
				FieldType:     model.CheckupFieldTypeNumber,
				Unit:          "kg",
				MinValue:      &minValue,
				MaxValue:      &maxValue,
				Options:       nil,
				IsProvisional: true,
				SortOrder:     3,
			},
			wantOptions: "[]",
		},
		{
			name: "empty options fall back to empty array",
			field: &model.CheckupTypeField{
				ID:        1,
				Options:   datatypes.JSON(""),
				FieldType: model.CheckupFieldTypeText,
			},
			wantOptions: "[]",
		},
		{
			name: "non-empty options pass through",
			field: &model.CheckupTypeField{
				ID:        1,
				Options:   datatypes.JSON(`["a","b"]`),
				FieldType: model.CheckupFieldTypeSingleSelect,
			},
			wantOptions: `["a","b"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toCheckupTypeFieldResponse(tt.field)

			assert.Equal(t, tt.field.ID, got.ID)
			assert.Equal(t, tt.field.CheckupTypeID, got.CheckupTypeID)
			assert.Equal(t, tt.field.Name, got.Name)
			assert.Equal(t, string(tt.field.FieldType), got.FieldType)
			assert.Equal(t, tt.field.Unit, got.Unit)
			assert.Equal(t, tt.field.MinValue, got.MinValue)
			assert.Equal(t, tt.field.MaxValue, got.MaxValue)
			assert.Equal(t, tt.field.IsProvisional, got.IsProvisional)
			assert.Equal(t, tt.field.SortOrder, got.SortOrder)
			assert.JSONEq(t, tt.wantOptions, string(got.Options))
		})
	}
}

func TestToCheckupFieldResultResponse(t *testing.T) {
	fieldID := uint64(10)
	valueNumber := 12.5
	valueBool := true

	tests := []struct {
		name         string
		result       *model.CheckupFieldResult
		wantValueLen int
	}{
		{
			name: "nil value list falls back to empty slice",
			result: &model.CheckupFieldResult{
				ID:                 1,
				CheckupID:          2,
				CheckupTypeFieldID: &fieldID,
				FieldName:          "体重",
				FieldType:          model.CheckupFieldTypeNumber,
				Unit:               "kg",
				ValueNumber:        &valueNumber,
				ValueText:          "text",
				ValueBool:          &valueBool,
				ValueList:          nil,
				IsAbnormal:         true,
				Status:             model.ExaminationResultStatusHigh,
				SortOrder:          5,
			},
			wantValueLen: 0,
		},
		{
			name: "non-empty value list passes through",
			result: &model.CheckupFieldResult{
				ID:        1,
				CheckupID: 2,
				FieldName: "既往歴",
				FieldType: model.CheckupFieldTypeChecklist,
				ValueList: []string{"a", "b"},
				Status:    model.ExaminationResultStatusNormal,
			},
			wantValueLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toCheckupFieldResultResponse(tt.result)

			assert.Equal(t, tt.result.ID, got.ID)
			assert.Equal(t, tt.result.CheckupID, got.CheckupID)
			assert.Equal(t, tt.result.CheckupTypeFieldID, got.CheckupTypeFieldID)
			assert.Equal(t, tt.result.FieldName, got.FieldName)
			assert.Equal(t, string(tt.result.FieldType), got.FieldType)
			assert.Equal(t, tt.result.Unit, got.Unit)
			assert.Equal(t, tt.result.ValueNumber, got.ValueNumber)
			assert.Equal(t, tt.result.ValueText, got.ValueText)
			assert.Equal(t, tt.result.ValueBool, got.ValueBool)
			assert.Equal(t, tt.result.IsAbnormal, got.IsAbnormal)
			assert.Equal(t, string(tt.result.Status), got.Status)
			assert.Equal(t, tt.result.SortOrder, got.SortOrder)
			require.Len(t, got.ValueList, tt.wantValueLen)
			assert.NotNil(t, got.ValueList)
		})
	}
}

func TestToPetCheckupResultResponse(t *testing.T) {
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name             string
		result           *model.CheckupFieldResult
		wantDate         string
		wantCheckupType  uint64
		wantCheckupTypeN string
	}{
		{
			name:             "nil checkup leaves date and type fields empty",
			result:           &model.CheckupFieldResult{ID: 1, FieldName: "体重"},
			wantDate:         "",
			wantCheckupType:  0,
			wantCheckupTypeN: "",
		},
		{
			name: "checkup without checkup type sets date and id only",
			result: &model.CheckupFieldResult{
				ID: 1,
				Checkup: &model.Checkup{
					Date:          date,
					CheckupTypeID: 7,
				},
			},
			wantDate:         "2026-06-01",
			wantCheckupType:  7,
			wantCheckupTypeN: "",
		},
		{
			name: "checkup with checkup type sets name",
			result: &model.CheckupFieldResult{
				ID: 1,
				Checkup: &model.Checkup{
					Date:          date,
					CheckupTypeID: 7,
					CheckupType:   &model.CheckupType{ID: 7, Name: "健診パッケージA"},
				},
			},
			wantDate:         "2026-06-01",
			wantCheckupType:  7,
			wantCheckupTypeN: "健診パッケージA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toPetCheckupResultResponse(tt.result)

			assert.Equal(t, tt.wantDate, got.Date)
			assert.Equal(t, tt.wantCheckupType, got.CheckupTypeID)
			assert.Equal(t, tt.wantCheckupTypeN, got.CheckupTypeName)
			assert.Equal(t, tt.result.ID, got.ID)
			assert.Equal(t, tt.result.FieldName, got.FieldName)
		})
	}
}
