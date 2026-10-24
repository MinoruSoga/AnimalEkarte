package medicalrecord

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// EMR-499 U1: the stay period is no longer a PATCH-able pair. A planned change belongs to extension
// proposals / acceptance proof and the actual period is derived from movement events, so a PATCH that
// carries the retired or stay-owned period fields is refused with a stable 422 (never silently dropped),
// while next_visit keeps its date-only binding.
func TestHospitalizationPeriodPatchContract(t *testing.T) {
	t.Run("period fields are refused with stay_lifecycle_write_unavailable", func(t *testing.T) {
		for name, body := range map[string]string{
			"retired start/end pair": `{"start_date":"2026-08-01T00:00:00+09:00","end_date":"2026-08-08T00:00:00+09:00","next_visit":"2026-08-29"}`,
			"retired start only":     `{"start_date":"2026-08-01T00:00:00+09:00"}`,
			"planned entry/exit":     `{"planned_entry_at":"2026-08-01T09:00:00+09:00","planned_exit_at":"2026-08-08T09:00:00+09:00"}`,
		} {
			t.Run(name, func(t *testing.T) {
				var req updateHospitalizationRequest
				require.NoError(t, json.Unmarshal([]byte(body), &req))
				_, err := req.toServiceInput()
				requireStayLifecycleWriteRefused(t, err)
			})
		}
	})

	t.Run("next_visit alone still binds as a local calendar date", func(t *testing.T) {
		var req updateHospitalizationRequest
		require.NoError(t, json.Unmarshal([]byte(`{"next_visit":"2026-08-29"}`), &req))

		input, err := req.toServiceInput()

		require.NoError(t, err)
		require.NotNil(t, input.NextVisit)
		assert.Equal(t, "2026-08-29", input.NextVisit.Format(time.DateOnly))
		assert.Equal(t,
			time.Date(2026, 8, 29, 0, 0, 0, 0, time.Local).Format(time.RFC3339),
			input.NextVisit.Format(time.RFC3339))
	})
}
