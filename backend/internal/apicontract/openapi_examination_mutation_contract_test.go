package apicontract

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestOpenAPIExaminationMutationContract(t *testing.T) {
	src, err := os.ReadFile(openapiPath)
	require.NoError(t, err)

	var spec struct {
		Components struct {
			Schemas map[string]yaml.Node `yaml:"schemas"`
		} `yaml:"components"`
		Paths map[string]map[string]yaml.Node `yaml:"paths"`
	}
	require.NoError(t, yaml.Unmarshal(src, &spec))

	var examination struct {
		Properties map[string]struct {
			Enum []string `yaml:"enum"`
		} `yaml:"properties"`
	}
	examinationNode, ok := spec.Components.Schemas["Examination"]
	require.True(t, ok, "missing Examination schema")
	require.NoError(t, examinationNode.Decode(&examination))
	assert.Equal(t, []string{
		"pending",
		"in_progress",
		"result_entered",
		"completed",
		"confirmed",
	}, examination.Properties["status"].Enum)

	for _, mutation := range []struct {
		method string
		path   string
	}{
		{method: "patch", path: "/examinations/{id}"},
		{method: "delete", path: "/examinations/{id}"},
		{method: "put", path: "/examinations/{id}/items"},
	} {
		operationNode, ok := spec.Paths[mutation.path][mutation.method]
		require.True(t, ok, "missing %s operation for %s", mutation.method, mutation.path)
		var operation struct {
			Responses map[string]yaml.Node `yaml:"responses"`
		}
		require.NoError(t, operationNode.Decode(&operation))
		assert.Contains(t, operation.Responses, "409", "missing confirmed-mutation conflict response for %s %s", mutation.method, mutation.path)
	}
}
