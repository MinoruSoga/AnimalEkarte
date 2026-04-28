package handler

// updateLstepSettingsRequest はLステップ設定更新リクエスト。空文字=変更なし。
type updateLstepSettingsRequest struct {
	LstepAPIKey            string `json:"lstep_api_key"`
	LstepBaseURL           string `json:"lstep_base_url"`
	LineChannelAccessToken string `json:"line_channel_access_token"`
	LineChannelSecret      string `json:"line_channel_secret"`
	LiffID                 string `json:"liff_id"`
	LineAccountName        string `json:"line_account_name"`
}
