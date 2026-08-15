package apperrors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapPayloadTooLarge(t *testing.T) {
	err := WrapPayloadTooLarge("request body exceeds size limit")

	require.Error(t, err)
	assert.True(t, IsPayloadTooLarge(err))
	assert.ErrorIs(t, err, ErrPayloadTooLarge)

	var appErr *AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "PAYLOAD_TOO_LARGE", appErr.Code)
}
