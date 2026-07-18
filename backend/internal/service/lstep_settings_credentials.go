package service

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

// encrypt は暗号化が必要なキーのみ暗号化する。cipher が nil なら平文のまま返す（開発環境）。
func (s *lstepSettingsService) encrypt(keyName, value string) (string, error) {
	if s.cipher == nil || !model.IsEncryptedKey(keyName) {
		return value, nil
	}
	return s.cipher.Encrypt(value)
}

// decrypt は暗号化済みキーを復号する。cipher が nil なら平文のまま返す（開発環境）。
func (s *lstepSettingsService) decrypt(keyName, value string) (string, error) {
	if s.cipher == nil || !model.IsEncryptedKey(keyName) {
		return value, nil
	}
	return s.cipher.Decrypt(value)
}

// GetRawCredentials は復号済みの Lステップ API キー・BASE URL・LINE アクセストークンを返す。
func (s *lstepSettingsService) GetRawCredentials(ctx context.Context, clinicID uint64) (apiKey, baseURL, lineToken string, err error) {
	records, err := s.repo.FindByClinicAndService(ctx, clinicID, model.IntegrationServiceLstep)
	if err != nil {
		return "", "", "", apperrors.Wrap(err, "failed to find lstep settings")
	}
	kvMap := make(map[string]string, len(records))
	for _, r := range records {
		val, decErr := s.decrypt(r.KeyName, r.KeyValue)
		if decErr != nil {
			slog.ErrorContext(ctx, "failed to decrypt integration value", "key_name", r.KeyName)
			val = ""
		}
		kvMap[r.KeyName] = val
	}
	base := kvMap[model.IntegrationKeyLstepBaseURL]
	if base == "" {
		base = lstep.DefaultBaseURL
	}
	return kvMap[model.IntegrationKeyLstepAPIKey], base, kvMap[model.IntegrationKeyLineChannelAccessToken], nil
}
