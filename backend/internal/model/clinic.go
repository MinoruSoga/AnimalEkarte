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

type PermissionType string

const (
	PermissionAccountAdmin    PermissionType = "account_admin"
	PermissionMedical         PermissionType = "medical"
	PermissionMedicalRead     PermissionType = "medical_read"
	PermissionTrimming        PermissionType = "trimming"
	PermissionBilling         PermissionType = "billing"
	PermissionReception       PermissionType = "reception"
	PermissionHospitalization PermissionType = "hospitalization"
	PermissionMasterAdmin     PermissionType = "master_admin"
	PermissionShiftAdmin      PermissionType = "shift_admin"
	PermissionInventory       PermissionType = "inventory"
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

type UserPermission struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	UserID     uint64         `gorm:"not null"                                       json:"user_id"`
	ClinicID   uint64         `gorm:"not null"                                       json:"clinic_id"`
	Permission PermissionType `gorm:"type:permission_type;not null"                  json:"permission"`
	GrantedBy  *uint64        `                                                      json:"granted_by,omitempty"`
	GrantedAt  time.Time      `gorm:"default:now()"                                  json:"granted_at"`
}

func (UserPermission) TableName() string { return "user_permissions" }
