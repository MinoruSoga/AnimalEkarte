package lstep

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/model"
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
	// P9 健診・予防タグ判定閾値
	HealthPreventionLookbackDays int `json:"health_prevention_lookback_days"`
	VaccineDeadlineDays          int `json:"vaccine_deadline_days"`
}

// UpdateLstepSettingsInput はPATCHリクエスト用（空文字=変更なし）
type UpdateLstepSettingsInput struct {
	LstepAPIKey            string
	LstepBaseURL           string
	LineChannelAccessToken string
	LineChannelSecret      string
	LiffID                 string
	ClearLiffID            bool
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
	// HealthPreventionLookbackDays / VaccineDeadlineDays が nil の場合は変更なし。有効値は 1 以上。
	HealthPreventionLookbackDays *int
	VaccineDeadlineDays          *int
}

// LstepConnectionTestResult は疎通確認結果
type LstepConnectionTestResult struct {
	LstepOK    bool   `json:"lstep_ok"`
	LstepError string `json:"lstep_error,omitempty"`
	LineOK     bool   `json:"line_ok"`
	LineError  string `json:"line_error,omitempty"`
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
	repo               LstepSettingsRepository
	syncSettingsRepo   LstepSyncSettingsRepository
	clinicSettingsRepo lstepClinicSettingsRepo
	cipher             *crypto.AESGCMCipher
	auditSvc           lstepAuditLogger
	transactor         Transactor
}

// NewLstepSettingsService は LstepSettingsService を初期化して返す。
// cipher が nil の場合は暗号化なしで動作する（開発環境で INTEGRATION_ENCRYPTION_KEY 未設定時）。
// auditSvc が nil の場合は監査ログをスキップする（CLI ツール等での使用を想定）。
// clinicSettingsRepo が nil の場合は clinic_settings の読み書きをスキップする。
// optional transactor: production must pass Transactor so UpdateSettings is atomic (LSA-06 / X-06).
// When omitted, a passthrough is used (test helpers); writes then do not share an ambient tx.
func NewLstepSettingsService(repo LstepSettingsRepository, syncSettingsRepo LstepSyncSettingsRepository, cipher *crypto.AESGCMCipher, auditSvc lstepAuditLogger, clinicSettingsRepo lstepClinicSettingsRepo, transactor ...Transactor) LstepSettingsService {
	var tx Transactor
	if len(transactor) > 0 && transactor[0] != nil {
		tx = transactor[0]
	} else {
		tx = passthroughSettingsTransactor{}
	}
	return &lstepSettingsService{repo: repo, syncSettingsRepo: syncSettingsRepo, clinicSettingsRepo: clinicSettingsRepo, cipher: cipher, auditSvc: auditSvc, transactor: tx}
}

// passthroughSettingsTransactor runs fn without opening a DB transaction (test fallback only).
type passthroughSettingsTransactor struct{}

func (passthroughSettingsTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func (s *lstepSettingsService) GetSettings(ctx context.Context, clinicID uint64) (*LstepSettingsResponse, error) {
	records, err := s.repo.FindByClinicAndService(ctx, clinicID, model.IntegrationServiceLstep)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find lstep settings")
	}

	kvMap := make(map[string]string, len(records))
	var lastUpdated *time.Time
	for _, r := range records {
		val, decErr := s.decrypt(r.KeyName, r.KeyValue)
		if decErr != nil {
			// LSB-04 / DEC-35: 復号失敗を空文字へ置換して握り潰さない（サイレント停止を防ぐ）
			return nil, apperrors.Wrap(decErr, "failed to decrypt integration value")
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
			return nil, apperrors.Wrap(csErr, "failed to find clinic settings")
		}
		applyClinicSettingsToLstepResponse(resp, cs)
	}

	return resp, nil
}

// applyClinicSettingsToLstepResponse は ClinicSettings から LstepSettingsResponse への
// CPM v1/v2・dormant・health 閾値の純粋なフィールドマッピングを行う（BE-refactor.md E-1）。
func applyClinicSettingsToLstepResponse(resp *LstepSettingsResponse, cs *model.ClinicSettings) {
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
	hpt := model.HealthPreventionThresholds{
		LookbackDays:    cs.HealthPreventionLookbackDays,
		VaccineDeadline: cs.VaccineDeadlineDays,
	}.WithDefaults()
	resp.HealthPreventionLookbackDays = hpt.LookbackDays
	resp.VaccineDeadlineDays = hpt.VaccineDeadline
}

func (s *lstepSettingsService) UpdateSettings(ctx context.Context, clinicID uint64, input *UpdateLstepSettingsInput, actorID *uint64) (*LstepSettingsResponse, error) {
	// LSA-01: reject non-allowlisted lstep_base_url before any write (fail closed at service boundary).
	if input != nil && input.LstepBaseURL != "" {
		normalized, err := ValidateLstepBaseURL(input.LstepBaseURL)
		if err != nil {
			return nil, err
		}
		input.LstepBaseURL = normalized
	}
	// Pure validation before opening a transaction (cpm_version enum etc. live in updateClinicSyncConfig helpers).
	// Business graph writes (credentials + sync flag + clinic_settings) share one ambient tx (LSA-06 / X-06).
	var resp *LstepSettingsResponse
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.updateIntegrationCredentials(txCtx, clinicID, input); err != nil {
			return apperrors.Wrap(err, "failed to update integration credentials")
		}
		if input.IsSyncEnabled != nil && s.syncSettingsRepo != nil {
			if err := s.updateSyncEnabled(txCtx, clinicID, *input.IsSyncEnabled); err != nil {
				return apperrors.Wrap(err, "failed to update sync enabled")
			}
		}
		if err := s.updateClinicSyncConfig(txCtx, clinicID, input); err != nil {
			return apperrors.Wrap(err, "failed to update clinic sync config")
		}
		// Reload inside the same tx so post-commit GetSettings failure cannot invert durable success (X-01).
		got, err := s.GetSettings(txCtx, clinicID)
		if err != nil {
			return err
		}
		resp = got
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to update lstep settings in transaction")
	}

	// Audit remains best-effort after durable success (existing contract / tests).
	if s.auditSvc != nil {
		if auditErr := s.auditSvc.LogLstepOperation(ctx, clinicID, actorID, "update_settings", "clinic", &clinicID); auditErr != nil {
			slog.WarnContext(ctx, "audit log failed for update lstep settings", "error", auditErr, "clinic_id", clinicID)
		}
	}
	return resp, nil
}

func (s *lstepSettingsService) DeleteSettings(ctx context.Context, clinicID uint64, actorID *uint64) error {
	if err := s.repo.DeleteByClinicAndService(ctx, clinicID, model.IntegrationServiceLstep); err != nil {
		return apperrors.Wrap(err, "failed to delete lstep settings")
	}
	if s.auditSvc != nil {
		if auditErr := s.auditSvc.LogLstepOperation(ctx, clinicID, actorID, "delete_settings", "clinic", &clinicID); auditErr != nil {
			slog.WarnContext(ctx, "audit log failed for delete lstep settings", "error", auditErr, "clinic_id", clinicID)
		}
	}
	return nil
}
