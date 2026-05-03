package model

import "time"

// LstepTagCache はLステップタグのカルテ側キャッシュ。
// タグ操作ごとに Upsert/Delete して Lステップ API 依存を最小化する。
type LstepTagCache struct {
	ID       uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID uint64    `gorm:"not null"                 json:"clinic_id"`
	OwnerID  uint64    `gorm:"not null"                 json:"owner_id"`
	TagName  string    `gorm:"column:tag_name;not null" json:"tag_name"`
	Category string    `gorm:"column:category;not null;default:'auto'" json:"category"`
	Reason   *string   `gorm:"column:reason"            json:"reason,omitempty"`
	SyncedAt time.Time `gorm:"column:synced_at;not null;default:now()" json:"synced_at"`
}

func (LstepTagCache) TableName() string { return "lstep_tag_cache" }
