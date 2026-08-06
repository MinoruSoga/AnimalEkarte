package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- モック実装 ----

type mockSettingLookup struct {
	findFn func(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
}

func (m *mockSettingLookup) FindByClinicID(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
	return m.findFn(ctx, clinicID)
}

type mockCustomerLookup struct {
	findOrCreateFn func(ctx context.Context, clinicID uint64, lineUserID, displayName string) (*model.LineCustomer, error)
}

func (m *mockCustomerLookup) FindOrCreateByLineUserID(ctx context.Context, clinicID uint64, lineUserID, displayName string) (*model.LineCustomer, error) {
	return m.findOrCreateFn(ctx, clinicID, lineUserID, displayName)
}

// ---- テストヘルパー ----

// newLiffAuthRouter はLiffAuth ミドルウェアを組み込んだテスト用 Gin ルーターを返す。
func newLiffAuthRouter(lookup LineCustomerLookup, settingLookup LineReservationSettingLookup) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		// clinicId パラメータを URL から取得するために /:clinicId ルートが必要
		c.Next()
	})
	r.GET("/api/liff/:clinicId/test", LiffAuth(lookup, settingLookup), func(c *gin.Context) {
		customerID, _ := ExtractLiffCustomerID(c)
		c.JSON(http.StatusOK, gin.H{
			"customer_id": customerID,
		})
	})
	return r
}

// startMockLINEServer は LINE ID Token 検証 API のモックサーバーを起動する。
func startMockLINEServer(t *testing.T, statusCode int, resp *lineVerifyResponse) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(resp) //nolint:errcheck // テスト用モックサーバーのレスポンス書き込みエラーは無視してよい
	}))
	t.Cleanup(func() { srv.Close() })
	return srv
}

// validSettingLookup は正常なクリニック設定を返すモック。
func validSettingLookup(liffID string) *mockSettingLookup {
	return &mockSettingLookup{
		findFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
			return &model.LineReservationSetting{ClinicID: 3, LiffID: liffID}, nil
		},
	}
}

// validCustomerLookup は正常な顧客を返すモック。
func validCustomerLookup() *mockCustomerLookup {
	return &mockCustomerLookup{
		findOrCreateFn: func(_ context.Context, _ uint64, _ string, _ string) (*model.LineCustomer, error) {
			return &model.LineCustomer{ID: 1, ClinicID: 3, DisplayName: "テストユーザー"}, nil
		},
	}
}

// ---- LiffAuth ミドルウェアテスト ----

func TestLiffAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		authHeader string
		clinicID   string
		setupLine  func(t *testing.T) // lineVerifyURL を設定
		setting    LineReservationSettingLookup
		customer   LineCustomerLookup
		wantStatus int
		wantBody   string
	}{
		{
			name:       "Authorizationヘッダなし → 401",
			authHeader: "",
			clinicID:   "3",
			setupLine:  nil,
			setting:    validSettingLookup("1234567890-abcdefgh"),
			customer:   validCustomerLookup(),
			wantStatus: http.StatusUnauthorized,
			wantBody:   "missing authorization header",
		},
		{
			name:       "Bearer形式でないトークン → 401",
			authHeader: "Token invalidtoken",
			clinicID:   "3",
			setupLine:  nil,
			setting:    validSettingLookup("1234567890-abcdefgh"),
			customer:   validCustomerLookup(),
			wantStatus: http.StatusUnauthorized,
			wantBody:   "invalid authorization header",
		},
		{
			name:       "Bearerのみでトークンなし → 401",
			authHeader: "Bearer",
			clinicID:   "3",
			setupLine:  nil,
			setting:    validSettingLookup("1234567890-abcdefgh"),
			customer:   validCustomerLookup(),
			wantStatus: http.StatusUnauthorized,
			wantBody:   "invalid authorization header",
		},
		{
			name:       "clinicId が数値でない → 400",
			authHeader: "Bearer validtoken",
			clinicID:   "abc",
			setupLine:  nil,
			setting:    validSettingLookup("1234567890-abcdefgh"),
			customer:   validCustomerLookup(),
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid clinic id",
		},
		{
			name:       "clinicId がゼロ → 400",
			authHeader: "Bearer validtoken",
			clinicID:   "0",
			setupLine:  nil,
			setting:    validSettingLookup("1234567890-abcdefgh"),
			customer:   validCustomerLookup(),
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid clinic id",
		},
		{
			name:       "設定取得失敗（DBエラー）→ 500",
			authHeader: "Bearer validtoken",
			clinicID:   "3",
			setupLine:  nil,
			setting: &mockSettingLookup{findFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, errors.New("connection refused")
			}},
			customer:   validCustomerLookup(),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "failed to load clinic setting",
		},
		{
			name:       "LiffID未設定（LINE予約未構成）→ 503",
			authHeader: "Bearer validtoken",
			clinicID:   "3",
			setupLine:  nil,
			setting:    validSettingLookup(""), // 空文字
			customer:   validCustomerLookup(),
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   "LINE reservation not configured",
		},
		{
			name:       "LINE API がエラーを返す → 401",
			authHeader: "Bearer invalidtoken",
			clinicID:   "3",
			setupLine: func(t *testing.T) {
				srv := startMockLINEServer(t, http.StatusBadRequest, &lineVerifyResponse{
					Error:     "invalid_token",
					ErrorDesc: "The ID token is invalid",
				})
				lineVerifyURL = srv.URL
			},
			setting:    validSettingLookup("1234567890-abcdefgh"),
			customer:   validCustomerLookup(),
			wantStatus: http.StatusUnauthorized,
			wantBody:   "invalid ID token",
		},
		{
			name:       "LINE API が Sub 空文字を返す → 401",
			authHeader: "Bearer emptysub",
			clinicID:   "3",
			setupLine: func(t *testing.T) {
				srv := startMockLINEServer(t, http.StatusOK, &lineVerifyResponse{
					Sub:  "", // empty
					Name: "テスト",
				})
				lineVerifyURL = srv.URL
			},
			setting:    validSettingLookup("1234567890-abcdefgh"),
			customer:   validCustomerLookup(),
			wantStatus: http.StatusUnauthorized,
			wantBody:   "invalid ID token",
		},
		{
			name:       "顧客の取得/作成失敗 → 500",
			authHeader: "Bearer validtoken",
			clinicID:   "3",
			setupLine: func(t *testing.T) {
				srv := startMockLINEServer(t, http.StatusOK, &lineVerifyResponse{
					Sub:  "U1234567890",
					Name: "テストユーザー",
				})
				lineVerifyURL = srv.URL
			},
			setting: validSettingLookup("1234567890-abcdefgh"),
			customer: &mockCustomerLookup{
				findOrCreateFn: func(_ context.Context, _ uint64, _, _ string) (*model.LineCustomer, error) {
					return nil, errors.New("db write error")
				},
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "failed to resolve customer",
		},
		{
			name:       "正常系: コンテキストに顧客情報をセット → 200",
			authHeader: "Bearer validtoken",
			clinicID:   "3",
			setupLine: func(t *testing.T) {
				srv := startMockLINEServer(t, http.StatusOK, &lineVerifyResponse{
					Sub:  "U1234567890",
					Name: "テストユーザー",
				})
				lineVerifyURL = srv.URL
			},
			setting:    validSettingLookup("1234567890-abcdefgh"),
			customer:   validCustomerLookup(),
			wantStatus: http.StatusOK,
			wantBody:   `"customer_id":1`,
		},
	}

	// LIFF_MOCK=true はローカル開発用バイパスで、テスト中は無効にする
	t.Setenv("LIFF_MOCK", "false")

	originalURL := lineVerifyURL
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト後に URL を復元
			lineVerifyURL = originalURL
			if tt.setupLine != nil {
				tt.setupLine(t)
			}

			r := newLiffAuthRouter(tt.customer, tt.setting)
			req := httptest.NewRequest(http.MethodGet, "/api/liff/"+tt.clinicID+"/test", http.NoBody)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "HTTP status mismatch")
			assert.Contains(t, w.Body.String(), tt.wantBody, "response body mismatch")
		})
	}
}

// ---- verifyLiffIDToken 単体テスト ----

func TestVerifyLiffIDToken(t *testing.T) {
	originalURL := lineVerifyURL
	t.Cleanup(func() { lineVerifyURL = originalURL })

	t.Run("LINE APIが正常レスポンスを返す", func(t *testing.T) {
		srv := startMockLINEServer(t, http.StatusOK, &lineVerifyResponse{
			Sub:      "U9999999",
			Name:     "田中太郎",
			ClientID: "1234567890",
		})
		lineVerifyURL = srv.URL

		result, err := verifyLiffIDToken(context.Background(), "mytoken", "1234567890")
		require.NoError(t, err)
		assert.Equal(t, "U9999999", result.Sub)
		assert.Equal(t, "田中太郎", result.Name)
	})

	t.Run("LINE APIがエラーフィールドを返す", func(t *testing.T) {
		srv := startMockLINEServer(t, http.StatusBadRequest, &lineVerifyResponse{
			Error:     "invalid_request",
			ErrorDesc: "The id_token is expired",
		})
		lineVerifyURL = srv.URL

		_, err := verifyLiffIDToken(context.Background(), "expiredtoken", "1234567890")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "The id_token is expired")
	})

	t.Run("Subフィールドがemptyのとき InvalidInput", func(t *testing.T) {
		srv := startMockLINEServer(t, http.StatusOK, &lineVerifyResponse{
			Sub:  "",
			Name: "不明ユーザー",
		})
		lineVerifyURL = srv.URL

		_, err := verifyLiffIDToken(context.Background(), "badtoken", "1234567890")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty line user id")
	})

	t.Run("モックサーバーが不正なJSONを返す", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("NOT JSON")) //nolint:errcheck // テスト用モックサーバーの書き込みエラーは無視してよい
		}))
		t.Cleanup(func() { srv.Close() })
		lineVerifyURL = srv.URL

		_, err := verifyLiffIDToken(context.Background(), "badtoken", "1234567890")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parse line verify response")
	})

	t.Run("サーバーに接続不可のとき error", func(t *testing.T) {
		lineVerifyURL = "http://127.0.0.1:1" // 閉じているポート

		_, err := verifyLiffIDToken(context.Background(), "token", "clientid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "line verify request failed")
	})
}

// ---- LIFF ID からチャンネルID抽出ロジックのテスト ----

// TestLiffChannelIDExtraction はLIFF ID の形式 "{channelID}-{appID}" から
// channelID が正しく取り出されることを検証する（LiffAuth 内のロジック）。
func TestLiffChannelIDExtraction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("LIFF_MOCK", "false")
	originalURL := lineVerifyURL
	t.Cleanup(func() { lineVerifyURL = originalURL })

	var capturedClientID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		capturedClientID = r.FormValue("client_id")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(lineVerifyResponse{Sub: "U0001", Name: "ユーザー"}) //nolint:errcheck // テスト用モックサーバーのレスポンス書き込みエラーは無視してよい
	}))
	t.Cleanup(func() { srv.Close() })
	lineVerifyURL = srv.URL

	setting := &mockSettingLookup{
		findFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
			// LIFF ID = "{channelID}-{appID}" 形式
			return &model.LineReservationSetting{ClinicID: 3, LiffID: "1234567890-abcdefgh"}, nil
		},
	}
	customer := validCustomerLookup()

	r := newLiffAuthRouter(customer, setting)
	req := httptest.NewRequest(http.MethodGet, "/api/liff/3/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer sometoken")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// LINEチャンネルID（ハイフン前の部分）が client_id として渡されること
	assert.Equal(t, "1234567890", capturedClientID)
}

// ---- Extract 関数テスト ----

func TestExtractLiffCustomerID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("存在する場合は取得できる", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(liffCustomerIDKey, uint64(42))

		id, ok := ExtractLiffCustomerID(c)
		assert.True(t, ok)
		assert.Equal(t, uint64(42), id)
	})

	t.Run("存在しない場合は (0, false)", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		id, ok := ExtractLiffCustomerID(c)
		assert.False(t, ok)
		assert.Equal(t, uint64(0), id)
	})
}

// BUG-008/014: LIFF_MOCK must not fall through to real auth on success or failure.
func TestLiffAuth_MockModeBypassesWithoutAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("LIFF_MOCK", "true")

	r := newLiffAuthRouter(validCustomerLookup(), validSettingLookup("1234567890-abcdefgh"))
	req := httptest.NewRequest(http.MethodGet, "/api/liff/3/test", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"customer_id":1`)
}

func TestLiffAuth_MockModeLookupFailureIsFailClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("LIFF_MOCK", "true")

	failing := &mockCustomerLookup{
		findOrCreateFn: func(context.Context, uint64, string, string) (*model.LineCustomer, error) {
			return nil, errors.New("db unavailable")
		},
	}
	r := newLiffAuthRouter(failing, validSettingLookup("1234567890-abcdefgh"))
	req := httptest.NewRequest(http.MethodGet, "/api/liff/3/test", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "LIFF mock authentication failed")
	assert.NotContains(t, w.Body.String(), "missing authorization header")
}

func TestLiffAuth_MockModeInvalidClinicID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("LIFF_MOCK", "true")

	r := newLiffAuthRouter(validCustomerLookup(), validSettingLookup("1234567890-abcdefgh"))
	req := httptest.NewRequest(http.MethodGet, "/api/liff/abc/test", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid clinic id")
}
