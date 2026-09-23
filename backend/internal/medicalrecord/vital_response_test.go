package medicalrecord

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// UAT-R2-EXCLUSIVE-LOCK: 応答 DTO が現在の行 version を返すこと。
// FE はこの値を次の PATCH の expectedVersion として送り返す。
func TestToVitalResponse_IncludesVersion(t *testing.T) {
	vital := &model.VitalRecord{
		ID:         12,
		ClinicID:   3,
		RecordedAt: time.Date(2026, 5, 28, 10, 30, 0, 0, time.UTC),
		Version:    4,
	}

	resp := toVitalResponse(vital)

	assert.Equal(t, 4, resp.Version)

	body, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"version":4`)
}
