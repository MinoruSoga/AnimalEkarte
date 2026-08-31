package medicalrecord

import (
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type labDeviceOpenAPISchema struct {
	Required   []string                            `yaml:"required"`
	Properties map[string]labDeviceOpenAPIProperty `yaml:"properties"`
}

type labDeviceOpenAPIProperty struct {
	Type     string `yaml:"type"`
	Format   string `yaml:"format"`
	Nullable bool   `yaml:"nullable"`
	Items    struct {
		Ref string `yaml:"$ref"`
	} `yaml:"items"`
}

func TestLabDeviceOpenAPIResponseParity(t *testing.T) {
	src, err := os.ReadFile("../../docs/api.yaml")
	require.NoError(t, err)

	var spec struct {
		Components struct {
			Schemas map[string]labDeviceOpenAPISchema `yaml:"schemas"`
		} `yaml:"components"`
	}
	require.NoError(t, yaml.Unmarshal(src, &spec))

	contracts := []struct {
		name   string
		typeOf reflect.Type
	}{
		{"LabDeviceJobCard", reflect.TypeOf(labDeviceJobCardResponse{})},
		{"LabDeviceBoard", reflect.TypeOf(labDeviceBoardResponse{})},
		{"LabDeviceTodayVisit", reflect.TypeOf(labDeviceTodayVisitResponse{})},
	}
	for _, contract := range contracts {
		t.Run(contract.name, func(t *testing.T) {
			schema, ok := spec.Components.Schemas[contract.name]
			require.True(t, ok, "missing OpenAPI schema")

			properties, required, nullable := responseJSONContract(contract.typeOf)
			assert.Equal(t, properties, sortedKeys(schema.Properties), "OpenAPI properties differ from Go JSON fields")
			assert.Equal(t, required, sortedStrings(schema.Required), "OpenAPI required fields differ from non-omitempty Go JSON fields")
			for name, wantNullable := range nullable {
				assert.Equal(t, wantNullable, schema.Properties[name].Nullable, "%s nullable differs from Go JSON output", name)
			}
		})
	}

	card := spec.Components.Schemas["LabDeviceJobCard"]
	assert.Equal(t, "boolean", card.Properties["clock_skew"].Type)
	assert.Equal(t, "string", card.Properties["review_reason"].Type)

	board := spec.Components.Schemas["LabDeviceBoard"]
	assert.Equal(t, "array", board.Properties["received"].Type)
	assert.Equal(t, "#/components/schemas/LabDeviceJobCard", board.Properties["received"].Items.Ref)
	assert.Equal(t, "array", board.Properties["today_visits"].Type)
	assert.Equal(t, "#/components/schemas/LabDeviceTodayVisit", board.Properties["today_visits"].Items.Ref)

	visit := spec.Components.Schemas["LabDeviceTodayVisit"]
	assert.Equal(t, "integer", visit.Properties["record_id"].Type)
	assert.Equal(t, "int64", visit.Properties["record_id"].Format)
	assert.Equal(t, "boolean", visit.Properties["pet_is_deceased"].Type)
}

func TestLabDeviceOpenAPISourceEnumsMatchRuntime(t *testing.T) {
	src, err := os.ReadFile("../../docs/api.yaml")
	require.NoError(t, err)
	var document yaml.Node
	require.NoError(t, yaml.Unmarshal(src, &document))

	var enums [][]string
	collectLabDeviceSourceEnums(&document, &enums)
	require.Len(t, enums, 5, "expected every lab-device source enum in schemas and query")

	runtime := make([]string, 0, len(labDeviceDefaults()))
	for _, device := range labDeviceDefaults() {
		runtime = append(runtime, string(device.SourceType))
	}
	runtime = sortedStrings(runtime)
	for _, enum := range enums {
		assert.Equal(t, runtime, sortedStrings(enum))
	}
}

func collectLabDeviceSourceEnums(node *yaml.Node, enums *[][]string) {
	if node.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(node.Content); i += 2 {
			key, value := node.Content[i], node.Content[i+1]
			if key.Value == "enum" && value.Kind == yaml.SequenceNode {
				values := make([]string, 0, len(value.Content))
				containsLabSource := false
				for _, item := range value.Content {
					values = append(values, item.Value)
					containsLabSource = containsLabSource || item.Value == "fuji_nx600"
				}
				if containsLabSource {
					*enums = append(*enums, values)
				}
			}
			collectLabDeviceSourceEnums(value, enums)
		}
		return
	}
	for _, child := range node.Content {
		collectLabDeviceSourceEnums(child, enums)
	}
}

func responseJSONContract(responseType reflect.Type) (properties, required []string, nullable map[string]bool) {
	nullable = make(map[string]bool)
	for i := 0; i < responseType.NumField(); i++ {
		field := responseType.Field(i)
		tag := strings.Split(field.Tag.Get("json"), ",")
		if tag[0] == "" || tag[0] == "-" {
			continue
		}
		properties = append(properties, tag[0])
		omitEmpty := len(tag) > 1 && tag[1] == "omitempty"
		if !omitEmpty {
			required = append(required, tag[0])
		}
		nullable[tag[0]] = field.Type.Kind() == reflect.Pointer && !omitEmpty
	}
	return sortedStrings(properties), sortedStrings(required), nullable
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return sortedStrings(keys)
}

func sortedStrings(values []string) []string {
	sort.Strings(values)
	return values
}
