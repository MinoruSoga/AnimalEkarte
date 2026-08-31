package medicalrecord

import (
	"encoding/json"
	"fmt"
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
		{"LabDeviceItemMaster", reflect.TypeOf(labDeviceItemMasterResponse{})},
		{"LabDeviceItemMasterEnsureResponse", reflect.TypeOf(labDeviceItemMasterEnsureResponse{})},
		{"LabDevice", reflect.TypeOf(labDeviceResponse{})},
		{"SaveLabDeviceConfigurationResponse", reflect.TypeOf(saveLabDeviceConfigurationResponse{})},
		{"LabDeviceJobItem", reflect.TypeOf(labDeviceJobItemResponse{})},
		{"LabDeviceJobCard", reflect.TypeOf(labDeviceJobCardResponse{})},
		{"LabDeviceWait", reflect.TypeOf(labDeviceWaitResponse{})},
		{"LabDeviceStation", reflect.TypeOf(labDeviceStationResponse{})},
		{"LabDeviceTodayVisit", reflect.TypeOf(labDeviceTodayVisitResponse{})},
		{"LabDeviceBoard", reflect.TypeOf(labDeviceBoardResponse{})},
		{"LabDeviceFramesResponse", reflect.TypeOf(labDeviceReceiveResponse{})},
		{"LabDeviceFrameResult", reflect.TypeOf(labDeviceFrameResultResponse{})},
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

func TestLabDeviceOpenAPIEnsureCountsMatchRuntime(t *testing.T) {
	src, err := os.ReadFile("../../docs/api.yaml")
	require.NoError(t, err)

	var spec struct {
		Paths map[string]struct {
			Post struct {
				Summary     string `yaml:"summary"`
				Description string `yaml:"description"`
			} `yaml:"post"`
		} `yaml:"paths"`
	}
	require.NoError(t, yaml.Unmarshal(src, &spec))

	operation, ok := spec.Paths["/lab-device-item-masters/ensure"]
	require.True(t, ok, "missing ensure endpoint")
	deviceCount := len(labDeviceDefaults())
	assert.Equal(t,
		fmt.Sprintf("観測カタログ%d行と既定機器%d行を医院へ冪等投入", LabDeviceItemCatalogCount, deviceCount),
		operation.Post.Summary,
	)
	assert.Equal(t,
		fmt.Sprintf("ResourceLabImport edit。既存行の exam_type_field_id は上書きしない。未知コードは追加しない。lab_devices の既定%d行も無ければ作る。", deviceCount),
		operation.Post.Description,
	)

	createOperation, ok := spec.Paths["/lab-devices"]
	require.True(t, ok, "missing create lab-device endpoint")
	sourceTypes := make(map[string]struct{}, deviceCount)
	for _, device := range labDeviceDefaults() {
		sourceTypes[string(device.SourceType)] = struct{}{}
	}
	assert.Equal(t,
		fmt.Sprintf("ResourceLabImport create。source_type は clinic 内で一意（対応プロトコルは%d種）。", len(sourceTypes)),
		createOperation.Post.Description,
	)
}

func TestLabDeviceConnectivityDocsMatchRuntime(t *testing.T) {
	doc, err := os.ReadFile("../../../docs/ops/deploy/LAB_DEVICE_CONNECTIVITY.md")
	require.NoError(t, err)

	var slots []struct {
		SourceType string `json:"source_type"`
	}
	require.NoError(t, json.Unmarshal([]byte(labDeviceDefaultSlotsJSON), &slots))
	text := string(doc)
	assert.Contains(t, text, fmt.Sprintf("%d種の `source_type`", len(labDeviceDefaults())))
	assert.Contains(t, text, fmt.Sprintf("既定slotは%d種", len(slots)))
	assert.Contains(t, text, "`idexx_vetlab` は decoder / persist / 既定slotを実装済み")
	assert.NotContains(t, text, "別 source_type。未実装")
	assert.NotContains(t, text, "既定3スロットに IDEXX フレームを足す")

	adr, err := os.ReadFile("../../../docs/architecture/adr/007-lab-device-receive-and-commit.md")
	require.NoError(t, err)
	adrText := string(adr)
	assert.NotContains(t, adrText, "に3種を allowlist")
	assert.NotContains(t, adrText, "Commit は3種を受けない")
	assert.NotContains(t, adrText, "### 7. マスタと 1測定=1 exam")
	assert.Contains(t, adrText, "`device_hint` | `NX600` / `AU10V` / `PU-4010` / `VetLab`")

	task, err := os.ReadFile("../../../.hex-skills/tasks/T001-vetlab-multi-exam-persist.md")
	require.NoError(t, err)
	taskText := string(task)
	assert.Contains(t, taskText, "**status**: Completed")
	assert.NotContains(t, taskText, "- [ ]")
	assert.NotContains(t, taskText, "現状（ADR-007 §7 の実装）は「`exam_type_id` が2種以上なら保存拒否")
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
