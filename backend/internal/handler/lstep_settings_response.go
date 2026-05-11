package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/service"
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
	FireHourJST                  int        `json:"fire_hour_jst"`
	CPMVersion                   string     `json:"cpm_version"`
	DormantPrevention180Days     int        `json:"dormant_prevention_180_days"`
	DormantPrevention210Days     int        `json:"dormant_prevention_210_days"`
	DormantPrevention240Days     int        `json:"dormant_prevention_240_days"`
	DormantPrevention365Days     int        `json:"dormant_prevention_365_days"`
}

// lstepConnectionTestResponse は疎通確認結果レスポンス
type lstepConnectionTestResponse struct {
	LstepOK    bool   `json:"lstep_ok"`
	LstepError string `json:"lstep_error,omitempty"`
	LineOK     bool   `json:"line_ok"`
	LineError  string `json:"line_error,omitempty"`
}

func toLstepSettingsResponse(s *service.LstepSettingsResponse) lstepSettingsResponse {
	return lstepSettingsResponse{
		LstepAPIKeyMasked:            s.LstepAPIKeyMasked,
		LstepBaseURL:                 s.LstepBaseURL,
		LineChannelAccessTokenMasked: s.LineChannelAccessTokenMasked,
		LineChannelSecretMasked:      s.LineChannelSecretMasked,
		LiffID:                       s.LiffID,
		LineAccountName:              s.LineAccountName,
		IsConfigured:                 s.IsConfigured,
		LastUpdatedAt:                s.LastUpdatedAt,
		IsSyncEnabled:                s.IsSyncEnabled,
		SyncEnabledAt:                s.SyncEnabledAt,
		FireHourJST:                  s.FireHourJST,
		CPMVersion:                   s.CPMVersion,
		DormantPrevention180Days:     s.DormantPrevention180Days,
		DormantPrevention210Days:     s.DormantPrevention210Days,
		DormantPrevention240Days:     s.DormantPrevention240Days,
		DormantPrevention365Days:     s.DormantPrevention365Days,
	}
}

func toLstepConnectionTestResponse(r *service.LstepConnectionTestResult) lstepConnectionTestResponse {
	return lstepConnectionTestResponse{
		LstepOK:    r.LstepOK,
		LstepError: r.LstepError,
		LineOK:     r.LineOK,
		LineError:  r.LineError,
	}
}
