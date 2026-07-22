package service

// line_reservation_setting_cipher_wiring_test.go — R⑥レビューMEDIUM対応:
// NewServices がtyped compatibility inputのencrypt/decrypt closureをLINE予約設定serviceへ
// そのまま注入することをSave()貫通で固定する。cmd/apiのproduction closure配線は
// lstep_dependencies_test.goが検証する。

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
	reservation.LineReservationSettingRepository
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
	repos := &repository.Repositories{}
	svcs := NewServices(
		repos,
		&reservation.ReservationNotificationConfig{},
		"test-jwt-secret",
		&mockAuditService{},
		&LegacyLstepDependencies{
			TagSync: &mockLstepTagSyncService{},
			EncryptCredential: func(value string) (string, error) {
				return cipher.Encrypt(value)
			},
			DecryptCredential: func(_ context.Context, value string) string {
				plaintext, decryptErr := cipher.Decrypt(value)
				require.NoError(t, decryptErr)
				return plaintext
			},
		},
		repo,
	)

	_, _, err = svcs.LineReservationSetting.Save(context.Background(), 1, &reservation.UpsertLineReservationSettingInput{
		LineChannelSecret: "plain-secret",
		LineAccessToken:   "plain-token",
	})
	require.NoError(t, err)
	require.NotNil(t, repo.persisted)
	// 実 closure（lstep.EncryptLineCredential(cipher, ...)）を通過していれば平文では保存されない
	assert.NotEqual(t, "plain-secret", repo.persisted.LineChannelSecret)
	got, err := cipher.Decrypt(repo.persisted.LineChannelSecret)
	require.NoError(t, err)
	assert.Equal(t, "plain-secret", got)
}
