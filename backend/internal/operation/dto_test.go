package operation

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestParseOperationKey(t *testing.T) {
	const canonical = "11111111-1111-4111-8111-111111111111"

	cases := []struct {
		name    string
		raw     string
		want    uuid.UUID
		wantErr bool
	}{
		{name: "canonical lowercase hyphen form", raw: canonical, want: uuid.MustParse(canonical)},
		{name: "uppercase digits and letters", raw: "11111111-1111-4111-8111-11111111111A", wantErr: true},
		{name: "braced form", raw: "{" + canonical + "}", wantErr: true},
		{name: "nil uuid", raw: "00000000-0000-0000-0000-000000000000", wantErr: true},
		{name: "empty", raw: "", wantErr: true},
		{name: "without hyphens", raw: "11111111111141118111111111111111", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseOperationKey(tc.raw)
			if tc.wantErr {
				require.Error(t, err)
				require.True(t, apperrors.IsInvalidInput(err), "error must be an invalid-input error: %v", err)
				require.Equal(t, uuid.UUID{}, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

// TestNewResultResponse asserts what crosses the wire: the result body is
// disclosed only for a definite success, and the internal effect summary never
// appears.
func TestNewResultResponse(t *testing.T) {
	key := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	recordedAt := time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)
	stored := &Result{
		ResultRefs:        map[string]string{"billing_id": "42"},
		CommittedVersions: map[string]string{"billing_version": "3"},
		EffectSummary:     map[string]string{"internal_note": "must-not-be-disclosed"},
		RecordedAt:        recordedAt,
	}

	t.Run("discloses the result for a definite success", func(t *testing.T) {
		for _, status := range []Status{StatusSucceeded, StatusReplayed} {
			response := NewResultResponse(key, Outcome{Status: status, Result: stored})

			require.Equal(t, status, response.Status)
			require.Equal(t, key.String(), response.OperationKey)
			require.Equal(t, stored.ResultRefs, response.ResultRefs)
			require.Equal(t, stored.CommittedVersions, response.CommittedVersions)
			require.NotNil(t, response.RecordedAt)
			require.Equal(t, recordedAt, *response.RecordedAt)

			raw, err := json.Marshal(response)
			require.NoError(t, err)
			payload := string(raw)
			require.Contains(t, payload, `"result_refs"`)
			require.Contains(t, payload, `"committed_versions"`)
			require.Contains(t, payload, `"recorded_at"`)
			require.NotContains(t, payload, "effect_summary")
			require.NotContains(t, payload, "must-not-be-disclosed")
		}
	})

	t.Run("hides the result for every non-success status", func(t *testing.T) {
		for _, status := range []Status{
			StatusNotSucceeded,
			StatusKeyConflict,
			StatusInProgress,
			StatusResultUnavailable,
			StatusScopeNotFound,
			StatusNotAuthorized,
			StatusValidation,
			StatusVersionConflict,
			StatusDependencyChanged,
			StatusEditRightLost,
		} {
			// The outcome deliberately carries a stored result to prove that the
			// status, not the presence of a result, decides disclosure.
			response := NewResultResponse(key, Outcome{Status: status, Result: stored})

			require.Equal(t, status, response.Status)
			require.Equal(t, key.String(), response.OperationKey)
			require.Nil(t, response.ResultRefs)
			require.Nil(t, response.CommittedVersions)
			require.Nil(t, response.RecordedAt)

			raw, err := json.Marshal(response)
			require.NoError(t, err)
			payload := string(raw)
			require.Contains(t, payload, `"status":"`+string(status)+`"`)
			require.NotContains(t, payload, "result_refs")
			require.NotContains(t, payload, "committed_versions")
			require.NotContains(t, payload, "recorded_at")
			require.NotContains(t, payload, "effect_summary")
			require.NotContains(t, payload, "must-not-be-disclosed")
		}
	})
}
