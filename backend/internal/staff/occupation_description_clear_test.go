package staff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// EMR-448: 職種の備考 "" は空で保存し、省略は既存値を保持する。職種分類や予約 gate は変えない。
func TestBuildOccupationUpdate_EmptyDescriptionIsWritten(t *testing.T) {
	empty := ""
	assert.Equal(t, map[string]any{"description": ""}, buildOccupationUpdate(&UpdateOccupationInput{Description: &empty}))
	assert.Empty(t, buildOccupationUpdate(&UpdateOccupationInput{}))
}
