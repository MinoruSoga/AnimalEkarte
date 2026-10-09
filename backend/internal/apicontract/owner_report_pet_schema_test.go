package apicontract

// owner_report_pet_schema_test.go — PR#614 残存指摘 P3: OwnerReportPet ↔ wire DTO parity gate.
//
// wire 側: pet/pet_handler.go の ownerReportResponse.DeceasedAt は *time.Time を
// `json:"deceased_at"`（omitempty 無し）で常時 emit する — 生存ペットでは null。
// schema 側: docs/api.yaml の OwnerReportPet は additionalProperties:false なのに
// deceased_at を宣言しておらず、実レスポンスがスキーマを逸脱していた。
// 本テストは両側を固定する: schema が deceased_at を type:string / format:date-time /
// nullable / required 内に持ち、DTO が omitempty 無しの *time.Time で emit し続けること。

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const ownerReportHandlerPath = "../pet/pet_handler.go"

// TestOwnerReportPetSchemaDeclaresDeceasedAt pins the schema half of the parity:
// OwnerReportPet must declare deceased_at as a nullable date-time string and list it in
// required, because the wire DTO always emits the key (null for living pets) and the
// schema is additionalProperties:false.
func TestOwnerReportPetSchemaDeclaresDeceasedAt(t *testing.T) {
	src, err := os.ReadFile(openapiPath)
	require.NoError(t, err)

	var spec struct {
		Components struct {
			Schemas map[string]yaml.Node `yaml:"schemas"`
		} `yaml:"components"`
	}
	require.NoError(t, yaml.Unmarshal(src, &spec))

	var ownerReportPet struct {
		Required             []string `yaml:"required"`
		AdditionalProperties *bool    `yaml:"additionalProperties"`
		Properties           map[string]struct {
			Type     string `yaml:"type"`
			Format   string `yaml:"format"`
			Nullable bool   `yaml:"nullable"`
		} `yaml:"properties"`
	}
	node, ok := spec.Components.Schemas["OwnerReportPet"]
	require.True(t, ok, "missing OwnerReportPet schema")
	require.NoError(t, node.Decode(&ownerReportPet))

	require.NotNil(t, ownerReportPet.AdditionalProperties)
	assert.False(t, *ownerReportPet.AdditionalProperties,
		"OwnerReportPet must stay additionalProperties:false; undeclared wire keys break the contract")

	deceasedAt, ok := ownerReportPet.Properties["deceased_at"]
	require.True(t, ok, "OwnerReportPet.properties must declare deceased_at: the wire DTO always emits it")
	assert.Equal(t, "string", deceasedAt.Type)
	assert.Equal(t, "date-time", deceasedAt.Format,
		"deceased_at is *time.Time on the wire (RFC3339), not a date-only string")
	assert.True(t, deceasedAt.Nullable,
		"living pets serialize deceased_at as null")
	assert.Contains(t, ownerReportPet.Required, "deceased_at",
		"json tag has no omitempty — the key is always present, so it belongs in required")
}

// TestOwnerReportPetWireDTOMatchesDeceasedAtContract pins the DTO half of the parity:
// ownerReportResponse must keep emitting deceased_at as *time.Time with no omitempty.
// If the DTO drifts (rename, omitempty, type change) the schema declaration must be
// revisited — this test is the tripwire in both directions.
func TestOwnerReportPetWireDTOMatchesDeceasedAtContract(t *testing.T) {
	src, err := os.ReadFile(ownerReportHandlerPath)
	require.NoError(t, err)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, ownerReportHandlerPath, src, 0)
	require.NoError(t, err)

	found := false
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name.Name != "ownerReportResponse" {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			require.True(t, ok, "ownerReportResponse must remain a struct")
			for _, field := range st.Fields.List {
				if field.Tag == nil {
					continue
				}
				tag, err := strconv.Unquote(field.Tag.Value)
				require.NoError(t, err)
				name, opts, _ := strings.Cut(reflect.StructTag(tag).Get("json"), ",")
				if name != "deceased_at" {
					continue
				}
				found = true
				assert.True(t, isTimeTimeType(field.Type),
					"deceased_at must stay *time.Time on the wire (RFC3339 date-time)")
				assert.NotContains(t, strings.Split(opts, ","), "omitempty",
					"omitempty would drop deceased_at for living pets and break the always-emit contract")
			}
		}
	}
	require.True(t, found, "ownerReportResponse must keep a field tagged json:\"deceased_at\"")
}
