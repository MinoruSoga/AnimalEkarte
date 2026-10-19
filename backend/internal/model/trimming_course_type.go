package model

import (
	"gorm.io/gorm"
	"time"
)

// TrimmingCourseType はクリニックごとのトリミングコース種別マスタ (issue #73)。
// payment_methods と同型の拡張可能マスタ。trimming_courses.course_type_id から参照される。
type TrimmingCourseType struct {
	ID        uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID  uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name      string         `gorm:"column:name;type:varchar(50);not null" json:"name"`
	SortOrder *int           `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	IsActive  bool           `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	CreatedAt time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
}

func (TrimmingCourseType) TableName() string { return "trimming_course_types" }
