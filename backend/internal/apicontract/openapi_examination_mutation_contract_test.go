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
			Type     string   `yaml:"type"`
			Format   string   `yaml:"format"`
			Enum     []string `yaml:"enum"`
			Minimum  int      `yaml:"minimum"`
			ReadOnly bool     `yaml:"readOnly"`
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
	revisionPointer, ok := examination.Properties["current_revision_version"]
	require.True(t, ok, "missing current_revision_version from Examination response contract")
	assert.Equal(t, "integer", revisionPointer.Type)
	assert.Equal(t, "int64", revisionPointer.Format)
	assert.Equal(t, 1, revisionPointer.Minimum)
	assert.True(t, revisionPointer.ReadOnly)

	var unconfirmRequest struct {
		Required   []string `yaml:"required"`
		Properties map[string]struct {
			Type      string `yaml:"type"`
			MinLength int    `yaml:"minLength"`
			MaxLength int    `yaml:"maxLength"`
			Pattern   string `yaml:"pattern"`
		} `yaml:"properties"`
	}
	unconfirmNode, ok := spec.Components.Schemas["UnconfirmExaminationRequest"]
	require.True(t, ok, "missing UnconfirmExaminationRequest schema")
	require.NoError(t, unconfirmNode.Decode(&unconfirmRequest))
	assert.Equal(t, []string{"reason"}, unconfirmRequest.Required)
	assert.Equal(t, "string", unconfirmRequest.Properties["reason"].Type)
	assert.Equal(t, 1, unconfirmRequest.Properties["reason"].MinLength)
	assert.Equal(t, 500, unconfirmRequest.Properties["reason"].MaxLength)
	assert.Equal(t, `\S`, unconfirmRequest.Properties["reason"].Pattern)

	for _, mutation := range []struct {
		method string
		path   string
		body   string
	}{
		{method: "patch", path: "/examinations/{id}"},
		{method: "delete", path: "/examinations/{id}"},
		{method: "put", path: "/examinations/{id}/items"},
		{method: "post", path: "/examinations/{id}/unconfirm", body: "#/components/schemas/UnconfirmExaminationRequest"},
	} {
		operationNode, ok := spec.Paths[mutation.path][mutation.method]
		require.True(t, ok, "missing %s operation for %s", mutation.method, mutation.path)
		var operation struct {
			Responses   map[string]yaml.Node `yaml:"responses"`
			RequestBody struct {
				Required bool `yaml:"required"`
				Content  map[string]struct {
					Schema struct {
						Ref string `yaml:"$ref"`
					} `yaml:"schema"`
				} `yaml:"content"`
			} `yaml:"requestBody"`
		}
		require.NoError(t, operationNode.Decode(&operation))
		assert.Contains(t, operation.Responses, "409", "missing confirmed-mutation conflict response for %s %s", mutation.method, mutation.path)
		if mutation.body != "" {
			assert.True(t, operation.RequestBody.Required)
			assert.Equal(t, mutation.body, operation.RequestBody.Content["application/json"].Schema.Ref)
			assert.Contains(t, operation.Responses, "200")
		}
	}
}
