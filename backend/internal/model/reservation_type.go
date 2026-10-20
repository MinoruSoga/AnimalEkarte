package model

import (
	"gorm.io/gorm"
	"time"
)

// ReservationDayOption defines which weekdays a service type is available for LINE reservation.
type ReservationDayOption string

const (
	DayOptionNone     ReservationDayOption = "none"
	DayOptionSaturday ReservationDayOption = "saturday"
	DayOptionWeekday  ReservationDayOption = "weekday"
	DayOptionAnyday   ReservationDayOption = "anyday"
)

// ReservationTypeCategory は予約区分のカテゴリ（'general' | 'trimming'）
type ReservationTypeCategory string

const (
	ReservationTypeCategoryGeneral  ReservationTypeCategory = "general"
	ReservationTypeCategoryTrimming ReservationTypeCategory = "trimming"
)

// UnavailableType は予約不可時間の種別
type UnavailableType string

const (
	UnavailableTypeWeekly   UnavailableType = "weekly"
	UnavailableTypeSpecific UnavailableType = "specific"
)

// AvailableSlotType は予約可能枠の種別
type AvailableSlotType string

const (
	AvailableSlotTypeWeekly   AvailableSlotType = "weekly"
	AvailableSlotTypeSpecific AvailableSlotType = "specific"
)

// ReservationType はサービス種別（予約区分）マスタ
type ReservationType struct {
	ID                     uint64                           `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID               uint64                           `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name                   string                           `gorm:"column:name;type:text;not null" json:"name"`
	IsActive               bool                             `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	Description            string                           `gorm:"column:description;type:text;not null;default:''" json:"description"`
	SortOrder              *int                             `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	GroupID                *uint64                          `gorm:"column:group_id;type:bigint" json:"group_id"`
	ReservationDisplayName string                           `gorm:"column:reservation_display_name;type:text;not null;default:''" json:"reservation_display_name"`
	DurationMinutes        int                              `gorm:"column:duration_minutes;type:int;not null" json:"duration_minutes"`
	ShortName              string                           `gorm:"column:short_name;type:text;not null;default:''" json:"short_name"`
	ShowShortName          bool                             `gorm:"column:show_short_name;type:boolean;not null;default:false" json:"show_short_name"`
	ReservationVisible     bool                             `gorm:"column:reservation_visible;type:boolean;not null;default:true" json:"reservation_visible"`
	ReservationComment     string                           `gorm:"column:reservation_comment;type:text;not null;default:''" json:"reservation_comment"`
	ReservationImageURL    string                           `gorm:"column:reservation_image_url;type:text;not null;default:''" json:"reservation_image_url"`
	ParentID               *uint64                          `gorm:"column:parent_id;type:bigint" json:"parent_id"`
	MaxConcurrent          *int                             `gorm:"column:max_concurrent;type:integer" json:"max_concurrent"`
	ReservationDayOption   ReservationDayOption             `gorm:"column:reservation_day_option;type:text;not null;default:'none'" json:"reservation_day_option"`
	IsInternal             bool                             `gorm:"column:is_internal;type:boolean;not null;default:false" json:"is_internal"`
	CreatedAt              time.Time                        `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time                        `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt              gorm.DeletedAt                   `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	BusinessKind           *string                          `gorm:"column:business_kind;type:text" json:"-"`
	UsageResumePending     *bool                            `gorm:"column:usage_resume_pending;type:boolean" json:"-"`
	NameRevision           *int64                           `gorm:"column:name_revision;type:bigint" json:"-"`
	Version                int64                            `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	Group                  *ReservationTypeGroup            `gorm:"foreignKey:GroupID;references:ID" json:"group,omitempty"`
	Parent                 *ReservationType                 `gorm:"foreignKey:ParentID;references:ID" json:"parent,omitempty"`
	Children               []ReservationType                `gorm:"foreignKey:ParentID;references:ID" json:"children,omitempty"`
	UnavailableTimes       []ReservationTypeUnavailableTime `gorm:"foreignKey:ReservationTypeID;references:ID" json:"unavailable_times,omitempty"`
	AvailableSlots         []ReservationTypeAvailableSlot   `gorm:"foreignKey:ReservationTypeID;references:ID" json:"available_slots,omitempty"`
	StaffCapabilities      []StaffReservationCapability     `gorm:"foreignKey:ReservationTypeID;references:ID" json:"staff_capabilities,omitempty"`
}

func (ReservationType) TableName() string { return "reservation_types" }

// ReservationTypeUnavailableTime は予約区分の予約不可時間帯（BE-115）
type ReservationTypeUnavailableTime struct {
	ID                uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID          uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	ReservationTypeID uint64          `gorm:"column:reservation_type_id;type:bigint;not null" json:"reservation_type_id"`
	UnavailableType   UnavailableType `gorm:"column:unavailable_type;type:text;not null" json:"unavailable_type"`
	DayOfWeek         *int16          `gorm:"column:day_of_week;type:smallint" json:"day_of_week,omitempty"`
	SpecificDate      *time.Time      `gorm:"column:specific_date;type:date" json:"specific_date,omitempty"`
	StartTime         string          `gorm:"column:start_time;type:varchar(5);not null" json:"start_time"`
	EndTime           string          `gorm:"column:end_time;type:varchar(5);not null" json:"end_time"`
	CreatedAt         time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time       `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (ReservationTypeUnavailableTime) TableName() string { return "reservation_type_unavailable_times" }

// ReservationTypeAvailableSlot は予約区分の予約可能な開始時刻
type ReservationTypeAvailableSlot struct {
	ID                uint64            `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID          uint64            `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	ReservationTypeID uint64            `gorm:"column:reservation_type_id;type:bigint;not null" json:"reservation_type_id"`
	AvailableType     AvailableSlotType `gorm:"column:available_type;type:text;not null" json:"available_type"`
	DayOfWeek         *int16            `gorm:"column:day_of_week;type:smallint" json:"day_of_week,omitempty"`
	SpecificDate      *time.Time        `gorm:"column:specific_date;type:date" json:"specific_date,omitempty"`
	StartTime         string            `gorm:"column:start_time;type:varchar(5);not null" json:"start_time"`
	IsActive          bool              `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	CreatedAt         time.Time         `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time         `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (ReservationTypeAvailableSlot) TableName() string { return "reservation_type_available_slots" }

// ReservationTypeOccupation は予約区分と職種の紐付け（M:N）（BE-115）
