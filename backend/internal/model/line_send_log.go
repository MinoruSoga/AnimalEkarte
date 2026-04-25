package model

import "time"

type LineSendLog struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID       uint64    `gorm:"not null"                 json:"clinic_id"`
	OwnerID        uint64    `gorm:"not null"                 json:"owner_id"`
	SentByUserID   uint64    `gorm:"not null"                 json:"sent_by_user_id"`
	MessageType    string    `gorm:"not null"                 json:"message_type"`
	ContentSummary string    `gorm:"not null"                 json:"content_summary"`
	LineMessageID  *string   `gorm:"default:null"             json:"line_message_id,omitempty"`
	Status         string    `gorm:"not null"                 json:"status"`
	ErrorMessage   *string   `gorm:"default:null"             json:"error_message,omitempty"`
	SentAt         time.Time `gorm:"not null;default:now()"   json:"sent_at"`
}

func (LineSendLog) TableName() string { return "line_send_logs" }
