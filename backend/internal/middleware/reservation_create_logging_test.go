package middleware

// reservation_create_logging_test.go — EMR-336 RED tests.
//
// POST /api/v1/reservations が 5xx で落ちたとき、request boundary の既存 request
// log 1 行へ安全な診断属性 (error_type / sqlstate / constraint_name) だけを載せる。
// pgconn.PgError の Message/Detail/Hint/Where、ラップ文、生の c.Errors.String()、
// 入力値・owner/pet/staff 由来の値はログにもレスポンスにも出さない。
// sqlstate は大文字英数字 5 文字、constraint_name は 63 bytes 以下の単純 DB 識別子のみ
// 受理し、malformed な値はログへ出さない。診断属性は予約作成 POST に限定し、
// 他 endpoint は従来の error attr 契約を維持する。

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

const reservationCreatePath = "/api/v1/reservations"

// requestLogRecord は request boundary の JSON ログ 1 行の型付き表現。
// optional 属性は *string で、key 欠落 (nil) と空値を区別する。
type requestLogRecord struct {
	Msg            string  `json:"msg"`
	Method         string  `json:"method"`
	Path           string  `json:"path"`
	Status         int     `json:"status"`
	RequestID      string  `json:"request_id"`
	Error          *string `json:"error"`
	ErrorType      *string `json:"error_type"`
	SQLState       *string `json:"sqlstate"`
	ConstraintName *string `json:"constraint_name"`
}

// emr336LeakSentinels はログ・レスポンスへ絶対に出てはいけない合成 sensitive 値と
// 生 error テキスト片。pg Message/Detail/Hint/Where、fmt wrap 文、request body 値を網羅する。
var emr336LeakSentinels = []string{
	"EMR336MESSAGE", "EMR336OWNER", "EMR336PET", "EMR336TIME",
	"EMR336HINT", "EMR336STAFF", "EMR336WHERE", "EMR336BODYVALUE",
	"conflicting key value", "reservation create failed",
}

func captureRequestLogBuffer(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return buf
}

// singleRequestLogRecord は request 境界のログがちょうど 1 行であることを確認して返す。
// 2 行目の専用エラーログ（重複 boundary log）はここで FAIL する。
func singleRequestLogRecord(t *testing.T, buf *bytes.Buffer) requestLogRecord {
	t.Helper()
	raw := strings.TrimSpace(buf.String())
	require.NotEmpty(t, raw, "request boundary must emit exactly one log record")
	lines := strings.Split(raw, "\n")
	require.Len(t, lines, 1,
		"safe diagnostic attrs must ride on the single existing request log, not a second record")
	var rec requestLogRecord
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &rec))
	return rec
}

// assertNoDiagnosticAttrs は新しい診断属性が一切出ていないことを確認する。
func assertNoDiagnosticAttrs(t *testing.T, rec requestLogRecord) {
	t.Helper()
	assert.Nil(t, rec.ErrorType, "error_type must not be emitted on this path")
	assert.Nil(t, rec.SQLState, "sqlstate must not be emitted on this path")
	assert.Nil(t, rec.ConstraintName, "constraint_name must not be emitted on this path")
}

// sensitiveReservationPgError builds a synthetic PgError whose Message/Detail/Hint/Where
// carry EMR336 sentinels. 23P01 on a constraint other than
// excl_appointments_doctor_timerange is fail-closed: it must stay a generic 500.
func sensitiveReservationPgError() *pgconn.PgError {
	return &pgconn.PgError{
		Code:           "23P01",
		ConstraintName: "excl_other_table_timerange",
		Severity:       "ERROR",
		Message:        `conflicting key value violates exclusion constraint "excl_other_table_timerange" EMR336MESSAGE`,
		Detail:         "Key (owner_id, pet_id, start_time)=(EMR336OWNER, EMR336PET, EMR336TIME) already exists.",
		Hint:           "EMR336HINT contact EMR336STAFF",
		Where:          "EMR336WHERE clause fragment",
	}
}

func reservationCreateRouter(handler gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.Use(RequestID())
	r.Use(RequestLoggingMiddleware())
	r.POST(reservationCreatePath, handler)
	return r
}

// RED: wrapped *pgconn.PgError on POST /api/v1/reservations must emit exactly one
// request-log record carrying safe diagnostic attrs and nothing raw.
func TestRequestLoggingMiddleware_ReservationCreate5xx_EmitsSafeDiagnosticAttrs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	wrappedForms := map[string]error{
		"fmt_wrapped": fmt.Errorf("reservation create failed: %w", sensitiveReservationPgError()),
		"fromgorm_wrapped": fmt.Errorf("reservation create failed: %w",
			apperrors.FromGORM(sensitiveReservationPgError(), "reservation", "")),
	}
	for name, handlerErr := range wrappedForms {
		t.Run(name, func(t *testing.T) {
			buf := captureRequestLogBuffer(t)
			r := reservationCreateRouter(func(c *gin.Context) {
				httpapi.RespondError(c, handlerErr)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, reservationCreatePath,
				strings.NewReader(`{"owner_name":"EMR336BODYVALUE"}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Request-ID", "EMR336REQID001")
			r.ServeHTTP(w, req)

			// Response stays generic: no pg internals, no sentinel values.
			assert.Equal(t, http.StatusInternalServerError, w.Code)
			for _, s := range emr336LeakSentinels {
				assert.NotContains(t, w.Body.String(), s, "response must not leak %q", s)
			}
			assert.NotContains(t, w.Body.String(), "23P01")
			assert.NotContains(t, w.Body.String(), "excl_other_table_timerange")

			rec := singleRequestLogRecord(t, buf)
			assert.Equal(t, "server error", rec.Msg)
			assert.Equal(t, http.StatusInternalServerError, rec.Status)
			assert.Equal(t, http.MethodPost, rec.Method)
			assert.Equal(t, reservationCreatePath, rec.Path)

			// Request correlation: log request_id must match the response header.
			assert.Equal(t, "EMR336REQID001", rec.RequestID)
			assert.Equal(t, w.Header().Get("X-Request-ID"), rec.RequestID)

			// Safe diagnostic class only.
			require.NotNil(t, rec.ErrorType)
			assert.Equal(t, "*pgconn.PgError", *rec.ErrorType)
			require.NotNil(t, rec.SQLState)
			assert.Equal(t, "23P01", *rec.SQLState)
			require.NotNil(t, rec.ConstraintName)
			assert.Equal(t, "excl_other_table_timerange", *rec.ConstraintName)
			assert.Nil(t, rec.Error, "raw c.Errors.String() must not be emitted on reservation create")

			line := buf.String()
			for _, s := range emr336LeakSentinels {
				assert.NotContains(t, line, s, "request log must not leak %q", s)
			}
		})
	}

	// RED: a non-pg error still emits a safe type name but never raw error text.
	t.Run("generic_wrapped_error", func(t *testing.T) {
		buf := captureRequestLogBuffer(t)
		r := reservationCreateRouter(func(c *gin.Context) {
			httpapi.RespondError(c,
				fmt.Errorf("reservation create failed: %w", errors.New("EMR336GENERIC raw detail")))
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, reservationCreatePath, http.NoBody)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		rec := singleRequestLogRecord(t, buf)
		require.NotNil(t, rec.ErrorType, "generic 5xx on reservation create must emit a safe error_type")
		assert.NotEmpty(t, *rec.ErrorType)
		assert.NotContains(t, *rec.ErrorType, " ", "error_type must be a type name, not free text")
		assert.Nil(t, rec.SQLState)
		assert.Nil(t, rec.ConstraintName)
		assert.Nil(t, rec.Error)
		assert.NotContains(t, buf.String(), "EMR336GENERIC")
	})
}

// RED: malformed SQLSTATE / constraint metadata must be rejected, never logged raw.
func TestRequestLoggingMiddleware_ReservationCreate5xx_RejectsMalformedPgMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := map[string]*pgconn.PgError{
		"sqlstate_too_short": {
			Code: "23P", ConstraintName: "excl_valid_name", Message: "EMR336MSHORT",
		},
		"sqlstate_lowercase": {
			Code: "23p01", ConstraintName: "excl_valid_name", Message: "EMR336MLOWER",
		},
		"sqlstate_too_long": {
			Code: "23P011", ConstraintName: "excl_valid_name", Message: "EMR336MLONG",
		},
		"constraint_with_spaces": {
			Code: "23P01", ConstraintName: "excl bad name", Message: "EMR336MSPACE",
		},
		"constraint_with_quote_injection": {
			Code: "23P01", ConstraintName: `excl" EMR336INJECT "x`, Message: "EMR336MINJECT",
		},
		"constraint_over_63_bytes": {
			Code: "23P01", ConstraintName: strings.Repeat("e", 64), Message: "EMR336M64B",
		},
	}
	for name, pgErr := range cases {
		t.Run(name, func(t *testing.T) {
			buf := captureRequestLogBuffer(t)
			r := reservationCreateRouter(func(c *gin.Context) {
				httpapi.RespondError(c, fmt.Errorf("reservation create failed: %w", pgErr))
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, reservationCreatePath, http.NoBody)
			r.ServeHTTP(w, req)
			require.Equal(t, http.StatusInternalServerError, w.Code)

			rec := singleRequestLogRecord(t, buf)
			line := buf.String()

			// The well-formed half of the pair is still emitted; the malformed half is
			// rejected — its raw text must not appear anywhere in the log line.
			if pgErr.Code == "23P01" {
				require.NotNil(t, rec.SQLState)
				assert.Equal(t, "23P01", *rec.SQLState)
			} else {
				if rec.SQLState != nil {
					assert.NotEqual(t, pgErr.Code, *rec.SQLState)
				}
				assert.NotContains(t, line, pgErr.Code,
					"malformed SQLSTATE must not reach the log")
			}
			if pgErr.ConstraintName == "excl_valid_name" {
				require.NotNil(t, rec.ConstraintName)
				assert.Equal(t, "excl_valid_name", *rec.ConstraintName)
			} else {
				if rec.ConstraintName != nil {
					assert.NotEqual(t, pgErr.ConstraintName, *rec.ConstraintName)
				}
				assert.NotContains(t, line, pgErr.ConstraintName,
					"malformed constraint name must not reach the log")
			}
			assert.Nil(t, rec.Error)
			assert.NotContains(t, line, pgErr.Message)
		})
	}
}

// Guard: success and 4xx on the same path must not emit the diagnostic class at all.
func TestRequestLoggingMiddleware_ReservationCreate_NoDiagnosticAttrsBelow5xx(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name    string
		handler gin.HandlerFunc
		status  int
	}{
		{"success_2xx", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"id": 1})
		}, http.StatusCreated},
		{"client_error_4xx", func(c *gin.Context) {
			httpapi.RespondError(c, apperrors.WrapInvalidInput("EMR336INVALID detail"))
		}, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := captureRequestLogBuffer(t)
			r := reservationCreateRouter(tc.handler)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, reservationCreatePath, http.NoBody)
			r.ServeHTTP(w, req)
			require.Equal(t, tc.status, w.Code)

			rec := singleRequestLogRecord(t, buf)
			assertNoDiagnosticAttrs(t, rec)
			assert.Nil(t, rec.Error)
			assert.NotContains(t, buf.String(), "EMR336INVALID")
		})
	}
}

// Guard: the safe diagnostic class is scoped to reservation create only. Other
// endpoints — including a non-POST on the same path — keep the existing raw
// error attr contract unchanged.
func TestRequestLoggingMiddleware_NonReservationEndpoint_KeepsRawErrorAttr(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name   string
		method string
		path   string
	}{
		{"other_post_endpoint", http.MethodPost, "/api/v1/other"},
		{"reservation_list_get", http.MethodGet, reservationCreatePath},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := captureRequestLogBuffer(t)
			r := gin.New()
			r.Use(RequestID())
			r.Use(RequestLoggingMiddleware())
			h := func(c *gin.Context) {
				httpapi.RespondError(c, fmt.Errorf("legacy path: %w", sensitiveReservationPgError()))
			}
			if tc.method == http.MethodPost {
				r.POST(tc.path, h)
			} else {
				r.GET(tc.path, h)
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, http.NoBody)
			r.ServeHTTP(w, req)
			require.Equal(t, http.StatusInternalServerError, w.Code)

			rec := singleRequestLogRecord(t, buf)
			require.NotNil(t, rec.Error, "non-create endpoints keep the existing raw error attr")
			assert.Contains(t, *rec.Error, "SQLSTATE 23P01")
			assertNoDiagnosticAttrs(t, rec)
		})
	}
}
