package model

import "time"

// ClinicIntegration はクリニックの外部サービス連携設定（AES-256-GCM暗号化済み）。
type ClinicIntegration struct {
	ID        uint64    `gorm:"primaryKey"  json:"id"`
	ClinicID  uint64    `gorm:"not null"    json:"clinic_id"`
	Service   string    `gorm:"not null"    json:"service"`
	KeyName   string    `gorm:"not null"    json:"key_name"`
	KeyValue  string    `gorm:"not null"    json:"-"` // 絶対にJSONに出力しない
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ClinicIntegration) TableName() string { return "clinic_integrations" }

// Lステップ連携の service 識別子
const IntegrationServiceLstep = "lstep"

// key_name 定数
const (
	IntegrationKeyLstepAPIKey            = "lstep_api_key"
	IntegrationKeyLstepBaseURL           = "lstep_base_url"
	IntegrationKeyLineChannelAccessToken = "line_channel_access_token"
	IntegrationKeyLineChannelSecret      = "line_channel_secret"
	IntegrationKeyLiffID                 = "liff_id"
	IntegrationKeyLineAccountName        = "line_account_name"
)

// 暗号化が必要な key_name セット（true=暗号化必要）
var integrationEncryptedKeys = map[string]bool{
	IntegrationKeyLstepAPIKey:            true,
	IntegrationKeyLineChannelAccessToken: true,
	IntegrationKeyLineChannelSecret:      true,
}

// IsEncryptedKey は当該 key_name が暗号化対象かを返す
func IsEncryptedKey(keyName string) bool {
	return integrationEncryptedKeys[keyName]
}
