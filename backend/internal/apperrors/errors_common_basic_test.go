package apperrors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// EMR-491: 共通基本情報の 422 コードは wire 値として固定する。
func TestCommonBasicCodes(t *testing.T) {
	assert.Equal(t, "common_basic_unresolved", CodeCommonBasicUnresolved)
	assert.Equal(t, "owner_basic_revision_pending", CodeOwnerBasicRevisionPending)
	assert.Equal(t, "owner_registration_pending", CodeOwnerRegistrationPending)
	assert.Equal(t, "pet_basic_revision_pending", CodePetBasicRevisionPending)
	assert.Equal(t, "pet_lifecycle_write_pending", CodePetLifecycleWritePending)
	assert.Equal(t, "pet_registration_pending", CodePetRegistrationPending)
}

func TestWrapCommonBasicUnresolved(t *testing.T) {
	err := WrapCommonBasicUnresolved("owner 1/2: no common clinic binding")

	assert.True(t, HasCode(err, CodeCommonBasicUnresolved))
	assert.True(t, errors.Is(err, ErrUnprocessable))
	assert.False(t, errors.Is(err, ErrConflict))
	assert.False(t, errors.Is(err, ErrInvalidInput))
	assert.True(t, HasCode(Wrap(err, "outer"), CodeCommonBasicUnresolved), "found through a wrapping error chain")

	var appErr *AppError
	require.True(t, errors.As(err, &appErr))
	assert.NotContains(t, appErr.Message, "owner 1/2", "response message must not carry the detail")
	assert.Contains(t, err.Error(), "owner 1/2: no common clinic binding", "detail stays available for server logs")
}
