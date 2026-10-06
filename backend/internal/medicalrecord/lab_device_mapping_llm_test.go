package medicalrecord

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAILabItemDrafter_DoesNotSendMeasuredValues(t *testing.T) {
	var got map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/responses", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.NoError(t, json.Unmarshal(body, &got))
		assert.NotContains(t, string(body), mappingSentinelValue)
		_, hasTools := got["tools"]
		assert.False(t, hasTools)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"output_text":"{\"exam_type_field_id\":7,\"field_name\":\"BUN\"}"}`))
	}))
	t.Cleanup(server.Close)

	drafter := NewOpenAILabItemDrafter(server.Client(), "test-key", server.URL)
	result, err := drafter.DraftItem(context.Background(), LabMappingLLMRequest{
		SourceType: "fuji_nx600",
		Species:    "canine",
		Code:       "BUN-P",
		DeviceUnit: "mg/dl",
		Candidates: []LabMappingFieldCandidate{{ID: 7, Name: "BUN", Unit: "mg/dl"}},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(7), result.ExamTypeFieldID)
	assert.Equal(t, "gpt-6-luna", got["model"])
	assert.Equal(t, false, got["store"])
	reasoning, _ := got["reasoning"].(map[string]any)
	assert.Equal(t, "none", reasoning["effort"])
	assert.NotContains(t, string(result.RequestJSON), "test-key")
	assert.NotContains(t, string(result.RequestJSON), mappingSentinelValue)
}

func TestOpenAILabItemDrafter_EmptyKeyAndHTTPFailure(t *testing.T) {
	drafter := NewOpenAILabItemDrafter(nil, "", "https://example.invalid")
	result, err := drafter.DraftItem(context.Background(), LabMappingLLMRequest{Code: "BUN-P"})
	require.NoError(t, err)
	assert.Equal(t, labMappingLLMUnconfigured, result.ErrorCategory)
	assert.NotEmpty(t, result.RequestJSON)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "secret-body", http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)
	drafter = NewOpenAILabItemDrafter(server.Client(), "test-key", server.URL)
	result, err = drafter.DraftItem(context.Background(), LabMappingLLMRequest{Code: "BUN-P"})
	require.NoError(t, err)
	assert.Equal(t, labMappingLLMHTTP, result.ErrorCategory)
	assert.Nil(t, result.ResponseJSON)
	assert.NotContains(t, string(result.RequestJSON), "secret-body")
}
