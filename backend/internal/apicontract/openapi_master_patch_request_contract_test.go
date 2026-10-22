package apicontract

import (
	"go/ast"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type masterPatchContractSpec struct {
	Paths map[string]map[string]struct {
		OperationID string `yaml:"operationId"`
		RequestBody struct {
			Content map[string]struct {
				Schema struct {
					Ref string `yaml:"$ref"`
				} `yaml:"schema"`
			} `yaml:"content"`
		} `yaml:"requestBody"`
	} `yaml:"paths"`
	Components struct {
		Schemas map[string]struct {
			Required   []string                  `yaml:"required"`
			Properties map[string]map[string]any `yaml:"properties"`
		} `yaml:"schemas"`
	} `yaml:"components"`
}

func loadMasterPatchContractSpec(t *testing.T) masterPatchContractSpec {
	t.Helper()
	src, err := os.ReadFile(openapiPath)
	require.NoError(t, err)
	var spec masterPatchContractSpec
	require.NoError(t, yaml.Unmarshal(src, &spec))
	return spec
}

func requestSchemaRef(t *testing.T, spec masterPatchContractSpec, path, method string) string {
	t.Helper()
	op, ok := spec.Paths[path][method]
	require.True(t, ok, "%s %s must exist", method, path)
	return op.RequestBody.Content["application/json"].Schema.Ref
}

// TestOpenAPIMasterPatchRequestContract は EMR-533 / EMR-536 の PATCH 専用 request schema を固定する。
// 共用の応答/POST schema を緩和・拡張せず、PATCH requestBody だけを専用 schema へ結線し、
// 専用 schema の他 property は共用 schema の既存契約と同一であることを静的に確認する（runtime/DB は別試験）。
func TestOpenAPIMasterPatchRequestContract(t *testing.T) {
	spec := loadMasterPatchContractSpec(t)
	schemas := spec.Components.Schemas

	for _, tc := range []struct {
		path, shared, patch, changed string
	}{
		{"/masters/hospitalization-plans", "HospitalizationPlan", "UpdateHospitalizationPlanRequest", "description"},
		{"/masters/medicines", "Medicine", "UpdateMedicineRequest", "clear_strength"},
	} {
		t.Run(tc.patch, func(t *testing.T) {
			assert.Equal(t, "#/components/schemas/"+tc.patch, requestSchemaRef(t, spec, tc.path+"/{id}", "patch"))
			assert.Equal(t, "#/components/schemas/"+tc.shared, requestSchemaRef(t, spec, tc.path, "post"),
				"POST must keep the shared schema")

			shared, ok := schemas[tc.shared]
			require.True(t, ok)
			patch, ok := schemas[tc.patch]
			require.True(t, ok)
			assert.Equal(t, shared.Required, patch.Required, "required must not drift")
			for name, prop := range shared.Properties {
				if prop["readOnly"] == true {
					assert.NotContains(t, patch.Properties, name, "readOnly %s is not a PATCH input", name)
					continue
				}
				if name == tc.changed {
					continue
				}
				assert.Equal(t, prop, patch.Properties[name], "%s.%s must equal the shared contract", tc.patch, name)
			}
		})
	}

	t.Run("EMR-533 description null is PATCH-only", func(t *testing.T) {
		patch := schemas["UpdateHospitalizationPlanRequest"].Properties["description"]
		assert.Equal(t, "string", patch["type"])
		assert.Equal(t, true, patch["nullable"])
		shared := schemas["HospitalizationPlan"].Properties["description"]
		assert.Equal(t, "string", shared["type"])
		assert.NotContains(t, shared, "nullable", "response/POST description must not be relaxed")
	})

	t.Run("EMR-536 clear_strength is PATCH-only", func(t *testing.T) {
		patch := schemas["UpdateMedicineRequest"].Properties["clear_strength"]
		assert.Equal(t, "boolean", patch["type"])
		assert.Equal(t, true, patch["nullable"], "Go bool binding accepts null as false")
		assert.NotContains(t, schemas["Medicine"].Properties, "clear_strength", "GET/POST must not expose the flag")
	})
}

// TestMasterPatchRequestGoWireMatchesOpenAPI は OpenAPI の null 許容が Go の decode 型と一致することを固定する。
func TestMasterPatchRequestGoWireMatchesOpenAPI(t *testing.T) {
	plan := collectStructJSONFields(t, "../medicalrecord/hospitalization_plan_request.go", "updateHospitalizationPlanRequest")
	description, ok := plan["description"]
	require.True(t, ok)
	pointer, ok := description.expr.(*ast.StarExpr)
	require.True(t, ok, "description null/omission must decode to nil (keep)")
	assert.True(t, isIdentType(pointer.X, "string"))

	medicine := collectStructJSONFields(t, "../medicalrecord/medicine_request.go", "updateMedicineRequest")
	clearStrength, ok := medicine["clear_strength"]
	require.True(t, ok)
	assert.True(t, isIdentType(clearStrength.expr, "bool"), "null/omission/false must all mean no clear")
}
