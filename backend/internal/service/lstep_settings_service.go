package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// LstepSettingsResponse はGETレスポンス用（マスク済み）
type LstepSettingsResponse struct {
	LstepAPIKeyMasked            string     `json:"lstep_api_key_masked"`
	LstepBaseURL                 string     `json:"lstep_base_url"`
	LineChannelAccessTokenMasked string     `json:"line_channel_access_token_masked"`
	LineChannelSecretMasked      string     `json:"line_channel_secret_masked"`
	LiffID                       string     `json:"liff_id"`
	LineAccountName              string     `json:"line_account_name"`
	IsConfigured                 bool       `json:"is_configured"`
	LastUpdatedAt                *time.Time `json:"last_updated_at"`
}

// UpdateLstepSettingsInput はPATCHリクエスト用（空文字=変更なし）
type UpdateLstepSettingsInput struct {
	LstepAPIKey            string
	LstepBaseURL           string
	LineChannelAccessToken string
	LineChannelSecret      string
	LiffID                 string
	LineAccountName        string
}

// LstepConnectionTestResult は疎通確認結果
type LstepConnectionTestResult struct {
	LstepOK    bool   `json:"lstep_ok"`
	LstepError string `json:"lstep_error,omitempty"`
	LineOK     bool   `json:"line_ok"`
	LineError  string `json:"line_error,omitempty"`
}

// LstepSettingsService は Lステップ/LINE連携設定の管理インターフェース。
type LstepSettingsService interface {
	GetSettings(ctx context.Context, clinicID uint64) (*LstepSettingsResponse, error)
	UpdateSettings(ctx context.Context, clinicID uint64, input UpdateLstepSettingsInput, actorID *uint64) (*LstepSettingsResponse, error)
	DeleteSettings(ctx context.Context, clinicID uint64, actorID *uint64) error
	TestConnection(ctx context.Context, clinicID uint64) (*LstepConnectionTestResult, error)
	// GetRawCredentials は復号済みの API キー・BASE URL・LINE アクセストークンを返す。
	// 設定が存在しない場合は空文字を返す（エラーにはならない）。
	GetRawCredentials(ctx context.Context, clinicID uint64) (apiKey, baseURL, lineToken string, err error)
}

type lstepSettingsService struct {
	repo     repository.LstepSettingsRepository
	cipher   *crypto.AESGCMCipher
	auditSvc AuditService
}

// NewLstepSettingsService は LstepSettingsService を初期化して返す。
// cipher が nil の場合は暗号化なしで動作する（開発環境で INTEGRATION_ENCRYPTION_KEY 未設定時）。
// auditSvc が nil の場合は監査ログをスキップする（CLI ツール等での使用を想定）。
func NewLstepSettingsService(repo repository.LstepSettingsRepository, cipher *crypto.AESGCMCipher, auditSvc AuditService) LstepSettingsService {
	return &lstepSettingsService{repo: repo, cipher: cipher, auditSvc: auditSvc}
}

func (s *lstepSettingsService) GetSettings(ctx context.Context, clinicID uint64) (*LstepSettingsResponse, error) {
	records, err := s.repo.FindByClinicAndService(ctx, clinicID, model.IntegrationServiceLstep)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find lstep settings", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to find lstep settings")
	}

	kvMap := make(map[string]string, len(records))
	var lastUpdated *time.Time
	for _, r := range records {
		val, decErr := s.decrypt(r.KeyName, r.KeyValue)
		if decErr != nil {
			slog.ErrorContext(ctx, "failed to decrypt integration value", "key_name", r.KeyName)
			val = ""
		}
		kvMap[r.KeyName] = val
		if lastUpdated == nil || r.UpdatedAt.After(*lastUpdated) {
			t := r.UpdatedAt
			lastUpdated = &t
		}
	}

	resp := buildLstepSettingsResponse(kvMap, lastUpdated)
	return resp, nil
}

func buildLstepSettingsResponse(kvMap map[string]string, lastUpdated *time.Time) *LstepSettingsResponse {
	mask := func(v string) string {
		if v == "" {
			return ""
		}
		return crypto.MaskValue(v)
	}
	apiKey := kvMap[model.IntegrationKeyLstepAPIKey]
	token := kvMap[model.IntegrationKeyLineChannelAccessToken]
	secret := kvMap[model.IntegrationKeyLineChannelSecret]
	isConfigured := apiKey != "" || token != ""
	return &LstepSettingsResponse{
		LstepAPIKeyMasked:            mask(apiKey),
		LstepBaseURL:                 kvMap[model.IntegrationKeyLstepBaseURL],
		LineChannelAccessTokenMasked: mask(token),
		LineChannelSecretMasked:      mask(secret),
		LiffID:                       kvMap[model.IntegrationKeyLiffID],
		LineAccountName:              kvMap[model.IntegrationKeyLineAccountName],
		IsConfigured:                 isConfigured,
		LastUpdatedAt:                lastUpdated,
	}
}

func (s *lstepSettingsService) UpdateSettings(ctx context.Context, clinicID uint64, input UpdateLstepSettingsInput, actorID *uint64) (*LstepSettingsResponse, error) {
	pairs := []struct {
		keyName string
		value   string
	}{
		{model.IntegrationKeyLstepAPIKey, input.LstepAPIKey},
		{model.IntegrationKeyLstepBaseURL, input.LstepBaseURL},
		{model.IntegrationKeyLineChannelAccessToken, input.LineChannelAccessToken},
		{model.IntegrationKeyLineChannelSecret, input.LineChannelSecret},
		{model.IntegrationKeyLiffID, input.LiffID},
		{model.IntegrationKeyLineAccountName, input.LineAccountName},
	}

	for _, pair := range pairs {
		if pair.value == "" {
			continue // 空文字=変更なし（誤上書き防止）
		}
		encrypted, err := s.encrypt(pair.keyName, pair.value)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to encrypt integration value")
		}
		record := &model.ClinicIntegration{
			ClinicID: clinicID,
			Service:  model.IntegrationServiceLstep,
			KeyName:  pair.keyName,
			KeyValue: encrypted,
		}
		if err := s.repo.Upsert(ctx, record); err != nil {
			slog.ErrorContext(ctx, "failed to upsert lstep setting", "error", err, "key_name", pair.keyName)
			return nil, apperrors.Wrap(err, "failed to update lstep setting")
		}
	}
	resp, err := s.GetSettings(ctx, clinicID)
	if err == nil && s.auditSvc != nil {
		_ = s.auditSvc.LogLstepOperation(ctx, clinicID, actorID, "update_settings", "clinic", &clinicID)
	}
	return resp, err
}

func (s *lstepSettingsService) DeleteSettings(ctx context.Context, clinicID uint64, actorID *uint64) error {
	if err := s.repo.DeleteByClinicAndService(ctx, clinicID, model.IntegrationServiceLstep); err != nil {
		slog.ErrorContext(ctx, "failed to delete lstep settings", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete lstep settings")
	}
	if s.auditSvc != nil {
		_ = s.auditSvc.LogLstepOperation(ctx, clinicID, actorID, "delete_settings", "clinic", &clinicID)
	}
	return nil
}

func (s *lstepSettingsService) TestConnection(ctx context.Context, clinicID uint64) (*LstepConnectionTestResult, error) {
	records, err := s.repo.FindByClinicAndService(ctx, clinicID, model.IntegrationServiceLstep)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find lstep settings for test", "error", err)
		return nil, apperrors.Wrap(err, "failed to load settings for connection test")
	}

	kvMap := make(map[string]string, len(records))
	for _, r := range records {
		val, _ := s.decrypt(r.KeyName, r.KeyValue)
		kvMap[r.KeyName] = val
	}

	result := &LstepConnectionTestResult{}

	// Lステップ疎通確認
	lstepKey := kvMap[model.IntegrationKeyLstepAPIKey]
	lstepBase := kvMap[model.IntegrationKeyLstepBaseURL]
	if lstepBase == "" {
		lstepBase = "https://api.lstep.jp"
	}
	if lstepKey != "" {
		if err := testLstepAPI(ctx, lstepBase, lstepKey); err != nil {
			result.LstepOK = false
			result.LstepError = err.Error()
		} else {
			result.LstepOK = true
		}
	}

	// LINE Messaging API疎通確認
	lineToken := kvMap[model.IntegrationKeyLineChannelAccessToken]
	if lineToken != "" {
		if err := testLineAPI(ctx, lineToken); err != nil {
			result.LineOK = false
			result.LineError = err.Error()
		} else {
			result.LineOK = true
		}
	}

	return result, nil
}

func testLstepAPI(ctx context.Context, baseURL, apiKey string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/v1/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()               //nolint:errcheck
	_, _ = io.Copy(io.Discard, resp.Body) //nolint:errcheck
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("authentication failed: HTTP %d", resp.StatusCode)
	}
	return nil
}

func testLineAPI(ctx context.Context, channelToken string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.line.me/v2/bot/info", nil)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+channelToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()               //nolint:errcheck
	_, _ = io.Copy(io.Discard, resp.Body) //nolint:errcheck
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("authentication failed: HTTP %d", resp.StatusCode)
	}
	return nil
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
			slog.ErrorContext(ctx, "failed to decrypt integration value", "key_name", r.KeyName)
			val = ""
		}
		kvMap[r.KeyName] = val
	}
	base := kvMap[model.IntegrationKeyLstepBaseURL]
	if base == "" {
		base = "https://api.lstep.jp"
	}
	return kvMap[model.IntegrationKeyLstepAPIKey], base, kvMap[model.IntegrationKeyLineChannelAccessToken], nil
}
