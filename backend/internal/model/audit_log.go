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
	// Metadata は LSTEP 操作の件数・抽出条件を保存する多次元メタデータ（ISSUE-010）。
	// resource_id 単一 ID では表現できない情報（例: 健診対象抽出のフィルタ条件 + 件数集計）を JSON で永続化する。
	Metadata  json.RawMessage `gorm:"type:jsonb"   json:"metadata"`
	IPAddress string          `json:"ip_address"`
	UserAgent string          `json:"user_agent"`
	CreatedAt time.Time       `json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_logs" }

// 監査アクション定数
const (
	AuditActorTypeStaff  = "staff"
	AuditActorTypeSystem = "system"

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

	// 取扱説明書（マニュアル）編集 監査アクション
	AuditActionManualArticleUpsert = "manual_article.upsert"
	AuditActionManualArticleDelete = "manual_article.delete"

	// 会計・返金 監査アクション（#122）
	AuditActionBillingCancel        = "billing.cancel"
	AuditActionBillingPostCloseEdit = "billing.post_close_edit"
	AuditActionBillingRefundCreate  = "billing_refund.create"
	// #189: 確定済み会計のクレジット（カード）金額の確定後訂正
	AuditActionBillingCreditCorrection = "billing.credit_correction"
)
