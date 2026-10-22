package medicalrecord

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// EMR-448: 非空の備考を空へ変更する PATCH は "" をそのまま update map へ渡し、省略は key を作らない。
func TestBuildCageUpdate_EmptyDescriptionIsWritten(t *testing.T) {
	empty := ""
	assert.Equal(t, map[string]any{"description": ""}, buildCageUpdate(&UpdateCageInput{Description: &empty}))
	assert.Empty(t, buildCageUpdate(&UpdateCageInput{}))
}
