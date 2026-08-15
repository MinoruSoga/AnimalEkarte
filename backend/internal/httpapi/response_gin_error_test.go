package httpapi

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

type errorResponderCase struct {
	name     string
	respond  func(*gin.Context, error)
	wantBody string
}

func errorResponderCases() []errorResponderCase {
	return []errorResponderCase{
		{
			name:     "RespondError",
			respond:  RespondError,
			wantBody: `{"error":"internal server error"}`,
		},
		{
			name: "RespondErrorWithExtras",
			respond: func(c *gin.Context, err error) {
				RespondErrorWithExtras(c, err, nil)
			},
			wantBody: `{"code":"","error":"internal server error"}`,
		},
	}
}

func TestRespondError_GinErrorRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantState  string
	}{
		{
			name: "unknown pg error maps to 500",
			err: fmt.Errorf("database error: %w", &pgconn.PgError{
				Code:    "42703",
				Message: "column pets.version does not exist",
			}),
			wantStatus: http.StatusInternalServerError,
			wantState:  "42703",
		},
		{
			name:       "bad gateway maps to 502",
			err:        apperrors.WrapBadGateway("upstream service failed"),
			wantStatus: http.StatusBadGateway,
		},
	}

	for _, responder := range errorResponderCases() {
		t.Run(responder.name, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					c, w := newTestContext()

					responder.respond(c, tt.err)

					assert.Equal(t, tt.wantStatus, w.Code)
					require.Len(t, c.Errors, 1)
					assert.ErrorIs(t, c.Errors[0].Err, tt.err)
					if tt.wantState != "" {
						assert.Contains(t, c.Errors.String(), tt.wantState)
					}
				})
			}
		})
	}
}

func TestRespondError_KnownPgErrorsDoNotRegisterGinError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	knownCodes := []string{"23503", "23505", "22003", "22P02", "23514"}

	for _, responder := range errorResponderCases() {
		t.Run(responder.name, func(t *testing.T) {
			for _, code := range knownCodes {
				t.Run(code, func(t *testing.T) {
					c, w := newTestContext()
					err := &pgconn.PgError{
						Code:           code,
						Message:        "sensitive user-derived value",
						ConstraintName: "sensitive_constraint",
						Detail:         "sensitive detail",
					}

					responder.respond(c, err)

					assert.Equal(t, http.StatusBadRequest, w.Code)
					assert.Empty(t, c.Errors)
				})
			}
		})
	}
}

func TestRespondError_UnknownPgResponseIsByteStableAndNonLeaking(t *testing.T) {
	gin.SetMode(gin.TestMode)

	err := fmt.Errorf("database error: %w", &pgconn.PgError{
		Code:           "42703",
		Message:        "column pets.version does not exist",
		ConstraintName: "pets_version_check",
		TableName:      "pets",
		Detail:         "SELECT pets.version FROM pets LEFT JOIN owners",
	})

	for _, responder := range errorResponderCases() {
		t.Run(responder.name, func(t *testing.T) {
			c, w := newTestContext()

			responder.respond(c, err)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
			assert.Equal(t, responder.wantBody, w.Body.String())
			for _, leaked := range []string{
				"42703",
				"SQLSTATE",
				"column pets.version does not exist",
				"pets_version_check",
				"pets",
				"SELECT",
			} {
				assert.Falsef(t, strings.Contains(w.Body.String(), leaked), "response body leaks %q", leaked)
			}
		})
	}
}

func TestRespondError_NilDoesNotPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, responder := range errorResponderCases() {
		t.Run(responder.name, func(t *testing.T) {
			c, w := newTestContext()

			assert.NotPanics(t, func() {
				responder.respond(c, nil)
			})

			assert.Equal(t, http.StatusInternalServerError, w.Code)
			assert.Empty(t, c.Errors)
			assert.Equal(t, responder.wantBody, w.Body.String())
		})
	}
}
