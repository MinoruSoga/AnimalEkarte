package httpapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCamelToSnake は CamelCase → snake_case 変換を検証する。
// BUG-LINE-010: 連続した大文字（頭字語）を 1 単語として扱うこと。
func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"OwnerName", "owner_name"},
		{"IsDangerous", "is_dangerous"},
		{"TypeID", "type_id"},         // BUG-LINE-010: 以前は "type_i_d"
		{"OwnerID", "owner_id"},       // 同上
		{"HTTPServer", "http_server"}, // 頭字語の末尾で区切る
		{"APIKey", "api_key"},
		{"URL", "url"},
		{"ID", "id"},
		{"Name", "name"},
		{"", ""},
		{"A", "a"},
		{"lowercase", "lowercase"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := camelToSnake(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
