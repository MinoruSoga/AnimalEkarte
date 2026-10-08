package model

import "time"

type MedicalRecordAddendum struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	MedicalRecordID uint64    `gorm:"not null"                 json:"medical_record_id"`
	ClinicID        uint64    `gorm:"not null"                 json:"clinic_id"`
	AuthorUserID    uint64    `gorm:"not null"                 json:"author_user_id"`
	BeforeText      string    `gorm:"not null;default:''"      json:"before_text"`
	AfterText       string    `gorm:"not null"                 json:"after_text"`
	Reason          string    `gorm:"not null"                 json:"reason"`
	CreatedAt       time.Time `gorm:"autoCreateTime"           json:"created_at"`

	// Relations
	// AuthorStaff は追記者の氏名解決専用の preload-only リレーション。
	// repository が「親カルテの clinic_id + deleted_at IS NULL」スコープでのみ
	// Preload する。wire 露出は response DTO の author_name が担うため
	// json:"-"（非直列化）かつ tygo:"-"（models.ts へ出さない: EMR-326）。
	AuthorStaff *Staff `gorm:"foreignKey:AuthorUserID" json:"-" tygo:"-"`
}

func (MedicalRecordAddendum) TableName() string { return "medical_record_addenda" }
