package medicalrecord

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// PR #851 review A1: snapshot=true must apply the requested lifecycle_state with the same derived-state
// filter as the list tabs, and total must be the matched count rather than the unfiltered candidate count.
func TestListHospitalizationsSnapshotAppliesDerivedLifecycleState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rows := []model.Hospitalization{{ID: 1, ClinicID: 1}, {ID: 2, ClinicID: 1}, {ID: 3, ClinicID: 1}}
	projections := map[uint64]HospitalizationStayProjection{
		1: {Lifecycle: lifecycleAdmitted},
		2: {Lifecycle: lifecycleDischarged},
		3: unknownHospitalizationStay(),
	}
	cases := []struct {
		query string
		want  []uint64
	}{
		{"snapshot=true", []uint64{1, 2, 3}},
		{"snapshot=true&lifecycle_state=admitted", []uint64{1}},
		{"snapshot=true&lifecycle_state=discharged", []uint64{2}},
		{"snapshot=true&lifecycle_state=unknown", []uint64{3}},
		{"snapshot=true&lifecycle_state=reserved", []uint64{}},
	}
	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			svc := &snapshotHospitalizationService{
				mockHospitalizationService: &mockHospitalizationService{
					stayProjectionsFn: func(context.Context, uint64, []model.Hospitalization) (map[uint64]HospitalizationStayProjection, error) {
						return projections, nil
					},
				},
				snapshotFn: func(context.Context, uint64, listHospitalizationFilters) ([]model.Hospitalization, int64, error) {
					return rows, int64(len(rows)), nil
				},
			}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/hospitalizations?"+tc.query, nil)
			setClinicID(c)

			newHandlerWithHospitalizationSvc(svc).ListHospitalizations(c)

			require.Equal(t, http.StatusOK, rec.Code)
			var body struct {
				Data  []HospitalizationResponse `json:"data"`
				Total int64                     `json:"total"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			got := make([]uint64, 0, len(body.Data))
			for _, row := range body.Data {
				got = append(got, row.ID)
			}
			assert.Equal(t, tc.want, got)
			assert.Equal(t, int64(len(tc.want)), body.Total)
		})
	}
}
