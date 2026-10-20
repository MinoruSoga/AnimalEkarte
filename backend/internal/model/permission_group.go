package model

import (
	"gorm.io/gorm"
	"time"
)

// PermissionGroup は権限グループマスタ（クリニック単位で権限を管理）
type PermissionGroup struct {
	ID          uint64                `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID    uint64                `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name        string                `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Description string                `gorm:"column:description;type:text;not null;default:''" json:"description"`
	Color       string                `gorm:"column:color;type:varchar(7);not null;default:'#6B7280'" json:"color"`
	IsActive    bool                  `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	SortOrder   int                   `gorm:"column:sort_order;type:int;not null;default:0" json:"sort_order"`
	CreatedAt   time.Time             `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time             `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt        `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	Rules       []PermissionGroupRule `gorm:"foreignKey:GroupID;references:ID" json:"rules,omitempty"`
	Staffs      []Staff               `gorm:"-" json:"staffs,omitempty"`
}

func (PermissionGroup) TableName() string { return "permission_groups" }

// PermissionGroupRule は権限グループ内のリソース×CRUD権限
type PermissionGroupRule struct {
	ID              uint64           `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	GroupID         uint64           `gorm:"column:group_id;type:bigint;not null" json:"group_id"`
	Resource        string           `gorm:"column:resource;type:varchar(50);not null" json:"resource"`
	CanView         bool             `gorm:"column:can_view;type:boolean;not null;default:false" json:"can_view"`
	CanCreate       bool             `gorm:"column:can_create;type:boolean;not null;default:false" json:"can_create"`
	CanEdit         bool             `gorm:"column:can_edit;type:boolean;not null;default:false" json:"can_edit"`
	CanDelete       bool             `gorm:"column:can_delete;type:boolean;not null;default:false" json:"can_delete"`
	CreatedAt       time.Time        `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time        `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt   `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	PermissionGroup *PermissionGroup `gorm:"foreignKey:GroupID;references:ID" json:"permission_group,omitempty"`
}

func (PermissionGroupRule) TableName() string { return "permission_group_rules" }

// StaffPermissionGroup はスタッフと権限グループの中間テーブル（M:N）
type StaffPermissionGroup struct {
	StaffID   uint64           `gorm:"column:staff_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"staff_id"`
	GroupID   uint64           `gorm:"column:group_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"group_id"`
	CreatedAt time.Time        `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	Staff     *Staff           `gorm:"foreignKey:StaffID;references:ID" json:"staff,omitempty"`
	Group     *PermissionGroup `gorm:"foreignKey:GroupID;references:ID" json:"group,omitempty"`
}

func (StaffPermissionGroup) TableName() string { return "staff_permission_groups" }
