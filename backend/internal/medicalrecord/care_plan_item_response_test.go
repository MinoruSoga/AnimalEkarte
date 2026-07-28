package medicalrecord

import (
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToCarePlanItemResponse(t *testing.T) {
	createdAt := time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 7, 2, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name   string
		item   *model.CarePlanItem
		wantFn func(t *testing.T, resp carePlanItemResponse)
	}{
		{
			name: "full item with all optional pointers and timing set",
			item: &model.CarePlanItem{
				ID:                    1,
				HospitalizationID:     10,
				Type:                  "medicine",
				Name:                  "抗生剤投与",
				Description:           "朝夕2回",
				Timing:                pq.StringArray{"morning", "evening"},
				Status:                "active",
				Notes:                 "経過観察",
				MedicineID:            uint64Ptr(5),
				ProcedureID:           uint64Ptr(6),
				HospitalizationPlanID: uint64Ptr(7),
				UnitPrice:             1500,
				Category:              "medicine",
				SortOrder:             3,
				CreatedAt:             createdAt,
				UpdatedAt:             updatedAt,
			},
			wantFn: func(t *testing.T, resp carePlanItemResponse) {
				assert.Equal(t, "1", resp.ID)
				assert.Equal(t, "10", resp.HospitalizationID)
				assert.Equal(t, "medicine", resp.Type)
				assert.Equal(t, "抗生剤投与", resp.Name)
				assert.Equal(t, "朝夕2回", resp.Description)
				assert.Equal(t, []string{"morning", "evening"}, resp.Timing)
				assert.Equal(t, "active", resp.Status)
				assert.Equal(t, "経過観察", resp.Notes)
				require.NotNil(t, resp.MedicineID)
				assert.Equal(t, "5", *resp.MedicineID)
				require.NotNil(t, resp.ProcedureID)
				assert.Equal(t, "6", *resp.ProcedureID)
				require.NotNil(t, resp.HospitalizationPlanID)
				assert.Equal(t, "7", *resp.HospitalizationPlanID)
				assert.Equal(t, int64(1500), resp.UnitPrice)
				assert.Equal(t, "medicine", resp.Category)
				assert.Equal(t, 3, resp.SortOrder)
			},
		},
		{
			name: "zero-value item with nil optional pointers and nil timing",
			item: &model.CarePlanItem{
				ID:                1,
				HospitalizationID: 10,
			},
			wantFn: func(t *testing.T, resp carePlanItemResponse) {
				assert.Nil(t, resp.MedicineID)
				assert.Nil(t, resp.ProcedureID)
				assert.Nil(t, resp.HospitalizationPlanID)
				assert.NotNil(t, resp.Timing, "nil Timing should be coalesced to an explicit empty slice")
				assert.Empty(t, resp.Timing)
			},
		},
		{
			name: "empty (non-nil) timing slice is preserved as-is",
			item: &model.CarePlanItem{
				ID:                2,
				HospitalizationID: 10,
				Timing:            pq.StringArray{},
			},
			wantFn: func(t *testing.T, resp carePlanItemResponse) {
				assert.NotNil(t, resp.Timing)
				assert.Empty(t, resp.Timing)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := toCarePlanItemResponse(tt.item)
			tt.wantFn(t, resp)
		})
	}
}

func uint64Ptr(v uint64) *uint64 {
	return &v
}
