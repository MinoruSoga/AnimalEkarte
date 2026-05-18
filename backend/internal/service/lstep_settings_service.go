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
	// P2 CPM V1 判定閾値
	CPMV1DormantDays      int   `json:"cpm_v1_dormant_days"`
	CPMV1NoahDays         int   `json:"cpm_v1_noah_days"`
	CPMV1NoahAnnualVisits int   `json:"cpm_v1_noah_annual_visits"`
	CPMV1NoahLTV          int64 `json:"cpm_v1_noah_ltv"`
	CPMV1CoreDays         int   `json:"cpm_v1_core_days"`
	CPMV1CoreAnnualVisits int   `json:"cpm_v1_core_annual_visits"`
	CPMV1CoreLTV          int64 `json:"cpm_v1_core_ltv"`
	CPMV1SpotMinAmount    int64 `json:"cpm_v1_spot_min_amount"`
	CPMV1SpotInactiveDays int   `json:"cpm_v1_spot_inactive_days"`
	CPMV1GrowingMaxDays   int   `json:"cpm_v1_growing_max_days"`
	CPMV1GrowingMinVisits int   `json:"cpm_v1_growing_min_visits"`
	CPMV1GrowingMaxVisits int   `json:"cpm_v1_growing_max_visits"`
	CPMV1LTVBreakLow      int64 `json:"cpm_v1_ltv_break_low"`
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
	// CPMV1* が nil の場合は変更なし。
	CPMV1DormantDays      *int
	CPMV1NoahDays         *int
	CPMV1NoahAnnualVisits *int
	CPMV1NoahLTV          *int64
	CPMV1CoreDays         *int
	CPMV1CoreAnnualVisits *int
	CPMV1CoreLTV          *int64
	CPMV1SpotMinAmount    *int64
	CPMV1SpotInactiveDays *int
	CPMV1GrowingMaxDays   *int
	CPMV1GrowingMinVisits *int
	CPMV1GrowingMaxVisits *int
	CPMV1LTVBreakLow      *int64
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
	// GetCPMV1Thresholds はクリニックの CPM V1 判定閾値を返す。DB 値が 0 以下なら default で補完する。
	GetCPMV1Thresholds(ctx context.Context, clinicID uint64) (model.CPMV1Thresholds, error)
	// GetHealthPreventionThresholds はクリニックの健診・予防タグ判定閾値を返す。DB 値が 0 以下なら default で補完する。
	GetHealthPreventionThresholds(ctx context.Context, clinicID uint64) (model.HealthPreventionThresholds, error)
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
		v1t := model.CPMV1Thresholds{
			DormantDays:      cs.CPMV1DormantDays,
			NoahDays:         cs.CPMV1NoahDays,
			NoahAnnualVisits: cs.CPMV1NoahAnnualVisits,
			NoahLTV:          cs.CPMV1NoahLTV,
			CoreDays:         cs.CPMV1CoreDays,
			CoreAnnualVisits: cs.CPMV1CoreAnnualVisits,
			CoreLTV:          cs.CPMV1CoreLTV,
			SpotMinAmount:    cs.CPMV1SpotMinAmount,
			SpotInactiveDays: cs.CPMV1SpotInactiveDays,
			GrowingMaxDays:   cs.CPMV1GrowingMaxDays,
			GrowingMinVisits: cs.CPMV1GrowingMinVisits,
			GrowingMaxVisits: cs.CPMV1GrowingMaxVisits,
			LTVBreakLow:      cs.CPMV1LTVBreakLow,
		}.WithDefaults()
		resp.CPMV1DormantDays = v1t.DormantDays
		resp.CPMV1NoahDays = v1t.NoahDays
		resp.CPMV1NoahAnnualVisits = v1t.NoahAnnualVisits
		resp.CPMV1NoahLTV = v1t.NoahLTV
		resp.CPMV1CoreDays = v1t.CoreDays
		resp.CPMV1CoreAnnualVisits = v1t.CoreAnnualVisits
		resp.CPMV1CoreLTV = v1t.CoreLTV
		resp.CPMV1SpotMinAmount = v1t.SpotMinAmount
		resp.CPMV1SpotInactiveDays = v1t.SpotInactiveDays
		resp.CPMV1GrowingMaxDays = v1t.GrowingMaxDays
		resp.CPMV1GrowingMinVisits = v1t.GrowingMinVisits
		resp.CPMV1GrowingMaxVisits = v1t.GrowingMaxVisits
		resp.CPMV1LTVBreakLow = v1t.LTVBreakLow
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

	if s.clinicSettingsRepo != nil &&
		(input.CPMV1DormantDays != nil || input.CPMV1NoahDays != nil || input.CPMV1NoahAnnualVisits != nil ||
			input.CPMV1NoahLTV != nil || input.CPMV1CoreDays != nil || input.CPMV1CoreAnnualVisits != nil ||
			input.CPMV1CoreLTV != nil || input.CPMV1SpotMinAmount != nil || input.CPMV1SpotInactiveDays != nil ||
			input.CPMV1GrowingMaxDays != nil || input.CPMV1GrowingMinVisits != nil || input.CPMV1GrowingMaxVisits != nil ||
			input.CPMV1LTVBreakLow != nil) {
		current, csErr := s.clinicSettingsRepo.FindByClinicID(ctx, clinicID)
		if csErr != nil {
			slog.ErrorContext(ctx, "failed to read clinic settings for cpm v1 merge", "error", csErr, "clinic_id", clinicID)
			return nil, apperrors.Wrap(csErr, "failed to find clinic settings")
		}
		v1t := model.CPMV1Thresholds{
			DormantDays:      current.CPMV1DormantDays,
			NoahDays:         current.CPMV1NoahDays,
			NoahAnnualVisits: current.CPMV1NoahAnnualVisits,
			NoahLTV:          current.CPMV1NoahLTV,
			CoreDays:         current.CPMV1CoreDays,
			CoreAnnualVisits: current.CPMV1CoreAnnualVisits,
			CoreLTV:          current.CPMV1CoreLTV,
			SpotMinAmount:    current.CPMV1SpotMinAmount,
			SpotInactiveDays: current.CPMV1SpotInactiveDays,
			GrowingMaxDays:   current.CPMV1GrowingMaxDays,
			GrowingMinVisits: current.CPMV1GrowingMinVisits,
			GrowingMaxVisits: current.CPMV1GrowingMaxVisits,
			LTVBreakLow:      current.CPMV1LTVBreakLow,
		}.WithDefaults()
		if input.CPMV1DormantDays != nil {
			if *input.CPMV1DormantDays < 1 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("cpm_v1_dormant_days must be >= 1, got %d", *input.CPMV1DormantDays))
			}
			v1t.DormantDays = *input.CPMV1DormantDays
		}
		if input.CPMV1NoahDays != nil {
			if *input.CPMV1NoahDays < 1 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("cpm_v1_noah_days must be >= 1, got %d", *input.CPMV1NoahDays))
			}
			v1t.NoahDays = *input.CPMV1NoahDays
		}
		if input.CPMV1NoahAnnualVisits != nil {
			if *input.CPMV1NoahAnnualVisits < 1 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("cpm_v1_noah_annual_visits must be >= 1, got %d", *input.CPMV1NoahAnnualVisits))
			}
			v1t.NoahAnnualVisits = *input.CPMV1NoahAnnualVisits
		}
		if input.CPMV1NoahLTV != nil {
			if *input.CPMV1NoahLTV < 0 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("cpm_v1_noah_ltv must be >= 0, got %d", *input.CPMV1NoahLTV))
			}
			v1t.NoahLTV = *input.CPMV1NoahLTV
		}
		if input.CPMV1CoreDays != nil {
			if *input.CPMV1CoreDays < 1 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("cpm_v1_core_days must be >= 1, got %d", *input.CPMV1CoreDays))
			}
			v1t.CoreDays = *input.CPMV1CoreDays
		}
		if input.CPMV1CoreAnnualVisits != nil {
			if *input.CPMV1CoreAnnualVisits < 1 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("cpm_v1_core_annual_visits must be >= 1, got %d", *input.CPMV1CoreAnnualVisits))
			}
			v1t.CoreAnnualVisits = *input.CPMV1CoreAnnualVisits
		}
		if input.CPMV1CoreLTV != nil {
			if *input.CPMV1CoreLTV < 0 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("cpm_v1_core_ltv must be >= 0, got %d", *input.CPMV1CoreLTV))
			}
			v1t.CoreLTV = *input.CPMV1CoreLTV
		}
		if input.CPMV1SpotMinAmount != nil {
			if *input.CPMV1SpotMinAmount < 0 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("cpm_v1_spot_min_amount must be >= 0, got %d", *input.CPMV1SpotMinAmount))
			}
			v1t.SpotMinAmount = *input.CPMV1SpotMinAmount
		}
		if input.CPMV1SpotInactiveDays != nil {
			if *input.CPMV1SpotInactiveDays < 1 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("cpm_v1_spot_inactive_days must be >= 1, got %d", *input.CPMV1SpotInactiveDays))
			}
			v1t.SpotInactiveDays = *input.CPMV1SpotInactiveDays
		}
		if input.CPMV1GrowingMaxDays != nil {
			if *input.CPMV1GrowingMaxDays < 1 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("cpm_v1_growing_max_days must be >= 1, got %d", *input.CPMV1GrowingMaxDays))
			}
			v1t.GrowingMaxDays = *input.CPMV1GrowingMaxDays
		}
		if input.CPMV1GrowingMinVisits != nil {
			if *input.CPMV1GrowingMinVisits < 1 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("cpm_v1_growing_min_visits must be >= 1, got %d", *input.CPMV1GrowingMinVisits))
			}
			v1t.GrowingMinVisits = *input.CPMV1GrowingMinVisits
		}
		if input.CPMV1GrowingMaxVisits != nil {
			if *input.CPMV1GrowingMaxVisits < 1 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("cpm_v1_growing_max_visits must be >= 1, got %d", *input.CPMV1GrowingMaxVisits))
			}
			v1t.GrowingMaxVisits = *input.CPMV1GrowingMaxVisits
		}
		if input.CPMV1LTVBreakLow != nil {
			if *input.CPMV1LTVBreakLow < 0 {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("cpm_v1_ltv_break_low must be >= 0, got %d", *input.CPMV1LTVBreakLow))
			}
			v1t.LTVBreakLow = *input.CPMV1LTVBreakLow
		}
		if err := s.clinicSettingsRepo.UpdateCPMV1Thresholds(ctx, clinicID, v1t); err != nil {
			slog.ErrorContext(ctx, "failed to update cpm v1 thresholds", "error", err, "clinic_id", clinicID)
			return nil, apperrors.Wrap(err, "failed to update cpm v1 thresholds")
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

// GetCPMV1Thresholds はクリニックの CPM V1 判定閾値を返す。DB 値が 0 以下なら default で補完する。
func (s *lstepSettingsService) GetCPMV1Thresholds(ctx context.Context, clinicID uint64) (model.CPMV1Thresholds, error) {
	if s.clinicSettingsRepo == nil {
		return model.CPMV1Thresholds{}.WithDefaults(), nil
	}
	settings, err := s.clinicSettingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get clinic settings for cpm v1 thresholds", "clinic_id", clinicID, "error", err)
		return model.CPMV1Thresholds{}, apperrors.Wrap(err, "failed to find clinic settings for cpm v1 thresholds")
	}
	return model.CPMV1Thresholds{
		DormantDays:      settings.CPMV1DormantDays,
		NoahDays:         settings.CPMV1NoahDays,
		NoahAnnualVisits: settings.CPMV1NoahAnnualVisits,
		NoahLTV:          settings.CPMV1NoahLTV,
		CoreDays:         settings.CPMV1CoreDays,
		CoreAnnualVisits: settings.CPMV1CoreAnnualVisits,
		CoreLTV:          settings.CPMV1CoreLTV,
		SpotMinAmount:    settings.CPMV1SpotMinAmount,
		SpotInactiveDays: settings.CPMV1SpotInactiveDays,
		GrowingMaxDays:   settings.CPMV1GrowingMaxDays,
		GrowingMinVisits: settings.CPMV1GrowingMinVisits,
		GrowingMaxVisits: settings.CPMV1GrowingMaxVisits,
		LTVBreakLow:      settings.CPMV1LTVBreakLow,
	}.WithDefaults(), nil
}

// GetHealthPreventionThresholds はクリニックの健診・予防タグ判定閾値を返す。DB 値が 0 以下なら default で補完する。
func (s *lstepSettingsService) GetHealthPreventionThresholds(ctx context.Context, clinicID uint64) (model.HealthPreventionThresholds, error) {
	if s.clinicSettingsRepo == nil {
		return model.HealthPreventionThresholds{}.WithDefaults(), nil
	}
	settings, err := s.clinicSettingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get clinic settings for health prevention thresholds", "clinic_id", clinicID, "error", err)
		return model.HealthPreventionThresholds{}, apperrors.Wrap(err, "failed to find clinic settings for health prevention thresholds")
	}
	return model.HealthPreventionThresholds{
		LookbackDays:    settings.HealthPreventionLookbackDays,
		VaccineDeadline: settings.VaccineDeadlineDays,
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
