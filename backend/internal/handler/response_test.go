package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	return c, w
}

func TestExtractClinicID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		setupContext    func(c *gin.Context)
		wantClinicID    uint64
		wantOK          bool
		wantStatus      int
		wantBodyContain string
	}{
		{
			name: "extracts valid numeric clinic_id from context",
			setupContext: func(c *gin.Context) {
				c.Set("clinic_id", "42")
			},
			wantClinicID: 42,
			wantOK:       true,
		},
		{
			name:            "returns false when clinic_id key is missing",
			setupContext:    func(_ *gin.Context) {},
			wantClinicID:    0,
			wantOK:          false,
			wantStatus:      http.StatusUnauthorized,
			wantBodyContain: "missing clinic context",
		},
		{
			name: "returns false when clinic_id is not a string",
			setupContext: func(c *gin.Context) {
				c.Set("clinic_id", 42) // int instead of string
			},
			wantClinicID:    0,
			wantOK:          false,
			wantStatus:      http.StatusBadRequest,
			wantBodyContain: "invalid clinic context",
		},
		{
			name: "returns false when clinic_id is non-numeric string",
			setupContext: func(c *gin.Context) {
				c.Set("clinic_id", "not-a-number")
			},
			wantClinicID:    0,
			wantOK:          false,
			wantStatus:      http.StatusBadRequest,
			wantBodyContain: "invalid clinic context",
		},
		{
			name: "returns false for negative numeric string",
			setupContext: func(c *gin.Context) {
				c.Set("clinic_id", "-1")
			},
			wantClinicID:    0,
			wantOK:          false,
			wantStatus:      http.StatusBadRequest,
			wantBodyContain: "invalid clinic context",
		},
		{
			name: "returns false for empty string",
			setupContext: func(c *gin.Context) {
				c.Set("clinic_id", "")
			},
			wantClinicID:    0,
			wantOK:          false,
			wantStatus:      http.StatusBadRequest,
			wantBodyContain: "invalid clinic context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupContext(c)

			clinicID, ok := extractClinicID(c)

			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.Equal(t, tt.wantClinicID, clinicID)
			} else {
				assert.Equal(t, uint64(0), clinicID)
				assert.Equal(t, tt.wantStatus, w.Code)
				assert.Contains(t, w.Body.String(), tt.wantBodyContain)
			}
		})
	}
}

func TestRespondError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
		notInBody  string
	}{
		{
			name:       "not found error maps to 404",
			err:        apperrors.WrapNotFound("owner", "123"),
			wantStatus: http.StatusNotFound,
			wantBody:   "not found",
		},
		{
			name:       "already exists error maps to 409",
			err:        apperrors.WrapAlreadyExists("owner", "test@example.com"),
			wantStatus: http.StatusConflict,
			wantBody:   "owner 'test@example.com' already exists",
		},
		{
			name:       "invalid input error maps to 400",
			err:        apperrors.WrapInvalidInput("name is required"),
			wantStatus: http.StatusBadRequest,
			wantBody:   "name is required",
		},
		{
			name:       "unauthorized sentinel maps to 401",
			err:        apperrors.ErrUnauthorized,
			wantStatus: http.StatusUnauthorized,
			wantBody:   "unauthorized",
		},
		{
			name:       "forbidden sentinel maps to 403",
			err:        apperrors.ErrForbidden,
			wantStatus: http.StatusForbidden,
			wantBody:   "forbidden",
		},
		{
			name:       "unknown error maps to 500 without sensitive info",
			err:        fmt.Errorf("some internal db error with sensitive info"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal server error",
			notInBody:  "sensitive info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := newTestContext()

			RespondError(c, tt.err)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.notInBody != "" {
				assert.NotContains(t, w.Body.String(), tt.notInBody)
			}
		})
	}
}

// thirdPartyCodeMessageError は Code/Message の exported string フィールドを持つが
// service.ReservationLimitError ではない任意のエラー型（サードパーティ SDK エラー等を模す）。
type thirdPartyCodeMessageError struct {
	Code    string
	Message string
}

func (e *thirdPartyCodeMessageError) Error() string {
	return e.Code + ": " + e.Message
}

// TestRespondError_UnclassifiedCustomErrorWithCodeMessageFields は、
// exported Code/Message フィールドを持つが service.ReservationLimitError ではない
// 非 AppError エラーが 500 + "internal server error" に落ちることを検証する
// （第5期回帰: reflection フォールバックにより 409 + 自身のメッセージが漏れていた）。
func TestRespondError_UnclassifiedCustomErrorWithCodeMessageFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newTestContext()

	err := &thirdPartyCodeMessageError{Code: "UPSTREAM_CODE", Message: "sensitive upstream detail"}
	RespondError(c, err)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal server error")
	assert.NotContains(t, w.Body.String(), "sensitive upstream detail")
	assert.NotContains(t, w.Body.String(), "UPSTREAM_CODE")
}

// TestRespondErrorWithExtras_BadGateway は RespondErrorWithExtras に
// apperrors.WrapBadGateway 相当のエラーを渡した場合、502 + 安全メッセージを
// 返すことを検証する（旧 resolveErrorResponse は ErrBadGateway ケースを持たず
// default の 409 + 生 err.Error() に落ちていた）。
func TestRespondErrorWithExtras_BadGateway(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newTestContext()

	err := apperrors.WrapBadGateway("upstream LINE API error")
	RespondErrorWithExtras(c, err, nil)

	assert.Equal(t, http.StatusBadGateway, w.Code)
	assert.Contains(t, w.Body.String(), "upstream LINE API error")
}

// TestRespondErrorWithExtras_NotFoundEmptyMessageUsesDefault は Message=="" の
// NotFound 系 AppError が渡された場合、デフォルトメッセージ "resource not found"
// にフォールバックすることを検証する（Message!="" ガードが無いケースは空文字列を
// そのまま応答してしまっていた）。
func TestRespondErrorWithExtras_NotFoundEmptyMessageUsesDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newTestContext()

	err := &apperrors.AppError{Code: "NOT_FOUND", Message: "", Err: apperrors.ErrNotFound}
	RespondErrorWithExtras(c, err, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "resource not found")
}

// TestRespondError_NotImplementedIgnoresCustomMessage は WrapNotImplemented(customMsg) の
// customMsg がレスポンス本文に反映されず、常に固定メッセージ "not implemented" になる
// ことを検証する（reservation_handler.go 等の既存呼出元の挙動を A-2 の統一で変えないための
// リグレッション防止。resolveErrorResponse の他ケースと異なり、この 1 ケースだけは
// 統一前から appErr.Message を無視する仕様だった）。
func TestRespondError_NotImplementedIgnoresCustomMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newTestContext()

	RespondError(c, apperrors.WrapNotImplemented("この機能は未実装です"))

	assert.Equal(t, http.StatusNotImplemented, w.Code)
	assert.Contains(t, w.Body.String(), "not implemented")
	assert.NotContains(t, w.Body.String(), "この機能は未実装です")
}

func TestParsePagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		query     string
		wantPage  int
		wantLimit int
		wantErr   bool
	}{
		{
			name:      "defaults return page 1 and limit 20",
			query:     "",
			wantPage:  1,
			wantLimit: 20,
			wantErr:   false,
		},
		{
			name:      "custom valid values are accepted",
			query:     "page=2&limit=50",
			wantPage:  2,
			wantLimit: 50,
			wantErr:   false,
		},
		{
			name:      "limit of 100 is accepted",
			query:     "page=1&limit=100",
			wantPage:  1,
			wantLimit: 100,
			wantErr:   false,
		},
		{
			name:    "page zero is invalid",
			query:   "page=0",
			wantErr: true,
		},
		{
			name:    "negative page is invalid",
			query:   "page=-1",
			wantErr: true,
		},
		{
			name:    "limit over 100 is invalid",
			query:   "limit=101",
			wantErr: true,
		},
		{
			name:    "limit zero is invalid",
			query:   "limit=0",
			wantErr: true,
		},
		{
			name:    "non-numeric page is invalid",
			query:   "page=abc",
			wantErr: true,
		},
		{
			name:    "non-numeric limit is invalid",
			query:   "limit=xyz",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)

			page, limit, err := parsePagination(c)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, 0, page)
				assert.Equal(t, 0, limit)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantPage, page)
				assert.Equal(t, tt.wantLimit, limit)
			}
		})
	}
}

// TestCamelToSnake は CamelCase → snake_case 変換を検証する。
// BUG-LINE-010: 連続した大文字（頭字語）を 1 単語として扱うこと。
func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"OwnerName", "owner_name"},
		{"IsDangerous", "is_dangerous"},
		{"TypeID", "type_id"},         // BUG-LINE-010: 以前は "type_i_d"
		{"OwnerID", "owner_id"},       // 同上
		{"HTTPServer", "http_server"}, // 頭字語の末尾で区切る
		{"APIKey", "api_key"},
		{"URL", "url"},
		{"ID", "id"},
		{"Name", "name"},
		{"", ""},
		{"A", "a"},
		{"lowercase", "lowercase"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := camelToSnake(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
