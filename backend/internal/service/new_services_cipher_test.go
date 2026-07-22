package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/reservation"
)

// TestNewServices_OwnerUsesInjectedCipherForLstepDecryption は X-2 の回帰テストである。
// service.go は LstepSettingsService を nil cipher で決め打ち構築していたため、
// Owner を含む6サービス（Owner/Pet/Accounting/Vaccination/Prescription/Aggregation）が
// 保持する LstepTagSyncService 経由の復号が常に平文パススルーになり、main.go の
// svcs.LstepSettings 再代入では波及しなかった（クロージャで既に古いインスタンスを注入済みのため）。
func TestNewServices_OwnerUsesInjectedCipherForLstepDecryption(t *testing.T) {
	cipher, err := crypto.NewAESGCMCipher(testIntegrationKeyHex)
	require.NoError(t, err)

	encAPIKey, err := cipher.Encrypt("secret-lstep-api-key")
	require.NoError(t, err)

	settingsRepo := &mockLstepSettingsRepository{
		findByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) ([]*model.ClinicIntegration, error) {
			return []*model.ClinicIntegration{
				{KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: encAPIKey},
			}, nil
		},
	}

	repos := &repository.Repositories{
		LstepSettings:     settingsRepo,
		LstepSyncSettings: &mockLstepSyncSettingsRepository{},
	}

	svcs := NewServices(repos, &reservation.ReservationNotificationConfig{}, cipher, nil, "test-jwt-secret")

	// Owner の tagSyncSvc が保持する settingsSvc は NewServices が公開する
	// svcs.LstepSettings と同一インスタンスでなければならない（DI 一貫性）。
	ownerImpl, ok := svcs.Owner.(*ownerService)
	require.True(t, ok, "svcs.Owner の実装型が想定と異なる")
	assert.Same(t, svcs.LstepTagSync, ownerImpl.tagSyncSvc,
		"Owner が参照する LstepTagSyncService は NewServices が返すインスタンスと同一であるべき")

	// NewServices に渡した cipher が Owner 経由の LstepSettingsService まで伝播しているかを検証する。
	// nil cipher で構築されていれば暗号文がそのまま返り、平文と一致しない。
	apiKey, _, _, err := svcs.LstepSettings.GetRawCredentials(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "secret-lstep-api-key", apiKey,
		"NewServices に渡された cipher が Owner 経由の LstepSettingsService に伝播していない（service.go の nil 決め打ちバグ）")
}
