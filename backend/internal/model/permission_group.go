package model

import (
	"time"

	"gorm.io/gorm"
)

// PermissionGroup は権限グループマスタ（クリニック単位で権限を管理）
type PermissionGroup struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64         `gorm:"not null"                                       json:"clinic_id"`
	Name        string         `gorm:"type:varchar(100);not null"                     json:"name"`
	Description string         `gorm:"default:''"                                     json:"description"`
	Color       string         `gorm:"type:varchar(7);default:'#6B7280'"              json:"color"`
	IsActive    bool           `gorm:"default:true"                                   json:"is_active"`
	SortOrder   int            `gorm:"type:integer;default:0"                         json:"sort_order"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt   gorm.DeletedAt `                                                      json:"-"`

	// Relations
	Rules  []PermissionGroupRule `gorm:"foreignKey:GroupID" json:"rules,omitempty"`
	Staffs []Staff               `gorm:"many2many:staff_permission_groups;foreignKey:ID;joinForeignKey:GroupID;References:ID;joinReferences:StaffID" json:"staffs,omitempty"`
}

func (PermissionGroup) TableName() string { return "permission_groups" }

// PermissionGroupRule は権限グループ内のリソース×CRUD権限
type PermissionGroupRule struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	GroupID   uint64         `gorm:"not null"                                       json:"group_id"`
	Resource  string         `gorm:"type:varchar(50);not null"                      json:"resource"`
	CanView   bool           `gorm:"default:false"                                  json:"can_view"`
	CanCreate bool           `gorm:"default:false"                                  json:"can_create"`
	CanEdit   bool           `gorm:"default:false"                                  json:"can_edit"`
	CanDelete bool           `gorm:"default:false"                                  json:"can_delete"`
	CreatedAt time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt gorm.DeletedAt `                                                      json:"-"`

	// Relations
	PermissionGroup *PermissionGroup `gorm:"foreignKey:GroupID" json:"permission_group,omitempty"`
}

func (PermissionGroupRule) TableName() string { return "permission_group_rules" }

// StaffPermissionGroup はスタッフと権限グループの中間テーブル（M:N）
type StaffPermissionGroup struct {
	StaffID   uint64    `gorm:"primaryKey;not null"                                  json:"staff_id"`
	GroupID   uint64    `gorm:"primaryKey;not null"                                  json:"group_id"`
	CreatedAt time.Time `gorm:"autoCreateTime"                                       json:"created_at"`

	// Relations
	Staff *Staff           `gorm:"foreignKey:StaffID;references:ID" json:"staff,omitempty"`
	Group *PermissionGroup `gorm:"foreignKey:GroupID;references:ID" json:"group,omitempty"`
}

func (StaffPermissionGroup) TableName() string { return "staff_permission_groups" }
