package apicontract

// openapi_support_chat_contract_test.go — EMR-410: POST /support/chat の
// request schema(oneOf新/レガシー)・required fields・enum・response codes を
// YAML parse して検証する（文字列 grep ではない）。

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type openAPISpec struct {
	Paths      map[string]any `yaml:"paths"`
	Components struct {
		Schemas map[string]map[string]any `yaml:"schemas"`
	} `yaml:"components"`
}

func loadSupportChatSpec(t *testing.T) openAPISpec {
	t.Helper()
	raw, err := os.ReadFile("../../docs/api.yaml")
	require.NoError(t, err)
	var spec openAPISpec
	require.NoError(t, yaml.Unmarshal(raw, &spec))
	return spec
}

func nav(node any, keys ...string) any {
	cur := node
	for _, k := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = m[k]
	}
	return cur
}

func TestOpenAPISupportChat(t *testing.T) {
	spec := loadSupportChatSpec(t)
	post := nav(spec.Paths, "/support/chat", "post")
	require.NotNil(t, post, "POST /support/chat must be documented")

	// request body: oneOf of the two named schemas
	oneOf := nav(post, "requestBody", "content", "application/json", "schema", "oneOf").([]any)
	require.Len(t, oneOf, 2)
	refs := []string{nav(oneOf[0], "$ref").(string), nav(oneOf[1], "$ref").(string)}
	assert.ElementsMatch(t, []string{
		"#/components/schemas/SupportChatRequest",
		"#/components/schemas/LegacySupportChatRequest",
	}, refs)

	// new protocol required + constraints
	newReq := spec.Components.Schemas["SupportChatRequest"]
	assert.ElementsMatch(t, []string{"message", "manual_revision", "keyword_candidates"}, newReq["required"])
	assert.Equal(t, "^[0-9a-f]{64}$", nav(newReq, "properties", "manual_revision", "pattern"))
	assert.Equal(t, 20, nav(newReq, "properties", "keyword_candidates", "maxItems"))
	assert.ElementsMatch(t, []any{"screens", "workflows"}, nav(newReq, "properties", "keyword_candidates", "items", "properties", "category", "enum"))

	// legacy: deprecated + title/text marked ignored
	legacy := spec.Components.Schemas["LegacySupportChatRequest"]
	assert.Equal(t, true, legacy["deprecated"])
	assert.ElementsMatch(t, []string{"message", "context"}, legacy["required"])
	assert.Contains(t, nav(legacy, "properties", "context", "items", "properties", "title", "description"), "ignored")
	assert.Contains(t, nav(legacy, "properties", "context", "items", "properties", "text", "description"), "ignored")

	// response codes
	responses := nav(post, "responses").(map[string]any)
	for _, code := range []string{"200", "400", "401", "409", "413", "429", "500", "501", "502", "503", "504"} {
		assert.Contains(t, responses, code)
	}

	// 200 schema: required reply/sources/retrieval + enums + revision pattern
	ok200 := nav(post, "responses", "200", "content", "application/json", "schema")
	assert.ElementsMatch(t, []string{"reply", "sources", "retrieval"}, ok200.(map[string]any)["required"])
	retrieval := nav(ok200, "properties", "retrieval")
	assert.ElementsMatch(t, []string{"mode", "reason", "manual_revision"}, retrieval.(map[string]any)["required"])
	assert.ElementsMatch(t, []any{"hybrid", "keyword", "none"}, nav(retrieval, "properties", "mode", "enum"))
	assert.ElementsMatch(t, []any{"ok", "disabled", "index_missing", "embedding_unavailable", "capacity", "no_evidence"}, nav(retrieval, "properties", "reason", "enum"))
	assert.Equal(t, "^[0-9a-f]{64}$", nav(retrieval, "properties", "manual_revision", "pattern"))
	// oneOf 曖昧性排除: 新プロトコルは context を持たない
	assert.Equal(t, []any{"context"}, nav(newReq, "not", "required"))
	// slug パターンは BE と一致
	assert.Equal(t, "^[A-Za-z0-9][A-Za-z0-9_-]{0,99}$", nav(newReq, "properties", "keyword_candidates", "items", "properties", "slug", "pattern"))
	assert.Equal(t, true, nav(newReq, "properties", "keyword_candidates", "uniqueItems"))
}
