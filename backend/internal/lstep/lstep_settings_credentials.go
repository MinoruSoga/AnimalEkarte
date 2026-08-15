package lstep

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

// allowedLstepAPIHosts is the host allowlist for LSTEP base URLs (LSA-01).
func allowedLstepAPIHosts() map[string]struct{} {
	hosts := map[string]struct{}{"api.lstep.jp": {}}
	if u, err := url.Parse(lstep.DefaultBaseURL); err == nil && u.Hostname() != "" {
		hosts[strings.ToLower(u.Hostname())] = struct{}{}
	}
	return hosts
}

// ValidateLstepBaseURL enforces https + host allowlist for lstep_base_url (LSA-01).
// Empty raw returns DefaultBaseURL.
func ValidateLstepBaseURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return lstep.DefaultBaseURL, nil
	}
	u, err := url.Parse(trimmed)
	if err != nil {
		return "", apperrors.WrapInvalidInput("lstep_base_url is not a valid URL")
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return "", apperrors.WrapInvalidInput("lstep_base_url host is required")
	}
	// Loopback http is allowed only for local connectivity probes (httptest); never for public hosts.
	isLoopback := host == "127.0.0.1" || host == "localhost" || host == "::1"
	if u.Scheme == "http" {
		if !isLoopback {
			return "", apperrors.WrapInvalidInput("lstep_base_url must use https")
		}
	} else if u.Scheme != "https" {
		return "", apperrors.WrapInvalidInput("lstep_base_url must use https")
	}
	if !isLoopback {
		if _, ok := allowedLstepAPIHosts()[host]; !ok {
			return "", apperrors.WrapInvalidInput(fmt.Sprintf("lstep_base_url host %q is not allowed", host))
		}
	}
	if u.User != nil {
		return "", apperrors.WrapInvalidInput("lstep_base_url must not include userinfo")
	}
	out := u.Scheme + "://" + u.Host
	if u.Path != "" && u.Path != "/" {
		out += strings.TrimRight(u.Path, "/")
	}
	return out, nil
}

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
			// LSB-04 / DEC-35: 復号失敗を空文字へ置換して握り潰さない（サイレント停止を防ぐ）
			slog.ErrorContext(ctx, "failed to decrypt integration value", "key_name", r.KeyName, "error", decErr)
			return "", "", "", apperrors.Wrap(decErr, "failed to decrypt integration value")
		}
		kvMap[r.KeyName] = val
	}
	base, err := ValidateLstepBaseURL(kvMap[model.IntegrationKeyLstepBaseURL])
	if err != nil {
		// Fail closed: never hand a non-allowlisted host to callers that attach the API key.
		return "", "", "", err
	}
	return kvMap[model.IntegrationKeyLstepAPIKey], base, kvMap[model.IntegrationKeyLineChannelAccessToken], nil
}
