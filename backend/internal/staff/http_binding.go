package staff

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

const (
	staffJSONBodyMaxBytes        int64 = 64 * 1024
	staffJSONBodyTooLargeMessage       = "request body exceeds size limit"
	staffJSONInvalidMessage            = "invalid request body"
)

// bindStaffJSON consumes the complete request within a fixed limit before
// binding. This rejects both chunked oversized bodies and additional JSON
// values that ShouldBindJSON alone could leave unread.
func bindStaffJSON(c *gin.Context, destination any) error {
	if c.Request.ContentLength > staffJSONBodyMaxBytes {
		return apperrors.WrapPayloadTooLarge(staffJSONBodyTooLargeMessage)
	}

	boundedBody := http.MaxBytesReader(c.Writer, c.Request.Body, staffJSONBodyMaxBytes)
	defer func() {
		_ = boundedBody.Close()
	}()

	body, err := io.ReadAll(boundedBody)
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return apperrors.WrapPayloadTooLarge(staffJSONBodyTooLargeMessage)
		}
		return apperrors.WrapInvalidInput(staffJSONInvalidMessage)
	}
	if hasTrailingJSONValue(body) {
		return apperrors.WrapInvalidInput(staffJSONInvalidMessage)
	}

	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if err := c.ShouldBindJSON(destination); err != nil {
		return apperrors.WrapInvalidInput(httpapi.ParseBindError(err))
	}
	return validateStaffJSONByteLengths(destination)
}

func hasTrailingJSONValue(body []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(body))
	var firstValue json.RawMessage
	if err := decoder.Decode(&firstValue); err != nil {
		return false
	}
	var trailingValue json.RawMessage
	return decoder.Decode(&trailingValue) != io.EOF
}

func validateStaffJSONByteLengths(destination any) error {
	switch request := destination.(type) {
	case *createStaffRequest:
		if len([]byte(request.Email)) > 254 {
			return apperrors.WrapInvalidInput("email must be at most 254 bytes")
		}
		if len([]byte(request.Password)) > 72 {
			return apperrors.WrapInvalidInput("password must be at most 72 bytes")
		}
	case *updateStaffRequest:
		if request.Password != nil && len([]byte(*request.Password)) > 72 {
			return apperrors.WrapInvalidInput("password must be at most 72 bytes")
		}
	}
	return nil
}
