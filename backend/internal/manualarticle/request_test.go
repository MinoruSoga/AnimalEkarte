package manualarticle

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

const validEditBody = `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"42","expected_version":3,"expected_content_version_id":"7","content":{"title":"タイトル","order_value":1.5,"section":"基本","body_markdown":"# 本文"}}`

func TestParseEditRequest_Valid(t *testing.T) {
	cmd, err := parseEditRequest([]byte(validEditBody), model.ManualCategoryScreens, "01-login")
	require.NoError(t, err)
	assert.Equal(t, uint64(42), cmd.ArticleID)
	assert.Equal(t, "42", cmd.ArticleIDText)
	assert.Equal(t, int64(3), cmd.ExpectedVersion)
	assert.Equal(t, uint64(7), cmd.ExpectedContentVersionID)
	assert.Equal(t, "タイトル", cmd.Title)
	assert.Equal(t, "1.5", cmd.OrderText)
	assert.Equal(t, "基本", cmd.Section)
	assert.Equal(t, "# 本文", cmd.BodyMarkdown)
	assert.Len(t, cmd.ContentDigestHex, 64)
	assert.Len(t, cmd.CommandDigestHex, 64)
}

func TestParseEditRequest_DigestDeterministic(t *testing.T) {
	a, err := parseEditRequest([]byte(validEditBody), model.ManualCategoryScreens, "01-login")
	require.NoError(t, err)
	b, err := parseEditRequest([]byte(validEditBody), model.ManualCategoryScreens, "01-login")
	require.NoError(t, err)
	assert.Equal(t, a.ContentDigestHex, b.ContentDigestHex)
	assert.Equal(t, a.CommandDigestHex, b.CommandDigestHex)
	// 別 slug は command_digest を変える（経路 selector 束縛）。
	c, err := parseEditRequest([]byte(validEditBody), model.ManualCategoryScreens, "other-slug")
	require.NoError(t, err)
	assert.NotEqual(t, a.CommandDigestHex, c.CommandDigestHex)
}

func TestParseEditRequest_Rejects(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"missing member", `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"42","expected_version":3,"expected_content_version_id":"7"}`},
		{"null member", `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":null,"expected_version":3,"expected_content_version_id":"7","content":{"title":"t","order_value":1,"section":"s","body_markdown":"b"}}`},
		{"extra member", `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"42","expected_version":3,"expected_content_version_id":"7","content":{"title":"t","order_value":1,"section":"s","body_markdown":"b"},"extra":1}`},
		{"duplicate member", `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"42","article_id":"43","expected_version":3,"expected_content_version_id":"7","content":{"title":"t","order_value":1,"section":"s","body_markdown":"b"}}`},
		{"empty title", `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"42","expected_version":3,"expected_content_version_id":"7","content":{"title":"","order_value":1,"section":"s","body_markdown":"b"}}`},
		{"empty section", `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"42","expected_version":3,"expected_content_version_id":"7","content":{"title":"t","order_value":1,"section":"","body_markdown":"b"}}`},
		{"order too precise", `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"42","expected_version":3,"expected_content_version_id":"7","content":{"title":"t","order_value":1.234,"section":"s","body_markdown":"b"}}`},
		{"order out of range", `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"42","expected_version":3,"expected_content_version_id":"7","content":{"title":"t","order_value":999999999.99,"section":"s","body_markdown":"b"}}`},
		{"article_id zero", `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"0","expected_version":3,"expected_content_version_id":"7","content":{"title":"t","order_value":1,"section":"s","body_markdown":"b"}}`},
		{"article_id number not string", `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":42,"expected_version":3,"expected_content_version_id":"7","content":{"title":"t","order_value":1,"section":"s","body_markdown":"b"}}`},
		{"expected_version zero", `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"42","expected_version":0,"expected_content_version_id":"7","content":{"title":"t","order_value":1,"section":"s","body_markdown":"b"}}`},
		{"bad operation key", `{"operation_key":"not-a-uuid","article_id":"42","expected_version":3,"expected_content_version_id":"7","content":{"title":"t","order_value":1,"section":"s","body_markdown":"b"}}`},
		{"uppercase uuid", `{"operation_key":"AAAAAAAA-1111-4111-8111-111111111111","article_id":"42","expected_version":3,"expected_content_version_id":"7","content":{"title":"t","order_value":1,"section":"s","body_markdown":"b"}}`},
		{"not json", `not-json`},
		{"lone surrogate escape", `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"42","expected_version":3,"expected_content_version_id":"7","content":{"title":"t\uD800x","order_value":1,"section":"s","body_markdown":"b"}}`},
		{"quoted expected_version", `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"42","expected_version":"3","expected_content_version_id":"7","content":{"title":"t","order_value":1,"section":"s","body_markdown":"b"}}`},
		{"quoted order_value", `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"42","expected_version":3,"expected_content_version_id":"7","content":{"title":"t","order_value":"1.25","section":"s","body_markdown":"b"}}`},
		{"trailing brace", `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"42","expected_version":3,"expected_content_version_id":"7","content":{"title":"t","order_value":1,"section":"s","body_markdown":"b"}}}`},
		{"decoded NUL in title", `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"42","expected_version":3,"expected_content_version_id":"7","content":{"title":"t\u0000x","order_value":1,"section":"s","body_markdown":"b"}}`},
		{"decoded NUL in section", `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"42","expected_version":3,"expected_content_version_id":"7","content":{"title":"t","order_value":1,"section":"\u0000","body_markdown":"b"}}`},
		{"decoded NUL in body", `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"42","expected_version":3,"expected_content_version_id":"7","content":{"title":"t","order_value":1,"section":"s","body_markdown":"b\u0000"}}`},
		{"trailing bracket", `[{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"42","expected_version":3,"expected_content_version_id":"7","content":{"title":"t","order_value":1,"section":"s","body_markdown":"b"}]]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseEditRequest([]byte(tt.body), model.ManualCategoryScreens, "s")
			require.Error(t, err)
		})
	}
}

func TestParseEditRequest_InvalidUTF8(t *testing.T) {
	_, err := parseEditRequest([]byte{0xff, 0xfe}, model.ManualCategoryScreens, "s")
	require.Error(t, err)
}

func TestParseEditRequest_EmptyBodyAllowed(t *testing.T) {
	body := `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"1","expected_version":1,"expected_content_version_id":"2","content":{"title":"t","order_value":0,"section":"s","body_markdown":""}}`
	cmd, err := parseEditRequest([]byte(body), model.ManualCategoryScreens, "s")
	require.NoError(t, err)
	assert.Equal(t, "", cmd.BodyMarkdown)
}

func TestParseEditRequest_ExactStringsPreserved(t *testing.T) {
	// trim・NFC・改行正規化はしない: 末尾空白・全角・\r\n をそのまま保持する。
	body := `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"1","expected_version":1,"expected_content_version_id":"2","content":{"title":"t ","order_value":0,"section":"s","body_markdown":"a\r\nb é"}}`
	cmd, err := parseEditRequest([]byte(body), model.ManualCategoryScreens, "s")
	require.NoError(t, err)
	assert.Equal(t, "t ", cmd.Title)
	assert.True(t, strings.Contains(cmd.BodyMarkdown, "\r\n"))
}

func TestParseEditRequest_LiteralBackslashU0000Accepted(t *testing.T) {
	// 6文字の literal `\u0000`（バックスラッシュ+u0000）は decoded U+0000
	// と区別して受理・そのまま保存/hash する。
	body := `{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"42","expected_version":3,"expected_content_version_id":"7","content":{"title":"t\\u0000x","order_value":1,"section":"s","body_markdown":"b"}}`
	cmd, err := parseEditRequest([]byte(body), model.ManualCategoryScreens, "s")
	require.NoError(t, err)
	assert.Equal(t, `t\u0000x`, cmd.Title)
}

func editBodyWith(t *testing.T, orderValue string, title, section, bodyMarkdown string) []byte {
	t.Helper()
	// digest の検査ではなく validation のみを見るため既知の固定値を組み立てる。
	return []byte(`{"operation_key":"11111111-1111-4111-8111-111111111111","article_id":"1","expected_version":1,"expected_content_version_id":"2","content":{"title":` +
		strconv.Quote(title) + `,"order_value":` + orderValue + `,"section":` + strconv.Quote(section) +
		`,"body_markdown":` + strconv.Quote(bodyMarkdown) + `}}`)
}

func TestParseEditRequest_OrderValueBoundaries(t *testing.T) {
	// 丸め・許容誤差はしない: JSON 数値を最短10進に写した小数部が2桁を超えると拒否。
	tests := []struct {
		name      string
		order     string
		wantErr   bool
		wantOrder string
	}{
		{"1.005 reject (exact 3dp)", "1.005", true, ""},
		{"1.004999999 reject", "1.004999999", true, ""},
		{"negative zero normalizes to 0", "-0", false, "0"},
		{"10.50 keeps one trailing digit", "10.50", false, "10.5"},
		{"max in range", "99999999.99", false, "99999999.99"},
		{"above range rejects", "100000000", true, ""},
		{"integer", "7", false, "7"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := parseEditRequest(editBodyWith(t, tt.order, "t", "s", "b"), model.ManualCategoryScreens, "s")
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantOrder, cmd.OrderText)
		})
	}
}

func TestParseEditRequest_LengthBoundaries(t *testing.T) {
	// 公開 API 契約 (ManualArticleEditContent) は文字数上限を課さない:
	// 旧 binding の max=255 / max=100000 は撤回済み。長い値は受理する。
	long255 := strings.Repeat("a", 255)
	long256 := strings.Repeat("a", 256)
	body100k := strings.Repeat("b", 100000)
	body100kPlus := strings.Repeat("b", 100001)

	tests := []struct {
		name    string
		title   string
		section string
		body    string
		wantErr bool
	}{
		{"title 255 ok", long255, "s", "b", false},
		{"title 256 accepted", long256, "s", "b", false},
		{"section 255 ok", "t", long255, "b", false},
		{"section 256 accepted", "t", long256, "b", false},
		{"body 100000 ok", "t", "s", body100k, false},
		{"body 100001 accepted", "t", "s", body100kPlus, false},
		{"empty title rejected", "", "s", "b", true},
		{"empty section rejected", "t", "", "b", true},
		{"empty body accepted", "t", "s", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseEditRequest(editBodyWith(t, "0", tt.title, tt.section, tt.body), model.ManualCategoryScreens, "s")
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
