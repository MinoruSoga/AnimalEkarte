package apicontract

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestOpenAPITreatmentDoseDeviationReasonContract は TASK-377 の
// dose_deviation_reason field が Create/Update treatment request に載ることを固定する。
func TestOpenAPITreatmentDoseDeviationReasonContract(t *testing.T) {
	src, err := os.ReadFile(openapiPath)
	require.NoError(t, err)

	var spec struct {
		Components struct {
			Schemas map[string]yaml.Node `yaml:"schemas"`
		} `yaml:"components"`
	}
	require.NoError(t, yaml.Unmarshal(src, &spec))

	for _, schemaName := range []string{"CreateTreatmentRequest", "UpdateTreatmentRequest"} {
		node, ok := spec.Components.Schemas[schemaName]
		require.True(t, ok, "missing %s schema", schemaName)
		var schema struct {
			Properties map[string]struct {
				Type      string `yaml:"type"`
				MinLength int    `yaml:"minLength"`
				MaxLength int    `yaml:"maxLength"`
			} `yaml:"properties"`
		}
		require.NoError(t, node.Decode(&schema))
		prop, ok := schema.Properties["dose_deviation_reason"]
		require.True(t, ok, "%s must declare dose_deviation_reason", schemaName)
		assert.Equal(t, "string", prop.Type)
		assert.Equal(t, 1, prop.MinLength)
		assert.Equal(t, 500, prop.MaxLength)
	}
}
