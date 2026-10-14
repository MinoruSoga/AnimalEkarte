package medicalrecord

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestHospitalizationPeriodPatchContract(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantErr   bool
		wantStart string
		wantEnd   string
		wantNext  string
	}{
		{
			name:      "FE paired JST-midnight period with date-only next_visit binds to service input",
			body:      `{"start_date":"2026-08-01T00:00:00+09:00","end_date":"2026-08-08T00:00:00+09:00","next_visit":"2026-08-29"}`,
			wantStart: "2026-08-01T00:00:00+09:00",
			wantEnd:   "2026-08-08T00:00:00+09:00",
			wantNext:  "2026-08-29",
		},
		{
			name:    "inverted paired period is rejected as invalid input",
			body:    `{"start_date":"2026-08-08T00:00:00+09:00","end_date":"2026-08-01T00:00:00+09:00"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req updateHospitalizationRequest
			require.NoError(t, json.Unmarshal([]byte(tt.body), &req))

			input, err := req.toServiceInput()
			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, input.StartDate)
			require.NotNil(t, input.EndDate)
			require.NotNil(t, input.NextVisit)
			assert.Equal(t, tt.wantStart, input.StartDate.Format(time.RFC3339))
			assert.Equal(t, tt.wantEnd, input.EndDate.Format(time.RFC3339))
			assert.Equal(t, tt.wantNext, input.NextVisit.Format(time.DateOnly))
			assert.Equal(t,
				time.Date(2026, 8, 29, 0, 0, 0, 0, time.Local).Format(time.RFC3339),
				input.NextVisit.Format(time.RFC3339))
		})
	}
}
