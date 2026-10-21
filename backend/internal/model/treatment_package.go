package model

import "time"

// TreatmentPackage は医院別診療項目パッケージマスタの定義行（EMR-518 §3.1 / EMR-521）。
// 削除は論理のみ。deleted_at/deleted_by は両方 NULL または両方非 NULL の pair-null 規約。
// 名前 unique は付けない（同名は identity/再送一致を証明しない）。
type TreatmentPackage struct {
	ID          uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID    uint64     `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name        string     `gorm:"column:name;type:text;not null" json:"name"`
	Description string     `gorm:"column:description;type:text;not null;default:''" json:"description"`
	IsActive    bool       `gorm:"column:is_active;type:boolean;not null" json:"is_active"`
	SortOrder   int        `gorm:"column:sort_order;type:integer;not null;default:0" json:"sort_order"`
	Version     int64      `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	CreatedBy   uint64     `gorm:"column:created_by;type:bigint;not null" json:"created_by"`
	UpdatedBy   uint64     `gorm:"column:updated_by;type:bigint;not null" json:"updated_by"`
	CreatedAt   time.Time  `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;type:timestamptz;not null;autoUpdateTime" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	DeletedBy   *uint64    `gorm:"column:deleted_by;type:bigint" json:"-"`
}

func (TreatmentPackage) TableName() string { return "treatment_packages" }
