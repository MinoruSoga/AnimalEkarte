package medicalrecord

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/animal-ekarte/backend/internal/model"

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
	assert.True(t, json.Valid(result.ResponseJSON), "HTTP failure body must be valid JSON")
	if json.Valid(result.ResponseJSON) {
		var retained string
		require.NoError(t, json.Unmarshal(result.ResponseJSON, &retained))
		assert.True(t, retained == "secret-body\n", "HTTP failure body must be retained")
	}
	assert.NotContains(t, string(result.RequestJSON), "secret-body")
	assert.NotContains(t, string(result.RequestJSON), "test-key")
	assert.NotContains(t, string(result.RequestJSON), "Authorization")
}

func TestOpenAILabItemDrafter_InvalidJSONDraftPersists(t *testing.T) {
	body := []byte("invalid model body")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(body) }))
	t.Cleanup(server.Close)
	_, store, masters, _ := newMappingHarness()
	svc := NewLabMappingDraftService(store, masters, NewOpenAILabItemDrafter(server.Client(), "test-key", server.URL), &scriptAudit{}, snapTx{store: store})
	payload := fujiOne("S1", "BUN-P", mappingSentinelValue, "mg/dl", "")
	view, err := svc.Create(context.Background(), 1, payload, "nx600")
	require.NoError(t, err)
	saved, err := svc.Get(context.Background(), 1, view.Draft.ID)
	require.NoError(t, err)
	require.Len(t, saved.Items, 1)
	assert.True(t, json.Valid(saved.Items[0].LLMResponse), "persisted response must be valid JSON")
	assert.Equal(t, model.LabMappingDraftStatusUnconfirmed, saved.Draft.Status)
	assert.True(t, string(payload) == string(saved.Draft.RawPayload), "raw payload must be retained")
	assert.Equal(t, model.LabMappingItemSourceBlank, saved.Items[0].ItemSource)
	assert.Equal(t, labMappingLLMParse, saved.Items[0].LLMError)
	assert.Nil(t, saved.Items[0].ExamTypeFieldID)
	if json.Valid(saved.Items[0].LLMResponse) {
		var retained string
		require.NoError(t, json.Unmarshal(saved.Items[0].LLMResponse, &retained))
		assert.True(t, retained == string(body), "invalid response must be retained")
	}
}

type labMappingRoundTripFunc func(*http.Request) (*http.Response, error)

func (f labMappingRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type labMappingFailReader struct{}

func (labMappingFailReader) Read([]byte) (int, error) { return 0, errors.New("read failure") }
func (labMappingFailReader) Close() error             { return nil }
func TestOpenAILabItemDrafter_HTTPBodyBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name           string
		body           io.ReadCloser
		transportError error
		wantBody       bool
	}{
		{name: "valid_json", body: io.NopCloser(strings.NewReader(`{"error":"unavailable"}`)), wantBody: true},
		{name: "transport_failure", transportError: errors.New("transport failure")},
		{name: "read_failure", body: labMappingFailReader{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &http.Client{Transport: labMappingRoundTripFunc(func(*http.Request) (*http.Response, error) {
				if tc.transportError != nil {
					return nil, tc.transportError
				}
				return &http.Response{StatusCode: http.StatusBadGateway, Body: tc.body, Header: make(http.Header)}, nil
			})}
			result, err := NewOpenAILabItemDrafter(client, "test-key", "https://example.invalid").DraftItem(context.Background(), LabMappingLLMRequest{Code: "BUN-P"})
			require.NoError(t, err)
			assert.Equal(t, labMappingLLMHTTP, result.ErrorCategory)
			if tc.wantBody {
				assert.True(t, string(result.ResponseJSON) == `{"error":"unavailable"}`, "valid JSON bytes must be preserved")
			} else {
				assert.Nil(t, result.ResponseJSON)
			}
		})
	}
	calls := 0
	client := &http.Client{Transport: labMappingRoundTripFunc(func(*http.Request) (*http.Response, error) { calls++; return nil, errors.New("unexpected HTTP") })}
	_, err := NewOpenAILabItemDrafter(client, "", "https://example.invalid").DraftItem(context.Background(), LabMappingLLMRequest{})
	require.NoError(t, err)
	assert.Zero(t, calls)
}
