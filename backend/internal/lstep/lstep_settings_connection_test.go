package lstep

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- testLstepAPI ----

func TestTestLstepAPI(t *testing.T) {
	t.Run("returns nil on 200 OK", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/tags", r.URL.Path)
			assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		err := testLstepAPI(context.Background(), srv.URL, "test-key")
		assert.NoError(t, err)
	})

	t.Run("returns nil on other non-auth status codes", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		err := testLstepAPI(context.Background(), srv.URL, "key")
		assert.NoError(t, err)
	})

	t.Run("returns authentication failed error on 401", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer srv.Close()

		err := testLstepAPI(context.Background(), srv.URL, "bad-key")
		assert.ErrorIs(t, err, errConnectionUnauthorized)
		assert.Equal(t, "unauthorized", classifyConnectionProbeError(err))
	})

	t.Run("returns authentication failed error on 403", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer srv.Close()

		err := testLstepAPI(context.Background(), srv.URL, "bad-key")
		assert.ErrorIs(t, err, errConnectionUnauthorized)
		assert.Equal(t, "unauthorized", classifyConnectionProbeError(err))
	})

	t.Run("returns error when request build fails", func(t *testing.T) {
		err := testLstepAPI(context.Background(), "://bad-url", "key")
		assert.Error(t, err)
		assert.Equal(t, "unreachable", classifyConnectionProbeError(err))
	})

	t.Run("returns error when connection fails", func(t *testing.T) {
		err := testLstepAPI(context.Background(), "http://127.0.0.1:1", "key")
		assert.Error(t, err)
		assert.Equal(t, "unreachable", classifyConnectionProbeError(err))
	})
}

// ---- testLineAPI ----

func TestTestLineAPI(t *testing.T) {
	t.Run("returns nil on 200 OK", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v2/bot/info", r.URL.Path)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		err := testLineAPI(context.Background(), srv.URL, "test-token")
		assert.NoError(t, err)
	})

	t.Run("returns nil on other non-auth status codes", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		err := testLineAPI(context.Background(), srv.URL, "token")
		assert.NoError(t, err)
	})

	t.Run("returns authentication failed error on 401", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer srv.Close()

		err := testLineAPI(context.Background(), srv.URL, "bad-token")
		assert.ErrorIs(t, err, errConnectionUnauthorized)
		assert.Equal(t, "unauthorized", classifyConnectionProbeError(err))
	})

	t.Run("returns authentication failed error on 403", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer srv.Close()

		err := testLineAPI(context.Background(), srv.URL, "bad-token")
		assert.ErrorIs(t, err, errConnectionUnauthorized)
		assert.Equal(t, "unauthorized", classifyConnectionProbeError(err))
	})

	t.Run("returns error when request build fails", func(t *testing.T) {
		err := testLineAPI(context.Background(), "://bad-url", "token")
		assert.Error(t, err)
		assert.Equal(t, "unreachable", classifyConnectionProbeError(err))
	})

	t.Run("returns error when connection fails", func(t *testing.T) {
		err := testLineAPI(context.Background(), "http://127.0.0.1:1", "token")
		assert.Error(t, err)
		assert.Equal(t, "unreachable", classifyConnectionProbeError(err))
	})
}

func TestValidateLstepBaseURL(t *testing.T) {
	t.Run("empty uses default", func(t *testing.T) {
		got, err := ValidateLstepBaseURL("")
		assert.NoError(t, err)
		assert.Contains(t, got, "https://")
	})
	t.Run("rejects http scheme for public hosts", func(t *testing.T) {
		_, err := ValidateLstepBaseURL("http://api.lstep.jp")
		assert.Error(t, err)
	})
	t.Run("rejects non-allowlisted host", func(t *testing.T) {
		_, err := ValidateLstepBaseURL("https://evil.example.com")
		assert.Error(t, err)
	})
	t.Run("accepts api.lstep.jp", func(t *testing.T) {
		got, err := ValidateLstepBaseURL("https://api.lstep.jp")
		assert.NoError(t, err)
		assert.Equal(t, "https://api.lstep.jp", got)
	})
	t.Run("allows http loopback for local probes", func(t *testing.T) {
		got, err := ValidateLstepBaseURL("http://127.0.0.1:1234")
		assert.NoError(t, err)
		assert.Equal(t, "http://127.0.0.1:1234", got)
	})
}

// ---- TestConnection (lstepSettingsService) ----

func TestLstepSettingsService_TestConnection(t *testing.T) {
	t.Run("returns error when repository fails", func(t *testing.T) {
		repo := &mockLstepSettingsRepository{
			findByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) ([]*model.ClinicIntegration, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewLstepSettingsService(repo, &mockLstepSyncSettingsRepository{}, nil, nil, nil)

		result, err := svc.TestConnection(context.Background(), 1)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	// レコードが無い場合: lstep_api_key / line_channel_access_token 双方が空なので
	// 両方の疎通確認をスキップする（lstepBase のデフォルト代入は実行されるが、
	// testLstepAPI/testLineAPI は呼ばれない）。
	t.Run("skips both checks when no keys are configured", func(t *testing.T) {
		repo := &mockLstepSettingsRepository{
			findByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) ([]*model.ClinicIntegration, error) {
				return []*model.ClinicIntegration{}, nil
			},
		}
		svc := NewLstepSettingsService(repo, &mockLstepSyncSettingsRepository{}, nil, nil, nil)

		result, err := svc.TestConnection(context.Background(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.LstepOK)
		assert.Empty(t, result.LstepError)
		assert.False(t, result.LineOK)
		assert.Empty(t, result.LineError)
	})

	t.Run("returns LstepOK true on successful connection", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		repo := &mockLstepSettingsRepository{
			findByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) ([]*model.ClinicIntegration, error) {
				return []*model.ClinicIntegration{
					{KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: "test-key"},
					{KeyName: model.IntegrationKeyLstepBaseURL, KeyValue: srv.URL},
				}, nil
			},
		}
		svc := NewLstepSettingsService(repo, &mockLstepSyncSettingsRepository{}, nil, nil, nil)

		result, err := svc.TestConnection(context.Background(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.LstepOK)
		assert.Empty(t, result.LstepError)
	})

	t.Run("returns LstepOK false with error message on authentication failure", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer srv.Close()

		repo := &mockLstepSettingsRepository{
			findByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) ([]*model.ClinicIntegration, error) {
				return []*model.ClinicIntegration{
					{KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: "bad-key"},
					{KeyName: model.IntegrationKeyLstepBaseURL, KeyValue: srv.URL},
				}, nil
			},
		}
		svc := NewLstepSettingsService(repo, &mockLstepSyncSettingsRepository{}, nil, nil, nil)

		result, err := svc.TestConnection(context.Background(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.LstepOK)
		assert.NotEmpty(t, result.LstepError)
	})

	t.Run("propagates decrypt errors instead of probing with empty secrets", func(t *testing.T) {
		cipher, err := crypto.NewAESGCMCipher(testIntegrationKeyHex)
		require.NoError(t, err)
		repo := &mockLstepSettingsRepository{
			findByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) ([]*model.ClinicIntegration, error) {
				return []*model.ClinicIntegration{
					{KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: "not-valid-base64!!"},
				}, nil
			},
		}
		svc := NewLstepSettingsService(repo, &mockLstepSyncSettingsRepository{}, cipher, nil, nil)

		result, err := svc.TestConnection(context.Background(), 1)
		require.Error(t, err, "LSB-04: 復号失敗を握り潰さない")
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "decrypt")
	})
}
