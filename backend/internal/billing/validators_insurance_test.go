package billing

// validators_insurance_test.go — BE9-2C B①: service/validators_test.go から補償率検証テストを
// 実装（validators_insurance.go）と同 package へ移動。

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCoverageRate(t *testing.T) {
	assert.NoError(t, ValidateCoverageRate(50))
	assert.Error(t, ValidateCoverageRate(-1))
	assert.Error(t, ValidateCoverageRate(101))
}

func TestValidateOptionalCoverageRate(t *testing.T) {
	assert.NoError(t, ValidateOptionalCoverageRate(nil))
	val := 50
	assert.NoError(t, ValidateOptionalCoverageRate(&val))
}
