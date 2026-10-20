package apicontract

import (
	"go/ast"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Static boundary check: package/runtime and database behavior are separate tests.
func TestNullableInsuranceResponseContract(t *testing.T) {
	for _, tc := range []struct{ path, name, key string }{
		{"../billing/billing_item_request.go", "BillingItemResponse", "is_insurance_applicable"},
		{"../billing/unbilled_revision.go", "unbilledRevisionItem", "is_insurance_applicable"},
		{"../medicalrecord/treatment_response.go", "treatmentResponse", "is_insurance"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fields := collectStructJSONFields(t, tc.path, tc.name)
			field, ok := fields[tc.key]
			require.True(t, ok, "field must be present")
			pointer, ok := field.expr.(*ast.StarExpr)
			require.True(t, ok, "legacy null needs *bool")
			require.True(t, isIdentType(pointer.X, "bool"))
			require.NotContains(t, field.opts, "omitempty", "null must be present in JSON")
		})
	}
	raw, err := os.ReadFile("../../docs/api.yaml")
	require.NoError(t, err)
	var spec struct {
		Components struct {
			Schemas map[string]struct {
				Properties map[string]struct {
					Type     string `yaml:"type"`
					Nullable bool   `yaml:"nullable"`
				} `yaml:"properties"`
			} `yaml:"schemas"`
		} `yaml:"components"`
	}
	require.NoError(t, yaml.Unmarshal(raw, &spec))
	for schema, key := range map[string]string{"BillingItem": "is_insurance_applicable", "Treatment": "is_insurance"} {
		field, ok := spec.Components.Schemas[schema].Properties[key]
		require.True(t, ok)
		require.Equal(t, "boolean", field.Type)
		require.True(t, field.Nullable)
	}
}
