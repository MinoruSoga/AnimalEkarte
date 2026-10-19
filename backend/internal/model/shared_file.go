package model

import (
	"gorm.io/gorm"
	"time"
)

// SharedFile はLINE個別送信用にアップロードされたファイルのメタデータ。
type SharedFile struct {
	ID             uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID       uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	OwnerID        *uint64        `gorm:"column:owner_id;type:bigint" json:"owner_id"`
	UploadedBy     uint64         `gorm:"column:uploaded_by;type:bigint;not null" json:"uploaded_by"`
	FileType       string         `gorm:"column:file_type;type:varchar(50);not null" json:"file_type"`
	FileName       string         `gorm:"column:file_name;type:varchar(255);not null" json:"file_name"`
	FileKey        string         `gorm:"column:file_key;type:varchar(500);not null" json:"-"`
	FileSize       int64          `gorm:"column:file_size;type:bigint;not null" json:"file_size"`
	Purpose        string         `gorm:"column:purpose;type:varchar(50);not null" json:"purpose"`
	ExpiresAt      *time.Time     `gorm:"column:expires_at;type:timestamptz" json:"expires_at"`
	CreatedAt      time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	RetentionClass string         `gorm:"column:retention_class;type:text;not null;default:'unclassified'" json:"-"`
}

func (SharedFile) TableName() string { return "shared_files" }

// ファイルタイプ定数
const (
	SharedFileTypePDF       = "pdf"
	SharedFileTypeImageJPEG = "image_jpeg"
	SharedFileTypeImagePNG  = "image_png"
)

// 用途定数
const (
	SharedFilePurposeInspectionResult = "inspection_result"
	SharedFilePurposeVaccineCert      = "vaccine_cert"
	SharedFilePurposeOther            = "other"
)

// 許可拡張子と対応 MIME タイプ
// GIF を含まないのは意図的: 共有ファイルは LINE 配信経路（line_send_service）で飼い主へ送るため、
// LINE Messaging API が対応する JPEG/PNG（+PDF）に限定する。院内保存専用のカルテ画像
// （handler/medical_record_image_request.go）は GIF 可であり、この乖離を同期してはならない。
var AllowedFileExtensions = map[string]string{
	".pdf":  "application/pdf",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
}

// AllowedFileTypes は許可されたファイルタイプの一覧
var AllowedFileTypes = map[string]string{
	".pdf":  SharedFileTypePDF,
	".jpg":  SharedFileTypeImageJPEG,
	".jpeg": SharedFileTypeImageJPEG,
	".png":  SharedFileTypeImagePNG,
}

// MaxSharedFileSizeBytes は許可する最大ファイルサイズ（LINE Messaging API制限: 10MB）
const MaxSharedFileSizeBytes = 10 * 1024 * 1024
