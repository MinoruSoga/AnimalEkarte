package apicontract

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

const (
	medicalRecordResponsePath       = "../medicalrecord/medical_record_response.go"
	medicalRecordWeightResponsePath = "../medicalrecord/medical_record_weight_response.go"
	medicalRecordResponsesTSPath    = "../../../frontend/src/types/generated/medicalrecord-responses.ts"
)

type weightWireField struct {
	expr ast.Expr
	opts []string
}

func collectStructJSONFields(t *testing.T, path, typeName string) map[string]weightWireField {
	t.Helper()
	src, err := os.ReadFile(path)
	require.NoError(t, err)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, 0)
	require.NoError(t, err)

	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name.Name != typeName {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			require.True(t, ok, "%s must remain a struct", typeName)
			fields := map[string]weightWireField{}
			for _, field := range st.Fields.List {
				if field.Tag == nil {
					continue
				}
				tag, err := strconv.Unquote(field.Tag.Value)
				require.NoError(t, err)
				name, opts, _ := strings.Cut(reflect.StructTag(tag).Get("json"), ",")
				if name == "" || name == "-" {
					continue
				}
				var optList []string
				if opts != "" {
					optList = strings.Split(opts, ",")
				}
				fields[name] = weightWireField{expr: field.Type, opts: optList}
			}
			return fields
		}
	}
	t.Fatalf("type %s not found in %s", typeName, path)
	return nil
}

func isIdentType(expr ast.Expr, name string) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == name
}

func isSelectorType(expr ast.Expr, pkg, name string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	x, ok := sel.X.(*ast.Ident)
	return ok && x.Name == pkg && sel.Sel.Name == name
}

func extractTSInterfaceBody(t *testing.T, src, name string) string {
	t.Helper()
	marker := "export interface " + name + " {"
	idx := strings.Index(src, marker)
	require.NotEqual(t, -1, idx, "generated TS must export interface %s", name)

	body := stripTSComments(src[idx+len(marker):])
	depth := 1
	for i, r := range body {
		switch r {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return body[:i]
			}
		}
	}
	t.Fatalf("unterminated interface %s in generated TS", name)
	return ""
}

func stripTSComments(src string) string {
	var out strings.Builder
	i := 0
	for i < len(src) {
		if i+1 < len(src) && src[i] == '/' && src[i+1] == '/' {
			for i < len(src) && src[i] != '\n' {
				i++
			}
			continue
		}
		if i+1 < len(src) && src[i] == '/' && src[i+1] == '*' {
			i += 2
			for i+1 < len(src) && !(src[i] == '*' && src[i+1] == '/') {
				i++
			}
			i += 2
			continue
		}
		out.WriteByte(src[i])
		i++
	}
	return out.String()
}

func normalizeTSFields(body string) []string {
	compact := strings.Map(func(r rune) rune {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, body)
	var fields []string
	for _, f := range strings.Split(compact, ";") {
		if f != "" {
			fields = append(fields, f)
		}
	}
	return fields
}

func TestMedicalRecordWeightOpenAPIContract(t *testing.T) {
	src, err := os.ReadFile(openapiPath)
	require.NoError(t, err)

	var spec struct {
		Components struct {
			Schemas map[string]yaml.Node `yaml:"schemas"`
		} `yaml:"components"`
	}
	require.NoError(t, yaml.Unmarshal(src, &spec))

	t.Run("MedicalRecord requires non-nullable weight_measurements array", func(t *testing.T) {
		node, ok := spec.Components.Schemas["MedicalRecord"]
		require.True(t, ok, "missing MedicalRecord schema")

		var mr struct {
			Required   []string `yaml:"required"`
			Properties map[string]struct {
				Type     string `yaml:"type"`
				Nullable bool   `yaml:"nullable"`
				ReadOnly bool   `yaml:"readOnly"`
				Items    struct {
					Ref      string `yaml:"$ref"`
					Nullable bool   `yaml:"nullable"`
				} `yaml:"items"`
			} `yaml:"properties"`
		}
		require.NoError(t, node.Decode(&mr))

		prop, ok := mr.Properties["weight_measurements"]
		require.True(t, ok,
			"MedicalRecord.properties must declare weight_measurements: the wire DTO always emits it")
		assert.Equal(t, "array", prop.Type)
		assert.Equal(t, "#/components/schemas/WeightMeasurementResponse", prop.Items.Ref,
			"weight_measurements items must reference the WeightMeasurementResponse component")
		assert.False(t, prop.Items.Nullable,
			"weight_measurements items must not be nullable")
		assert.True(t, prop.ReadOnly,
			"weight_measurements is a server-side projection; clients never send it")
		assert.False(t, prop.Nullable,
			"no measurements serialize as [], never null")
		assert.Contains(t, mr.Required, "weight_measurements",
			"json tag has no omitempty — the key is always present, so it belongs in required")
	})

	t.Run("WeightMeasurementResponse component pins the four-key shape", func(t *testing.T) {
		node, ok := spec.Components.Schemas["WeightMeasurementResponse"]
		require.True(t, ok, "missing WeightMeasurementResponse schema")

		var wm struct {
			Type       string   `yaml:"type"`
			Nullable   bool     `yaml:"nullable"`
			Required   []string `yaml:"required"`
			Properties map[string]struct {
				Type     string   `yaml:"type"`
				Format   string   `yaml:"format"`
				Nullable bool     `yaml:"nullable"`
				Enum     []string `yaml:"enum"`
			} `yaml:"properties"`
		}
		require.NoError(t, node.Decode(&wm))

		assert.Equal(t, "object", wm.Type)
		assert.False(t, wm.Nullable, "WeightMeasurementResponse component must not be nullable")
		assert.ElementsMatch(t, []string{"id", "value", "unit", "recorded_at"}, wm.Required,
			"all four fields are always emitted (no omitempty), so all belong in required")
		require.Len(t, wm.Properties, 4,
			"WeightMeasurementResponse must expose exactly id/value/unit/recorded_at")

		id, ok := wm.Properties["id"]
		require.True(t, ok)
		assert.Equal(t, "integer", id.Type)
		assert.Equal(t, "int64", id.Format, "id is uint64 on the wire")
		assert.False(t, id.Nullable)

		value, ok := wm.Properties["value"]
		require.True(t, ok)
		assert.Equal(t, "number", value.Type)
		assert.Equal(t, "double", value.Format, "value is float64 on the wire")
		assert.False(t, value.Nullable)

		unit, ok := wm.Properties["unit"]
		require.True(t, ok)
		assert.Equal(t, "string", unit.Type)
		assert.ElementsMatch(t, []string{"Kg", "g"}, unit.Enum,
			"unit must stay the model.BodyWeightUnit enum (BodyWeightUnitKg/BodyWeightUnitG)")
		assert.False(t, unit.Nullable)

		recordedAt, ok := wm.Properties["recorded_at"]
		require.True(t, ok)
		assert.Equal(t, "string", recordedAt.Type)
		assert.Equal(t, "date-time", recordedAt.Format,
			"recorded_at is time.Time on the wire (RFC3339), not a date-only string")
		assert.False(t, recordedAt.Nullable)
	})
}

func TestMedicalRecordWeightGeneratedTSContract(t *testing.T) {
	src, err := os.ReadFile(medicalRecordResponsesTSPath)
	require.NoError(t, err, "generated TS must exist at %s", medicalRecordResponsesTSPath)

	mr := extractTSInterfaceBody(t, string(src), "MedicalRecordResponse")
	var wmField string
	for _, f := range normalizeTSFields(mr) {
		if strings.HasPrefix(f, "weight_measurements") {
			wmField = f
		}
	}
	require.NotEmpty(t, wmField,
		"MedicalRecordResponse must declare weight_measurements (regenerate medicalrecord-responses.ts via tygo)")
	assert.Equal(t, "weight_measurements:WeightMeasurementResponse[]", wmField,
		"weight_measurements must be a required non-nullable array (no `?`, no `| null`): [] means no measurements")

	wm := extractTSInterfaceBody(t, string(src), "WeightMeasurementResponse")
	assert.ElementsMatch(t, []string{
		"id:number",
		"value:number",
		`unit:"Kg"|"g"`,
		"recorded_at:string",
	}, normalizeTSFields(wm),
		"WeightMeasurementResponse must be exactly id/value/unit/recorded_at, all required")
}

func TestMedicalRecordWeightWireDTOContract(t *testing.T) {
	weight := collectStructJSONFields(t, medicalRecordWeightResponsePath, "WeightMeasurementResponse")
	require.Len(t, weight, 4,
		"WeightMeasurementResponse must emit exactly id/value/unit/recorded_at")
	for name, field := range weight {
		assert.NotContains(t, field.opts, "omitempty",
			"omitempty on %s would drop the key and break the fixed four-key wire shape", name)
	}

	id, ok := weight["id"]
	require.True(t, ok)
	assert.True(t, isIdentType(id.expr, "uint64"), "id must stay uint64")

	value, ok := weight["value"]
	require.True(t, ok)
	assert.True(t, isIdentType(value.expr, "float64"), "value must stay float64")

	unit, ok := weight["unit"]
	require.True(t, ok)
	assert.True(t, isSelectorType(unit.expr, "model", "BodyWeightUnit"),
		"unit must stay model.BodyWeightUnit (Kg/g enum)")

	recordedAt, ok := weight["recorded_at"]
	require.True(t, ok)
	assert.True(t, isSelectorType(recordedAt.expr, "time", "Time"),
		"recorded_at must stay exactly time.Time (not *time.Time) on the wire (RFC3339 date-time)")

	mr := collectStructJSONFields(t, medicalRecordResponsePath, "MedicalRecordResponse")
	wm, ok := mr["weight_measurements"]
	require.True(t, ok,
		"MedicalRecordResponse must keep a field tagged json:\"weight_measurements\"")
	assert.NotContains(t, wm.opts, "omitempty",
		"omitempty would drop weight_measurements and break the always-emit contract")
	arr, isArr := wm.expr.(*ast.ArrayType)
	require.True(t, isArr, "weight_measurements must stay a slice ([] on the wire, never null)")
	require.Nil(t, arr.Len, "weight_measurements must be a slice, not a fixed-length array")
	assert.True(t, isIdentType(arr.Elt, "WeightMeasurementResponse"),
		"weight_measurements element type must stay WeightMeasurementResponse")
}
