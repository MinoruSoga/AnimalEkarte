package apicontract

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestOpenAPIExaminationPrintSnapshotContract(t *testing.T) {
	src, err := os.ReadFile(openapiPath)
	require.NoError(t, err)

	var spec struct {
		Components struct {
			Schemas map[string]yaml.Node `yaml:"schemas"`
		} `yaml:"components"`
		Paths map[string]map[string]yaml.Node `yaml:"paths"`
	}
	require.NoError(t, yaml.Unmarshal(src, &spec))

	snapshotNode, ok := spec.Components.Schemas["ExaminationPrintSnapshot"]
	require.True(t, ok, "missing ExaminationPrintSnapshot schema")
	var snapshot struct {
		Required   []string `yaml:"required"`
		Properties map[string]struct {
			Type   string   `yaml:"type"`
			Enum   []string `yaml:"enum"`
			Format string   `yaml:"format"`
		} `yaml:"properties"`
	}
	require.NoError(t, snapshotNode.Decode(&snapshot))
	assert.Contains(t, snapshot.Required, "print_boundary")
	assert.Contains(t, snapshot.Required, "version")
	assert.Contains(t, snapshot.Required, "items")
	assert.Equal(t, []string{"working", "official"}, snapshot.Properties["kind"].Enum)
	assert.Equal(t, []string{"official", "draft"}, snapshot.Properties["print_boundary"].Enum)
	_, hasDanger := snapshot.Properties["danger_reason"]
	assert.False(t, hasDanger, "print snapshot must not expose danger_reason")

	itemNode, ok := spec.Components.Schemas["ExaminationPrintItem"]
	require.True(t, ok, "missing ExaminationPrintItem schema")
	var item struct {
		Required   []string `yaml:"required"`
		Properties map[string]struct {
			Type string `yaml:"type"`
		} `yaml:"properties"`
	}
	require.NoError(t, itemNode.Decode(&item))
	assert.Contains(t, item.Required, "is_assessed")
	assert.Equal(t, "boolean", item.Properties["is_assessed"].Type)

	operationNode, ok := spec.Paths["/examinations/{id}/print-snapshot"]["get"]
	require.True(t, ok, "missing GET /examinations/{id}/print-snapshot")
	var operation struct {
		OperationID string `yaml:"operationId"`
		Parameters  []struct {
			Name     string `yaml:"name"`
			In       string `yaml:"in"`
			Required bool   `yaml:"required"`
		} `yaml:"parameters"`
		Responses map[string]struct {
			Content map[string]struct {
				Schema struct {
					Ref string `yaml:"$ref"`
				} `yaml:"schema"`
			} `yaml:"content"`
		} `yaml:"responses"`
	}
	require.NoError(t, operationNode.Decode(&operation))
	assert.Equal(t, "getExaminationPrintSnapshot", operation.OperationID)
	var hasVersionQuery bool
	for _, p := range operation.Parameters {
		if p.Name == "version" && p.In == "query" {
			hasVersionQuery = true
			assert.False(t, p.Required)
		}
	}
	assert.True(t, hasVersionQuery, "version query parameter is required by the print contract")
	assert.Equal(t, "#/components/schemas/ExaminationPrintSnapshot", operation.Responses["200"].Content["application/json"].Schema.Ref)
	assert.Contains(t, operation.Responses, "404")
}
