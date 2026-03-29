package model

import (
	"time"

	"gorm.io/gorm"
)

type UserType string

const (
	UserTypeSystemAdmin UserType = "system_admin"
	UserTypeClinicAdmin UserType = "clinic_admin"
	UserTypeStaff       UserType = "staff"
)

type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"
	AccountStatusInactive AccountStatus = "inactive"
	AccountStatusLocked   AccountStatus = "locked"
)

type Clinic struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	CompanyID          uint64    `gorm:"not null"                                       json:"company_id"`
	Name               string    `gorm:"not null"                                       json:"name"`
	PostalCode         string    `gorm:"default:''"                                     json:"postal_code"`
	Address            string    `gorm:"default:''"                                     json:"address"`
	PhoneNumber        string    `gorm:"default:''"                                     json:"phone_number"`
	FaxNumber          string    `gorm:"default:''"                                     json:"fax_number"`
	RegistrationNumber string    `gorm:"default:''"                                     json:"registration_number"`
	DirectorName       string    `gorm:"default:''"                                     json:"director_name"`
	Email              string    `gorm:"default:''"                                     json:"email"`
	Website            string    `gorm:"default:''"                                     json:"website"`
	LogoURL            string    `gorm:"default:''"                                     json:"logo_url"`
	IsActive           bool      `gorm:"default:true"                                   json:"is_active"`
	StandardTaxRate    float64   `gorm:"type:numeric;not null;default:0.10"             json:"standard_tax_rate"`
	ReducedTaxRate     float64   `gorm:"type:numeric;not null;default:0.08"             json:"reduced_tax_rate"`
	CreatedAt          time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Clinic) TableName() string { return "clinics" }

type UserAccount struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	Email           string         `gorm:"not null;uniqueIndex"                           json:"email"`
	DisplayName     string         `gorm:"not null"                                       json:"display_name"`
	DisplayNameKana string         `gorm:"default:''"                                     json:"display_name_kana"`
	UserType        UserType       `gorm:"type:user_type;not null;default:'staff'"        json:"user_type"`
	JobTitleID      *uint64        `                                                      json:"job_title_id,omitempty"`
	Status          AccountStatus  `gorm:"type:account_status;default:'active'"           json:"status"`
	AvatarURL       string         `gorm:"default:''"                                     json:"avatar_url"`
	StaffID         *uint64        `                                                      json:"staff_id,omitempty"`
	PasswordHash    string         `gorm:"not null;default:''"                            json:"-"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt       gorm.DeletedAt `                                                      json:"-" swaggerignore:"true"`

	// Relations
	Staff    *Staff    `gorm:"foreignKey:StaffID"    json:"staff,omitempty"`
	JobTitle *JobTitle `gorm:"foreignKey:JobTitleID" json:"job_title,omitempty"`
}

func (UserAccount) TableName() string { return "user_accounts" }

type UserClinicMembership struct {
	ID       uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	UserID   uint64    `gorm:"not null"                                       json:"user_id"`
	ClinicID uint64    `gorm:"not null"                                       json:"clinic_id"`
	IsMain   bool      `gorm:"default:false"                                  json:"is_main"`
	JoinedAt time.Time `gorm:"default:now()"                                  json:"joined_at"`
}

func (UserClinicMembership) TableName() string { return "user_clinic_memberships" }

// PermissionGroup は権限グループ定義（company単位で管理）
type PermissionGroup struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CompanyID   uint64         `gorm:"not null"                 json:"company_id"`
	Name        string         `gorm:"not null"                 json:"name"`
	Description string         `gorm:"default:''"               json:"description"`
	Color       string         `gorm:"default:'#6B7280'"        json:"color"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"           json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"           json:"updated_at"`
	DeletedAt   gorm.DeletedAt `                                json:"deleted_at,omitempty"`

	// Relations
	Rules []PermissionGroupRule `gorm:"foreignKey:GroupID" json:"rules,omitempty"`
}

func (PermissionGroup) TableName() string { return "permission_groups" }

// PermissionGroupRule はグループ×ページ×CRUD権限
type PermissionGroupRule struct {
	ID        uint64   `gorm:"primaryKey;autoIncrement" json:"id"`
	GroupID   uint64   `gorm:"not null"                 json:"group_id"`
	Resource  Resource `gorm:"not null"                 json:"resource"`
	CanView   bool     `gorm:"not null;default:false"   json:"can_view"`
	CanCreate bool     `gorm:"not null;default:false"   json:"can_create"`
	CanEdit   bool     `gorm:"not null;default:false"   json:"can_edit"`
	CanDelete bool     `gorm:"not null;default:false"   json:"can_delete"`
}

func (PermissionGroupRule) TableName() string { return "permission_group_rules" }

// UserPermissionGroup はユーザー→グループ紐付け
type UserPermissionGroup struct {
	UserID  uint64 `gorm:"primaryKey" json:"user_id"`
	GroupID uint64 `gorm:"primaryKey" json:"group_id"`
}

func (UserPermissionGroup) TableName() string { return "user_permission_groups" }
