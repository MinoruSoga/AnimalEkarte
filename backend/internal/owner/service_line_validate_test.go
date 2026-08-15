package owner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateLineUserID(t *testing.T) {
	assert.NoError(t, validateLineUserID("Uabcdef123"))
	assert.Error(t, validateLineUserID(""))
	assert.Error(t, validateLineUserID("../evil"))
	assert.Error(t, validateLineUserID("id?x=1"))
	assert.Error(t, validateLineUserID(string(make([]byte, 65))))
}
