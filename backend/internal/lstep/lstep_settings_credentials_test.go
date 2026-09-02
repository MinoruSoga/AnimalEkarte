package lstep

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- encrypt / decrypt ----

// testIntegrationKeyHex は be9_2c_r1_mock_carriers_test.go（R⑥でreservationへ移動した旧line_reservation_setting_service_test.go由来のcarrier） で定義済みの
// 32 バイト（AES-256）ダミー暗号鍵を再利用する。

func TestLstepSettingsService_Encrypt(t *testing.T) {
	cipher, err := crypto.NewAESGCMCipher(testIntegrationKeyHex)
	require.NoError(t, err)

	tests := []struct {
		name       string
		cipher     *crypto.AESGCMCipher
		keyName    string
		value      string
		wantChange bool
	}{
		{
			name:       "cipher が nil の場合は平文のまま返す",
			cipher:     nil,
			keyName:    model.IntegrationKeyLstepAPIKey,
			value:      "raw-api-key",
			wantChange: false,
		},
		{
			name:       "暗号化対象外キーはそのまま返す",
			cipher:     cipher,
			keyName:    model.IntegrationKeyLstepBaseURL,
			value:      "https://example.com",
			wantChange: false,
		},
		{
			name:       "暗号化対象キーは暗号化される",
			cipher:     cipher,
			keyName:    model.IntegrationKeyLstepAPIKey,
			value:      "secret-api-key",
			wantChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &lstepSettingsService{cipher: tt.cipher}
			got, err := s.encrypt(tt.keyName, tt.value)
			require.NoError(t, err)
			if tt.wantChange {
				assert.NotEqual(t, tt.value, got)
			} else {
				assert.Equal(t, tt.value, got)
			}
		})
	}
}

func TestLstepSettingsService_Decrypt(t *testing.T) {
	cipher, err := crypto.NewAESGCMCipher(testIntegrationKeyHex)
	require.NoError(t, err)

	t.Run("cipher が nil の場合は平文のまま返す", func(t *testing.T) {
		s := &lstepSettingsService{cipher: nil}
		got, err := s.decrypt(model.IntegrationKeyLstepAPIKey, "raw-value")
		require.NoError(t, err)
		assert.Equal(t, "raw-value", got)
	})

	t.Run("暗号化対象外キーはそのまま返す", func(t *testing.T) {
		s := &lstepSettingsService{cipher: cipher}
		got, err := s.decrypt(model.IntegrationKeyLstepBaseURL, "https://example.com")
		require.NoError(t, err)
		assert.Equal(t, "https://example.com", got)
	})

	t.Run("暗号化済み値は復号される（round-trip）", func(t *testing.T) {
		s := &lstepSettingsService{cipher: cipher}
		enc, encErr := s.encrypt(model.IntegrationKeyLstepAPIKey, "secret-api-key")
		require.NoError(t, encErr)
		got, decErr := s.decrypt(model.IntegrationKeyLstepAPIKey, enc)
		require.NoError(t, decErr)
		assert.Equal(t, "secret-api-key", got)
	})

	t.Run("不正な暗号文はエラーを返す", func(t *testing.T) {
		s := &lstepSettingsService{cipher: cipher}
		_, err := s.decrypt(model.IntegrationKeyLstepAPIKey, "not-valid-base64-ciphertext!!")
		assert.Error(t, err)
	})
}

// ---- GetRawCredentials ----

func TestLstepSettingsService_GetRawCredentials(t *testing.T) {
	cipher, err := crypto.NewAESGCMCipher(testIntegrationKeyHex)
	require.NoError(t, err)

	t.Run("レコードなし → 空文字とデフォルトBASE URLを返す", func(t *testing.T) {
		repo := &mockLstepSettingsRepository{
			findByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) ([]*model.ClinicIntegration, error) {
				return []*model.ClinicIntegration{}, nil
			},
		}
		svc := NewLstepSettingsService(repo, &mockLstepSyncSettingsRepository{}, nil, nil, nil)
		apiKey, baseURL, lineToken, err := svc.GetRawCredentials(context.Background(), 1)
		require.NoError(t, err)
		assert.Equal(t, "", apiKey)
		assert.Equal(t, "https://api.lstep.jp", baseURL)
		assert.Equal(t, "", lineToken)
	})

	t.Run("cipher なし（平文保存）→ そのまま返す", func(t *testing.T) {
		repo := &mockLstepSettingsRepository{
			findByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) ([]*model.ClinicIntegration, error) {
				return []*model.ClinicIntegration{
					{KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: "plain-api-key"},
					{KeyName: model.IntegrationKeyLstepBaseURL, KeyValue: "https://api.lstep.jp"},
					{KeyName: model.IntegrationKeyLineChannelAccessToken, KeyValue: "plain-line-token"},
				}, nil
			},
		}
		svc := NewLstepSettingsService(repo, &mockLstepSyncSettingsRepository{}, nil, nil, nil)
		apiKey, baseURL, lineToken, err := svc.GetRawCredentials(context.Background(), 1)
		require.NoError(t, err)
		assert.Equal(t, "plain-api-key", apiKey)
		assert.Equal(t, "https://api.lstep.jp", baseURL)
		assert.Equal(t, "plain-line-token", lineToken)
	})

	t.Run("cipher あり（暗号化済み値）→ 復号して返す", func(t *testing.T) {
		s := &lstepSettingsService{cipher: cipher}
		encAPIKey, encErr := s.encrypt(model.IntegrationKeyLstepAPIKey, "secret-api-key")
		require.NoError(t, encErr)
		encLineToken, encErr := s.encrypt(model.IntegrationKeyLineChannelAccessToken, "secret-line-token")
		require.NoError(t, encErr)

		repo := &mockLstepSettingsRepository{
			findByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) ([]*model.ClinicIntegration, error) {
				return []*model.ClinicIntegration{
					{KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: encAPIKey},
					{KeyName: model.IntegrationKeyLineChannelAccessToken, KeyValue: encLineToken},
				}, nil
			},
		}
		svc := NewLstepSettingsService(repo, &mockLstepSyncSettingsRepository{}, cipher, nil, nil)
		apiKey, baseURL, lineToken, err := svc.GetRawCredentials(context.Background(), 1)
		require.NoError(t, err)
		assert.Equal(t, "secret-api-key", apiKey)
		assert.Equal(t, "https://api.lstep.jp", baseURL)
		assert.Equal(t, "secret-line-token", lineToken)
	})

	t.Run("復号に失敗した値は error を返す（空文字フォールバックしない）", func(t *testing.T) {
		// LSB-04 / DEC-35: silent empty substitution hides key-rotation failures.
		repo := &mockLstepSettingsRepository{
			findByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) ([]*model.ClinicIntegration, error) {
				return []*model.ClinicIntegration{
					{KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: "corrupt-ciphertext"},
				}, nil
			},
		}
		svc := NewLstepSettingsService(repo, &mockLstepSyncSettingsRepository{}, cipher, nil, nil)
		apiKey, _, _, err := svc.GetRawCredentials(context.Background(), 1)
		require.Error(t, err)
		assert.Equal(t, "", apiKey)
	})

	t.Run("リポジトリエラー → ラップして返す", func(t *testing.T) {
		repo := &mockLstepSettingsRepository{
			findByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) ([]*model.ClinicIntegration, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewLstepSettingsService(repo, &mockLstepSyncSettingsRepository{}, nil, nil, nil)
		_, _, _, err := svc.GetRawCredentials(context.Background(), 1)
		assert.Error(t, err)
	})

	t.Run("stored loopback base URL is rejected fail-closed", func(t *testing.T) {
		repo := &mockLstepSettingsRepository{
			findByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) ([]*model.ClinicIntegration, error) {
				return []*model.ClinicIntegration{
					{KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: "plain-api-key"},
					{KeyName: model.IntegrationKeyLstepBaseURL, KeyValue: "http://127.0.0.1:9"},
				}, nil
			},
		}
		svc := NewLstepSettingsService(repo, &mockLstepSyncSettingsRepository{}, nil, nil, nil)
		apiKey, baseURL, _, err := svc.GetRawCredentials(context.Background(), 1)
		require.Error(t, err)
		assert.Equal(t, "", apiKey)
		assert.Equal(t, "", baseURL)
	})
}
