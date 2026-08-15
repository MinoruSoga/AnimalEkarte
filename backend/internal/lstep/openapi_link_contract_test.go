package lstep

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type liffLinkOpenAPISchema struct {
	Required             []string                           `yaml:"required"`
	AdditionalProperties *bool                              `yaml:"additionalProperties"`
	Properties           map[string]liffLinkOpenAPIProperty `yaml:"properties"`
}

type liffLinkOpenAPIProperty struct {
	Type      string `yaml:"type"`
	MaxLength *int   `yaml:"maxLength"`
}

type liffLinkOpenAPIResponse struct {
	Content map[string]any `yaml:"content"`
}

func TestOpenAPI_LiffLinkAccountContractMatchesBoundary(t *testing.T) {
	specBytes, err := os.ReadFile(filepath.Join("..", "..", "docs", "api.yaml"))
	require.NoError(t, err)

	var spec struct {
		Components struct {
			Schemas map[string]liffLinkOpenAPISchema `yaml:"schemas"`
		} `yaml:"components"`
		Paths map[string]struct {
			Post struct {
				Responses map[string]liffLinkOpenAPIResponse `yaml:"responses"`
			} `yaml:"post"`
		} `yaml:"paths"`
	}
	require.NoError(t, yaml.Unmarshal(specBytes, &spec))

	schema, ok := spec.Components.Schemas["LiffLinkAccountRequest"]
	require.True(t, ok)
	assert.ElementsMatch(t, []string{"link_token", "line_id_token"}, schema.Required)
	require.NotNil(t, schema.AdditionalProperties)
	assert.False(t, *schema.AdditionalProperties)
	assert.NotContains(t, schema.Properties, "force")

	tests := []struct {
		name          string
		property      string
		wantMaxLength int
	}{
		{
			name:          "link token limit",
			property:      "link_token",
			wantMaxLength: maxLineLinkTokenChars,
		},
		{
			name:          "LINE ID token limit",
			property:      "line_id_token",
			wantMaxLength: maxLineIDTokenChars,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			property, ok := schema.Properties[tt.property]
			require.True(t, ok)
			assert.Equal(t, "string", property.Type)
			require.NotNil(t, property.MaxLength)
			assert.Equal(t, tt.wantMaxLength, *property.MaxLength)
		})
	}

	linkPath, ok := spec.Paths["/api/liff/{clinicId}/link"]
	require.True(t, ok)
	response, ok := linkPath.Post.Responses["204"]
	require.True(t, ok, "successful public LINE link must return no owner record")
	assert.Empty(t, response.Content)
	assert.NotContains(t, linkPath.Post.Responses, "200")
	for _, status := range []string{"400", "401", "404", "409", "413", "500", "502"} {
		assert.Contains(t, linkPath.Post.Responses, status)
	}
}
