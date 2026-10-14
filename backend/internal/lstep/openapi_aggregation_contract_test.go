package lstep

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type aggregationContractSchema struct {
	Ref                  string                               `yaml:"$ref"`
	Type                 string                               `yaml:"type"`
	Required             []string                             `yaml:"required"`
	Properties           map[string]aggregationContractSchema `yaml:"properties"`
	Enum                 []string                             `yaml:"enum"`
	AdditionalProperties yaml.Node                            `yaml:"additionalProperties"`
}

var aggregationContractCountKeys = []string{
	"cpm_encounter", "cpm_growing", "cpm_core", "cpm_spot",
	"cpm_noah", "cpm_dormant", "cpm_unclassified",
}

func TestOpenAPI_AggregationCountsContract(t *testing.T) {
	spec := readAggregationOpenAPIContract(t)
	response := spec.Paths["/clinics/{clinic_id}/owners/aggregations"].Get.Responses["200"]
	assert.Equal(t, "#/components/schemas/OwnerAggregationList", response.Content["application/json"].Schema.Ref)
	parent, ok := spec.Components.Schemas["OwnerAggregationList"]
	require.True(t, ok)
	assert.Contains(t, parent.Required, "cpm_stage_counts")
	schema, ok := parent.Properties["cpm_stage_counts"]
	require.True(t, ok)
	assert.Equal(t, "object", schema.Type)
	assert.ElementsMatch(t, aggregationContractCountKeys, schema.Required)
	propertyKeys := make([]string, 0, len(schema.Properties))
	for key, property := range schema.Properties {
		propertyKeys = append(propertyKeys, key)
		assert.Equal(t, "integer", property.Type, key)
	}
	assert.ElementsMatch(t, aggregationContractCountKeys, propertyKeys)
	var additional bool
	require.Equal(t, "!!bool", schema.AdditionalProperties.Tag,
		"additionalProperties must explicitly be false, not an arbitrary integer map")
	require.NoError(t, schema.AdditionalProperties.Decode(&additional))
	assert.False(t, additional)
}

// This checks only the count object's declared keys and integer wire values;
// it is not a general OpenAPI validator.
func aggregationCountsMatchContract(schema aggregationContractSchema, raw []byte) bool {
	var body map[string]json.RawMessage
	if json.Unmarshal(raw, &body) != nil || body == nil {
		return false
	}
	for _, key := range schema.Required {
		if _, ok := body[key]; !ok {
			return false
		}
	}
	for key, value := range body {
		property, declared := schema.Properties[key]
		if !declared || property.Type != "integer" || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return false
		}
		var count int64
		if json.Unmarshal(value, &count) != nil {
			return false
		}
	}
	return true
}

func TestOpenAPI_AggregationCountsPayloads(t *testing.T) {
	schema := readAggregationOpenAPIContract(t).Components.Schemas["OwnerAggregationList"].Properties["cpm_stage_counts"]
	valid := map[string]json.RawMessage{
		"cpm_encounter": json.RawMessage("1"), "cpm_growing": json.RawMessage("0"),
		"cpm_core": json.RawMessage("0"), "cpm_spot": json.RawMessage("2"),
		"cpm_noah": json.RawMessage("0"), "cpm_dormant": json.RawMessage("0"),
		"cpm_unclassified": json.RawMessage("1"),
	}
	raw, err := json.Marshal(valid)
	require.NoError(t, err)
	require.True(t, aggregationCountsMatchContract(schema, raw))
	for _, key := range aggregationContractCountKeys {
		t.Run("missing_"+key, func(t *testing.T) {
			body := make(map[string]json.RawMessage, len(valid))
			for field, value := range valid {
				if field != key {
					body[field] = value
				}
			}
			raw, err := json.Marshal(body)
			require.NoError(t, err)
			assert.False(t, aggregationCountsMatchContract(schema, raw))
		})
	}
	for _, invalid := range []struct {
		name, key, value string
	}{
		{"extra_key", "unexpected", "0"},
		{"string_count", "cpm_core", `"0"`},
		{"fraction_count", "cpm_core", "0.5"},
		{"null_count", "cpm_core", "null"},
	} {
		t.Run(invalid.name, func(t *testing.T) {
			body := make(map[string]json.RawMessage, len(valid)+1)
			for key, value := range valid {
				body[key] = value
			}
			body[invalid.key] = json.RawMessage(invalid.value)
			raw, err := json.Marshal(body)
			require.NoError(t, err)
			assert.False(t, aggregationCountsMatchContract(schema, raw))
		})
	}
	t.Run("null_object", func(t *testing.T) {
		assert.False(t, aggregationCountsMatchContract(schema, []byte("null")))
	})
}

func TestOpenAPI_AggregationCountsHandlerPayload(t *testing.T) {
	schema := readAggregationOpenAPIContract(t).Components.Schemas["OwnerAggregationList"].Properties["cpm_stage_counts"]
	gin.SetMode(gin.TestMode)
	repo := &mockLtvRepositoryCapture{
		fn: func(context.Context, *FindOwnerLTVParams) ([]OwnerLTVRow, error) { return nil, nil },
	}
	h := newHandlerWithAggregationSvc(NewAggregationService(repo, &mockLstepSettingsService{}))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	setClinicID(c)
	h.ListOwnerAggregation(c)
	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	counts, exists := body["cpm_stage_counts"]
	require.True(t, exists)
	assert.True(t, aggregationCountsMatchContract(schema, counts), "handler counts must match the OpenAPI declaration")
}

type aggregationOpenAPIContract struct {
	Components struct {
		Schemas map[string]aggregationContractSchema `yaml:"schemas"`
	} `yaml:"components"`
	Paths map[string]struct {
		Get struct {
			Responses map[string]struct {
				Content map[string]struct {
					Schema aggregationContractSchema `yaml:"schema"`
				} `yaml:"content"`
			} `yaml:"responses"`
		} `yaml:"get"`
	} `yaml:"paths"`
}

func readAggregationOpenAPIContract(t *testing.T) aggregationOpenAPIContract {
	t.Helper()
	src, err := os.ReadFile("../../docs/api.yaml")
	require.NoError(t, err)
	var spec aggregationOpenAPIContract
	require.NoError(t, yaml.Unmarshal(src, &spec))
	return spec
}

func TestOpenAPI_AggregationTimeoutPayloads(t *testing.T) {
	spec := readAggregationOpenAPIContract(t)
	schema := spec.Paths["/clinics/{clinic_id}/owners/aggregations"].Get.Responses["503"].Content["application/json"].Schema
	for _, tt := range []struct {
		name, body string
		valid      bool
	}{
		{"valid", `{"error":"aggregation timed out","code":"aggregation_timeout"}`, true},
		{"missing_error", `{"code":"aggregation_timeout"}`, false},
		{"missing_code", `{"error":"timeout"}`, false},
		{"wrong_code", `{"error":"timeout","code":"other"}`, false},
		{"nonstring_error", `{"error":42,"code":"aggregation_timeout"}`, false},
		{"nonstring_code", `{"error":"timeout","code":42}`, false},
		{"null_error", `{"error":null,"code":"aggregation_timeout"}`, false},
		{"null_body", `null`, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, aggregationTimeoutMatchesContract(schema, []byte(tt.body)))
		})
	}
}

func aggregationTimeoutMatchesContract(schema aggregationContractSchema, raw []byte) bool {
	var body map[string]json.RawMessage
	if json.Unmarshal(raw, &body) != nil || body == nil || schema.Type != "object" {
		return false
	}
	for _, field := range schema.Required {
		value, ok := body[field]
		if !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return false
		}
		property := schema.Properties[field]
		var text string
		if property.Type != "string" || json.Unmarshal(value, &text) != nil {
			return false
		}
		if len(property.Enum) > 0 {
			allowed := false
			for _, candidate := range property.Enum {
				allowed = allowed || candidate == text
			}
			if !allowed {
				return false
			}
		}
	}
	return true
}

func TestOpenAPI_AggregationTimeoutContract(t *testing.T) {
	spec := readAggregationOpenAPIContract(t)
	path, ok := spec.Paths["/clinics/{clinic_id}/owners/aggregations"]
	require.True(t, ok)
	response, ok := path.Get.Responses["503"]
	require.True(t, ok)
	content, ok := response.Content["application/json"]
	require.True(t, ok, "503 must declare its JSON body")
	schema := content.Schema
	assert.Equal(t, "object", schema.Type)
	assert.ElementsMatch(t, []string{"error", "code"}, schema.Required)
	assert.Equal(t, "string", schema.Properties["error"].Type)
	assert.Equal(t, "string", schema.Properties["code"].Type)
	assert.Equal(t, []string{"aggregation_timeout"}, schema.Properties["code"].Enum)

	gin.SetMode(gin.TestMode)
	repo := &mockLtvRepositoryCapture{
		fn: func(ctx context.Context, _ *FindOwnerLTVParams) ([]OwnerLTVRow, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		},
	}
	svc := &aggregationService{
		repo: repo, settingsSvc: &mockLstepSettingsService{}, queryTimeout: time.Nanosecond,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	setClinicID(c)
	newHandlerWithAggregationSvc(svc).ListOwnerAggregation(c)
	require.Equal(t, http.StatusServiceUnavailable, w.Code)
	var body map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	for _, field := range schema.Required {
		value, exists := body[field]
		require.True(t, exists, "handler must return required field %s", field)
		var text string
		require.NoError(t, json.Unmarshal(value, &text))
		assert.NotEmpty(t, text)
		if allowed := schema.Properties[field].Enum; len(allowed) > 0 {
			assert.Contains(t, allowed, text)
		}
	}
}
