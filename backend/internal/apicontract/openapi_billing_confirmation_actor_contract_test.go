package apicontract

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type billingConfirmationRequestSchema struct {
	Required             []string                                     `yaml:"required"`
	Properties           map[string]billingConfirmationPropertySchema `yaml:"properties"`
	AdditionalProperties *bool                                        `yaml:"additionalProperties"`
}

type billingConfirmationPropertySchema struct {
	MaxLength int    `yaml:"maxLength"`
	Pattern   string `yaml:"pattern"`
}

type billingConfirmationMediaType struct {
	Schema billingConfirmationRequestSchema `yaml:"schema"`
}

type billingConfirmationRequestBody struct {
	Content map[string]billingConfirmationMediaType `yaml:"content"`
}

type billingConfirmationOperation struct {
	RequestBody billingConfirmationRequestBody `yaml:"requestBody"`
	Responses   map[string]yaml.Node           `yaml:"responses"`
}

func TestOpenAPIBillingConfirmationRequestsUseAuthenticatedActor(t *testing.T) {
	src, err := os.ReadFile(openapiPath)
	require.NoError(t, err)

	var spec struct {
		Paths map[string]map[string]yaml.Node `yaml:"paths"`
	}
	require.NoError(t, yaml.Unmarshal(src, &spec))

	tests := []struct {
		name           string
		path           string
		wantProperties []string
		wantRequired   []string
		wantMaxLengths map[string]int
		wantPatterns   map[string]string
		forbiddenActor string
	}{
		{
			name:           "confirm",
			path:           "/medical-records/{id}/billing-confirmation/confirm",
			wantProperties: []string{"memo"},
			wantMaxLengths: map[string]int{"memo": 1000},
			forbiddenActor: "confirmed_by",
		},
		{
			name:           "return",
			path:           "/medical-records/{id}/billing-confirmation/return",
			wantProperties: []string{"memo", "return_reason"},
			wantRequired:   []string{"return_reason"},
			wantMaxLengths: map[string]int{
				"memo":          1000,
				"return_reason": 500,
			},
			wantPatterns:   map[string]string{"return_reason": `\S`},
			forbiddenActor: "returned_by",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operationNode, ok := spec.Paths[tt.path]["post"]
			require.True(t, ok, "missing POST operation for %s", tt.path)
			var operation billingConfirmationOperation
			require.NoError(t, operationNode.Decode(&operation))
			mediaType, ok := operation.RequestBody.Content["application/json"]
			require.True(t, ok, "missing application/json request body")

			propertyNames := make([]string, 0, len(mediaType.Schema.Properties))
			for name := range mediaType.Schema.Properties {
				propertyNames = append(propertyNames, name)
			}
			assert.ElementsMatch(t, tt.wantProperties, propertyNames)
			assert.ElementsMatch(t, tt.wantRequired, mediaType.Schema.Required)
			assert.NotContains(t, mediaType.Schema.Properties, tt.forbiddenActor)
			require.NotNil(t, mediaType.Schema.AdditionalProperties)
			assert.False(t, *mediaType.Schema.AdditionalProperties)
			for property, maxLength := range tt.wantMaxLengths {
				assert.Equal(t, maxLength, mediaType.Schema.Properties[property].MaxLength)
			}
			for property, pattern := range tt.wantPatterns {
				assert.Equal(t, pattern, mediaType.Schema.Properties[property].Pattern)
			}
			for _, status := range []string{
				"200",
				"400",
				"401",
				"403",
				"404",
				"409",
				"413",
				"415",
				"500",
			} {
				assert.Contains(
					t,
					operation.Responses,
					status,
					"missing documented %s response",
					status,
				)
			}
		})
	}
}
