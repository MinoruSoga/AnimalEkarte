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
