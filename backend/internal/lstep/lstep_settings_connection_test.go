package lstep

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

func allowLoopbackProbeClient(t *testing.T) {
	t.Helper()
	orig := connectionProbeClient
	connectionProbeClient = &http.Client{
		Timeout:       connectionProbeTimeout,
		CheckRedirect: probeCheckRedirect,
	}
	t.Cleanup(func() { connectionProbeClient = orig })
}

func useHttptestProbeClient(t *testing.T, srv *httptest.Server) {
	t.Helper()
	target, err := url.Parse(srv.URL)
	require.NoError(t, err)
	orig := connectionProbeClient
	connectionProbeClient = &http.Client{
		Timeout: connectionProbeTimeout,
		Transport: hostRewriteRoundTripper{
			target: target,
			next:   srv.Client().Transport,
		},
		CheckRedirect: probeCheckRedirect,
	}
	t.Cleanup(func() { connectionProbeClient = orig })
}

type hostRewriteRoundTripper struct {
	target *url.URL
	next   http.RoundTripper
}

func (h hostRewriteRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = h.target.Scheme
	clone.URL.Host = h.target.Host
	clone.Host = h.target.Host
	clone.RequestURI = ""
	next := h.next
	if next == nil {
		next = http.DefaultTransport
	}
	return next.RoundTrip(clone)
}

// ---- testLstepAPI ----

func TestTestLstepAPI(t *testing.T) {
	t.Run("returns nil on 200 OK", func(t *testing.T) {
		allowLoopbackProbeClient(t)
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
		allowLoopbackProbeClient(t)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		err := testLstepAPI(context.Background(), srv.URL, "key")
		assert.NoError(t, err)
	})

	t.Run("returns authentication failed error on 401", func(t *testing.T) {
		allowLoopbackProbeClient(t)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer srv.Close()

		err := testLstepAPI(context.Background(), srv.URL, "bad-key")
		assert.ErrorIs(t, err, errConnectionUnauthorized)
		assert.Equal(t, "unauthorized", classifyConnectionProbeError(err))
	})

	t.Run("returns authentication failed error on 403", func(t *testing.T) {
		allowLoopbackProbeClient(t)
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

	t.Run("does not forward Authorization on redirect to loopback", func(t *testing.T) {
		allowLoopbackProbeClient(t)
		var targetHits int
		var targetAuth string
		target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			targetHits++
			targetAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(target.Close)

		origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, target.URL+"/secret", http.StatusFound)
		}))
		t.Cleanup(origin.Close)

		err := testLstepAPI(context.Background(), origin.URL, "secret-bearer")
		assert.Error(t, err)
		assert.Equal(t, "unreachable", classifyConnectionProbeError(err))
		assert.Equal(t, 0, targetHits)
		assert.Empty(t, targetAuth)
	})
}

// ---- testLineAPI ----

func TestTestLineAPI(t *testing.T) {
	t.Run("returns nil on 200 OK", func(t *testing.T) {
		allowLoopbackProbeClient(t)
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
		allowLoopbackProbeClient(t)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		err := testLineAPI(context.Background(), srv.URL, "token")
		assert.NoError(t, err)
	})

	t.Run("returns authentication failed error on 401", func(t *testing.T) {
		allowLoopbackProbeClient(t)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer srv.Close()

		err := testLineAPI(context.Background(), srv.URL, "bad-token")
		assert.ErrorIs(t, err, errConnectionUnauthorized)
		assert.Equal(t, "unauthorized", classifyConnectionProbeError(err))
	})

	t.Run("returns authentication failed error on 403", func(t *testing.T) {
		allowLoopbackProbeClient(t)
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
		assert.Equal(t, lstep.DefaultBaseURL, got)
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
	t.Run("normalizes default https port to canonical origin", func(t *testing.T) {
		got, err := ValidateLstepBaseURL("https://api.lstep.jp:443/ignored")
		assert.NoError(t, err)
		assert.Equal(t, "https://api.lstep.jp", got)
	})
	t.Run("rejects http loopback", func(t *testing.T) {
		for _, raw := range []string{
			"http://127.0.0.1",
			"http://127.0.0.1:1234",
			"http://localhost",
			"http://localhost:9",
			"http://[::1]",
			"http://[::1]:8080",
		} {
			_, err := ValidateLstepBaseURL(raw)
			assert.Error(t, err, raw)
		}
	})
	t.Run("rejects https loopback and IP literals", func(t *testing.T) {
		for _, raw := range []string{
			"https://127.0.0.1",
			"https://localhost",
			"https://[::1]",
			"https://10.0.0.1",
			"https://169.254.169.254",
			"https://0.0.0.0",
		} {
			_, err := ValidateLstepBaseURL(raw)
			assert.Error(t, err, raw)
		}
	})
	t.Run("rejects userinfo", func(t *testing.T) {
		_, err := ValidateLstepBaseURL("https://user:pass@api.lstep.jp")
		assert.Error(t, err)
	})
	t.Run("rejects custom port on allowlisted host", func(t *testing.T) {
		_, err := ValidateLstepBaseURL("https://api.lstep.jp:8443")
		assert.Error(t, err)
	})
}

func TestProbeDialContext_RejectsForbiddenResolvedAddresses(t *testing.T) {
	tests := []struct {
		name string
		ip   string
	}{
		{name: "loopback v4", ip: "127.0.0.1"},
		{name: "loopback v6", ip: "::1"},
		{name: "unspecified", ip: "0.0.0.0"},
		{name: "private", ip: "10.1.2.3"},
		{name: "link-local", ip: "169.254.1.1"},
		{name: "cgnat", ip: "100.64.0.1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := lookupProbeIPAddr
			lookupProbeIPAddr = func(context.Context, string) ([]net.IPAddr, error) {
				return []net.IPAddr{{IP: net.ParseIP(tt.ip)}}, nil
			}
			t.Cleanup(func() { lookupProbeIPAddr = orig })

			_, err := probeDialContext(context.Background(), "tcp", "api.lstep.jp:443")
			assert.ErrorIs(t, err, errBlockedDialAddress)
		})
	}
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
		useHttptestProbeClient(t, srv)

		repo := &mockLstepSettingsRepository{
			findByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) ([]*model.ClinicIntegration, error) {
				return []*model.ClinicIntegration{
					{KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: "test-key"},
					{KeyName: model.IntegrationKeyLstepBaseURL, KeyValue: "https://api.lstep.jp"},
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
		useHttptestProbeClient(t, srv)

		repo := &mockLstepSettingsRepository{
			findByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) ([]*model.ClinicIntegration, error) {
				return []*model.ClinicIntegration{
					{KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: "bad-key"},
					{KeyName: model.IntegrationKeyLstepBaseURL, KeyValue: "https://api.lstep.jp"},
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

	t.Run("rejects persisted loopback base URL without probing", func(t *testing.T) {
		repo := &mockLstepSettingsRepository{
			findByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) ([]*model.ClinicIntegration, error) {
				return []*model.ClinicIntegration{
					{KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: "test-key"},
					{KeyName: model.IntegrationKeyLstepBaseURL, KeyValue: "http://127.0.0.1:9/secret"},
				}, nil
			},
		}
		svc := NewLstepSettingsService(repo, &mockLstepSyncSettingsRepository{}, nil, nil, nil)

		result, err := svc.TestConnection(context.Background(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.LstepOK)
		assert.Equal(t, "invalid_base_url", result.LstepError)
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
