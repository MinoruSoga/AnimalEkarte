package model

import (
	"encoding/json"
	"time"
)

// AuditLog は権限変更・認証操作の記録。削除禁止テーブル。
type AuditLog struct {
	ID         uint64          `gorm:"primaryKey"   json:"id"`
	ClinicID   *uint64         `json:"clinic_id"`
	ActorID    *uint64         `json:"actor_id"`
	ActorType  string          `gorm:"not null"     json:"actor_type"`
	Action     string          `gorm:"not null"     json:"action"`
	Resource   string          `gorm:"not null"     json:"resource"`
	ResourceID *uint64         `json:"resource_id"`
	OldValue   json.RawMessage `gorm:"type:jsonb"   json:"old_value"`
	NewValue   json.RawMessage `gorm:"type:jsonb"   json:"new_value"`
	IPAddress  string          `json:"ip_address"`
	UserAgent  string          `json:"user_agent"`
	CreatedAt  time.Time       `json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_logs" }

// 監査アクション定数
const (
	AuditActionPermissionGroupCreate = "permission_group.create"
	AuditActionPermissionGroupUpdate = "permission_group.update"
	AuditActionPermissionGroupDelete = "permission_group.delete"
	AuditActionPermissionRulesUpdate = "permission_rules.update"
	AuditActionAuthLoginSuccess      = "auth.login.success"
	AuditActionAuthLoginFailure      = "auth.login.failure"
	AuditActionAuthLogout            = "auth.logout"

	// Lステップ / LINE連携 監査アクション
	AuditActionLstepSettingsSave     = "lstep.settings.save"
	AuditActionLstepTagSync          = "lstep.tag.sync"
	AuditActionLstepTagSyncBulk      = "lstep.tag.sync_bulk"
	AuditActionLineNotificationSend  = "line.notification.send"
	AuditActionOwnerLineUserIDUpdate = "owner.line_user_id.update"
	AuditActionOwnerLineUserIDUnlink = "owner.line_user_id.unlink"
)
