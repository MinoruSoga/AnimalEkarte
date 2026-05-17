package handler

// updateLstepSettingsRequest はLステップ設定更新リクエスト。空文字=変更なし。
type updateLstepSettingsRequest struct {
	LstepAPIKey              string  `json:"lstep_api_key"`
	LstepBaseURL             string  `json:"lstep_base_url"`
	LineChannelAccessToken   string  `json:"line_channel_access_token"`
	LineChannelSecret        string  `json:"line_channel_secret"`
	LiffID                   string  `json:"liff_id"`
	LineAccountName          string  `json:"line_account_name"`
	IsSyncEnabled            *bool   `json:"is_sync_enabled"`
	CPMVersion               *string `json:"cpm_version"`
	DormantPrevention180Days *int    `json:"dormant_prevention_180_days"`
	DormantPrevention210Days *int    `json:"dormant_prevention_210_days"`
	DormantPrevention240Days *int    `json:"dormant_prevention_240_days"`
	DormantPrevention365Days *int    `json:"dormant_prevention_365_days"`
	CPMV2ComingThreshold     *int    `json:"cpm_v2_coming_threshold"`
	CPMV2GoodThreshold       *int    `json:"cpm_v2_good_threshold"`
	CPMV2FamilyThreshold     *int    `json:"cpm_v2_family_threshold"`
	CPMV2NoahThreshold       *int    `json:"cpm_v2_noah_threshold"`
}
