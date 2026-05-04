package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// LstepCsvImport は Lステップ CSV インポート履歴。
type LstepCsvImport struct {
	ID               uuid.UUID      `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID         uint64         `gorm:"column:clinic_id;not null"                                json:"clinic_id"`
	CsvType          string         `gorm:"column:csv_type;type:varchar(50);not null"                json:"csv_type"`
	FileName         string         `gorm:"column:file_name;type:varchar(255);not null"              json:"file_name"`
	UploadedByUserID *uint64        `gorm:"column:uploaded_by_user_id"                               json:"uploaded_by_user_id,omitempty"`
	RowCount         int            `gorm:"column:row_count;not null;default:0"                      json:"row_count"`
	SuccessCount     int            `gorm:"column:success_count;not null;default:0"                  json:"success_count"`
	ErrorCount       int            `gorm:"column:error_count;not null;default:0"                    json:"error_count"`
	Status           string         `gorm:"column:status;type:varchar(20);not null;default:pending"  json:"status"`
	ErrorLog         datatypes.JSON `gorm:"column:error_log;type:jsonb"                              json:"error_log,omitempty"`
	ImportedAt       *time.Time     `gorm:"column:imported_at"                                       json:"imported_at,omitempty"`
	CreatedAt        time.Time      `gorm:"column:created_at;not null;autoCreateTime"                json:"created_at"`
}

func (LstepCsvImport) TableName() string { return "lstep_csv_imports" }
