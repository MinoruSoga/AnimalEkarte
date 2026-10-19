package model

import (
	"gorm.io/gorm"
	"time"
)

// StaffType represents the role of a staff member in the reservation system.
type StaffType string

const (
	StaffTypeDoctor   StaffType = "doctor"
	StaffTypeNurse    StaffType = "nurse"
	StaffTypeTrimmer  StaffType = "trimmer"
	StaffTypeResource StaffType = "resource"
)

type Staff struct {
	ID                     uint64                       `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID               uint64                       `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	AccountID              *uint64                      `gorm:"column:account_id;type:bigint" json:"account_id,omitempty"`
	Name                   string                       `gorm:"column:name;type:text;not null" json:"name"`
	IsActive               bool                         `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	LicenseNumber          string                       `gorm:"column:license_number;type:text;not null;default:''" json:"license_number"`
	OccupationID           *uint64                      `gorm:"column:occupation_id;type:bigint" json:"occupation_id,omitempty"`
	SortOrder              *int                         `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	StaffType              StaffType                    `gorm:"column:staff_type;type:staff_type;not null;default:'doctor'" json:"staff_type"`
	ReservationDisplayName string                       `gorm:"column:reservation_display_name;type:text;not null;default:''" json:"reservation_display_name"`
	ReservationVisible     bool                         `gorm:"column:reservation_visible;type:boolean;not null;default:true" json:"reservation_visible"`
	ReservationComment     string                       `gorm:"column:reservation_comment;type:text;not null;default:''" json:"reservation_comment"`
	ReservationImageURL    string                       `gorm:"column:reservation_image_url;type:text;not null;default:''" json:"reservation_image_url"`
	CreatedAt              time.Time                    `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time                    `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt              gorm.DeletedAt               `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	PositionName           *string                      `gorm:"column:position_name;type:text" json:"-"`
	EmploymentCategory     *string                      `gorm:"column:employment_category;type:text" json:"-"`
	Account                *Account                     `gorm:"foreignKey:AccountID;references:ID" json:"account,omitempty"`
	Occupation             *Occupation                  `gorm:"foreignKey:OccupationID;references:ID" json:"occupation,omitempty"`
	ClinicAssignments      []StaffClinicAssignment      `gorm:"foreignKey:StaffID;references:ID" json:"clinic_assignments,omitempty"`
	Capabilities           []StaffReservationCapability `gorm:"foreignKey:StaffID;references:ID" json:"capabilities,omitempty"`
}

func (Staff) TableName() string { return "staffs" }

type ShiftType string

const (
	ShiftTypeFull      ShiftType = "full"
	ShiftTypeMorning   ShiftType = "morning"
	ShiftTypeAfternoon ShiftType = "afternoon"
	ShiftTypeOff       ShiftType = "off"
	ShiftTypePaidLeave ShiftType = "paid_leave"
)

type ShiftEntry struct {
	ID        uint64            `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID  uint64            `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	StaffID   uint64            `gorm:"column:staff_id;type:bigint;not null" json:"staff_id"`
	Date      time.Time         `gorm:"column:date;type:date;not null" json:"date"`
	ShiftType ShiftType         `gorm:"column:shift_type;type:shift_type;not null" json:"shift_type"`
	Notes     string            `gorm:"column:notes;type:text;not null;default:''" json:"notes"`
	CreatedAt time.Time         `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt time.Time         `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	StartsAt  *time.Time        `gorm:"column:starts_at;type:timestamptz" json:"-"`
	EndsAt    *time.Time        `gorm:"column:ends_at;type:timestamptz" json:"-"`
	Version   int64             `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	Staff     *Staff            `gorm:"foreignKey:StaffID;references:ID" json:"staff,omitempty"`
	Breaks    []ShiftEntryBreak `gorm:"foreignKey:ShiftEntryID;references:ID" json:"breaks,omitempty"`
}

func (ShiftEntry) TableName() string { return "shift_entries" }

// ShiftTemplate はシフトテンプレートマスタを表す
type ShiftTemplate struct {
	ID        uint64               `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID  uint64               `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name      string               `gorm:"column:name;type:varchar(100);not null" json:"name"`
	ShiftType ShiftType            `gorm:"column:shift_type;type:shift_type;not null;default:'full'" json:"shift_type"`
	StartTime *string              `gorm:"column:start_time;type:time" json:"start_time,omitempty"`
	EndTime   *string              `gorm:"column:end_time;type:time" json:"end_time,omitempty"`
	Notes     string               `gorm:"column:notes;type:text;not null;default:''" json:"notes"`
	SortOrder int                  `gorm:"column:sort_order;type:integer;not null;default:0" json:"sort_order"`
	IsActive  bool                 `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	CreatedAt time.Time            `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt time.Time            `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt       `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	Breaks    []ShiftTemplateBreak `gorm:"foreignKey:ShiftTemplateID;references:ID"                      json:"breaks,omitempty"`
}

func (ShiftTemplate) TableName() string { return "shift_templates" }

// ShiftTemplateBreak はシフトテンプレートの休憩時間を表す
type ShiftTemplateBreak struct {
	ID              uint64 `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ShiftTemplateID uint64 `gorm:"column:shift_template_id;type:bigint;not null" json:"shift_template_id"`
	BreakStart      string `gorm:"column:break_start;type:time;not null" json:"break_start"`
	BreakEnd        string `gorm:"column:break_end;type:time;not null" json:"break_end"`
}

func (ShiftTemplateBreak) TableName() string { return "shift_template_breaks" }
