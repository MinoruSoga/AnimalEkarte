package model

import (
	"time"

	"gorm.io/gorm"
)

// TrimmingCourseType はクリニックごとのトリミングコース種別マスタ (issue #73)。
// payment_methods と同型の拡張可能マスタ。trimming_courses.course_type_id から参照される。
type TrimmingCourseType struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID  uint64         `gorm:"not null"                 json:"clinic_id"`
	Name      string         `gorm:"not null"                 json:"name"`
	SortOrder int            `gorm:"not null;default:0"       json:"sort_order"`
	IsActive  bool           `gorm:"not null;default:true"    json:"is_active"`
	CreatedAt time.Time      `gorm:"autoCreateTime"           json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"           json:"updated_at"`
	DeletedAt gorm.DeletedAt `                                json:"-"`
}

func (TrimmingCourseType) TableName() string { return "trimming_course_types" }
