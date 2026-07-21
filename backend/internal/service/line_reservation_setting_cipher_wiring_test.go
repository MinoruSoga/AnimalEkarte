package service

// line_reservation_setting_cipher_wiring_test.go — R⑥レビューMEDIUM対応:
// NewServices が LINE 予約設定サービスへ注入する encrypt/decrypt closure が
// 実 encryptLineCredential/decryptLineCredential（実 cipher）へ正しく配線されることを
// Save() 貫通で固定する統合テスト（reservation 側テストは複製 closure を使うため、
// 実配線はここが唯一の検証点・new_services_cipher_test.go の X-2 と同型）。

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

type wiringLineSettingRepo struct {
	repository.LineReservationSettingRepository
	persisted *model.LineReservationSetting
}

func (r *wiringLineSettingRepo) FindByClinicID(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
	return r.persisted, nil
}

func (r *wiringLineSettingRepo) Save(_ context.Context, _ uint64, setting *model.LineReservationSetting) error {
	r.persisted = setting
	return nil
}

func TestNewServices_LineReservationSettingUsesInjectedCipherClosures(t *testing.T) {
	cipher, err := crypto.NewAESGCMCipher(testIntegrationKeyHex)
	require.NoError(t, err)

	repo := &wiringLineSettingRepo{}
	repos := &repository.Repositories{
		LineReservationSetting: repo,
		LstepSettings:          &mockLstepSettingsRepository{},
		LstepSyncSettings:      &mockLstepSyncSettingsRepository{},
	}
	svcs := NewServices(repos, &reservation.ReservationNotificationConfig{}, cipher, nil, "test-jwt-secret")

	_, _, err = svcs.LineReservationSetting.Save(context.Background(), 1, &reservation.UpsertLineReservationSettingInput{
		LineChannelSecret: "plain-secret",
		LineAccessToken:   "plain-token",
	})
	require.NoError(t, err)
	require.NotNil(t, repo.persisted)
	// 実 closure（encryptLineCredential(cipher, ...)）を通過していれば平文では保存されない
	assert.NotEqual(t, "plain-secret", repo.persisted.LineChannelSecret)
	got, err := cipher.Decrypt(repo.persisted.LineChannelSecret)
	require.NoError(t, err)
	assert.Equal(t, "plain-secret", got)
}
