package trimming

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// EMR-448: トリミングコース/オプションの備考 "" は空で保存し、省略は既存値を保持する。
func TestBuildTrimmingUpdate_EmptyDescriptionIsWritten(t *testing.T) {
	empty := ""
	assert.Equal(t, map[string]any{"description": ""}, buildTrimmingCourseUpdate(&UpdateTrimmingCourseInput{Description: &empty}))
	assert.Equal(t, map[string]any{"description": ""}, buildTrimmingOptionUpdate(&UpdateTrimmingOptionInput{Description: &empty}))
	assert.Empty(t, buildTrimmingCourseUpdate(&UpdateTrimmingCourseInput{}))
	assert.Empty(t, buildTrimmingOptionUpdate(&UpdateTrimmingOptionInput{}))
}
