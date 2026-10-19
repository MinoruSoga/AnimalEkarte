package support

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchTermsSegmentTailMerge(t *testing.T) {
	// 末尾 1 コードユニットの断片は直前パートへマージ（JS と同仕様）。
	got := segmentJaTerm("新しい飼い主X")
	assert.Equal(t, []string{"新しい", "飼い主X"}, got)
	got = segmentJaTerm("新しい飼い主😀")
	assert.Equal(t, []string{"新しい", "飼い主", "😀"}, got)
}

func TestSearchTermsStopwordOnlyNoGroups(t *testing.T) {
	assert.Empty(t, extractSearchTermGroups("方法"))
	assert.Empty(t, extractSearchTermGroups("どうやって"))
}

// EMR-441: サロゲートペア1文字 (𠮷) は JS length 契約 (UTF-16=2) で 2 文字扱い。
// rune 数で棄却すると検索グループが空になり「該当なし」誤応答になる。
func TestSearchTermsKeepsSupplementaryPlaneTerm(t *testing.T) {
	got := extractSearchTermGroups("𠮷")
	assert.NotEmpty(t, got)
	for _, group := range got {
		for _, term := range group {
			assert.Equal(t, "𠮷", term)
		}
	}
}
