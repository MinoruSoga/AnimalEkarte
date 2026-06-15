package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeKana(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// ひらがな検索 → カタカナ DB 値がヒットすることを保証するケース
		{"ぴ で ピーター がヒット (ひらがな→変換なし)", "ぴ", "ぴ"},
		{"ピ で ピーター がヒット (カタカナ→ひらがな)", "ピ", "ぴ"},
		{"ぴーたー で ピーター がヒット", "ぴーたー", "ぴーたー"},
		{"ピーター で ピーター がヒット", "ピーター", "ぴーたー"},
		// 双方向等価
		{"ひらがな→変換なし", "あいうえお", "あいうえお"},
		{"カタカナ→ひらがな", "アイウエオ", "あいうえお"},
		// 混在
		{"カタカナひらがな混在", "ポチたろう", "ぽちたろう"},
		{"漢字・記号は変換しない", "山田ハナコ", "山田はなこ"},
		// 小文字カタカナ
		{"小文字カタカナ ァィゥェォ", "ァィゥェォ", "ぁぃぅぇぉ"},
		// 範囲外の文字
		{"ASCII は変換しない", "abc123", "abc123"},
		{"記号は変換しない", "!@#%_\\", "!@#%_\\"},
		// 特殊カタカナ: ヴヵヶ (U+30F4〜U+30F6)
		{"ヴ→ゔ", "ヴ", "ゔ"},
		{"ヵヶ→ゕゖ", "ヵヶ", "ゕゖ"},
		// 空文字
		{"空文字", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeKana(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
