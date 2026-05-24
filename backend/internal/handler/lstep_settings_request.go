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
	// CPM V1 判定閾値。nil = 変更なし。
	CPMV1DormantDays      *int   `json:"cpm_v1_dormant_days"`
	CPMV1NoahDays         *int   `json:"cpm_v1_noah_days"`
	CPMV1NoahAnnualVisits *int   `json:"cpm_v1_noah_annual_visits"`
	CPMV1NoahLTV          *int64 `json:"cpm_v1_noah_ltv"`
	CPMV1CoreDays         *int   `json:"cpm_v1_core_days"`
	CPMV1CoreAnnualVisits *int   `json:"cpm_v1_core_annual_visits"`
	CPMV1CoreLTV          *int64 `json:"cpm_v1_core_ltv"`
	CPMV1SpotMinAmount    *int64 `json:"cpm_v1_spot_min_amount"`
	CPMV1SpotInactiveDays *int   `json:"cpm_v1_spot_inactive_days"`
	CPMV1GrowingMaxDays   *int   `json:"cpm_v1_growing_max_days"`
	CPMV1GrowingMinVisits *int   `json:"cpm_v1_growing_min_visits"`
	CPMV1GrowingMaxVisits *int   `json:"cpm_v1_growing_max_visits"`
	CPMV1LTVBreakLow      *int64 `json:"cpm_v1_ltv_break_low"`
	// 健診・予防タグ判定閾値。nil = 変更なし。
	HealthPreventionLookbackDays *int `json:"health_prevention_lookback_days"`
	VaccineDeadlineDays          *int `json:"vaccine_deadline_days"`
}
