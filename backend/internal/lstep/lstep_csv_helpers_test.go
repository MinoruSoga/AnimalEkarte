package lstep

import (
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func encodeShiftJIS(t *testing.T, s string) []byte {
	t.Helper()
	encoder := japanese.ShiftJIS.NewEncoder()
	out, err := io.ReadAll(transform.NewReader(strings.NewReader(s), encoder))
	if err != nil {
		t.Fatalf("encode SJIS failed: %v", err)
	}
	return out
}

func TestResolveCsvHeaders(t *testing.T) {
	cases := []struct {
		name     string
		headers  []string
		wantKeys []string
		wantErr  bool
	}{
		{
			name:     "必須 line_user_id (英) のみ",
			headers:  []string{"LINE ID"},
			wantKeys: []string{"line_user_id"},
		},
		{
			name:     "大文字小文字無視",
			headers:  []string{"line id"},
			wantKeys: []string{"line_user_id"},
		},
		{
			name:     "前後空白除去",
			headers:  []string{"  LINE ID  ", "  表示名 "},
			wantKeys: []string{"line_user_id", "display_name"},
		},
		{
			name:     "全カラム alias 解決",
			headers:  []string{"LINE ID", "表示名", "登録日", "タグ", "シナリオ", "流入元", "ブロック", "最終メッセージ"},
			wantKeys: []string{"line_user_id", "display_name", "registered_at", "tags", "scenarios", "traffic_source", "block_status", "last_message_at"},
		},
		{
			name:    "必須 line_user_id 欠落",
			headers: []string{"表示名", "タグ"},
			wantErr: true,
		},
		{
			name:    "空ヘッダー",
			headers: []string{},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveCsvHeaders(tc.headers)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			for _, key := range tc.wantKeys {
				if _, ok := got[key]; !ok {
					t.Errorf("key %q not found in resolved map", key)
				}
			}
		})
	}
}

func TestSplitMultiValue(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []string
	}{
		{"空文字", "", nil},
		{"単値", "tag1", []string{"tag1"}},
		{"カンマ区切り", "tag1,tag2,tag3", []string{"tag1", "tag2", "tag3"}},
		{"セミコロン区切り", "tag1;tag2;tag3", []string{"tag1", "tag2", "tag3"}},
		{"前後空白除去", " tag1 , tag2 ", []string{"tag1", "tag2"}},
		{"空要素除去", "tag1,,tag2", []string{"tag1", "tag2"}},
		{"混在 (, と ;)", "tag1;tag2,tag3", []string{"tag1", "tag2", "tag3"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := splitMultiValue(tc.input)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseCsvTime(t *testing.T) {
	mustParseTime := func(layout, value string) *time.Time {
		parsed, err := time.ParseInLocation(layout, value, time.Local)
		if err != nil {
			panic(err)
		}
		return &parsed
	}

	cases := []struct {
		name  string
		input string
		want  *time.Time
	}{
		{"空文字", "", nil},
		{"空白のみ", "   ", nil},
		{"不正フォーマット", "not-a-date", nil},
		{"YYYY-MM-DD", "2026-05-01", mustParseTime("2006-01-02", "2026-05-01")},
		{"YYYY/MM/DD", "2026/05/01", mustParseTime("2006/01/02", "2026/05/01")},
		{"YYYY-MM-DD HH:MM:SS", "2026-05-01 12:30:00", mustParseTime("2006-01-02 15:04:05", "2026-05-01 12:30:00")},
		{"YYYY/MM/DD HH:MM:SS", "2026/05/01 12:30:00", mustParseTime("2006/01/02 15:04:05", "2026/05/01 12:30:00")},
		{"RFC3339 without zone", "2026-05-01T12:30:00", mustParseTime("2006-01-02T15:04:05", "2026-05-01T12:30:00")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseCsvTime(tc.input)
			if tc.want == nil {
				if got != nil {
					t.Errorf("got %v, want nil", *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("got nil, want %v", *tc.want)
				return
			}
			if !got.Equal(*tc.want) {
				t.Errorf("got %v, want %v", *got, *tc.want)
			}
		})
	}
}
