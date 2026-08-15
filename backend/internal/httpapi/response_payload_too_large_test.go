package httpapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestRespondError_PayloadTooLargeMapsTo413(t *testing.T) {
	c, recorder := newTestContext()

	RespondError(c, apperrors.WrapPayloadTooLarge("request body exceeds size limit"))

	assert.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "request body exceeds size limit")
}
