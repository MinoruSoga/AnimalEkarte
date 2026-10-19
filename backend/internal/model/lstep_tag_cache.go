package model

import (
	"time"
)

// LstepTagCache はLステップタグのカルテ側キャッシュ。
// タグ操作ごとに Upsert/Delete して Lステップ API 依存を最小化する。
type LstepTagCache struct {
	ID       uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	OwnerID  uint64    `gorm:"column:owner_id;type:bigint;not null" json:"owner_id"`
	TagName  string    `gorm:"column:tag_name;type:varchar(100);not null" json:"tag_name"`
	Category string    `gorm:"column:category;type:varchar(20);not null;default:'auto'" json:"category"`
	SyncedAt time.Time `gorm:"column:synced_at;type:timestamptz;not null;default:now()" json:"synced_at"`
	Reason   *string   `gorm:"column:reason;type:text" json:"reason,omitempty"`
}

func (LstepTagCache) TableName() string { return "lstep_tag_cache" }
