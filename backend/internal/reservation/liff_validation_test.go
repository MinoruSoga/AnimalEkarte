package reservation

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// TestValidateLiffReservationInput は BUG-LINE-012 の入力サイズ制限を検証する。
func TestValidateLiffReservationInput(t *testing.T) {
	tests := []struct {
		name      string
		build     func() liffCreateReservationRequest
		wantErr   bool
		errSubstr string
	}{
		{
			name: "正常系: 空の customer_fields",
			build: func() liffCreateReservationRequest {
				return liffCreateReservationRequest{}
			},
			wantErr: false,
		},
		{
			name: "正常系: 通常の customer_fields",
			build: func() liffCreateReservationRequest {
				cf := map[string]any{
					"customer_name": "山田 太郎",
					"phone":         "09012345678",
					"pets":          []any{map[string]any{"name": "ポチ", "type": "犬"}},
				}
				b, _ := json.Marshal(cf)
				return liffCreateReservationRequest{
					CustomerFields: b,
					RequestText:    "よろしくお願いします",
				}
			},
			wantErr: false,
		},
		{
			name: "request_text が長過ぎる → 400",
			build: func() liffCreateReservationRequest {
				return liffCreateReservationRequest{
					RequestText: strings.Repeat("a", maxRequestTextLength+1),
				}
			},
			wantErr:   true,
			errSubstr: "request_text",
		},
		{
			name: "customer_fields 全体サイズ超過 → 400",
			build: func() liffCreateReservationRequest {
				// 10KB 超の JSON を生成
				huge := strings.Repeat("a", maxCustomerFieldsSize+100)
				return liffCreateReservationRequest{
					CustomerFields: json.RawMessage(`"` + huge + `"`),
				}
			},
			wantErr:   true,
			errSubstr: "サイズが大きすぎます",
		},
		{
			name: "customer_fields が JSON オブジェクトでない（配列） → 400",
			build: func() liffCreateReservationRequest {
				return liffCreateReservationRequest{
					CustomerFields: json.RawMessage(`[1,2,3]`),
				}
			},
			wantErr:   true,
			errSubstr: "JSON オブジェクト",
		},
		{
			name: "customer_fields のキー数超過 → 400",
			build: func() liffCreateReservationRequest {
				m := make(map[string]any)
				for i := 0; i <= maxCustomerFieldsKeys; i++ {
					m[strings.Repeat("k", 1)+string(rune('a'+i%26))+string(rune('0'+i%10))] = "v"
					if len(m) > maxCustomerFieldsKeys {
						break
					}
				}
				// 念のため強制的に 21 個入れる
				for i := len(m); i <= maxCustomerFieldsKeys; i++ {
					m["extra_key_"+string(rune('a'+i))] = "v"
				}
				b, _ := json.Marshal(m)
				return liffCreateReservationRequest{CustomerFields: b}
			},
			wantErr:   true,
			errSubstr: "キー数",
		},
		{
			name: "個別 string 値が長過ぎる → 400",
			build: func() liffCreateReservationRequest {
				cf := map[string]any{"customer_name": strings.Repeat("a", maxCustomerFieldValueLength+1)}
				b, _ := json.Marshal(cf)
				return liffCreateReservationRequest{CustomerFields: b}
			},
			wantErr:   true,
			errSubstr: "customer_name",
		},
		{
			name: "ネスト配列内の string も長さ検証される → 400",
			build: func() liffCreateReservationRequest {
				cf := map[string]any{
					"pets": []any{
						map[string]any{"name": strings.Repeat("a", maxCustomerFieldValueLength+1)},
					},
				}
				b, _ := json.Marshal(cf)
				return liffCreateReservationRequest{CustomerFields: b}
			},
			wantErr:   true,
			errSubstr: "pets[0].name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.build()
			err := validateLiffReservationInput(&req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err), "expected invalid input error, got %v", err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
