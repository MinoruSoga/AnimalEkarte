package auth

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
	authJSONBodyMaxBytes        int64 = 16 * 1024
	authJSONBodyTooLargeMessage       = "request body exceeds size limit"
	authJSONInvalidMessage            = "invalid request body"
)

// bindAuthJSON reads the complete bounded request before decoding. Reading to
// EOF is required because a decoder can otherwise accept a valid first JSON
// value without observing oversized or additional trailing input.
func bindAuthJSON(c *gin.Context, destination any) error {
	if c.Request.ContentLength > authJSONBodyMaxBytes {
		return apperrors.WrapPayloadTooLarge(authJSONBodyTooLargeMessage)
	}

	boundedBody := http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		authJSONBodyMaxBytes,
	)
	defer func() {
		_ = boundedBody.Close()
	}()
	body, err := io.ReadAll(boundedBody)
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return apperrors.WrapPayloadTooLarge(authJSONBodyTooLargeMessage)
		}
		return apperrors.WrapInvalidInput(authJSONInvalidMessage)
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	var firstValue json.RawMessage
	if err := decoder.Decode(&firstValue); err == nil {
		var trailingValue json.RawMessage
		if trailingErr := decoder.Decode(&trailingValue); trailingErr != io.EOF {
			return apperrors.WrapInvalidInput(authJSONInvalidMessage)
		}
	}

	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if err := c.ShouldBindJSON(destination); err != nil {
		return apperrors.WrapInvalidInput(httpapi.ParseBindError(err))
	}
	return validateAuthJSONByteLengths(destination)
}

func validateAuthJSONByteLengths(destination any) error {
	switch request := destination.(type) {
	case *LoginInput:
		if len([]byte(request.Email)) > 254 {
			return apperrors.WrapInvalidInput("メールアドレスは254バイト以下で入力してください")
		}
		if len([]byte(request.Password)) > 72 {
			return apperrors.WrapInvalidInput("パスワードは72 バイト以下で入力してください")
		}
	case *ForgotPasswordRequest:
		if len([]byte(request.Email)) > 254 {
			return apperrors.WrapInvalidInput("メールアドレスは254バイト以下で入力してください")
		}
	case *ResetPasswordRequest:
		if len([]byte(request.Token)) > 128 {
			return apperrors.WrapInvalidInput("トークンは128バイト以下で入力してください")
		}
		if len([]byte(request.Password)) > 72 {
			return apperrors.WrapInvalidInput("パスワードは72 バイト以下で入力してください")
		}
	case *ChangeMyPasswordRequest:
		if len([]byte(request.CurrentPassword)) > 72 {
			return apperrors.WrapInvalidInput("現在のパスワードは72 バイト以下で入力してください")
		}
		if len([]byte(request.NewPassword)) > 72 {
			return apperrors.WrapInvalidInput("新しいパスワードは72 バイト以下で入力してください")
		}
	}
	return nil
}
