package lstep

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

// lstepSettingsResponse はGETレスポンス（センシティブ値はマスク済み）
type lstepSettingsResponse struct {
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
	CPMV2ComingThreshold         int        `json:"cpm_v2_coming_threshold"`
	CPMV2GoodThreshold           int        `json:"cpm_v2_good_threshold"`
	CPMV2FamilyThreshold         int        `json:"cpm_v2_family_threshold"`
	CPMV2NoahThreshold           int        `json:"cpm_v2_noah_threshold"`
	// CPM V1 判定閾値
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
	// 健診・予防タグ判定閾値
	HealthPreventionLookbackDays int `json:"health_prevention_lookback_days"`
	VaccineDeadlineDays          int `json:"vaccine_deadline_days"`
}

// lstepConnectionTestResponse は疎通確認結果レスポンス
type lstepConnectionTestResponse struct {
	LstepOK    bool   `json:"lstep_ok"`
	LstepError string `json:"lstep_error,omitempty"`
	LineOK     bool   `json:"line_ok"`
	LineError  string `json:"line_error,omitempty"`
}

func toLstepSettingsResponse(s *LstepSettingsResponse) lstepSettingsResponse {
	return lstepSettingsResponse{
		LstepAPIKeyMasked:            s.LstepAPIKeyMasked,
		LstepBaseURL:                 s.LstepBaseURL,
		LineChannelAccessTokenMasked: s.LineChannelAccessTokenMasked,
		LineChannelSecretMasked:      s.LineChannelSecretMasked,
		LiffID:                       s.LiffID,
		LineAccountName:              s.LineAccountName,
		IsConfigured:                 s.IsConfigured,
		LastUpdatedAt:                httpapi.LocalTimePtr(s.LastUpdatedAt),
		IsSyncEnabled:                s.IsSyncEnabled,
		SyncEnabledAt:                httpapi.LocalTimePtr(s.SyncEnabledAt),
		CPMVersion:                   s.CPMVersion,
		DormantPrevention180Days:     s.DormantPrevention180Days,
		DormantPrevention210Days:     s.DormantPrevention210Days,
		DormantPrevention240Days:     s.DormantPrevention240Days,
		DormantPrevention365Days:     s.DormantPrevention365Days,
		CPMV2ComingThreshold:         s.CPMV2ComingThreshold,
		CPMV2GoodThreshold:           s.CPMV2GoodThreshold,
		CPMV2FamilyThreshold:         s.CPMV2FamilyThreshold,
		CPMV2NoahThreshold:           s.CPMV2NoahThreshold,
		CPMV1DormantDays:             s.CPMV1DormantDays,
		CPMV1NoahDays:                s.CPMV1NoahDays,
		CPMV1NoahAnnualVisits:        s.CPMV1NoahAnnualVisits,
		CPMV1NoahLTV:                 s.CPMV1NoahLTV,
		CPMV1CoreDays:                s.CPMV1CoreDays,
		CPMV1CoreAnnualVisits:        s.CPMV1CoreAnnualVisits,
		CPMV1CoreLTV:                 s.CPMV1CoreLTV,
		CPMV1SpotMinAmount:           s.CPMV1SpotMinAmount,
		CPMV1SpotInactiveDays:        s.CPMV1SpotInactiveDays,
		CPMV1GrowingMaxDays:          s.CPMV1GrowingMaxDays,
		CPMV1GrowingMinVisits:        s.CPMV1GrowingMinVisits,
		CPMV1GrowingMaxVisits:        s.CPMV1GrowingMaxVisits,
		CPMV1LTVBreakLow:             s.CPMV1LTVBreakLow,
		HealthPreventionLookbackDays: s.HealthPreventionLookbackDays,
		VaccineDeadlineDays:          s.VaccineDeadlineDays,
	}
}

func toLstepConnectionTestResponse(r *LstepConnectionTestResult) lstepConnectionTestResponse {
	return lstepConnectionTestResponse{
		LstepOK:    r.LstepOK,
		LstepError: r.LstepError,
		LineOK:     r.LineOK,
		LineError:  r.LineError,
	}
}
