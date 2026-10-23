package apicontract

import (
	"go/ast"
	"os"
	"slices"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// masterPatchOperation は operation だけを decode する。path item には配列の path-level
// parameters もあるため、Paths は yaml.Node で受けて必要な method だけをこの型へ decode する。
type masterPatchOperation struct {
	OperationID string `yaml:"operationId"`
	RequestBody struct {
		Content map[string]struct {
			Schema struct {
				Ref string `yaml:"$ref"`
			} `yaml:"schema"`
		} `yaml:"content"`
	} `yaml:"requestBody"`
	Responses map[string]yaml.Node `yaml:"responses"`
}

type masterPatchContractSpec struct {
	Paths      map[string]map[string]yaml.Node `yaml:"paths"`
	Components struct {
		Schemas map[string]struct {
			Required   []string             `yaml:"required"`
			Properties map[string]yaml.Node `yaml:"properties"`
		} `yaml:"schemas"`
	} `yaml:"components"`
}

// canonicalYAMLNode は property schema の部分木を、行・列・コメント・style・mapping の key 順に
// 依らず比較するための正規形。検査範囲を狭めないよう、子孫の key / 値 / tag をすべて保持する。
type canonicalYAMLNode struct {
	Kind  yaml.Kind
	Tag   string
	Value string
	Items []canonicalYAMLNode
	Pairs []canonicalYAMLPair
}

type canonicalYAMLPair struct {
	Key   string
	Value canonicalYAMLNode
}

func canonicalizeYAMLNode(node *yaml.Node) canonicalYAMLNode {
	if node.Kind == yaml.AliasNode {
		return canonicalizeYAMLNode(node.Alias)
	}
	canonical := canonicalYAMLNode{Kind: node.Kind, Tag: node.ShortTag(), Value: node.Value}
	if node.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(node.Content); i += 2 {
			canonical.Pairs = append(canonical.Pairs, canonicalYAMLPair{
				Key:   node.Content[i].Value,
				Value: canonicalizeYAMLNode(node.Content[i+1]),
			})
		}
		sort.Slice(canonical.Pairs, func(a, b int) bool { return canonical.Pairs[a].Key < canonical.Pairs[b].Key })
		return canonical
	}
	for _, child := range node.Content {
		canonical.Items = append(canonical.Items, canonicalizeYAMLNode(child))
	}
	return canonical
}

// masterPatchProperty は property schema の mapping から key の値を正規形で返す。key が無ければ ok=false。
func masterPatchProperty(prop yaml.Node, key string) (canonicalYAMLNode, bool) {
	canonical := canonicalizeYAMLNode(&prop)
	for _, pair := range canonical.Pairs {
		if pair.Key == key {
			return pair.Value, true
		}
	}
	return canonicalYAMLNode{}, false
}

func (n canonicalYAMLNode) isTrue() bool {
	return n.Kind == yaml.ScalarNode && n.Tag == "!!bool" && n.Value == "true"
}

func yamlScalarOf(tag, value string) canonicalYAMLNode {
	return canonicalYAMLNode{Kind: yaml.ScalarNode, Tag: tag, Value: value}
}

func yamlStringOf(value string) canonicalYAMLNode { return yamlScalarOf("!!str", value) }

var yamlTrue = yamlScalarOf("!!bool", "true")

func loadMasterPatchContractSpec(t *testing.T) masterPatchContractSpec {
	t.Helper()
	src, err := os.ReadFile(openapiPath)
	require.NoError(t, err)
	var spec masterPatchContractSpec
	require.NoError(t, yaml.Unmarshal(src, &spec))
	return spec
}

func masterPatchOperationOf(t *testing.T, spec masterPatchContractSpec, path, method string) masterPatchOperation {
	t.Helper()
	node, ok := spec.Paths[path][method]
	require.True(t, ok, "%s %s must exist", method, path)
	var op masterPatchOperation
	require.NoError(t, node.Decode(&op))
	return op
}

func requestSchemaRef(t *testing.T, spec masterPatchContractSpec, path, method string) string {
	t.Helper()
	return masterPatchOperationOf(t, spec, path, method).RequestBody.Content["application/json"].Schema.Ref
}

// TestOpenAPIMasterPatchRequestContract は EMR-533 / EMR-536 の PATCH 専用 request schema を固定する。
// 共用の応答/POST schema を緩和・拡張せず、PATCH requestBody だけを専用 schema へ結線し、
// 専用 schema の他 property は共用 schema の既存契約と同一であることを静的に確認する（runtime/DB は別試験）。
func TestOpenAPIMasterPatchRequestContract(t *testing.T) {
	spec := loadMasterPatchContractSpec(t)
	schemas := spec.Components.Schemas

	for _, tc := range []struct {
		path, shared, patch string
		changed             []string
	}{
		{"/masters/hospitalization-plans", "HospitalizationPlan", "UpdateHospitalizationPlanRequest", []string{"description", "body_size", "billing_unit"}},
		{"/masters/medicines", "Medicine", "UpdateMedicineRequest", []string{"clear_strength", "is_bill_discount_excluded", "version"}},
	} {
		t.Run(tc.patch, func(t *testing.T) {
			assert.Equal(t, "#/components/schemas/"+tc.patch, requestSchemaRef(t, spec, tc.path+"/{id}", "patch"))

			shared, ok := schemas[tc.shared]
			require.True(t, ok)
			patch, ok := schemas[tc.patch]
			require.True(t, ok)
			var writableRequired []string
			for _, name := range shared.Required {
				if readOnly, ok := masterPatchProperty(shared.Properties[name], "readOnly"); !ok || !readOnly.isTrue() {
					writableRequired = append(writableRequired, name)
				}
			}
			assert.Equal(t, writableRequired, patch.Required, "required must not drift")
			for name, prop := range shared.Properties {
				if slices.Contains(tc.changed, name) {
					continue
				}
				if readOnly, ok := masterPatchProperty(prop, "readOnly"); ok && readOnly.isTrue() {
					assert.NotContains(t, patch.Properties, name, "readOnly %s is not a PATCH input", name)
					continue
				}
				patchProp, ok := patch.Properties[name]
				require.True(t, ok, "%s.%s must exist in the PATCH schema", tc.patch, name)
				assert.Equal(t, canonicalizeYAMLNode(&prop), canonicalizeYAMLNode(&patchProp), "%s.%s must equal the shared contract", tc.patch, name)
			}
		})
	}
	assert.Equal(t, "#/components/schemas/HospitalizationPlan", requestSchemaRef(t, spec, "/masters/hospitalization-plans", "post"),
		"POST must keep the shared schema")
	assert.Equal(t, "#/components/schemas/CreateMedicineRequest", requestSchemaRef(t, spec, "/masters/medicines", "post"),
		"EMR-445 A01 splits the medicine POST request from the response schema")

	t.Run("EMR-533 description null is PATCH-only", func(t *testing.T) {
		patch := schemas["UpdateHospitalizationPlanRequest"].Properties["description"]
		patchType, _ := masterPatchProperty(patch, "type")
		assert.Equal(t, yamlStringOf("string"), patchType)
		patchNullable, _ := masterPatchProperty(patch, "nullable")
		assert.Equal(t, yamlTrue, patchNullable)
		shared := schemas["HospitalizationPlan"].Properties["description"]
		sharedType, _ := masterPatchProperty(shared, "type")
		assert.Equal(t, yamlStringOf("string"), sharedType)
		_, sharedNullable := masterPatchProperty(shared, "nullable")
		assert.False(t, sharedNullable, "response/POST description must not be relaxed")
	})

	t.Run("EMR-533 body_size and billing_unit enums list null only in PATCH", func(t *testing.T) {
		for name, want := range map[string][]string{
			"body_size":    {"small", "medium", "large"},
			"billing_unit": {"per_day", "per_night"},
		} {
			patch := schemas["UpdateHospitalizationPlanRequest"].Properties[name]
			patchEnum, ok := masterPatchProperty(patch, "enum")
			require.True(t, ok, "%s PATCH enum must exist", name)
			wantItems := make([]canonicalYAMLNode, 0, len(want)+1)
			for _, value := range want {
				wantItems = append(wantItems, yamlStringOf(value))
			}
			assert.Equal(t, wantItems, patchEnum.Items[:len(patchEnum.Items)-1], "%s PATCH enum values must equal the shared values", name)
			last := patchEnum.Items[len(patchEnum.Items)-1]
			assert.Equal(t, "!!null", last.Tag, "%s PATCH enum must end with an explicit null", name)
			patchNullable, _ := masterPatchProperty(patch, "nullable")
			assert.Equal(t, yamlTrue, patchNullable)

			shared := schemas["HospitalizationPlan"].Properties[name]
			sharedEnum, ok := masterPatchProperty(shared, "enum")
			require.True(t, ok, "%s shared enum must exist", name)
			assert.Equal(t, wantItems, sharedEnum.Items, "%s shared enum must not list null", name)
		}
	})

	t.Run("EMR-536 clear_strength is PATCH-only", func(t *testing.T) {
		patch := schemas["UpdateMedicineRequest"].Properties["clear_strength"]
		patchType, _ := masterPatchProperty(patch, "type")
		assert.Equal(t, yamlStringOf("boolean"), patchType)
		patchNullable, _ := masterPatchProperty(patch, "nullable")
		assert.Equal(t, yamlTrue, patchNullable, "Go bool binding accepts null as false")
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
