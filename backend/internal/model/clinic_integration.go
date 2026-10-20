package model

import (
	"time"
)

// ClinicIntegration はクリニックの外部サービス連携設定（AES-256-GCM暗号化済み）。
type ClinicIntegration struct {
	ID        uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID  uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Service   string    `gorm:"column:service;type:varchar(50);not null" json:"service"`
	KeyName   string    `gorm:"column:key_name;type:varchar(100);not null" json:"key_name"`
	KeyValue  string    `gorm:"column:key_value;type:text;not null" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (ClinicIntegration) TableName() string { return "clinic_integrations" }

// Lステップ連携の service 識別子
const IntegrationServiceLstep = "lstep"

// key_name 定数
//
//nolint:gosec // these are key-name identifiers (column values), not actual credentials
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
