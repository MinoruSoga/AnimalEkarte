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
	IsSyncEnabled                bool       `json:"is_sync_enabled"`
	SyncEnabledAt                *time.Time `json:"sync_enabled_at"`
	CPMVersion                   string     `json:"cpm_version"`
	DormantPrevention180Days     int        `json:"dormant_prevention_180_days"`
	DormantPrevention210Days     int        `json:"dormant_prevention_210_days"`
	DormantPrevention240Days     int        `json:"dormant_prevention_240_days"`
	DormantPrevention365Days     int        `json:"dormant_prevention_365_days"`
	// P1 CPM V2 来院回数閾値
	CPMV2ComingThreshold int `json:"cpm_v2_coming_threshold"`
	CPMV2GoodThreshold   int `json:"cpm_v2_good_threshold"`
	CPMV2FamilyThreshold int `json:"cpm_v2_family_threshold"`
	CPMV2NoahThreshold   int `json:"cpm_v2_noah_threshold"`
}

// UpdateLstepSettingsInput はPATCHリクエスト用（空文字=変更なし）
type UpdateLstepSettingsInput struct {
	LstepAPIKey            string
	LstepBaseURL           string
	LineChannelAccessToken string
	LineChannelSecret      string
	LiffID                 string
	LineAccountName        string
	// IsSyncEnabled が nil の場合は変更なし。false→true に変わった時のみ SyncEnabledAt を現在時刻にセット。
	IsSyncEnabled *bool
	// CPMVersion が nil の場合は変更なし。有効値は "v1" / "v2"。
	CPMVersion *string
	// DormantPrevention*Days が nil の場合は変更なし。有効値は 1 以上。
	DormantPrevention180Days *int
	DormantPrevention210Days *int
	DormantPrevention240Days *int
	DormantPrevention365Days *int
	// CPMV2*Threshold が nil の場合は変更なし。有効値は 1 以上。
	CPMV2ComingThreshold *int
	CPMV2GoodThreshold   *int
	CPMV2FamilyThreshold *int
	CPMV2NoahThreshold   *int
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
	UpdateSettings(ctx context.Context, clinicID uint64, input *UpdateLstepSettingsInput, actorID *uint64) (*LstepSettingsResponse, error)
	DeleteSettings(ctx context.Context, clinicID uint64, actorID *uint64) error
	TestConnection(ctx context.Context, clinicID uint64) (*LstepConnectionTestResult, error)
	// GetRawCredentials は復号済みの API キー・BASE URL・LINE アクセストークンを返す。
	// 設定が存在しない場合は空文字を返す（エラーにはならない）。
	GetRawCredentials(ctx context.Context, clinicID uint64) (apiKey, baseURL, lineToken string, err error)
	// IsSyncEnabled はクリニックのLステップ同期が有効かどうかを返す。
	// lstep_settings レコードが存在しない場合は false を返す（エラーにはならない）。
	IsSyncEnabled(ctx context.Context, clinicID uint64) (bool, error)
	// GetCPMVersion はクリニックの CPM 判定方式を返す。未設定時は "v1" を返す。
	GetCPMVersion(ctx context.Context, clinicID uint64) (string, error)
	// GetDormantThresholds はクリニックの dormant prevention 4 段階閾値を返す。DB 値が 0 以下なら default で補完する。
	GetDormantThresholds(ctx context.Context, clinicID uint64) (model.DormantThresholds, error)
	// GetCPMV2Thresholds はクリニックの CPM V2 来院回数閾値を返す。DB 値が 0 以下なら default で補完する。
	GetCPMV2Thresholds(ctx context.Context, clinicID uint64) (model.CPMV2Thresholds, error)
}

type lstepSettingsService struct {
	repo               repository.LstepSettingsRepository
	syncSettingsRepo   repository.LstepSyncSettingsRepository
	clinicSettingsRepo repository.ClinicSettingsRepository
	cipher             *crypto.AESGCMCipher
	auditSvc           AuditService
}

// NewLstepSettingsService は LstepSettingsService を初期化して返す。
// cipher が nil の場合は暗号化なしで動作する（開発環境で INTEGRATION_ENCRYPTION_KEY 未設定時）。
// auditSvc が nil の場合は監査ログをスキップする（CLI ツール等での使用を想定）。
// clinicSettingsRepo が nil の場合は clinic_settings の読み書きをスキップする。
func NewLstepSettingsService(repo repository.LstepSettingsRepository, syncSettingsRepo repository.LstepSyncSettingsRepository, cipher *crypto.AESGCMCipher, auditSvc AuditService, clinicSettingsRepo repository.ClinicSettingsRepository) LstepSettingsService {
	return &lstepSettingsService{repo: repo, syncSettingsRepo: syncSettingsRepo, clinicSettingsRepo: clinicSettingsRepo, cipher: cipher, auditSvc: auditSvc}
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

	if s.syncSettingsRepo != nil {
		syncSettings, syncErr := s.syncSettingsRepo.FindByClinicID(ctx, clinicID)
		if syncErr != nil && !apperrors.IsNotFound(syncErr) {
			slog.ErrorContext(ctx, "failed to find lstep sync settings", "error", syncErr, "clinic_id", clinicID)
			return nil, apperrors.Wrap(syncErr, "failed to find lstep sync settings")
		}
		if syncSettings != nil {
			resp.IsSyncEnabled = syncSettings.IsSyncEnabled
			resp.SyncEnabledAt = syncSettings.SyncEnabledAt
		}
	}

	if s.clinicSettingsRepo != nil {
		cs, csErr := s.clinicSettingsRepo.FindByClinicID(ctx, clinicID)
		if csErr != nil {
			slog.ErrorContext(ctx, "failed to find clinic settings", "error", csErr, "clinic_id", clinicID)
			return nil, apperrors.Wrap(csErr, "failed to find clinic settings")
		}
		if cs.CPMVersion == "" {
			resp.CPMVersion = "v1"
		} else {
			resp.CPMVersion = cs.CPMVersion
		}
		thresholds := model.DormantThresholds{
			Stage180: cs.DormantPrevention180Days,
			Stage210: cs.DormantPrevention210Days,
			Stage240: cs.DormantPrevention240Days,
			Stage365: cs.DormantPrevention365Days,
		}.WithDefaults()
		resp.DormantPrevention180Days = thresholds.Stage180
		resp.DormantPrevention210Days = thresholds.Stage210
		resp.DormantPrevention240Days = thresholds.Stage240
		resp.DormantPrevention365Days = thresholds.Stage365
		v2t := model.CPMV2Thresholds{
			Coming: cs.CPMV2ComingThreshold,
			Good:   cs.CPMV2GoodThreshold,
			Family: cs.CPMV2FamilyThreshold,
			Noah:   cs.CPMV2NoahThreshold,
		}.WithDefaults()
		resp.CPMV2ComingThreshold = v2t.Coming
		resp.CPMV2GoodThreshold = v2t.Good
		resp.CPMV2FamilyThreshold = v2t.Family
		resp.CPMV2NoahThreshold = v2t.Noah
	}

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

func (s *lstepSettingsService) UpdateSettings(ctx context.Context, clinicID uint64, input *UpdateLstepSettingsInput, actorID *uint64) (*LstepSettingsResponse, error) {
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

	if input.IsSyncEnabled != nil && s.syncSettingsRepo != nil {
		if err := s.updateSyncEnabled(ctx, clinicID, *input.IsSyncEnabled); err != nil {
			return nil, err
		}
	}

	if input.CPMVersion != nil && s.clinicSettingsRepo != nil {
		ver := *input.CPMVersion
		if ver != "v1" && ver != "v2" {
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("cpm_version must be 'v1' or 'v2', got %q", ver))
		}
		if err := s.clinicSettingsRepo.UpdateCPMVersion(ctx, clinicID, ver); err != nil {
			slog.ErrorContext(ctx, "failed to update cpm_version", "error", err, "clinic_id", clinicID)
			return nil, apperrors.Wrap(err, "failed to update cpm_version")
		}
	}

	if s.clinicSettingsRepo != nil &&
		(input.DormantPrevention180Days != nil || input.DormantPrevention210Days != nil ||
			input.DormantPrevention240Days != nil || input.DormantPrevention365Days != nil) {
		current, csErr := s.clinicSettingsRepo.FindByClinicID(ctx, clinicID)
		if csErr != nil {
			slog.ErrorContext(ctx, "failed to read clinic settings for dormant merge", "error", csErr, "clinic_id", clinicID)
			return nil, apperrors.Wrap(csErr, "failed to find clinic settings")
		}
		thresholds := model.DormantThresholds{
			Stage180: current.DormantPrevention180Days,
			Stage210: current.DormantPrevention210Days,
			Stage240: current.DormantPrevention240Days,
			Stage365: current.DormantPrevention365Days,
		}.WithDefaults()
		if input.DormantPrevention180Days != nil {
			if *input.DormantPrevention180Days < 1 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("dormant_prevention_180_days must be >= 1, got %d", *input.DormantPrevention180Days))
			}
			thresholds.Stage180 = *input.DormantPrevention180Days
		}
		if input.DormantPrevention210Days != nil {
			if *input.DormantPrevention210Days < 1 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("dormant_prevention_210_days must be >= 1, got %d", *input.DormantPrevention210Days))
			}
			thresholds.Stage210 = *input.DormantPrevention210Days
		}
		if input.DormantPrevention240Days != nil {
			if *input.DormantPrevention240Days < 1 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("dormant_prevention_240_days must be >= 1, got %d", *input.DormantPrevention240Days))
			}
			thresholds.Stage240 = *input.DormantPrevention240Days
		}
		if input.DormantPrevention365Days != nil {
			if *input.DormantPrevention365Days < 1 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("dormant_prevention_365_days must be >= 1, got %d", *input.DormantPrevention365Days))
			}
			thresholds.Stage365 = *input.DormantPrevention365Days
		}
		if err := s.clinicSettingsRepo.UpdateDormantThresholds(ctx, clinicID, thresholds); err != nil {
			slog.ErrorContext(ctx, "failed to update dormant thresholds", "error", err, "clinic_id", clinicID)
			return nil, apperrors.Wrap(err, "failed to update dormant thresholds")
		}
	}

	if s.clinicSettingsRepo != nil &&
		(input.CPMV2ComingThreshold != nil || input.CPMV2GoodThreshold != nil ||
			input.CPMV2FamilyThreshold != nil || input.CPMV2NoahThreshold != nil) {
		current, csErr := s.clinicSettingsRepo.FindByClinicID(ctx, clinicID)
		if csErr != nil {
			slog.ErrorContext(ctx, "failed to read clinic settings for cpm v2 merge", "error", csErr, "clinic_id", clinicID)
			return nil, apperrors.Wrap(csErr, "failed to find clinic settings")
		}
		v2t := model.CPMV2Thresholds{
			Coming: current.CPMV2ComingThreshold,
			Good:   current.CPMV2GoodThreshold,
			Family: current.CPMV2FamilyThreshold,
			Noah:   current.CPMV2NoahThreshold,
		}.WithDefaults()
		if input.CPMV2ComingThreshold != nil {
			if *input.CPMV2ComingThreshold < 1 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("cpm_v2_coming_threshold must be >= 1, got %d", *input.CPMV2ComingThreshold))
			}
			v2t.Coming = *input.CPMV2ComingThreshold
		}
		if input.CPMV2GoodThreshold != nil {
			if *input.CPMV2GoodThreshold < 1 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("cpm_v2_good_threshold must be >= 1, got %d", *input.CPMV2GoodThreshold))
			}
			v2t.Good = *input.CPMV2GoodThreshold
		}
		if input.CPMV2FamilyThreshold != nil {
			if *input.CPMV2FamilyThreshold < 1 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("cpm_v2_family_threshold must be >= 1, got %d", *input.CPMV2FamilyThreshold))
			}
			v2t.Family = *input.CPMV2FamilyThreshold
		}
		if input.CPMV2NoahThreshold != nil {
			if *input.CPMV2NoahThreshold < 1 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("cpm_v2_noah_threshold must be >= 1, got %d", *input.CPMV2NoahThreshold))
			}
			v2t.Noah = *input.CPMV2NoahThreshold
		}
		if err := s.clinicSettingsRepo.UpdateCPMV2Thresholds(ctx, clinicID, v2t); err != nil {
			slog.ErrorContext(ctx, "failed to update cpm v2 thresholds", "error", err, "clinic_id", clinicID)
			return nil, apperrors.Wrap(err, "failed to update cpm v2 thresholds")
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/v1/tags", http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()               //nolint:errcheck // close error on connectivity probe is not actionable
	_, _ = io.Copy(io.Discard, resp.Body) //nolint:errcheck // body drain failure on connectivity probe is not actionable
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("authentication failed: HTTP %d", resp.StatusCode)
	}
	return nil
}

func testLineAPI(ctx context.Context, channelToken string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.line.me/v2/bot/info", http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+channelToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()               //nolint:errcheck // close error on connectivity probe is not actionable
	_, _ = io.Copy(io.Discard, resp.Body) //nolint:errcheck // body drain failure on connectivity probe is not actionable
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

// GetCPMVersion はクリニックの CPM 判定方式を返す。レコード未存在または空文字時は "v1" を返す。
func (s *lstepSettingsService) GetCPMVersion(ctx context.Context, clinicID uint64) (string, error) {
	if s.clinicSettingsRepo == nil {
		return "v1", nil
	}
	settings, err := s.clinicSettingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return "", apperrors.Wrap(err, "failed to find clinic settings")
	}
	if settings.CPMVersion == "" {
		return "v1", nil
	}
	return settings.CPMVersion, nil
}

// GetDormantThresholds はクリニックの dormant prevention 4 段階閾値を返す。DB 値が 0 以下なら default で補完する。
func (s *lstepSettingsService) GetDormantThresholds(ctx context.Context, clinicID uint64) (model.DormantThresholds, error) {
	if s.clinicSettingsRepo == nil {
		return model.DormantThresholds{}.WithDefaults(), nil
	}
	settings, err := s.clinicSettingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get clinic settings for dormant thresholds", "clinic_id", clinicID, "error", err)
		return model.DormantThresholds{}, apperrors.Wrap(err, "failed to find clinic settings for dormant thresholds")
	}
	return model.DormantThresholds{
		Stage180: settings.DormantPrevention180Days,
		Stage210: settings.DormantPrevention210Days,
		Stage240: settings.DormantPrevention240Days,
		Stage365: settings.DormantPrevention365Days,
	}.WithDefaults(), nil
}

// GetCPMV2Thresholds はクリニックの CPM V2 来院回数閾値を返す。DB 値が 0 以下なら default で補完する。
func (s *lstepSettingsService) GetCPMV2Thresholds(ctx context.Context, clinicID uint64) (model.CPMV2Thresholds, error) {
	if s.clinicSettingsRepo == nil {
		return model.CPMV2Thresholds{}.WithDefaults(), nil
	}
	settings, err := s.clinicSettingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get clinic settings for cpm v2 thresholds", "clinic_id", clinicID, "error", err)
		return model.CPMV2Thresholds{}, apperrors.Wrap(err, "failed to find clinic settings for cpm v2 thresholds")
	}
	return model.CPMV2Thresholds{
		Coming: settings.CPMV2ComingThreshold,
		Good:   settings.CPMV2GoodThreshold,
		Family: settings.CPMV2FamilyThreshold,
		Noah:   settings.CPMV2NoahThreshold,
	}.WithDefaults(), nil
}

// IsSyncEnabled はクリニックの同期有効フラグを返す。レコード未作成時は false を返す。
func (s *lstepSettingsService) IsSyncEnabled(ctx context.Context, clinicID uint64) (bool, error) {
	if s.syncSettingsRepo == nil {
		return false, nil
	}
	settings, err := s.syncSettingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return false, nil
		}
		return false, apperrors.Wrap(err, "failed to find lstep sync settings")
	}
	return settings.IsSyncEnabled, nil
}

// updateSyncEnabled は IsSyncEnabled の変化に応じて lstep_settings を Upsert する。
// false → true の場合のみ SyncEnabledAt を現在時刻でセットする。true → false では保持する。
func (s *lstepSettingsService) updateSyncEnabled(ctx context.Context, clinicID uint64, newEnabled bool) error {
	current, err := s.syncSettingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil && !apperrors.IsNotFound(err) {
		slog.ErrorContext(ctx, "failed to find lstep sync settings", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to find lstep sync settings")
	}

	record := &model.LstepSettings{
		ClinicID:      clinicID,
		IsSyncEnabled: newEnabled,
	}

	if current != nil {
		record.SyncEnabledAt = current.SyncEnabledAt
	}

	// false → true の遷移時のみ SyncEnabledAt をセット
	currentEnabled := current != nil && current.IsSyncEnabled
	if !currentEnabled && newEnabled {
		now := time.Now()
		record.SyncEnabledAt = &now
	}

	if _, err := s.syncSettingsRepo.Upsert(ctx, record); err != nil {
		slog.ErrorContext(ctx, "failed to upsert lstep sync settings", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to update lstep sync settings")
	}
	return nil
}
