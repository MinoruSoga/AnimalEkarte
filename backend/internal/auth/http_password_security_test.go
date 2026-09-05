package auth

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type passwordResetHTTPStub struct {
	forgotPasswordErr error
	forgotPasswordFn  func(context.Context, string) error
}

func (s passwordResetHTTPStub) ForgotPassword(ctx context.Context, email string) error {
	if s.forgotPasswordFn != nil {
		return s.forgotPasswordFn(ctx, email)
	}
	return s.forgotPasswordErr
}

func (passwordResetHTTPStub) ResetPassword(context.Context, string, string) error {
	return nil
}

func (passwordResetHTTPStub) ResetPasswordWithResult(
	context.Context,
	string,
	string,
	CredentialMutationAudit,
) (*PasswordResetCompletion, error) {
	return &PasswordResetCompletion{AccountID: 1}, nil
}

func (passwordResetHTTPStub) Wait() {}

func executeForgotPasswordHTTP(
	t *testing.T,
	service PasswordResetService,
	email string,
) *httptest.ResponseRecorder {
	t.Helper()
	handler := NewHTTPHandler(
		HTTPDependencies{PasswordReset: service},
		CookieConfigForProduction(false),
	)
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/forgot-password",
		strings.NewReader(`{"email":"`+email+`"}`),
	)
	c.Request.Header.Set("Content-Type", "application/json")
	handler.ForgotPassword(c)
	return response
}

func TestHTTPHandler_ForgotPassword_InternalFailuresKeepUniformPublicResponseAndRedactPII(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	const privateEmail = "private-owner@example.test"
	const privateToken = "private-reset-token"

	success := executeForgotPasswordHTTP(t, passwordResetHTTPStub{}, privateEmail)

	var logs bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})
	failure := executeForgotPasswordHTTP(
		t,
		passwordResetHTTPStub{
			forgotPasswordErr: errors.New(
				"database failure for " + privateEmail + " token=" + privateToken,
			),
		},
		privateEmail,
	)

	require.Equal(t, http.StatusOK, success.Code)
	assert.Equal(t, success.Code, failure.Code)
	assert.JSONEq(t, success.Body.String(), failure.Body.String())
	assert.NotContains(t, logs.String(), privateEmail)
	assert.NotContains(t, logs.String(), privateToken)
	assert.Contains(t, logs.String(), "forgot password")
}

func TestHTTPHandler_ForgotPassword_RejectsMissingOrInvalidEmailBeforeService(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	serviceCalls := 0
	service := passwordResetHTTPStub{
		forgotPasswordFn: func(context.Context, string) error {
			serviceCalls++
			return nil
		},
	}
	tests := []struct {
		name string
		body string
	}{
		{name: "missing email", body: `{}`},
		{name: "invalid email", body: `{"email":"not-an-email"}`},
		{name: "incomplete domain email", body: `{"email":"a@"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeCalls := serviceCalls
			handler := NewHTTPHandler(
				HTTPDependencies{PasswordReset: service},
				CookieConfigForProduction(false),
			)
			response := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(response)
			c.Request = httptest.NewRequest(
				http.MethodPost,
				"/api/v1/auth/forgot-password",
				strings.NewReader(test.body),
			)
			c.Request.Header.Set("Content-Type", "application/json")

			handler.ForgotPassword(c)

			assert.Equal(t, http.StatusBadRequest, response.Code)
			assert.Equal(t, beforeCalls, serviceCalls)
		})
	}
}
